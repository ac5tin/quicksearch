package textprocessor

import (
	"testing"

	"github.com/joho/godotenv"
)

func TestLangDetect(t *testing.T) {
	godotenv.Load("../.env")
	tp := new(TextProcessor)
	lang := new(string)
	if err := tp.LangDetect("I love to eat apple", lang); err != nil {
		t.Error(err)
	}

	if *lang != "en" {
		t.Errorf("Expected: en, got: %s", *lang)
	}

}
