// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

var toolArgWhitelist = map[string][]string{
	"garuda.briefing":                    {"workspace", "tenant_id", "agent_id"},
	"garuda.entities":                    {"workspace", "package", "kind", "limit", "tenant_id"},
	"garuda.find_entity":                 {"workspace", "name", "name_pattern", "kind", "package", "file_path_contains", "exported_only", "limit", "tenant_id"},
	"garuda.inspect":                     {"workspace", "entity_id", "name", "tenant_id"},
	"garuda.neighbors":                   {"workspace", "entity_id", "name", "edge_kinds", "include_external", "limit", "tenant_id"},
	"garuda.subclasses":                  {"workspace", "entity_id", "name", "language", "limit", "tenant_id"},
	"garuda.implementers":                {"workspace", "entity_id", "name", "limit", "tenant_id"},
	"garuda.blast_radius":                {"workspace", "entity_id", "name", "symbol", "depth", "min_confidence", "tenant_id"},
	"garuda.query":                       {},
	"garuda.query_claims":                {"workspace", "subject", "status", "tenant_id"},
	"garuda.policy.list":                 {"tenant_id"},
	"garuda.policy.evaluate":             {"workspace"},
	"garuda.verify_policy_evaluation":    {"evaluation_id", "tenant_id"},
	"garuda.verify_evidence":             {"evidence_id", "evaluation_id", "decision_id", "workspace", "tenant_id"},
	"garuda.governance.status":           {"workspace", "tenant_id"},
	"garuda.check_drift":                 {"workspace", "tenant_id"},
	"garuda.get_lineage":                 {"decision_id", "actor", "symbol", "name", "entity_id", "workspace", "tenant_id"},
	"garuda.get_impact":                  {"decision_id", "entity_id", "symbol", "name", "workspace", "tenant_id"},
	"garuda.detect_contradictions":       {"workspace", "scope", "status", "tenant_id"},
	"garuda.propose_decision":            {"scope_domain", "scope_system", "workspace", "tenant_id"},
	"garuda.create_task":                 {"workspace", "title", "description", "priority", "owner_agent_id", "target_agent_id", "agent_id", "task_id", "scope_domain", "scope_system", "tenant_id"},
	"garuda.handoff":                     {"task_id", "source_agent_id", "target_agent_id", "workspace", "tenant_id"},
	"garuda.resume":                      {"agent_id", "checkpoint_id", "workspace", "tenant_id"},
	"garuda.talon_pre_edit":              {"workspace", "target", "symbol", "file", "tenant_id"},
	"garuda.talon_verify_edit":           {"workspace", "repo_dir", "staged", "diff", "files", "save_ledger", "save_to_ledger", "tenant_id"},
	"garuda.disambiguate_logic":          {"workspace", "team", "concept", "tenant_id"},
	"garuda.decide_logic_variant":        {"workspace", "team", "variant_id", "decision_action", "custom_weight", "approved_by", "tenant_id"},
	"garuda.audit_data_discrepancy":      {"workspace", "metric", "team_a_name", "team_a_cr", "team_a_crore", "team_a_val", "team_b_name", "team_b_cr", "team_b_crore", "team_b_val", "tenant_id"},
	"garuda.reconcile_data_audit":        {"workspace", "audit_id", "metric", "approved_by", "tenant_id"},
	"garuda.audit_financial_discrepancy": {"workspace", "metric", "team_a_name", "team_a_cr", "team_a_crore", "team_a_val", "team_b_name", "team_b_cr", "team_b_crore", "team_b_val", "tenant_id"},
	"garuda.reconcile_financial_audit":   {"workspace", "audit_id", "metric", "approved_by", "tenant_id"},
	"garuda.synthesize_rulebook":         {"workspace", "output", "output_path", "path", "tenant_id"},
	"garuda.approve_base_rulebook":       {"workspace", "approved_by", "output", "output_path", "path", "tenant_id"},
	"garuda.criticality_tiers":           {"workspace", "tier_filter", "surface_name", "package_pattern", "criticality_tier", "tenant_id"},
	"garuda.docs_graph":                  {"workspace", "doc_type", "tenant_id"},
	"garuda.docs_ingest":                 {"workspace", "path", "title", "doc_type", "tenant_id"},
	"garuda.doctor_diagnose":             {"workspace", "fix", "tenant_id"},
	"garuda.build_matrix_check":          {"workspace", "repo_dir", "diff", "tenant_id"},
	"garuda.hygiene_report":              {"workspace", "limit", "tenant_id"},
	"garuda.reset_audit_state":           {"workspace", "tenant_id"},
}

// filterArgs returns a map containing only whitelisted keys for the
// given tool. Values are passed through unchanged. Empty whitelist →
// empty map. Unknown tool → empty map (fails closed).
func filterArgs(toolName string, args map[string]interface{}) map[string]interface{} {
	allow, ok := toolArgWhitelist[toolName]
	if !ok || len(allow) == 0 {
		return map[string]interface{}{}
	}
	out := make(map[string]interface{}, len(allow))
	for _, key := range allow {
		if v, ok := args[key]; ok {
			out[key] = v
		}
	}
	return out
}
