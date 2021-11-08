package indexer

import (
	"quicksearch/textprocessor"
)

func (ind *Indexer) QueryFullText(qry string, t *[]Post) error {
	// TODO
	// - tokenise
	tp := new(textprocessor.TextProcessor)
	tokens := new([]textprocessor.Token)
	if err := tp.Tokenise(qry, tokens); err != nil {
		return err
	}
	// - get URLs per token
	tokenMap := make(map[*textprocessor.Token]*[]Post)
	// -- use loop instead of parallelism to reduce store i/o
	for _, token := range *tokens {
		posts := new([]Post)
		if err := ind.QueryToken(token.Text, posts); err != nil {
			return err
		}
		tokenMap[&token] = posts
	}
	// - rank by highest score
	// -- post score = token score + token heuristics + site score + timestamp score + matches

	return nil
}

func (ind *Indexer) QueryToken(token string, t *[]Post) error {
	// TODO
	// - get all postID back from token
	postIds := new([]string)
	if err := ind.Store.getPostIDListFromToken(token, postIds); err != nil {
		return err
	}

	// - get all post back from postID
	if err := ind.Store.getPostFromPostIDs(postIds, t); err != nil {
		return err
	}
	return nil
}
