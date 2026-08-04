package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// GetDecisionsActiveAt returns decisions that were valid at a specific point in time.
func (s *PostgresStore) GetDecisionsActiveAt(ctx context.Context, tenantID uuid.UUID, at time.Time, scope types.Scope, statuses []types.DecisionStatus) ([]*types.Decision, error) {
	statusStrs := make([]string, len(statuses))
	for i, st := range statuses {
		statusStrs[i] = string(st)
	}

	query := `
		SELECT id, title, status, scope, owner, confidence,
		       created_at, updated_at, approved_at, valid_from, valid_to
		FROM decisions
		WHERE tenant_id = $1
		  AND valid_from <= $2
		  AND (valid_to IS NULL OR valid_to >= $2)
		  AND ($3 = '' OR scope->>'domain' = $3)
		  AND ($4 = '' OR scope->>'system' = $4)
		  AND (cardinality($5::text[]) = 0 OR status = ANY($5))
		ORDER BY valid_from ASC
	`
	rows, err := s.pool.Query(ctx, query, tenantID, at, scope.Domain, scope.System, statusStrs)
	if err != nil {
		return nil, fmt.Errorf("failed to query active decisions: %w", err)
	}
	defer rows.Close()

	var results []*types.Decision
	for rows.Next() {
		var d types.Decision
		var statusStr string
		var scopeJSON []byte
		var validTo *time.Time
		err := rows.Scan(
			&d.ID, &d.Title, &statusStr, &scopeJSON, &d.Owner, &d.Confidence,
			&d.CreatedAt, &d.UpdatedAt, &d.ApprovedAt, &d.ValidFrom, &validTo,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan decision row: %w", err)
		}
		d.TenantID = tenantID
		d.Status = types.DecisionStatus(statusStr)
		if validTo != nil {
			d.ValidTo = validTo
		}
		if len(scopeJSON) > 0 {
			_ = json.Unmarshal(scopeJSON, &d.Scope)
		}
		results = append(results, &d)
	}
	return results, nil
}

// GetDecisionHistory returns all versions of a decision with temporal validity.
func (s *PostgresStore) GetDecisionHistory(ctx context.Context, tenantID, decisionID uuid.UUID) ([]*types.Decision, error) {
	query := `
		SELECT id, title, status, scope, owner, confidence,
		       created_at, updated_at, approved_at, valid_from, valid_to
		FROM decisions
		WHERE tenant_id = $1 AND id = $2
		ORDER BY valid_from ASC
	`
	rows, err := s.pool.Query(ctx, query, tenantID, decisionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query decision history: %w", err)
	}
	defer rows.Close()

	var results []*types.Decision
	for rows.Next() {
		var d types.Decision
		var statusStr string
		var scopeJSON []byte
		var validTo *time.Time
		err := rows.Scan(
			&d.ID, &d.Title, &statusStr, &scopeJSON, &d.Owner, &d.Confidence,
			&d.CreatedAt, &d.UpdatedAt, &d.ApprovedAt, &d.ValidFrom, &validTo,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan decision row: %w", err)
		}
		d.TenantID = tenantID
		d.Status = types.DecisionStatus(statusStr)
		if validTo != nil {
			d.ValidTo = validTo
		}
		if len(scopeJSON) > 0 {
			_ = json.Unmarshal(scopeJSON, &d.Scope)
		}
		results = append(results, &d)
	}
	return results, nil
}
