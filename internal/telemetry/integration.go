package telemetry

import (
	"context"
	"time"
)

// Global unexported pointer instance tracking the telemetry collection runtime.
var globalCollector *Collector

// InitTelemetry initializes the global telemetry collector instance.
func InitTelemetry(cfg *Config) {
	globalCollector = NewCollector(cfg)
}

// ShutdownTelemetry drains remaining data buffers and gracefully shuts down the telemetry loop.
func ShutdownTelemetry(ctx context.Context) error {
	if globalCollector == nil {
		return nil
	}
	return globalCollector.Shutdown(ctx)
}

// ============================================================
// Handler Wrappers
// ============================================================

// RecordDecisionProposed tracks an incoming proposed architectural decision path.
func RecordDecisionProposed(reused bool) {
	if globalCollector == nil {
		return
	}
	globalCollector.RecordDecision("proposed", reused)
}

// RecordDecisionRejected records a structural decision evaluation failure.
func RecordDecisionRejected() {
	if globalCollector == nil {
		return
	}
	globalCollector.RecordDecision("rejected", false)
}

// RecordDecisionStale flags when a historic governance decision falls out of compliance context.
func RecordDecisionStale() {
	if globalCollector == nil {
		return
	}
	globalCollector.RecordDecision("stale", false)
}

// RecordContradictionDetected signals when conflicting operational rules intersect.
func RecordContradictionDetected() {
	if globalCollector == nil {
		return
	}
	globalCollector.RecordContradiction("detected")
}

// RecordContradictionResolved logs the successful mitigation of an operational contradiction.
func RecordContradictionResolved() {
	if globalCollector == nil {
		return
	}
	globalCollector.RecordContradiction("resolved")
}

// RecordContradictionFalsePositive logs when an automated warning check evaluates as clean.
func RecordContradictionFalsePositive() {
	if globalCollector == nil {
		return
	}
	globalCollector.RecordContradiction("false_positive")
}

// RecordColdStartLatency records execution latency spikes on fresh container spin-ups.
func RecordColdStartLatency(d time.Duration) {
	if globalCollector == nil {
		return
	}
	globalCollector.RecordLatency("cold_start", d)
}

// RecordWarmStartLatency records execution latency windows for cached engine instances.
func RecordWarmStartLatency(d time.Duration) {
	if globalCollector == nil {
		return
	}
	globalCollector.RecordLatency("warm_start", d)
}

// RecordAPILatency measures standard end-to-end network handler response timelines.
func RecordAPILatency(d time.Duration) {
	if globalCollector == nil {
		return
	}
	globalCollector.RecordLatency("api", d)
}

// RecordCostSaving evaluates token budget efficiencies against targeted model types.
func RecordCostSaving(tokens int64, model string) {
	if globalCollector == nil {
		return
	}
	globalCollector.RecordCostSaving(tokens, model)
}

// RecordFeatureUsage increments tracking counters on internal system functions.
func RecordFeatureUsage(feature string) {
	if globalCollector == nil {
		return
	}
	globalCollector.RecordFeatureUsage(feature)
}

// RecordError pipes internal exception tracking out to telemetry ingestion loops.
func RecordError(errType, errMsg string) {
	if globalCollector == nil {
		return
	}
	globalCollector.RecordError(errType, errMsg)
}
