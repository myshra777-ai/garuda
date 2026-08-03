package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/myshra777-ai/garuda/internal/types"
)

type ContradictionEngine struct {
	store types.DecisionStore
}

func NewContradictionEngine(store types.DecisionStore) *ContradictionEngine {
	return &ContradictionEngine{store: store}
}

// DetectAndQuarantine evaluates rules and quarantines new contradictory decisions.
func (e *ContradictionEngine) DetectAndQuarantine(ctx context.Context,
	newDecision *types.Decision) (*types.Contradiction, error) {

	// Fetch active scope decisions
	existing, err := e.store.GetDecisionsByScope(ctx, newDecision.TenantID,
		newDecision.Scope.Domain, newDecision.Scope.System)
	if err != nil {
		return nil, err
	}

	var conflicting *types.Decision
	for _, d := range existing {
		// SKIP self AND skip decisions that are already quarantined or superseded
		if d.ID == newDecision.ID || string(d.Status) == "quarantined" || string(d.Status) == "superseded" {
			continue
		}

		if isContradictory(newDecision, d) {
			conflicting = d
			break
		}
	}

	if conflicting == nil {
		return nil, nil // No contradiction found against active decisions
	}

	reason := fmt.Sprintf("Proposed decision '%s' contradicts active decision '%s' (ID: %s) in scope %s/%s",
		newDecision.Title, conflicting.Title, conflicting.ID.String(),
		newDecision.Scope.Domain, newDecision.Scope.System)

	// Atomically quarantine decision and write record
	record, err := e.store.QuarantineDecision(ctx, newDecision.TenantID,
		newDecision.ID, conflicting.ID,
		newDecision.Scope.Domain, newDecision.Scope.System, reason)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func isContradictory(a, b *types.Decision) bool {
	if a.Title == b.Title {
		return true
	}

	aLower := strings.ToLower(a.Title)
	bLower := strings.ToLower(b.Title)

	if a.Scope.System == "db" || a.Scope.System == "database" {
		if (strings.Contains(aLower, "postgres") && strings.Contains(bLower, "mongodb")) ||
			(strings.Contains(aLower, "mongodb") && strings.Contains(bLower, "postgres")) {
			return true
		}
	}
	return false
}
