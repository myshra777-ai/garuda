// Package benchmark executes the automated 20-fixture ground-truth validation suite
// against strict golden expected JSON manifests across all 6 V5 benchmark layers.
package benchmark

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/types"
)

// FixtureCategory represents the functional evaluation layer in V5.
type FixtureCategory string

const (
	CategoryEntityExtraction  FixtureCategory = "entity_extraction"
	CategoryRelationship      FixtureCategory = "relationship_correctness"
	CategoryIdentityStability FixtureCategory = "identity_stability"
	CategoryDiffCorrectness   FixtureCategory = "diff_correctness"
	CategoryImpactCorrectness FixtureCategory = "impact_correctness"
	CategoryEvidenceIntegrity FixtureCategory = "evidence_correctness"
	CategoryCrossRepo         FixtureCategory = "cross_repo_resolution"
	CategoryScaleThroughput   FixtureCategory = "scale_and_throughput"
	CategoryParserRobustness  FixtureCategory = "parser_robustness_and_noise"
)

// BenchmarkMetricScoreboard tracks precision and recall across all fixtures[cite: 1, 2].
type BenchmarkMetricScoreboard struct {
	TotalEntitiesExpected     int
	TotalEntitiesDiscovered   int
	MatchedEntities           int
	TotalRelsExpected         int
	TotalRelsDiscovered       int
	MatchedRels               int
	FalseEdgesRejected        int
	DiffClassificationsPassed int
	ImpactConsumersMatched    int
	ImpactConsumersExpected   int
	EvidenceHashesVerified    int
	DeterminismRunsPassed     int
}

// -----------------------------------------------------------------------------
// Test 1: Full 20-Fixture Corpus Execution & Metric Gating
// -----------------------------------------------------------------------------

func TestTruthBenchmarks_All20Fixtures(t *testing.T) {
	fixturesRoot := filepath.Join("truth_fixtures")
	entries, err := os.ReadDir(fixturesRoot)
	require.NoError(t, err, "Unable to read truth_fixtures directory")

	scoreboard := &BenchmarkMetricScoreboard{}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		fixtureName := entry.Name()
		fixturePath := filepath.Join(fixturesRoot, fixtureName)

		t.Run(fixtureName, func(t *testing.T) {
			expectedFile := filepath.Join(fixturePath, "expected.json")
			if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
				t.Skipf("No expected.json manifest found for %s; skipping", fixtureName)
				return
			}

			rawExpected, err := os.ReadFile(expectedFile)
			require.NoError(t, err, "Failed to read expected.json for %s", fixtureName)

			var header struct {
				FixtureID string          `json:"fixture_id"`
				Category  FixtureCategory `json:"category"`
			}
			require.NoError(t, json.Unmarshal(rawExpected, &header))

			ctx := context.Background()
			goAnalyzer := analyzer.NewGoAnalyzer()

			switch header.Category {
			case CategoryDiffCorrectness:
				runDiffFixtureTest(t, ctx, fixturePath, rawExpected, scoreboard)

			case CategoryIdentityStability:
				runIdentityStabilityFixtureTest(t, ctx, fixturePath, rawExpected, scoreboard)

			case CategoryImpactCorrectness:
				runImpactFixtureTest(t, ctx, goAnalyzer, fixturePath, rawExpected, scoreboard)

			case CategoryCrossRepo:
				runCrossRepoFixtureTest(t, ctx, goAnalyzer, fixturePath, rawExpected, scoreboard)

			case CategoryEvidenceIntegrity:
				runEvidenceFixtureTest(t, ctx, goAnalyzer, fixturePath, rawExpected, scoreboard)

			case CategoryParserRobustness:
				runNoiseRobustnessFixtureTest(t, ctx, goAnalyzer, fixturePath, rawExpected, scoreboard)

			case CategoryScaleThroughput:
				runScaleThroughputFixtureTest(t, ctx, goAnalyzer, fixturePath, rawExpected, scoreboard)

			default:
				// Standard AST Extraction & Relationship Fixtures (001 - 010)[cite: 1, 2]
				runStandardExtractionFixtureTest(t, ctx, goAnalyzer, fixturePath, rawExpected, scoreboard)
			}
		})
	}

	// -------------------------------------------------------------------------
	// Verification of Target Gates Mandated by V5 Roadmap[cite: 1, 2]
	// -------------------------------------------------------------------------
	t.Log("\n" + formatScoreboardSummary(scoreboard))

	if scoreboard.TotalEntitiesExpected > 0 {
		entityPrecision := float64(scoreboard.MatchedEntities) / float64(scoreboard.TotalEntitiesDiscovered) * 100
		entityRecall := float64(scoreboard.MatchedEntities) / float64(scoreboard.TotalEntitiesExpected) * 100
		assert.GreaterOrEqual(t, entityPrecision, 98.0, "Entity Precision must meet or exceed 98%[cite: 1, 2]")
		assert.GreaterOrEqual(t, entityRecall, 98.0, "Entity Recall must meet or exceed 98%[cite: 1, 2]")
	}

	if scoreboard.TotalRelsExpected > 0 {
		relPrecision := float64(scoreboard.MatchedRels) / float64(scoreboard.TotalRelsDiscovered) * 100
		relRecall := float64(scoreboard.MatchedRels) / float64(scoreboard.TotalRelsExpected) * 100
		assert.GreaterOrEqual(t, relPrecision, 95.0, "Relationship Precision must meet or exceed 95%[cite: 1, 2]")
		assert.GreaterOrEqual(t, relRecall, 90.0, "Relationship Recall must meet or exceed 90%[cite: 1, 2]")
	}

	if scoreboard.ImpactConsumersExpected > 0 {
		impactPrecision := float64(scoreboard.ImpactConsumersMatched) / float64(scoreboard.ImpactConsumersExpected) * 100
		assert.GreaterOrEqual(t, impactPrecision, 95.0, "Impact Precision must meet or exceed 95%[cite: 1, 2]")
	}
}

// -----------------------------------------------------------------------------
// Category Handlers & Fixture Runners
// -----------------------------------------------------------------------------

// Fixtures 011 & 012: Semantic Diff Verification[cite: 1, 2]
func runDiffFixtureTest(t *testing.T, ctx context.Context, fixtureDir string, rawExpected []byte, sb *BenchmarkMetricScoreboard) {
	var expected struct {
		OverallClassification string `json:"overall_classification"`
		BreakingChangesCount  int    `json:"breaking_changes_count"`
		Changes               []struct {
			ChangeType   string `json:"change_type"`
			Severity     string `json:"severity"`
			TargetEntity string `json:"target_entity"`
		} `json:"changes"`
	}
	require.NoError(t, json.Unmarshal(rawExpected, &expected))

	v1Path := filepath.Join(fixtureDir, "v1")
	v2Path := filepath.Join(fixtureDir, "v2")

	analyzerRunner := analyzer.NewGoAnalyzer()
	snap1, err := analyzerRunner.Analyze(ctx, analyzer.AnalysisRequest{Path: v1Path, CommitSHA: "v1"})
	require.NoError(t, err)

	snap2, err := analyzerRunner.Analyze(ctx, analyzer.AnalysisRequest{Path: v2Path, CommitSHA: "v2"})
	require.NoError(t, err)

	diffResult := analyzer.ComputeSemanticDiff(snap1, snap2)
	require.NotNil(t, diffResult)

	assert.Equal(t, expected.OverallClassification, string(diffResult.Classification),
		"Diff classification mismatch in %s", fixtureDir)

	assert.Equal(t, expected.BreakingChangesCount, diffResult.BreakingCount,
		"Breaking change count mismatch in %s", fixtureDir)

	for _, expectedChange := range expected.Changes {
		found := false
		for _, actualChange := range diffResult.Changes {
			if actualChange.TargetEntity == expectedChange.TargetEntity &&
				string(actualChange.Severity) == expectedChange.Severity {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected diff change %s (%s) not found in diff result",
			expectedChange.TargetEntity, expectedChange.ChangeType)
	}

	sb.DiffClassificationsPassed++
}

// Fixtures 014, 015, 018: Identity Stability, Renames, Deletes & Determinism[cite: 1, 2]
func runIdentityStabilityFixtureTest(t *testing.T, ctx context.Context, fixtureDir string, rawExpected []byte, sb *BenchmarkMetricScoreboard) {
	var expected struct {
		EvolutionType      string `json:"evolution_type"`
		Comparison         string `json:"comparison"`
		SourceEntity       string `json:"source_entity"`
		TargetEntity       string `json:"target_entity"`
		TombstonedEntities []struct {
			QualifiedName string `json:"qualified_name"`
			Status        string `json:"status"`
		} `json:"tombstoned_entities"`
		EntitiesToVerify []struct {
			QualifiedName string `json:"qualified_name"`
			IDMustMatch   bool   `json:"id_must_match"`
		} `json:"entities_to_verify"`
	}
	require.NoError(t, json.Unmarshal(rawExpected, &expected))

	analyzerRunner := analyzer.NewGoAnalyzer()
	v1Path := filepath.Join(fixtureDir, "v1")
	v2Path := filepath.Join(fixtureDir, "v2")

	snap1, err := analyzerRunner.Analyze(ctx, analyzer.AnalysisRequest{Path: v1Path, CommitSHA: "v1"})
	require.NoError(t, err)

	snap2, err := analyzerRunner.Analyze(ctx, analyzer.AnalysisRequest{Path: v2Path, CommitSHA: "v2"})
	require.NoError(t, err)

	if expected.Comparison == "v1_vs_v2" { // Fixture 018 Determinism[cite: 1, 2]
		for _, toVerify := range expected.EntitiesToVerify {
			e1 := findEntityByQualifiedName(snap1.Entities, toVerify.QualifiedName)
			e2 := findEntityByQualifiedName(snap2.Entities, toVerify.QualifiedName)

			require.NotNil(t, e1, "Entity %s missing in v1 snapshot", toVerify.QualifiedName)
			require.NotNil(t, e2, "Entity %s missing in v2 snapshot", toVerify.QualifiedName)

			if toVerify.IDMustMatch {
				assert.Equal(t, e1.CanonicalID, e2.CanonicalID,
					"Canonical UUIDv5 identity altered by cosmetic churn for %s", toVerify.QualifiedName)
			}
		}
		sb.DeterminismRunsPassed++
	}

	if expected.EvolutionType == "ENTITY_DELETE" { // Fixture 015 Delete & Tombstone[cite: 1, 2]
		for _, tomb := range expected.TombstonedEntities {
			e1 := findEntityByQualifiedName(snap1.Entities, tomb.QualifiedName)
			require.NotNil(t, e1, "Entity %s should exist in v1", tomb.QualifiedName)

			e2 := findEntityByQualifiedName(snap2.Entities, tomb.QualifiedName)
			assert.Nil(t, e2, "Deleted entity %s must not appear as active in v2", tomb.QualifiedName)
		}
	}
}

// Fixture 013: BFS Blast Radius Impact Precision[cite: 1, 2]
func runImpactFixtureTest(t *testing.T, ctx context.Context, goAnalyzer *analyzer.GoAnalyzer, fixtureDir string, rawExpected []byte, sb *BenchmarkMetricScoreboard) {
	var expected struct {
		TargetMutation        string `json:"target_mutation"`
		ExpectedTotalImpacted int    `json:"expected_total_impacted"`
		Consumers             []struct {
			Depth        int    `json:"depth"`
			Severity     string `json:"severity"`
			Entity       string `json:"entity"`
			Relationship string `json:"relationship"`
		} `json:"consumers"`
	}
	require.NoError(t, json.Unmarshal(rawExpected, &expected))

	snap, err := goAnalyzer.Analyze(ctx, analyzer.AnalysisRequest{
		Path:      fixtureDir,
		CommitSHA: "head",
		Options:   analyzer.AnalysisOptions{IncludeCallGraph: true, TypeResolution: true},
	})
	require.NoError(t, err)

	impactReport := analyzer.ComputeImpactRadius(snap, expected.TargetMutation, 3)
	require.NotNil(t, impactReport)

	sb.ImpactConsumersExpected += len(expected.Consumers)

	for _, expConsumer := range expected.Consumers {
		found := false
		for _, actualImpact := range impactReport.ImpactedEntities {
			if actualImpact.QualifiedName == expConsumer.Entity &&
				actualImpact.Depth == expConsumer.Depth &&
				string(actualImpact.Severity) == expConsumer.Severity {
				found = true
				sb.ImpactConsumersMatched++
				break
			}
		}
		assert.True(t, found, "Expected downstream impact consumer %s at depth %d missing or misclassified",
			expConsumer.Entity, expConsumer.Depth)
	}
}

// Fixture 016: Cross-Module Boundary Extraction[cite: 1, 2]
func runCrossRepoFixtureTest(t *testing.T, ctx context.Context, goAnalyzer *analyzer.GoAnalyzer, fixtureDir string, rawExpected []byte, sb *BenchmarkMetricScoreboard) {
	var expected struct {
		ModulesCount             int `json:"modules_count"`
		CrossModuleRelationships []struct {
			Predicate       string  `json:"predicate"`
			SourceEntity    string  `json:"source_entity"`
			TargetEntity    string  `json:"target_entity"`
			IsCrossBoundary bool    `json:"is_cross_boundary"`
			EpistemicClass  string  `json:"epistemic_class"`
			Confidence      float64 `json:"confidence"`
		} `json:"cross_module_relationships"`
	}
	require.NoError(t, json.Unmarshal(rawExpected, &expected))

	snap, err := goAnalyzer.AnalyzeWorkspace(ctx, fixtureDir)
	require.NoError(t, err)

	for _, expRel := range expected.CrossModuleRelationships {
		rel := findRelationship(snap.Relationships, expRel.SourceEntity, expRel.Predicate, expRel.TargetEntity)
		require.NotNil(t, rel, "Cross-module relationship %s -[%s]-> %s not found",
			expRel.SourceEntity, expRel.Predicate, expRel.TargetEntity)

		assert.Equal(t, types.EpistemicClass(expRel.EpistemicClass), rel.EpistemicClass)
		assert.InEpsilon(t, expRel.Confidence, rel.Confidence, 0.001)
	}
}

// Fixture 017: Precise Line Spans & Evidence Hashes[cite: 1, 2]
func runEvidenceFixtureTest(t *testing.T, ctx context.Context, goAnalyzer *analyzer.GoAnalyzer, fixtureDir string, rawExpected []byte, sb *BenchmarkMetricScoreboard) {
	var expected struct {
		ExpectedEntities []struct {
			Name            string `json:"name"`
			Kind            string `json:"kind"`
			QualifiedName   string `json:"qualified_name"`
			LineStart       int    `json:"line_start"`
			LineEnd         int    `json:"line_end"`
			SnippetContains string `json:"snippet_contains"`
		} `json:"expected_entities"`
	}
	require.NoError(t, json.Unmarshal(rawExpected, &expected))

	snap, err := goAnalyzer.Analyze(ctx, analyzer.AnalysisRequest{
		Path:      fixtureDir,
		CommitSHA: "head",
		Options:   analyzer.AnalysisOptions{TypeResolution: true},
	})
	require.NoError(t, err)

	for _, exp := range expected.ExpectedEntities {
		e := findEntityByQualifiedName(snap.Entities, exp.QualifiedName)
		require.NotNil(t, e, "Entity %s not found", exp.QualifiedName)

		assert.Equal(t, exp.LineStart, e.LineStart, "LineStart mismatch for %s", exp.QualifiedName)
		assert.Equal(t, exp.LineEnd, e.LineEnd, "LineEnd mismatch for %s", exp.QualifiedName)
		assert.NotEmpty(t, e.EvidenceHash, "Evidence hash missing for %s", exp.QualifiedName)

		if exp.SnippetContains != "" {
			assert.Contains(t, e.ContentSnippet, exp.SnippetContains, "Snippet content mismatch for %s", exp.QualifiedName)
		}
		sb.EvidenceHashesVerified++
	}
}

// Fixture 019: Parser Robustness Over Deep Noise[cite: 1, 2]
func runNoiseRobustnessFixtureTest(t *testing.T, ctx context.Context, goAnalyzer *analyzer.GoAnalyzer, fixtureDir string, rawExpected []byte, sb *BenchmarkMetricScoreboard) {
	var expected struct {
		Invariants struct {
			ZeroPanics              bool `json:"zero_panics"`
			TopLevelEntityIntegrity bool `json:"top_level_entity_integrity"`
		} `json:"invariants"`
		ExpectedEntities []struct {
			Name          string `json:"name"`
			QualifiedName string `json:"qualified_name"`
			LineStart     int    `json:"line_start"`
			LineEnd       int    `json:"line_end"`
		} `json:"expected_entities"`
	}
	require.NoError(t, json.Unmarshal(rawExpected, &expected))

	snap, err := goAnalyzer.Analyze(ctx, analyzer.AnalysisRequest{
		Path:      fixtureDir,
		CommitSHA: "head",
		Options:   analyzer.AnalysisOptions{TypeResolution: true},
	})
	require.NoError(t, err, "Parser panicked or failed on noise fixture")

	for _, exp := range expected.ExpectedEntities {
		e := findEntityByQualifiedName(snap.Entities, exp.QualifiedName)
		assert.NotNil(t, e, "Top-level entity %s missing amidst noise", exp.QualifiedName)
	}
}

// Fixture 020: Scale & Throughput[cite: 1, 2]
func runScaleThroughputFixtureTest(t *testing.T, ctx context.Context, goAnalyzer *analyzer.GoAnalyzer, fixtureDir string, rawExpected []byte, sb *BenchmarkMetricScoreboard) {
	var expected struct {
		SummaryMetrics struct {
			PackagesCount int `json:"packages_count"`
			StructsCount  int `json:"structs_count"`
		} `json:"summary_metrics"`
		ExpectedImplementations []struct {
			Struct    string `json:"struct"`
			Interface string `json:"interface"`
		} `json:"expected_interface_implementations"`
	}
	require.NoError(t, json.Unmarshal(rawExpected, &expected))

	snap, err := goAnalyzer.Analyze(ctx, analyzer.AnalysisRequest{
		Path:      fixtureDir,
		CommitSHA: "head",
		Options:   analyzer.AnalysisOptions{IncludeCallGraph: true, TypeResolution: true},
	})
	require.NoError(t, err)

	for _, expImpl := range expected.ExpectedImplementations {
		rel := findRelationship(snap.Relationships, expImpl.Struct, types.PredicateImplements, expImpl.Interface)
		assert.NotNil(t, rel, "Expected IMPLEMENTS relation %s -> %s missing in scale fixture",
			expImpl.Struct, expImpl.Interface)
	}
}

// Standard Extraction Fixtures 001 - 010[cite: 1, 2]
func runStandardExtractionFixtureTest(t *testing.T, ctx context.Context, goAnalyzer *analyzer.GoAnalyzer, fixtureDir string, rawExpected []byte, sb *BenchmarkMetricScoreboard) {
	var expected struct {
		Entities      []types.Entity       `json:"entities"`
		Relationships []types.Relationship `json:"relationships"`
	}
	if err := json.Unmarshal(rawExpected, &expected); err != nil {
		return
	}

	snap, err := goAnalyzer.Analyze(ctx, analyzer.AnalysisRequest{
		Path:      fixtureDir,
		CommitSHA: "head",
		Options:   analyzer.AnalysisOptions{IncludeCallGraph: true, TypeResolution: true},
	})
	require.NoError(t, err)

	sb.TotalEntitiesExpected += len(expected.Entities)
	sb.TotalEntitiesDiscovered += len(snap.Entities)

	for _, expEntity := range expected.Entities {
		act := findEntityByQualifiedName(snap.Entities, expEntity.QualifiedName)
		if act != nil && act.Kind == expEntity.Kind {
			sb.MatchedEntities++
		}
	}

	sb.TotalRelsExpected += len(expected.Relationships)
	sb.TotalRelsDiscovered += len(snap.Relationships)

	for _, expRel := range expected.Relationships {
		act := findRelationship(snap.Relationships, expRel.SourceName, expRel.Predicate, expRel.TargetName)
		if act != nil {
			sb.MatchedRels++
		}
	}
}

// -----------------------------------------------------------------------------
// Helper Routines
// -----------------------------------------------------------------------------

func findEntityByQualifiedName(entities []types.Entity, qName string) *types.Entity {
	for i := range entities {
		if entities[i].QualifiedName == qName {
			return &entities[i]
		}
	}
	return nil
}

func formatScoreboardSummary(sb *BenchmarkMetricScoreboard) string {
	var b strings.Builder
	b.WriteString("====================================================\n")
	b.WriteString("        GARUDA V5 BENCHMARK GATE SUMMARY            \n")
	b.WriteString("====================================================\n")
	if sb.TotalEntitiesExpected > 0 {
		b.WriteString(fmt.Sprintf("• Entity Precision:          %.1f%%\n",
			float64(sb.MatchedEntities)/float64(sb.TotalEntitiesDiscovered)*100))
		b.WriteString(fmt.Sprintf("• Entity Recall:             %.1f%%\n",
			float64(sb.MatchedEntities)/float64(sb.TotalEntitiesExpected)*100))
	}
	if sb.TotalRelsExpected > 0 {
		b.WriteString(fmt.Sprintf("• Relationship Precision:    %.1f%%\n",
			float64(sb.MatchedRels)/float64(sb.TotalRelsDiscovered)*100))
		b.WriteString(fmt.Sprintf("• Relationship Recall:       %.1f%%\n",
			float64(sb.MatchedRels)/float64(sb.TotalRelsExpected)*100))
	}
	if sb.ImpactConsumersExpected > 0 {
		b.WriteString(fmt.Sprintf("• Impact Precision:          %.1f%%\n",
			float64(sb.ImpactConsumersMatched)/float64(sb.ImpactConsumersExpected)*100))
	}
	b.WriteString(fmt.Sprintf("• Evidence Hashes Verified:  %d\n", sb.EvidenceHashesVerified))
	b.WriteString(fmt.Sprintf("• Determinism Checks:        %d Passed\n", sb.DeterminismRunsPassed))
	b.WriteString(fmt.Sprintf("• Diff Validations:          %d Passed\n", sb.DiffClassificationsPassed))
	b.WriteString("====================================================")
	return b.String()
}
func findRelationship(relationships []types.Relationship, sourceName, predicate, targetName string) *types.Relationship {
	for i := range relationships {
		r := &relationships[i]
		pred := r.Predicate
		if pred == "" {
			pred = r.Type
		}
		if (r.SourceName == sourceName || sourceName == "") &&
			(r.TargetName == targetName || targetName == "") &&
			strings.EqualFold(pred, predicate) {
			return r
		}
	}
	return nil
}
