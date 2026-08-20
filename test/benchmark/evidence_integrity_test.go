// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package benchmark

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/types"
)

// TestEvidenceIntegrity_DirectObservationAnchors verifies Invariant 1-04:
// Every direct observation entity and relationship must carry a valid file path, line range, and evidence hash.
func TestEvidenceIntegrity_DirectObservationAnchors(t *testing.T) {
	ctx := context.Background()
	goAnalyzer := analyzer.NewGoAnalyzer()

	casePath := "../../garuda-bench/corpus/cases/001-basic"
	if _, err := os.Stat(casePath); os.IsNotExist(err) {
		casePath = "garuda-bench/corpus/cases/001-basic"
	}

	snap, err := goAnalyzer.Analyze(ctx, analyzer.AnalysisRequest{
		Path:      casePath,
		CommitSHA: "test-evidence-sha",
		Options: analyzer.AnalysisOptions{
			IncludeCallGraph: true,
			TypeResolution:   true,
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, snap.Entities)

	for _, e := range snap.Entities {
		assert.NotEmpty(t, e.FilePath, "Entity %s missing FilePath", e.QualifiedName)
		assert.Greater(t, e.LineStart, 0, "Entity %s LineStart must be positive", e.QualifiedName)
		assert.GreaterOrEqual(t, e.LineEnd, e.LineStart, "Entity %s LineEnd must be >= LineStart", e.QualifiedName)
		assert.NotEmpty(t, e.EvidenceHash, "Entity %s missing EvidenceHash", e.QualifiedName)
		assert.NotEmpty(t, e.ContentSnippet, "Entity %s missing ContentSnippet", e.QualifiedName)

		// Verify cryptographic anchor: EvidenceHash == SHA256(ContentSnippet)
		expectedHashBytes := sha256.Sum256([]byte(e.ContentSnippet))
		expectedHash := hex.EncodeToString(expectedHashBytes[:])
		assert.Equal(t, expectedHash, e.EvidenceHash, "Cryptographic hash mismatch for entity snippet %s", e.QualifiedName)
	}

	for _, r := range snap.Relationships {
		if r.EpistemicClass == types.EpistemicClassObservation {
			assert.NotEmpty(t, r.Predicate, "Observation relationship missing Predicate")
		}
	}
}

// TestEvidenceIntegrity_TamperDetection verifies Invariant 1-05 & P0-08:
// Modifying a single byte in the underlying source file alters the computed evidence hash[cite: 1].
func TestEvidenceIntegrity_TamperDetection(t *testing.T) {
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "service.go")

	originalCode := `package service

type AuthService struct {
	Timeout int
}

func (a *AuthService) Authenticate(token string) bool {
	return len(token) > 0
}
`
	require.NoError(t, os.WriteFile(sourceFile, []byte(originalCode), 0644))

	ctx := context.Background()
	goAnalyzer := analyzer.NewGoAnalyzer()

	// Analyze original artifact
	snap1, err := goAnalyzer.Analyze(ctx, analyzer.AnalysisRequest{
		Path:      tmpDir,
		CommitSHA: "v1-original",
		Options:   analyzer.AnalysisOptions{TypeResolution: true},
	})
	require.NoError(t, err)

	authMethod1 := findEntityByQualifiedName(snap1.Entities, "example.com/service.(*AuthService).Authenticate")
	require.NotNil(t, authMethod1)
	originalEvidenceHash := authMethod1.EvidenceHash

	// Tamper: alter whitespace/return statement
	tamperedCode := `package service

type AuthService struct {
	Timeout int
}

func (a *AuthService) Authenticate(token string) bool {
	return len(token) > 8 // Tampered logic
}
`
	require.NoError(t, os.WriteFile(sourceFile, []byte(tamperedCode), 0644))

	// Analyze tampered artifact
	snap2, err := goAnalyzer.Analyze(ctx, analyzer.AnalysisRequest{
		Path:      tmpDir,
		CommitSHA: "v2-tampered",
		Options:   analyzer.AnalysisOptions{TypeResolution: true},
	})
	require.NoError(t, err)

	authMethod2 := findEntityByQualifiedName(snap2.Entities, "example.com/service.(*AuthService).Authenticate")
	require.NotNil(t, authMethod2)
	tamperedEvidenceHash := authMethod2.EvidenceHash

	// Invariant 1-05: Source modification must alter evidence hash
	assert.NotEqual(t, originalEvidenceHash, tamperedEvidenceHash,
		"Tampered source code must produce a distinct evidence hash")

	// Invariant: Canonical ID remains stable across body mutations if signature is unchanged
	assert.Equal(t, authMethod1.CanonicalID, authMethod2.CanonicalID,
		"Canonical identity must remain stable across implementation changes")
}
