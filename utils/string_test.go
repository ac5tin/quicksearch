package utils

import "testing"

func TestTruncateString(t *testing.T) {
	input := "Hello world!"
	var max uint16 = 5
	TruncateString(&input, &max)
	t.Logf("Truncated string into: %s", input)
	if len(input) != 5 {
		t.Error("Truncate string length failed")
	}
	if input != "Hello" {
		t.Error("Unexpected truncate string result")
	}

	input = "Hello world!"
	max = 20
	TruncateString(&input, &max)
	if input != "Hello world!" {
		t.Error("Truncated string who's length is already less than max")

	}
}
