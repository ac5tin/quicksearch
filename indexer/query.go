package indexer

import (
	"quicksearch/textprocessor"
	"sort"
)

func (ind *Indexer) QueryFullText(qry string, num, offset uint32, t *[]Post) error {
	// TODO
	// - tokenise
	tp := new(textprocessor.TextProcessor)

	lang := new(string)
	if err := tp.LangDetect(qry, lang); err != nil {
		return err
	}

	tokens := new([]textprocessor.Token)
	if err := tp.Tokenise(textprocessor.InputText{Lang: *lang, Text: qry}, tokens); err != nil {
		return err
	}
	// - get URLs per token
	tokenMap := make(map[*textprocessor.Token]*[]Post) // posts for each token

	type post struct {
		Post
		score float32
	}
	allPosts := new([]post)
	postMap := make(map[string]*post) // post for each postID
	// -- use loop instead of parallelism to reduce store i/o
	for _, token := range *tokens {
		posts := new([]Post)
		if err := ind.QueryToken(token.Token, posts); err != nil {
			return err
		}
		tokenMap[&token] = posts
		for _, p := range *posts {
			// -- post score = token score + token heuristics + site score + timestamp score + matches
			var score float32 = 0.0
			if s, ok := postMap[p.ID]; ok {
				score = s.score
			}

			score += token.Score + 10.0
			postMap[p.ID] = &post{p, score}

		}
	}
	for _, p := range postMap {
		*allPosts = append(*allPosts, *p)
	}
	// - rank by highest score
	sort.SliceStable(*allPosts, func(i, j int) bool {
		return (*allPosts)[i].score > (*allPosts)[j].score
	})

	// seach num offset
	if uint32(len(*allPosts)) > num+offset {
		*allPosts = (*allPosts)[offset : offset+num]
	}

	for _, p := range *allPosts {
		*t = append(*t, p.Post)
	}

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
