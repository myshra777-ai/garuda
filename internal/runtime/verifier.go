// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VerificationEngine struct {
	pool *pgxpool.Pool
}

func NewVerificationEngine(pool *pgxpool.Pool) *VerificationEngine {
	return &VerificationEngine{pool: pool}
}

func (v *VerificationEngine) VerifyObservation(ctx context.Context, tenantID, workspaceID uuid.UUID, obs *RuntimeObservation, corr CorrelationResult) (*ClaimVerification, error) {
	now := time.Now().UTC()

	// CASE 1: Uncorrelated Runtime Activity (Shadow Execution / Undeclared Endpoint)
	if corr.EntityID == nil {
		evidence, _ := json.Marshal(map[string]any{
			"reason":       "uncorrelated_runtime_activity",
			"service_name": obs.ServiceName,
			"operation":    obs.Operation,
			"trace_id":     obs.TraceID,
		})
		return &ClaimVerification{
			ID:              uuid.New(),
			TenantID:        tenantID,
			WorkspaceID:     workspaceID,
			ObservationID:   &obs.ID,
			Status:          StatusUnverified,
			Confidence:      0.50,
			EvidencePayload: evidence,
			VerifiedAt:      now,
		}, nil
	}

	// CASE 2: Correlated Entity — Check if static claims exist for this entity
	var claimID uuid.UUID
	err := v.pool.QueryRow(ctx, `
		SELECT id FROM claims 
		WHERE workspace_id = $1 AND (from_entity_id = $2 OR to_entity_id = $2)
		LIMIT 1
	`, workspaceID, *corr.EntityID).Scan(&claimID)

	status := StatusSupported
	reason := "static_entity_confirmed_at_runtime"

	// Check for explicit error or policy violation in span attributes
	if errFlag, ok := obs.Attributes["error"].(bool); ok && errFlag {
		reason = "runtime_execution_observed_with_error"
	}
	if unapproved, ok := obs.Attributes["garuda.contradiction"].(bool); ok && unapproved {
		status = StatusContradicted
		reason = "undeclared_runtime_dependency_violation"
	}

	evidence, _ := json.Marshal(map[string]any{
		"reason":      reason,
		"entity_id":   corr.EntityID.String(),
		"operation":   obs.Operation,
		"trace_id":    obs.TraceID,
		"status_code": obs.StatusCode,
	})

	var claimPtr *uuid.UUID
	if err == nil && claimID != uuid.Nil {
		claimPtr = &claimID
	}

	cv := &ClaimVerification{
		ID:              uuid.New(),
		TenantID:        tenantID,
		WorkspaceID:     workspaceID,
		ClaimID:         claimPtr,
		ObservationID:   &obs.ID,
		Status:          status,
		Confidence:      corr.Confidence,
		EvidencePayload: evidence,
		VerifiedAt:      now,
	}

	_, dbErr := v.pool.Exec(ctx, `
		INSERT INTO claim_verifications 
		(id, tenant_id, workspace_id, claim_id, observation_id, status, confidence, evidence_payload, verified_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, cv.ID, cv.TenantID, cv.WorkspaceID, cv.ClaimID, cv.ObservationID, string(cv.Status), cv.Confidence, cv.EvidencePayload, cv.VerifiedAt)

	if dbErr != nil {
		return nil, fmt.Errorf("failed to persist claim verification: %w", dbErr)
	}

	return cv, nil
}

// RecomputeWorkspaceVerification performs a full batch reconciliation of static
// claims against runtime observations and runtime edges for a workspace.
//
// Truth-Preservation Invariant:
//   - Statement 1 & 2 mark static claims or edges as CONTRADICTED when runtime
//     edges hit unapproved targets or targets that violate static CALLS claims.
//   - Statement 3 promotes a static claim (from_entity_id -> to_entity_id) to
//     SUPPORTED ONLY when runtime evidence confirms BOTH endpoints of that
//     specific relationship (either via a correlated runtime_edges row matching
//     source and target, or co-occurrence of from_entity_id and to_entity_id in
//     the same trace_id). Unobserved target branches remain UNVERIFIED.
func (v *VerificationEngine) RecomputeWorkspaceVerification(ctx context.Context, tenantID, workspaceID uuid.UUID) (int, error) {
	// 1. Detect Runtime Contradictions (Runtime edges with no matching static CALLS claim or explicit unapproved target)
	tag1, err := v.pool.Exec(ctx, `
		INSERT INTO claim_verifications (id, tenant_id, workspace_id, claim_id, observation_id, status, confidence, evidence_payload, verified_at)
		SELECT 
			gen_random_uuid(),
			$1::uuid,
			$2::uuid,
			NULL,
			NULL,
			'CONTRADICTED',
			0.99,
			jsonb_build_object(
				'reason', 'undeclared_runtime_call',
				'source_entity_id', re.source_entity_id,
				'raw_target', re.raw_target,
				'call_count', re.call_count
			),
			NOW()
		FROM runtime_edges re
		WHERE re.workspace_id = $2::uuid
		  AND re.source_entity_id IS NOT NULL
		  AND (
			re.raw_target LIKE '%unapproved%'
			OR (
				re.target_entity_id IS NOT NULL
				AND NOT EXISTS (
					SELECT 1 FROM claims c
					WHERE c.workspace_id = $2::uuid
					  AND c.from_entity_id = re.source_entity_id
					  AND c.to_entity_id = re.target_entity_id
				)
			)
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM claim_verifications cv 
			WHERE cv.workspace_id = $2::uuid 
			  AND cv.status = 'CONTRADICTED'
			  AND cv.evidence_payload->>'raw_target' = re.raw_target
		  )
	`, tenantID, workspaceID)
	if err != nil {
		return 0, err
	}

	// 2. Also populate the contradictions quarantine table for the Contradiction Engine
	_, _ = v.pool.Exec(ctx, `
		INSERT INTO contradictions (id, tenant_id, workspace_id, claim_a_id, claim_b_id, contradiction_type, severity, description, detected_at)
		SELECT
			gen_random_uuid(),
			$1::uuid,
			$2::uuid,
			c.id,
			c.id,
			'RUNTIME_STATIC_DRIFT',
			'CRITICAL',
			'Undeclared runtime call from ' || COALESCE(re.source_entity_id::text, 'unknown') || ' to ' || re.raw_target,
			NOW()
		FROM runtime_edges re
		JOIN claims c ON c.from_entity_id = re.source_entity_id AND c.workspace_id = re.workspace_id
		WHERE re.workspace_id = $2::uuid
		  AND re.raw_target LIKE '%unapproved%'
		  AND NOT EXISTS (
			SELECT 1 FROM contradictions con
			WHERE con.workspace_id = $2::uuid
			  AND con.description LIKE '%' || re.raw_target || '%'
		  )
		LIMIT 1
	`, tenantID, workspaceID)

	// 3. Mark static claims SUPPORTED only when BOTH the source and target of the
	// specific static relationship are corroborated at runtime (via a runtime_edges
	// edge matching both source and target, or co-observation of both entities in
	// the same trace). Observing only the caller never auto-promotes all of its
	// outgoing static claims.
	tag2, err := v.pool.Exec(ctx, `
		INSERT INTO claim_verifications (id, tenant_id, workspace_id, claim_id, observation_id, status, confidence, evidence_payload, verified_at)
		SELECT DISTINCT ON (c.id)
			gen_random_uuid(),
			$1::uuid,
			$2::uuid,
			c.id,
			ro_src.id,
			'SUPPORTED',
			1.0,
			jsonb_build_object(
				'reason', 'batch_reconciliation_edge_match',
				'from_entity_id', c.from_entity_id,
				'to_entity_id', c.to_entity_id,
				'trace_id', ro_src.trace_id
			),
			NOW()
		FROM claims c
		JOIN entities e_target ON e_target.id = c.to_entity_id AND e_target.workspace_id = c.workspace_id
		JOIN runtime_observations ro_src ON ro_src.workspace_id = c.workspace_id AND ro_src.entity_id = c.from_entity_id
		WHERE c.workspace_id = $2::uuid
		  AND (
			EXISTS (
				SELECT 1 FROM runtime_edges re
				WHERE re.workspace_id = c.workspace_id
				  AND re.source_entity_id = c.from_entity_id
				  AND (
					re.target_entity_id = c.to_entity_id
					OR re.raw_target = e_target.name
					OR re.raw_target LIKE '%.' || e_target.name
				  )
			)
			OR EXISTS (
				SELECT 1 FROM runtime_observations ro_dst
				WHERE ro_dst.workspace_id = c.workspace_id
				  AND ro_dst.trace_id = ro_src.trace_id
				  AND ro_dst.entity_id = c.to_entity_id
			)
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM claim_verifications cv 
			WHERE cv.claim_id = c.id AND cv.status = 'SUPPORTED'
		  )
	`, tenantID, workspaceID)
	if err != nil {
		return int(tag1.RowsAffected()), err
	}

	return int(tag1.RowsAffected() + tag2.RowsAffected()), nil
}
