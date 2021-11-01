package main

import (
	"fmt"
	"os"
	"quicksearch/db"
	"quicksearch/indexer"

	gr "github.com/ac5tin/goredis"
)

func main() {
	// init indexer
	// -- init store for indexer
	// ---- initialise postgres
	pg, err := db.Db(os.Getenv("DB_STRING"), os.Getenv("DB_SCHEMA"))
	if err != nil {
		panic(err)
	}
	// ---- initialise redis
	rc := gr.NewRedisClient(fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")), 0, "")
	s := indexer.NewStore(rc, pg)
	indexer.I = new(indexer.Indexer)
	indexer.I.Store = &s

}
