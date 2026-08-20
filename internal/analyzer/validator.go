// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package analyzer

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// ValidationError captures invariant violations detected in a semantic snapshot.
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("pre-ledger validation violation on %s: %s", e.Field, e.Msg)
}

// PreLedgerValidator enforces graph integrity, identity uniqueness, and endpoint existence.
type PreLedgerValidator struct{}

// NewPreLedgerValidator creates a validator instance.
func NewPreLedgerValidator() *PreLedgerValidator {
	return &PreLedgerValidator{}
}

// Validate executes read-only validation against a snapshot before database persistence.
func (v *PreLedgerValidator) Validate(snap types.Snapshot) error {
	entitySet := make(map[uuid.UUID]bool, len(snap.Entities))
	canonicalIDSet := make(map[string]uuid.UUID, len(snap.Entities))

	// 1. Entity Invariants (Section 9)
	for _, e := range snap.Entities {
		if e.ID == uuid.Nil {
			return &ValidationError{Field: "entity.id", Msg: "entity ID cannot be nil"}
		}
		if entitySet[e.ID] {
			return &ValidationError{Field: "entity.id", Msg: fmt.Sprintf("duplicate entity UUID detected: %s", e.ID)}
		}
		entitySet[e.ID] = true

		if e.CanonicalID == "" {
			return &ValidationError{Field: "entity.canonical_id", Msg: fmt.Sprintf("entity %s missing canonical_id", e.ID)}
		}
		if existing, exists := canonicalIDSet[e.CanonicalID]; exists {
			return &ValidationError{Field: "entity.canonical_id", Msg: fmt.Sprintf("canonical_id collision (%s) between entities %s and %s", e.CanonicalID, existing, e.ID)}
		}
		canonicalIDSet[e.CanonicalID] = e.ID

		if e.Kind == "" {
			return &ValidationError{Field: "entity.kind", Msg: fmt.Sprintf("entity %s has empty kind", e.ID)}
		}
	}

	// 2. Relationship Invariants (Section 9)
	for _, r := range snap.Relationships {
		if r.SourceID == uuid.Nil {
			return &ValidationError{Field: "relationship.source_id", Msg: "relationship source ID cannot be nil"}
		}
		if !entitySet[r.SourceID] {
			return &ValidationError{Field: "relationship.source_id", Msg: fmt.Sprintf("source entity %s does not exist in snapshot", r.SourceID)}
		}

		// Resolved relationships MUST point to a valid target entity
		isResolved := r.ResolutionStatus == "" || r.ResolutionStatus == types.ResolutionStatusResolved
		if isResolved {
			if r.TargetID == uuid.Nil {
				return &ValidationError{Field: "relationship.target_id", Msg: fmt.Sprintf("resolved relationship from %s has nil target ID", r.SourceID)}
			}
			if !entitySet[r.TargetID] {
				return &ValidationError{Field: "relationship.target_id", Msg: fmt.Sprintf("resolved relationship target entity %s does not exist in snapshot", r.TargetID)}
			}
		}

		// Predicate/Type check
		if r.Predicate == "" && r.Type == "" {
			return &ValidationError{Field: "relationship.predicate", Msg: fmt.Sprintf("relationship from %s missing predicate or type", r.SourceID)}
		}
	}

	return nil
}
