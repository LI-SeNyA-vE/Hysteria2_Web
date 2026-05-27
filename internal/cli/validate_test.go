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
	if err := validateUsername("sub_alice"); err != nil {
		t.Fatalf("valid username rejected: %v", err)
	}

	// Cyrillic "е" + Latin — common RU keyboard mistake.
	if err := validateUsername("еtest"); err == nil {
		t.Fatal("expected error for cyrillic username")
	}
}

func TestWithUsernamePrefix(t *testing.T) {
	t.Parallel()

	if got := withUsernamePrefix("alice"); got != "sub_alice" {
		t.Fatalf("withUsernamePrefix(alice) = %q, want sub_alice", got)
	}
	if got := withUsernamePrefix("sub_bob"); got != "sub_bob" {
		t.Fatalf("withUsernamePrefix(sub_bob) = %q, want sub_bob", got)
	}
}
