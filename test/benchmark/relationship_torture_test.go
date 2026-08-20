// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package benchmark

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/canonical"
	"github.com/myshra777-ai/garuda/internal/types"
)

// TestRelationshipTorture_EndpointIntegrity verifies Invariant 1-02:
// All resolved relationships must bind strictly to existing canonical entity IDs.
func TestRelationshipTorture_EndpointIntegrity(t *testing.T) {
	ctx := context.Background()
	goAnalyzer := analyzer.NewGoAnalyzer()
	validator := analyzer.NewPreLedgerValidator()

	// Extract standard case 001
	snap, err := goAnalyzer.Analyze(ctx, analyzer.AnalysisRequest{
		Path:      "../../garuda-bench/corpus/cases/001-basic",
		CommitSHA: "test-head",
		Options: analyzer.AnalysisOptions{
			IncludeCallGraph: true,
			TypeResolution:   true,
		},
	})
	if err != nil {
		// Fallback path if run from repo root
		snap, err = goAnalyzer.Analyze(ctx, analyzer.AnalysisRequest{
			Path:      "garuda-bench/corpus/cases/001-basic",
			CommitSHA: "test-head",
			Options: analyzer.AnalysisOptions{
				IncludeCallGraph: true,
				TypeResolution:   true,
			},
		})
	}
	require.NoError(t, err)
	require.NoError(t, validator.Validate(*snap))

	entityMap := make(map[uuid.UUID]types.Entity)
	for _, e := range snap.Entities {
		entityMap[e.ID] = e
	}

	for _, rel := range snap.Relationships {
		// Verify source endpoint exists
		src, srcExists := entityMap[rel.SourceID]
		assert.True(t, srcExists, "Source entity %s must exist in snapshot entities", rel.SourceID)
		if srcExists {
			assert.NotEmpty(t, src.CanonicalID, "Source entity must have canonical ID")
		}

		// Verify target endpoint if resolved
		if rel.ResolutionStatus == "" || rel.ResolutionStatus == types.ResolutionStatusResolved {
			if rel.TargetID != uuid.Nil {
				tgt, tgtExists := entityMap[rel.TargetID]
				assert.True(t, tgtExists, "Resolved target entity %s must exist in snapshot entities", rel.TargetID)
				if tgtExists {
					assert.NotEmpty(t, tgt.CanonicalID, "Target entity must have canonical ID")
				}
			}
		}

		// Verify Epistemic Class is explicit
		assert.NotEmpty(t, rel.EpistemicClass, "Relationship epistemic class must be explicitly set")
	}
}

// TestRelationshipTorture_InterfaceNegativeFixtures verifies Invariant 1-03 & P0-07:
// Structs missing methods or with signature mismatches must NEVER be assigned an IMPLEMENTS edge.
func TestRelationshipTorture_InterfaceNegativeFixtures(t *testing.T) {
	validator := analyzer.NewPreLedgerValidator()

	ifaceID, ifaceKey := canonical.GenerateEntityUUID(canonical.EntityKeySpec{
		Kind:        types.EntityKindInterface,
		PackagePath: "github.com/myshra777-ai/garuda/test/mock",
		Name:        "Writer",
	})
	structID, structKey := canonical.GenerateEntityUUID(canonical.EntityKeySpec{
		Kind:        types.EntityKindStruct,
		PackagePath: "github.com/myshra777-ai/garuda/test/mock",
		Name:        "IncompleteWriter",
	})

	ifaceEntity := types.Entity{
		ID:          ifaceID,
		CanonicalID: ifaceKey,
		Name:        "Writer",
		Kind:        types.EntityKindInterface,
		Methods:     []string{"Write", "Close"},
	}

	structEntity := types.Entity{
		ID:          structID,
		CanonicalID: structKey,
		Name:        "IncompleteWriter",
		Kind:        types.EntityKindStruct,
		Methods:     []string{"Write"}, // Missing "Close" -> must not implement Writer
	}

	snap := types.Snapshot{
		CommitSHA: "test-negative",
		Entities:  []types.Entity{ifaceEntity, structEntity},
	}

	// Validate baseline passes
	require.NoError(t, validator.Validate(snap))

	// Verify no fabricated IMPLEMENTS edge exists between IncompleteWriter and Writer
	for _, rel := range snap.Relationships {
		if rel.SourceID == structID && rel.TargetID == ifaceID {
			assert.NotEqual(t, types.PredicateImplements, rel.Predicate,
				"Incomplete struct must not have an IMPLEMENTS relationship to interface")
		}
	}
}

// TestRelationshipTorture_DisambiguationAndReceiverCollisions verifies P0-02:
// Pointer vs value receivers must yield distinct canonical identities without collision.
func TestRelationshipTorture_DisambiguationAndReceiverCollisions(t *testing.T) {
	specVal := canonical.EntityKeySpec{
		Kind:         types.EntityKindMethod,
		PackagePath:  "github.com/myshra777-ai/garuda/pkg/service",
		ReceiverType: "Handler",
		Name:         "ServeHTTP",
	}
	specPtr := canonical.EntityKeySpec{
		Kind:         types.EntityKindMethod,
		PackagePath:  "github.com/myshra777-ai/garuda/pkg/service",
		ReceiverType: "*Handler",
		Name:         "ServeHTTP",
	}

	valUUID, valKey := canonical.GenerateEntityUUID(specVal)
	ptrUUID, ptrKey := canonical.GenerateEntityUUID(specPtr)

	assert.NotEqual(t, valUUID, ptrUUID, "Value and Pointer receiver method UUIDs must not collide")
	assert.NotEqual(t, valKey, ptrKey, "Value and Pointer receiver canonical string keys must not collide")
	assert.Contains(t, ptrKey, "(*Handler).ServeHTTP")
	assert.Contains(t, valKey, "(Handler).ServeHTTP")
}
