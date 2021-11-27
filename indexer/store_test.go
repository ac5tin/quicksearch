package indexer

import (
	"fmt"
	"os"
	"quicksearch/db"
	"testing"
	"time"

	gr "github.com/ac5tin/goredis"
	"github.com/joho/godotenv"
)

func TestStore(t *testing.T) {
	godotenv.Load("../.env")
	// setup store
	pg, err := db.Db(os.Getenv("DB_STRING"), os.Getenv("DB_SCHEMA"))
	if err != nil {
		t.Error(err)
	}
	rc := gr.NewRedisClient(fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")), 0, "")
	s := NewStore(rc, pg)
	I = new(Indexer)
	I.Store = &s

	post := Post{
		ID:            1,
		Author:        "test author",
		Title:         "test",
		Tokens:        map[string]float32{"test": 1.2},
		Summary:       "testing test 123 abc",
		URL:           "https://example.com",
		Timestamp:     uint64(time.Now().Unix()),
		Language:      "en",
		InternalLinks: []string{"https://example.com/abc"},
		ExternalLinks: []string{"https://abc.com"},
		Entities:      map[string]float32{"testing": 2.1, "test": 11.3},
	}
	t.Log("Inserting post")
	if err := I.Store.InsertPost(&post); err != nil {
		t.Error(err)
	}

	t.Log("Deleting post")
	url := "https://example.com"
	if err := I.Store.DeletePost(&url); err != nil {
		t.Error(err)
	}
}

func TestResetStore(t *testing.T) {
	godotenv.Load("../.env")
	// setup store
	pg, err := db.Db(os.Getenv("DB_STRING"), os.Getenv("DB_SCHEMA"))
	if err != nil {
		t.Error(err)
	}
	rc := gr.NewRedisClient(fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")), 0, "")
	s := NewStore(rc, pg)
	I = new(Indexer)
	I.Store = &s

	if err := I.Store.Reset(); err != nil {
		t.Error(err)
	}
}

func TestRemovePost(t *testing.T) {
	godotenv.Load("../.env")
	// setup store
	pg, err := db.Db(os.Getenv("DB_STRING"), os.Getenv("DB_SCHEMA"))
	if err != nil {
		t.Error(err)
	}
	rc := gr.NewRedisClient(fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")), 0, "")
	s := NewStore(rc, pg)
	I = new(Indexer)
	I.Store = &s

	url := "https://www.fasta.ai"
	if errr := I.Store.DeletePost(&url); err != nil {
		t.Error(errr)
	}
}

func TestSetSiteToken(t *testing.T) {
	godotenv.Load("../.env")
	// setup store
	pg, err := db.Db(os.Getenv("DB_STRING"), os.Getenv("DB_SCHEMA"))
	if err != nil {
		t.Error(err)
	}
	rc := gr.NewRedisClient(fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")), 0, "")
	s := NewStore(rc, pg)
	I = new(Indexer)
	I.Store = &s

	site := "en.wikipedia.org"
	tokens := make(map[string]float32)
	tokens["wiki"] = 40.5
	tokens["en"] = 4.5
	tokens["english"] = 4.0
	if err := I.Store.SetSiteTokens(&site, &tokens); err != nil {
		t.Error(err)
	}
}

func TestSetHToken(t *testing.T) {
	godotenv.Load("../.env")
	// setup store
	pg, err := db.Db(os.Getenv("DB_STRING"), os.Getenv("DB_SCHEMA"))
	if err != nil {
		t.Error(err)
	}
	rc := gr.NewRedisClient(fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")), 0, "")
	s := NewStore(rc, pg)
	I = new(Indexer)
	I.Store = &s

	url := "https://www.facebook.com/"
	tokens := make(map[string]float32)
	tokens["facebook"] = 500.5
	if err := I.Store.SetPostHTokens(&url, &tokens); err != nil {
		t.Error(err)
	}
}
