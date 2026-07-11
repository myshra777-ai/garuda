package telemetry

import (
	"context"
	"time"
)

// TelemetryCollector is a global instance.
var Collector *Collector

// InitTelemetry initializes the global telemetry collector.
func InitTelemetry(cfg *Config) {
	Collector = NewCollector(cfg)
}

// ShutdownTelemetry shuts down the global collector.
func ShutdownTelemetry(ctx context.Context) error {
	if Collector == nil {
		return nil
	}
	return Collector.Shutdown(ctx)
}

// ============================================================
// Handler Wrappers
// ============================================================

// RecordDecisionProposed records a decision proposal.
func RecordDecisionProposed(reused bool) {
	if Collector == nil {
		return
	}
	Collector.RecordDecision("proposed", reused)
}

// RecordDecisionRejected records a rejected decision.
func RecordDecisionRejected() {
	if Collector == nil {
		return
	}
	Collector.RecordDecision("rejected", false)
}

// RecordDecisionStale records a decision marked stale.
func RecordDecisionStale() {
	if Collector == nil {
		return
	}
	Collector.RecordDecision("stale", false)
}

// RecordContradictionDetected records a contradiction detection.
func RecordContradictionDetected() {
	if Collector == nil {
		return
	}
	Collector.RecordContradiction("detected")
}

// RecordContradictionResolved records a resolved contradiction.
func RecordContradictionResolved() {
	if Collector == nil {
		return
	}
	Collector.RecordContradiction("resolved")
}

// RecordContradictionFalsePositive records a false positive contradiction.
func RecordContradictionFalsePositive() {
	if Collector == nil {
		return
	}
	Collector.RecordContradiction("false_positive")
}

// RecordColdStartLatency records cold start latency.
func RecordColdStartLatency(d time.Duration) {
	if Collector == nil {
		return
	}
	Collector.RecordLatency("cold_start", d)
}

// RecordWarmStartLatency records warm start latency.
func RecordWarmStartLatency(d time.Duration) {
	if Collector == nil {
		return
	}
	Collector.RecordLatency("warm_start", d)
}

// RecordAPILatency records API call latency.
func RecordAPILatency(d time.Duration) {
	if Collector == nil {
		return
	}
	Collector.RecordLatency("api", d)
}

// RecordCostSaving records token and cost savings.
func RecordCostSaving(tokens int64, model string) {
	if Collector == nil {
		return
	}
	Collector.RecordCostSaving(tokens, model)
}

// RecordFeatureUsage records a feature being used.
func RecordFeatureUsage(feature string) {
	if Collector == nil {
		return
	}
	Collector.RecordFeatureUsage(feature)
}

// RecordError records an error.
func RecordError(errType, errMsg string) {
	if Collector == nil {
		return
	}
	Collector.RecordError(errType, errMsg)
}
