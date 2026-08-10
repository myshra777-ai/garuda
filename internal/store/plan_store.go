package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// GetPlan assembles a plan from decisions, tasks, handoffs, milestones, and lineage.
func (s *PostgresStore) GetPlan(ctx context.Context, tenantID uuid.UUID, req *types.PlanRequest) (*types.PlanResult, error) {
	// 1. Build scope
	scope := types.Scope{
		Domain: req.ScopeDomain,
		System: req.ScopeSystem,
	}

	// 2. Fetch decisions active at the given time (or now)
	at := time.Now().UTC()
	if req.At != nil {
		at = *req.At
	}
	statuses := []types.DecisionStatus{types.StatusCanonical, types.StatusApproved}
	if len(req.Statuses) > 0 {
		statuses = parseStatuses(req.Statuses)
	}
	decisions, err := s.GetDecisionsActiveAt(ctx, tenantID, at, scope, statuses)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch decisions: %w", err)
	}

	// 3. Fetch tasks (in scope, optionally filtered by status)
	taskStatuses := []string{"in_progress", "paused", "pending"}
	if len(req.Statuses) > 0 {
		taskStatuses = req.Statuses
	}
	tasks, err := s.ListTasksByScope(ctx, tenantID, scope.Domain, scope.System, taskStatuses)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks: %w", err)
	}

	// 4. Fetch handoffs
	handoffs, err := s.ListHandoffsByScope(ctx, tenantID, scope.Domain, scope.System)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch handoffs: %w", err)
	}

	// 5. Fetch milestones
	milestones, err := s.ListMilestonesByScope(ctx, tenantID, scope.Domain, scope.System)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch milestones: %w", err)
	}

	// 6. Fetch lineage edges (dependencies) for tasks in scope
	var taskIDs []uuid.UUID
	for _, t := range tasks {
		taskIDs = append(taskIDs, t.ID)
	}
	dependencies, err := s.ListLineageEdgesByTasks(ctx, tenantID, taskIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lineage edges: %w", err)
	}

	return &types.PlanResult{
		TenantID:     tenantID,
		Scope:        scope,
		Decisions:    decisions,
		Tasks:        tasks,
		Handoffs:     handoffs,
		Milestones:   milestones,
		Dependencies: dependencies,
		GeneratedAt:  time.Now().UTC(),
	}, nil
}

// Helper: parse status strings to DecisionStatus slice.
func parseStatuses(statuses []string) []types.DecisionStatus {
	var ds []types.DecisionStatus
	for _, s := range statuses {
		switch s {
		case "draft":
			ds = append(ds, types.StatusDraft)
		case "review":
			ds = append(ds, types.StatusReview)
		case "approved":
			ds = append(ds, types.StatusApproved)
		case "canonical":
			ds = append(ds, types.StatusCanonical)
		case "superseded":
			ds = append(ds, types.StatusSuperseded)
		case "archived":
			ds = append(ds, types.StatusArchived)
		case "quarantined":
			ds = append(ds, types.StatusQuarantined)
		}
	}
	return ds
}
