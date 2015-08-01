package utils

import (
	"crypto/rand"
	"math/big"
)

const randomStringCharset = "abcdefghijklmnopqrstuvxyzABCDEFGHIJKLMNOPQRSTUVXYZ0123456789"

var randomStringCharsetLength = big.NewInt(int64(len(randomStringCharset)))

func RandomString(length int) string {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		value, err := rand.Int(rand.Reader, randomStringCharsetLength)
		if err != nil {
			panic(err)
		}
		result[i] = randomStringCharset[value.Int64()]
	}
	return string(result)
}

func StripQuotes(str string) string {
	if len(str) > 1 && str[0] == '"' && str[len(str)-1] == '"' {
		return str[1 : len(str)-1]
	}
	return str
}
