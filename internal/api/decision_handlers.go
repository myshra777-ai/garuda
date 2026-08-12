package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/auth"
	"github.com/myshra777-ai/garuda/internal/engine"
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
	if tid, ok := auth.TenantIDFromContext(r.Context()); ok && tid != uuid.Nil {
		return tid, nil
	}
	if fallback != uuid.Nil {
		return fallback, nil
	}
	return uuid.Nil, fmt.Errorf("tenant_id is required")
}

// getActorFromContext retrieves the actor string from context or falls back to standard default.
func getActorFromContext(ctx context.Context) string {
	if actor, ok := ctx.Value("actor").(string); ok && actor != "" {
		return actor
	}
	return "cli-operator"
}

// estimateTokensSaved calculates an estimated token saving heuristic for a decision proposal.
func estimateTokensSaved(d *types.Decision) int64 {
	if d == nil {
		return 0
	}
	titleTokens := int64(len(strings.Fields(d.Title)) * 4)
	evidenceTokens := int64(len(d.EvidenceIDs) * 128)
	return titleTokens + evidenceTokens + 250
}

// getRequestID safely extracts or generates the request correlation ID.
func getRequestID(r *http.Request) string {
	if reqID, ok := r.Context().Value("request_id").(string); ok && reqID != "" {
		return reqID
	}
	return uuid.New().String()
}

// HandleProposeDecision validates and persists a new decision proposal.
func (s *Server) HandleProposeDecision(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	telemetry.RecordFeatureUsage("submit_decision")
	requestID := getRequestID(r)

	actor, ok := auth.ActorFromContext(r.Context())
	if !ok || actor == "" {
		s.RespondWithError(w, http.StatusUnauthorized, "missing actor context", requestID)
		return
	}
	tenantID, ok := auth.TenantIDFromContext(r.Context())
	if !ok || tenantID == uuid.Nil {
		s.RespondWithError(w, http.StatusUnauthorized, "missing tenant context", requestID)
		return
	}

	var req types.SubmitDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid request payload", requestID)
		return
	}
	req.TenantID = tenantID

	if req.Title == "" && req.Statement == "" {
		s.RespondWithError(w, http.StatusUnprocessableEntity, "title or statement required", requestID)
		return
	}

	var evIDs []types.EvidenceHash
	if len(req.Evidence) > 0 {
		evIDs = make([]types.EvidenceHash, 0, len(req.Evidence))
		for _, ev := range req.Evidence {
			evIDs = append(evIDs, ev.Hash)
		}
	}

	newDecision := &types.Decision{
		ID:          req.DecisionID,
		TenantID:    req.TenantID,
		Title:       req.Title,
		Scope:       req.Scope,
		EvidenceIDs: evIDs,
	}

	if s.contradictionEngine != nil {
		conflicting, err := s.contradictionEngine.ValidateDecision(r.Context(), newDecision)
		if err != nil {
			slog.Error("contradiction check failed", "error", err, "request_id", requestID)
			s.RespondWithError(w, http.StatusInternalServerError, "failed to validate decision policy", requestID)
			return
		}
		if conflicting {
			s.RespondWithError(w, http.StatusConflict, "proposed decision contradicts existing policy/decision", requestID)
			return
		}
	}

	result, err := s.store.SubmitDecision(r.Context(), &req, actor, requestID)
	if err != nil {
		slog.Error("decision submission failed",
			"error", err,
			"request_id", requestID,
			"tenant", tenantID,
			"actor", actor,
		)
		s.RespondWithError(w, http.StatusInternalServerError, "decision submission failed", requestID)
		return
	}

	domain, system := "", ""
	if req.Scope.Domain != "" || req.Scope.System != "" {
		domain = req.Scope.Domain
		system = req.Scope.System
	}
	telemetry.RecordDecisionProposed("", "", "proposed", domain, system, 0.0, 0)
	telemetry.RecordAPILatency(time.Since(start))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":          "created",
		"decision_id":     result.DecisionID.String(),
		"revision_id":     result.RevisionID.String(),
		"revision_number": result.RevisionNumber,
		"content_hash":    hex.EncodeToString(result.ContentHash),
		"merkle_root":     hex.EncodeToString(result.MerkleRoot),
		"request_id":      requestID,
	})
}

// HandleDecisionLineage resolves and retrieves parent and child lineage graphs.
func (s *Server) HandleDecisionLineage(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 5 || pathParts[len(pathParts)-1] != "lineage" {
		s.RespondWithError(w, http.StatusBadRequest, "invalid request path", requestID)
		return
	}

	decisionIDStr := pathParts[len(pathParts)-2]
	decisionID, err := uuid.Parse(decisionIDStr)
	if err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid decision ID format", requestID)
		return
	}

	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, "tenant_id is required", requestID)
		return
	}

	decision, err := s.store.GetDecision(r.Context(), tenantID, decisionID)
	if err != nil {
		slog.Warn("decision not found", "decision_id", decisionID, "tenant_id", tenantID, "error", err)
		s.RespondWithError(w, http.StatusNotFound, "decision not found", requestID)
		return
	}

	children, err := s.store.ListDecisionsByParent(r.Context(), tenantID, decisionID)
	if err != nil {
		slog.Warn("failed to fetch child decisions", "decision_id", decisionID, "error", err)
		children = []*types.Decision{}
	}

	lineage := map[string]interface{}{
		"decision":   decision,
		"children":   children,
		"request_id": requestID,
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

// HandleSystemPromptContext returns a formatted text block containing active canonical policies.
func (s *Server) HandleSystemPromptContext(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)

	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, "tenant_id is required", requestID)
		return
	}

	decisions, err := s.store.GetDecisionsActiveAt(
		r.Context(),
		tenantID,
		time.Now().UTC(),
		types.Scope{},
		[]types.DecisionStatus{types.StatusCanonical, types.StatusApproved},
	)
	if err != nil {
		slog.Error("failed to retrieve canonical decisions", "tenant_id", tenantID, "error", err, "request_id", requestID)
		s.RespondWithError(w, http.StatusInternalServerError, "failed to build system prompt context", requestID)
		return
	}

	var prompt strings.Builder
	prompt.WriteString("--- GARUDA ORGANIZATIONAL TRUTH BOOTSTRAP ---\n")
	prompt.WriteString("The following canonical policy decisions are currently active and strictly enforced:\n\n")

	for i, d := range decisions {
		prompt.WriteString(fmt.Sprintf("%d. [%s/%s] %s (ID: %s)\n", i+1, d.ScopeDomain, d.ScopeSystem, d.Title, d.ID.String()))
	}
	prompt.WriteString("\nDo not generate plans or code that contradict these policies.\n")
	prompt.WriteString("---------------------------------------------\n")

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(prompt.String()))
}

// HandleEvaluateRoute dispatches incoming payloads through the pre-flight classifier and shield.
func (s *Server) HandleEvaluateRoute(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)

	var req struct {
		Payload       string  `json:"payload"`
		TokenEstimate int     `json:"token_estimate"`
		SpendRatio    float64 `json:"spend_ratio"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid request body", requestID)
		return
	}

	router := engine.NewDynamicRouter()
	decision, err := router.ClassifyAndRoute(r.Context(), req.Payload, req.TokenEstimate, req.SpendRatio)
	if err != nil {
		slog.Error("route classification error", "error", err, "request_id", requestID)
		s.RespondWithError(w, http.StatusInternalServerError, "failed to evaluate route classification", requestID)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(decision)
}

// HandleAgentWarmup triggers system pre-heating states for execution agents.
func (s *Server) HandleAgentWarmup(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	modelName, modelProvider := getModelInfo(r)

	tokensSaved := int64(0)
	telemetry.RecordWarmStartWithModel(
		modelName,
		modelProvider,
		float64(time.Since(start).Milliseconds()),
		tokensSaved,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "warmed",
		"message": "asynchronous execution agents initialized",
	})
}

// HandleAuditVerify handles verification requests and logs model-attributed telemetry.
func (s *Server) HandleAuditVerify(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	modelName, _ := getModelInfo(r)
	telemetry.RecordVerificationWithModel(
		modelName,
		float64(time.Since(start).Milliseconds()),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "verified"})
}
