package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// HandleDecisionsActiveAt returns decisions that were valid at a specific timestamp.
func (s *Server) HandleDecisionsActiveAt(w http.ResponseWriter, r *http.Request) {
	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, "tenant_id required")
		return
	}

	atStr := r.URL.Query().Get("at")
	if atStr == "" {
		s.RespondWithError(w, http.StatusBadRequest, "missing 'at' timestamp parameter")
		return
	}
	at, err := time.Parse(time.RFC3339, atStr)
	if err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid timestamp format (must use RFC3339)")
		return
	}

	domain := r.URL.Query().Get("domain")
	system := r.URL.Query().Get("system")

	scope := types.Scope{Domain: domain, System: system}
	statuses := []types.DecisionStatus{types.StatusCanonical, types.StatusApproved, types.StatusDraft}

	decisions, err := s.store.GetDecisionsActiveAt(r.Context(), tenantID, at, scope, statuses)
	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"active_at": at,
		"count":     len(decisions),
		"decisions": decisions,
	})
}

// HandleDecisionHistory returns the temporal history of a decision.
func (s *Server) HandleDecisionHistory(w http.ResponseWriter, r *http.Request) {
	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, "tenant_id required")
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		s.RespondWithError(w, http.StatusBadRequest, "invalid request path")
		return
	}
	decisionIDStr := pathParts[len(pathParts)-2] // /api/v1/decisions/{id}/history

	decisionID, err := uuid.Parse(decisionIDStr)
	if err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid decision ID")
		return
	}

	history, err := s.store.GetDecisionHistory(r.Context(), tenantID, decisionID)
	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"decision_id": decisionID,
		"revisions":   history,
	})
}
