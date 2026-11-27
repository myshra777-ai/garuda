// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"strings"
	"testing"
)

func TestContainsContradictionMarker(t *testing.T) {
	cases := map[string]bool{
		"payments -> unapproved-ledger-write": true,
		"data exfiltration attempt":           true,
		"handler opened database driver":      true,
		"auth bypass via debug flag":          true,
		"payments -> repository.Save":         false,
		"":                                    false,
		// Matching is case-sensitive on the Go side; the SQL side uses ILIKE.
		"UNAPPROVED write": false,
	}
	for target, want := range cases {
		if got := ContainsContradictionMarker(target); got != want {
			t.Errorf("ContainsContradictionMarker(%q) = %v, want %v", target, got, want)
		}
	}
}

func TestContradictionMarkerClause(t *testing.T) {
	got := ContradictionMarkerClause("re.raw_target")
	want := "(re.raw_target ILIKE '%unapproved%' OR re.raw_target ILIKE '%exfiltration%' OR " +
		"re.raw_target ILIKE '%driver%' OR re.raw_target ILIKE '%bypass%')"
	if got != want {
		t.Fatalf("clause =\n%s\nwant\n%s", got, want)
	}
}

func TestContradictionMarkers_AreSafeForSQL(t *testing.T) {
	// ContradictionMarkerClause inlines markers into SQL, so a marker must
	// never contain quotes, LIKE wildcards or other metacharacters.
	for _, m := range ContradictionMarkers {
		if m == "" {
			t.Fatal("empty marker would match every row")
		}
		if strings.ContainsAny(m, `'"%_\;-`) {
			t.Fatalf("marker %q contains a SQL metacharacter", m)
		}
		if m != strings.ToLower(m) {
			t.Fatalf("marker %q must be lower case to match the Go-side check", m)
		}
	}
}

func TestContradictionMarkers_GoAndSQLAgree(t *testing.T) {
	// Every marker the SQL clause matches must also be caught by the Go check.
	clause := ContradictionMarkerClause("c")
	for _, m := range ContradictionMarkers {
		if !strings.Contains(clause, "'%"+m+"%'") {
			t.Fatalf("marker %q missing from the SQL clause", m)
		}
		if !ContainsContradictionMarker("prefix-" + m + "-suffix") {
			t.Fatalf("marker %q not caught by the Go check", m)
		}
	}
}
