package cli

import "testing"

func TestValidateUsername(t *testing.T) {
	t.Parallel()

	if err := validateUsername("testuserpanel1"); err != nil {
		t.Fatalf("valid username rejected: %v", err)
	}
	if err := validateUsername("user_123"); err != nil {
		t.Fatalf("valid username rejected: %v", err)
	}

	// Cyrillic "е" + Latin — common RU keyboard mistake.
	if err := validateUsername("еtest"); err == nil {
		t.Fatal("expected error for cyrillic username")
	}
}
