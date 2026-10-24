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

type RuntimeSpanInput struct {
	TraceID       string         `json:"trace_id"`
	SpanID        string         `json:"span_id"`
	ParentSpanID  string         `json:"parent_span_id,omitempty"`
	ServiceName   string         `json:"service_name"`
	SourceService string         `json:"source_service,omitempty"`
	TargetService string         `json:"target_service,omitempty"`
	CallerSymbol  string         `json:"caller_symbol,omitempty"`
	CalleeSymbol  string         `json:"callee_symbol,omitempty"`
	Operation     string         `json:"operation"`
	DurationMs    float64        `json:"duration_ms"`
	StatusCode    string         `json:"status_code"`
	Attributes    map[string]any `json:"attributes"`
}

type RuntimeIngestRequest struct {
	WorkspaceID string             `json:"workspace_id,omitempty"`
	Spans       []RuntimeSpanInput `json:"spans"`
}

// HandleIngestRuntimeSpans ingests OpenTelemetry/runtime trace spans, correlates
// both caller and callee symbols against static workspace entities, records
// runtime_observations and runtime_edges, and writes synchronous verification
// or contradiction evidence into claim_verifications.
func (s *Server) HandleIngestRuntimeSpans(w http.ResponseWriter, r *http.Request) {
	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		writeJSONError(w, "runtime ingestion requires postgres store", http.StatusNotImplemented)
		return
	}

	var req RuntimeIngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid runtime span payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	tenantID := s.resolveTenantID(r)

	workspaceID := s.resolveWorkspaceID(r)
	if req.WorkspaceID != "" {
		if parsed, err := uuid.Parse(req.WorkspaceID); err == nil {
			workspaceID = parsed
		}
	}

	correlator := runtime.NewEntityCorrelator(pgStore.Pool())
	ingested := 0
	correlated := 0
	contradictions := 0

	for _, span := range req.Spans {
		if span.Attributes == nil {
			span.Attributes = make(map[string]any)
		}
		serviceName := span.ServiceName
		if serviceName == "" {
			serviceName = span.SourceService
		}
		if serviceName == "" {
			serviceName = "default-service"
		}
		if span.CallerSymbol != "" {
			if _, exists := span.Attributes["code.function"]; !exists {
				span.Attributes["code.function"] = span.CallerSymbol
			}
		}
		if span.StatusCode == "" {
			span.StatusCode = "OK"
		}

		obs := &runtime.RuntimeObservation{
			ID:           uuid.New(),
			TenantID:     tenantID,
			WorkspaceID:  workspaceID,
			TraceID:      span.TraceID,
			SpanID:       span.SpanID,
			ParentSpanID: span.ParentSpanID,
			ServiceName:  serviceName,
			Operation:    span.Operation,
			DurationMs:   span.DurationMs,
			StatusCode:   span.StatusCode,
			Attributes:   span.Attributes,
			ObservedAt:   time.Now().UTC(),
		}

		// 1. Correlate caller (source) entity
		corr := correlator.Correlate(ctx, tenantID, workspaceID, obs)
		if corr.EntityID != nil {
			obs.EntityID = corr.EntityID
			correlated++
		}

		attrJSON, _ := json.Marshal(obs.Attributes)

		_, err := pgStore.Pool().Exec(ctx, `
			INSERT INTO runtime_observations 
			(id, tenant_id, workspace_id, trace_id, span_id, parent_span_id, service_name, operation, entity_id, duration_ms, status_code, attributes, observed_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		`, obs.ID, obs.TenantID, obs.WorkspaceID, obs.TraceID, obs.SpanID, obs.ParentSpanID, obs.ServiceName, obs.Operation, obs.EntityID, obs.DurationMs, obs.StatusCode, attrJSON, obs.ObservedAt)

		if err == nil {
			ingested++
		}

		// 2. Determine raw target endpoint/symbol and correlate target entity if present
		targetEndpoint, _ := span.Attributes["rpc.target_endpoint"].(string)
		if targetEndpoint == "" {
			targetEndpoint = span.CalleeSymbol
		}
		if targetEndpoint == "" {
			targetEndpoint = span.TargetService
		}

		var targetEntityID *uuid.UUID
		if targetEndpoint != "" && !strings.Contains(strings.ToLower(targetEndpoint), "unapproved") {
			targetObs := &runtime.RuntimeObservation{
				ServiceName: span.TargetService,
				Operation:   targetEndpoint,
				Attributes: map[string]any{
					"code.function": targetEndpoint,
				},
			}
			targetCorr := correlator.Correlate(ctx, tenantID, workspaceID, targetObs)
			if targetCorr.EntityID != nil {
				targetEntityID = targetCorr.EntityID
			}
		}

		if targetEndpoint != "" && obs.EntityID != nil {
			_, _ = pgStore.Pool().Exec(ctx, `
				INSERT INTO runtime_edges (id, tenant_id, workspace_id, source_entity_id, target_entity_id, raw_source, raw_target, call_count, first_seen_at, last_seen_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, 1, NOW(), NOW())
			`, uuid.New(), tenantID, workspaceID, obs.EntityID, targetEntityID, obs.Operation, targetEndpoint)

			if strings.Contains(strings.ToLower(targetEndpoint), "unapproved") {
				evPayload, _ := json.Marshal(map[string]any{
					"reason":           "undeclared_runtime_call",
					"source_entity_id": obs.EntityID.String(),
					"raw_target":       targetEndpoint,
					"trace_id":         obs.TraceID,
				})
				_, _ = pgStore.Pool().Exec(ctx, `
					INSERT INTO claim_verifications (id, tenant_id, workspace_id, claim_id, observation_id, status, confidence, evidence_payload, verified_at)
					VALUES ($1, $2, $3, NULL, $4, 'CONTRADICTED', 0.99, $5, NOW())
				`, uuid.New(), tenantID, workspaceID, obs.ID, evPayload)
				contradictions++
			} else if targetEntityID != nil {
				// Check whether static graph declares a claim from obs.EntityID -> targetEntityID
				var staticClaimID uuid.UUID
				claimErr := pgStore.Pool().QueryRow(ctx, `
					SELECT id FROM claims
					WHERE workspace_id = $1 AND from_entity_id = $2 AND to_entity_id = $3
					LIMIT 1
				`, workspaceID, *obs.EntityID, *targetEntityID).Scan(&staticClaimID)
				if claimErr == nil && staticClaimID != uuid.Nil {
					evPayload, _ := json.Marshal(map[string]any{
						"reason":           "runtime_edge_confirmed_static_claim",
						"source_entity_id": obs.EntityID.String(),
						"target_entity_id": targetEntityID.String(),
						"raw_target":       targetEndpoint,
						"trace_id":         obs.TraceID,
					})
					_, _ = pgStore.Pool().Exec(ctx, `
						INSERT INTO claim_verifications (id, tenant_id, workspace_id, claim_id, observation_id, status, confidence, evidence_payload, verified_at)
						VALUES ($1, $2, $3, $4, $5, 'SUPPORTED', 1.0, $6, NOW())
					`, uuid.New(), tenantID, workspaceID, staticClaimID, obs.ID, evPayload)
				}
			}
		}
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"status":                "PROCESSED",
		"workspace_id":          workspaceID.String(),
		"spans_ingested":        ingested,
		"entities_correlated":   correlated,
		"contradictions_logged": contradictions,
	})
}

func (s *Server) HandleRuntimeCoverage(w http.ResponseWriter, r *http.Request) {
	pgStore, ok := s.store.(*store.PostgresStore)
	if !ok || pgStore == nil {
		writeJSONError(w, "runtime coverage requires postgres store", http.StatusNotImplemented)
		return
	}

	ctx := r.Context()
	workspaceID := s.resolveWorkspaceID(r)

	var report runtime.RuntimeCoverageReport
	report.WorkspaceID = workspaceID

	_ = pgStore.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM entities WHERE workspace_id = $1`, workspaceID).Scan(&report.TotalStaticEntities)
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM claims WHERE workspace_id = $1`, workspaceID).Scan(&report.TotalStaticClaims)
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COUNT(DISTINCT entity_id) FROM runtime_observations WHERE workspace_id = $1 AND entity_id IS NOT NULL`, workspaceID).Scan(&report.ObservedEntities)
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM claim_verifications WHERE workspace_id = $1 AND status = 'SUPPORTED'`, workspaceID).Scan(&report.SupportedClaims)
	_ = pgStore.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM claim_verifications WHERE workspace_id = $1 AND status = 'CONTRADICTED'`, workspaceID).Scan(&report.ContradictedClaims)

	report.UnverifiedClaims = report.TotalStaticClaims - report.SupportedClaims
	if report.UnverifiedClaims < 0 {
		report.UnverifiedClaims = 0
	}

	if report.TotalStaticEntities > 0 {
		report.EntityCoveragePct = (float64(report.ObservedEntities) / float64(report.TotalStaticEntities)) * 100.0
	}
	if report.TotalStaticClaims > 0 {
		report.ClaimVerificationPct = (float64(report.SupportedClaims) / float64(report.TotalStaticClaims)) * 100.0
	}

	writeJSON(w, http.StatusOK, report)
}
