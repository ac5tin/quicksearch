package textprocessor

import (
	"testing"

	"github.com/joho/godotenv"
)

func TestEntity(t *testing.T) {
	godotenv.Load("../.env")
	tp := new(TextProcessor)

	ents := new([]string)
	if err := tp.EntityRecognition(InputText{Text: "With longer battery life, upgraded storage, camera updates and the new A15 Bionic processor, Apple's iPhone 13 is a tempting choice when picking a new iPhone", Lang: "en"}, ents); err != nil {
		t.Error(err)
	}
	t.Log(*ents)
	if len(*ents) == 0 {
		t.Error("Entity recognition failed")
	}
}
