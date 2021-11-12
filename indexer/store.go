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

// Post DB table
type Post struct {
	ID            string             `db:"id" json:"-"`
	Author        string             `db:"author" json:"author"`
	Site          string             `db:"site" json:"site"`
	Title         string             `db:"title" json:"title"`
	Tokens        map[string]float32 `db:"tokens" json:"tokens"`
	Summary       string             `db:"summary" json:"summary"`
	URL           string             `db:"url" json:"url"`
	Timestamp     uint64             `db:"timestamp" json:"timestamp"`
	Language      string             `db:"language" json:"language"`
	InternalLinks []string           `db:"internal_links" json:"internal_links"`
	ExternalLinks []string           `db:"external_links" json:"external_links"`
	Entities      map[string]float32 `db:"entities" json:"entities"`
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

// generate post id
func (s *Store) genPostID(url string) string {
	return fmt.Sprintf("%x", sha512.Sum512([]byte(url)))
}

// retrieve all post IDs of a given token
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

// add a post to a given token
func (s *Store) addPostLink(token string, url string) error {
	postID := s.genPostID(url)
	rconn := (*s.rc).Get()
	defer rconn.Close()

	if _, err := rconn.Do("LPUSH", token, postID); err != nil {
		return err
	}

	return nil
}

// get full post data from a post ID
func (s *Store) getPostFromPostIDs(postID *[]string, p *[]Post) error {
	conn, err := s.pg.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	str := ""
	for i, id := range *postID {
		if i > 0 && i < len(*postID)-1 {
			str += ","
		}
		str += fmt.Sprintf("'%s'", id)
	}

	if err := pgxscan.Select(context.Background(), conn, p, `
        SELECT id,title,url,timestamp,site,author,language,summary,tokens,internal_links,external_links,entities
        FROM posts
		WHERE id IN (%s)
    `, str); err != nil {
		return err
	}

	return nil
}

// insert (index) a post to store (redis and postgres)
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
            (id,author,site,title,url,timestamp,language,summary,tokens,internal_links,external_links,entities)
            VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			ON CONFLICT (id)
			DO UPDATE SET
				author = $2,
				site = $3,
				title = $4,
				url = $5,
				timestamp = $6,
				language = $7,
				summary = $8,
				tokens = $9,
				internal_links = $10,
				external_links = $11,
				entities = $12
    `, s.genPostID(p.URL), p.Author, p.Site, p.Title, p.URL, p.Timestamp, p.Language, p.Summary, p.Tokens, p.InternalLinks, p.ExternalLinks, p.Entities); err != nil {
		tx.Rollback(context.Background())
		return err
	}
	if err := tx.Commit(context.Background()); err != nil {
		return err
	}

	tokens := make(map[string]interface{})
	for k := range p.Tokens {
		tokens[k] = struct{}{}
	}
	for k := range p.Entities {
		tokens[k] = struct{}{}
	}

	for k := range tokens {
		if err := s.addPostLink(k, p.URL); err != nil {
			tx.Rollback(context.Background())
			return err
		}
	}
	return nil
}
