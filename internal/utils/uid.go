package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateID(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	// base64 URL encoding makes it URL-safe and readable
	return base64.RawURLEncoding.EncodeToString(b)
}
