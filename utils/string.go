package utils

func TruncateString(input *string, max *uint32) {
	if uint32(len(*input)) > *max {
		*input = (*input)[0:*max]
	}
}
