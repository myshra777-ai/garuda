// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

var toolArgWhitelist = map[string][]string{
	"garuda.briefing":                 {"workspace", "tenant_id"},
	"garuda.entities":                 {"workspace", "package", "kind", "limit"},
	"garuda.find_entity":              {"workspace", "name_pattern", "kind", "package", "limit"},
	"garuda.inspect":                  {"workspace", "entity_id"},
	"garuda.neighbors":                {"workspace", "entity_id", "limit"},
	"garuda.subclasses":               {"workspace", "entity_id", "name", "language", "limit"},
	"garuda.implementers":             {"workspace", "entity_id", "name", "limit"},
	"garuda.blast_radius":             {"workspace", "entity_id", "depth", "min_confidence"},
	"garuda.query":                    {},
	"garuda.query_claims":             {"workspace", "subject"},
	"garuda.policy.list":              {"tenant_id"},
	"garuda.policy.evaluate":          {"workspace"},
	"garuda.verify_policy_evaluation": {"evaluation_id", "tenant_id"},
	"garuda.governance.status":        {"workspace"},
	"garuda.check_drift":              {"workspace"},
	"garuda.get_lineage":              {"decision_id", "tenant_id"},
	"garuda.get_impact":               {"decision_id", "tenant_id"},
	"garuda.detect_contradictions":    {"tenant_id"},
	"garuda.propose_decision":         {"scope_domain", "scope_system", "tenant_id"},
	"garuda.handoff":                  {"task_id", "source_agent_id", "target_agent_id"},
	"garuda.resume":                   {"agent_id", "checkpoint_id"},
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
