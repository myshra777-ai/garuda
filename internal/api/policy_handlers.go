package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/telemetry"
	"github.com/myshra777-ai/garuda/internal/types"
)

// HandleRememberPolicy creates a new policy (equivalent to "/garuda remember").
func (s *Server) HandleRememberPolicy(w http.ResponseWriter, r *http.Request) {
	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, "tenant_id required")
		return
	}

	var req struct {
		Statement   string     `json:"statement"`
		ScopeDomain string     `json:"scope_domain"`
		ScopeSystem string     `json:"scope_system"`
		ValidTo     *time.Time `json:"valid_to,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Statement == "" {
		s.RespondWithError(w, http.StatusBadRequest, "statement is required")
		return
	}
	if req.ScopeDomain == "" || req.ScopeSystem == "" {
		s.RespondWithError(w, http.StatusBadRequest, "scope_domain and scope_system are required")
		return
	}

	actor := getActorFromContext(r.Context())

	policy := &types.Policy{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Statement:   req.Statement,
		ScopeDomain: req.ScopeDomain,
		ScopeSystem: req.ScopeSystem,
		Actor:       actor,
		Status:      "active",
		ValidFrom:   time.Now().UTC(),
		ValidTo:     req.ValidTo,
		Metadata:    map[string]interface{}{"source": "cli"},
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := s.store.SavePolicy(r.Context(), policy); err != nil {
		telemetry.RecordError("policy_save_failed", err.Error())
		s.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	telemetry.RecordFeatureUsage("policy_created")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        policy.ID.String(),
		"status":    "active",
		"statement": policy.Statement,
		"scope":     fmt.Sprintf("%s/%s", policy.ScopeDomain, policy.ScopeSystem),
	})
}

// HandleListPolicies lists active policies for a scope.
func (s *Server) HandleListPolicies(w http.ResponseWriter, r *http.Request) {
	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, "tenant_id required")
		return
	}
	domain := r.URL.Query().Get("domain")
	system := r.URL.Query().Get("system")
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "active"
	}

	policies, err := s.store.GetActivePolicies(r.Context(), tenantID, domain, system)
	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(policies)
}

// HandleSupersedePolicy supersedes an existing policy with a new one.
func (s *Server) HandleSupersedePolicy(w http.ResponseWriter, r *http.Request) {
	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, "tenant_id required")
		return
	}
	// Extract policy ID from URL
	policyIDStr := r.PathValue("id")
	if policyIDStr == "" {
		s.RespondWithError(w, http.StatusBadRequest, "policy ID required")
		return
	}
	policyID, err := uuid.Parse(policyIDStr)
	if err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid policy ID")
		return
	}

	var req struct {
		Statement   string     `json:"statement"`
		ScopeDomain string     `json:"scope_domain"`
		ScopeSystem string     `json:"scope_system"`
		ValidTo     *time.Time `json:"valid_to,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Statement == "" {
		s.RespondWithError(w, http.StatusBadRequest, "statement is required")
		return
	}

	actor := getActorFromContext(r.Context())

	// Create new policy
	newPolicy := &types.Policy{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Statement:   req.Statement,
		ScopeDomain: req.ScopeDomain,
		ScopeSystem: req.ScopeSystem,
		Actor:       actor,
		Status:      "active",
		ValidFrom:   time.Now().UTC(),
		ValidTo:     req.ValidTo,
		Metadata:    map[string]interface{}{"supersedes": policyID.String()},
	}
	if err := s.store.SavePolicy(r.Context(), newPolicy); err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Mark old as superseded
	if err := s.store.SupersedePolicy(r.Context(), policyID, newPolicy.ID); err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"superseded_policy": policyID.String(),
		"new_policy":        newPolicy.ID.String(),
		"status":            "superseded",
	})
}
