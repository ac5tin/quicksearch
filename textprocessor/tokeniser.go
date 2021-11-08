package textprocessor

type Token struct {
	Text  string  `json:"text"`
	Score float32 `json:"score"`
}

func (tp *TextProcessor) Tokenise(text string, tokens *[]Token) error {
	return nil
}
