// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package garudaexporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Config struct {
	Endpoint     string        `json:"endpoint"`
	TenantID     string        `json:"tenant_id"`
	WorkspaceID  string        `json:"workspace_id"`
	Timeout      time.Duration `json:"timeout"`
	MaxBatchSize int           `json:"max_batch_size"`
}

func DefaultConfig() Config {
	return Config{
		Endpoint:     "http://localhost:8080/api/v1/telemetry/spans",
		TenantID:     "00000000-0000-0000-0000-000000000001",
		WorkspaceID:  "532a8e33-975d-48a3-8f88-221cef52fec4",
		Timeout:      5 * time.Second,
		MaxBatchSize: 1000,
	}
}

type SpanPayload struct {
	ServiceName    string `json:"service_name"`
	CallerSymbol   string `json:"caller_symbol"`
	CallerPackage  string `json:"caller_package"`
	TargetEndpoint string `json:"target_endpoint"`
	Timestamp      string `json:"timestamp"`
}

type IngestRequest struct {
	Spans []SpanPayload `json:"spans"`
}

type GarudaSpanExporter struct {
	cfg    Config
	client *http.Client
}

func New(cfg Config) (*GarudaSpanExporter, error) {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "http://localhost:8080/api/v1/telemetry/spans"
	}
	if cfg.TenantID == "" {
		cfg.TenantID = "00000000-0000-0000-0000-000000000001"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.MaxBatchSize <= 0 {
		cfg.MaxBatchSize = 1000
	}

	return &GarudaSpanExporter{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}, nil
}

func (e *GarudaSpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	if len(spans) == 0 {
		return nil
	}

	payloads := make([]SpanPayload, 0, len(spans))
	for _, span := range spans {
		serviceName := "unknown_service"
		for _, attr := range span.Resource().Attributes() {
			if string(attr.Key) == "service.name" {
				serviceName = attr.Value.AsString()
			}
		}

		callerSymbol := span.Name()
		callerPkg := span.InstrumentationScope().Name
		targetEndpoint := span.Name()

		for _, attr := range span.Attributes() {
			switch string(attr.Key) {
			case "code.function":
				callerSymbol = attr.Value.AsString()
			case "code.namespace":
				callerPkg = attr.Value.AsString()
			case "peer.service", "net.peer.name", "http.url", "db.name":
				targetEndpoint = attr.Value.AsString()
			}
		}

		payloads = append(payloads, SpanPayload{
			ServiceName:    serviceName,
			CallerSymbol:   callerSymbol,
			CallerPackage:  callerPkg,
			TargetEndpoint: targetEndpoint,
			Timestamp:      span.StartTime().UTC().Format(time.RFC3339),
		})
	}

	return e.sendBatches(ctx, payloads)
}

func (e *GarudaSpanExporter) sendBatches(ctx context.Context, payloads []SpanPayload) error {
	batchSize := e.cfg.MaxBatchSize
	for i := 0; i < len(payloads); i += batchSize {
		end := i + batchSize
		if end > len(payloads) {
			end = len(payloads)
		}

		batch := payloads[i:end]
		reqBody, err := json.Marshal(IngestRequest{Spans: batch})
		if err != nil {
			return fmt.Errorf("failed to encode spans to JSON: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.cfg.Endpoint, bytes.NewReader(reqBody))
		if err != nil {
			return fmt.Errorf("failed to create http request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Garuda-Tenant-ID", e.cfg.TenantID)
		req.Header.Set("X-Garuda-Workspace-ID", e.cfg.WorkspaceID)

		resp, err := e.client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to post spans to garuda endpoint (%s): %w", e.cfg.Endpoint, err)
		}
		resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("garuda ingestion returned HTTP %d", resp.StatusCode)
		}
	}
	return nil
}

func (e *GarudaSpanExporter) Shutdown(ctx context.Context) error {
	return nil
}
