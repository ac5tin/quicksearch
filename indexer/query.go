package indexer

import (
	"context"
	"log"
	"math"
	"net/url"
	"quicksearch/textprocessor"
	"sort"
	"strings"
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
	tokenMap := make(map[*textprocessor.Token]*[]fullpost) // posts for each token

	type post struct {
		fullpost
		score       float32            // final score (sum of all scores)
		baseScore   float32            // initial score
		tokenScores map[string]float32 // score of each token
	}
	allPosts := new([]post)
	postMap := make(map[uint64]*post) // post for each postID
	//postMatches := make(map[uint64]uint32)

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
			posts := new([]fullpost)
			if err := ind.QueryToken(&t.Token, posts); err != nil {
				return err
			}
			tokenMap[&t] = posts

			for _, p := range *posts {
				// base score
				{
					// -- post score = token score + token heuristics + site tokens + site score + timestamp score + matches + path length
					var score float32 = 0.0
					if _, ok := postMap[p.ID]; !ok {
						// first time we see this post
						// add site score and timestamp score
						// - site score from db
						siteScore := p.SiteScore
						if siteScore > 300 {
							siteScore = 300
						}
						score += p.SiteScore * float32(math.Pow(SITE_MULTIPLIER, 2.0))
						// - timestamp score = 1.0 / (1.0 + (timestamp - now) / (24 * 60 * 60))
						if p.Timestamp == 0 {
							p.Timestamp = uint64(time.Now().Unix() - 604800) // 1 week ago
						}

						tsScore := float32(1.0 / float64(1.0+float64(int64(p.Timestamp-uint64(time.Now().Unix())))/float64(24*60*60)))
						if tsScore < 0 {
							tsScore = 0.001
						}
						if tsScore > TIME_MULTIPLIER {
							tsScore = TIME_MULTIPLIER
						}
						score += tsScore * TIME_MULTIPLIER

						// language score
						if *lang == p.Language {
							score *= LANGUAGE_MULTIPLIER
						}

						// url based scores

						{

							if u, err := url.Parse(p.URL); err == nil {
								// successfully parsed url
								// protocol score (https)
								switch u.Scheme {
								case "https":
									score *= PROTOCOL_MULTIPLIER
								default:
									// none
								}
								paths := len(strings.Split(u.Path, "/"))
								queries := len(strings.Split(u.RawQuery, "="))
								// paths size greater = less likely homepage = less score
								pathScore := 1 / float32(paths+(queries*PATH_QUERY_MULTIPLIER)) * PATH_MULTIPLIER
								score += pathScore
							}
						}
						// set base score
						postMap[p.ID] = &post{p, 0, score, make(map[string]float32)}
					}
				}

				// token score
				{
					var score float32 = 0.0
					// human tokens
					{

						if h, ok := p.TokensH[t.Token]; ok {
							score += h
						}
					}
					// site tokens
					{

						if h, ok := p.SiteTokens[t.Token]; ok {
							score += h
						}
					}
					// ai tokens
					{
						score += p.Tokens[t.Token]
					}
					postMap[p.ID].tokenScores[t.Token] = score
				}

				/*
					if v, ok := postMatches[p.ID]; ok {
						postMatches[p.ID] = v + 1
					} else {
						postMatches[p.ID] = 1
					}
					score *= float32(postMatches[p.ID]) * MATCH_MULTIPLIER

					postMap[p.ID] = &post{p, score}
				*/
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	for _, p := range postMap {
		// post final score
		{
			p.score = p.baseScore
			// final score = (average of all token scores) + (2nd highest token score) * TOKEN_MULTIPLIER
			if len(p.tokenScores) > 1 {
				scores := make([]float32, 0, len(p.tokenScores))
				var sum float32 = 0.0
				for _, v := range p.tokenScores {
					scores = append(scores, v)
					sum += v
				}
				// sort tokenscores in ascending order (small to big a > b)
				sort.Slice(scores, func(i, j int) bool {
					return scores[i] < scores[j]
				})

				score1 := scores[len(scores)-2]

				var diffScore float32 = 0

				{
					avg := sum / float32(len(scores))
					var avgs float32 = 0
					for _, s := range scores {
						avgs += (s / avg)
					}
					diffScore = avgs / float32(len(scores))
				}

				p.score += (sum + (score1 * MATCH_MULTIPLIER * float32(len(scores)))) * float32(diffScore)
			} else {
				// since only one token, just use that token's score
				for _, v := range p.tokenScores {
					p.score += v
				}
			}
		}

		// finally push the the posts list
		// log.Println(p.ID, p.Title, p.baseScore, p.tokenScores, p.score) // debug (print score)
		*allPosts = append(*allPosts, *p)
	}
	// - rank by highest score
	sort.Slice(*allPosts, func(i, j int) bool {
		return (*allPosts)[i].score > (*allPosts)[j].score
	})

	// seach num offset
	if uint32(len(*allPosts)) > num+offset {
		*allPosts = (*allPosts)[offset : offset+num]
	}

	//log.Println("top posts")
	for _, p := range *allPosts {
		// log.Println(p.ID, p.Title, p.baseScore, p.tokenScores, p.score) // debug (print score)
		*t = append(*t, p.Post)
	}

	return nil
}

func (ind *Indexer) QueryToken(token *string, t *[]fullpost) error {
	// TODO
	// - get all postID back from token
	postIds := new([]uint64)
	if err := ind.Store.getPostIDListFromToken(token, postIds); err != nil {
		return err
	}

	// - get all post back from postID
	// t0 := time.Now()
	if err := ind.Store.qs.GetData(postIds, t); err != nil {
		return err
	}
	// log.Printf("Query %s [%d rows] from db took %v\n", *token, len(*t), time.Since(t0))
	return nil
}
