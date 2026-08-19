// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

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

type Config struct {
	Enabled       bool
	Endpoint      string
	BatchSize     int
	FlushInterval time.Duration
	MaxQueueSize  int
	Mode          string // active, passive
	Salt          string
}

func DefaultConfig() *Config {
	return &Config{
		Enabled:       true,
		Endpoint:      os.Getenv("GARUDA_TELEMETRY_ENDPOINT"),
		BatchSize:     100,
		FlushInterval: 5 * time.Minute,
		MaxQueueSize:  10000,
		Mode:          os.Getenv("GARUDA_MODE"),
	}
}

type Collector struct {
	cfg                *Config
	client             *http.Client
	events             chan TelemetryEvent
	batch              []TelemetryEvent
	decisions          DecisionMetrics
	contradictions     ContradictionMetrics
	coldStartLatencies []float64
	warmStartLatencies []float64
	apiLatencies       []float64
	cost               CostMetrics
	featureUsage       map[string]int64
	errors             map[string]int64
	mu                 sync.Mutex
	wg                 sync.WaitGroup
	shutdown           chan struct{}
	flushTimer         *time.Ticker
	instanceHash       string
	sessionID          string
}

func NewCollector(cfg *Config) (*Collector, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if !cfg.Enabled || !IsConsentGiven() {
		slog.Info("Telemetry disabled (opt‑out or no consent)")
		return nil, fmt.Errorf("telemetry disabled")
	}
	if cfg.Endpoint == "" {
		slog.Warn("No telemetry endpoint configured; skipping")
		return nil, fmt.Errorf("no endpoint")
	}

	instanceHash := computeInstanceHash()
	sessionID := generateSessionID()

	c := &Collector{
		cfg:                cfg,
		client:             &http.Client{Timeout: 10 * time.Second},
		events:             make(chan TelemetryEvent, cfg.MaxQueueSize),
		batch:              make([]TelemetryEvent, 0, cfg.BatchSize),
		decisions:          DecisionMetrics{},
		contradictions:     ContradictionMetrics{},
		coldStartLatencies: make([]float64, 0, 0),
		warmStartLatencies: make([]float64, 0, 0),
		apiLatencies:       make([]float64, 0, 0),
		cost:               CostMetrics{},
		featureUsage:       make(map[string]int64),
		errors:             make(map[string]int64),
		shutdown:           make(chan struct{}),
		flushTimer:         time.NewTicker(cfg.FlushInterval),
		instanceHash:       instanceHash,
		sessionID:          sessionID,
	}
	c.wg.Add(1)
	go c.worker()
	return c, nil
}

func (c *Collector) Emit(evt TelemetryEvent) {
	if c.cfg.Mode == "passive" {
		evt.Mode = "passive"
	} else {
		evt.Mode = "active"
	}
	evt.InstanceHash = c.instanceHash
	evt.SessionID = c.sessionID
	evt.Timestamp = time.Now().UTC()
	if evt.GarudaVersion == "" {
		evt.GarudaVersion = "dev"
	}

	select {
	case c.events <- evt:
	default:
		slog.Warn("telemetry queue full, dropping event")
		slog.Debug("Telemetry event emitted", "event", evt)

	}
}

func (c *Collector) worker() {
	defer c.wg.Done()
	for {
		select {
		case <-c.shutdown:
			c.flush()
			return
		case <-c.flushTimer.C:
			c.flush()
		case evt, ok := <-c.events:
			if !ok {
				return
			}
			c.mu.Lock()
			c.batch = append(c.batch, evt)
			if len(c.batch) >= c.cfg.BatchSize {
				c.mu.Unlock()
				c.flush()
			} else {
				c.mu.Unlock()
			}
		}
	}
}

func (c *Collector) flush() {
	c.mu.Lock()
	if len(c.batch) == 0 {
		c.mu.Unlock()
		return
	}
	batch := c.batch
	c.batch = make([]TelemetryEvent, 0, c.cfg.BatchSize)
	c.mu.Unlock()
	_ = c.sendBatch(batch) // fail‑safe: drops on error
}

func (c *Collector) sendBatch(events []TelemetryEvent) error {
	payload := struct {
		Version string           `json:"version"`
		Events  []TelemetryEvent `json:"events"`
		Count   int              `json:"count"`
	}{
		Version: "1.0.0",
		Events:  events,
		Count:   len(events),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.cfg.Endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Garuda-Telemetry/1.0")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}
	return nil
}

func (c *Collector) Shutdown(ctx context.Context) error {
	close(c.shutdown)
	c.wg.Wait()
	return nil
}

func computeInstanceHash() string {
	hostname, _ := os.Hostname()
	cwd, _ := os.Getwd()
	// Machine + project + salt (if provided)
	salt := os.Getenv("GARUDA_TELEMETRY_SALT")
	if salt == "" {
		salt = "garuda-default-salt"
	}
	data := hostname + "|" + cwd + "|" + salt
	hash := sha256.Sum256([]byte(data))
	return base64.StdEncoding.EncodeToString(hash[:16])
}

func generateSessionID() string {
	hash := sha256.Sum256([]byte(time.Now().String()))
	return base64.StdEncoding.EncodeToString(hash[:8])
}
