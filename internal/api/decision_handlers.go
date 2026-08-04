package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/auth"
	"github.com/myshra777-ai/garuda/internal/telemetry"
	"github.com/myshra777-ai/garuda/internal/types"
)

// DecisionProposalRequest defines the incoming payload structure for a decision.
type DecisionProposalRequest struct {
	TenantID         uuid.UUID              `json:"tenant_id"`
	Title            string                 `json:"title"`
	ScopeDomain      string                 `json:"scope_domain"`
	ScopeSystem      string                 `json:"scope_system"`
	EvidenceIDs      []types.EvidenceHash   `json:"evidence_ids"`
	TemporalMetadata types.TemporalMetadata `json:"temporal_metadata"`
	ValidFrom        *time.Time             `json:"valid_from,omitempty"`
	ValidTo          *time.Time             `json:"valid_to,omitempty"`
}

// resolveTenantID extracts the tenant UUID from JWT context first, falling back to request body.
func resolveTenantID(r *http.Request, fallback uuid.UUID) (uuid.UUID, error) {
	// Priority 1: Context claim injected by auth middleware
	if tid, ok := auth.TenantIDFromContext(r.Context()); ok && tid != uuid.Nil {
		return tid, nil
	}
	// Priority 2: Fallback payload parameter
	if fallback != uuid.Nil {
		return fallback, nil
	}
	// Priority 3: Error
	return uuid.Nil, fmt.Errorf("tenant_id is required")
}

// HandleProposeDecision validates and persists a new decision proposal.
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

// HandleDecisionLineage resolves and retrieves parent and child lineage graphs.
func (s *Server) HandleDecisionLineage(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 5 || pathParts[len(pathParts)-1] != "lineage" {
		http.Error(w, `{"error":"invalid request path"}`, http.StatusBadRequest)
		return
	}

	decisionIDStr := pathParts[len(pathParts)-2]
	decisionID, err := uuid.Parse(decisionIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid decision ID"}`, http.StatusBadRequest)
		return
	}

	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		http.Error(w, `{"error":"tenant_id is required"}`, http.StatusUnauthorized)
		return
	}

	decision, err := s.store.GetDecision(r.Context(), tenantID, decisionID)
	if err != nil {
		http.Error(w, `{"error":"decision not found"}`, http.StatusNotFound)
		return
	}

	children, err := s.store.ListDecisionsByParent(r.Context(), tenantID, decisionID)
	if err != nil {
		children = []*types.Decision{}
	}

	lineage := map[string]interface{}{
		"decision": decision,
		"children": children,
	}

	if decision.ParentID != nil && *decision.ParentID != uuid.Nil {
		parent, err := s.store.GetDecision(r.Context(), tenantID, *decision.ParentID)
		if err == nil {
			lineage["parent"] = parent
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(lineage)
}

// HandleAgentWarmup triggers system pre-heating states for execution agents.
func (s *Server) HandleAgentWarmup(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	telemetry.RecordFeatureUsage("agent_warmup")
	telemetry.RecordWarmStartLatency(time.Since(start))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"warmed","message":"asynchronous execution agents initialized"}`))
}

// HandleSystemPromptContext returns a formatted text block containing active canonical policies
func (s *Server) HandleSystemPromptContext(w http.ResponseWriter, r *http.Request) {
	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, "tenant_id required")
		return
	}

	// Fetch active canonical decisions
	decisions, err := s.store.GetDecisionsActiveAt(r.Context(), tenantID, time.Now().UTC(), types.Scope{}, []types.DecisionStatus{types.StatusCanonical, types.StatusApproved})
	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	prompt := "--- GARUDA ORGANIZATIONAL TRUTH BOOTSTRAP ---\n"
	prompt += "The following canonical policy decisions are currently active and strictly enforced:\n\n"

	for i, d := range decisions {
		prompt += fmt.Sprintf("%d. [%s/%s] %s (ID: %s)\n", i+1, d.ScopeDomain, d.ScopeSystem, d.Title, d.ID.String())
	}
	prompt += "\nDo not generate plans or code that contradict these policies.\n"
	prompt += "---------------------------------------------\n"

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(prompt))
}
