package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// ListMilestonesByScope returns milestones linked to tasks in a given scope.
func (s *PostgresStore) ListMilestonesByScope(ctx context.Context, tenantID uuid.UUID, domain, system string) ([]*types.Milestone, error) {
	query := `
		SELECT m.id, m.tenant_id, m.task_id, m.title, m.description, m.status,
		       m.due_date, m.completed_at, m.created_at, m.updated_at
		FROM milestones m
		JOIN tasks t ON m.task_id = t.id
		WHERE m.tenant_id = $1 AND t.scope_domain = $2 AND t.scope_system = $3
		ORDER BY m.due_date ASC NULLS LAST
	`
	rows, err := s.pool.Query(ctx, query, tenantID, domain, system)
	if err != nil {
		return nil, fmt.Errorf("failed to list milestones: %w", err)
	}
	defer rows.Close()

	var milestones []*types.Milestone
	for rows.Next() {
		var m types.Milestone
		var taskID *uuid.UUID
		var dueDate, completedAt *time.Time

		err := rows.Scan(&m.ID, &m.TenantID, &taskID, &m.Title, &m.Description, &m.Status,
			&dueDate, &completedAt, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan milestone: %w", err)
		}
		m.TaskID = taskID
		if dueDate != nil {
			m.DueDate = dueDate
		}
		if completedAt != nil {
			m.CompletedAt = completedAt
		}
		milestones = append(milestones, &m)
	}
	return milestones, nil
}

// SaveMilestone inserts or updates a milestone.
func (s *PostgresStore) SaveMilestone(ctx context.Context, m *types.Milestone) error {
	query := `
		INSERT INTO milestones (id, tenant_id, task_id, title, description, status, due_date, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			task_id = EXCLUDED.task_id,
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			status = EXCLUDED.status,
			due_date = EXCLUDED.due_date,
			completed_at = EXCLUDED.completed_at,
			updated_at = NOW()
	`
	_, err := s.pool.Exec(ctx, query,
		m.ID, m.TenantID, m.TaskID, m.Title, m.Description, m.Status, m.DueDate, m.CompletedAt,
	)
	return err
}
