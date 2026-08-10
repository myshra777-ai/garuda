package topology

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// Executor runs a topology sequentially with pre-flight checks.
type Executor struct {
	store  TopologyStore
	shield Shield
}

// Shield defines the pre-flight validation contract.
type Shield interface {
	ValidateTask(ctx context.Context, task *types.Task, topology *types.Topology) (bool, string, error)
}

// TopologyStore defines the persistence methods needed by the executor.
type TopologyStore interface {
	GetTopology(ctx context.Context, id uuid.UUID) (*types.Topology, error)
	GetTasksByTopology(ctx context.Context, topologyID uuid.UUID) ([]*types.Task, error)
	UpdateTask(ctx context.Context, task *types.Task) error
	UpdateTopologyStatus(ctx context.Context, id uuid.UUID, status types.TopologyStatus) error
	UpdateTopologyTokens(ctx context.Context, id uuid.UUID, tokens int64) error
}

// NewExecutor creates a new executor.
func NewExecutor(store TopologyStore, shield Shield) *Executor {
	return &Executor{
		store:  store,
		shield: shield,
	}
}

// Execute runs all tasks in the topology in order.
func (e *Executor) Execute(ctx context.Context, topologyID uuid.UUID) error {
	// 1. Fetch topology and tasks
	topology, err := e.store.GetTopology(ctx, topologyID)
	if err != nil {
		return fmt.Errorf("failed to get topology: %w", err)
	}
	if topology.Status != types.TopologyPending && topology.Status != types.TopologyPaused {
		return fmt.Errorf("topology %s is not executable (status: %s)", topologyID, topology.Status)
	}

	// Update status to ACTIVE
	if err := e.store.UpdateTopologyStatus(ctx, topologyID, types.TopologyActive); err != nil {
		return fmt.Errorf("failed to update topology status: %w", err)
	}

	// 2. Get tasks
	tasks, err := e.store.GetTasksByTopology(ctx, topologyID)
	if err != nil {
		return err
	}

	// 3. Run tasks sequentially
	for _, task := range tasks {
		// Skip already completed tasks
		if task.Status == types.TaskCompleted {
			continue
		}

		// Check dependencies: ensure they are completed
		for _, depID := range task.DependsOn {
			dep, err := e.getTaskByID(ctx, tasks, depID)
			if err != nil {
				return err
			}
			if dep.Status != types.TaskCompleted {
				return fmt.Errorf("dependency %s not completed", depID)
			}
		}

		// Pre-flight check
		allowed, reason, err := e.shield.ValidateTask(ctx, task, topology)
		if err != nil {
			return err
		}
		if !allowed {
			// Quarantine task
			task.Status = types.TaskQuarantined
			_ = e.store.UpdateTask(ctx, task)
			_ = e.store.UpdateTopologyStatus(ctx, topologyID, types.TopologyPaused)
			return fmt.Errorf("task quarantined: %s", reason)
		}

		// Execute task (mock execution for now)
		if err := e.executeTask(ctx, task, topology); err != nil {
			task.Status = types.TaskFailed
			_ = e.store.UpdateTask(ctx, task)
			_ = e.store.UpdateTopologyStatus(ctx, topologyID, types.TopologyPaused)
			return fmt.Errorf("task execution failed: %w", err)
		}

		// Mark as completed
		now := time.Now().UTC()
		task.Status = types.TaskCompleted
		task.CompletedAt = &now
		if err := e.store.UpdateTask(ctx, task); err != nil {
			return err
		}

		// Update topology token consumption
		if err := e.store.UpdateTopologyTokens(ctx, topologyID, task.TokensUsed); err != nil {
			return err
		}
	}

	// Mark topology as completed
	if err := e.store.UpdateTopologyStatus(ctx, topologyID, types.TopologyCompleted); err != nil {
		return err
	}
	return nil
}

func (e *Executor) getTaskByID(ctx context.Context, tasks []*types.Task, id uuid.UUID) (*types.Task, error) {
	for _, t := range tasks {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf("task %s not found", id)
}

func (e *Executor) executeTask(ctx context.Context, task *types.Task, topology *types.Topology) error {
	// In MVP, simulate execution by consuming some tokens.
	// In production, invoke an agent (e.g., via MCP or provider API).
	// For now, we'll consume half the budget.
	task.TokensUsed = task.TokenBudget / 2
	if task.TokensUsed == 0 {
		task.TokensUsed = 10
	}
	// Simulate work time
	time.Sleep(500 * time.Millisecond)
	return nil
}
