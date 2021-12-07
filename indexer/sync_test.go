package indexer

import (
	"fmt"
	"os"
	"quicksearch/db"
	"testing"

	gr "github.com/ac5tin/goredis"
	"github.com/joho/godotenv"
)

func TestSync(t *testing.T) {
	godotenv.Load("../.env")
	// setup store
	pg, err := db.Db(os.Getenv("DB_STRING"), os.Getenv("DB_SCHEMA"))
	if err != nil {
		t.Error(err)
	}
	rc := gr.NewRedisClient(fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")), 0, "")

	mkey := "THIS_IS_A_MASTER_KEY"
	s := NewStore(rc, pg, true, &mkey)
	I = new(Indexer)
	I.Store = &s

	t.Log("Resetting index store")
	if err := I.Store.ResetReverseIndexStore(); err != nil {
		t.Error(err)
	}
	t.Log("Successfully resetted index store")

	t.Log("Syncing ...")
	if err := I.Store.Sync(); err != nil {
		t.Error(err)
	}
	t.Log(" -- Sync complete -- ")
}
