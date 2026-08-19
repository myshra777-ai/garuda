// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/budget"
	"github.com/myshra777-ai/garuda/internal/types"
)

// BudgetMiddleware performs an active pre-flight check before passing the request down.
func (s *Server) BudgetMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID, err := resolveTenantID(r, uuid.Nil)
		if err != nil || tenantID == uuid.Nil {
			next.ServeHTTP(w, r)
			return
		}

		b, err := s.store.GetTenantBudget(r.Context(), tenantID)
		if err != nil {
			http.Error(w, `{"error":"failed to fetch tenant budget"}`, http.StatusInternalServerError)
			return
		}

		if b.Status == "exhausted" || b.TokenBalance <= 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPaymentRequired)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "insufficient budget",
				"balance": b.TokenBalance,
			})
			return
		}

		ctx := context.WithValue(r.Context(), "budget", b)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ConsumeBudgetForRequest deducts estimated tokens after successful handler execution.
func (s *Server) ConsumeBudgetForRequest(r *http.Request, operation string, payload interface{}) error {
	tenantID, err := resolveTenantID(r, uuid.Nil)
	if err != nil || tenantID == uuid.Nil {
		return nil
	}

	agentID := r.Header.Get("X-Agent-ID")
	if agentID == "" {
		if a, ok := r.Context().Value("agent_id").(string); ok {
			agentID = a
		}
	}
	if agentID == "" {
		agentID = "unknown-agent"
	}

	estimatedTokens := budget.EstimateTokens(operation, payload)
	req := types.BudgetConsumptionRequest{
		AgentID:        agentID,
		TokensUsed:     estimatedTokens,
		ExecutionsUsed: 1,
		Operation:      operation,
	}

	resp, err := s.store.ConsumeBudgetDeduct(r.Context(), tenantID, req)
	if err != nil {
		return err
	}
	if !resp.Allowed {
		return &BudgetExceededError{Balance: resp.RemainingTokens}
	}
	return nil
}

type BudgetExceededError struct {
	Balance int64
}

func (e *BudgetExceededError) Error() string {
	return "budget exhausted"
}
