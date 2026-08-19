// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

// HandleGetBudget returns current budget usage and limits for the tenant.
func (s *Server) HandleGetBudget(w http.ResponseWriter, r *http.Request) {
	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, "tenant_id is required")
		return
	}

	budget, err := s.store.GetTenantBudget(r.Context(), tenantID)
	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, "failed to retrieve budget")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(budget)
}

// HandleConsumeBudget deducts tokens or executions for an agent action.
func (s *Server) HandleConsumeBudget(w http.ResponseWriter, r *http.Request) {
	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil {
		s.RespondWithError(w, http.StatusUnauthorized, "tenant_id is required")
		return
	}

	var req types.BudgetConsumptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "invalid budget request payload")
		return
	}

	if req.AgentID == "" {
		s.RespondWithError(w, http.StatusUnprocessableEntity, "agent_id is required")
		return
	}

	resp, err := s.store.ConsumeBudgetDeduct(r.Context(), tenantID, req)
	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if !resp.Allowed {
		w.WriteHeader(http.StatusPaymentRequired) // 402 Payment Required
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_ = json.NewEncoder(w).Encode(resp)
}
