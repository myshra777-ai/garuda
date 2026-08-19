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
	"github.com/myshra777-ai/garuda/internal/types"
)

// QuarantineDecision atomically updates decision status to 'quarantined' and logs the contradiction record.
func (s *PostgresStore) QuarantineDecision(ctx context.Context, tenantID uuid.UUID,
	proposedID, conflictingID uuid.UUID, domain, system, reason string) (*types.Contradiction, error) {

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start quarantine transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Atomically update decision status to quarantined
	updateQuery := `
		UPDATE decisions
		SET status = 'quarantined', updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2;
	`
	if _, err := tx.Exec(ctx, updateQuery, tenantID, proposedID); err != nil {
		return nil, fmt.Errorf("failed to set decision status to quarantined: %w", err)
	}

	// 2. Build contradiction record
	record := &types.Contradiction{
		ID:                 uuid.New(),
		TenantID:           tenantID,
		DecisionA:          conflictingID,
		DecisionB:          proposedID,
		Severity:           "medium",
		Quarantined:        true,
		Resolved:           false,
		ResolutionStrategy: "human",
		CreatedAt:          time.Now().UTC(),
	}

	insertQuery := `
		INSERT INTO contradictions (
			id, tenant_id, decision_a, decision_b, severity,
			quarantined, resolved, resolution_strategy, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`
	if _, err := tx.Exec(ctx, insertQuery,
		record.ID, record.TenantID, record.DecisionA, record.DecisionB,
		record.Severity, record.Quarantined, record.Resolved,
		record.ResolutionStrategy, record.CreatedAt); err != nil {
		return nil, fmt.Errorf("failed to insert contradiction record: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit quarantine transaction: %w", err)
	}

	return record, nil
}

// ListUnresolvedContradictions retrieves all unresolved quarantined contradictions.
func (s *PostgresStore) ListUnresolvedContradictions(ctx context.Context, tenantID uuid.UUID) ([]types.Contradiction, error) {
	query := `
		SELECT id, tenant_id, decision_a, decision_b, severity, quarantined, resolved,
		       resolution_strategy, created_at, resolved_at, auto_resolved_at
		FROM contradictions
		WHERE tenant_id = $1 AND resolved = false
		ORDER BY created_at ASC;
	`
	rows, err := s.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []types.Contradiction
	for rows.Next() {
		var c types.Contradiction
		var resolvedAt, autoResolvedAt *time.Time
		err := rows.Scan(
			&c.ID, &c.TenantID, &c.DecisionA, &c.DecisionB, &c.Severity,
			&c.Quarantined, &c.Resolved, &c.ResolutionStrategy,
			&c.CreatedAt, &resolvedAt, &autoResolvedAt,
		)
		if err != nil {
			return nil, err
		}
		if resolvedAt != nil {
			c.ResolvedAt = resolvedAt
		}
		if autoResolvedAt != nil {
			c.AutoResolvedAt = autoResolvedAt
		}
		results = append(results, c)
	}
	return results, nil
}

// ResolveContradiction marks a contradiction as resolved.
func (s *PostgresStore) ResolveContradiction(ctx context.Context, id uuid.UUID, strategy string) error {
	query := `
		UPDATE contradictions
		SET resolved = true,
		    resolution_strategy = $1,
		    resolved_at = NOW(),
		    auto_resolved_at = CASE WHEN $1 = 'auto_supersede' THEN NOW() ELSE NULL END
		WHERE id = $2;
	`
	_, err := s.pool.Exec(ctx, query, strategy, id)
	return err
}

// GetContradiction retrieves a single contradiction by ID for a tenant.
func (s *PostgresStore) GetContradiction(ctx context.Context, tenantID, id uuid.UUID) (*types.Contradiction, error) {
	query := `
		SELECT id, tenant_id, decision_a, decision_b, severity, quarantined, resolved,
		       resolution_strategy, created_at, resolved_at, auto_resolved_at
		FROM contradictions
		WHERE tenant_id = $1 AND id = $2;
	`
	var c types.Contradiction
	var resolvedAt, autoResolvedAt *time.Time

	err := s.pool.QueryRow(ctx, query, tenantID, id).Scan(
		&c.ID, &c.TenantID, &c.DecisionA, &c.DecisionB, &c.Severity,
		&c.Quarantined, &c.Resolved, &c.ResolutionStrategy,
		&c.CreatedAt, &resolvedAt, &autoResolvedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch contradiction: %w", err)
	}

	if resolvedAt != nil {
		c.ResolvedAt = resolvedAt
	}
	if autoResolvedAt != nil {
		c.AutoResolvedAt = autoResolvedAt
	}

	return &c, nil
}

// ListContradictions retrieves all contradictions for a tenant, filtered by resolution status.
func (s *PostgresStore) ListContradictions(ctx context.Context, tenantID uuid.UUID, resolved bool) ([]types.Contradiction, error) {
	query := `
		SELECT id, tenant_id, decision_a, decision_b, severity, quarantined, resolved,
		       resolution_strategy, created_at, resolved_at, auto_resolved_at
		FROM contradictions
		WHERE tenant_id = $1 AND resolved = $2
		ORDER BY created_at DESC;
	`
	rows, err := s.pool.Query(ctx, query, tenantID, resolved)
	if err != nil {
		return nil, fmt.Errorf("failed to list contradictions: %w", err)
	}
	defer rows.Close()

	var results []types.Contradiction
	for rows.Next() {
		var c types.Contradiction
		var resolvedAt, autoResolvedAt *time.Time
		err := rows.Scan(
			&c.ID, &c.TenantID, &c.DecisionA, &c.DecisionB, &c.Severity,
			&c.Quarantined, &c.Resolved, &c.ResolutionStrategy,
			&c.CreatedAt, &resolvedAt, &autoResolvedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contradiction row: %w", err)
		}
		if resolvedAt != nil {
			c.ResolvedAt = resolvedAt
		}
		if autoResolvedAt != nil {
			c.AutoResolvedAt = autoResolvedAt
		}
		results = append(results, c)
	}

	return results, nil
}
