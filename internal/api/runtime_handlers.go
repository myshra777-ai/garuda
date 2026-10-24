// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/runtime"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/tenant"
)

// ─────────────────────────────────────────────────────────────────────────────
// Runtime coverage: state-aware response fields
// ─────────────────────────────────────────────────────────────────────────────

type coverageField struct {
	State  string `json:"state"`
	Value  *int64 `json:"value,omitempty"`
	Reason string `json:"reason,omitempty"`
}

func observed(v int64) coverageField {
	return coverageField{State: "observed", Value: &v}
}

func unavailable(reason string) coverageField {
	return coverageField{State: "unavailable", Reason: reason}
}

func unavailableDerived(reason string) coverageField {
	return coverageField{State: "unavailable_derived", Reason: reason}
}

type RuntimeCoverageResponse struct {
	TotalStatic       coverageField `json:"total_static"`
	ObservedEntities  coverageField `json:"observed_entities"`
	SupportedCount    coverageField `json:"supported_count"`
	UnverifiedCount   coverageField `json:"unverified_count"`
	ContradictedCount coverageField `json:"contradicted_count"`
	TotalClaims       coverageField `json:"total_claims"`
	CoveragePercent   coverageField `json:"coverage_percent"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Runtime span ingestion
// ─────────────────────────────────────────────────────────────────────────────

type IngestTelemetrySpanDTO struct {
	TraceID       string                 `json:"trace_id"`
	SpanID        string                 `json:"span_id"`
	ServiceName   string                 `json:"service_name"`
	SourceService string                 `json:"source_service,omitempty"`
	TargetService string                 `json:"target_service"`
	CallerSymbol  string                 `json:"caller_symbol,omitempty"`
	CalleeSymbol  string                 `json:"callee_symbol,omitempty"`
	Operation     string                 `json:"operation"`
	DurationMS    float64                `json:"duration_ms"`
	StatusCode    string                 `json:"status_code"`
	Attributes    map[string]interface{} `json:"attributes"`
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

	tenantID := tenant.CanonicalID
	if ctxTenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID); ok && ctxTenantID != uuid.Nil {
		tenantID = ctxTenantID
	}

	workspaceName := r.URL.Query().Get("workspace")
	if workspaceName == "" {
		workspaceName = "default"
	}

	workspaceID, err := store.ResolveWorkspaceID(ctx, pgStore.Pool(), tenantID, workspaceName)
	if err != nil {
		http.Error(w, "no workspaces exist for tenant", http.StatusNotFound)
		return
	}

	var req IngestTelemetryRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	correlator := runtime.NewEntityCorrelator(pgStore.Pool())
	ingestedCount := 0

	for _, span := range req.Spans {
		if span.Attributes == nil {
			span.Attributes = make(map[string]interface{})
		}
		if span.ServiceName == "" && span.SourceService != "" {
			span.ServiceName = span.SourceService
		}
		if span.CallerSymbol != "" {
			if _, exists := span.Attributes["code.function"]; !exists {
				span.Attributes["code.function"] = span.CallerSymbol
			}
		}

		rawTarget, _ := span.Attributes["rpc.target_endpoint"].(string)
		if rawTarget == "" && span.CalleeSymbol != "" {
			rawTarget = span.CalleeSymbol
		}
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

		corrResult := correlator.Correlate(ctx, tenantID, workspaceID, &obs)
		entityID := corrResult.EntityID

		var insertedID uuid.UUID
		err = pgStore.Pool().QueryRow(ctx, `
            INSERT INTO runtime_observations (
                workspace_id, tenant_id, trace_id, span_id,
                source_service, target_service, service_name, operation,
                entity_id, duration_ms, status_code, attributes, started_at
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12::jsonb, $13)
            RETURNING id
        `, workspaceID, tenantID, obs.TraceID, obs.SpanID,
			obs.ServiceName, span.TargetService, obs.ServiceName, obs.Operation,
			entityID, int(obs.DurationMs), obs.StatusCode, string(attrJSON), obs.StartedAt).Scan(&insertedID)

		if err != nil {
			slog.Error("runtime_observations insert failed",
				"trace_id", span.TraceID,
				"span_id", span.SpanID,
				"error", err)
			continue
		}
		ingestedCount++

		if rawTarget != "" && entityID != nil {
			var targetEntityID *uuid.UUID
			if !runtime.ContainsContradictionMarker(rawTarget) {
				targetObs := runtime.RuntimeObservation{
					WorkspaceID: workspaceID,
					ServiceName: span.TargetService,
					Operation:   rawTarget,
					Attributes: map[string]interface{}{
						"code.function": rawTarget,
					},
				}
				targetCorr := correlator.Correlate(ctx, tenantID, workspaceID, &targetObs)
				targetEntityID = targetCorr.EntityID
			}

			_, err = pgStore.Pool().Exec(ctx, `
                INSERT INTO runtime_edges (
                    id, workspace_id, tenant_id, source_entity_id, target_entity_id,
                    raw_target, invocation_count, last_seen_at
                ) VALUES ($1, $2, $3, $4, $5, $6, 1, NOW())
                ON CONFLICT (workspace_id, tenant_id, source_entity_id, raw_target)
                DO UPDATE SET
                    target_entity_id = COALESCE(EXCLUDED.target_entity_id, runtime_edges.target_entity_id),
                    invocation_count = runtime_edges.invocation_count + 1,
                    last_seen_at = NOW()
            `, uuid.New(), workspaceID, tenantID, *entityID, targetEntityID, rawTarget)
			if err != nil {
				slog.Error("runtime_edges insert failed",
					"trace_id", span.TraceID,
					"raw_target", rawTarget,
					"error", err)
			}

			if runtime.ContainsContradictionMarker(rawTarget) {
				evidencePayload, _ := json.Marshal(map[string]interface{}{
					"raw_target":    rawTarget,
					"last_trace_id": span.TraceID,
					"service":       span.ServiceName,
				})

				_, err = pgStore.Pool().Exec(ctx, `
                    INSERT INTO claim_verifications (
                        id, workspace_id, tenant_id, source_entity_id, target_entity_id,
                        status, reason, static_edge_exists, runtime_observed_count,
                        evidence_payload, last_evaluated_at
                    ) VALUES (
                        $1, $2, $3, $4, md5($5)::uuid,
                        'CONTRADICTED', 'POLICY_VIOLATION_DETECTED', FALSE, 1,
                        $6::jsonb, NOW()
                    )
                    ON CONFLICT (workspace_id, tenant_id, source_entity_id, target_entity_id)
                    DO UPDATE SET
                        status = 'CONTRADICTED',
                        runtime_observed_count = claim_verifications.runtime_observed_count + 1,
                        evidence_payload = $6::jsonb,
                        last_evaluated_at = NOW()
                `, uuid.New(), workspaceID, tenantID, *entityID, rawTarget, evidencePayload)
				if err != nil {
					slog.Error("claim_verifications insert failed",
						"trace_id", span.TraceID,
						"raw_target", rawTarget,
						"error", err)
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if ingestedCount == 0 && len(req.Spans) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ingested": 0,
			"received": len(req.Spans),
			"status":   "rejected",
		})
		return
	}
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ingested": ingestedCount,
		"received": len(req.Spans),
		"status":   "accepted",
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Runtime coverage
// ─────────────────────────────────────────────────────────────────────────────

func (s *Server) HandleGetRuntimeCoverage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}

	scope, err := workspaceScopeFromRequest(ctx, r, pgStore)
	if err != nil {
		writeWorkspaceNotFound(w, r, r.URL.Query().Get("workspace"))
		return
	}
	workspaceID := scope.WorkspaceID
	tenantID := scope.TenantID

	verifier := runtime.NewVerificationEngine(pgStore.Pool())
	_, _ = verifier.RecomputeWorkspaceVerification(ctx, tenantID, workspaceID)

	queryInt := func(sql string) (int64, error) {
		var v int64
		err := pgStore.Pool().QueryRow(ctx, sql, workspaceID).Scan(&v)
		return v, err
	}

	var totalStatic, observedEntities, supportedCount, unverifiedCount, contradictedCount, totalClaims coverageField

	if v, err := queryInt(`SELECT COUNT(*)::bigint FROM entities WHERE workspace_id = $1`); err == nil {
		totalStatic = observed(v)
	} else {
		slog.Error("coverage: entities count failed", "error", err, "workspace_id", workspaceID)
		totalStatic = unavailable("entities count query failed")
	}

	if v, err := queryInt(`SELECT COUNT(DISTINCT entity_id)::bigint FROM runtime_observations WHERE workspace_id = $1 AND entity_id IS NOT NULL`); err == nil {
		observedEntities = observed(v)
	} else {
		slog.Error("coverage: observed entities count failed", "error", err, "workspace_id", workspaceID)
		observedEntities = unavailable("runtime_observations count query failed")
	}

	if v, err := queryInt(`SELECT COUNT(*)::bigint FROM claim_verifications WHERE workspace_id = $1 AND status = 'SUPPORTED'`); err == nil {
		supportedCount = observed(v)
	} else {
		slog.Error("coverage: supported count failed", "error", err, "workspace_id", workspaceID)
		supportedCount = unavailable("claim_verifications SUPPORTED count query failed")
	}

	if v, err := queryInt(`SELECT COUNT(*)::bigint FROM claim_verifications WHERE workspace_id = $1 AND status = 'UNVERIFIED'`); err == nil {
		unverifiedCount = observed(v)
	} else {
		slog.Error("coverage: unverified count failed", "error", err, "workspace_id", workspaceID)
		unverifiedCount = unavailable("claim_verifications UNVERIFIED count query failed")
	}

	if v, err := queryInt(`SELECT COUNT(*)::bigint FROM claim_verifications WHERE workspace_id = $1 AND status = 'CONTRADICTED'`); err == nil {
		contradictedCount = observed(v)
	} else {
		slog.Error("coverage: contradicted count failed", "error", err, "workspace_id", workspaceID)
		contradictedCount = unavailable("claim_verifications CONTRADICTED count query failed")
	}

	if v, err := queryInt(`SELECT COUNT(*)::bigint FROM claims WHERE workspace_id = $1`); err == nil {
		totalClaims = observed(v)
	} else {
		slog.Error("coverage: claims count failed", "error", err, "workspace_id", workspaceID)
		totalClaims = unavailable("claims count query failed")
	}

	var coveragePercent coverageField
	if totalStatic.State == "observed" && observedEntities.State == "observed" && *totalStatic.Value > 0 {
		pct := (*observedEntities.Value * 10000) / *totalStatic.Value
		coveragePercent = observed(pct)
	} else {
		coveragePercent = unavailableDerived("total_static or observed_entities unavailable")
	}

	resp := RuntimeCoverageResponse{
		TotalStatic:       totalStatic,
		ObservedEntities:  observedEntities,
		SupportedCount:    supportedCount,
		UnverifiedCount:   unverifiedCount,
		ContradictedCount: contradictedCount,
		TotalClaims:       totalClaims,
		CoveragePercent:   coveragePercent,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// ─────────────────────────────────────────────────────────────────────────────
// Merkle state
// ─────────────────────────────────────────────────────────────────────────────

func (s *Server) HandleGetMerkleState(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}

	tenantIDStr := r.URL.Query().Get("tenant_id")
	tenantID := tenant.CanonicalID
	if tenantIDStr != "" {
		if parsed, err := uuid.Parse(tenantIDStr); err == nil {
			tenantID = parsed
		}
	}

	snap, err := pgStore.GetLatestMerkleSnapshot(ctx, tenantID)
	if err != nil || snap == nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":              "GENESIS",
			"block_height":        1,
			"snapshot_hash":       "Genesis verified",
			"static_root_hash":    "Genesis",
			"runtime_root_hash":   "Genesis",
			"verified_claims":     0,
			"contradicted_claims": 0,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snap)
}
