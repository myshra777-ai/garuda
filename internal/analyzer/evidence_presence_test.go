// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package analyzer_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/myshra777-ai/garuda/internal/analyzer"
)

// TestEvidencePopulation_NoEmptyEdges asserts that every relationship
// emitted by the workspace analyzer carries non-empty source evidence.
//
// A failure here means a new emitter was added without consulting the
// evidence index, or an existing emitter stopped looking up positions.
// Either way, an edge without evidence breaks the "click to source"
// guarantee that the trust narrative depends on.
func TestEvidencePopulation_NoEmptyEdges(t *testing.T) {
	corpusDir := filepath.Join("..", "..", "garuda-bench", "corpus", "cases")
	entries, err := os.ReadDir(corpusDir)
	if err != nil {
		t.Skipf("corpus not found: %v", err)
	}

	var totalEdges, emptyFile, emptyLine int

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		casePath := filepath.Join(corpusDir, entry.Name())
		ws, err := analyzer.DiscoverWorkspace(casePath)
		if err != nil {
			continue
		}
		result, err := analyzer.AnalyzeWorkspace(context.Background(), ws)
		if err != nil {
			continue
		}
		for _, rel := range result.Relationships {
			totalEdges++
			if rel.Evidence.File == "" {
				emptyFile++
				t.Errorf("%s: edge %s -[%s]-> %s has empty Evidence.File",
					entry.Name(), rel.From, rel.Type, rel.To)
			}
			if rel.Evidence.LineStart == 0 {
				emptyLine++
				t.Errorf("%s: edge %s -[%s]-> %s has zero Evidence.LineStart",
					entry.Name(), rel.From, rel.Type, rel.To)
			}
		}
	}

	t.Logf("checked %d edges across %d fixtures: %d empty-file, %d empty-line",
		totalEdges, len(entries), emptyFile, emptyLine)
}
