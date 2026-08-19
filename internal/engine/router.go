// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package engine

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/myshra777-ai/garuda/internal/types"
)

type DynamicRouter struct {
	shield *ShieldEngine
	rng    *rand.Rand
}

func NewDynamicRouter() *DynamicRouter {
	return &DynamicRouter{
		shield: NewShieldEngine(),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ClassifyAndRoute evaluates payload context, executes redaction, and dispatches target models.
func (r *DynamicRouter) ClassifyAndRoute(ctx context.Context, payload string, tokenEstimate int, spendRatio float64) (*types.RoutingDecision, error) {
	cleanPayload, redactedCount := r.shield.Redact(payload)
	domain := r.classifyDomain(cleanPayload, tokenEstimate)
	budgetShifted := spendRatio >= 0.85
	primary, fallback := r.selectModels(domain, budgetShifted)

	riskyPayload := strings.Contains(strings.ToLower(cleanPayload), "drop ") || strings.Contains(strings.ToLower(cleanPayload), "delete ") || strings.Contains(strings.ToLower(cleanPayload), "truncate ")
	shadowExecuted := riskyPayload || r.rng.Float64() <= 0.01
	consensusRequired := domain == types.DomainCodeGov && riskyPayload

	return &types.RoutingDecision{
		Domain:            domain,
		SelectedModel:     primary,
		FallbackModel:     fallback,
		BudgetShifted:     budgetShifted,
		ShadowExecuted:    shadowExecuted,
		ConsensusRequired: consensusRequired,
		CleanPayload:      cleanPayload,
		RedactedCount:     redactedCount,
		Timestamp:         time.Now().UTC(),
	}, nil
}

func (r *DynamicRouter) classifyDomain(payload string, tokenEstimate int) types.TaskDomain {
	if tokenEstimate > 100000 {
		return types.DomainLargeRAG
	}

	lower := strings.ToLower(payload)
	if strings.Contains(lower, "proof") || strings.Contains(lower, "theorem") || strings.Contains(lower, "math") || strings.Contains(lower, "ltl") {
		return types.DomainLogic
	}

	if strings.Contains(lower, "func ") || strings.Contains(lower, "sql") || strings.Contains(lower, "refactor") || strings.Contains(lower, "class ") || strings.Contains(lower, "drop ") || strings.Contains(lower, "delete ") || strings.Contains(lower, "truncate ") {
		return types.DomainCodeGov
	}

	return types.DomainRoutine
}

func (r *DynamicRouter) selectModels(domain types.TaskDomain, budgetShifted bool) (string, string) {
	switch domain {
	case types.DomainCodeGov:
		if budgetShifted {
			return "deepseek-v3", "local-ollama"
		}
		return "claude-3-5-sonnet", "deepseek-v3"

	case types.DomainLogic:
		if budgetShifted {
			return "claude-3-5-sonnet", "deepseek-v3"
		}
		return "deepseek-r1", "claude-3-5-sonnet"

	case types.DomainLargeRAG:
		if budgetShifted {
			return "claude-3-5-sonnet", "gemini-1-5-flash"
		}
		return "gemini-1-5-pro", "claude-3-5-sonnet"

	case types.DomainRoutine:
		return "local-ollama", "gemini-1-5-flash"

	default:
		return "local-ollama", "gemini-1-5-flash"
	}
}
