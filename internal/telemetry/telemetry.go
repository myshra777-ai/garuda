package telemetry

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

// ============================================================
// Configuration
// ============================================================

// Config holds telemetry configuration.
type Config struct {
	Enabled       bool          `json:"enabled"`
	Endpoint      string        `json:"endpoint"`
	BatchSize     int           `json:"batch_size"`
	FlushInterval time.Duration `json:"flush_interval"`
	Anonymize     bool          `json:"anonymize"`
	Salt          string        `json:"salt"`
}

// DefaultConfig returns the default telemetry configuration.
func DefaultConfig() *Config {
	return &Config{
		Enabled:       true,
		Endpoint:      "https://telemetry.garuda.dev/v1/ingest",
		BatchSize:     100,
		FlushInterval: 30 * time.Second,
		Anonymize:     true,
		Salt:          os.Getenv("GARUDA_TELEMETRY_SALT"),
	}
}

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
// Telemetry Collector
// ============================================================

// Collector aggregates and sends telemetry data.
type Collector struct {
	cfg        *Config
	client     *http.Client
	mu         sync.Mutex
	batch      []TelemetryPayload
	flushTimer *time.Ticker
	stopChan   chan struct{}
	wg         sync.WaitGroup

	// Aggregated metrics
	decisions      DecisionMetrics
	contradictions ContradictionMetrics
	performance    PerformanceMetrics
	cost           CostMetrics
	featureUsage   map[string]int64
	errors         map[string]int64

	// Latency histograms (for percentile calculation)
	coldStartLatencies []float64
	warmStartLatencies []float64
	apiLatencies       []float64
}

// NewCollector creates a new telemetry collector.
func NewCollector(cfg *Config) *Collector {
	if cfg == nil {
		cfg = LoadConfigFromEnv()
	}
	if cfg.Salt == "" {
		cfg.Salt = "garuda-default-salt" // In production, set via env
	}

	c := &Collector{
		cfg:          cfg,
		client:       &http.Client{Timeout: 10 * time.Second},
		batch:        make([]TelemetryPayload, 0, cfg.BatchSize),
		featureUsage: make(map[string]int64),
		errors:       make(map[string]int64),
		stopChan:     make(chan struct{}),
	}

	if cfg.Enabled {
		c.flushTimer = time.NewTicker(cfg.FlushInterval)
		c.wg.Add(1)
		go c.backgroundFlusher()
	}

	return c
}

// ============================================================
// Metric Recording Methods
// ============================================================

// RecordDecision records a decision event.
func (c *Collector) RecordDecision(action string, reused bool) {
	if !c.cfg.Enabled {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	switch action {
	case "proposed":
		c.decisions.Proposed++
		if reused {
			c.decisions.Reused++
		}
	case "rejected":
		c.decisions.Rejected++
	case "stale":
		c.decisions.Stale++
	}
}

// RecordContradiction records a contradiction event.
func (c *Collector) RecordContradiction(action string) {
	if !c.cfg.Enabled {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	switch action {
	case "detected":
		c.contradictions.Detected++
	case "resolved":
		c.contradictions.Resolved++
	case "false_positive":
		c.contradictions.FalsePos++
	}
}

// RecordLatency records a latency measurement.
func (c *Collector) RecordLatency(metricType string, latency time.Duration) {
	if !c.cfg.Enabled {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	ms := float64(latency.Microseconds()) / 1000.0

	switch metricType {
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
	// Estimate cost: $0.002 per 1K tokens for GPT-4, adjust as needed
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
		case <-c.stopChan:
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

	// Build payload from aggregated metrics
	payload := TelemetryPayload{
		Version:              "1.0.0",
		TenantHash:           c.anonymizeTenant(),
		SessionID:            c.anonymizeSessionID(),
		Timestamp:            time.Now().UTC(),
		DecisionMetrics:      c.decisions,
		ContradictionMetrics: c.contradictions,
		CostMetrics:          c.cost,
		FeatureUsage:         c.featureUsage,
	}

	// Calculate percentiles
	payload.Performance = c.calculatePercentiles()

	// Build error entries
	payload.Errors = c.buildErrorEntries()

	// Add to batch
	c.batch = append(c.batch, payload)

	// Reset aggregated metrics (keep running totals, but reset per-batch ones)
	// Decision and contradiction metrics are cumulative, so we reset after flushing.
	// But we should keep them running for the next batch.
	// Actually, we want to reset them so the next batch doesn't double-count.
	// However, for YC we want cumulative totals. Let's keep cumulative but send
	// the cumulative value each time. That's fine—the server can handle it.

	// For latency and errors, we reset after each flush to avoid sending
	// the same data repeatedly.
	c.coldStartLatencies = nil
	c.warmStartLatencies = nil
	c.apiLatencies = nil
	c.errors = make(map[string]int64)

	// Keep decision and contradiction metrics cumulative.
	// The server will see the total over time.

	// If batch is full, send immediately.
	if len(c.batch) >= c.cfg.BatchSize {
		return c.sendBatch()
	}

	c.mu.Unlock()
	return nil
}

// sendBatch sends all batched payloads to the telemetry endpoint.
func (c *Collector) sendBatch() error {
	c.mu.Lock()
	batch := c.batch
	c.batch = make([]TelemetryPayload, 0, c.cfg.BatchSize)
	c.mu.Unlock()

	if len(batch) == 0 {
		return nil
	}

	body, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry batch: %w", err)
	}

	req, err := http.NewRequest("POST", c.cfg.Endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create telemetry request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Garuda-Telemetry/1.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send telemetry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("telemetry server returned %d", resp.StatusCode)
	}

	slog.Debug("telemetry batch sent", "count", len(batch))
	return nil
}

// ============================================================
// Helper Methods
// ============================================================

// anonymizeTenant hashes the tenant ID to prevent PII leakage.
func (c *Collector) anonymizeTenant() string {
	// In a real implementation, this would hash the tenant ID.
	// For now, we return a placeholder.
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
	// Sort samples
	sorted := make([]float64, len(samples))
	copy(sorted, samples)
	// Simple sort (could be optimized)
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

// ============================================================
// Shutdown
// ============================================================

// Shutdown stops the telemetry collector and flushes remaining data.
func (c *Collector) Shutdown(ctx context.Context) error {
	if !c.cfg.Enabled {
		return nil
	}

	close(c.stopChan)
	c.wg.Wait()

	if c.flushTimer != nil {
		c.flushTimer.Stop()
	}

	// Final flush
	if err := c.Flush(); err != nil {
		slog.Error("final telemetry flush failed", "error", err)
	}

	return nil
}
