package indexer

import (
	"fmt"
	"os"
	"quicksearch/db"
	"testing"

	gr "github.com/ac5tin/goredis"
	"github.com/joho/godotenv"
)

func TestQuery(t *testing.T) {
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

	posts := new([]Post)

	qry := "a.i data science"
	lang := "en"
	if err := I.QueryFullText(&qry, &lang, 10, 0, posts); err != nil {
		t.Error(err)
	}

	if len(*posts) == 0 {
		t.Error("Failed to query posts")
	}

	t.Logf("Number of results: %d", len(*posts))

	for _, p := range *posts {
		t.Logf("%s", p.URL)
	}
}

func TestTokenQuery(t *testing.T) {
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

	posts := new([]fullpost)
	if err := I.QueryToken("ai", posts); err != nil {
		t.Error(err)
	}

	t.Logf("Number of results: %d", len(*posts))
}
