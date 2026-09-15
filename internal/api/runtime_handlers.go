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

// coverageField is one value in the runtime coverage summary. A field
// whose query failed carries state "unavailable" and a null value;
// the client must render it as unknown, not as zero.
//
// This mirrors the three-state claim model from the README:
//
//	observed            — the value is a fact
//	unavailable         — the query failed; the value is unknown
//	unavailable_derived — depends on an unavailable upstream field
//
// A zero value and an unknown value are different facts. A workspace
// with zero SUPPORTED claims and a workspace whose SUPPORTED count
// query failed are not the same workspace. Reporting both as 0 is the
// same class of lie the claim_verifications table refuses to make.
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

// RuntimeCoverageResponse is the state-aware response shape for
// HandleGetRuntimeCoverage. It supersedes runtime.RuntimeCoverageSummary
// for this endpoint, which could only express ints and therefore could
// not distinguish "zero" from "unknown".
//
// CoveragePercent is expressed in basis points (100 = 1%). A value of
// 5000 means 50.00%. This keeps the state-aware field shape uniform
// across every field in the response.
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
	TargetService string                 `json:"target_service"`
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

	// TODO(session-E): this handler is invoked by external agents, not
	// browsers, and has no session to derive a tenant from. Until
	// agent authentication carries a tenant claim, this uses the
	// canonical tenant. The runtime coverage handler above is
	// session-scoped and does not have this limitation.
	workspaceName := r.URL.Query().Get("workspace")
	if workspaceName == "" {
		workspaceName = "default"
	}

	// Resolve via the tenant-scoped helper. Empty name selects the
	// most recently updated workspace for the tenant. The previous
	// code fell back to `SELECT id FROM workspaces LIMIT 1` with no
	// tenant filter, which could return a workspace owned by another
	// tenant.
	workspaceID, err := store.ResolveWorkspaceID(ctx, pgStore.Pool(), tenant.CanonicalID, workspaceName)
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
	tenantID := tenant.CanonicalID

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
			// The unique constraint is unq_runtime_edge over four
			// columns: (workspace_id, tenant_id, source_entity_id,
			// raw_target). The ON CONFLICT target must name the
			// same four columns.
			_, err = pgStore.Pool().Exec(ctx, `
                INSERT INTO runtime_edges (
                    id, workspace_id, tenant_id, source_entity_id,
                    raw_target, invocation_count, last_seen_at
                ) VALUES ($1, $2, $3, $4, $5, 1, NOW())
                ON CONFLICT (workspace_id, tenant_id, source_entity_id, raw_target)
                DO UPDATE SET
                    invocation_count = runtime_edges.invocation_count + 1,
                    last_seen_at = NOW()
            `, uuid.New(), workspaceID, tenantID, *entityID, rawTarget)
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

				// The unique constraint is unq_claim_verification
				// over four columns: (workspace_id, tenant_id,
				// source_entity_id, target_entity_id).
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

	// Each field is queried and its state recorded independently. A
	// failed query yields state "unavailable"; a downstream field that
	// depends on an unavailable value yields "unavailable_derived".
	// Neither is reported as zero. Absence of evidence is not evidence
	// of absence.
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

	// coverage_percent is derived. If either input is unavailable, the
	// percentage is unavailable_derived, not zero. The value is
	// expressed in basis points (100 = 1%) so it fits the same int64
	// shape as every other field.
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
