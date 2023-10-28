package utils

func Substring(str string, start, end int) string {
	if len(str) < end {
		return str
	}

	runes := []rune(str)

	return string(runes[start:end])
}
