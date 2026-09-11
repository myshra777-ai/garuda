// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/myshra777-ai/garuda/internal/analyzer"
)

func TestAnalyze_SelfRepo(t *testing.T) {
	// The analyzer requires a Go module. Walk up to find go.mod,
	// unless we're in a temp dir without one — in that case, skip.
	root, err := findModuleRoot(".")
	if err != nil {
		t.Skipf("no go.mod found from %q; skipping workspace-level test", ".")
	}

	result, err := analyzer.AnalyzeDirectory(context.Background(), root)
	if err != nil {
		t.Fatalf("AnalyzeDirectory failed on %s: %v", root, err)
	}

	if result.Stats.Files == 0 {
		t.Error("expected at least one file")
	}
	if result.Fingerprint == "" {
		t.Error("expected fingerprint")
	}
	if len(result.Entities) == 0 {
		t.Error("expected at least one entity")
	}
	if len(result.Relationships) == 0 {
		t.Error("expected at least one relationship")
	}

	// Every relationship must carry an honest classification.
	for i, r := range result.Relationships {
		if r.ResolutionMethod == "" {
			t.Errorf("relationship %d (%s -> %s) has empty resolution_method", i, r.From, r.To)
		}
		if r.Confidence <= 0 {
			t.Errorf("relationship %d (%s -> %s) has non-positive confidence: %f", i, r.From, r.To, r.Confidence)
		}
	}
}

func findModuleRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
