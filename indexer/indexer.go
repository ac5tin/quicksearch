package indexer

const TIME_MULTIPLIER = 0.5
const MATCH_MULTIPLIER = 50.0
const SITE_MULTIPLIER = 0.1
const LANGUAGE_MULTIPLIER = 5.5
const PROTOCOL_MULTIPLIER = 2.5
const PATH_MULTIPLIER = 150.5
const PATH_QUERY_MULTIPLIER = 8.0

type Indexer struct {
	Store *Store
}

var I *Indexer = nil
