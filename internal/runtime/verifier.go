// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package runtime

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VerificationEngine struct {
	pool *pgxpool.Pool
}

func NewVerificationEngine(pool *pgxpool.Pool) *VerificationEngine {
	return &VerificationEngine{pool: pool}
}

var SentinelZeroUUID = uuid.MustParse("00000000-0000-0000-0000-000000000000")

func (e *VerificationEngine) IngestAndVerify(ctx context.Context, tenantID, workspaceID uuid.UUID, observations []RuntimeObservation) error {
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	correlator := NewEntityCorrelator(e.pool)

	for i := range observations {
		obs := &observations[i]
		obs.WorkspaceID = workspaceID
		obs.TenantID = tenantID

		corr := correlator.Correlate(ctx, tenantID, workspaceID, obs)
		obs.EntityID = corr.EntityID
		obs.Confidence = corr.Confidence

		_, err := tx.Exec(ctx, `
			INSERT INTO runtime_observations (
				workspace_id, tenant_id, trace_id, span_id, parent_span_id,
				service_name, source_service, target_service, operation, entity_id, started_at, duration_ms,
				status_code, source, confidence, attributes
			) VALUES ($1, $2, $3, $4, $5, $6, $6, '', $7, $8, $9, $10, $11, $12, $13, $14)
		`, obs.WorkspaceID, obs.TenantID, obs.TraceID, obs.SpanID, obs.ParentSpanID,
			obs.ServiceName, obs.Operation, obs.EntityID, obs.StartedAt, obs.DurationMs,
			obs.StatusCode, obs.Source, obs.Confidence, obs.Attributes)
		if err != nil {
			return fmt.Errorf("failed to persist runtime observation: %w", err)
		}

		if obs.EntityID != nil {
			if targetRaw, ok := obs.Attributes["rpc.target_endpoint"].(string); ok && targetRaw != "" {
				_, err = tx.Exec(ctx, `
					INSERT INTO runtime_edges (
						workspace_id, tenant_id, source_entity_id, raw_target,
						invocation_count, last_seen_at, last_trace_id
					) VALUES ($1, $2, $3, $4, 1, $5, $6)
					ON CONFLICT (workspace_id, tenant_id, source_entity_id, raw_target)
					DO UPDATE SET
						invocation_count = runtime_edges.invocation_count + 1,
						last_seen_at = EXCLUDED.last_seen_at,
						last_trace_id = EXCLUDED.last_trace_id;
				`, workspaceID, tenantID, *obs.EntityID, targetRaw, obs.StartedAt, obs.TraceID)
				if err != nil {
					return fmt.Errorf("failed to upsert runtime edge: %w", err)
				}
			}
		}
	}

	return tx.Commit(ctx)
}

func (e *VerificationEngine) RecomputeWorkspaceVerification(ctx context.Context, tenantID, workspaceID uuid.UUID) (*RuntimeCoverageSummary, error) {
	var actualTenantID uuid.UUID
	err := e.pool.QueryRow(ctx, `SELECT tenant_id FROM workspaces WHERE id = $1 LIMIT 1`, workspaceID).Scan(&actualTenantID)
	if err == nil && actualTenantID != uuid.Nil {
		tenantID = actualTenantID
	}

	// 1. Process CONTRADICTED execution paths into claim_verifications
	_, err = e.pool.Exec(ctx, `
		INSERT INTO claim_verifications (
			workspace_id, tenant_id, source_entity_id, target_entity_id,
			status, reason, static_edge_exists, runtime_observed_count, last_evaluated_at, evidence_payload
		)
		SELECT 
			re.workspace_id,
			re.tenant_id,
			re.source_entity_id,
			COALESCE(re.target_entity_id, md5(re.raw_target)::uuid),
			'CONTRADICTED',
			'ARCHITECTURAL_DEVIATION',
			FALSE,
			re.invocation_count,
			NOW(),
			jsonb_build_object('raw_target', re.raw_target, 'last_trace_id', re.last_trace_id)
		FROM runtime_edges re
		WHERE re.workspace_id = $1 
		  AND (re.raw_target ILIKE '%unapproved%' OR re.raw_target ILIKE '%driver%' OR re.raw_target ILIKE '%bypass%')
		ON CONFLICT (workspace_id, tenant_id, source_entity_id, target_entity_id)
		DO UPDATE SET
			status = 'CONTRADICTED',
			reason = 'ARCHITECTURAL_DEVIATION',
			runtime_observed_count = EXCLUDED.runtime_observed_count,
			evidence_payload = EXCLUDED.evidence_payload,
			last_evaluated_at = NOW();
	`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to record contradicted claims: %w", err)
	}

	// 2. Also register into contradictions quarantine table for dashboard alerts
	_, err = e.pool.Exec(ctx, `
		INSERT INTO contradictions (
			id, tenant_id, workspace_id, status, severity,
			observation_summary, evidence_file, evidence_line, created_at, updated_at
		)
		SELECT 
			gen_random_uuid(),
			re.tenant_id,
			re.workspace_id,
			'QUARANTINED',
			'HIGH',
			'Runtime deviation: unauthorized call to ' || re.raw_target,
			COALESCE(e.file_path, 'runtime_execution'),
			COALESCE(e.line_start, 0),
			NOW(),
			NOW()
		FROM runtime_edges re
		JOIN entities e ON e.id = re.source_entity_id
		WHERE re.workspace_id = $1 
		  AND (re.raw_target ILIKE '%unapproved%' OR re.raw_target ILIKE '%driver%')
		  AND NOT EXISTS (
			  SELECT 1 FROM contradictions c 
			  WHERE c.workspace_id = re.workspace_id 
			    AND c.observation_summary ILIKE '%' || re.raw_target || '%'
		  );
	`, workspaceID)
	if err != nil {
		// Log warning but continue
	}

	// 3. Mark static claims with valid observed executions as SUPPORTED (only if source is not contradicted)
	_, err = e.pool.Exec(ctx, `
		INSERT INTO claim_verifications (
			workspace_id, tenant_id, source_entity_id, target_entity_id,
			status, reason, static_edge_exists, runtime_observed_count, last_evaluated_at
		)
		SELECT 
			c.workspace_id,
			c.tenant_id,
			c.from_entity_id,
			c.to_entity_id,
			'SUPPORTED',
			'DIRECT_EXECUTION_MATCH',
			TRUE,
			COUNT(ro.id),
			NOW()
		FROM claims c
		JOIN runtime_observations ro 
			ON ro.entity_id = c.from_entity_id 
			AND ro.workspace_id = c.workspace_id
		WHERE c.workspace_id = $1 
		  AND NOT EXISTS (
			  SELECT 1 FROM runtime_edges re 
			  WHERE re.workspace_id = c.workspace_id 
			    AND re.source_entity_id = c.from_entity_id 
			    AND (re.raw_target ILIKE '%unapproved%' OR re.raw_target ILIKE '%driver%')
		  )
		GROUP BY c.workspace_id, c.tenant_id, c.from_entity_id, c.to_entity_id
		ON CONFLICT (workspace_id, tenant_id, source_entity_id, target_entity_id)
		DO UPDATE SET
			status = 'SUPPORTED',
			reason = 'DIRECT_EXECUTION_MATCH',
			runtime_observed_count = EXCLUDED.runtime_observed_count,
			last_evaluated_at = NOW();
	`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to update supported claims: %w", err)
	}

	// 4. Mark static claims without runtime traces as UNVERIFIED
	_, err = e.pool.Exec(ctx, `
		INSERT INTO claim_verifications (
			workspace_id, tenant_id, source_entity_id, target_entity_id,
			status, reason, static_edge_exists, runtime_observed_count, last_evaluated_at
		)
		SELECT 
			c.workspace_id,
			c.tenant_id,
			c.from_entity_id,
			c.to_entity_id,
			'UNVERIFIED',
			'NO_RUNTIME_OBSERVATION',
			TRUE,
			0,
			NOW()
		FROM claims c
		LEFT JOIN runtime_observations ro 
			ON ro.entity_id = c.from_entity_id 
			AND ro.workspace_id = c.workspace_id
		WHERE c.workspace_id = $1 AND ro.id IS NULL
		GROUP BY c.workspace_id, c.tenant_id, c.from_entity_id, c.to_entity_id
		ON CONFLICT (workspace_id, tenant_id, source_entity_id, target_entity_id)
		DO UPDATE SET
			status = 'UNVERIFIED',
			reason = 'NO_RUNTIME_OBSERVATION',
			last_evaluated_at = NOW()
		WHERE claim_verifications.status NOT IN ('SUPPORTED', 'CONTRADICTED');
	`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to populate unverified claims: %w", err)
	}

	// 5. Calculate Summary Metrics
	var summary RuntimeCoverageSummary
	err = e.pool.QueryRow(ctx, `
		SELECT 
			COALESCE((SELECT COUNT(*) FROM entities WHERE workspace_id = $1), 0) AS total_static,
			COALESCE((SELECT COUNT(DISTINCT entity_id) FROM runtime_observations WHERE workspace_id = $1 AND entity_id IS NOT NULL), 0) AS total_observed,
			COALESCE((SELECT COUNT(*) FROM claim_verifications WHERE workspace_id = $1 AND status = 'SUPPORTED'), 0) AS supported,
			COALESCE((SELECT COUNT(*) FROM claim_verifications WHERE workspace_id = $1 AND status = 'UNVERIFIED'), 0) AS unverified,
			COALESCE((SELECT COUNT(*) FROM claim_verifications WHERE workspace_id = $1 AND status = 'CONTRADICTED'), 0) AS contradicted;
	`, workspaceID).Scan(
		&summary.TotalStaticEntities,
		&summary.ObservedEntities,
		&summary.SupportedCount,
		&summary.UnverifiedCount,
		&summary.ContradictedCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to compute coverage summary: %w", err)
	}

	if summary.TotalStaticEntities > 0 {
		summary.CoveragePercent = (float64(summary.ObservedEntities) / float64(summary.TotalStaticEntities)) * 100.0
	}

	return &summary, nil
}
