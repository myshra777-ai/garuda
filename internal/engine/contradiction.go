package engine

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/techtaytor/garuda/internal/types"
)

// ContradictionType defines the kind of conflict detected.
type ContradictionType int

const (
	DirectConflict ContradictionType = iota
	DependencyConflict
	ScopeOverlap
)

func (ct ContradictionType) String() string {
	return [...]string{
		"DirectConflict",
		"DependencyConflict",
		"ScopeOverlap",
	}[ct]
}

// ContradictionRecord represents a detected conflict.
type ContradictionRecord struct {
	Type        ContradictionType
	DecisionA   uuid.UUID
	DecisionB   uuid.UUID
	Explanation string
	DetectedAt  time.Time
	Resolved    bool
}

// ContradictionEngine is the main detection engine.
type ContradictionEngine struct {
	store types.DecisionStore
}

// NewContradictionEngine creates a new engine.
func NewContradictionEngine(store types.DecisionStore) *ContradictionEngine {
	return &ContradictionEngine{
		store: store,
	}
}

// ValidateDecision checks a decision against all canonical decisions.
func (e *ContradictionEngine) ValidateDecision(ctx context.Context, tenantID uuid.UUID, candidate *types.Decision) ([]ContradictionRecord, error) {
	// ... implementation will be added later
	return nil, nil
}
