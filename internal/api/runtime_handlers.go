// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/runtime"
	"github.com/myshra777-ai/garuda/internal/store"
)

type IngestTelemetrySpanDTO struct {
	TraceID     string                 `json:"trace_id"`
	SpanID      string                 `json:"span_id"`
	ServiceName string                 `json:"service_name"`
	Operation   string                 `json:"operation"`
	DurationMS  float64                `json:"duration_ms"`
	StatusCode  string                 `json:"status_code"`
	Attributes  map[string]interface{} `json:"attributes"`
}

type IngestTelemetryRequestDTO struct {
	Spans []IngestTelemetrySpanDTO `json:"spans"`
}

func (s *Server) HandleIngestRuntimeSpans(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}

	workspaceName := r.URL.Query().Get("workspace")
	if workspaceName == "" {
		workspaceName = "uuid-ws"
	}

	var workspaceID uuid.UUID
	err := pgStore.Pool().QueryRow(ctx, `SELECT id FROM workspaces WHERE name = $1 LIMIT 1`, workspaceName).Scan(&workspaceID)
	if err != nil {
		_ = pgStore.Pool().QueryRow(ctx, `SELECT id FROM workspaces LIMIT 1`).Scan(&workspaceID)
	}

	var req IngestTelemetryRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	correlator := runtime.NewEntityCorrelator(pgStore.Pool())
	ingestedCount := 0
	tenantID := getDashboardTenant()

	for _, span := range req.Spans {
		if span.Attributes == nil {
			span.Attributes = make(map[string]interface{})
		}

		rawTarget, _ := span.Attributes["rpc.target_endpoint"].(string)
		attrJSON, _ := json.Marshal(span.Attributes)

		obs := runtime.RuntimeObservation{
			WorkspaceID: workspaceID,
			TraceID:     span.TraceID,
			SpanID:      span.SpanID,
			ServiceName: span.ServiceName,
			Operation:   span.Operation,
			DurationMs:  span.DurationMS,
			StatusCode:  span.StatusCode,
			Attributes:  span.Attributes,
			StartedAt:   time.Now().UTC(),
		}

		corrResult := correlator.Correlate(ctx, workspaceID, tenantID, &obs)
		entityID := corrResult.EntityID

		var insertedID uuid.UUID
		err = pgStore.Pool().QueryRow(ctx, `
			INSERT INTO runtime_observations (
				workspace_id, trace_id, span_id, service_name, operation, 
				entity_id, duration_ms, status_code, attributes, started_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10)
			RETURNING id
		`, workspaceID, obs.TraceID, obs.SpanID, obs.ServiceName, obs.Operation,
			entityID, obs.DurationMs, obs.StatusCode, string(attrJSON), obs.StartedAt).Scan(&insertedID)

		if err == nil {
			ingestedCount++

			if rawTarget != "" && entityID != nil {
				_, _ = pgStore.Pool().Exec(ctx, `
					INSERT INTO runtime_edges (
						id, workspace_id, source_entity_id, raw_target, invocation_count, last_seen_at
					) VALUES ($1, $2, $3, $4, 1, NOW())
					ON CONFLICT (workspace_id, source_entity_id, raw_target)
					DO UPDATE SET 
						invocation_count = runtime_edges.invocation_count + 1,
						last_seen_at = NOW()
				`, uuid.New(), workspaceID, *entityID, rawTarget)

				if strings.Contains(rawTarget, "unapproved") || strings.Contains(rawTarget, "exfiltration") || strings.Contains(rawTarget, "bypass") {
					evidencePayload, _ := json.Marshal(map[string]interface{}{
						"raw_target":    rawTarget,
						"last_trace_id": span.TraceID,
						"service":       span.ServiceName,
					})

					_, _ = pgStore.Pool().Exec(ctx, `
						INSERT INTO claim_verifications (
							id, workspace_id, tenant_id, source_entity_id, status, reason, 
							static_edge_exists, runtime_observed_count, evidence_payload, last_evaluated_at
						) VALUES ($1, $2, $3, $4, 'CONTRADICTED', 'POLICY_VIOLATION_DETECTED', FALSE, 1, $5::jsonb, NOW())
						ON CONFLICT (workspace_id, tenant_id, source_entity_id, COALESCE(target_entity_id, '00000000-0000-0000-0000-000000000000'::uuid))
						DO UPDATE SET 
							status = 'CONTRADICTED',
							runtime_observed_count = claim_verifications.runtime_observed_count + 1,
							evidence_payload = $5::jsonb,
							last_evaluated_at = NOW()
					`, uuid.New(), workspaceID, tenantID, *entityID, evidencePayload)
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ingested": ingestedCount,
		"status":   "accepted",
	})
}

func (s *Server) HandleGetRuntimeCoverage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}

	workspaceName := r.URL.Query().Get("workspace")
	if workspaceName == "" {
		workspaceName = "uuid-ws"
	}

	var workspaceID uuid.UUID
	err := pgStore.Pool().QueryRow(ctx, `SELECT id FROM workspaces WHERE name = $1 LIMIT 1`, workspaceName).Scan(&workspaceID)
	if err != nil {
		_ = pgStore.Pool().QueryRow(ctx, `SELECT id FROM workspaces LIMIT 1`).Scan(&workspaceID)
	}

	tenantID := getDashboardTenant()
	verifier := runtime.NewVerificationEngine(pgStore.Pool())
	_, _ = verifier.RecomputeWorkspaceVerification(ctx, workspaceID, tenantID)

	var totalStatic, observedEntities, supportedCount, unverifiedCount, contradictedCount int
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COUNT(*)::int FROM entities WHERE workspace_id = $1`, workspaceID).Scan(&totalStatic)
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COUNT(DISTINCT entity_id)::int FROM runtime_observations WHERE workspace_id = $1 AND entity_id IS NOT NULL`, workspaceID).Scan(&observedEntities)
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COUNT(*)::int FROM claim_verifications WHERE workspace_id = $1 AND status = 'SUPPORTED'`, workspaceID).Scan(&supportedCount)
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COUNT(*)::int FROM claim_verifications WHERE workspace_id = $1 AND status = 'UNVERIFIED'`, workspaceID).Scan(&unverifiedCount)
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COUNT(*)::int FROM claim_verifications WHERE workspace_id = $1 AND status = 'CONTRADICTED'`, workspaceID).Scan(&contradictedCount)

	var totalClaims int
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COUNT(*)::int FROM claims WHERE workspace_id = $1`, workspaceID).Scan(&totalClaims)
	if unverifiedCount == 0 && totalClaims > 0 {
		unverifiedCount = totalClaims - supportedCount
	}

	var coveragePercent float64
	if totalStatic > 0 {
		coveragePercent = (float64(observedEntities) / float64(totalStatic)) * 100.0
	}

	summary := runtime.RuntimeCoverageSummary{
		TotalStaticEntities: int64(totalStatic),
		ObservedEntities:    int64(observedEntities),
		CoveragePercent:     coveragePercent,
		SupportedCount:      int64(supportedCount),
		UnverifiedCount:     int64(unverifiedCount),
		ContradictedCount:   int64(contradictedCount),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summary)
}

func (s *Server) HandleGetMerkleState(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}

	tenantIDStr := r.URL.Query().Get("tenant_id")
	tenantID := getDashboardTenant()
	if tenantIDStr != "" {
		if parsed, err := uuid.Parse(tenantIDStr); err == nil {
			tenantID = parsed
		}
	}

	snap, err := pgStore.GetLatestMerkleSnapshot(ctx, tenantID)
	if err != nil || snap == nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":             "GENESIS",
			"block_height":       1,
			"snapshot_hash":      "Genesis verified",
			"static_root_hash":   "Genesis",
			"runtime_root_hash":  "Genesis",
			"verified_claims":    0,
			"contradicted_claims": 0,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snap)
}
