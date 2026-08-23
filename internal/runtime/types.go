// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package runtime

import (
	"time"

	"github.com/google/uuid"
)

type VerificationStatus string

const (
	StatusSupported    VerificationStatus = "SUPPORTED"
	StatusUnverified   VerificationStatus = "UNVERIFIED"
	StatusContradicted VerificationStatus = "CONTRADICTED"
)

type VerificationReason string

const (
	ReasonDirectExecutionMatch   VerificationReason = "DIRECT_EXECUTION_MATCH"
	ReasonNoRuntimeObservation   VerificationReason = "NO_RUNTIME_OBSERVATION"
	ReasonLowConfidenceMatch     VerificationReason = "LOW_CONFIDENCE_MATCH"
	ReasonArchitecturalDeviation VerificationReason = "ARCHITECTURAL_DEVIATION"
	ReasonQuarantinedTarget      VerificationReason = "QUARANTINED_TARGET"
)

type RuntimeObservation struct {
	ID           uuid.UUID              `json:"id"`
	WorkspaceID  uuid.UUID              `json:"workspace_id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	TraceID      string                 `json:"trace_id"`
	SpanID       string                 `json:"span_id"`
	ParentSpanID string                 `json:"parent_span_id"`
	ServiceName  string                 `json:"service_name"`
	Operation    string                 `json:"operation"`
	EntityID     *uuid.UUID             `json:"entity_id,omitempty"`
	StartedAt    time.Time              `json:"started_at"`
	DurationMs   float64                `json:"duration_ms"`
	StatusCode   string                 `json:"status_code"`
	Source       string                 `json:"source"`
	Confidence   float64                `json:"confidence"`
	Attributes   map[string]interface{} `json:"attributes"`
}

type CorrelationResult struct {
	EntityID   *uuid.UUID
	Confidence float64
	Strategy   string
}

type VerificationRecord struct {
	SourceEntityID       uuid.UUID          `json:"source_entity_id"`
	TargetEntityID       *uuid.UUID         `json:"target_entity_id,omitempty"`
	SourceSymbol         string             `json:"source_symbol"`
	TargetSymbol         string             `json:"target_symbol"`
	Status               VerificationStatus `json:"status"`
	Reason               VerificationReason `json:"reason"`
	StaticEdgeExists     bool               `json:"static_edge_exists"`
	RuntimeObservedCount int64              `json:"runtime_observed_count"`
	LastTraceID          string             `json:"last_trace_id,omitempty"`
	LastEvaluatedAt      time.Time          `json:"last_evaluated_at"`
}

type RuntimeCoverageSummary struct {
	TotalStaticEntities int64   `json:"total_static_entities"`
	ObservedEntities    int64   `json:"observed_entities"`
	CoveragePercent     float64 `json:"coverage_percent"`
	SupportedCount      int64   `json:"supported_count"`
	UnverifiedCount     int64   `json:"unverified_count"`
	ContradictedCount   int64   `json:"contradicted_count"`
}
