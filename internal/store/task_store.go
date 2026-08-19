// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// ListTasksByScope returns tasks filtered by scope and status.
func (s *PostgresStore) ListTasksByScope(ctx context.Context, tenantID uuid.UUID, domain, system string, statuses []string) ([]*types.Task, error) {
	query := `
		SELECT id, tenant_id, title, description, status, priority,
		       owner_agent_id, parent_task_id, scope_domain, scope_system,
		       version, created_at, updated_at, completed_at
		FROM tasks
		WHERE tenant_id = $1
		  AND scope_domain = $2
		  AND scope_system = $3
		  AND (cardinality($4::text[]) = 0 OR status = ANY($4))
		ORDER BY priority DESC, created_at ASC
	`
	rows, err := s.pool.Query(ctx, query, tenantID, domain, system, statuses)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*types.Task
	for rows.Next() {
		t := &types.Task{}
		err := rows.Scan(
			&t.ID, &t.TenantID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.OwnerAgentID, &t.ParentTaskID, &t.ScopeDomain, &t.ScopeSystem,
			&t.Version, &t.CreatedAt, &t.UpdatedAt, &t.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// GetTask retrieves a single task by ID.
func (s *PostgresStore) GetTask(ctx context.Context, tenantID, taskID uuid.UUID) (*types.Task, error) {
	query := `
		SELECT id, tenant_id, title, description, status, priority,
		       owner_agent_id, parent_task_id, scope_domain, scope_system,
		       version, created_at, updated_at, completed_at
		FROM tasks
		WHERE tenant_id = $1 AND id = $2
	`
	var t types.Task
	err := s.pool.QueryRow(ctx, query, tenantID, taskID).Scan(
		&t.ID, &t.TenantID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.OwnerAgentID, &t.ParentTaskID, &t.ScopeDomain, &t.ScopeSystem,
		&t.Version, &t.CreatedAt, &t.UpdatedAt, &t.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &t, nil
}

// SaveTask inserts or updates a task.
func (s *PostgresStore) SaveTask(ctx context.Context, t *types.Task) error {
	query := `
		INSERT INTO tasks (id, tenant_id, title, description, status, priority,
		                   owner_agent_id, parent_task_id, scope_domain, scope_system,
		                   version, created_at, updated_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			status = EXCLUDED.status,
			priority = EXCLUDED.priority,
			owner_agent_id = EXCLUDED.owner_agent_id,
			parent_task_id = EXCLUDED.parent_task_id,
			scope_domain = EXCLUDED.scope_domain,
			scope_system = EXCLUDED.scope_system,
			version = tasks.version + 1,
			updated_at = NOW(),
			completed_at = EXCLUDED.completed_at
	`
	_, err := s.pool.Exec(ctx, query,
		t.ID, t.TenantID, t.Title, t.Description, t.Status, t.Priority,
		t.OwnerAgentID, t.ParentTaskID, t.ScopeDomain, t.ScopeSystem,
		t.Version, t.CreatedAt, t.UpdatedAt, t.CompletedAt,
	)
	return err
}
