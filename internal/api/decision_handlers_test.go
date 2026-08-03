package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/auth"
	"github.com/myshra777-ai/garuda/internal/telemetry"
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

func (f *fakeDecisionStore) ListContradictions(ctx context.Context, tenantID uuid.UUID, resolved bool) ([]types.Contradiction, error) {
	return nil, nil
}

func (f *fakeDecisionStore) GetContradiction(ctx context.Context, tenantID, id uuid.UUID) (*types.Contradiction, error) {
	return nil, nil
}

func (f *fakeDecisionStore) IngestEvidence(ctx context.Context, tenantID uuid.UUID, evidence []types.Evidence) error {
	return nil
}

func (f *fakeDecisionStore) ConsumeBudget(ctx context.Context, tenantID uuid.UUID, tokens int) error {
	return nil
}

// HandleProposeDecision validates and persists a new decision proposal.
func (s *Server) HandleProposeDecision(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	telemetry.RecordFeatureUsage("submit_decision")

	var req DecisionProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		telemetry.RecordError("malformed_json", err.Error())
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	tenantID, err := resolveTenantID(r, req.TenantID)
	if err != nil {
		telemetry.RecordError("validation_failure", err.Error())
		http.Error(w, `{"error":"tenant_id is required"}`, http.StatusUnprocessableEntity)
		return
	}

	if req.Title == "" {
		telemetry.RecordError("validation_failure", "missing title")
		http.Error(w, `{"error":"title is required"}`, http.StatusUnprocessableEntity)
		return
	}

	now := time.Now().UTC()
	actor, _ := auth.ActorFromContext(r.Context())
	if actor == "" {
		actor = "system"
	}

	decision := &types.Decision{
		ID:               uuid.New(),
		TenantID:         tenantID,
		Title:            req.Title,
		Status:           types.StatusDraft,
		ScopeDomain:      req.ScopeDomain, // Populated flat field
		ScopeSystem:      req.ScopeSystem, // Populated flat field
		Scope:            types.Scope{Domain: req.ScopeDomain, System: req.ScopeSystem},
		Owner:            actor,
		Confidence:       0.8,
		EvidenceIDs:      req.EvidenceIDs,
		TemporalMetadata: req.TemporalMetadata,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// 1. Save decision and expose exact error on failure
	if err := s.store.SaveDecision(r.Context(), decision); err != nil {
		log.Printf("ERROR: SaveDecision failed: %v", err)
		telemetry.RecordError("db_save_failed", err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("failed to save decision: %v", err),
		})
		return
	}

	// 2. Run Autonomous Contradiction Quarantine Engine
	var quarantineResult *types.Contradiction
	if s.contradictionEngine != nil {
		q, err := s.contradictionEngine.DetectAndQuarantine(r.Context(), decision)
		if err != nil {
			telemetry.RecordError("quarantine_evaluation_failed", err.Error())
		} else if q != nil {
			quarantineResult = q
			telemetry.RecordContradictionDetected()
		}
	}

	// 3. Consume budget post-execution
	if err := s.ConsumeBudgetForRequest(r, "propose_decision", req); err != nil {
		telemetry.RecordError("budget_consume_failed", err.Error())
	}

	telemetry.RecordDecisionProposed(quarantineResult != nil)
	telemetry.RecordAPILatency(time.Since(start))

	// 4. Build response payload
	resp := map[string]interface{}{
		"id":        decision.ID.String(),
		"status":    decision.Status,
		"tenant_id": decision.TenantID.String(),
		"title":     decision.Title,
	}

	if quarantineResult != nil {
		resp["quarantined"] = true
		resp["status"] = types.StatusQuarantined
		resp["contradiction_id"] = quarantineResult.ID.String()
		resp["conflicting_decision_id"] = quarantineResult.DecisionA.String()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
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

	server := NewServer(store, nil, jwtConfig, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/decisions/"+childID.String()+"/lineage", nil)
	req = req.WithContext(auth.ContextWithActorAndTenant(req.Context(), "alice", tenantID.String()))
	rr := httptest.NewRecorder()

	server.HandleDecisionLineage(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}
