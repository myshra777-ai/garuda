// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/store"
)

// TraceSpanPayload defines the minimal OTel span format Garuda needs.
type TraceSpanPayload struct {
	TraceID       string `json:"trace_id"`
	SpanID        string `json:"span_id"`
	SourceService string `json:"source_service"`
	TargetService string `json:"target_service"`
	Operation     string `json:"operation"`
	StatusCode    string `json:"status_code"`
	DurationMS    int    `json:"duration_ms"`
	Environment   string `json:"environment"`
	Workspace     string `json:"workspace"`
}

type TraceIngestRequest struct {
	Spans []TraceSpanPayload `json:"spans"`
}

// HandleIngestTraces accepts minimal OpenTelemetry trace batches.
func (s *Server) HandleIngestTraces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}

	var req TraceIngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	ingested := 0

	for _, span := range req.Spans {
		wsName := span.Workspace
		if wsName == "" {
			wsName = "uuid-ws"
		}

		var wsID uuid.UUID
		err := pgStore.Pool().QueryRow(
			ctx,
			`SELECT id FROM workspaces WHERE name = $1 LIMIT 1`,
			wsName,
		).Scan(&wsID)
		if err != nil {
			continue
		}

		env := span.Environment
		if env == "" {
			env = "production"
		}

		statusCode := span.StatusCode
		if statusCode == "" {
			statusCode = "OK"
		}

		_, err = pgStore.Pool().Exec(
			ctx,
			`
			INSERT INTO runtime_observations (
				workspace_id, trace_id, span_id, source_service, target_service,
				operation, status_code, duration_ms, environment, observed_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			`,
			wsID, span.TraceID, span.SpanID, span.SourceService, span.TargetService,
			span.Operation, statusCode, span.DurationMS, env, time.Now().UTC(),
		)
		if err == nil {
			ingested++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "INGESTED",
		"count":    ingested,
		"received": len(req.Spans),
	})
}
