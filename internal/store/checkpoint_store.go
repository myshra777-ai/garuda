// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/myshra777-ai/garuda/internal/types"
)

// GetCheckpoint retrieves a specific checkpoint snapshot by tenant ID and checkpoint ID.
func (s *PostgresStore) GetCheckpoint(ctx context.Context, tenantID, id uuid.UUID) (*types.Checkpoint, error) {
	query := `
		SELECT id, tenant_id, agent_id, checkpoint_name, task_id, checkpoint_data, status,
		       created_at, updated_at, expires_at
		FROM agent_checkpoints
		WHERE tenant_id = $1 AND id = $2;
	`

	var c types.Checkpoint
	var taskID *uuid.UUID
	var expiresAt *time.Time

	err := s.pool.QueryRow(ctx, query, tenantID, id).Scan(
		&c.ID,
		&c.TenantID,
		&c.AgentID,
		&c.CheckpointName,
		&taskID,
		&c.CheckpointData,
		&c.Status,
		&c.CreatedAt,
		&c.UpdatedAt,
		&expiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("checkpoint not found (tenant: %s, id: %s): %w", tenantID, id, err)
		}
		return nil, fmt.Errorf("failed to query checkpoint: %w", err)
	}

	c.TaskID = taskID
	if expiresAt != nil {
		c.ExpiresAt = expiresAt
	}

	return &c, nil
}

// SaveCheckpoint inserts a new checkpoint record or updates existing state on primary key or (tenant_id, checkpoint_name) collision.
func (s *PostgresStore) SaveCheckpoint(ctx context.Context, c *types.Checkpoint) error {
	query := `
		INSERT INTO agent_checkpoints (
			id, tenant_id, agent_id, checkpoint_name, task_id, checkpoint_data, status, created_at, updated_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (tenant_id, checkpoint_name) DO UPDATE SET
			agent_id = EXCLUDED.agent_id,
			task_id = EXCLUDED.task_id,
			checkpoint_data = EXCLUDED.checkpoint_data,
			status = EXCLUDED.status,
			updated_at = NOW(),
			expires_at = EXCLUDED.expires_at;
	`
	_, err := s.pool.Exec(ctx, query,
		c.ID, c.TenantID, c.AgentID, c.CheckpointName, c.TaskID, c.CheckpointData,
		c.Status, c.CreatedAt, c.UpdatedAt, c.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save checkpoint: %w", err)
	}
	return nil
}

// ListCheckpointsByAgent retrieves recent checkpoints for a specific agent in reverse chronological order.
func (s *PostgresStore) ListCheckpointsByAgent(ctx context.Context, tenantID uuid.UUID, agentID string, limit int) ([]*types.Checkpoint, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	query := `
		SELECT id, tenant_id, agent_id, checkpoint_name, task_id, checkpoint_data, status,
		       created_at, updated_at, expires_at
		FROM agent_checkpoints
		WHERE tenant_id = $1 AND agent_id = $2
		ORDER BY created_at DESC
		LIMIT $3;
	`
	rows, err := s.pool.Query(ctx, query, tenantID, agentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list checkpoints: %w", err)
	}
	defer rows.Close()

	var results []*types.Checkpoint
	for rows.Next() {
		var c types.Checkpoint
		var taskID *uuid.UUID
		var expiresAt *time.Time
		if err := rows.Scan(
			&c.ID, &c.TenantID, &c.AgentID, &c.CheckpointName, &taskID, &c.CheckpointData,
			&c.Status, &c.CreatedAt, &c.UpdatedAt, &expiresAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan checkpoint row: %w", err)
		}
		c.TaskID = taskID
		if expiresAt != nil {
			c.ExpiresAt = expiresAt
		}
		results = append(results, &c)
	}
	return results, nil
}
