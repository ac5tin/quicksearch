package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Db database connection using pgx instead of sqlx
func Db(connstring, schema string) (*pgxpool.Pool, error) {
	// config
	config, err := pgxpool.ParseConfig(connstring)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		if _, err := conn.Exec(context.Background(), fmt.Sprintf(`SET search_path TO %s;`, schema)); err != nil {
			return err
		}
		return nil
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	// succesfully created pool instance

	// set default schema
	pool.Exec(context.Background(), fmt.Sprintf(`SET search_path TO %s;`, schema))
	log.Println("Connected to db")
	return pool, nil
}
