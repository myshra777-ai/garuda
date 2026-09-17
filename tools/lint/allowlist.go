// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package lint

// AllowEntry suppresses one violation. Entries are keyed on
// (File, Check, Hash), not line numbers: line numbers drift when a
// file is edited; literal content does not. Every entry must carry a
// non-empty Reason; a test in lint_test.go enforces this.
//
// File paths are the strings produced by the walker, which starts
// from this package's directory. That is why they carry the "../../"
// prefix. A future refinement will normalize them to repo-relative.
//
// An empty allowlist is the target state. Entries accumulate only when
// a violation is a genuine false positive that the check cannot be
// refined to exclude, or when the violation is a known defect tracked
// elsewhere (in which case the Reason cites the tracker item and the
// entry is removed when the defect is fixed).
type AllowEntry struct {
	File   string
	Check  string
	Hash   string
	Reason string
}

// Allowlist is the current set of suppressed violations.
var Allowlist = []AllowEntry{
	{
		File:   "../../cmd/garuda/stats.go",
		Check:  "sql-scoping",
		Hash:   "746a8045d56395c296303c22d685dbbb9e206db0b53ea2147861fb8b2eb03ce2",
		Reason: "garuda stats reports tenant-wide aggregate counts by design. The command has no workspace dimension; adding one changes the product surface, not the code correctness.",
	},
	{
		File:   "../../cmd/garuda/stats.go",
		Check:  "sql-scoping",
		Hash:   "f9be995839b44e94b55c08b4e16059aba0d11f00873488845a219b6a5b903787",
		Reason: "Same as stats.go:67 — claims count is a tenant-wide aggregate.",
	},
	{
		File:   "../../cmd/garuda/stats.go",
		Check:  "sql-scoping",
		Hash:   "cc59489ac3156c0abfc6a2302953b9325ce119e0e6d1cbd36e0b1b57aa1d92d3",
		Reason: "Same as stats.go:67 — cross_repo_edges count is a tenant-wide aggregate.",
	},
	{
		File:   "../../internal/store/contradiction_store.go",
		Check:  "sql-scoping",
		Hash:   "6a4e812132fbfb17ac7af9f71be0a47fa5ee614ecd71c3cbb6e50fb47ea485a4",
		Reason: "KNOWN DEFECT, tracked as B28. INSERT into contradictions does not set workspace_id; rows are written with NULL. The check fires correctly. Removing this entry is part of the B28 fix, which will derive workspace from decision_a. Until then, the only consumer is the tenant-wide contradiction report, which does not filter by workspace, so the omission is not user-visible.",
	},
	{
		File:   "../../internal/store/contradiction_store.go",
		Check:  "sql-scoping",
		Hash:   "3d66e168315d582f1c45d86e521ca47c79cae3711ed753955212a49422ec92b0",
		Reason: "Deliberate tenant-wide contradiction report driven by the garuda.detect_contradictions MCP tool. contradictions.workspace_id is nullable and historical rows carry NULL; workspace-scoping this query would hide them.",
	},
	{
		File:   "../../internal/store/contradiction_store.go",
		Check:  "sql-scoping",
		Hash:   "d2f82751400211687120b8acefcf07b7f3d9acc75ca330dcedad26207d4b1e43",
		Reason: "Same as contradiction_store.go:72 — second tenant-wide contradiction query on the same table.",
	},
}

// IsAllowed reports whether (file, check, hash) is present in the
// allowlist.
func IsAllowed(file, check, hash string) bool {
	for _, e := range Allowlist {
		if e.File == file && e.Check == check && e.Hash == hash {
			return true
		}
	}
	return false
}
