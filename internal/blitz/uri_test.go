package blitz

import "testing"

func TestRelabelHy2Remark(t *testing.T) {
	t.Parallel()

	got := RelabelHy2Remark("hy2://user:pass@1.2.3.4:443/?insecure=1#IPv4", "Test")
	want := "hy2://user:pass@1.2.3.4:443/?insecure=1#Test"
	if got != want {
		t.Fatalf("RelabelHy2Remark() = %q, want %q", got, want)
	}
}

func TestCollectRelabeledHy2URIsKeepsMultipleServers(t *testing.T) {
	t.Parallel()

	ipv4 := "hy2://user@1.2.3.4:443/?insecure=1#IPv4"
	ipv6 := "hy2://user@1.2.3.4:443/?insecure=1#IPv6"
	resp := UserURIResponse{
		IPv4: &ipv4,
		IPv6: &ipv6,
	}

	got := CollectRelabeledHy2URIs(resp, "Test")
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2: %v", len(got), got)
	}
}

func TestCollectRelabeledHy2URIsDifferentServerNames(t *testing.T) {
	t.Parallel()

	uri := "hy2://user@185.1.2.3:443/?insecure=1#IPv4"
	resp := UserURIResponse{IPv4: &uri}

	a := CollectRelabeledHy2URIs(resp, "Test")
	b := CollectRelabeledHy2URIs(resp, "VK")
	if len(a) != 1 || len(b) != 1 {
		t.Fatalf("unexpected counts: a=%d b=%d", len(a), len(b))
	}
	if a[0] == b[0] {
		t.Fatalf("expected different remarks, got %q", a[0])
	}
}
