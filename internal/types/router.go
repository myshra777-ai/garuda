// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package types

import "time"

type TaskDomain string

const (
	DomainCodeGov  TaskDomain = "code_governance"
	DomainLogic    TaskDomain = "logic_math"
	DomainLargeRAG TaskDomain = "large_rag"
	DomainRoutine  TaskDomain = "routine_parsing"
)

type RouteTarget struct {
	PrimaryProvider string `json:"primary_provider"`
	FallbackTarget  string `json:"fallback_target"`
	Optimization    string `json:"optimization"`
}

type RoutingDecision struct {
	Domain            TaskDomain `json:"domain"`
	SelectedModel     string     `json:"selected_model"`
	FallbackModel     string     `json:"fallback_model"`
	BudgetShifted     bool       `json:"budget_shifted"`
	ShadowExecuted    bool       `json:"shadow_executed"`
	ConsensusRequired bool       `json:"consensus_required"`
	CleanPayload      string     `json:"clean_payload"`
	RedactedCount     int        `json:"redacted_count"`
	Timestamp         time.Time  `json:"timestamp"`
}
