package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// HandleTopologyRecommend recommends a topology.
func (s *Server) HandleTopologyRecommend(w http.ResponseWriter, r *http.Request) {
	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, "tenant_id required")
		return
	}

	var req struct {
		Goal            string `json:"goal"`
		ScopeDomain     string `json:"scope_domain"`
		ScopeSystem     string `json:"scope_system"`
		MaxBudgetTokens int64  `json:"max_budget_tokens"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.MaxBudgetTokens == 0 {
		req.MaxBudgetTokens = 50000
	}

	topology, err := s.topologyGenerator.Recommend(r.Context(), tenantID, req.Goal, req.ScopeDomain, req.ScopeSystem, req.MaxBudgetTokens)
	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Save topology and tasks to DB
	if err := s.store.SaveTopology(r.Context(), topology); err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, "failed to save topology: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(topology)
}

// HandleTopologyExecute executes a topology.
func (s *Server) HandleTopologyExecute(w http.ResponseWriter, r *http.Request) {
	topologyIDStr := r.PathValue("id")
	if topologyIDStr == "" {
		s.RespondWithError(w, http.StatusBadRequest, "topology ID required")
		return
	}
	topologyID, err := uuid.Parse(topologyIDStr)
	if err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid topology ID")
		return
	}

	go func() {
		ctx := r.Context()
		_ = s.topologyExecutor.Execute(ctx, topologyID)
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":      "execution_started",
		"topology_id": topologyID.String(),
	})
}

// HandleTopologyStatus returns topology details and SSE stream.
func (s *Server) HandleTopologyStatus(w http.ResponseWriter, r *http.Request) {
	// For MVP, return basic status
	topologyIDStr := r.PathValue("id")
	if topologyIDStr == "" {
		s.RespondWithError(w, http.StatusBadRequest, "topology ID required")
		return
	}
	// Fetch from DB and return
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"active"}`))
}
