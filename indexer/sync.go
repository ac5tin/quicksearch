package indexer

// Sync local redis store with upstream database
func (s *Store) Sync() error {
	posts := new([]fullpost)
	if err := s.getAllPosts(posts); err != nil {
		return err
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
}
