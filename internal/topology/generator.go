package topology

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/types"
)

// Generator creates execution topologies from natural language goals.
type Generator struct {
	store *store.PostgresStore
}

func NewGenerator(s *store.PostgresStore) *Generator {
	return &Generator{store: s}
}

// Recommend analyzes a goal and returns a proposed topology.
// Recommend analyzes a goal and returns a proposed topology.
func (g *Generator) Recommend(ctx context.Context, tenantID uuid.UUID, goal, scopeDomain, scopeSystem string, maxBudget int64) (*types.Topology, error) {
	// 1. Determine roles in logical order
	roles := determineRoles(goal) // returns [AUDITOR, ARCHITECT, ENGINEER] – but we want ARCHITECT first
	// Reorder: Architect first, then Auditor, then Engineer
	orderedRoles := []types.AgentRole{}
	for _, r := range roles {
		if r == types.RoleArchitect {
			orderedRoles = append(orderedRoles, r)
		}
	}
	for _, r := range roles {
		if r == types.RoleAuditor {
			orderedRoles = append(orderedRoles, r)
		}
	}
	for _, r := range roles {
		if r == types.RoleEngineer {
			orderedRoles = append(orderedRoles, r)
		}
	}
	if len(orderedRoles) == 0 {
		orderedRoles = []types.AgentRole{types.RoleArchitect, types.RoleAuditor, types.RoleEngineer}
	}

	// 2. Build topology with proper ID
	topology := &types.Topology{
		ID:             uuid.New(),
		TenantID:       tenantID,
		Goal:           goal,
		ScopeDomain:    scopeDomain,
		ScopeSystem:    scopeSystem,
		Status:         types.TopologyPending,
		MaxTokenBudget: maxBudget,
		TokensConsumed: 0,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	// 3. Generate tasks with correct fields
	tasks := []*types.Task{}
	tokenBudget := int64(5000) // starting budget per task
	for i, role := range orderedRoles {
		title, desc := getTaskTitleDesc(role, goal, scopeDomain, scopeSystem)
		task := &types.Task{
			ID:           uuid.New(),
			TenantID:     tenantID,
			TopologyID:   topology.ID,
			SequenceNo:   i + 1,
			Title:        title,
			Description:  desc,
			RequiredRole: role,
			Scope:        scopeDomain + ":" + scopeSystem,
			Status:       types.TaskPending,
			TokenBudget:  tokenBudget,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
			AssignedTo:   nil,
		}
		tokenBudget += 1000 // increment for next task
		tasks = append(tasks, task)
	}

	// 4. Set dependencies: each task depends on the previous one
	for i, task := range tasks {
		if i > 0 {
			task.DependsOn = []uuid.UUID{tasks[i-1].ID}
		}
	}

	topology.Tasks = tasks
	return topology, nil
}

func getTaskTitleDesc(role types.AgentRole, goal, scopeDomain, scopeSystem string) (string, string) {
	switch role {
	case types.RoleArchitect:
		return "Design solution architecture",
			fmt.Sprintf("Create a design proposal for '%s' in scope %s/%s", goal, scopeDomain, scopeSystem)
	case types.RoleAuditor:
		return "Security & policy pre-flight check",
			fmt.Sprintf("Verify design aligns with active policies in %s/%s", scopeDomain, scopeSystem)
	case types.RoleEngineer:
		return "Implement solution",
			fmt.Sprintf("Execute the approved design for '%s'", goal)
	default:
		return "Task for " + string(role), "Execute task"
	}
}

func determineRoles(goal string) []types.AgentRole {
	goalLower := strings.ToLower(goal)
	roles := []types.AgentRole{types.RoleAuditor} // always include auditor

	if strings.Contains(goalLower, "design") ||
		strings.Contains(goalLower, "architecture") ||
		strings.Contains(goalLower, "plan") {
		roles = append(roles, types.RoleArchitect)
	}
	if strings.Contains(goalLower, "code") ||
		strings.Contains(goalLower, "implement") ||
		strings.Contains(goalLower, "build") {
		roles = append(roles, types.RoleEngineer)
	}
	if len(roles) == 1 {
		roles = append(roles, types.RoleArchitect)
	}
	return roles
}

func generateTasks(goal, scopeDomain, scopeSystem string, roles []types.AgentRole) []*types.Task {
	var tasks []*types.Task
	for _, role := range roles {
		var title, desc string
		switch role {
		case types.RoleArchitect:
			title = "Design solution architecture"
			desc = fmt.Sprintf("Create a design proposal for '%s' in scope %s/%s", goal, scopeDomain, scopeSystem)
		case types.RoleAuditor:
			title = "Security & policy pre-flight check"
			desc = fmt.Sprintf("Verify design aligns with active policies in %s/%s", scopeDomain, scopeSystem)
		case types.RoleEngineer:
			title = "Implement solution"
			desc = fmt.Sprintf("Execute the approved design for '%s'", goal)
		default:
			title = "Task: " + string(role)
			desc = "Execute task for " + string(role)
		}
		tasks = append(tasks, &types.Task{
			Title:        title,
			Description:  desc,
			RequiredRole: role,
			Scope:        scopeDomain + ":" + scopeSystem,
		})
	}
	return tasks
}
