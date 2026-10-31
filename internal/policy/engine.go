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
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/store"
)

// Engine coordinates policy loading, evaluation, and persistence.
type Engine struct {
	pool      *pgxpool.Pool
	evaluator *Evaluator
	anchors   *Anchor
}

// LoadSupersededPolicyIDs returns the set of policy IDs (lowercase) whose latest
// status in the tenant's policies table is 'superseded', 'deprecated', or 'disabled'.
func LoadSupersededPolicyIDs(ctx context.Context, pool *pgxpool.Pool, tenantID uuid.UUID) map[string]bool {
	superseded := make(map[string]bool)
	if pool == nil || tenantID == uuid.Nil {
		return superseded
	}
	rows, err := pool.Query(ctx, `
		SELECT COALESCE(statement, ''), COALESCE(status, 'active')
		FROM policies
		WHERE tenant_id = $1
		ORDER BY created_at ASC
	`, tenantID)
	if err != nil {
		return superseded
	}
	defer rows.Close()
	latestStatus := make(map[string]string)
	for rows.Next() {
		var stmt, st string
		if rows.Scan(&stmt, &st) == nil {
			key := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(stmt, "POLICY:")))
			if key != "" {
				latestStatus[key] = strings.ToLower(strings.TrimSpace(st))
			}
		}
	}
	for k, st := range latestStatus {
		if st == "superseded" || st == "deprecated" || st == "disabled" || st == "archived" {
			superseded[k] = true
		}
	}
	return superseded
}

// LoadActivePolicies is the Single Policy Registry across policy.list, policy.evaluate,
// briefing, docs_graph, talon_pre_edit, and talon_verify_edit (fixes B-03, B-10, B-16).
// It merges YAML policies from policyDir with active database policies (such as
// warn-missing-runtime), excludes superseded policies, and assigns one canonical
// title per policy ID.
func LoadActivePolicies(ctx context.Context, pool *pgxpool.Pool, tenantID uuid.UUID, policyDir string) ([]ParsedPolicy, []string, []string, error) {
	supersededMap := LoadSupersededPolicyIDs(ctx, pool, tenantID)
	supersededSet := make(map[string]bool)
	for k := range supersededMap {
		supersededSet[k] = true
	}

	byID := make(map[string]ParsedPolicy)
	if policyDir != "" {
		if parsed, err := ParseDirectory(policyDir); err == nil {
			for _, pp := range parsed {
				if pp.Policy == nil {
					continue
				}
				pid := strings.ToLower(strings.TrimSpace(pp.Policy.ID))
				if pid == "" {
					continue
				}
				if supersededMap[pid] {
					supersededSet[pid] = true
					continue
				}
				if pp.Policy.ExpiresAt != nil && time.Now().After(*pp.Policy.ExpiresAt) {
					supersededSet[pid] = true
					continue
				}
				pp.Policy.ID = pid
				if strings.TrimSpace(pp.Policy.Title) == "" {
					pp.Policy.Title = pid
				}
				byID[pid] = pp
			}
		}
	}

	if pool != nil && tenantID != uuid.Nil {
		rows, err := pool.Query(ctx, `
			SELECT COALESCE(statement, ''), COALESCE(scope_domain, ''),
			       COALESCE(scope_system, ''), COALESCE(actor, ''), COALESCE(status, 'active'),
			       COALESCE(metadata, '{}'::jsonb), COALESCE(source_path, ''), COALESCE(source_hash, '')
			FROM policies
			WHERE tenant_id = $1
			ORDER BY created_at ASC
		`, tenantID)
		if err == nil {
			for rows.Next() {
				var stmt, sDom, sSys, actor, status, srcPath, srcHash string
				var metaBytes []byte
				if rows.Scan(&stmt, &sDom, &sSys, &actor, &status, &metaBytes, &srcPath, &srcHash) != nil {
					continue
				}
				pid := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(stmt, "POLICY:")))
				if pid == "" {
					continue
				}
				if strings.EqualFold(status, "superseded") || supersededMap[pid] {
					supersededSet[pid] = true
					delete(byID, pid)
					continue
				}
				if !strings.EqualFold(status, "active") {
					continue
				}

				var meta map[string]any
				_ = json.Unmarshal(metaBytes, &meta)
				metaTitle, _ := meta["title"].(string)
				metaDesc, _ := meta["description"].(string)
				metaVer, _ := meta["version"].(string)
				if metaVer == "" {
					metaVer = "v1"
				}

				// If already loaded from YAML, keep YAML definition and synchronize canonical title
				if existing, ok := byID[pid]; ok {
					if existing.Policy.Title == "" && metaTitle != "" {
						existing.Policy.Title = metaTitle
						byID[pid] = existing
					}
					continue
				}

				// Reconstruct active DB-only policy (e.g. warn-missing-runtime) so it is NEVER skipped (fixes B-03 & B-10)
				title := strings.TrimSpace(metaTitle)
				if title == "" {
					if pid == "warn-missing-runtime" {
						title = "Exported functions should have runtime evidence"
					} else {
						title = pid
					}
				}
				pol := &Policy{
					ID:          pid,
					Version:     metaVer,
					Title:       title,
					Description: metaDesc,
					Priority:    100,
					Language:    "go",
					Authority:   actor,
					Scope: Scope{
						Domain: sDom,
						System: sSys,
					},
					When: []Predicate{
						{
							Type: "verification_missing",
							Params: map[string]any{
								"kind":     "function",
								"exported": true,
								"limit":    25,
							},
						},
					},
					Then: Outcome{
						Decision: DecisionWarn,
						Reason:   title,
					},
				}
				if srcPath == "" {
					srcPath = "policies/" + pid + ".yaml"
				}
				byID[pid] = ParsedPolicy{
					Policy:     pol,
					SourcePath: srcPath,
					SourceHash: srcHash,
				}
			}
			rows.Close()
		}
	}

	activeList := make([]ParsedPolicy, 0, len(byID))
	activeIDs := make([]string, 0, len(byID))
	for pid, pp := range byID {
		activeIDs = append(activeIDs, pid)
		activeList = append(activeList, pp)
	}
	sort.Strings(activeIDs)
	sort.SliceStable(activeList, func(i, j int) bool {
		if activeList[i].Policy.Priority != activeList[j].Policy.Priority {
			return activeList[i].Policy.Priority > activeList[j].Policy.Priority
		}
		return activeList[i].Policy.ID < activeList[j].Policy.ID
	})

	supersededIDs := make([]string, 0, len(supersededSet))
	for k := range supersededSet {
		supersededIDs = append(supersededIDs, k)
	}
	sort.Strings(supersededIDs)

	return activeList, activeIDs, supersededIDs, nil
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
	WorkspaceID             uuid.UUID                        `json:"workspace_id"`
	Evaluations             []*Evaluation                    `json:"evaluations"`
	FinalDecision           Decision                         `json:"final_decision"` // highest-severity across all
	BlockedBy               *uuid.UUID                       `json:"blocked_by,omitempty"`
	BlockedByPolicyID       string                           `json:"blocked_by_policy_id,omitempty"`
	BlockedByTitle          string                           `json:"blocked_by_title,omitempty"`
	NetNewViolationsCount   int                              `json:"net_new_violations_count"`
	BaselineViolationsCount int                              `json:"baseline_violations_count"`
	Summary                 string                           `json:"summary"`
	ReconcileCount          int                              `json:"reconcile_count"`
	Preview                 bool                             `json:"preview"`
	BuildContext            *analyzer.BuildContext           `json:"build_context,omitempty"`
	EnvironmentDrift        *analyzer.EnvironmentDriftReport `json:"environment_drift,omitempty"`
}

// RunOptions tunes the behavior of one policy pass.
type RunOptions struct {
	DryRun          bool
	Reconcile       bool
	DiffScope       *DiffScopeFilter
	RepoDir         string
	BuildTarget     *analyzer.BuildTargetSpec
	BuildContext    *analyzer.BuildContext
	IncludeFixtures bool
}

// Run loads all active policies from the Single Policy Registry (YAML + DB),
// evaluates every single active policy against the workspace (guaranteeing
// len(Evaluations) == len(activePolicies), fixing B-03), prevents vacuous ALLOW
// when 0 entities match scope (fixing B-04), and anchors non-dry-run evaluations.
func (e *Engine) Run(
	ctx context.Context,
	tenantID, workspaceID uuid.UUID,
	policyDir string,
	subjectKind, subjectID, actor string,
	opts RunOptions,
) (*RunResult, error) {
	dryRun := opts.DryRun

	activePolicies, _, _, err := LoadActivePolicies(ctx, e.pool, tenantID, policyDir)
	if err != nil {
		return nil, fmt.Errorf("load active policies: %w", err)
	}

	repoDir := opts.RepoDir
	if repoDir == "" && policyDir != "" {
		repoDir = filepath.Dir(policyDir)
	}
	if repoDir == "" {
		repoDir = "."
	}
	wsLabel := subjectID
	if wsLabel == "" {
		wsLabel = "default"
	}

	activeBuildCtx := opts.BuildContext
	if activeBuildCtx == nil {
		activeBuildCtx = analyzer.ResolveBuildContext(repoDir, "", wsLabel, opts.BuildTarget)
	}
	recordedBuildCtx := store.GetLatestRecordedBuildContext(ctx, e.pool, tenantID, workspaceID)
	if recordedBuildCtx != nil && recordedBuildCtx.GraphFingerprint != "" {
		activeBuildCtx.GraphFingerprint = recordedBuildCtx.GraphFingerprint
		activeBuildCtx.ReceiptHash = activeBuildCtx.ComputeReceiptHash()
	}
	driftReport := analyzer.DetectEnvironmentDrift(recordedBuildCtx, activeBuildCtx, wsLabel)

	result := &RunResult{
		WorkspaceID:      workspaceID,
		FinalDecision:    DecisionAllow,
		Preview:          dryRun,
		BuildContext:     activeBuildCtx,
		EnvironmentDrift: driftReport,
	}

	seenStatements := make(map[string]bool, len(activePolicies))
	diffScope := opts.DiffScope
	if opts.IncludeFixtures {
		if diffScope == nil {
			diffScope = &DiffScopeFilter{IncludeFixtures: true}
		} else {
			diffScope.IncludeFixtures = true
		}
	}

	for _, pp := range activePolicies {
		p := pp.Policy
		seenStatements["POLICY:"+p.ID] = true

		ev, evalErr := e.evaluator.EvaluateWithScope(ctx, p, tenantID, workspaceID, subjectKind, subjectID, actor, diffScope)
		if evalErr != nil {
			slog.Error("policy evaluation failed (fail-closed)",
				"policy_id", p.ID, "workspace_id", workspaceID, "error", evalErr)
			return nil, fmt.Errorf("fail-closed: policy %s evaluation failed: %w", p.ID, evalErr)
		}
		ev.PolicyExternalID = p.ID
		ev.PolicyTitle = p.Title
		ev.EvaluatedAt = time.Now().UTC()
		ev.BuildContext = activeBuildCtx
		ev.Evidence.BuildContext = activeBuildCtx

		result.NetNewViolationsCount += ev.NetNewViolationsCount
		result.BaselineViolationsCount += ev.BaselineViolationsCount

		if dryRun {
			if ev.PolicyID == uuid.Nil {
				ev.PolicyID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(p.ID))
			}
			result.Evaluations = append(result.Evaluations, ev)
			if ev.Decision.Severity() > result.FinalDecision.Severity() {
				result.FinalDecision = ev.Decision
				id := ev.PolicyID
				result.BlockedBy = &id
				result.BlockedByPolicyID = p.ID
				result.BlockedByTitle = p.Title
			}
			continue
		}

		policyUUID, uErr := e.upsertPolicy(ctx, tenantID, pp)
		if uErr != nil {
			return nil, fmt.Errorf("fail-closed: failed to persist policy %s: %w", p.ID, uErr)
		}
		ev.PolicyID = policyUUID

		blockHeight, proof, aErr := e.anchors.AppendEvaluation(ctx, ev)
		if aErr != nil {
			return nil, fmt.Errorf("fail-closed: failed to anchor evaluation for policy %s: %w", p.ID, aErr)
		}
		ev.MerkleBlockHeight = &blockHeight

		if pErr := e.persistEvaluation(ctx, ev, proof); pErr != nil {
			return nil, fmt.Errorf("fail-closed: failed to persist evaluation for policy %s: %w", p.ID, pErr)
		}

		result.Evaluations = append(result.Evaluations, ev)
		if ev.Decision.Severity() > result.FinalDecision.Severity() {
			result.FinalDecision = ev.Decision
			id := ev.PolicyID
			result.BlockedBy = &id
			result.BlockedByPolicyID = p.ID
			result.BlockedByTitle = p.Title
		}
	}

	if !dryRun && opts.Reconcile {
		seenList := make([]string, 0, len(seenStatements))
		for s := range seenStatements {
			seenList = append(seenList, s)
		}
		if orphaned, oErr := e.pool.Exec(ctx, `
			UPDATE policies
			   SET status = 'superseded',
			       updated_at = NOW()
			 WHERE tenant_id = $1
			   AND status = 'active'
			   AND NOT (statement = ANY($2::text[]))
		`, tenantID, seenList); oErr == nil {
			result.ReconcileCount = int(orphaned.RowsAffected())
		}
	}

	result.Summary = fmt.Sprintf(
		"workspace=%s policies=%d evaluations=%d final=%s net_new=%d baseline=%d (%s)",
		workspaceID, len(activePolicies), len(result.Evaluations), result.FinalDecision,
		result.NetNewViolationsCount, result.BaselineViolationsCount,
		activeBuildCtx.SummaryLine(),
	)
	return result, nil
}

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
