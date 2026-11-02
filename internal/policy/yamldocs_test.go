// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package policy

import "testing"

func TestSplitYAMLDocuments(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"single no marker", "a: 1\nb: 2\n", 1},
		{"leading marker", "---\na: 1\n", 1},
		{"two docs", "a: 1\n---\nb: 2\n", 2},
		{"leading + two docs", "---\na: 1\n---\nb: 2\n", 2},
		{"three docs CRLF", "a: 1\r\n---\r\nb: 2\r\n---\r\nc: 3\r\n", 3},
		{"trailing marker ignored", "a: 1\n---\n", 1},
		{"comment-only doc ignored", "# header\n---\na: 1\n---\n# nothing\n", 1},
		{"empty", "", 0},
		{"marker with trailing space", "a: 1\n--- \nb: 2\n", 2},
		{"indented dashes are content", "a: |\n  ---\n  text\n", 1},
		{"end marker", "a: 1\n...\nb: 2\n", 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := len(splitYAMLDocuments([]byte(tc.in))); got != tc.want {
				t.Fatalf("got %d docs, want %d", got, tc.want)
			}
		})
	}
}

func TestSplitYAMLDocumentsPreservesContent(t *testing.T) {
	docs := splitYAMLDocuments([]byte("id: one\nname: A\n---\nid: two\nname: B\n"))
	if len(docs) != 2 || string(docs[0]) != "id: one\nname: A\n" || string(docs[1]) != "id: two\nname: B\n" {
		t.Fatalf("unexpected split: %q", docs)
	}
}
