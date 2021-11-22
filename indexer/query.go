package indexer

import (
	"context"
	"log"
	"net/url"
	"quicksearch/textprocessor"
	"sort"
	"time"

	"golang.org/x/sync/errgroup"
)

func (ind *Indexer) QueryFullText(qry, lang *string, num, offset uint32, t *[]Post) error {
	// TODO
	// - tokenise
	tp := new(textprocessor.TextProcessor)

	tokens := new([]textprocessor.Token)
	if err := tp.Tokenise(textprocessor.InputText{Lang: *lang, Text: *qry}, tokens); err != nil {
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
	postMatches := make(map[string]uint32)

	// -- use parallelism to speed up query process
	g, ctx := errgroup.WithContext(context.Background())
	for _, token := range *tokens {
		t := token
		g.Go(func() error {
			select {
			case <-ctx.Done():
				log.Println("Error occured in another goroutine")
				return nil
			default:
				// dont block
			}
			posts := new([]Post)
			if err := ind.QueryToken(t.Token, posts); err != nil {
				return err
			}
			tokenMap[&t] = posts

			for _, p := range *posts {
				// -- post score = token score + token heuristics + site score + timestamp score + matches
				var score float32 = 0.0
				if s, ok := postMap[p.ID]; ok {
					score = s.score
				} else {
					// first time we see this post
					// add site score and timestamp score
					// - site score from db
					// - timestamp score = 1.0 / (1.0 + (timestamp - now) / (24 * 60 * 60))
					if p.Timestamp == 0 {
						p.Timestamp = uint64(time.Now().Unix() - 604800) // 1 week ago
					}

					tsScore := float32(1.0 / float64(1.0+float64(int64(p.Timestamp-uint64(time.Now().Unix())))/float64(24*60*60)))
					score += tsScore * TIME_MULTIPLIER
				}
				if h, ok := p.TokensH[t.Token]; ok {
					score += h
				}
				score += p.Tokens[t.Token]

				if v, ok := postMatches[p.ID]; ok {
					postMatches[p.ID] = v + 1
				} else {
					postMatches[p.ID] = 1
				}
				score *= float32(postMatches[p.ID]) * MATCH_MULTIPLIER

				// language score
				if *lang == p.Language {
					score *= LANGUAGE_MULTIPLIER
				}

				// protocol score (https)
				{

					if u, err := url.Parse(p.URL); err == nil {
						switch u.Scheme {
						case "https":
							score *= PROTOCOL_MULTIPLIER
						default:
							// none
						}
					}
				}

				postMap[p.ID] = &post{p, score}
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return err
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
		//log.Println(p.URL, p.score) // debug (print score)
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
