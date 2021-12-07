package indexer

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// Sync local redis store with upstream database
func (s *Store) Sync() error {
	posts := new([]fullpost)
	if err := s.getAllPosts(posts); err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// dont block
		}

		for _, p := range *posts {
			tokens := make(map[string]interface{})
			for k := range p.Tokens {
				tokens[k] = struct{}{}
			}
			for k := range p.TokensH {
				tokens[k] = struct{}{}
			}
			for k := range p.SiteTokens {
				tokens[k] = struct{}{}
			}
			for token := range tokens {
				if err := s.addPostLink(&token, &p.ID); err != nil {
					return err
				}
			}
		}
		return nil
	})
	g.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// dont block
		}
		if err := s.qs.SetData(posts); err != nil {
			return err
		}

		return nil
	})
	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
