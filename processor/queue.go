package processor

import (
	"log"
	"sync"
	"time"
)

const MAX_PARALLEL = 10
const MAX_RETRY = 5

var queue []*Results = []*Results{}

var QChan chan *Results = make(chan *Results)

func queueProcessor(q []*Results) {
	wg := new(sync.WaitGroup)
	wg.Add(len(q))
	for _, x := range q {
		go func(r *Results) {
			defer wg.Done()
			// retry
			for i := 0; i < MAX_RETRY; i++ {
				if err := processPostResults(r); err != nil {
					if i == MAX_RETRY-1 {
						log.Printf("Failed to process %s too many times, aborting ... | ERR: %s", r.URL, err.Error())
						return
					}
					log.Printf("Failed to process %s, retrying ... ", r.URL)
					time.Sleep(5 * time.Second)
					continue
				}
				break
			}
			log.Printf("Successfully indexed %s", r.URL) // debug
		}(x)
	}
	wg.Wait()
}

func ProcessQueue() {
	go func() {
		for {
			r := <-QChan
			queue = append(queue, r)
		}
	}()

	for {
		capAt := len(queue)
		log.Printf("Current Queue Length: %d", capAt)
		if len(queue) > MAX_PARALLEL {
			capAt = MAX_PARALLEL
		}
		cappedQ, newq := queue[:capAt], queue[capAt:]
		queue = newq
		queueProcessor(cappedQ)
		time.Sleep(10 * time.Second)
	}
}
