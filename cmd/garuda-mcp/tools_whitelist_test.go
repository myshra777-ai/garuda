// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import "testing"

func TestToolArgWhitelistCoversAllTools(t *testing.T) {
	declared := make(map[string]bool, len(allTools))
	for _, tool := range allTools {
		name, _ := tool["name"].(string)
		if name == "" {
			t.Errorf("tool in allTools has no name")
			continue
		}
		declared[name] = true
		if _, ok := toolArgWhitelist[name]; !ok {
			t.Errorf("tool %q is declared in allTools but has no whitelist entry", name)
		}
	}
	for name := range toolArgWhitelist {
		if !declared[name] {
			t.Errorf("tool %q has a whitelist entry but is not in allTools", name)
		}
	}
}

func TestFilterArgsDropsUnlistedKeys(t *testing.T) {
	got := filterArgs("garuda.query", map[string]interface{}{
		"query":     "should be dropped",
		"workspace": "should also be dropped",
	})
	if len(got) != 0 {
		t.Errorf("garuda.query should filter all args, got %v", got)
	}

	got = filterArgs("garuda.policy.evaluate", map[string]interface{}{
		"workspace":  "go-validation-10",
		"policy_dir": "/secret/path",
	})
	if got["workspace"] != "go-validation-10" {
		t.Errorf("workspace should be kept, got %v", got)
	}
	if _, ok := got["policy_dir"]; ok {
		t.Errorf("policy_dir should be dropped, got %v", got["policy_dir"])
	}
}
