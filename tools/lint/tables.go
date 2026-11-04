// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

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
	"api_contracts":                   true,
	"claim_verifications":             true,
	"claims":                          true,
	"contradictions":                  true,
	"cross_module_edges":              true,
	"cross_repo_edges":                true,
	"data_logic_audit_ledger":         true,
	"decisions":                       true,
	"document_claims":                 true,
	"entities":                        true,
	"financial_audit_reconciliations": true,
	"impact_assessments":              true,
	"logic_variants_registry":         true,
	"mcp_agent_watermarks":            true,
	"mcp_sessions":                    true,
	"policy_evaluations":              true,
	"relationships":                   true,
	"repositories":                    true,
	"runtime_edges":                   true,
	"runtime_observations":            true,
	"surface_criticality_tiers":       true,
	"synthesized_rulebooks":           true,
	"workspace_documents":             true,
	"workspace_members":               true,
	"workspace_modules":               true,
}
