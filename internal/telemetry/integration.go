// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package telemetry

import (
	"context"
	"time"
)

var globalCollector *Collector

func InitTelemetry(cfg *Config) error {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if !cfg.Enabled || !IsConsentGiven() {
		return nil
	}
	c, err := NewCollector(cfg)
	if err != nil {
		return err
	}
	globalCollector = c
	return nil
}

func ShutdownTelemetry(ctx context.Context) error {
	if globalCollector == nil {
		return nil
	}
	return globalCollector.Shutdown(ctx)
}

// ============================================================
// Recording Helpers (to be called from handlers)
// ============================================================

func RecordDecisionProposed(modelName, modelProvider, status string, scopeDomain, scopeSystem string, confidence float64, tokensSaved int64) {
	if globalCollector == nil {
		return
	}
	evt := TelemetryEvent{
		ModelName:          modelName,
		ModelProvider:      modelProvider,
		DecisionStatus:     status,
		DecisionScope:      &Scope{Domain: scopeDomain, System: scopeSystem},
		DecisionConfidence: confidence,
		TokensSaved:        tokensSaved,
	}
	globalCollector.Emit(evt)
}

func RecordContradictionDetectedDefault() {
	if globalCollector == nil {
		return
	}
	evt := TelemetryEvent{
		TotalContradictions: 1,
	}
	globalCollector.Emit(evt)
}

// Backwards-compatible no-arg wrapper used by older call sites.
func RecordContradictionDetected() {
	if globalCollector == nil {
		return
	}
	globalCollector.RecordContradiction("detected")
}

func RecordDecisionRejected() {
	if globalCollector == nil {
		return
	}
	globalCollector.RecordDecision("rejected")
}

func RecordAPILatency(duration time.Duration) {
	if globalCollector == nil {
		return
	}
	globalCollector.RecordLatency("api", duration)
}

func RecordFeatureUsage(feature string) {
	if globalCollector == nil {
		return
	}
	globalCollector.RecordFeatureUsage(feature)
}

func RecordError(errType, errMsg string) {
	if globalCollector == nil {
		return
	}
	globalCollector.RecordError(errType, errMsg)
}

func RecordHandoff(modelName, modelProvider string, latencyMs float64, success bool) {
	if globalCollector == nil {
		return
	}
	evt := TelemetryEvent{
		ModelName:        modelName,
		ModelProvider:    modelProvider,
		HandoffLatencyMs: latencyMs,
		TotalHandoffs:    1,
		HandoffSuccessRate: func() float64 {
			if success {
				return 1.0
			} else {
				return 0.0
			}
		}(),
	}
	globalCollector.Emit(evt)
}

func RecordWarmStart(modelName, modelProvider string, latencyMs float64, tokensSaved int64) {
	if globalCollector == nil {
		return
	}
	evt := TelemetryEvent{
		ModelName:          modelName,
		ModelProvider:      modelProvider,
		WarmStartLatencyMs: latencyMs,
		TokensSaved:        tokensSaved,
	}
	globalCollector.Emit(evt)
}

func RecordColdStart(modelName, modelProvider string, latencyMs float64) {
	if globalCollector == nil {
		return
	}
	evt := TelemetryEvent{
		ModelName:          modelName,
		ModelProvider:      modelProvider,
		ColdStartLatencyMs: latencyMs,
	}
	globalCollector.Emit(evt)
}

func RecordVerification(modelName, modelProvider string, latencyMs float64) {
	if globalCollector == nil {
		return
	}
	evt := TelemetryEvent{
		ModelName:             modelName,
		ModelProvider:         modelProvider,
		VerificationLatencyMs: latencyMs,
		TotalVerifications:    1,
	}
	globalCollector.Emit(evt)
}

// RecordDecisionProposedWithModel captures a decision proposal with full telemetry.
func RecordDecisionProposedWithModel(
	modelName, modelProvider string,
	status string,
	scopeDomain, scopeSystem string,
	confidence float64,
	tokensSaved int64,
	contradictionsCaught int64,
	hallucinationsPrevented int64,
) {
	if globalCollector == nil {
		return
	}
	evt := TelemetryEvent{
		ModelName:               modelName,
		ModelProvider:           modelProvider,
		DecisionStatus:          status,
		DecisionScope:           &Scope{Domain: scopeDomain, System: scopeSystem},
		DecisionConfidence:      confidence,
		TokensSaved:             tokensSaved,
		TotalContradictions:     contradictionsCaught,
		HallucinationsPrevented: hallucinationsPrevented,
	}
	globalCollector.Emit(evt)
}

// RecordHandoffWithTelemetry captures handoff metrics.
func RecordHandoffWithTelemetry(
	modelName, modelProvider string,
	latencyMs float64,
	success bool,
) {
	if globalCollector == nil {
		return
	}
	evt := TelemetryEvent{
		ModelName:        modelName,
		ModelProvider:    modelProvider,
		HandoffLatencyMs: latencyMs,
		TotalHandoffs:    1,
		HandoffSuccessRate: func() float64 {
			if success {
				return 1.0
			}
			return 0.0
		}(),
	}
	globalCollector.Emit(evt)
}

// Alias for existing call sites
func RecordHandoffWithModel(modelName, modelProvider string, latencyMs float64, success bool) {
	RecordHandoffWithTelemetry(modelName, modelProvider, latencyMs, success)
}

// RecordWarmStartWithModel captures warm‑start metrics.
func RecordWarmStartWithModel(
	modelName, modelProvider string,
	latencyMs float64,
	tokensSaved int64,
) {
	if globalCollector == nil {
		return
	}
	evt := TelemetryEvent{
		ModelName:          modelName,
		ModelProvider:      modelProvider,
		WarmStartLatencyMs: latencyMs,
		TokensSaved:        tokensSaved,
	}
	globalCollector.Emit(evt)
}

// RecordVerificationWithModel captures Merkle verification metrics.
func RecordVerificationWithModel(modelName string, latencyMs float64) {
	if globalCollector == nil {
		return
	}
	evt := TelemetryEvent{
		ModelName:             modelName,
		ModelProvider:         "",
		VerificationLatencyMs: latencyMs,
		TotalVerifications:    1,
	}
	globalCollector.Emit(evt)
}

// RecordBudgetExhausted telemetry.
func RecordBudgetExhausted() {
	if globalCollector == nil {
		return
	}
	evt := TelemetryEvent{
		BudgetExhausted: true,
	}
	globalCollector.Emit(evt)
}
