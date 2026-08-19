// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/myshra777-ai/garuda/internal/types"
)

type ContradictionEngine struct {
	store types.DecisionStore
}

func NewContradictionEngine(store types.DecisionStore) *ContradictionEngine {
	return &ContradictionEngine{store: store}
}

// ValidateDecision reports whether a proposed decision conflicts with existing ones in scope.
func (e *ContradictionEngine) ValidateDecision(ctx context.Context, newDecision *types.Decision) (bool, error) {
	if e == nil || e.store == nil || newDecision == nil {
		return false, nil
	}

	existing, err := e.store.GetDecisionsByScope(ctx, newDecision.TenantID, newDecision.Scope.Domain, newDecision.Scope.System)
	if err != nil {
		return false, err
	}

	for _, d := range existing {
		if d == nil || d.ID == newDecision.ID || string(d.Status) == "quarantined" || string(d.Status) == "superseded" {
			continue
		}
		if isContradictory(newDecision, d) {
			return true, nil
		}
	}

	return false, nil
}

// DetectAndQuarantine evaluates rules and quarantines new contradictory decisions.
func (e *ContradictionEngine) DetectAndQuarantine(ctx context.Context, newDecision *types.Decision) (*types.Contradiction, error) {
	conflicting, err := e.ValidateDecision(ctx, newDecision)
	if err != nil {
		return nil, err
	}
	if !conflicting {
		return nil, nil
	}

	// Fetch the conflicting decision to build a stable quarantine reason.
	existing, err := e.store.GetDecisionsByScope(ctx, newDecision.TenantID, newDecision.Scope.Domain, newDecision.Scope.System)
	if err != nil {
		return nil, err
	}

	var conflictingDecision *types.Decision
	for _, d := range existing {
		if d != nil && d.ID != newDecision.ID && !strings.EqualFold(string(d.Status), "quarantined") && !strings.EqualFold(string(d.Status), "superseded") && isContradictory(newDecision, d) {
			conflictingDecision = d
			break
		}
	}
	if conflictingDecision == nil {
		return nil, nil
	}

	reason := fmt.Sprintf("Proposed decision '%s' contradicts active decision '%s' (ID: %s) in scope %s/%s",
		newDecision.Title, conflictingDecision.Title, conflictingDecision.ID.String(),
		newDecision.Scope.Domain, newDecision.Scope.System)

	record, err := e.store.QuarantineDecision(ctx, newDecision.TenantID,
		newDecision.ID, conflictingDecision.ID,
		newDecision.Scope.Domain, newDecision.Scope.System, reason)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func isContradictory(a, b *types.Decision) bool {
	if a == nil || b == nil {
		return false
	}
	if strings.TrimSpace(a.Title) == "" || strings.TrimSpace(b.Title) == "" {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(a.Title), strings.TrimSpace(b.Title)) {
		return true
	}

	aLower := strings.ToLower(a.Title)
	bLower := strings.ToLower(b.Title)

	if a.Scope.Domain == b.Scope.Domain && a.Scope.System == b.Scope.System {
		if aLower == bLower {
			return true
		}
		if aLower == strings.ToLower(b.Title) {
			return true
		}
	}

	techA, hasTechA := detectTechnologyToken(aLower)
	techB, hasTechB := detectTechnologyToken(bLower)
	if hasTechA && hasTechB && techA != techB {
		return true
	}

	return false
}

func detectTechnologyToken(s string) (string, bool) {
	for _, token := range []string{"postgres", "postgresql", "mysql", "mongodb", "redis", "sqlite", "sqlserver", "oracle", "cockroach", "elasticsearch"} {
		if strings.Contains(s, token) {
			return token, true
		}
	}
	return "", false
}

// CheckPolicyViolation returns true if the proposal violates any active policy.
func (e *ContradictionEngine) CheckPolicyViolation(ctx context.Context, proposal *types.Decision) (bool, string, *types.Policy, error) {
	policies, err := e.store.GetActivePolicies(ctx, proposal.TenantID, proposal.Scope.Domain, proposal.Scope.System)
	if err != nil {
		return false, "", nil, err
	}
	for _, p := range policies {
		if isProposalViolatingPolicy(proposal, p) {
			return true, p.Statement, p, nil
		}
	}
	return false, "", nil, nil
}

func isProposalViolatingPolicy(proposal *types.Decision, policy *types.Policy) bool {
	// Check for explicit contradictions (simplified)
	if policy.Statement == "do not change schema" && strings.Contains(proposal.Title, "change schema") {
		return true
	}
	// Check if policy statement appears in proposal title or rationale
	if strings.Contains(strings.ToLower(proposal.Title), strings.ToLower(policy.Statement)) {
		return true
	}
	return false
}
