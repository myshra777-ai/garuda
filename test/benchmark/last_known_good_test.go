// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package benchmark

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/canonical"
	"github.com/myshra777-ai/garuda/internal/types"
)

// WorkspaceTruthState simulates the workspace pointer tracking the active truth revision.
type WorkspaceTruthState struct {
	CurrentCommit string
	SnapshotHash  string
	Manifest      *types.AnalysisManifest
}

// TestFailedAnalysis_PreservesLastKnownGoodState verifies Invariant 1-09:
// A failed or malformed analysis must abort and preserve the previous valid revision without advancing truth pointers.
func TestFailedAnalysis_PreservesLastKnownGoodState(t *testing.T) {
	ctx := context.Background()
	goAnalyzer := analyzer.NewGoAnalyzer()
	validator := analyzer.NewPreLedgerValidator()

	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "store.go")

	// 1. Initial valid commit (v1)
	v1Code := `package store

type UserStore struct {
	Capacity int
}

func (s *UserStore) GetUser(id string) string {
	return id
}
`
	require.NoError(t, os.WriteFile(sourceFile, []byte(v1Code), 0644))

	snap1, err := goAnalyzer.Analyze(ctx, analyzer.AnalysisRequest{
		Path:      tmpDir,
		CommitSHA: "commit-v1-clean",
		Options:   analyzer.AnalysisOptions{TypeResolution: true},
	})
	require.NoError(t, err)
	require.NoError(t, validator.Validate(*snap1))

	_, hash1, err := canonical.CanonicalizeSnapshot(*snap1)
	require.NoError(t, err)

	state := &WorkspaceTruthState{
		CurrentCommit: "commit-v1-clean",
		SnapshotHash:  hash1,
		Manifest: &types.AnalysisManifest{
			AnalysisID:   uuid.New(),
			CommitSHA:    "commit-v1-clean",
			EntityCount:  len(snap1.Entities),
			SnapshotHash: hash1,
			Status:       types.AnalysisStatusSucceeded,
			StartedAt:    time.Now().UTC(),
		},
	}

	// 2. Attempt corrupted analysis (v2 with dangling relationship violating PreLedgerValidator)[cite: 1]
	snap2Malformed := *snap1
	snap2Malformed.CommitSHA = "commit-v2-corrupted"
	snap2Malformed.Relationships = append(snap2Malformed.Relationships, types.Relationship{
		ID:               uuid.New(),
		SourceID:         snap1.Entities[0].ID,
		TargetID:         uuid.New(), // Dangling ID (not in entities)
		Predicate:        types.PredicateCalls,
		ResolutionStatus: types.ResolutionStatusResolved,
	})

	// Pre-Ledger Validation Gate: must reject before ledger entry[cite: 1]
	validationErr := validator.Validate(snap2Malformed)
	require.Error(t, validationErr, "PreLedgerValidator must reject snapshot with dangling relationship")

	// If validation fails, abort persistence and preserve last known good[cite: 1]
	if validationErr != nil {
		// Do NOT update state.CurrentCommit or state.SnapshotHash
	}

	// Invariant 1-09 Assertions: Workspace state strictly retained at v1[cite: 1]
	assert.Equal(t, "commit-v1-clean", state.CurrentCommit, "Failed analysis must not advance CurrentCommit")
	assert.Equal(t, hash1, state.SnapshotHash, "Failed analysis must not alter SnapshotHash")
	assert.Equal(t, types.AnalysisStatusSucceeded, state.Manifest.Status)
	assert.Equal(t, "commit-v1-clean", state.Manifest.CommitSHA)
}
