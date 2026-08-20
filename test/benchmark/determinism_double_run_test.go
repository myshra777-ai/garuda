// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package benchmark

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/canonical"
)

// findCorpusDirs locates all test case directories across the benchmark suite.
func findCorpusDirs() []string {
	candidates := []string{
		filepath.Join("..", "..", "garuda-bench", "corpus", "cases"),
		"corpus/cases",
	}

	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			entries, _ := os.ReadDir(dir)
			var caseDirs []string
			for _, e := range entries {
				if e.IsDir() {
					caseDirs = append(caseDirs, filepath.Join(dir, e.Name()))
				}
			}
			return caseDirs
		}
	}
	return nil
}

// TestDoubleRunDeterminism validates Invariant 1-06: Repeated analysis on the same source
// must produce identical canonical bytes and snapshot hashes.
func TestDoubleRunDeterminism(t *testing.T) {
	caseDirs := findCorpusDirs()
	if len(caseDirs) == 0 {
		t.Fatal("benchmark corpus directory not found in garuda-bench/corpus/cases")
	}

	ctx := context.Background()
	goAnalyzer := analyzer.NewGoAnalyzer()
	validator := analyzer.NewPreLedgerValidator()

	for _, casePath := range caseDirs {
		caseName := filepath.Base(casePath)

		t.Run(caseName, func(t *testing.T) {
			req := analyzer.AnalysisRequest{
				Path:      casePath,
				CommitSHA: "test-commit-sha",
				Options: analyzer.AnalysisOptions{
					TypeResolution:   true,
					IncludeCallGraph: true,
				},
			}

			// Run 1: Primary extraction
			snap1, err := goAnalyzer.Analyze(ctx, req)
			if err != nil {
				t.Fatalf("run 1 extraction failed: %v", err)
			}
			if err := validator.Validate(*snap1); err != nil {
				t.Fatalf("run 1 failed pre-ledger validation: %v", err)
			}
			bytes1, hash1, err := canonical.CanonicalizeSnapshot(*snap1)
			if err != nil {
				t.Fatalf("run 1 canonicalization failed: %v", err)
			}

			// Run 2: Replay extraction
			snap2, err := goAnalyzer.Analyze(ctx, req)
			if err != nil {
				t.Fatalf("run 2 extraction failed: %v", err)
			}
			if err := validator.Validate(*snap2); err != nil {
				t.Fatalf("run 2 failed pre-ledger validation: %v", err)
			}
			bytes2, hash2, err := canonical.CanonicalizeSnapshot(*snap2)
			if err != nil {
				t.Fatalf("run 2 canonicalization failed: %v", err)
			}

			// Invariant 1-06: Snapshot hash determinism
			if hash1 != hash2 {
				t.Fatalf("determinism failure in %s: hash1 (%s) != hash2 (%s)", caseName, hash1, hash2)
			}

			// Canonical byte stream equality
			if !bytes.Equal(bytes1, bytes2) {
				t.Fatalf("determinism failure in %s: canonical byte streams differ despite identical inputs", caseName)
			}

			// Entity & relationship count invariants
			if len(snap1.Entities) != len(snap2.Entities) {
				t.Fatalf("entity count mismatch: %d != %d", len(snap1.Entities), len(snap2.Entities))
			}
			if len(snap1.Relationships) != len(snap2.Relationships) {
				t.Fatalf("relationship count mismatch: %d != %d", len(snap1.Relationships), len(snap2.Relationships))
			}
		})
	}
}
