package indexer

import (
	gr "github.com/ac5tin/goredis"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Post struct {
	ID        string
	Message   string
	Title     string
	URL       string
	Timestamp uint64
	Domain    string
	Score     float32
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
