// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestIncrementalWorkspaceAnalysis_CacheSkipping(t *testing.T) {
	tempDir := t.TempDir()

	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module github.com/garuda/incremental\n\ngo 1.22\n"), 0644)
	pkgA := filepath.Join(tempDir, "pkg_a")
	_ = os.MkdirAll(pkgA, 0755)
	_ = os.WriteFile(filepath.Join(pkgA, "a.go"), []byte("package pkg_a\ntype ModelA struct { ID string }\n"), 0644)

	ws, err := DiscoverWorkspace(tempDir)
	if err != nil {
		t.Fatalf("DiscoverWorkspace failed: %v", err)
	}

	cache := NewMemoryPackageCache()
	tenantID := uuid.New()
	opts := WorkspaceAnalysisOptions{TenantID: tenantID, Cache: cache}

	// 1. Cold run (populates cache)
	res1, err := AnalyzeWorkspaceWithOptions(context.Background(), ws, opts)
	if err != nil {
		t.Fatalf("Cold run failed: %v", err)
	}
	if res1.Stats.Files != 1 {
		t.Fatalf("expected 1 parsed file in cold run, got %d", res1.Stats.Files)
	}

	// 2. Warm run (must hit cache: 0 files parsed)
	res2, err := AnalyzeWorkspaceWithOptions(context.Background(), ws, opts)
	if err != nil {
		t.Fatalf("Warm run failed: %v", err)
	}
	if res2.Stats.Files != 0 {
		t.Fatalf("expected 0 files parsed in warm run due to cache hit, got %d", res2.Stats.Files)
	}

	// 3. Assert results are equivalent
	if len(res1.Entities) != len(res2.Entities) {
		t.Errorf("entity count mismatch: cold=%d, warm=%d", len(res1.Entities), len(res2.Entities))
	}
	if res1.Fingerprint != res2.Fingerprint {
		t.Errorf("fingerprint mismatch between cold and warm runs")
	}
}
