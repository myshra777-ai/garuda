package engine

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// LineageEngine manages decision lineage.
type LineageEngine struct {
	store types.DecisionStore
}

// NewLineageEngine creates a new lineage engine.
func NewLineageEngine(store types.DecisionStore) *LineageEngine {
	return &LineageEngine{
		store: store,
	}
}

// RegisterDecision validates that a decision exists and is safe to track.
func (e *LineageEngine) RegisterDecision(ctx context.Context, tenantID, decisionID uuid.UUID) error {
	if e == nil || e.store == nil {
		return fmt.Errorf("lineage engine store is not initialized")
	}
	if decisionID == uuid.Nil {
		return fmt.Errorf("decision id is required")
	}
	if _, err := e.store.GetDecision(ctx, tenantID, decisionID); err != nil {
		return fmt.Errorf("failed to validate decision %s: %w", decisionID, err)
	}
	return nil
}

// GetDecisionLineage returns the full lineage of a decision.
func (e *LineageEngine) GetDecisionLineage(ctx context.Context, tenantID, decisionID uuid.UUID) (*DecisionLineage, error) {
	if e == nil || e.store == nil {
		return nil, fmt.Errorf("lineage engine store is not initialized")
	}
	if decisionID == uuid.Nil {
		return nil, fmt.Errorf("decision id is required")
	}

	decision, err := e.store.GetDecision(ctx, tenantID, decisionID)
	if err != nil {
		return nil, err
	}
	if decision == nil {
		return nil, fmt.Errorf("decision %s not found", decisionID)
	}

	lineage := &DecisionLineage{
		DecisionID: decisionID,
	}

	if decision.ParentID != nil {
		parent, err := e.store.GetDecision(ctx, tenantID, *decision.ParentID)
		if err == nil && parent != nil {
			lineage.Parent = parent
			lineage.Dependencies = []*types.Decision{parent}
		}
	}

	children, err := e.store.ListDecisionsByParent(ctx, tenantID, decisionID)
	if err == nil {
		lineage.Children = children
		lineage.Dependents = children
	}

	lineage.SupersedingChain = e.buildSupersedingChain(ctx, tenantID, decision)
	lineage.ImpactSet, err = e.GetImpactSet(ctx, tenantID, decisionID)
	if err != nil {
		lineage.ImpactSet = nil
	}

	return lineage, nil
}

func (e *LineageEngine) buildSupersedingChain(ctx context.Context, tenantID uuid.UUID, decision *types.Decision) []uuid.UUID {
	if decision == nil {
		return nil
	}
	chain := []uuid.UUID{decision.ID}
	current := decision
	for current != nil && current.ParentID != nil {
		parent, err := e.store.GetDecision(ctx, tenantID, *current.ParentID)
		if err != nil || parent == nil {
			break
		}
		chain = append(chain, parent.ID)
		current = parent
	}
	return chain
}

// GetImpactSet returns all decisions that depend on the given decision (directly or indirectly).
func (e *LineageEngine) GetImpactSet(ctx context.Context, tenantID, decisionID uuid.UUID) ([]uuid.UUID, error) {
	if e == nil || e.store == nil {
		return nil, fmt.Errorf("lineage engine store is not initialized")
	}
	if decisionID == uuid.Nil {
		return nil, fmt.Errorf("decision id is required")
	}

	var result []uuid.UUID
	visited := make(map[uuid.UUID]bool)

	var visit func(id uuid.UUID) error
	visit = func(id uuid.UUID) error {
		visited[id] = true
		children, err := e.store.ListDecisionsByParent(ctx, tenantID, id)
		if err != nil {
			return err
		}
		for _, child := range children {
			if child == nil || child.ID == uuid.Nil {
				continue
			}
			if !visited[child.ID] {
				result = append(result, child.ID)
				if err := visit(child.ID); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if err := visit(decisionID); err != nil {
		return nil, err
	}

	filtered := []uuid.UUID{}
	for _, id := range result {
		if id != decisionID {
			filtered = append(filtered, id)
		}
	}
	return filtered, nil
}

// DecisionLineage represents the lineage of a decision.
type DecisionLineage struct {
	DecisionID       uuid.UUID
	Parent           *types.Decision
	Children         []*types.Decision
	Dependencies     []*types.Decision
	Dependents       []*types.Decision
	SupersedingChain []uuid.UUID
	ImpactSet        []uuid.UUID
}
