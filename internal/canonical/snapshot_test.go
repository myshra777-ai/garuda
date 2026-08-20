// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package canonical

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

func TestCanonicalizeSnapshot_OrderIndependence(t *testing.T) {
	idA := uuid.New()
	idB := uuid.New()

	entity1 := types.Entity{
		ID:          idA,
		CanonicalID: "go-canonical-v1:struct:pkgA:User",
		Name:        "User",
		Kind:        types.EntityKindStruct,
		Fields:      []string{"id", "email", "created_at"},
	}
	entity2 := types.Entity{
		ID:          idB,
		CanonicalID: "go-canonical-v1:function:pkgB:NewUser",
		Name:        "NewUser",
		Kind:        types.EntityKindFunction,
	}

	rel1 := types.Relationship{
		SourceID:  idB,
		Predicate: types.PredicateCalls,
		TargetID:  idA,
	}

	// Snapshot with order [entity1, entity2]
	snap1 := types.Snapshot{
		CommitSHA:     "abcd1234efgh5678",
		Entities:      []types.Entity{entity1, entity2},
		Relationships: []types.Relationship{rel1},
	}

	// Snapshot with inverted order [entity2, entity1] and scrambled fields
	entity1Scrambled := entity1
	entity1Scrambled.Fields = []string{"created_at", "id", "email"}
	snap2 := types.Snapshot{
		CommitSHA:     "abcd1234efgh5678",
		Entities:      []types.Entity{entity2, entity1Scrambled},
		Relationships: []types.Relationship{rel1},
	}

	bytes1, hash1, err := CanonicalizeSnapshot(snap1)
	if err != nil {
		t.Fatalf("failed to canonicalize snap1: %v", err)
	}

	bytes2, hash2, err := CanonicalizeSnapshot(snap2)
	if err != nil {
		t.Fatalf("failed to canonicalize snap2: %v", err)
	}

	if hash1 != hash2 {
		t.Fatalf("snapshot hash mismatch under different input order: %s != %s", hash1, hash2)
	}

	if !bytes.Equal(bytes1, bytes2) {
		t.Fatal("canonical byte streams are not identical")
	}
}
