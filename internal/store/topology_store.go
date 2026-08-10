package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// SaveTopology persists a topology and its tasks.
func (s *PostgresStore) SaveTopology(ctx context.Context, top *types.Topology) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Insert topology
	topQuery := `
		INSERT INTO topologies (
			id, tenant_id, goal, scope_domain, scope_system, status,
			max_token_budget, tokens_consumed, merkle_root, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = tx.Exec(ctx, topQuery,
		top.ID, top.TenantID, top.Goal, top.ScopeDomain, top.ScopeSystem,
		string(top.Status), top.MaxTokenBudget, top.TokensConsumed, top.MerkleRoot,
		top.CreatedAt, top.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert topology: %w", err)
	}

	// Insert tasks
	for _, task := range top.Tasks {
		taskQuery := `
			INSERT INTO topology_tasks (
				id, topology_id, sequence_no, title, description,
				required_role, assigned_to, scope, status, depends_on,
				token_budget, tokens_used, created_at, completed_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		`
		dependsOnJSON, _ := json.Marshal(task.DependsOn)
		var assignedTo *uuid.UUID
		if task.AssignedTo != nil && *task.AssignedTo != uuid.Nil {
			assignedTo = task.AssignedTo
		}
		var completedAt *time.Time
		if task.CompletedAt != nil {
			completedAt = task.CompletedAt
		}
		_, err = tx.Exec(ctx, taskQuery,
			task.ID, task.TopologyID, task.SequenceNo, task.Title, task.Description,
			string(task.RequiredRole), assignedTo, task.Scope, string(task.Status),
			dependsOnJSON, task.TokenBudget, task.TokensUsed, task.CreatedAt, completedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert task %s: %w", task.ID, err)
		}
	}

	return tx.Commit(ctx)
}

// GetTopology retrieves a topology by ID.
func (s *PostgresStore) GetTopology(ctx context.Context, id uuid.UUID) (*types.Topology, error) {
	query := `
		SELECT id, tenant_id, goal, scope_domain, scope_system, status,
		       max_token_budget, tokens_consumed, merkle_root,
		       created_at, updated_at, completed_at
		FROM topologies
		WHERE id = $1
	`
	var top types.Topology
	var completedAt *time.Time
	var statusStr string
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&top.ID, &top.TenantID, &top.Goal, &top.ScopeDomain, &top.ScopeSystem,
		&statusStr, &top.MaxTokenBudget, &top.TokensConsumed, &top.MerkleRoot,
		&top.CreatedAt, &top.UpdatedAt, &completedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get topology: %w", err)
	}
	top.Status = types.TopologyStatus(statusStr)
	if completedAt != nil {
		top.CompletedAt = completedAt
	}
	return &top, nil
}

// UpdateTopologyStatus updates the status of a topology.
func (s *PostgresStore) UpdateTopologyStatus(ctx context.Context, id uuid.UUID, status types.TopologyStatus) error {
	query := `
		UPDATE topologies
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := s.pool.Exec(ctx, query, string(status), id)
	return err
}

// GetTasksByTopology retrieves all tasks for a topology.
func (s *PostgresStore) GetTasksByTopology(ctx context.Context, topologyID uuid.UUID) ([]*types.Task, error) {
	query := `
		SELECT id, topology_id, sequence_no, title, description,
		       required_role, assigned_to, scope, status, depends_on,
		       token_budget, tokens_used, created_at, completed_at
		FROM topology_tasks
		WHERE topology_id = $1
		ORDER BY sequence_no ASC
	`
	rows, err := s.pool.Query(ctx, query, topologyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*types.Task
	for rows.Next() {
		var t types.Task
		var assignedTo *uuid.UUID
		var completedAt *time.Time
		var dependsOnJSON []byte
		var requiredRoleStr string
		var statusStr string

		err := rows.Scan(
			&t.ID, &t.TopologyID, &t.SequenceNo, &t.Title, &t.Description,
			&requiredRoleStr, &assignedTo, &t.Scope, &statusStr, &dependsOnJSON,
			&t.TokenBudget, &t.TokensUsed, &t.CreatedAt, &completedAt,
		)
		if err != nil {
			return nil, err
		}
		t.RequiredRole = types.AgentRole(requiredRoleStr)
		t.Status = types.TaskStatus(statusStr)
		t.AssignedTo = assignedTo
		if completedAt != nil {
			t.CompletedAt = completedAt
		}
		if len(dependsOnJSON) > 0 {
			if err := json.Unmarshal(dependsOnJSON, &t.DependsOn); err != nil {
				return nil, err
			}
		}
		tasks = append(tasks, &t)
	}
	return tasks, nil
}

// UpdateTask updates a single task.
func (s *PostgresStore) UpdateTask(ctx context.Context, task *types.Task) error {
	var assignedTo *uuid.UUID
	if task.AssignedTo != nil && *task.AssignedTo != uuid.Nil {
		assignedTo = task.AssignedTo
	}
	query := `
		UPDATE topology_tasks
		SET status = $1, assigned_to = $2, tokens_used = $3, completed_at = $4, updated_at = NOW()
		WHERE id = $5
	`
	_, err := s.pool.Exec(ctx, query,
		string(task.Status), assignedTo, task.TokensUsed, task.CompletedAt, task.ID,
	)
	return err
}

// UpdateTopologyTokens updates the token consumption for a topology.
func (s *PostgresStore) UpdateTopologyTokens(ctx context.Context, id uuid.UUID, tokens int64) error {
	query := `
		UPDATE topologies
		SET tokens_consumed = tokens_consumed + $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := s.pool.Exec(ctx, query, tokens, id)
	return err
}
