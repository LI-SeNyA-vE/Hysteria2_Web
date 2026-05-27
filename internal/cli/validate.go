package cli

import (
	"fmt"
	"strings"
	"unicode"
)

const usernamePrefix = "sub_"

func withUsernamePrefix(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return name
	}
	if strings.HasPrefix(name, usernamePrefix) {
		return name
	}
	return usernamePrefix + name
}

func normalizePanelUsername(name string) (string, error) {
	name = withUsernamePrefix(name)
	if err := validateUsername(name); err != nil {
		return "", err
	}
	return name, nil
}

func validateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username не может быть пустым")
	}

	for i, r := range username {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			continue
		default:
			return fmt.Errorf(
				"недопустимый символ %q (позиция %d) в username %q%s",
				r, i+1, username, invalidRuneHint(r),
			)
		}
	}
	return nil
}

func invalidRuneHint(r rune) string {
	if unicode.Is(unicode.Cyrillic, r) {
		switch r {
		case 'е', 'Е':
			return " — кириллическая «е», нужна латинская (переключите раскладку EN)"
		default:
			return " — кириллическая буква, нужна латиница (раскладка EN)"
		}
	}
	return ". Допустимы только a-z, A-Z, 0-9 и _"
}
