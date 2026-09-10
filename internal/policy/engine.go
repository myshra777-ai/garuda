// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Engine coordinates policy loading, evaluation, and persistence.
type Engine struct {
	pool      *pgxpool.Pool
	evaluator *Evaluator
	anchors   *Anchor
}

func NewEngine(pool *pgxpool.Pool) *Engine {
	return &Engine{
		pool:      pool,
		evaluator: NewEvaluator(pool),
		anchors:   NewAnchor(pool),
	}
}

// RunResult is the aggregate outcome of one policy pass.
type RunResult struct {
	WorkspaceID   uuid.UUID     `json:"workspace_id"`
	Evaluations   []*Evaluation `json:"evaluations"`
	FinalDecision Decision      `json:"final_decision"` // highest-severity across all
	BlockedBy     *uuid.UUID    `json:"blocked_by,omitempty"`
	Summary       string        `json:"summary"`
}

// Run loads all policies under policyDir, evaluates them against the workspace,
// persists the evaluations, and anchors them to the Merkle ledger.
//
// The final decision is the highest-severity decision across all evaluations.
// BLOCK > REVIEW > WARN > ALLOW.
func (e *Engine) Run(
	ctx context.Context,
	tenantID, workspaceID uuid.UUID,
	policyDir string,
	subjectKind, subjectID, actor string,
) (*RunResult, error) {

	policies, hashes, err := ParseDirectory(policyDir)
	if err != nil {
		return nil, fmt.Errorf("load policies: %w", err)
	}

	// Sort by priority descending (higher = evaluated first)
	sort.SliceStable(policies, func(i, j int) bool {
		return policies[i].Priority > policies[j].Priority
	})

	result := &RunResult{
		WorkspaceID:   workspaceID,
		FinalDecision: DecisionAllow,
	}

	for _, p := range policies {
		// Skip expired policies
		if p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt) {
			slog.Debug("policy expired, skipping", "policy_id", p.ID)
			continue
		}

		// Filter by language if the policy targets a specific one
		if p.Language != "" && p.Language != "any" {
			// Quick existence check — cheaper than full evaluation
			hasLang, _, _ := e.evaluator.evalLanguageMatches(
				ctx, p, &Predicate{Type: "language_matches", Params: map[string]any{"language": p.Language}},
				tenantID, workspaceID,
			)
			if !hasLang {
				continue
			}
		}

		ev, err := e.evaluator.Evaluate(ctx, p, tenantID, workspaceID, subjectKind, subjectID, actor)
		if err != nil {
			slog.Error("policy evaluation failed",
				"policy_id", p.ID, "workspace_id", workspaceID, "error", err)
			continue
		}
		ev.EvaluatedAt = time.Now().UTC()
		// Persist the policy row (idempotent upsert by policy_id)
		policyUUID, err := e.upsertPolicy(ctx, tenantID, p, hashes)
		if err != nil {
			slog.Error("failed to upsert policy", "policy_id", p.ID, "error", err)
			continue
		}
		ev.PolicyID = policyUUID

		// Anchor to Merkle BEFORE persisting the evaluation, so we can store the block height
		blockHeight, proof, err := e.anchors.AppendEvaluation(ctx, ev)
		if err != nil {
			slog.Error("failed to anchor evaluation", "policy_id", p.ID, "error", err)
			// Continue without anchor — the evaluation is still recorded
		} else {
			ev.MerkleBlockHeight = &blockHeight
			_ = proof // proof is also stored in the DB row
		}

		if err := e.persistEvaluation(ctx, ev, proof); err != nil {
			slog.Error("failed to persist evaluation", "policy_id", p.ID, "error", err)
			continue
		}

		result.Evaluations = append(result.Evaluations, ev)

		// Update final decision
		if ev.Decision.Severity() > result.FinalDecision.Severity() {
			result.FinalDecision = ev.Decision
			id := ev.PolicyID
			result.BlockedBy = &id
		}
	}

	result.Summary = fmt.Sprintf(
		"workspace=%s policies=%d evaluations=%d final=%s",
		workspaceID, len(policies), len(result.Evaluations), result.FinalDecision,
	)
	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Persistence
// ─────────────────────────────────────────────────────────────────────────────

func (e *Engine) upsertPolicy(
	ctx context.Context,
	tenantID uuid.UUID,
	p *Policy,
	hashes map[string]string,
) (uuid.UUID, error) {

	// Look up source path and hash (if any) from the hashes map
	var sourcePath, sourceHash string
	for path, h := range hashes {
		if h == hashPolicyYAML(p) {
			sourcePath = path
			sourceHash = h
			break
		}
	}

	// Find existing policy by tenant + external id
	var id uuid.UUID
	err := e.pool.QueryRow(ctx, `
		SELECT id FROM policies
		WHERE tenant_id = $1 AND statement = $2
		LIMIT 1
	`, tenantID, "POLICY:"+p.ID).Scan(&id)
	if err == nil {
		// Update existing
		_, err = e.pool.Exec(ctx, `
			UPDATE policies
			SET scope_domain = $2,
			    scope_system = $3,
			    actor = $4,
			    metadata = $5,
			    updated_at = NOW()
			WHERE id = $1
		`, id, p.Scope.Domain, p.Scope.System, p.Authority, policyMetadataJSON(p))
		return id, err
	}

	// Insert new
	err = e.pool.QueryRow(ctx, `
		INSERT INTO policies (
			tenant_id, statement, scope_domain, scope_system,
			actor, status, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, 'active', $6, NOW(), NOW()
		) RETURNING id
	`, tenantID,
		"POLICY:"+p.ID,
		p.Scope.Domain,
		p.Scope.System,
		p.Authority,
		policyMetadataJSON(p),
	).Scan(&id)
	_ = sourcePath
	_ = sourceHash
	return id, err
}

func (e *Engine) persistEvaluation(ctx context.Context, ev *Evaluation, proof []byte) error {
	evidenceJSON, err := json.Marshal(ev.Evidence)
	if err != nil {
		return fmt.Errorf("marshal evidence: %w", err)
	}
	_, err = e.pool.Exec(ctx, `
		INSERT INTO policy_evaluations (
			id, tenant_id, workspace_id, policy_id, policy_version,
			decision, reason, evidence, evaluated_at,
			merkle_block_height, merkle_proof,
			subject_kind, subject_id, actor
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`,
		ev.ID, ev.TenantID, ev.WorkspaceID, ev.PolicyID, ev.PolicyVersion,
		string(ev.Decision), ev.Reason, evidenceJSON, ev.EvaluatedAt,
		ev.MerkleBlockHeight, proof,
		ev.SubjectKind, ev.SubjectID, ev.Actor,
	)
	if err != nil {
		return fmt.Errorf("insert evaluation: %w", err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func policyMetadataJSON(p *Policy) []byte {
	b, _ := json.Marshal(map[string]any{
		"external_id": p.ID,
		"title":       p.Title,
		"description": p.Description,
		"priority":    p.Priority,
		"language":    p.Language,
		"version":     p.Version,
		"when":        p.When,
		"then":        p.Then,
	})
	return b
}

func hashPolicyYAML(p *Policy) string {
	// Placeholder — the parser computes the real hash. Kept here for
	// symmetry with ParseDirectory so we can look it up if needed.
	return ""
}
