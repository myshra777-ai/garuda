// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package analyzer

import (
	"testing"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

func TestPreLedgerValidator(t *testing.T) {
	validator := NewPreLedgerValidator()

	idA := uuid.New()
	idB := uuid.New()

	validEntityA := types.Entity{
		ID:          idA,
		CanonicalID: "go-canonical-v1:struct:pkgA:User",
		Name:        "User",
		Kind:        types.EntityKindStruct,
	}
	validEntityB := types.Entity{
		ID:          idB,
		CanonicalID: "go-canonical-v1:function:pkgA:NewUser",
		Name:        "NewUser",
		Kind:        types.EntityKindFunction,
	}

	validRel := types.Relationship{
		SourceID:         idB,
		Predicate:        types.PredicateCalls,
		TargetID:         idA,
		ResolutionStatus: types.ResolutionStatusResolved,
	}

	// 1. Valid snapshot passes
	t.Run("ValidSnapshotPasses", func(t *testing.T) {
		snap := types.Snapshot{
			Entities:      []types.Entity{validEntityA, validEntityB},
			Relationships: []types.Relationship{validRel},
		}
		if err := validator.Validate(snap); err != nil {
			t.Fatalf("expected valid snapshot to pass, got error: %v", err)
		}
	})

	// 2. Dangling resolved target fails
	t.Run("DanglingResolvedTargetFails", func(t *testing.T) {
		snap := types.Snapshot{
			Entities:      []types.Entity{validEntityB}, // validEntityA missing
			Relationships: []types.Relationship{validRel},
		}
		if err := validator.Validate(snap); err == nil {
			t.Fatal("expected validator to catch dangling relationship target")
		}
	})

	// 3. Canonical ID collision fails
	t.Run("CanonicalIDCollisionFails", func(t *testing.T) {
		collidingEntity := validEntityB
		collidingEntity.ID = uuid.New()
		collidingEntity.CanonicalID = validEntityA.CanonicalID // Duplicate canonical key

		snap := types.Snapshot{
			Entities: []types.Entity{validEntityA, collidingEntity},
		}
		if err := validator.Validate(snap); err == nil {
			t.Fatal("expected validator to catch canonical ID collision")
		}
	})
}
