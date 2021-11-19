package processor

type Results struct {
	RawHTML              string   `json:"rawHTML"`
	URL                  string   `json:"url"`
	Title                string   `json:"title"`
	Summary              string   `json:"summary"`
	Author               string   `json:"author"`
	MainContent          string   `json:"mainContent"`
	Timestamp            uint64   `json:"timestamp"`
	Site                 string   `json:"site"`
	Country              string   `json:"country"`
	Lang                 string   `json:"lang"`
	Type                 string   `json:"type"`
	RelatedInternalLinks []string `json:"relatedInternalLinks"`
	RelatedExternalLinks []string `json:"relatedExternalLinks"`
	Tokens               []string `json:"tokens"`
}
