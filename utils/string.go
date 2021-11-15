package utils

func TruncateString(input *string, max *uint16) {
	if uint16(len(*input)) > *max {
		*input = (*input)[0:*max]
	}
}
