// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package talon

import (
	"reflect"
	"strings"
	"testing"
)

// jsonKeys lists the JSON keys a struct declares, in field order.
func jsonKeys(t reflect.Type) []string {
	keys := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		if name == "" || name == "-" {
			continue
		}
		keys = append(keys, name)
	}
	return keys
}

// TestGuardVerdictJSONContract pins the keys of `talon guard --json` and
// garuda.talon_verify_edit. Agents, hooks and CI parse these; add new keys
// at will, but renaming or removing one is a breaking change and must be
// done on purpose by updating this list.
func TestGuardVerdictJSONContract(t *testing.T) {
	want := []string{
		"tenant_id", "workspace", "workspace_id", "outcome", "agent_should_proceed",
		"save_to_ledger", "side_effects", "touched_files", "touched_symbols", "touched_packages",
		"blast_radius_affected", "by_severity", "net_new_violations_count",
		"baseline_violations_count", "adr_contradictions_count", "baseline_id", "violations",
		"baseline_violations", "self_repair_instructions", "self_repair_prompt",
		"merkle_block_height", "build_context", "environment_drift", "matrix_results",
		"reproduce_command", "evidence_coverage", "duration_ms",
	}
	if got := jsonKeys(reflect.TypeOf(GuardVerdict{})); !reflect.DeepEqual(got, want) {
		t.Fatalf("GuardVerdict JSON keys changed:\n got  %v\n want %v", got, want)
	}
}
