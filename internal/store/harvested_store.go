// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/harvester"
)

// SaveHarvestedDecision persists or updates a harvested candidate record.
// SaveHarvestedDecision persists a harvested decision.
func (s *PostgresStore) SaveHarvestedDecision(ctx context.Context, hd *harvester.HarvestedDecision) error {
	query := `
		INSERT INTO harvested_decisions (
			id, tenant_id, source_type, source_id, source_url, raw_text,
			extracted_decision, confidence, human_validated, decision_id,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := s.pool.Exec(ctx, query,
		hd.ID,
		hd.TenantID, // ✅ SaveHarvestedDecision Uses TenantID
		hd.SourceType,
		hd.SourceID,
		hd.SourceURL,
		hd.RawText,
		hd.ExtractedDecision,
		hd.Confidence,
		hd.HumanValidated,
		hd.DecisionID,
		hd.CreatedAt,
		hd.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save harvested decision: %w", err)
	}
	return nil
}

// GetHarvestedDecision retrieves a harvested decision candidate by UUID.
func (s *PostgresStore) GetHarvestedDecision(ctx context.Context, id uuid.UUID) (*harvester.HarvestedDecision, error) {
	query := `
		SELECT id, tenant_id, source_type, source_id, source_url, raw_text,
		       extracted_decision, confidence, human_validated, decision_id,
		       created_at, updated_at
		FROM harvested_decisions
		WHERE id = $1
	`
	var hd harvester.HarvestedDecision

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&hd.ID, &hd.TenantID, &hd.SourceType, &hd.SourceID, &hd.SourceURL, &hd.RawText,
		&hd.ExtractedDecision, &hd.Confidence, &hd.HumanValidated, &hd.DecisionID,
		&hd.CreatedAt, &hd.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get harvested decision: %w", err)
	}
	return &hd, nil
}

// ListHarvestedDecisions returns harvested decisions filtered by source and validation state.
func (s *PostgresStore) ListHarvestedDecisions(ctx context.Context, sourceType string, validatedOnly bool) ([]*harvester.HarvestedDecision, error) {
	query := `
		SELECT id, tenant_id, source_type, source_id, source_url, raw_text,
		       extracted_decision, confidence, human_validated, decision_id,
		       created_at, updated_at
		FROM harvested_decisions
		WHERE ($1 = '' OR source_type = $1)
		  AND ($2 = false OR human_validated = true)
		ORDER BY created_at DESC
	`
	rows, err := s.pool.Query(ctx, query, sourceType, validatedOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to list harvested decisions: %w", err)
	}
	defer rows.Close()

	var results []*harvester.HarvestedDecision
	for rows.Next() {
		var hd harvester.HarvestedDecision
		err := rows.Scan(
			&hd.ID, &hd.TenantID, &hd.SourceType, &hd.SourceID, &hd.SourceURL, &hd.RawText,
			&hd.ExtractedDecision, &hd.Confidence, &hd.HumanValidated, &hd.DecisionID,
			&hd.CreatedAt, &hd.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan harvested decision: %w", err)
		}
		results = append(results, &hd)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during harvested decisions row iteration: %w", err)
	}

	return results, nil
}

// SaveDecision persists a promoted canonical decision.
