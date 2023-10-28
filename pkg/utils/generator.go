package utils

import (
	mr "math/rand"
)

var (
	nums    = "1234567890"
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func GeneratePincodeString(max int) string {
	return generate(nums, max)
}

func GenerateTokenString(max int) string {
	return generate(letters, max)
}

func generate(source string, max int) string {
	b := make([]byte, max)
	for i := range b {
		b[i] = source[mr.Intn(len(source))]
	}
	return string(b)
}
