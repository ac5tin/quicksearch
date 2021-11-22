package indexer

const TIME_MULTIPLIER = 4.0
const MATCH_MULTIPLIER = 10.0
const SITE_MULTIPLIER = 0.1
const LANGUAGE_MULTIPLIER = 5.5

type Indexer struct {
	Store *Store
}

var I *Indexer = nil
