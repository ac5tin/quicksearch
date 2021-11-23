package tokeniser

type token struct {
	AIScore float32 `json:"ai_score"`
	DocFreq float32 `json:"doc_freq"`
	Word    string  `json:"word"`
}

func (t *Tokeniser) Score(input *[]string, out *[]token) error {
	return nil
}
