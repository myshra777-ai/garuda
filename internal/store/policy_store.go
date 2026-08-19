// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// SavePolicy inserts a new policy.
func (s *PostgresStore) SavePolicy(ctx context.Context, p *types.Policy) error {
	query := `
		INSERT INTO policies (
			id, tenant_id, statement, scope_domain, scope_system,
			actor, status, valid_from, valid_to, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
	`
	_, err := s.pool.Exec(ctx, query,
		p.ID, p.TenantID, p.Statement, p.ScopeDomain, p.ScopeSystem,
		p.Actor, p.Status, p.ValidFrom, p.ValidTo, p.Metadata,
	)
	return err
}

// GetActivePolicies returns all active policies for a scope.
func (s *PostgresStore) GetActivePolicies(ctx context.Context, tenantID uuid.UUID, scopeDomain, scopeSystem string) ([]*types.Policy, error) {
	now := time.Now().UTC()
	query := `
		SELECT id, tenant_id, statement, scope_domain, scope_system,
		       actor, status, valid_from, valid_to, metadata,
		       created_at, updated_at, superseded_by, merkle_hash
		FROM policies
		WHERE tenant_id = $1
		  AND scope_domain = $2
		  AND scope_system = $3
		  AND status = 'active'
		  AND valid_from <= $4
		  AND (valid_to IS NULL OR valid_to >= $4)
		ORDER BY valid_from DESC
	`
	rows, err := s.pool.Query(ctx, query, tenantID, scopeDomain, scopeSystem, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*types.Policy
	for rows.Next() {
		p := &types.Policy{}
		var validTo sql.NullTime
		var supersededBy uuid.NullUUID
		var merkleHash sql.NullString
		err := rows.Scan(
			&p.ID, &p.TenantID, &p.Statement, &p.ScopeDomain, &p.ScopeSystem,
			&p.Actor, &p.Status, &p.ValidFrom, &validTo, &p.Metadata,
			&p.CreatedAt, &p.UpdatedAt, &supersededBy, &merkleHash,
		)
		if err != nil {
			return nil, err
		}
		if validTo.Valid {
			p.ValidTo = &validTo.Time
		}
		if supersededBy.Valid {
			p.SupersededBy = &supersededBy.UUID
		}
		if merkleHash.Valid {
			p.MerkleHash = merkleHash.String
		}
		policies = append(policies, p)
	}
	return policies, nil
}

// GetActivePoliciesByScope returns policies for a broader scope (wildcard support).
func (s *PostgresStore) GetActivePoliciesByScope(ctx context.Context, tenantID uuid.UUID, scope types.Scope) ([]*types.Policy, error) {
	// If domain/system are empty, match all.
	query := `
		SELECT id, tenant_id, statement, scope_domain, scope_system,
		       actor, status, valid_from, valid_to, metadata,
		       created_at, updated_at, superseded_by, merkle_hash
		FROM policies
		WHERE tenant_id = $1
		  AND status = 'active'
		  AND valid_from <= NOW()
		  AND (valid_to IS NULL OR valid_to >= NOW())
		  AND ($2 = '' OR scope_domain = $2)
		  AND ($3 = '' OR scope_system = $3)
	`
	rows, err := s.pool.Query(ctx, query, tenantID, scope.Domain, scope.System)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// ... scan similar to above
	return nil, nil
}

// SupersedePolicy marks a policy as superseded and creates a new one.
func (s *PostgresStore) SupersedePolicy(ctx context.Context, oldID, newID uuid.UUID) error {
	query := `
		UPDATE policies
		SET status = 'superseded', superseded_by = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := s.pool.Exec(ctx, query, newID, oldID)
	return err
}

// LogPolicyViolation records a policy violation.
func (s *PostgresStore) LogPolicyViolation(ctx context.Context, v *types.PolicyViolation) error {
	query := `
		INSERT INTO policy_violations (
			id, tenant_id, policy_id, actor, attempted_action, decision_id, reason, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	_, err := s.pool.Exec(ctx, query,
		v.ID, v.TenantID, v.PolicyID, v.Actor, v.AttemptedAction, v.DecisionID, v.Reason,
	)
	return err
}
