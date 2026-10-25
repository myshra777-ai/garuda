// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package policy

import (
	"testing"

	"github.com/google/uuid"
)

func TestFilterEvidenceByDiffScope_PartitionsNetNewAndBaseline(t *testing.T) {
	touchedID := uuid.New()
	legacyFromID := uuid.New()
	legacyToID := uuid.New()
	newToID := uuid.New()

	ev := Evidence{
		ClaimIDs:  []uuid.UUID{uuid.New(), uuid.New()},
		EntityIDs: []uuid.UUID{legacyFromID, legacyToID, touchedID, newToID},
		ReasoningNotes: []string{
			"layer violation (CALLS): legacy_api.OldHandler -> store.SaveOrder",
			"layer violation (CALLS): checkout_api.NewHandler -> store.DirectSQL",
		},
	}

	scope := &DiffScopeFilter{
		TouchedEntityIDs: map[uuid.UUID]bool{touchedID: true},
		TouchedPackages:  map[string]bool{"checkout_api": true},
		OnlyNetNew:       true,
	}

	filtered, netNew, baseline := filterEvidenceByDiffScope("layer_violation", ev, scope)
	if netNew != 1 {
		t.Fatalf("expected 1 net-new violation, got %d", netNew)
	}
	if baseline != 1 {
		t.Fatalf("expected 1 baseline violation, got %d", baseline)
	}
	if len(filtered.ReasoningNotes) != 1 || filtered.ReasoningNotes[0] != "layer violation (CALLS): checkout_api.NewHandler -> store.DirectSQL" {
		t.Errorf("unexpected net-new notes: %+v", filtered.ReasoningNotes)
	}
	if len(filtered.BaselineNotes) != 1 || filtered.BaselineNotes[0] != "layer violation (CALLS): legacy_api.OldHandler -> store.SaveOrder" {
		t.Errorf("unexpected baseline notes: %+v", filtered.BaselineNotes)
	}
}

func TestFilterEvidenceByDiffScope_ZeroTouchedEntitiesSuppressesAllBaseline(t *testing.T) {
	legacyFromID := uuid.New()
	legacyToID := uuid.New()

	ev := Evidence{
		ClaimIDs:  []uuid.UUID{uuid.New()},
		EntityIDs: []uuid.UUID{legacyFromID, legacyToID},
		ReasoningNotes: []string{
			"layer violation (CALLS): legacy_api.OldHandler -> store.SaveOrder",
		},
	}

	// Simulate a documentation-only commit (e.g., README.md, PLAYBOOK.md) where
	// OnlyNetNew is true and 0 AST symbols/packages were touched.
	scope := &DiffScopeFilter{
		TouchedEntityIDs: map[uuid.UUID]bool{},
		TouchedPackages:  map[string]bool{},
		OnlyNetNew:       true,
	}

	filtered, netNew, baseline := filterEvidenceByDiffScope("layer_violation", ev, scope)
	if netNew != 0 {
		t.Fatalf("expected 0 net-new violations for doc-only commit, got %d", netNew)
	}
	if baseline != 1 {
		t.Fatalf("expected 1 baseline violation for doc-only commit, got %d", baseline)
	}
	if len(filtered.ReasoningNotes) != 0 {
		t.Errorf("expected 0 net-new reasoning notes, got %+v", filtered.ReasoningNotes)
	}
	if len(filtered.BaselineNotes) != 1 {
		t.Errorf("expected 1 baseline reasoning note, got %+v", filtered.BaselineNotes)
	}
}
