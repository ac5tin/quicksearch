package indexer

// Sync local redis store with upstream database
func (s *Store) Sync() error {
	posts := new([]Post)
	if err := s.getAllPosts(posts); err != nil {
		return err
	}

	for _, p := range *posts {
		for k := range p.Tokens {
			if err := s.addPostLink(&k, &p.ID); err != nil {
				return err
			}
		}
	}
	return nil
}
