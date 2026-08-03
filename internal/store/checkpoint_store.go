package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// SaveCheckpoint inserts or updates an agent checkpoint.
func (s *PostgresStore) SaveCheckpoint(ctx context.Context, c *types.Checkpoint) error {
	query := `
		INSERT INTO agent_checkpoints (
			id, tenant_id, agent_id, task_id, checkpoint_data, status, created_at, updated_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (tenant_id, id) DO UPDATE SET
			agent_id = EXCLUDED.agent_id,
			task_id = EXCLUDED.task_id,
			checkpoint_data = EXCLUDED.checkpoint_data,
			status = EXCLUDED.status,
			updated_at = NOW(),
			expires_at = EXCLUDED.expires_at;
	`
	_, err := s.pool.Exec(ctx, query,
		c.ID, c.TenantID, c.AgentID, c.TaskID, c.CheckpointData,
		c.Status, c.CreatedAt, c.UpdatedAt, c.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save checkpoint: %w", err)
	}
	return nil
}

// GetCheckpoint retrieves a checkpoint by ID (tenant-scoped).
func (s *PostgresStore) GetCheckpoint(ctx context.Context, tenantID, id uuid.UUID) (*types.Checkpoint, error) {
	var c types.Checkpoint
	var taskID *uuid.UUID
	var expiresAt *time.Time

	query := `
		SELECT id, tenant_id, agent_id, task_id, checkpoint_data, status,
		       created_at, updated_at, expires_at
		FROM agent_checkpoints
		WHERE tenant_id = $1 AND id = $2;
	`
	err := s.pool.QueryRow(ctx, query, tenantID, id).Scan(
		&c.ID, &c.TenantID, &c.AgentID, &taskID, &c.CheckpointData,
		&c.Status, &c.CreatedAt, &c.UpdatedAt, &expiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint: %w", err)
	}
	c.TaskID = taskID
	if expiresAt != nil {
		c.ExpiresAt = expiresAt
	}
	return &c, nil
}

// ListCheckpointsByAgent retrieves active checkpoints for an agent.
func (s *PostgresStore) ListCheckpointsByAgent(ctx context.Context, tenantID uuid.UUID, agentID string, limit int) ([]*types.Checkpoint, error) {
	query := `
		SELECT id, tenant_id, agent_id, task_id, checkpoint_data, status,
		       created_at, updated_at, expires_at
		FROM agent_checkpoints
		WHERE tenant_id = $1 AND agent_id = $2 AND status = 'active'
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
			&c.ID, &c.TenantID, &c.AgentID, &taskID, &c.CheckpointData,
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
