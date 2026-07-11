package engine

import (
	"context"

	"github.com/google/uuid"
	"github.com/techtaytor/garuda/internal/types"
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

// RegisterDecision registers a decision in the lineage graph.
func (e *LineageEngine) RegisterDecision(ctx context.Context, tenantID, decisionID uuid.UUID) error {
	// ... implementation will be added later
	return nil
}
