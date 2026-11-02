// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package policy

import "strings"

// splitYAMLDocuments splits a YAML stream on document separators ("---" at column 0)
// and drops documents that contain no content (blank or comment-only). The "..."
// end marker is treated as a separator too. Each returned slice keeps its original
// text so it can be hashed and unmarshalled independently.
func splitYAMLDocuments(data []byte) [][]byte {
	var docs [][]byte
	var cur []string
	flush := func() {
		if hasYAMLContent(cur) {
			for len(cur) > 0 && strings.TrimSpace(cur[len(cur)-1]) == "" {
				cur = cur[:len(cur)-1]
			}
			docs = append(docs, []byte(strings.Join(cur, "\n")+"\n"))
		}
		cur = cur[:0]
	}
	for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		trimmed := strings.TrimRight(line, " \t")
		if trimmed == "---" || strings.HasPrefix(trimmed, "--- ") || trimmed == "..." {
			flush()
			// "--- # comment" or "--- key: v" openers: keep inline content after the marker.
			if strings.HasPrefix(trimmed, "--- ") {
				rest := strings.TrimSpace(trimmed[4:])
				if rest != "" && !strings.HasPrefix(rest, "#") {
					cur = append(cur, rest)
				}
			}
			continue
		}
		cur = append(cur, line)
	}
	flush()
	return docs
}

func hasYAMLContent(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if t != "" && !strings.HasPrefix(t, "#") {
			return true
		}
	}
	return false
}
