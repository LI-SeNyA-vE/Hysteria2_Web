package blitz

import (
	"fmt"
	"strings"
)

// RelabelHy2Remark replaces the #fragment (display name) in a hy2:// URI.
func RelabelHy2Remark(uri, remark string) string {
	uri = strings.TrimSpace(uri)
	remark = strings.TrimSpace(remark)
	if !strings.HasPrefix(uri, "hy2://") || remark == "" {
		return uri
	}
	base, _, _ := strings.Cut(uri, "#")
	return base + "#" + remark
}

type hy2Entry struct {
	uri   string
	label string
}

func hy2Entries(resp UserURIResponse) []hy2Entry {
	var entries []hy2Entry
	seen := make(map[string]struct{})

	add := func(raw *string, label string) {
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
		entries = append(entries, hy2Entry{uri: line, label: label})
	}

	add(resp.IPv4, "IPv4")
	add(resp.IPv6, "IPv6")
	add(resp.NormalSub, "Sub")
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
		label := strings.TrimSpace(node.Name)
		if label == "" {
			label = "node"
		}
		entries = append(entries, hy2Entry{uri: line, label: label})
	}

	return entries
}

// CollectRelabeledHy2URIs returns hy2 links with #fragment set to the server name.
// Multiple links from the same server stay distinct (e.g. IPv4 + IPv6).
func CollectRelabeledHy2URIs(resp UserURIResponse, serverName string) []string {
	serverName = strings.TrimSpace(serverName)
	entries := hy2Entries(resp)
	if len(entries) == 0 || serverName == "" {
		return nil
	}

	remarks := make([]string, len(entries))
	for i := range entries {
		remarks[i] = serverName
	}

	relabeled := func(i int) string {
		return RelabelHy2Remark(entries[i].uri, remarks[i])
	}

	seen := make(map[string]int)
	for i := range entries {
		line := relabeled(i)
		if count, ok := seen[line]; ok {
			switch entries[i].label {
			case "IPv4", "IPv6", "Sub":
				remarks[i] = fmt.Sprintf("%s · %d", serverName, count+1)
			default:
				remarks[i] = serverName + " · " + entries[i].label
			}
			line = relabeled(i)
		}
		seen[line]++
	}

	out := make([]string, 0, len(entries))
	outSeen := make(map[string]struct{}, len(entries))
	for i := range entries {
		line := relabeled(i)
		if _, ok := outSeen[line]; ok {
			continue
		}
		outSeen[line] = struct{}{}
		out = append(out, line)
	}
	return out
}
