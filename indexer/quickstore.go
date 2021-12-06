package indexer

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

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

func (q *QuickStore) SetData(data *[]fullpost) error {
	for _, d := range *data {
		b, err := json.Marshal(&d)
		if err != nil {
			return err
		}
		id := fmt.Sprintf("%d", d.ID)
		q.set(&id, &b)
	}
	return nil
}

func (q *QuickStore) GetData(idList *[]uint64, data *[]fullpost) error {
	max := 10
	cwg := new(sync.WaitGroup)
	cwg.Add(0)
	c := make(chan fullpost, len(*idList))
	go func() {
		defer cwg.Done()
		for {
			r := <-c
			*data = append(*data, r)
			if len(*data) == len(*idList) {
				break
			}
		}
	}()
	multi := max
	for _, id := range *idList {
		wg := new(sync.WaitGroup)
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := fmt.Sprintf("%d", id)
			b := new([]byte)
			if err := q.get(&key, b); err != nil {
				log.Printf("error getting data from quickstore: %s", err.Error())
				return
			}
			t := new(fullpost)
			if err := json.Unmarshal(*b, t); err != nil {
				log.Printf("error unmarshalling data from quickstore: %s", err.Error())
				return
			}
		}()
		multi--
		if multi == 0 {
			wg.Wait()
			multi = max
		}
	}
	cwg.Wait()
	return nil
}
