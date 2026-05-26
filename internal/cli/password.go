package cli

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const passwordAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generatePassword(length int) (string, error) {
	if length < 3 {
		return "", fmt.Errorf("password length must be at least 3")
	}

	lowers := "abcdefghijklmnopqrstuvwxyz"
	uppers := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits := "0123456789"

	buf := make([]byte, length)
	required := []string{lowers, uppers, digits}
	for i, charset := range required {
		c, err := randomChar(charset)
		if err != nil {
			return "", err
		}
		buf[i] = c
	}
	for i := len(required); i < length; i++ {
		c, err := randomChar(passwordAlphabet)
		if err != nil {
			return "", err
		}
		buf[i] = c
	}
	shuffleBytes(buf)
	return string(buf), nil
}

func randomChar(charset string) (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
	if err != nil {
		return 0, err
	}
	return charset[n.Int64()], nil
}

func shuffleBytes(b []byte) {
	for i := len(b) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			continue
		}
		idx := int(j.Int64())
		b[i], b[idx] = b[idx], b[i]
	}
}
