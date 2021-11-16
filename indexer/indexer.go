package indexer

const TIME_MULTIPLIER = 3.0
const MATCH_MULTIPLIER = 10.0

type Indexer struct {
	Store *Store
}

var I *Indexer = nil
