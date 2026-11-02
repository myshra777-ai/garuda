// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package store

import "testing"

func TestParseModuleDirective(t *testing.T) {
	cases := []struct {
		line string
		want string
		ok   bool
	}{
		{"module github.com/x/y", "github.com/x/y", true},
		{"module github.com/x/y // v1.9.8", "github.com/x/y", true},
		{"module   github.com/x/y   //c", "github.com/x/y", true},
		{"module \"github.com/x/y\"", "github.com/x/y", true},
		{"module\tgithub.com/x/y", "github.com/x/y", true},
		{"modulefoo", "", false},
		{"module", "", false},
		{"module // only comment", "", false},
		{"go 1.26", "", false},
		{"// module a/b", "", false},
	}
	for _, tc := range cases {
		got, ok := parseModuleDirective(tc.line)
		if got != tc.want || ok != tc.ok {
			t.Errorf("parseModuleDirective(%q) = (%q,%v), want (%q,%v)", tc.line, got, ok, tc.want, tc.ok)
		}
	}
}
