// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/merkle"
)

type Engine struct {
	pool      *pgxpool.Pool
	evaluator *Evaluator
	anchors   *merkle.AnchorChain
}

func NewEngine(pool *pgxpool.Pool) *Engine {
	return &Engine{
		pool:      pool,
		evaluator: NewEvaluator(pool),
		anchors:   merkle.NewAnchorChain(pool),
	}
}

type RunResult struct {
	WorkspaceID   uuid.UUID          `json:"workspace_id"`
	WorkspaceName string             `json:"workspace_name"`
	EvaluatedAt   time.Time          `json:"evaluated_at"`
	FinalDecision DecisionStatus     `json:"final_decision"`
	TotalPolicies int                `json:"total_policies"`
	PassedCount   int                `json:"passed_count"`
	WarnCount     int                `json:"warn_count"`
	BlockedCount  int                `json:"blocked_count"`
	Evaluations   []PolicyEvaluation `json:"evaluations"`
	BlockedBy     *uuid.UUID         `json:"blocked_by,omitempty"`
}

func resolveDecisionOverride(ctx context.Context, pool *pgxpool.Pool, tenantID, workspaceID uuid.UUID, policyName string, now time.Time) (*uuid.UUID, bool) {
	if pool == nil {
		return nil, false
	}

	var id uuid.UUID
	err := pool.QueryRow(ctx, `
		SELECT id FROM decisions
		WHERE tenant_id = $1
		  AND workspace_id = $2
		  AND status = 'ACCEPTED'
		  AND (valid_until IS NULL OR valid_until > $3)
		  AND overridden_policies @> to_jsonb($4::text)
		ORDER BY created_at DESC
		LIMIT 1
	`, tenantID, workspaceID, now, policyName).Scan(&id)

	if err == nil && id != uuid.Nil {
		return &id, true
	}
	return nil, false
}

type overrideLookupFunc func(ctx context.Context, policyName string, now time.Time) (*uuid.UUID, bool)

// Run evaluates all loaded policies against the target workspace.
//
// Fail-Closed Invariant:
//   - Any predicate evaluation error aborts the run and returns an error (never
//     defaulting to ALLOW on a failed query).
//   - When dryRun is false, policy registration, Merkle anchor chain appending,
//     and evaluation audit persistence must succeed; any persistence or anchoring
//     failure returns an error rather than silently dropping the audit trail.
func (e *Engine) Run(ctx context.Context, tenantID, workspaceID uuid.UUID, workspaceName string, policies []LoadedPolicy, dryRun bool) (*RunResult, error) {
	var snapshotHash string
	_ = e.pool.QueryRow(ctx, `
		SELECT COALESCE(snapshot_hash, 'unanchored-snapshot') 
		FROM canonical_revisions 
		WHERE workspace_id = $1 
		ORDER BY created_at DESC LIMIT 1
	`, workspaceID).Scan(&snapshotHash)
	if snapshotHash == "" {
		snapshotHash = "genesis-snapshot"
	}

	res := &RunResult{
		WorkspaceID:   workspaceID,
		WorkspaceName: workspaceName,
		EvaluatedAt:   time.Now().UTC(),
		FinalDecision: DecisionAllow,
		TotalPolicies: len(policies),
	}

	lookupOverride := func(ctx context.Context, policyName string, now time.Time) (*uuid.UUID, bool) {
		return resolveDecisionOverride(ctx, e.pool, tenantID, workspaceID, policyName, now)
	}

	for _, p := range policies {
		var policyID uuid.UUID
		if !dryRun {
			var err error
			policyID, err = e.upsertPolicy(ctx, tenantID, workspaceID, p)
			if err != nil {
				return nil, fmt.Errorf("fail-closed: failed to persist policy %q: %w", p.Policy.Name, err)
			}
		} else {
			policyID = uuid.New()
		}

		ev, err := e.evaluator.Evaluate(ctx, tenantID, workspaceID, policyID, p.Policy, snapshotHash)
		if err != nil {
			return nil, fmt.Errorf("fail-closed: policy %q evaluation failed: %w", p.Policy.Name, err)
		}

		applyEvaluationOutcome(res, ev, lookupOverride, ctx)

		if !dryRun {
			leafHash := merkle.ComputeEvaluationLeafHash(
				ev.ID,
				ev.PolicyID,
				ev.WorkspaceID,
				ev.SnapshotHash,
				string(ev.Decision),
				string(ev.EvidencePayload),
				ev.EvaluatedAt,
			)
			anchorID, err := e.anchors.AppendEvaluation(ctx, tenantID, workspaceID, leafHash)
			if err != nil {
				return nil, fmt.Errorf("fail-closed: failed to append Merkle anchor for policy %q: %w", p.Policy.Name, err)
			}
			ev.MerkleAnchorID = &anchorID

			if err := e.persistEvaluation(ctx, ev); err != nil {
				return nil, fmt.Errorf("fail-closed: failed to persist evaluation audit record for policy %q: %w", p.Policy.Name, err)
			}
		}

		res.Evaluations = append(res.Evaluations, *ev)
	}

	return res, nil
}

func applyEvaluationOutcome(res *RunResult, ev *PolicyEvaluation, lookup overrideLookupFunc, ctx context.Context) {
	if ev.Decision == DecisionBlock || ev.Decision == DecisionRequireApproval {
		if lookup != nil {
			if decID, ok := lookup(ctx, ev.PolicyName, res.EvaluatedAt); ok {
				ev.Decision = DecisionOverridden
				ev.OverriddenByDecisionID = decID

				var payload map[string]any
				_ = json.Unmarshal(ev.EvidencePayload, &payload)
				if payload == nil {
					payload = make(map[string]any)
				}
				payload["overridden_by_decision_id"] = decID.String()
				ev.EvidencePayload, _ = json.Marshal(payload)
			}
		}
	}

	switch ev.Decision {
	case DecisionAllow:
		res.PassedCount++
	case DecisionOverridden:
		res.PassedCount++
	case DecisionWarn:
		res.WarnCount++
		if res.FinalDecision == DecisionAllow {
			res.FinalDecision = DecisionWarn
		}
	case DecisionRequireApproval:
		res.WarnCount++
		if res.FinalDecision == DecisionAllow || res.FinalDecision == DecisionWarn {
			res.FinalDecision = DecisionRequireApproval
		}
	case DecisionBlock:
		res.BlockedCount++
		res.FinalDecision = DecisionBlock
		if res.BlockedBy == nil {
			id := ev.ID
			res.BlockedBy = &id
		}
	}
}

func (e *Engine) upsertPolicy(ctx context.Context, tenantID, workspaceID uuid.UUID, lp LoadedPolicy) (uuid.UUID, error) {
	var id uuid.UUID
	err := e.pool.QueryRow(ctx, `
		INSERT INTO policies (id, tenant_id, workspace_id, name, version, severity, action, yaml_content, content_hash, active, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, true, NOW())
		ON CONFLICT (workspace_id, name, content_hash) 
		DO UPDATE SET active = true
		RETURNING id
	`, tenantID, workspaceID, lp.Policy.Name, lp.Policy.Version, string(lp.Policy.Severity), string(lp.Policy.Action), lp.Content, lp.Hash).Scan(&id)
	return id, err
}

func (e *Engine) persistEvaluation(ctx context.Context, ev *PolicyEvaluation) error {
	_, err := e.pool.Exec(ctx, `
		INSERT INTO policy_evaluations 
		(id, tenant_id, workspace_id, policy_id, snapshot_hash, decision, violation_count, evidence_payload, merkle_anchor_id, overridden_by_decision_id, evaluated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, ev.ID, ev.TenantID, ev.WorkspaceID, ev.PolicyID, ev.SnapshotHash, string(ev.Decision), ev.ViolationCount, ev.EvidencePayload, ev.MerkleAnchorID, ev.OverriddenByDecisionID, ev.EvaluatedAt)
	return err
}
