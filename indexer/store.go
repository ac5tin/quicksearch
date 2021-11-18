package indexer

import (
	"context"
	"crypto/sha512"
	"fmt"
	"net/url"
	"quicksearch/utils"

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
	TokensH       map[string]float32 `db:"tokens_h" json:"tokens_h"`
	Summary       string             `db:"summary" json:"summary"`
	URL           string             `db:"url" json:"url"`
	Timestamp     uint64             `db:"timestamp" json:"timestamp"`
	Language      string             `db:"language" json:"language"`
	InternalLinks []string           `db:"internal_links" json:"internal_links"`
	ExternalLinks []string           `db:"external_links" json:"external_links"`
	Entities      map[string]float32 `db:"entities" json:"entities"`
	// logging down scores added to external sites
	ExternalSiteScores map[string]float32 `db:"external_site_scores"`
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

func (s *Store) delPostLink(token, url string) error {
	postID := s.genPostID(url)
	rconn := (*s.rc).Get()
	defer rconn.Close()

	if _, err := rconn.Do("LREM", token, 0, postID); err != nil {
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
		if i > 0 && i <= len(*postID)-1 {
			str += ","
		}
		str += fmt.Sprintf("'%s'", id)
	}

	if err := pgxscan.Select(context.Background(), conn, p, fmt.Sprintf(`
	SELECT id,title,url,timestamp,site,author,language,summary,tokens,tokens_h,internal_links,external_links,entities,external_site_scores
	FROM posts
	WHERE id IN (%s)
	`, str)); err != nil {
		return err
	}

	return nil
}

// insert (index) a post to store (redis and postgres)
func (s *Store) InsertPost(p *Post) error {
	// check if already exist
	update := false
	posts := new([]Post)
	if err := s.getPostFromPostIDs(&[]string{s.genPostID(p.URL)}, posts); err != nil {
		return err
	}
	if len(*posts) > 0 {
		update = true
	}
	// handle update
	{
		if update {
			post := (*posts)[0]
			// remove from redis
			for t := range post.Tokens {
				s.delPostLink(t, post.URL)
			}
			for t := range post.Entities {
				s.delPostLink(t, post.URL)
			}
			// subtract from site scores
			for k, v := range post.ExternalSiteScores {
				score := new(float32)
				if err := s.getSiteScore(&k, score); err != nil {
					return err
				}
				*score -= v
				if err := s.upsertSiteScore(&k, score); err != nil {
					return err
				}
			}
		}
	}
	// handle external and internal links
	{
		p.ExternalSiteScores = make(map[string]float32)
		// - if self.site no score then set self.site.score = 0.1
		// - each external link.score += self.site.score * 0.1  + 0.1 // max cap = 1
		selfScore := new(float32)
		if err := s.getSiteScore(&p.Site, selfScore); err != nil {
			return err
		}
		if *selfScore == 0 {
			// first time we see this site, init
			*selfScore = 0.1
			if err := s.upsertSiteScore(&p.Site, selfScore); err != nil {
				return err
			}
			p.ExternalSiteScores[p.Site] = *selfScore
		}

		addScore := *selfScore*0.1 + 0.1
		if addScore > 1 {
			addScore = 1
		}

		// dedupe host
		hosts := make(map[string]interface{})
		for _, link := range p.ExternalLinks {
			u, err := url.Parse(link)
			if err != nil {
				return err
			}
			if _, ok := hosts[u.Host]; ok {
				continue
			} else {
				hosts[u.Host] = struct{}{}
			}

			score := new(float32)
			if err := s.getSiteScore(&u.Host, score); err != nil {
				return err
			}
			*score += addScore
			if err := s.upsertSiteScore(&u.Host, score); err != nil {
				return err
			}
			// external site scores
			p.ExternalSiteScores[u.Host] = addScore

		}
		hosts = nil // gc
	}

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

	var max_str_len uint32 = 255
	utils.TruncateString(&p.Title, &max_str_len)
	utils.TruncateString(&p.Author, &max_str_len)

	if _, err = tx.Exec(context.Background(), `
        INSERT INTO posts
            (id,author,site,title,url,timestamp,language,summary,tokens,internal_links,external_links,entities,external_site_scores)
            VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
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
				entities = $12,
				external_site_scores = $13
    `, s.genPostID(p.URL), p.Author, p.Site, p.Title, p.URL, p.Timestamp, p.Language, p.Summary, p.Tokens, p.InternalLinks, p.ExternalLinks, p.Entities, p.ExternalSiteScores); err != nil {
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

func (s *Store) upsertSiteScore(site *string, score *float32) error {
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
	if _, err := tx.Exec(context.Background(), `
		INSERT INTO sites (site, score) VALUES ($1, $2)
		ON CONFLICT (site) DO UPDATE SET score = $2
	`, *site, *score); err != nil {
		return err
	}
	if err := tx.Commit(context.Background()); err != nil {
		return err
	}
	return nil
}

func (s *Store) getSiteScore(site *string, score *float32) error {
	conn, err := s.pg.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	scores := new([]float32)
	if err := pgxscan.Select(context.Background(), conn, scores, `
		SELECT score FROM sites WHERE site = $1
	`, *site); err != nil {
		return err
	}

	if len(*scores) == 0 {
		*score = 0
		return nil
	}

	*score = (*scores)[0]
	return nil
}

func (s *Store) DeletePost(url string) error {
	posts := new([]Post)
	if err := s.getPostFromPostIDs(&[]string{s.genPostID(url)}, posts); err != nil {
		return err
	}
	// cleanup , reset scores
	{
		for _, p := range *posts {
			// remove from redis
			for t := range p.Tokens {
				s.delPostLink(t, p.URL)
			}
			for t := range p.Entities {
				s.delPostLink(t, p.URL)
			}
			// subtract from site scores
			for k, v := range p.ExternalSiteScores {
				score := new(float32)
				if err := s.getSiteScore(&k, score); err != nil {
					return err
				}
				*score -= v
				if err := s.upsertSiteScore(&k, score); err != nil {
					return err
				}
			}

		}
	}

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
	if _, err := tx.Exec(context.Background(), `
		DELETE FROM posts
		WHERE id = $1
	`, s.genPostID(url)); err != nil {
		tx.Rollback(context.Background())
		return err
	}

	if err := tx.Commit(context.Background()); err != nil {
		return err
	}

	return nil
}

// Full reeset store
func (s *Store) Reset() error {
	rconn := (*s.rc).Get()
	defer rconn.Close()

	if _, err := rconn.Do("FLUSHALL"); err != nil {
		return err
	}

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
	if _, err := tx.Exec(context.Background(), `
		DELETE FROM posts
	`); err != nil {
		tx.Rollback(context.Background())
		return err
	}

	if _, err := tx.Exec(context.Background(), `
		DELETE FROM sites
	`); err != nil {
		tx.Rollback(context.Background())
		return err
	}

	if err := tx.Commit(context.Background()); err != nil {
		return err
	}

	return nil
}

func (s *Store) getAllPosts(posts *[]Post) error {
	conn, err := s.pg.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	if err := pgxscan.Select(context.Background(), conn, posts, `
	SELECT id,title,url,timestamp,site,author,language,summary,tokens,tokens_h,internal_links,external_links,entities,external_site_scores
	FROM posts
	`); err != nil {
		return err
	}

	return nil
}

func (s *Store) ResetReverseIndexStore() error {
	rconn := (*s.rc).Get()
	defer rconn.Close()

	if _, err := rconn.Do("FLUSHALL"); err != nil {
		return err
	}

	return nil
}
