package blitz

import "strings"

func CollectHy2URIs(resp UserURIResponse) []string {
	var uris []string
	seen := make(map[string]struct{})

	add := func(raw *string) {
		if raw == nil {
			return
		}
		line := strings.TrimSpace(*raw)
		if !strings.HasPrefix(line, "hy2://") {
			return
		}
		if _, ok := seen[line]; ok {
			return
		}
		seen[line] = struct{}{}
		uris = append(uris, line)
	}

	add(resp.IPv4)
	add(resp.IPv6)
	add(resp.NormalSub)
	for _, node := range resp.Nodes {
		if node.URI == "" {
			continue
		}
		line := strings.TrimSpace(node.URI)
		if !strings.HasPrefix(line, "hy2://") {
			continue
		}
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		uris = append(uris, line)
	}

	return uris
}
