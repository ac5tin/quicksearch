package indexer

import (
	"context"
	"net/url"
	"quicksearch/utils"

	gr "github.com/ac5tin/goredis"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Post DB table
type Post struct {
	ID            uint64             `db:"id" json:"-"`
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

type fullpost struct {
	Post
	SiteScore float32 `db:"site_score"`
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

// retrieve all post IDs of a given token
func (s *Store) getPostIDListFromToken(token *string, t *[]uint64) error {
	rconn := (*s.rc).Get()
	defer rconn.Close()
	res, err := redis.Int64s((rconn.Do("LRANGE", token, 0, -1)))
	if err != nil {
		return err
	}

	for _, r := range res {
		*t = append(*t, uint64(r))
	}
	return nil
}

// add a post to a given token
func (s *Store) addPostLink(token *string, id *uint64) error {
	rconn := (*s.rc).Get()
	defer rconn.Close()

	if _, err := rconn.Do("LPUSH", token, id); err != nil {
		return err
	}

	return nil
}

func (s *Store) delPostLink(token *string, id *uint64) error {
	rconn := (*s.rc).Get()
	defer rconn.Close()

	if _, err := rconn.Do("LREM", token, 0, id); err != nil {
		return err
	}

	return nil
}

// get full post data from a post ID
func (s *Store) getPostFromPostIDs(postID *[]uint64, p *[]fullpost) error {
	conn, err := s.pg.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	if err := pgxscan.Select(context.Background(), conn, p, `
	SELECT id,title,url,timestamp,posts.site,author,language,summary,tokens,tokens_h,internal_links,external_links,entities,external_site_scores,sites.score as site_score
	FROM posts
	LEFT JOIN sites
    	ON sites.site = posts.site
	WHERE id =ANY($1::VARCHAR[])
	`, *postID); err != nil {
		return err
	}

	return nil
}

// returns true if exist
func (s *Store) getPostsFromURL(url *string, p *[]fullpost) error {
	conn, err := s.pg.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	if err := pgxscan.Select(context.Background(), conn, p, `
	SELECT id,title,url,timestamp,posts.site,author,language,summary,tokens,tokens_h,internal_links,external_links,entities,external_site_scores,sites.score as site_score
	FROM posts
	LEFT JOIN sites
    	ON sites.site = posts.site
	WHERE url = $1
	`, *url); err != nil {
		return err
	}

	return nil
}

// insert (index) a post to store (redis and postgres)
func (s *Store) InsertPost(p *Post) error {
	// check if already exist
	update := false
	posts := new([]fullpost)
	if err := s.getPostsFromURL(&p.URL, posts); err != nil {
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
				s.delPostLink(&t, &post.ID)
			}
			for t := range post.Entities {
				s.delPostLink(&t, &post.ID)
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
			*selfScore = SITE_MULTIPLIER
			if err := s.upsertSiteScore(&p.Site, selfScore); err != nil {
				return err
			}
			p.ExternalSiteScores[p.Site] = *selfScore
		}

		addScore := *selfScore*SITE_MULTIPLIER + SITE_MULTIPLIER
		if addScore > (SITE_MULTIPLIER * 10) {
			addScore = (SITE_MULTIPLIER * 10)
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

	var max_str_len uint32 = 255
	utils.TruncateString(&p.Title, &max_str_len)
	utils.TruncateString(&p.Author, &max_str_len)

	rowID := new(uint64)

	if err = pgxscan.Get(context.Background(), conn, rowID, `
        INSERT INTO posts
            (author,site,title,url,timestamp,language,summary,tokens,internal_links,external_links,entities,external_site_scores)
            VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			ON CONFLICT (url)
			DO UPDATE SET
				author = $1,
				site = $2,
				title = $3,
				url = $4,
				timestamp = $5,
				language = $6,
				summary = $7,
				tokens = $8,
				internal_links = $9,
				external_links = $10,
				entities = $11,
				external_site_scores = $12
		RETURNING id
    `, p.Author, p.Site, p.Title, p.URL, p.Timestamp, p.Language, p.Summary, p.Tokens, p.InternalLinks, p.ExternalLinks, p.Entities, p.ExternalSiteScores); err != nil {
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
		if err := s.addPostLink(&k, &p.ID); err != nil {
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

func (s *Store) DeletePost(url *string) error {
	posts := new([]fullpost)
	if err := s.getPostsFromURL(url, posts); err != nil {
		return err
	}
	// cleanup , reset scores
	{
		for _, p := range *posts {
			// remove from redis
			for t := range p.Tokens {
				s.delPostLink(&t, &p.ID)
			}
			for t := range p.Entities {
				s.delPostLink(&t, &p.ID)
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
		WHERE url = $1
	`, url); err != nil {
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
