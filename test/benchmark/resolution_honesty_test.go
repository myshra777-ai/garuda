// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Enforces that the analyzer does not claim certainty it has not earned.
// Any relationship with confidence >= 1.0 MUST declare ResolutionMethod = GO_TYPES.

package benchmark

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/types"
)

func TestResolutionHonesty_NoUnearnedCertainty(t *testing.T) {
	ctx := context.Background()

	fixturesRoot := filepath.Join("truth_fixtures")
	entries, err := os.ReadDir(fixturesRoot)
	if err != nil {
		t.Fatalf("cannot read truth_fixtures: %v", err)
	}

	var checked, violations int
	var methodCounts = map[string]int{}
	var statusCounts = map[string]int{}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		fixturePath := filepath.Join(fixturesRoot, e.Name())

		a := analyzer.NewGoAnalyzer()
		snap, err := a.Analyze(ctx, analyzer.AnalysisRequest{
			Path:      fixturePath,
			CommitSHA: "test-head",
			Options: analyzer.AnalysisOptions{
				IncludeCallGraph: true,
				TypeResolution:   true,
			},
		})
		if err != nil {
			t.Logf("skip %s: %v", e.Name(), err)
			continue
		}

		for _, rel := range snap.Relationships {
			checked++
			methodCounts[string(rel.ResolutionMethod)]++
			statusCounts[string(rel.ResolutionStatus)]++

			// Rule 1: Confidence == 1.0 must mean type-checked.
			if rel.Confidence >= 1.0 && rel.ResolutionMethod != types.ResolutionMethodGoTypes {
				violations++
				t.Errorf("UNEARNED CERTAINTY in %s: %s --%s--> %s "+
					"(confidence=%.2f, method=%q, status=%q)",
					e.Name(), rel.SourceName, rel.Predicate, rel.TargetName,
					rel.Confidence, rel.ResolutionMethod, rel.ResolutionStatus)
			}

			// Rule 2: An empty resolution method is a bug, not a default.
			if rel.ResolutionMethod == "" {
				violations++
				t.Errorf("MISSING METHOD in %s: %s --%s--> %s (confidence=%.2f)",
					e.Name(), rel.SourceName, rel.Predicate, rel.TargetName, rel.Confidence)
			}
		}
	}

	t.Logf("\n"+
		"═══════════════════════════════════════════════\n"+
		"  RESOLUTION HONESTY AUDIT\n"+
		"═══════════════════════════════════════════════\n"+
		"  Fixtures scanned:  %d\n"+
		"  Relationships:     %d\n"+
		"  Violations:        %d\n"+
		"\n"+
		"  By method:\n"+
		"    GO_TYPES:           %d\n"+
		"    IMPORT_RESOLUTION:  %d\n"+
		"    AST_EXACT:          %d\n"+
		"    HEURISTIC:          %d\n"+
		"    (empty):            %d\n"+
		"\n"+
		"  By status:\n"+
		"    RESOLVED:           %d\n"+
		"    AMBIGUOUS:          %d\n"+
		"    UNRESOLVED:         %d\n"+
		"    INFERRED:           %d\n"+
		"    (empty):            %d\n"+
		"═══════════════════════════════════════════════",
		len(entries), checked, violations,
		methodCounts["GO_TYPES"],
		methodCounts["IMPORT_RESOLUTION"],
		methodCounts["AST_EXACT"],
		methodCounts["HEURISTIC"],
		methodCounts[""],
		statusCounts["RESOLVED"],
		statusCounts["AMBIGUOUS"],
		statusCounts["UNRESOLVED"],
		statusCounts["INFERRED"],
		statusCounts[""],
	)
}
