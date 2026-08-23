// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package benchmark

import "time"

type BenchmarkMode string

const (
	ModeNaive     BenchmarkMode = "NAIVE_UNASSISTED"
	ModeGarudaMCP BenchmarkMode = "GARUDA_MCP_GROUNDED"
)

type GroundTruthTarget struct {
	Symbol              string   `json:"symbol"`
	Package             string   `json:"package"`
	ExpectedUpstream    []string `json:"expected_upstream"`
	ExpectedDownstream  []string `json:"expected_downstream"`
	ForbiddenEndpoints  []string `json:"forbidden_endpoints"`
	KnownContradictions []string `json:"known_contradictions"`
}

type BenchmarkTask struct {
	ID          string            `json:"id"`
	Category    string            `json:"category"`
	Title       string            `json:"title"`
	Prompt      string            `json:"prompt"`
	Target      GroundTruthTarget `json:"target"`
}

type TaskExecutionResult struct {
	TaskID                 string        `json:"task_id"`
	Title                  string        `json:"title"`
	Mode                   BenchmarkMode `json:"mode"`
	HallucinationDetected  bool          `json:"hallucination_detected"`
	HallucinatedSymbols    []string      `json:"hallucinated_symbols"`
	ViolationsQuarantined  bool          `json:"violations_quarantined"`
	UpstreamRecall         float64       `json:"upstream_recall"`
	DownstreamRecall       float64       `json:"downstream_recall"`
	PrecisionScore         float64       `json:"precision_score"`
	ContextSizeTokens      int           `json:"context_size_tokens"`
	DurationMs             int64         `json:"duration_ms"`
}

type BenchmarkSummaryReport struct {
	Timestamp            time.Time             `json:"timestamp"`
	Workspace            string                `json:"workspace"`
	TotalTasks           int                   `json:"total_tasks"`
	NaiveMetrics         AggregateMetrics      `json:"naive_metrics"`
	GarudaMCPMetrics     AggregateMetrics      `json:"garuda_mcp_metrics"`
	ImprovementFactor    map[string]float64    `json:"improvement_factor"`
	DetailedTaskResults  []TaskExecutionResult `json:"detailed_task_results"`
}

type AggregateMetrics struct {
	AvgPrecision        float64 `json:"avg_precision"`
	AvgUpstreamRecall   float64 `json:"avg_upstream_recall"`
	AvgDownstreamRecall float64 `json:"avg_downstream_recall"`
	HallucinationRate   float64 `json:"hallucination_rate"`
	ViolationCatchRate  float64 `json:"violation_catch_rate"`
	AvgContextTokens    int     `json:"avg_context_tokens"`
}
