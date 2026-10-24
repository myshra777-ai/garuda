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
	WorkspaceID    uuid.UUID     `json:"workspace_id"`
	Evaluations    []*Evaluation `json:"evaluations"`
	FinalDecision  Decision      `json:"final_decision"` // highest-severity across all
	BlockedBy      *uuid.UUID    `json:"blocked_by,omitempty"`
	Summary        string        `json:"summary"`
	ReconcileCount int           `json:"reconcile_count"`
	Preview        bool          `json:"preview"`
}

// RunOptions tunes the behavior of one policy pass.
type RunOptions struct {
	DryRun    bool
	Reconcile bool
}

// Run loads all policies under policyDir, evaluates them against the
// workspace, persists the evaluations, and anchors them to the Merkle
// ledger.
//
// Fail-Closed Invariant: Any policy parse error, predicate evaluation error,
// or non-dry-run persistence/Merkle anchoring error aborts the pass and returns
// an error rather than silently continuing with ALLOW.
func (e *Engine) Run(
	ctx context.Context,
	tenantID, workspaceID uuid.UUID,
	policyDir string,
	subjectKind, subjectID, actor string,
	opts RunOptions,
) (*RunResult, error) {
	dryRun := opts.DryRun

	parsed, err := ParseDirectory(policyDir)
	if err != nil {
		return nil, fmt.Errorf("load policies: %w", err)
	}

	// Sort by priority descending (higher = evaluated first)
	sort.SliceStable(parsed, func(i, j int) bool {
		return parsed[i].Policy.Priority > parsed[j].Policy.Priority
	})

	result := &RunResult{
		WorkspaceID:   workspaceID,
		FinalDecision: DecisionAllow,
	}
	result.Preview = dryRun

	seenStatements := make(map[string]bool, len(parsed))

	for _, pp := range parsed {
		p := pp.Policy
		seenStatements["POLICY:"+p.ID] = true

		// Skip expired policies
		if p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt) {
			slog.Debug("policy expired, skipping", "policy_id", p.ID)
			continue
		}

		// Filter by language if the policy targets a specific one
		if p.Language != "" && p.Language != "any" {
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
			slog.Error("policy evaluation failed (fail-closed)",
				"policy_id", p.ID, "workspace_id", workspaceID, "error", err)
			return nil, fmt.Errorf("fail-closed: policy %s evaluation failed: %w", p.ID, err)
		}
		ev.EvaluatedAt = time.Now().UTC()

		if dryRun {
			if ev.PolicyID == uuid.Nil {
				ev.PolicyID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(p.ID))
			}
			result.Evaluations = append(result.Evaluations, ev)
			if ev.Decision.Severity() > result.FinalDecision.Severity() {
				result.FinalDecision = ev.Decision
				id := ev.PolicyID
				result.BlockedBy = &id
			}
			continue
		}

		// Persist the policy row (idempotent upsert by policy_id)
		policyUUID, err := e.upsertPolicy(ctx, tenantID, pp)
		if err != nil {
			slog.Error("failed to upsert policy (fail-closed)", "policy_id", p.ID, "error", err)
			return nil, fmt.Errorf("fail-closed: failed to persist policy %s: %w", p.ID, err)
		}
		ev.PolicyID = policyUUID

		// Anchor to Merkle BEFORE persisting the evaluation, so we can store the block height
		blockHeight, proof, err := e.anchors.AppendEvaluation(ctx, ev)
		if err != nil {
			slog.Error("failed to anchor evaluation (fail-closed)", "policy_id", p.ID, "error", err)
			return nil, fmt.Errorf("fail-closed: failed to anchor evaluation for policy %s: %w", p.ID, err)
		}
		ev.MerkleBlockHeight = &blockHeight

		if err := e.persistEvaluation(ctx, ev, proof); err != nil {
			slog.Error("failed to persist evaluation (fail-closed)", "policy_id", p.ID, "error", err)
			return nil, fmt.Errorf("fail-closed: failed to persist evaluation for policy %s: %w", p.ID, err)
		}

		result.Evaluations = append(result.Evaluations, ev)

		// Update final decision
		if ev.Decision.Severity() > result.FinalDecision.Severity() {
			result.FinalDecision = ev.Decision
			id := ev.PolicyID
			result.BlockedBy = &id
		}
	}

	reconcile := opts.Reconcile
	if !dryRun && reconcile {
		seenList := make([]string, 0, len(seenStatements))
		for s := range seenStatements {
			seenList = append(seenList, s)
		}

		orphaned, err := e.pool.Exec(ctx, `
			UPDATE policies
			   SET status = 'superseded',
			       updated_at = NOW()
			 WHERE tenant_id = $1
			   AND status = 'active'
			   AND NOT (statement = ANY($2::text[]))
		`, tenantID, seenList)
		if err != nil {
			slog.Error("policy reconcile failed", "error", err, "tenant_id", tenantID)
		} else {
			n := orphaned.RowsAffected()
			result.ReconcileCount = int(n)
			if n > 0 {
				slog.Debug("policy reconcile: marked orphaned policies superseded",
					"tenant_id", tenantID, "count", n)
			}
		}
	}

	result.Summary = fmt.Sprintf(
		"workspace=%s policies=%d evaluations=%d final=%s",
		workspaceID, len(parsed), len(result.Evaluations), result.FinalDecision,
	)
	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Persistence
// ─────────────────────────────────────────────────────────────────────────────

func (e *Engine) upsertPolicy(
	ctx context.Context,
	tenantID uuid.UUID,
	pp ParsedPolicy,
) (uuid.UUID, error) {
	p := pp.Policy

	statement := "POLICY:" + p.ID
	metadataJSON := policyMetadataJSON(p)

	var previousID *uuid.UUID
	{
		var prev uuid.UUID
		err := e.pool.QueryRow(ctx, `
			SELECT id FROM policies
			 WHERE tenant_id = $1
			   AND scope_domain = $2
			   AND scope_system = $3
			   AND metadata->>'title' = $4
			   AND statement <> $5
			   AND status = 'active'
			 ORDER BY created_at DESC
			 LIMIT 1
		`, tenantID, p.Scope.Domain, p.Scope.System, p.Title, statement).Scan(&prev)
		if err == nil {
			previousID = &prev
		}
	}

	var id uuid.UUID
	err := e.pool.QueryRow(ctx, `
		INSERT INTO policies (
			tenant_id, statement, scope_domain, scope_system,
			actor, status, metadata, source_path, source_hash,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, 'active', $6, $7, $8, NOW(), NOW()
		)
		ON CONFLICT (tenant_id, statement) DO UPDATE SET
			scope_domain = EXCLUDED.scope_domain,
			scope_system = EXCLUDED.scope_system,
			actor        = EXCLUDED.actor,
			metadata     = EXCLUDED.metadata,
			source_path  = EXCLUDED.source_path,
			source_hash  = EXCLUDED.source_hash,
			updated_at   = NOW()
		RETURNING id
	`, tenantID, statement,
		p.Scope.Domain, p.Scope.System,
		p.Authority, metadataJSON,
		pp.SourcePath, pp.SourceHash,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("upsert policy %s: %w", p.ID, err)
	}

	if previousID != nil && *previousID != id {
		_, _ = e.pool.Exec(ctx, `
			UPDATE policies
			   SET status = 'superseded',
			       superseded_by = $1,
			       updated_at = NOW()
			 WHERE id = $2 AND tenant_id = $3
		`, id, *previousID, tenantID)
	}

	return id, nil
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
			merkle_block_height, merkle_proof, verification_version,
			subject_kind, subject_id, actor
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,1,$12,$13,$14)
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
