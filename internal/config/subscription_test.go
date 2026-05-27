package config

import "testing"

func TestSubscriptionURL(t *testing.T) {
	t.Setenv("SUB_PUBLIC_URL", "https://panel.example.com")
	if got := SubscriptionURL("abc"); got != "https://panel.example.com/sub/abc" {
		t.Fatalf("SubscriptionURL() = %q", got)
	}
}

func TestSubscriptionPublicBaseFromHTTPAddr(t *testing.T) {
	t.Setenv("SUB_PUBLIC_URL", "")
	t.Setenv("HTTP_ADDR", "0.0.0.0:8787")
	if got := SubscriptionPublicBase(); got != "http://127.0.0.1:8787" {
		t.Fatalf("SubscriptionPublicBase() = %q", got)
	}
}

func TestUsingLocalSubscriptionURL(t *testing.T) {
	t.Setenv("SUB_PUBLIC_URL", "")
	if !UsingLocalSubscriptionURL() {
		t.Fatal("expected local URL mode")
	}
	t.Setenv("SUB_PUBLIC_URL", "http://10.0.0.1:8080")
	if UsingLocalSubscriptionURL() {
		t.Fatal("expected public URL mode")
	}
}
