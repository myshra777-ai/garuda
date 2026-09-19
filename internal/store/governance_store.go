// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RevisionChainEntry struct {
	ID                   uuid.UUID
	RevisionNumber       int
	DecisionHash         []byte
	PreviousRevisionHash []byte
	CreatedAt            time.Time
}

type DecisionExplanation struct {
	ID                   uuid.UUID
	Title                string
	Statement            string
	ScopeJSON            []byte
	Owner                string
	Confidence           float64
	Status               string
	CreatedAt            time.Time
	RevisionNumber       int
	CanonicalJSON        []byte
	DecisionHash         []byte
	PreviousRevisionHash []byte
	RevisionCreatedAt    time.Time
	MerkleRoot           []byte
}

// GetRevisionChain retrieves the revision chain for one decision.
//
// The chain is per-decision: revision N's previous_revision_hash
// equals revision N-1's decision_hash, and revision 1's
// previous_revision_hash is the zero hash (its predecessor is
// genesis). The previous version of this function took only tenantID
// and returned every revision for the tenant in a flat list ordered
// by created_at. The verify command walked that flat list assuming a
// single chain; the first time a decision boundary was crossed it
// reported a false break.
//
// The chain returned is ordered by revision_number ASC so the caller
// can compare each entry to its predecessor without re-sorting.
func (s *PostgresStore) GetRevisionChain(ctx context.Context, tenantID, decisionID uuid.UUID) ([]RevisionChainEntry, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, revision_number, decision_hash, previous_revision_hash, created_at
		FROM decision_revisions
		WHERE tenant_id = $1 AND decision_id = $2
		ORDER BY revision_number ASC
	`, tenantID, decisionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query revision chain: %w", err)
	}
	defer rows.Close()

	var chain []RevisionChainEntry
	for rows.Next() {
		var entry RevisionChainEntry
		if err := rows.Scan(
			&entry.ID,
			&entry.RevisionNumber,
			&entry.DecisionHash,
			&entry.PreviousRevisionHash,
			&entry.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan revision entry: %w", err)
		}
		chain = append(chain, entry)
	}
	return chain, rows.Err()
}

// ListDecisionsWithRevisions returns every decision ID in a tenant
// that has at least one revision. Used by the verify command to
// iterate over every decision chain.
func (s *PostgresStore) ListDecisionsWithRevisions(ctx context.Context, tenantID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT decision_id
		  FROM decision_revisions
		 WHERE tenant_id = $1
		 ORDER BY decision_id
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list decisions with revisions: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan decision id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// GetDecisionExplanation retrieves the full decision, latest revision, and merkle root.
func (s *PostgresStore) GetDecisionExplanation(ctx context.Context, tenantID, decisionID uuid.UUID) (*DecisionExplanation, error) {
	var exp DecisionExplanation
	exp.ID = decisionID

	err := s.pool.QueryRow(ctx, `
		SELECT title, statement, scope, owner, confidence, status, created_at
		FROM decisions
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, decisionID).Scan(
		&exp.Title, &exp.Statement, &exp.ScopeJSON, &exp.Owner,
		&exp.Confidence, &exp.Status, &exp.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("decision not found: %w", err)
	}

	_ = s.pool.QueryRow(ctx, `
		SELECT revision_number, canonical_json, decision_hash, previous_revision_hash, created_at
		FROM decision_revisions
		WHERE tenant_id = $1 AND decision_id = $2
		ORDER BY revision_number DESC LIMIT 1
	`, tenantID, decisionID).Scan(
		&exp.RevisionNumber, &exp.CanonicalJSON, &exp.DecisionHash,
		&exp.PreviousRevisionHash, &exp.RevisionCreatedAt,
	)

	_ = s.pool.QueryRow(ctx, `SELECT root_hash FROM merkle_roots WHERE tenant_id = $1`, tenantID).Scan(&exp.MerkleRoot)

	return &exp, nil
}
