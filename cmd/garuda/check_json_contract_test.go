// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"reflect"
	"strings"
	"testing"
)

// TestCheckJSONContract pins the keys of `garuda check --json`, which CI
// integrations (including the GitHub Action) parse. Adding a key is fine;
// renaming or removing one is a breaking change and must be made on purpose
// by updating this list.
func TestCheckJSONContract(t *testing.T) {
	want := []string{
		"workspace", "target_symbol", "diff_mode", "touched_files", "touched_symbols",
		"touched_packages", "affected_entities", "by_severity", "evidence_state",
		"adr_contradictions", "policies_evaluated", "net_new_violations", "baseline_violations",
		"violations", "self_repair_instructions", "blocked_by", "blocked_by_policy",
		"governance_outcome", "merkle_block_height", "build_context", "environment_drift",
		"matrix_results", "reproduce_command", "evidence_coverage", "impacted_entities",
		"impacted_entities_truncated", "duration_ms",
	}
	typ := reflect.TypeOf(ChangeControlDecision{})
	got := make([]string, 0, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		if name := strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]; name != "" && name != "-" {
			got = append(got, name)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("check --json keys changed:\n got  %v\n want %v", got, want)
	}
}
