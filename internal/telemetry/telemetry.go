package telemetry

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"time"
)

// ============================================================
// Configuration Helpers
// ============================================================

// LoadConfigFromEnv loads telemetry config from environment variables.
func LoadConfigFromEnv() *Config {
	cfg := DefaultConfig()

	if v := os.Getenv("GARUDA_TELEMETRY_ENABLED"); v != "" {
		cfg.Enabled = v == "true" || v == "1"
	}
	if v := os.Getenv("GARUDA_TELEMETRY_ENDPOINT"); v != "" {
		cfg.Endpoint = v
	}
	if v := os.Getenv("GARUDA_TELEMETRY_BATCH_SIZE"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.BatchSize)
	}
	if v := os.Getenv("GARUDA_TELEMETRY_FLUSH_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.FlushInterval = d
		}
	}
	if v := os.Getenv("GARUDA_TELEMETRY_SALT"); v != "" {
		cfg.Salt = v
	}

	return cfg
}

// ============================================================
// Metrics Types
// ============================================================

// DecisionMetrics tracks decision-related operations.
type DecisionMetrics struct {
	Proposed int64 `json:"proposed"`
	Reused   int64 `json:"reused"`
	Rejected int64 `json:"rejected"`
	Stale    int64 `json:"stale"`
}

// ContradictionMetrics tracks contradiction detection.
type ContradictionMetrics struct {
	Detected int64 `json:"detected"`
	Resolved int64 `json:"resolved"`
	FalsePos int64 `json:"false_pos"`
}

// PerformanceMetrics tracks latency and performance.
type PerformanceMetrics struct {
	ColdStartLatencyP50 float64 `json:"cold_start_latency_p50"`
	ColdStartLatencyP95 float64 `json:"cold_start_latency_p95"`
	ColdStartLatencyP99 float64 `json:"cold_start_latency_p99"`
	WarmStartLatencyP50 float64 `json:"warm_start_latency_p50"`
	WarmStartLatencyP95 float64 `json:"warm_start_latency_p95"`
	WarmStartLatencyP99 float64 `json:"warm_start_latency_p99"`
	APILatencyP50       float64 `json:"api_latency_p50"`
	APILatencyP95       float64 `json:"api_latency_p95"`
	APILatencyP99       float64 `json:"api_latency_p99"`
}

// CostMetrics tracks token and cost savings.
type CostMetrics struct {
	TokensSaved   int64   `json:"tokens_saved"`
	EstimatedCost float64 `json:"estimated_cost"`
	ModelType     string  `json:"model_type"`
}

// TelemetryPayload is the complete metrics payload sent to the telemetry server.
type TelemetryPayload struct {
	Version              string               `json:"version"`
	TenantHash           string               `json:"tenant_hash"`
	SessionID            string               `json:"session_id"`
	Timestamp            time.Time            `json:"timestamp"`
	DecisionMetrics      DecisionMetrics      `json:"decisions"`
	ContradictionMetrics ContradictionMetrics `json:"contradictions"`
	Performance          PerformanceMetrics   `json:"performance"`
	CostMetrics          CostMetrics          `json:"cost"`
	FeatureUsage         map[string]int64     `json:"feature_usage"`
	Errors               []ErrorEntry         `json:"errors,omitempty"`
}

// ErrorEntry captures a single error event.
type ErrorEntry struct {
	Type    string `json:"type"`
	Count   int64  `json:"count"`
	Message string `json:"message,omitempty"`
}

// ============================================================
// Metric Recording Methods
// ============================================================

// RecordDecision records a decision event.
func (c *Collector) RecordDecision(status string) {
	if !c.cfg.Enabled {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	switch status {
	case "proposed":
		c.decisions.Proposed++
	case "reused":
		c.decisions.Reused++
	case "rejected":
		c.decisions.Rejected++
	case "stale":
		c.decisions.Stale++
	}
}

// RecordContradiction records a contradiction event.
func (c *Collector) RecordContradiction(status string) {
	if !c.cfg.Enabled {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	switch status {
	case "detected":
		c.contradictions.Detected++
	case "resolved":
		c.contradictions.Resolved++
	case "false_positive":
		c.contradictions.FalsePos++
	}
}

// RecordLatency records a latency measurement.
func (c *Collector) RecordLatency(latencyType string, d time.Duration) {
	if !c.cfg.Enabled {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	ms := float64(d.Milliseconds())
	switch latencyType {
	case "cold_start":
		c.coldStartLatencies = append(c.coldStartLatencies, ms)
	case "warm_start":
		c.warmStartLatencies = append(c.warmStartLatencies, ms)
	case "api":
		c.apiLatencies = append(c.apiLatencies, ms)
	}
}

// RecordCostSaving records token and cost savings.
func (c *Collector) RecordCostSaving(tokens int64, model string) {
	if !c.cfg.Enabled {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cost.TokensSaved += tokens
	c.cost.ModelType = model
	// Estimate cost: $0.002 per 1K tokens for GPT-4 baseline
	c.cost.EstimatedCost += float64(tokens) * 0.000002
}

// RecordFeatureUsage tracks which features are being used.
func (c *Collector) RecordFeatureUsage(feature string) {
	if !c.cfg.Enabled {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.featureUsage[feature]++
}

// RecordError tracks errors by type.
func (c *Collector) RecordError(errType, errMsg string) {
	if !c.cfg.Enabled {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errors[errType]++
}

// ============================================================
// Batch & Flush Logic
// ============================================================

// backgroundFlusher runs in a goroutine and flushes metrics periodically.
func (c *Collector) backgroundFlusher() {
	defer c.wg.Done()

	for {
		select {
		case <-c.flushTimer.C:
			if err := c.Flush(); err != nil {
				slog.Error("telemetry flush failed", "error", err)
			}
		case <-c.shutdown:
			// Final flush before exit
			if err := c.Flush(); err != nil {
				slog.Error("final telemetry flush failed", "error", err)
			}
			return
		}
	}
}

// Flush sends the current batch of telemetry data to the endpoint.
func (c *Collector) Flush() error {
	if !c.cfg.Enabled {
		return nil
	}

	c.mu.Lock()

	// Build an aggregated TelemetryEvent from internal counters
	perf := c.calculatePercentiles()
	totalDecisions := c.decisions.Proposed + c.decisions.Reused + c.decisions.Rejected + c.decisions.Stale
	totalContradictions := c.contradictions.Detected + c.contradictions.Resolved + c.contradictions.FalsePos

	event := TelemetryEvent{
		InstanceHash:          c.instanceHash,
		SessionID:             c.sessionID,
		Mode:                  c.cfg.Mode,
		GarudaVersion:         "1.0.0",
		TotalDecisions:        totalDecisions,
		TotalContradictions:   totalContradictions,
		TokensSaved:           c.cost.TokensSaved,
		CostSavedUSD:          c.cost.EstimatedCost,
		ColdStartLatencyMs:    perf.ColdStartLatencyP95,
		WarmStartLatencyMs:    perf.WarmStartLatencyP95,
		VerificationLatencyMs: perf.APILatencyP95,
	}

	// Add to batch of events to send
	c.batch = append(c.batch, event)

	// Reset sampled data after flush
	c.coldStartLatencies = nil
	c.warmStartLatencies = nil
	c.apiLatencies = nil
	c.errors = make(map[string]int64)
	c.featureUsage = make(map[string]int64)
	c.decisions = DecisionMetrics{}
	c.contradictions = ContradictionMetrics{}
	c.cost = CostMetrics{}

	// If batch is full, trigger sending batch
	if len(c.batch) >= c.cfg.BatchSize {
		c.mu.Unlock()
		return c.flushBatch()
	}

	c.mu.Unlock()
	return nil
}

// flushBatch sends all batched payloads to the endpoint safely.
func (c *Collector) flushBatch() error {
	// Delegate to collector's sendBatch which handles []TelemetryEvent
	c.mu.Lock()
	batch := c.batch
	c.batch = make([]TelemetryEvent, 0, c.cfg.BatchSize)
	c.mu.Unlock()

	if len(batch) == 0 {
		return nil
	}

	if err := c.sendBatch(batch); err != nil {
		return fmt.Errorf("failed to send telemetry batch: %w", err)
	}
	slog.Debug("telemetry batch sent", "count", len(batch))
	return nil
}

// ============================================================
// Helper Methods
// ============================================================

// anonymizeTenant hashes the tenant ID to prevent PII leakage.
func (c *Collector) anonymizeTenant() string {
	tenantID := os.Getenv("GARUDA_TENANT_ID")
	if tenantID == "" {
		return "anonymous"
	}
	hash := sha256.Sum256([]byte(tenantID + c.cfg.Salt))
	return base64.StdEncoding.EncodeToString(hash[:8])
}

// anonymizeSessionID hashes the session ID.
func (c *Collector) anonymizeSessionID() string {
	sessionID := os.Getenv("GARUDA_SESSION_ID")
	if sessionID == "" {
		return "anonymous"
	}
	hash := sha256.Sum256([]byte(sessionID + c.cfg.Salt))
	return base64.StdEncoding.EncodeToString(hash[:8])
}

// calculatePercentiles computes p50, p95, p99 from latency samples.
func (c *Collector) calculatePercentiles() PerformanceMetrics {
	return PerformanceMetrics{
		ColdStartLatencyP50: c.percentile(c.coldStartLatencies, 0.50),
		ColdStartLatencyP95: c.percentile(c.coldStartLatencies, 0.95),
		ColdStartLatencyP99: c.percentile(c.coldStartLatencies, 0.99),
		WarmStartLatencyP50: c.percentile(c.warmStartLatencies, 0.50),
		WarmStartLatencyP95: c.percentile(c.warmStartLatencies, 0.95),
		WarmStartLatencyP99: c.percentile(c.warmStartLatencies, 0.99),
		APILatencyP50:       c.percentile(c.apiLatencies, 0.50),
		APILatencyP95:       c.percentile(c.apiLatencies, 0.95),
		APILatencyP99:       c.percentile(c.apiLatencies, 0.99),
	}
}

func (c *Collector) percentile(samples []float64, p float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	sorted := make([]float64, len(samples))
	copy(sorted, samples)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	idx := int(float64(len(sorted)) * p)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func (c *Collector) buildErrorEntries() []ErrorEntry {
	entries := make([]ErrorEntry, 0, len(c.errors))
	for errType, count := range c.errors {
		entries = append(entries, ErrorEntry{
			Type:  errType,
			Count: count,
		})
	}
	return entries
}
