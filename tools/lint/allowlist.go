// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

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
		Hash:   "e39622de75bf2ec884b1b9e5f04e22df4db8ea2050208959a484a12cfe2d58ca",
		Reason: "Correct by design. Migration 084 added a `kind` discriminator and split the table's shape: decision_vs_decision rows carry decision_a and decision_b, runtime_vs_static rows carry workspace_id and source_entity_id. This INSERT writes a decision_vs_decision row; per the schema's chk_contradictions_kind_decision constraint it must have decision_a and decision_b, and MUST NOT require workspace_id. decisions has no workspace_id column, so the caller has no workspace context to pass. The lint check cannot see kind-discrimination logic. See migrations/084_contradictions_runtime_kind.sql.",
	},
	{
		File:   "../../internal/api/control_plane_tenants.go",
		Check:  "sql-scoping",
		Hash:   "c06e17bf704b0045ef0178ba86e849e28da7fca6210e1d10b1cab705ec4e2c2d",
		Reason: "The Control Plane Tenants list intentionally aggregates across every workspace of a tenant. The entities and mcp_sessions counts are per-tenant, not per-workspace. Adding a workspace_id filter would break the Tenants tab metric. See docs/specs/control-plane-metrics.md §Tenants tab.",
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
	{
		File:   "../../cmd/garuda/stats.go",
		Check:  "sql-scoping",
		Hash:   "ebeda374f15f79fe1aa96057dd4db1897fae171d0a08e6631ec32b327e5099e9",
		Reason: "Same as stats.go:67 — garuda stats reports tenant-wide aggregate decision count by design.",
	},
	{
		File:   "../../internal/store/analysis.go",
		Check:  "sql-scoping",
		Hash:   "6012cca43a1cce55cce3fb451861aa47b4ec0d4f82c76c9e00e625ea5971f130",
		Reason: "SaveAnalysisResult persists a tenant-level analysis decision record using scope_system for the workspace name; Migration 098 added nullable workspace_id for dashboard scoping while keeping legacy DecisionStore callers backward-compatible.",
	},
	{
		File:   "../../internal/store/decision_store.go",
		Check:  "sql-scoping",
		Hash:   "40152e48058d25b6f8edea0ec041e6ccdd49e4b1667f08145d4181ac3961a35c",
		Reason: "SaveDecision writes a canonical types.Decision (whose RFC 6962 canonical byte encoding in internal/merkle/decision.go commits to scope_domain and scope_system rather than workspace_id).",
	},
	{
		File:   "../../internal/store/decision_store.go",
		Check:  "sql-scoping",
		Hash:   "6fd44960698ce493dec283177b704358cb6ae81e670e20b1ab929b032240eed2",
		Reason: "ListDecisions(ctx, tenantID, filter) is a tenant-level store method that filters by scope_domain / scope_system; dashboard handlers scope by workspace_id directly.",
	},
	{
		File:   "../../internal/store/decision_store.go",
		Check:  "sql-scoping",
		Hash:   "a5fed9f2e04b1762201871a5c5a03ddd47aa1dd0e3a7168a96a9cff07da82f87",
		Reason: "Same as decision_store.go:181 — GetDecisionsByScope filters by tenant_id and scope_domain/scope_system.",
	},
	{
		File:   "../../internal/store/postgres.go",
		Check:  "sql-scoping",
		Hash:   "7c5346829f747468e6d154c79677c6406ae83019376650c462266ad33d5877ca",
		Reason: "GetDecisions(ctx, tenantID) is a legacy tenant-level store interface method; workspace-scoped callers use dashboard_handlers.go with workspace_id.",
	},
	{
		File:   "../../internal/store/revision_store.go",
		Check:  "sql-scoping",
		Hash:   "58bb5f5a01a1b100f77ba70b04160c3869eb711ee2e9541c74edc15da0c02d93",
		Reason: "CommitRevision inserts a revision-linked decision row keyed by tenant_id and scope_system.",
	},
	{
		File:   "../../internal/store/temporal_store.go",
		Check:  "sql-scoping",
		Hash:   "689cc7130925a7caafda5bbea02c45de7359a8c7c5fb60c65f1d961bda068099",
		Reason: "GetDecisionsAtTime queries historical tenant decisions across a temporal window.",
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
