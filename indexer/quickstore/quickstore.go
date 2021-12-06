package quickstore

import (
	"fmt"

	"golang.org/x/sync/syncmap"
)

type QuickStore struct {
	master    bool
	slaves    []string
	masterKey string
	data      syncmap.Map
}

func NewQuickStore(master bool, slaves *[]string, masterKey *string) *QuickStore {
	return &QuickStore{
		master:    master,
		slaves:    *slaves,
		masterKey: *masterKey,
		data:      syncmap.Map{},
	}
}

// set value to store
func (q *QuickStore) set(key *string, value *[]byte) {
	q.data.Store(*key, *value)
}

// get value back from store
func (q *QuickStore) get(key *string, value *[]byte) error {
	if v, ok := q.data.Load(*key); ok {
		*value = v.([]byte)
		return nil
	}
	return fmt.Errorf("key not found")
}
