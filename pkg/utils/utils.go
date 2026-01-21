package utils

import (
	"crypto/rand"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateID(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b) // generates cryptographically secure random bytes
	if err != nil {
		return "", err
	}

	for i := 0; i < length; i++ {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b), nil
}
