package indexer

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sync"

	"golang.org/x/sync/syncmap"
)

type quickStore struct {
	master    bool
	slaves    []string
	masterKey string
	data      syncmap.Map
}

func newQuickStore(master bool, slaves *[]string, masterKey *string) *quickStore {
	return &quickStore{
		master:    master,
		slaves:    *slaves,
		masterKey: *masterKey,
		data:      syncmap.Map{},
	}
}

// set value to store
func (q *quickStore) set(key *string, value *[]byte) {
	q.data.Store(*key, *value)
}

// get value back from store
func (q *quickStore) get(key *string, value *[]byte) error {
	if v, ok := q.data.Load(*key); ok {
		*value = v.([]byte)
		return nil
	}
	return fmt.Errorf("key not found")
}

func (q *quickStore) SetData(data *[]fullpost) error {
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

func (q *quickStore) GetData(idList *[]uint64, data *[]fullpost) error {
	if len(*idList) == 0 {
		return nil
	}
	max := runtime.NumCPU()
	cwg := new(sync.WaitGroup)
	cwg.Add(1)
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
	wg := new(sync.WaitGroup)
	for _, id := range *idList {
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
			c <- *t
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
