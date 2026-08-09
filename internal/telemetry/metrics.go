package telemetry

import (
	"time"
)

// TelemetryEvent represents a single anonymous metric payload.
type TelemetryEvent struct {
	// Instance & session
	InstanceHash  string `json:"instance_hash"`
	SessionID     string `json:"session_id"`
	Mode          string `json:"mode"` // active / passive
	GarudaVersion string `json:"garuda_version"`
	AgentRuntime  string `json:"agent_runtime"`

	// Decisions
	DecisionStatus        string  `json:"decision_status,omitempty"`
	DecisionScope         *Scope  `json:"decision_scope,omitempty"`
	DecisionConfidence    float64 `json:"decision_confidence,omitempty"`
	ContradictionResolved *bool   `json:"contradiction_resolved,omitempty"`

	// Model
	ModelProvider string `json:"model_provider,omitempty"`
	ModelName     string `json:"model_name,omitempty"`
	ModelRoute    string `json:"model_route,omitempty"`

	// Cost & Savings
	TokensEstimated int64   `json:"tokens_estimated,omitempty"`
	TokensSaved     int64   `json:"tokens_saved,omitempty"`
	CostSavedUSD    float64 `json:"cost_saved_usd,omitempty"`
	BudgetRemaining int64   `json:"budget_remaining,omitempty"`

	// Performance
	ColdStartLatencyMs    float64 `json:"cold_start_latency_ms,omitempty"`
	WarmStartLatencyMs    float64 `json:"warm_start_latency_ms,omitempty"`
	HandoffLatencyMs      float64 `json:"handoff_latency_ms,omitempty"`
	VerificationLatencyMs float64 `json:"verification_latency_ms,omitempty"`

	// Usage
	ActiveAgents        int   `json:"active_agents,omitempty"`
	TotalHandoffs       int64 `json:"total_handoffs,omitempty"`
	TotalContradictions int64 `json:"total_contradictions,omitempty"`
	TotalDecisions      int64 `json:"total_decisions,omitempty"`
	TotalVerifications  int64 `json:"total_verifications,omitempty"`
	BudgetExhausted     bool  `json:"budget_exhausted,omitempty"`

	// Coordination
	HandoffSuccessRate         float64 `json:"handoff_success_rate,omitempty"`
	ContradictionReductionRate float64 `json:"contradiction_reduction_rate,omitempty"`
	TokenReuseRate             float64 `json:"token_reuse_rate,omitempty"`
	DuplicateWorkReduction     float64 `json:"duplicate_work_reduction,omitempty"`
	CoordinationScore          float64 `json:"coordination_score,omitempty"`

	// Hallucinations
	HallucinationsPrevented        int64   `json:"hallucinations_prevented,omitempty"`
	HallucinationReductionPerModel float64 `json:"hallucination_reduction_per_model,omitempty"`

	// Timestamp
	Timestamp time.Time `json:"timestamp"`
}

type Scope struct {
	Domain string `json:"domain"`
	System string `json:"system"`
}
