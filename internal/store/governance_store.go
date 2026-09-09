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
	DecisionHash         []byte
	PreviousRevisionHash []byte
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

// GetRevisionChain retrieves all decision revisions for chain verification.
func (s *PostgresStore) GetRevisionChain(ctx context.Context, tenantID uuid.UUID) ([]RevisionChainEntry, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, decision_hash, previous_revision_hash
		FROM decision_revisions
		WHERE tenant_id = $1
		ORDER BY created_at ASC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query revision chain: %w", err)
	}
	defer rows.Close()

	var chain []RevisionChainEntry
	for rows.Next() {
		var entry RevisionChainEntry
		if err := rows.Scan(&entry.ID, &entry.DecisionHash, &entry.PreviousRevisionHash); err != nil {
			return nil, fmt.Errorf("failed to scan revision entry: %w", err)
		}
		chain = append(chain, entry)
	}
	return chain, rows.Err()
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
