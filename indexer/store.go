package indexer

import (
	"context"
	"crypto/sha512"
	"fmt"

	gr "github.com/ac5tin/goredis"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Post struct {
	ID        string  `db:"id" json:"id"`
	Message   string  `db:"message" json:"message"`
	Title     string  `db:"title" json:"title"`
	URL       string  `db:"url" json:"url"`
	Timestamp uint64  `db:"timestamp" json:"timestamp"`
	Domain    string  `db:"domain" json:"domain"`
	Score     float32 `db:"score" json:"score"`
}

type Store struct {
	rc *gr.Client
	pg *pgxpool.Pool
}

func NewStore(rc *gr.Client, pg *pgxpool.Pool) Store {
	return Store{
		rc,
		pg,
	}
}

func (s *Store) genPostID(url string) string {
	return fmt.Sprintf("%x", sha512.Sum512([]byte(url)))
}

// reverse index
func (s *Store) getPostIDListFromToken(token string, t *[]string) error {
	rconn := (*s.rc).Get()
	defer rconn.Close()
	res, err := redis.Strings((rconn.Do("LRANGE", token, 0, -1)))
	if err != nil {
		return err
	}
	*t = res
	return nil
}

func (s *Store) addPostLink(token string, url string) error {
	postID := s.genPostID(url)
	rconn := (*s.rc).Get()
	defer rconn.Close()
	if _, err := rconn.Do("LPUSH", token, postID); err != nil {
		return err
	}
	return nil
}

// post data
func (s *Store) getPostFromPostID(postID string, p *Post) error {
	conn, err := s.pg.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()
	if err := pgxscan.Get(context.Background(), conn, p, `
        SELECT id,message,title,url,timestamp,domain,score
        FROM posts
    `); err != nil {
		return err
	}

	return nil
}

func (s *Store) InsertPost(p *Post) error {
	conn, err := s.pg.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()
	tx, err := conn.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	if _, err = tx.Exec(context.Background(), `
        INSERT INTO posts
            (id,message,title,url,timestamp,domain,score)
            VALUES ($1,$2,$3,$4,$5,$6,$7)
    `, p.ID, p.Message, p.Title, p.URL, p.Timestamp, p.Domain, p.Score); err != nil {
		tx.Rollback(context.Background())
		return err
	}
	if err := tx.Commit(context.Background()); err != nil {
		return err
	}
	return nil
}
