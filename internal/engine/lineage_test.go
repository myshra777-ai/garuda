package engine

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

type stubDecisionStore struct {
	decisions map[uuid.UUID]*types.Decision
}

func (s *stubDecisionStore) GetDecision(ctx context.Context, tenantID, decisionID uuid.UUID) (*types.Decision, error) {
	return s.decisions[decisionID], nil
}

func (s *stubDecisionStore) SaveDecision(ctx context.Context, d *types.Decision) error {
	return nil
}

func (s *stubDecisionStore) GetDecisionsByScope(ctx context.Context, tenantID uuid.UUID, domain, system string) ([]*types.Decision, error) {
	return nil, nil
}

func (s *stubDecisionStore) GetDecisionRevisions(ctx context.Context, tenantID, decisionID uuid.UUID) ([]types.DecisionRevision, error) {
	return nil, nil
}

func (s *stubDecisionStore) ListDecisions(ctx context.Context, tenantID uuid.UUID, scope types.Scope, statuses []types.DecisionStatus) ([]*types.Decision, error) {
	return nil, nil
}

func (s *stubDecisionStore) IngestEvidence(ctx context.Context, tenantID uuid.UUID, evidence []types.Evidence) error {
	return nil
}

func (s *stubDecisionStore) ConsumeBudget(ctx context.Context, tenantID uuid.UUID, tokens int) error {
	return nil
}

func (s *stubDecisionStore) ListDecisionsByParent(ctx context.Context, tenantID, parentID uuid.UUID) ([]*types.Decision, error) {
	var children []*types.Decision
	for _, d := range s.decisions {
		if d.ParentID != nil && *d.ParentID == parentID {
			children = append(children, d)
		}
	}
	return children, nil
}

func (s *stubDecisionStore) ListByScope(ctx context.Context, tenantID uuid.UUID, scope types.Scope, statuses []types.DecisionStatus) ([]*types.Decision, error) {
	return nil, nil
}

func (s *stubDecisionStore) ListContradictions(ctx context.Context, tenantID uuid.UUID, resolved bool) ([]types.Contradiction, error) {
	return nil, nil
}

func (s *stubDecisionStore) GetContradiction(ctx context.Context, tenantID, id uuid.UUID) (*types.Contradiction, error) {
	return nil, nil
}

func TestGetDecisionLineageIncludesParentAndChildren(t *testing.T) {
	parentID := uuid.New()
	childID := uuid.New()
	store := &stubDecisionStore{decisions: map[uuid.UUID]*types.Decision{
		parentID: {ID: parentID, Status: types.StatusCanonical},
		childID:  {ID: childID, ParentID: &parentID, Status: types.StatusCanonical},
	}}

	engine := NewLineageEngine(store)
	lineage, err := engine.GetDecisionLineage(context.Background(), uuid.Nil, childID)
	if err != nil {
		t.Fatalf("GetDecisionLineage returned error: %v", err)
	}
	if lineage.Parent == nil || lineage.Parent.ID != parentID {
		t.Fatalf("expected parent %s, got %+v", parentID, lineage.Parent)
	}
	if len(lineage.Children) != 0 {
		t.Fatalf("expected no direct children for lineage root, got %d", len(lineage.Children))
	}
}
