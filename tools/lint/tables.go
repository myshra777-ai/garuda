// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package lint

// WorkspaceScopedTables is the set of tables that carry a workspace_id
// column. A query that reads or writes one of these tables must scope
// by workspace_id, repository_id, or a primary-key lookup.
//
// This map is checked against migrations/*.sql by
// TestTablesMatchMigrations. Do not edit it by hand unless the test is
// failing and the reason is that a migration added or dropped the
// column — in which case edit the map to match the migration and re-run.
//
// The values below are the state at the time this package was written.
// The test is the source of truth.
var WorkspaceScopedTables = map[string]bool{
	"api_contracts":        true,
	"claim_verifications":  true,
	"claims":               true,
	"contradictions":       true,
	"cross_module_edges":   true,
	"cross_repo_edges":     true,
	"document_claims":      true,
	"entities":             true,
	"impact_assessments":   true,
	"mcp_agent_watermarks": true,
	"policy_evaluations":   true,
	"relationships":        true,
	"repositories":         true,
	"runtime_edges":        true,
	"runtime_observations": true,
	"workspace_members":    true,
	"workspace_modules":    true,
}
