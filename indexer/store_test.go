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
		ID:            "test",
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
	if err := I.Store.DeletePost("https://example.com"); err != nil {
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

	if errr := I.Store.DeletePost("https://www.fasta.ai"); err != nil {
		t.Error(errr)
	}
}
