// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package runtime

import "strings"

// ContradictionMarkers are the substrings in a runtime target that
// indicate an observation contradicts the static architecture or a
// policy. Both the ingest handler (internal/api/runtime_handlers.go)
// and the verifier (internal/runtime/verifier.go) match against this
// list.
//
// Every marker here is load-bearing somewhere in the codebase:
//
//	unapproved   — present in every check. Universal.
//	exfiltration — benchmark corpus. runner.go:82.
//	driver       — README example: "payment services must not access
//	               the database driver directly".
//	bypass       — benchmark corpus. runner.go:71.
//
// Adding a marker here causes both writer paths to recognise it. The
// SQL form in ContradictionMarkerClause uses only the marker text, so
// markers must not contain SQL metacharacters or LIKE wildcards.
var ContradictionMarkers = []string{"unapproved", "exfiltration", "driver", "bypass"}

// ContainsContradictionMarker reports whether target contains any
// marker. Used by the ingest handler's Go-side check.
func ContainsContradictionMarker(target string) bool {
	for _, m := range ContradictionMarkers {
		if strings.Contains(target, m) {
			return true
		}
	}
	return false
}

// ContradictionMarkerClause returns a SQL fragment matching any of
// the markers against the named column, for use in a WHERE clause.
// Example: ContradictionMarkerClause("re.raw_target") returns
// "(re.raw_target ILIKE '%unapproved%' OR re.raw_target ILIKE
// '%exfiltration%' OR ...)".
//
// The column name is caller-supplied and must be a trusted literal;
// it is not parameterised because Postgres does not permit a
// parameter in the column position of an ILIKE expression. The
// marker strings are compile-time constants with no metacharacters.
func ContradictionMarkerClause(column string) string {
	parts := make([]string, len(ContradictionMarkers))
	for i, m := range ContradictionMarkers {
		parts[i] = column + " ILIKE '%" + m + "%'"
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}
