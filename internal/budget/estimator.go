// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package budget

import (
	"encoding/json"
)

// EstimateTokens returns an estimated token count for an operation.
func EstimateTokens(operation string, payload interface{}) int {
	switch operation {
	case "propose_decision", "mcp_propose_decision":
		return estimateProposeDecision(payload)
	case "query", "mcp_query":
		return estimateQuery(payload)
	case "get_lineage", "mcp_get_lineage":
		return 30
	case "detect_contradictions", "mcp_detect_contradictions":
		return 50
	case "get_impact", "mcp_get_impact":
		return 30
	case "checkpoint_save", "mcp_checkpoint_save":
		return 50
	case "checkpoint_restore":
		return 20
	case "handoff":
		return 40
	default:
		return 100 // fallback
	}
}

func estimateProposeDecision(payload interface{}) int {
	text := extractString(payload, "title")
	if text == "" {
		text = extractString(payload, "statement")
	}
	if text == "" {
		text = extractString(payload, "decision")
	}
	return 100 + len(text)/4
}

func estimateQuery(payload interface{}) int {
	query := extractString(payload, "query")
	return 50 + len(query)/4
}

func extractString(payload interface{}, field string) string {
	if payload == nil {
		return ""
	}
	switch v := payload.(type) {
	case map[string]interface{}:
		if val, ok := v[field]; ok {
			if s, ok := val.(string); ok {
				return s
			}
		}
	case string:
		if field == "query" || field == "title" || field == "statement" {
			return v
		}
	}

	if bytes, err := json.Marshal(payload); err == nil {
		var m map[string]interface{}
		if json.Unmarshal(bytes, &m) == nil {
			if val, ok := m[field]; ok {
				if s, ok := val.(string); ok {
					return s
				}
			}
		}
	}
	return ""
}
