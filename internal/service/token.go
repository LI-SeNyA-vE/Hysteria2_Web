package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func generateSubToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate sub token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
