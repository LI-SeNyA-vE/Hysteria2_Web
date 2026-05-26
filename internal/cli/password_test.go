package cli

import (
	"strings"
	"testing"
	"unicode"
)

func TestGeneratePassword(t *testing.T) {
	t.Parallel()

	password, err := generatePassword(33)
	if err != nil {
		t.Fatalf("generatePassword() error = %v", err)
	}
	if len(password) != 33 {
		t.Fatalf("len = %d, want 33", len(password))
	}

	var hasLower, hasUpper, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			t.Fatalf("unexpected character %q in password", r)
		}
		if strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:'\",.<>?/`~", r) {
			t.Fatalf("special character found: %q", r)
		}
	}
	if !hasLower || !hasUpper || !hasDigit {
		t.Fatalf("missing required char classes: lower=%v upper=%v digit=%v", hasLower, hasUpper, hasDigit)
	}
}
