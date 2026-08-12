package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/auth"
	"github.com/myshra777-ai/garuda/internal/engine"
	"github.com/myshra777-ai/garuda/internal/types"
)

type fakeDecisionStore struct {
	decisions map[uuid.UUID]*types.Decision
}

func (f *fakeDecisionStore) GetDecision(ctx context.Context, tenantID, decisionID uuid.UUID) (*types.Decision, error) {
	return f.decisions[decisionID], nil
}

func (f *fakeDecisionStore) SaveDecision(ctx context.Context, d *types.Decision) error {
	if f.decisions == nil {
		f.decisions = make(map[uuid.UUID]*types.Decision)
	}
	f.decisions[d.ID] = d
	return nil
}

func (f *fakeDecisionStore) GetDecisionsByScope(ctx context.Context, tenantID uuid.UUID, domain, system string) ([]*types.Decision, error) {
	var out []*types.Decision
	for _, d := range f.decisions {
		if d != nil && d.TenantID == tenantID && d.Scope.Domain == domain && d.Scope.System == system {
			out = append(out, d)
		}
	}
	return out, nil
}

func (f *fakeDecisionStore) GetDecisionRevisions(ctx context.Context, tenantID, decisionID uuid.UUID) ([]types.DecisionRevision, error) {
	return nil, nil
}

func (f *fakeDecisionStore) ListDecisions(ctx context.Context, tenantID uuid.UUID, scope types.Scope, statuses []types.DecisionStatus) ([]*types.Decision, error) {
	return nil, nil
}

func (f *fakeDecisionStore) ListDecisionsByParent(ctx context.Context, tenantID, parentID uuid.UUID) ([]*types.Decision, error) {
	return nil, nil
}

func (f *fakeDecisionStore) SaveCheckpoint(ctx context.Context, c *types.Checkpoint) error {
	return nil
}

func (f *fakeDecisionStore) GetCheckpoint(ctx context.Context, tenantID, id uuid.UUID) (*types.Checkpoint, error) {
	return nil, nil
}

func (f *fakeDecisionStore) ListCheckpointsByAgent(ctx context.Context, tenantID uuid.UUID, agentID string, limit int) ([]*types.Checkpoint, error) {
	return nil, nil
}

func (f *fakeDecisionStore) GetTenantBudget(ctx context.Context, tenantID uuid.UUID) (*types.TenantBudget, error) {
	return nil, nil
}

func (f *fakeDecisionStore) PreflightCheckAndReserve(ctx context.Context, tenantID uuid.UUID, estimatedTokens int) error {
	return nil
}

func (f *fakeDecisionStore) ConsumeBudgetDeduct(ctx context.Context, tenantID uuid.UUID, req types.BudgetConsumptionRequest) (*types.BudgetConsumptionResponse, error) {
	return &types.BudgetConsumptionResponse{Allowed: true}, nil
}

func (f *fakeDecisionStore) QuarantineDecision(ctx context.Context, tenantID uuid.UUID, proposedID, conflictingID uuid.UUID, domain, system, reason string) (*types.Contradiction, error) {
	return nil, nil
}

func (f *fakeDecisionStore) ListUnresolvedContradictions(ctx context.Context, tenantID uuid.UUID) ([]types.Contradiction, error) {
	return nil, nil
}

func (f *fakeDecisionStore) ResolveContradiction(ctx context.Context, id uuid.UUID, strategy string) error {
	return nil
}

func (f *fakeDecisionStore) GetMerkleRoot(ctx context.Context, tenantID uuid.UUID) (*types.MerkleRoot, error) {
	return nil, nil
}

func (f *fakeDecisionStore) AppendMerkleChain(ctx context.Context, tenantID uuid.UUID, decisionHash string) (*types.MerkleRoot, error) {
	return nil, nil
}

func (f *fakeDecisionStore) AddEvidenceBlock(ctx context.Context, tenantID, decisionID uuid.UUID, payload any) (*types.EvidenceBlock, error) {
	return nil, nil
}

func (f *fakeDecisionStore) ListAllTenants(ctx context.Context) ([]uuid.UUID, error) { return nil, nil }

func (f *fakeDecisionStore) GetLatestMerkleSnapshot(ctx context.Context, tenantID uuid.UUID) (*types.MerkleSnapshot, error) {
	return nil, nil
}

func (f *fakeDecisionStore) SaveMerkleSnapshot(ctx context.Context, snap *types.MerkleSnapshot) error {
	return nil
}

func (f *fakeDecisionStore) ListMerkleSnapshots(ctx context.Context, tenantID uuid.UUID, limit int) ([]types.MerkleSnapshot, error) {
	return nil, nil
}

func (f *fakeDecisionStore) GetDecisionsActiveAt(ctx context.Context, tenantID uuid.UUID, at time.Time, scope types.Scope, statuses []types.DecisionStatus) ([]*types.Decision, error) {
	return nil, nil
}

func (f *fakeDecisionStore) GetDecisionHistory(ctx context.Context, tenantID, decisionID uuid.UUID) ([]*types.Decision, error) {
	return nil, nil
}

func (f *fakeDecisionStore) ListContradictions(ctx context.Context, tenantID uuid.UUID, resolved bool) ([]types.Contradiction, error) {
	return nil, nil
}

func (f *fakeDecisionStore) GetContradiction(ctx context.Context, tenantID, id uuid.UUID) (*types.Contradiction, error) {
	return nil, nil
}

func (f *fakeDecisionStore) GetActivePolicies(ctx context.Context, tenantID uuid.UUID, scopeDomain, scopeSystem string) ([]*types.Policy, error) {
	return nil, nil
}

func (f *fakeDecisionStore) GetActivePoliciesByScope(ctx context.Context, tenantID uuid.UUID, scope types.Scope) ([]*types.Policy, error) {
	return nil, nil
}

func (f *fakeDecisionStore) SavePolicy(ctx context.Context, p *types.Policy) error {
	return nil
}

func (f *fakeDecisionStore) SupersedePolicy(ctx context.Context, oldID, newID uuid.UUID) error {
	return nil
}

func (f *fakeDecisionStore) LogPolicyViolation(ctx context.Context, v *types.PolicyViolation) error {
	return nil
}

func (f *fakeDecisionStore) IngestEvidence(ctx context.Context, tenantID uuid.UUID, evidence []types.Evidence) error {
	return nil
}

func (f *fakeDecisionStore) ConsumeBudget(ctx context.Context, tenantID uuid.UUID, tokens int) error {
	return nil
}

// Audit Trail Implementations for fakeDecisionStore
func (f *fakeDecisionStore) LogAuditEvent(ctx context.Context, tenantID uuid.UUID, eventType string, eventID uuid.UUID, actor string, payload interface{}) (*types.AuditEvent, error) {
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

func (f *fakeDecisionStore) VerifyAuditEvent(ctx context.Context, tenantID, eventID uuid.UUID) (*types.AuditVerification, error) {
	return &types.AuditVerification{
		EventID:    eventID,
		IsVerified: true,
	}, nil
}

func (f *fakeDecisionStore) ListAuditEvents(ctx context.Context, tenantID uuid.UUID, since time.Time) ([]types.AuditEvent, error) {
	return []types.AuditEvent{}, nil
}

func (f *fakeDecisionStore) SaveTopology(ctx context.Context, top *types.Topology) error {
	return nil // stub
}
func (f *fakeDecisionStore) GetTopology(ctx context.Context, id uuid.UUID) (*types.Topology, error) {
	return nil, nil
}
func (f *fakeDecisionStore) GetTasksByTopology(ctx context.Context, topologyID uuid.UUID) ([]*types.Task, error) {
	return nil, nil
}
func (f *fakeDecisionStore) UpdateTask(ctx context.Context, task *types.Task) error {
	return nil
}
func (f *fakeDecisionStore) UpdateTopologyStatus(ctx context.Context, id uuid.UUID, status types.TopologyStatus) error {
	return nil
}
func (f *fakeDecisionStore) UpdateTopologyTokens(ctx context.Context, id uuid.UUID, tokens int64) error {
	return nil
}

func TestHandleProposeDecisionRejectsContradictions(t *testing.T) {
	store := &fakeDecisionStore{decisions: map[uuid.UUID]*types.Decision{}}
	jwtConfig, err := auth.NewJWTConfig("garuda", "garuda-api", 5*time.Minute)
	if err != nil {
		t.Fatalf("new jwt config: %v", err)
	}
	tenantID := uuid.New()
	token, err := jwtConfig.GenerateTokenWithTenant("alice", tenantID.String())
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	server := NewServer(store, nil, jwtConfig, engine.NewContradictionEngine(store), nil, nil, nil)

	first := []byte(`{"title":"Use PostgreSQL for financial records","scope_domain":"finance","scope_system":"ledger"}`)
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/decisions/submit", bytes.NewReader(first))
	req1.Header.Set("Authorization", "Bearer "+token)
	req1 = req1.WithContext(auth.ContextWithActorAndTenant(req1.Context(), "alice", tenantID.String()))
	rr1 := httptest.NewRecorder()
	server.HandleProposeDecision(rr1, req1)
	if rr1.Code != http.StatusCreated {
		t.Fatalf("first decision expected 201, got %d: %s", rr1.Code, rr1.Body.String())
	}

	second := []byte(`{"title":"Use MongoDB for financial records","scope_domain":"finance","scope_system":"ledger"}`)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/decisions/submit", bytes.NewReader(second))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2 = req2.WithContext(auth.ContextWithActorAndTenant(req2.Context(), "alice", tenantID.String()))
	rr2 := httptest.NewRecorder()
	server.HandleProposeDecision(rr2, req2)
	if rr2.Code != http.StatusConflict {
		t.Fatalf("second decision expected 409, got %d: %s", rr2.Code, rr2.Body.String())
	}
}

func TestHandleDecisionLineageUsesTenantFromContext(t *testing.T) {
	store := &fakeDecisionStore{decisions: map[uuid.UUID]*types.Decision{}}
	jwtConfig, err := auth.NewJWTConfig("garuda", "garuda-api", 5*time.Minute)
	if err != nil {
		t.Fatalf("new jwt config: %v", err)
	}
	tenantID := uuid.New()
	parentID := uuid.New()
	childID := uuid.New()
	store.decisions[parentID] = &types.Decision{ID: parentID, TenantID: tenantID, Title: "parent", Status: types.StatusCanonical}
	store.decisions[childID] = &types.Decision{ID: childID, TenantID: tenantID, ParentID: &parentID, Title: "child", Status: types.StatusDraft}

	server := NewServer(store, nil, jwtConfig, nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/decisions/"+childID.String()+"/lineage", nil)
	req = req.WithContext(auth.ContextWithActorAndTenant(req.Context(), "alice", tenantID.String()))
	rr := httptest.NewRecorder()

	server.HandleDecisionLineage(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}
func (f *fakeDecisionStore) GetPlan(ctx context.Context, tenantID uuid.UUID, req *types.PlanRequest) (*types.PlanResult, error) {
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

// Minimal SubmitDecision stub for fakeDecisionStore used in API tests.
func (f *fakeDecisionStore) SubmitDecision(ctx context.Context, req *types.SubmitDecisionRequest, actor, requestID string) (*types.SubmitDecisionResult, error) {
	if f.decisions == nil {
		f.decisions = make(map[uuid.UUID]*types.Decision)
	}
	id := req.DecisionID
	if id == uuid.Nil {
		id = uuid.New()
	}
	f.decisions[id] = &types.Decision{ID: id, TenantID: req.TenantID, Title: req.Title, Status: types.StatusDraft}
	return &types.SubmitDecisionResult{DecisionID: id, RevisionID: uuid.New(), RevisionNumber: 1, ContentHash: []byte{}, MerkleRoot: []byte{}}, nil
}
