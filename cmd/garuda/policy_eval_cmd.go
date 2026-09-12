// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"

	"github.com/myshra777-ai/garuda/internal/policy"
)

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Deterministic policy enforcement: list, validate, evaluate, verify",
	Long: `Garuda's policy enforcement layer.

A policy is a declarative YAML file expressing intent (authority, scope, rules).
Every evaluation produces an enforcement decision (ALLOW / WARN / REVIEW / BLOCK)
with evidence, anchored to the tenant's Merkle ledger.

Commands:
  list                          List active policies in the DB
  validate <dir>                Parse + validate all policies without evaluating
  evaluate <dir>                Run all policies against the workspace, persist results
  show <evaluation-id>          Show a specific evaluation with evidence
  verify <evaluation-id>        Verify the Merkle anchor for an evaluation`,
}

// ─────────────────────────────────────────────────────────────────────────────
// policy list
// ─────────────────────────────────────────────────────────────────────────────

var policyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active policies registered in the workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		pool, err := openPolicyPool()
		if err != nil {
			return err
		}
		defer pool.Close()

		tenantID := getTenantID()

		rows, err := pool.Query(context.Background(), `
			SELECT id, statement, scope_domain, scope_system, actor, status, created_at
			FROM policies
			WHERE tenant_id = $1
			ORDER BY created_at DESC
		`, tenantID)
		if err != nil {
			return fmt.Errorf("query policies: %w", err)
		}
		defer rows.Close()

		count := 0
		fmt.Printf("%-38s %-32s %-12s %-12s %-10s\n", "ID", "STATEMENT", "DOMAIN", "SYSTEM", "STATUS")
		fmt.Println(strings.Repeat("─", 110))
		for rows.Next() {
			var id uuid.UUID
			var statement, domain, system, actor, status string
			var createdAt time.Time
			if err := rows.Scan(&id, &statement, &domain, &system, &actor, &status, &createdAt); err != nil {
				continue
			}
			stmt := statement
			if len(stmt) > 30 {
				stmt = stmt[:27] + "..."
			}
			fmt.Printf("%-38s %-32s %-12s %-12s %-10s\n", id, stmt, domain, system, status)
			count++
		}
		fmt.Printf("\n%d policies\n", count)
		return nil
	},
}

// ─────────────────────────────────────────────────────────────────────────────
// policy validate <dir>
// ─────────────────────────────────────────────────────────────────────────────

var policyValidateCmd = &cobra.Command{
	Use:   "validate [dir]",
	Short: "Parse and validate all policy YAML files without evaluating",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		info, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf("policy dir not found: %w", err)
		}
		if !info.IsDir() {
			return fmt.Errorf("%s is not a directory", dir)
		}

		policies, hashes, err := policy.ParseDirectory(dir)
		if err != nil {
			return fmt.Errorf("parse directory: %w", err)
		}

		if len(policies) == 0 {
			fmt.Println("⚠️  No .yaml or .yml policy files found.")
			return nil
		}

		fmt.Printf("Validated %d policies under %s\n\n", len(policies), dir)
		for _, p := range policies {
			var sourcePath string
			for path, h := range hashes {
				if strings.Contains(path, p.ID) {
					sourcePath = path
					_ = h
					break
				}
			}
			fmt.Printf("✓ %s\n", p.ID)
			fmt.Printf("    title:     %s\n", p.Title)
			fmt.Printf("    version:   %s\n", p.Version)
			fmt.Printf("    priority:  %d\n", p.Priority)
			fmt.Printf("    language:  %s\n", p.Language)
			fmt.Printf("    authority: %s\n", p.Authority)
			fmt.Printf("    decision:  %s\n", p.Then.Decision)
			fmt.Printf("    predicates: %d\n", len(p.When))
			for i, pred := range p.When {
				fmt.Printf("        [%d] %s\n", i+1, pred.Type)
			}
			if sourcePath != "" {
				fmt.Printf("    source:    %s\n", sourcePath)
			}
			fmt.Println()
		}
		return nil
	},
}

// ─────────────────────────────────────────────────────────────────────────────
// policy evaluate <dir>
// ─────────────────────────────────────────────────────────────────────────────

var (
	policyWorkspaceFlag string
	policySubjectKind   string
	policySubjectID     string
	policyActorFlag     string
	policyJSONFlag      bool
	policyFailOnBlock   bool
)

var policyEvaluateCmd = &cobra.Command{
	Use:   "evaluate [dir]",
	Short: "Run all policies against the workspace and anchor decisions to the ledger",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]

		pool, err := openPolicyPool()
		if err != nil {
			return err
		}
		defer pool.Close()

		tenantID := getTenantID()

		wsName := policyWorkspaceFlag
		if wsName == "" {
			wsName = os.Getenv("GARUDA_WORKSPACE")
		}
		if wsName == "" {
			wsName = "default"
		}

		var workspaceID uuid.UUID
		err = pool.QueryRow(context.Background(), `
			SELECT id FROM workspaces WHERE tenant_id = $1 AND name = $2 LIMIT 1
		`, tenantID, wsName).Scan(&workspaceID)
		if err != nil {
			return fmt.Errorf("workspace %q not found: %w", wsName, err)
		}

		actor := policyActorFlag
		if actor == "" {
			actor = os.Getenv("GARUDA_AGENT")
		}
		if actor == "" {
			actor = "cli-operator"
		}

		ctx := context.Background()
		engine := policy.NewEngine(pool)

		start := time.Now()
		result, err := engine.Run(ctx, tenantID, workspaceID, dir,
			policySubjectKind, policySubjectID, actor)
		if err != nil {
			return fmt.Errorf("engine run failed: %w", err)
		}
		elapsed := time.Since(start)

		if policyJSONFlag {
			out, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(out))
		} else {
			printPolicyRunResult(result, elapsed, wsName)
		}

		if policyFailOnBlock && result.FinalDecision == policy.DecisionBlock {
			os.Exit(2)
		}
		return nil
	},
}

func printPolicyRunResult(r *policy.RunResult, elapsed time.Duration, wsName string) {
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║               GARUDA POLICY ENFORCEMENT                            ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════╝")
	fmt.Printf("Workspace:   %s\n", wsName)
	fmt.Printf("Duration:    %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("Policies:    %d evaluated\n", len(r.Evaluations))
	fmt.Println()

	if len(r.Evaluations) == 0 {
		fmt.Println("No policies fired. All checks passed or none applied.")
		return
	}

	// Group by decision, highest severity first
	sorted := append([]*policy.Evaluation{}, r.Evaluations...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Decision.Severity() > sorted[j].Decision.Severity()
	})

	for _, ev := range sorted {
		var badge string
		switch ev.Decision {
		case policy.DecisionBlock:
			badge = "🛑 BLOCK"
		case policy.DecisionReview:
			badge = "🔍 REVIEW"
		case policy.DecisionWarn:
			badge = "⚠️  WARN"
		default:
			badge = "✅ ALLOW"
		}

		fmt.Printf("%s  policy=%s\n", badge, shortUUID(ev.PolicyID))
		fmt.Printf("      reason: %s\n", strings.TrimSpace(ev.Reason))
		if len(ev.Evidence.MatchedPredicates) > 0 {
			fmt.Printf("      matched: %s\n", strings.Join(ev.Evidence.MatchedPredicates, ", "))
		}
		fmt.Printf("      entities: %d  claims: %d  contradictions: %d\n",
			len(ev.Evidence.EntityIDs), len(ev.Evidence.ClaimIDs), len(ev.Evidence.ContradictionIDs))
		if ev.MerkleBlockHeight != nil {
			fmt.Printf("      merkle:   block #%d\n", *ev.MerkleBlockHeight)
		} else {
			fmt.Printf("      merkle:   (unanchored)\n")
		}
		if len(ev.Evidence.ReasoningNotes) > 0 {
			fmt.Printf("      evidence:\n")
			max := 3
			for i, note := range ev.Evidence.ReasoningNotes {
				if i >= max {
					fmt.Printf("        ... and %d more\n", len(ev.Evidence.ReasoningNotes)-max)
					break
				}
				fmt.Printf("        - %s\n", note)
			}
		}
		fmt.Println()
	}

	fmt.Println("───")
	fmt.Printf("FINAL DECISION: %s\n", r.FinalDecision)
	if r.BlockedBy != nil {
		fmt.Printf("Blocked by:     %s\n", *r.BlockedBy)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// policy show <evaluation-id>
// ─────────────────────────────────────────────────────────────────────────────

var policyShowCmd = &cobra.Command{
	Use:   "show [evaluation-id]",
	Short: "Show a specific policy evaluation with evidence",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pool, err := openPolicyPool()
		if err != nil {
			return err
		}
		defer pool.Close()

		evalID, err := uuid.Parse(args[0])
		if err != nil {
			return fmt.Errorf("invalid evaluation id: %w", err)
		}

		var ev policy.Evaluation
		var evidenceJSON []byte
		var decision, reason string

		err = pool.QueryRow(context.Background(), `
			SELECT id, tenant_id, workspace_id, policy_id, policy_version,
			       decision, reason, evidence, evaluated_at,
			       merkle_block_height, merkle_proof,
			       subject_kind, subject_id, actor
			FROM policy_evaluations
			WHERE id = $1
		`, evalID).Scan(
			&ev.ID, &ev.TenantID, &ev.WorkspaceID, &ev.PolicyID, &ev.PolicyVersion,
			&decision, &reason, &evidenceJSON, &ev.EvaluatedAt,
			&ev.MerkleBlockHeight, &ev.MerkleProof,
			&ev.SubjectKind, &ev.SubjectID, &ev.Actor,
		)
		if err != nil {
			return fmt.Errorf("evaluation not found: %w", err)
		}
		ev.Decision = policy.Decision(decision)
		ev.Reason = reason
		_ = json.Unmarshal(evidenceJSON, &ev.Evidence)

		fmt.Println()
		fmt.Println("POLICY EVALUATION")
		fmt.Println(strings.Repeat("═", 65))
		fmt.Printf("  ID:           %s\n", ev.ID)
		fmt.Printf("  Policy:       %s (v%s)\n", ev.PolicyID, ev.PolicyVersion)
		fmt.Printf("  Workspace:    %s\n", ev.WorkspaceID)
		fmt.Printf("  Subject:      %s/%s\n", ev.SubjectKind, ev.SubjectID)
		fmt.Printf("  Actor:        %s\n", ev.Actor)
		fmt.Printf("  Evaluated:    %s\n", ev.EvaluatedAt.Format(time.RFC3339))
		fmt.Println()
		fmt.Printf("  Decision:     %s\n", ev.Decision)
		fmt.Printf("  Reason:       %s\n", strings.TrimSpace(ev.Reason))
		fmt.Println()
		fmt.Println("  Evidence:")
		fmt.Printf("    Claims:         %d\n", len(ev.Evidence.ClaimIDs))
		for _, id := range ev.Evidence.ClaimIDs {
			fmt.Printf("      - %s\n", id)
		}
		fmt.Printf("    Entities:       %d\n", len(ev.Evidence.EntityIDs))
		for _, id := range ev.Evidence.EntityIDs {
			fmt.Printf("      - %s\n", id)
		}
		fmt.Printf("    Contradictions: %d\n", len(ev.Evidence.ContradictionIDs))
		for _, id := range ev.Evidence.ContradictionIDs {
			fmt.Printf("      - %s\n", id)
		}
		if len(ev.Evidence.MatchedPredicates) > 0 {
			fmt.Printf("    Predicates:     %s\n", strings.Join(ev.Evidence.MatchedPredicates, ", "))
		}
		if len(ev.Evidence.ReasoningNotes) > 0 {
			fmt.Println("    Notes:")
			for _, n := range ev.Evidence.ReasoningNotes {
				fmt.Printf("      - %s\n", n)
			}
		}
		fmt.Println()
		if ev.MerkleBlockHeight != nil {
			fmt.Printf("  Merkle:       block #%d\n", *ev.MerkleBlockHeight)
			fmt.Printf("  Proof bytes:  %d\n", len(ev.MerkleProof))
		} else {
			fmt.Printf("  Merkle:       (unanchored)\n")
		}
		fmt.Println()
		return nil
	},
}

// ─────────────────────────────────────────────────────────────────────────────
// policy verify <evaluation-id>
// ─────────────────────────────────────────────────────────────────────────────

var policyVerifyCmd = &cobra.Command{
	Use:   "verify [evaluation-id]",
	Short: "Verify the Merkle anchor for a policy evaluation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pool, err := openPolicyPool()
		if err != nil {
			return err
		}
		defer pool.Close()

		evalID, err := uuid.Parse(args[0])
		if err != nil {
			return fmt.Errorf("invalid evaluation id: %w", err)
		}

		ev, err := loadEvaluation(context.Background(), pool, evalID)
		if err != nil {
			return err
		}

		anchor := policy.NewAnchor(pool)
		ok, err := anchor.VerifyEvaluation(context.Background(), ev)
		if err != nil {
			fmt.Printf("❌ Verification FAILED: %v\n", err)
			os.Exit(1)
		}
		if ok {
			fmt.Printf("✅ Merkle anchor VALID for evaluation %s at block #%d\n",
				ev.ID, *ev.MerkleBlockHeight)
		} else {
			fmt.Printf("❌ Merkle anchor INVALID for evaluation %s\n", ev.ID)
			os.Exit(1)
		}
		return nil
	},
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func openPolicyPool() (*pgxpool.Pool, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable must be set")
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	return pool, nil
}

func shortUUID(id uuid.UUID) string {
	s := id.String()
	if len(s) >= 8 {
		return s[:8]
	}
	return s
}

func loadEvaluation(ctx context.Context, pool *pgxpool.Pool, evalID uuid.UUID) (*policy.Evaluation, error) {
	var ev policy.Evaluation
	var evidenceJSON []byte
	var decision, reason string
	err := pool.QueryRow(ctx, `
		SELECT id, tenant_id, workspace_id, policy_id, policy_version,
		       decision, reason, evidence, evaluated_at,
		       merkle_block_height, merkle_proof,
		       subject_kind, subject_id, actor
		FROM policy_evaluations WHERE id = $1
	`, evalID).Scan(
		&ev.ID, &ev.TenantID, &ev.WorkspaceID, &ev.PolicyID, &ev.PolicyVersion,
		&decision, &reason, &evidenceJSON, &ev.EvaluatedAt,
		&ev.MerkleBlockHeight, &ev.MerkleProof,
		&ev.SubjectKind, &ev.SubjectID, &ev.Actor,
	)
	if err != nil {
		return nil, fmt.Errorf("evaluation not found: %w", err)
	}
	ev.Decision = policy.Decision(decision)
	ev.Reason = reason
	_ = json.Unmarshal(evidenceJSON, &ev.Evidence)
	return &ev, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Registration
// ─────────────────────────────────────────────────────────────────────────────

func init() {
	policyEvaluateCmd.Flags().StringVarP(&policyWorkspaceFlag, "workspace", "w", "", "Workspace name (default: GARUDA_WORKSPACE env)")
	policyEvaluateCmd.Flags().StringVar(&policySubjectKind, "subject-kind", "manual", "Subject kind (commit | pr | agent_action | manual)")
	policyEvaluateCmd.Flags().StringVar(&policySubjectID, "subject-id", "", "Subject identifier (commit SHA, PR number, etc.)")
	policyEvaluateCmd.Flags().StringVar(&policyActorFlag, "actor", "", "Actor (default: GARUDA_AGENT env)")
	policyEvaluateCmd.Flags().BoolVar(&policyJSONFlag, "json", false, "Output JSON")
	policyEvaluateCmd.Flags().BoolVar(&policyFailOnBlock, "fail-on-block", false, "Exit non-zero if any policy returns BLOCK")

	policyCmd.AddCommand(policyListCmd)
	policyCmd.AddCommand(policyValidateCmd)
	policyCmd.AddCommand(policyEvaluateCmd)
	policyCmd.AddCommand(policyShowCmd)
	policyCmd.AddCommand(policyVerifyCmd)

	rootCmd.AddCommand(policyCmd)
	_ = filepath.Separator // silence unused import if filepath not otherwise used
}
