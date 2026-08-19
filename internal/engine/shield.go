// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package engine

import (
	"context"
	"fmt"
	"regexp"

	"github.com/myshra777-ai/garuda/internal/types"
)

// ShieldEngine handles pattern matching and inline content redaction.
type ShieldEngine struct {
	patterns []*regexp.Regexp
}

// NewShieldEngine initializes a new ShieldEngine with default sensitive patterns.
func NewShieldEngine() *ShieldEngine {
	return &ShieldEngine{
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(api_key|apikey|secret|password|bearer|auth_token)\s*=\s*['"][a-zA-Z0-9_\-]{16,}['"]`),
			regexp.MustCompile(`(sk-[a-zA-Z0-9]{32,})`),
			regexp.MustCompile(`(ghp_[a-zA-Z0-9]{36})`),
			regexp.MustCompile(`(xox[baprs]-[a-zA-Z0-9]{10,})`),
			regexp.MustCompile(`(?i)\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`), // Email/PII
		},
	}
}

// Redact sanitizes sensitive data inline prior to provider dispatch.
func (s *ShieldEngine) Redact(input string) (string, int) {
	cleaned := input
	count := 0

	for _, pattern := range s.patterns {
		matches := pattern.FindAllString(cleaned, -1)
		count += len(matches)
		cleaned = pattern.ReplaceAllString(cleaned, "[REDACTED_SECRET]")
	}

	return cleaned, count
}

// PreFlightShield intercepts and validates agent actions before execution.
type PreFlightShield struct {
	contradictionEngine *ContradictionEngine
	shieldEngine        *ShieldEngine
}

// NewPreFlightShield creates a new pre-flight shield.
func NewPreFlightShield(ce *ContradictionEngine) *PreFlightShield {
	return &PreFlightShield{
		contradictionEngine: ce,
		shieldEngine:        NewShieldEngine(),
	}
}

// ShieldRequest represents an action to validate.
type ShieldRequest struct {
	AgentID   string
	Role      types.AgentRole
	Scope     types.Scope
	Decision  *types.Decision
	TokenCost int64
}

// ShieldResult contains the validation outcome.
type ShieldResult struct {
	Allowed        bool                 `json:"allowed"`
	Reason         string               `json:"reason"`
	Contradiction  *types.Contradiction `json:"contradiction,omitempty"`
	BudgetCheck    *BudgetResult        `json:"budget_check,omitempty"`
	RedactionCount int                  `json:"redaction_count,omitempty"`
}

// BudgetResult contains budget validation details.
type BudgetResult struct {
	WithinBudget bool   `json:"within_budget"`
	Remaining    int64  `json:"remaining"`
	WarningLevel string `json:"warning_level"` // ok, warning, critical
}

// Validate checks all pre-flight conditions.
func (s *PreFlightShield) Validate(ctx context.Context, req *ShieldRequest) (*ShieldResult, error) {
	redactionCount := 0

	// 1. Role scope check
	if !s.checkRoleScope(req.Role, req.Scope) {
		return &ShieldResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Role %s cannot modify scope %s/%s", req.Role, req.Scope.Domain, req.Scope.System),
		}, nil
	}

	// 2. Contradiction check & inline decision text redaction
	if req.Decision != nil {
		if req.Decision.Statement != "" {
			redactedStatement, count := s.shieldEngine.Redact(req.Decision.Statement)
			req.Decision.Statement = redactedStatement
			redactionCount += count
		}

		conflict, err := s.contradictionEngine.ValidateDecision(ctx, req.Decision)
		if err != nil {
			return nil, err
		}
		if conflict {
			return &ShieldResult{
				Allowed:        false,
				Reason:         "Contradiction detected",
				RedactionCount: redactionCount,
			}, nil
		}
	}

	// 3. Budget check
	// Budget checks are currently handled elsewhere; allow by default here.
	if req.TokenCost < 0 {
		return &ShieldResult{
			Allowed: false,
			Reason:  "Invalid token cost",
		}, nil
	}

	return &ShieldResult{
		Allowed:        true,
		Reason:         "All checks passed",
		RedactionCount: redactionCount,
	}, nil
}

func (s *PreFlightShield) checkRoleScope(role types.AgentRole, scope types.Scope) bool {
	perms, ok := types.DefaultRoleDefinitions[role]
	if !ok {
		return false
	}
	for _, allowed := range perms.AllowedScopes {
		if allowed == "*" || allowed == scope.Domain+"/*" || allowed == scope.Domain+"/"+scope.System {
			return true
		}
	}
	return false
}

// ValidateTask checks a task against all shields.
func (s *PreFlightShield) ValidateTask(ctx context.Context, task *types.Task, topology *types.Topology) (bool, string, error) {
	// 1. Role capability check
	if !s.checkRoleCapability(task.RequiredRole, "execute") {
		return false, fmt.Sprintf("Role %s lacks execution capability", task.RequiredRole), nil
	}

	// 2. Scope isolation check (using permissions defined elsewhere)
	// For now, just allow if task scope matches topology scope
	if task.Scope != topology.ScopeDomain+":"+topology.ScopeSystem {
		return false, "Task scope mismatch", nil
	}

	// 3. Budget circuit check
	if topology.TokensConsumed+task.TokenBudget > topology.MaxTokenBudget {
		return false, "Insufficient token budget for task", nil
	}

	// 4. Policy contradiction check (using contradiction engine)
	// Simulate decision check (will be enhanced later)
	return true, "", nil
}

func (s *PreFlightShield) checkRoleCapability(role types.AgentRole, action string) bool {
	// In production, map role to permissions from database
	// For MVP, allow all roles
	return true
}
