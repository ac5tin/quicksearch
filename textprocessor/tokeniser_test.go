package textprocessor

import (
	"testing"

	"github.com/joho/godotenv"
)

func TestTokeniser(t *testing.T) {
	godotenv.Load("../.env")
	tp := new(TextProcessor)

	tokens := new([]Token)
	if err := tp.Tokenise(InputText{Text: "I love to eat apple", Lang: "en"}, tokens); err != nil {
		t.Error(err)
	}

	t.Log(*tokens)
	if len(*tokens) == 0 {
		t.Errorf("Failed to tokenise sentence")
	}
}
