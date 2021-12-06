package indexer

import "testing"

func TestQuickstore(t *testing.T) {
	slaves := []string{}
	mkey := "key"
	qs := NewQuickStore(true, &slaves, &mkey)

	key := "hello"
	value := "world"
	b := []byte(value)
	qs.set(&key, &b)

	r := new([]byte)
	if err := qs.get(&key, r); err != nil {
		t.Error("Error getting value")
	}
	if string(*r) != value {
		t.Errorf("Expected %s, got %s", value, *r)
	}

}
