package blitz

import "testing"

func TestCollectHy2URIs(t *testing.T) {
	t.Parallel()

	ipv4 := "hy2://v4-link"
	ipv6 := "http://not-hy2"
	normal := "hy2://sub-link"
	resp := UserURIResponse{
		IPv4:      &ipv4,
		IPv6:      &ipv6,
		NormalSub: &normal,
		Nodes: []NodeURI{
			{Name: "node1", URI: "hy2://node-link"},
			{Name: "dup", URI: "hy2://node-link"},
		},
	}

	got := CollectHy2URIs(resp)
	if len(got) != 3 {
		t.Fatalf("CollectHy2URIs() len = %d, want 3: %v", len(got), got)
	}
}
