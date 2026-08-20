// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package canonical

import (
	"testing"

	"github.com/myshra777-ai/garuda/internal/types"
)

func TestCanonicalEntityIdentity(t *testing.T) {
	// Invariant: Same symbol name in distinct packages must produce different IDs
	specA := EntityKeySpec{
		Kind:        types.EntityKindFunction,
		PackagePath: "github.com/myshra777-ai/garuda/internal/store",
		Name:        "NewStore",
	}
	specB := EntityKeySpec{
		Kind:        types.EntityKindFunction,
		PackagePath: "github.com/myshra777-ai/garuda/internal/engine",
		Name:        "NewStore",
	}

	idA, keyA := GenerateEntityUUID(specA)
	idB, keyB := GenerateEntityUUID(specB)

	if idA == idB || keyA == keyB {
		t.Fatalf("collision detected across distinct packages: %s == %s", keyA, keyB)
	}

	// Invariant: Determinism (identical spec must yield identical UUID and key)
	idA2, keyA2 := GenerateEntityUUID(specA)
	if idA != idA2 || keyA != keyA2 {
		t.Fatalf("determinism broken: %s != %s", keyA, keyA2)
	}

	// Invariant: Pointer vs Value receiver distinction
	specVal := EntityKeySpec{
		Kind:         types.EntityKindMethod,
		PackagePath:  "github.com/myshra777-ai/garuda/internal/engine",
		ReceiverType: "LineageEngine",
		Name:         "Evaluate",
	}
	specPtr := EntityKeySpec{
		Kind:         types.EntityKindMethod,
		PackagePath:  "github.com/myshra777-ai/garuda/internal/engine",
		ReceiverType: "*LineageEngine",
		Name:         "Evaluate",
	}

	idVal, _ := GenerateEntityUUID(specVal)
	idPtr, _ := GenerateEntityUUID(specPtr)

	if idVal == idPtr {
		t.Fatal("pointer and value receiver methods must not produce identical UUIDs")
	}
}
