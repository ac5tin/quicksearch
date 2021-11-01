package indexer

func (ind *Indexer) QueryFullText(qry string, t *[]Post) error {
	// TODO
	// - tokenise
	// - get URLs per token
	// - rank by highest score
	return nil
}

func (ind *Indexer) QueryToken(token string, t *[]string) error {
	// TODO
	// - get all postID back from token
	return nil
}
