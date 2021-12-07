package indexer

import (
	"testing"
	"time"
)

func TestQuickstore(t *testing.T) {
	slaves := []string{}
	mkey := "key"
	qs := newQuickStore(true, &slaves, &mkey)

	key := "hello"
	value := "helloworld"
	b := []byte(value)
	qs.set(&key, &b)

	r := new([]byte)
	if err := qs.get(&key, r); err != nil {
		t.Error("Error getting value")
	}
	if string(*r) != value {
		t.Errorf("Expected %s, got %s", value, *r)
	}

	post := Post{
		ID:            1,
		Author:        "test author",
		Title:         "test",
		Tokens:        map[string]float32{"test": 1.2},
		Summary:       "testing test 123 abc",
		URL:           "https://example.com",
		Timestamp:     uint64(time.Now().Unix()),
		Language:      "en",
		InternalLinks: []string{"https://example.com/abc"},
		ExternalLinks: []string{"https://abc.com"},
		Entities:      map[string]float32{"testing": 2.1, "test": 11.3},
	}
	d := make([]fullpost, 0)
	d = append(d, fullpost{Post: post})
	qs.SetData(&d)
	idList := []uint64{1}
	ret := new([]fullpost)
	if err := qs.GetData(&idList, ret); err != nil {
		t.Error("Error getting data")
	}
	if len(*ret) == 0 {
		t.Error("Failed to get data")
	}

}
