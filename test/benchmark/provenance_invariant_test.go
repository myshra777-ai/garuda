package benchmark

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myshra777-ai/garuda/internal/analyzer"
)

// TestProvenance_SourceMutationAltersEvidenceHash verifies the core trust invariant:
// Source Mutation -> Evidence Hash Change -> Artifact Fingerprint Change[cite: 1, 2].
func TestProvenance_SourceMutationAltersEvidenceHash(t *testing.T) {
	tempDir := t.TempDir()
	sourceFile := filepath.Join(tempDir, "service.go")

	initialCode := `package service

func ProcessPayment(amount int64) bool {
	return amount > 0
}
`
	require.NoError(t, os.WriteFile(sourceFile, []byte(initialCode), 0644))

	// 1. Initial Analysis Run
	req1 := analyzer.AnalysisRequest{
		Path:      tempDir,
		CommitSHA: "commit-v1",
		Options:   analyzer.AnalysisOptions{TypeResolution: true},
	}
	res1, err := analyzer.NewGoAnalyzer().Analyze(context.Background(), req1)
	require.NoError(t, err)
	require.Len(t, res1.Entities, 1)

	entityV1 := res1.Entities[0]
	assert.Equal(t, "ProcessPayment", entityV1.Name)
	assert.NotEmpty(t, entityV1.EvidenceHash)
	assert.Equal(t, 3, entityV1.LineStart)
	assert.Equal(t, 5, entityV1.LineEnd)

	// 2. Mutate exactly 1 character in source code (amount > 0 -> amount >= 0)
	mutatedCode := `package service

func ProcessPayment(amount int64) bool {
	return amount >= 0
}
`
	require.NoError(t, os.WriteFile(sourceFile, []byte(mutatedCode), 0644))

	// 3. Second Analysis Run
	req2 := analyzer.AnalysisRequest{
		Path:      tempDir,
		CommitSHA: "commit-v2",
		Options:   analyzer.AnalysisOptions{TypeResolution: true},
	}
	res2, err := analyzer.NewGoAnalyzer().Analyze(context.Background(), req2)
	require.NoError(t, err)
	require.Len(t, res2.Entities, 1)

	entityV2 := res2.Entities[0]

	// 4. Assert Invariants
	// Invariant A: Canonical identity remains stable across minor edits[cite: 1, 2]
	assert.Equal(t, entityV1.CanonicalID, entityV2.CanonicalID,
		"Canonical UUIDv5 identity must remain unchanged when symbol name and signature are preserved[cite: 1, 2]")

	// Invariant B: Evidence hash MUST change because the underlying source bytes changed[cite: 1, 2]
	assert.NotEqual(t, entityV1.EvidenceHash, entityV2.EvidenceHash,
		"CRITICAL: Evidence hash failed to change following source mutation[cite: 1, 2]")

	// Invariant C: Analysis payload fingerprint changes[cite: 1, 2]
	assert.NotEqual(t, res1.Fingerprint, res2.Fingerprint,
		"Semantic artifact fingerprint must change when evidence changes[cite: 1, 2]")
}

// TestProvenance_DeterministicReplayIdentity verifies that re-running analysis
// on identical commits produces bit-for-bit identical outputs (100% Deterministic Repeatability)[cite: 1, 2].
func TestProvenance_DeterministicReplayIdentity(t *testing.T) {
	fixtureDir := filepath.Join("truth_fixtures", "001-basic")

	req := analyzer.AnalysisRequest{
		Path:      fixtureDir,
		CommitSHA: "static-commit-001",
		Options:   analyzer.AnalysisOptions{TypeResolution: true},
	}

	runner := analyzer.NewGoAnalyzer()

	run1, err := runner.Analyze(context.Background(), req)
	require.NoError(t, err)

	run2, err := runner.Analyze(context.Background(), req)
	require.NoError(t, err)

	// Verify Identical Artifact Fingerprint & Entity Counts
	assert.Equal(t, run1.Fingerprint, run2.Fingerprint, "Replay analysis produced conflicting artifact fingerprints")
	assert.Equal(t, len(run1.Entities), len(run2.Entities), "Replay analysis produced conflicting entity count")
	assert.Equal(t, len(run1.Relationships), len(run2.Relationships), "Replay analysis produced conflicting edge count")

	for i := range run1.Entities {
		assert.Equal(t, run1.Entities[i].CanonicalID, run2.Entities[i].CanonicalID)
		assert.Equal(t, run1.Entities[i].EvidenceHash, run2.Entities[i].EvidenceHash)
	}
}
