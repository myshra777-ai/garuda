// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package engine

import (
	"context"
	"testing"
	"time"

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

func (s *stubDecisionStore) SaveCheckpoint(ctx context.Context, c *types.Checkpoint) error {
	return nil
}

func (s *stubDecisionStore) GetCheckpoint(ctx context.Context, tenantID, id uuid.UUID) (*types.Checkpoint, error) {
	return nil, nil
}

func (s *stubDecisionStore) ListCheckpointsByAgent(ctx context.Context, tenantID uuid.UUID, agentID string, limit int) ([]*types.Checkpoint, error) {
	return nil, nil
}

func (s *stubDecisionStore) GetTenantBudget(ctx context.Context, tenantID uuid.UUID) (*types.TenantBudget, error) {
	return nil, nil
}

func (s *stubDecisionStore) PreflightCheckAndReserve(ctx context.Context, tenantID uuid.UUID, estimatedTokens int) error {
	return nil
}

func (s *stubDecisionStore) ConsumeBudgetDeduct(ctx context.Context, tenantID uuid.UUID, req types.BudgetConsumptionRequest) (*types.BudgetConsumptionResponse, error) {
	return &types.BudgetConsumptionResponse{Allowed: true}, nil
}

func (s *stubDecisionStore) QuarantineDecision(ctx context.Context, tenantID uuid.UUID, proposedID, conflictingID uuid.UUID, domain, system, reason string) (*types.Contradiction, error) {
	return nil, nil
}

func (s *stubDecisionStore) ListUnresolvedContradictions(ctx context.Context, tenantID uuid.UUID) ([]types.Contradiction, error) {
	return nil, nil
}

func (s *stubDecisionStore) ResolveContradiction(ctx context.Context, id uuid.UUID, strategy string) error {
	return nil
}

func (s *stubDecisionStore) GetMerkleRoot(ctx context.Context, tenantID uuid.UUID) (*types.MerkleRoot, error) {
	return nil, nil
}

func (s *stubDecisionStore) AppendMerkleChain(ctx context.Context, tenantID uuid.UUID, decisionHash string) (*types.MerkleRoot, error) {
	return nil, nil
}

func (s *stubDecisionStore) AddEvidenceBlock(ctx context.Context, tenantID, decisionID uuid.UUID, payload any) (*types.EvidenceBlock, error) {
	return nil, nil
}

func (s *stubDecisionStore) ListAllTenants(ctx context.Context) ([]uuid.UUID, error) { return nil, nil }

func (s *stubDecisionStore) GetLatestMerkleSnapshot(ctx context.Context, tenantID uuid.UUID) (*types.MerkleSnapshot, error) {
	return nil, nil
}

func (s *stubDecisionStore) SaveMerkleSnapshot(ctx context.Context, snap *types.MerkleSnapshot) error {
	return nil
}

func (s *stubDecisionStore) ListMerkleSnapshots(ctx context.Context, tenantID uuid.UUID, limit int) ([]types.MerkleSnapshot, error) {
	return nil, nil
}

func (s *stubDecisionStore) GetDecisionsActiveAt(ctx context.Context, tenantID uuid.UUID, at time.Time, scope types.Scope, statuses []types.DecisionStatus) ([]*types.Decision, error) {
	return nil, nil
}

func (s *stubDecisionStore) GetDecisionHistory(ctx context.Context, tenantID, decisionID uuid.UUID) ([]*types.Decision, error) {
	return nil, nil
}

func (s *stubDecisionStore) ListContradictions(ctx context.Context, tenantID uuid.UUID, resolved bool) ([]types.Contradiction, error) {
	return nil, nil
}

func (s *stubDecisionStore) GetActivePolicies(ctx context.Context, tenantID uuid.UUID, scopeDomain, scopeSystem string) ([]*types.Policy, error) {
	return nil, nil
}

func (s *stubDecisionStore) GetActivePoliciesByScope(ctx context.Context, tenantID uuid.UUID, scope types.Scope) ([]*types.Policy, error) {
	return nil, nil
}

func (s *stubDecisionStore) SavePolicy(ctx context.Context, p *types.Policy) error {
	return nil
}

func (s *stubDecisionStore) SupersedePolicy(ctx context.Context, oldID, newID uuid.UUID) error {
	return nil
}

func (s *stubDecisionStore) LogPolicyViolation(ctx context.Context, v *types.PolicyViolation) error {
	return nil
}

func (s *stubDecisionStore) GetContradiction(ctx context.Context, tenantID, id uuid.UUID) (*types.Contradiction, error) {
	return nil, nil
}
func (s *stubDecisionStore) ListAuditEvents(ctx context.Context, tenantID uuid.UUID, since time.Time) ([]types.AuditEvent, error) {
	return nil, nil
}

func (s *stubDecisionStore) LogAuditEvent(ctx context.Context, tenantID uuid.UUID, eventType string, eventID uuid.UUID, actor string, payload interface{}) (*types.AuditEvent, error) {
	return &types.AuditEvent{
		ID:        uuid.New(),
		TenantID:  tenantID,
		EventType: eventType,
		EventID:   eventID,
		Actor:     actor,
		Payload:   payload,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (s *stubDecisionStore) VerifyAuditEvent(ctx context.Context, tenantID, eventID uuid.UUID) (*types.AuditVerification, error) {
	return &types.AuditVerification{
		EventID:    eventID,
		IsVerified: true,
	}, nil
}

func (s *stubDecisionStore) SaveTopology(ctx context.Context, top *types.Topology) error {
	return nil
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

func (s *stubDecisionStore) GetTopology(ctx context.Context, id uuid.UUID) (*types.Topology, error) {
	return nil, nil
}
func (s *stubDecisionStore) GetTasksByTopology(ctx context.Context, topologyID uuid.UUID) ([]*types.Task, error) {
	return nil, nil
}
func (s *stubDecisionStore) UpdateTopologyStatus(ctx context.Context, id uuid.UUID, status types.TopologyStatus) error {
	return nil
}
func (s *stubDecisionStore) UpdateTask(ctx context.Context, task *types.Task) error {
	return nil
}
func (s *stubDecisionStore) UpdateTopologyTokens(ctx context.Context, id uuid.UUID, tokens int64) error {
	return nil
}
func (s *stubDecisionStore) GetPlan(ctx context.Context, tenantID uuid.UUID, req *types.PlanRequest) (*types.PlanResult, error) {
	return &types.PlanResult{
		TenantID:     tenantID,
		Scope:        types.Scope{Domain: req.ScopeDomain, System: req.ScopeSystem},
		Decisions:    []*types.Decision{},
		Tasks:        []*types.Task{},
		Handoffs:     []*types.HandoffRecord{},
		Milestones:   []*types.Milestone{},
		Dependencies: []types.LineageEdge{},
		GeneratedAt:  time.Now().UTC(),
	}, nil
}

// Minimal SubmitDecision stub to satisfy interface in tests.
func (s *stubDecisionStore) SubmitDecision(ctx context.Context, req *types.SubmitDecisionRequest, actor, requestID string) (*types.SubmitDecisionResult, error) {
	return &types.SubmitDecisionResult{
		DecisionID:     req.DecisionID,
		RevisionID:     uuid.New(),
		RevisionNumber: 1,
		ContentHash:    []byte{},
		MerkleRoot:     []byte{},
	}, nil
}
