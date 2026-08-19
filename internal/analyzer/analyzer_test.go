// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer_test

import (
	"testing"

	"github.com/myshra777-ai/garuda/internal/analyzer"
)

func TestAnalyze(t *testing.T) {
	// Use the current directory as a test repo
	result, err := analyzer.Analyze(".")
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result.Stats.Files == 0 {
		t.Error("expected at least one file")
	}
	if result.Fingerprint == "" {
		t.Error("expected fingerprint")
	}

	// Check that JSON serialization works
	// (We'll rely on the report generation)
}
