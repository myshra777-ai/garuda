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

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

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

type garudaExporter struct {
	cfg    *Config
	client *http.Client
	logger *zap.Logger
}

func newGarudaExporter(cfg *Config, logger *zap.Logger) *garudaExporter {
	return &garudaExporter{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger,
	}
}

func (e *garudaExporter) pushTraces(ctx context.Context, td ptrace.Traces) error {
	var payloads []SpanPayload

	resourceSpans := td.ResourceSpans()
	for i := 0; i < resourceSpans.Len(); i++ {
		rs := resourceSpans.At(i)
		serviceName := "unknown_service"
		if sn, ok := rs.Resource().Attributes().Get("service.name"); ok {
			serviceName = sn.AsString()
		}

		scopeSpans := rs.ScopeSpans()
		for j := 0; j < scopeSpans.Len(); j++ {
			ss := scopeSpans.At(j)
			packageName := ss.Scope().Name()

			spans := ss.Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				payload := e.convertSpan(serviceName, packageName, span)
				if payload != nil {
					payloads = append(payloads, *payload)
				}
			}
		}
	}

	if len(payloads) == 0 {
		return nil
	}

	return e.sendBatches(ctx, payloads)
}

func (e *garudaExporter) convertSpan(serviceName, scopeName string, span ptrace.Span) *SpanPayload {
	callerSymbol := span.Name()
	callerPkg := scopeName
	targetEndpoint := ""

	// 1. Check code attributes if available
	span.Attributes().Range(func(k string, v pcommon.Value) bool {
		switch k {
		case "code.function":
			callerSymbol = v.AsString()
		case "code.namespace":
			callerPkg = v.AsString()
		case "peer.service", "net.peer.name", "http.url", "db.name":
			if targetEndpoint == "" {
				targetEndpoint = v.AsString()
			}
		case "db.system":
			if targetEndpoint == "" {
				targetEndpoint = fmt.Sprintf("%s-instance", v.AsString())
			}
		}
		return true
	})

	if targetEndpoint == "" {
		targetEndpoint = span.Name()
	}

	ts := span.StartTimestamp().AsTime()
	if ts.IsZero() {
		ts = time.Now()
	}

	return &SpanPayload{
		ServiceName:    serviceName,
		CallerSymbol:   callerSymbol,
		CallerPackage:  callerPkg,
		TargetEndpoint: targetEndpoint,
		Timestamp:      ts.UTC().Format(time.RFC3339),
	}
}

func (e *garudaExporter) sendBatches(ctx context.Context, payloads []SpanPayload) error {
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

	e.logger.Debug("Pushed runtime spans to Garuda", zap.Int("count", len(payloads)))
	return nil
}
