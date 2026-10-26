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
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"

	"github.com/myshra777-ai/garuda/internal/policy"
	"github.com/myshra777-ai/garuda/internal/store"
)

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Deterministic policy enforcement: list, validate, evaluate, verify",
	Long: `Garuda's policy enforcement layer.

A policy is a declarative YAML file expressing intent (authority, scope, rules).
Every evaluation produces an enforcement decision (ALLOW / WARN / REVIEW / BLOCK)
with evidence and a BuildContext receipt (build tags, target platform, CGO state,
exact git commit SHA, graph fingerprint, and 1-command reproduction instruction),
anchored to the tenant's Merkle ledger.

Commands:
  list                          List active policies in the DB
  validate <dir>                Parse + validate all policies without evaluating
  evaluate <dir>                Run all policies against the workspace, persist results
  show <evaluation-id>          Show a specific evaluation with evidence & build receipt
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

var policyReconcileFlag bool

var policySaveFlag bool

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

		parsed, err := policy.ParseDirectory(dir)
		if err != nil {
			return fmt.Errorf("parse directory: %w", err)
		}

		if len(parsed) == 0 {
			fmt.Println("ℹ️  No .yaml or .yml policy files found.")
			return nil
		}

		fmt.Printf("Validated %d policies under %s\n\n", len(parsed), dir)
		for _, pp := range parsed {
			p := pp.Policy
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
			if pp.SourcePath != "" {
				fmt.Printf("    source:    %s\n", pp.SourcePath)
			}
			if pp.SourceHash != "" {
				fmt.Printf("    hash:      %s\n", pp.SourceHash)
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
	policyTagsFlag      string
	policyTargetFlag    string
	policyCGOFlag       bool
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
			wsName = getWorkspaceName()
		}

		workspaceID, err := store.ResolveWorkspaceID(context.Background(), pool, tenantID, wsName)
		if err != nil {
			return fmt.Errorf("resolve workspace %q: %w", wsName, err)
		}

		actor := policyActorFlag
		if actor == "" {
			actor = os.Getenv("GARUDA_AGENT")
		}
		if actor == "" {
			actor = "cli-operator"
		}
		if policyReconcileFlag && !policySaveFlag {
			return fmt.Errorf("--reconcile requires --save; both are write operations and must be explicit")
		}
		cgoChanged := cmd.Flags().Changed("cgo") || os.Getenv("CGO_ENABLED") != ""
		cgoVal := policyCGOFlag
		if !cmd.Flags().Changed("cgo") && os.Getenv("CGO_ENABLED") != "" {
			cgoVal = os.Getenv("CGO_ENABLED") == "1" || strings.EqualFold(os.Getenv("CGO_ENABLED"), "true")
		}
		targetSpec := buildTargetFromFlags(policyTargetFlag, policyTagsFlag, cgoChanged, cgoVal)

		ctx := context.Background()
		engine := policy.NewEngine(pool)

		start := time.Now()
		result, err := engine.Run(ctx, tenantID, workspaceID, dir, policySubjectKind, policySubjectID, actor,
			policy.RunOptions{
				DryRun:      !policySaveFlag,
				Reconcile:   policyReconcileFlag,
				BuildTarget: targetSpec,
			})
		if err != nil {
			return fmt.Errorf("engine run failed: %w", err)
		}
		elapsed := time.Since(start)

		if policyJSONFlag {
			out, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(out))
		} else {
			printPolicyRunResult(result, elapsed, wsName, !policySaveFlag)
		}

		if policyFailOnBlock && result.FinalDecision == policy.DecisionBlock {
			os.Exit(2)
		}
		return nil
	},
}

func printPolicyRunResult(r *policy.RunResult, elapsed time.Duration, wsName string, dryRun bool) {
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║               GARUDA POLICY ENFORCEMENT                            ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════╝")
	fmt.Printf("Workspace:   %s\n", wsName)
	if dryRun {
		fmt.Printf("Mode:        preview (nothing written)\n")
	} else {
		fmt.Printf("Mode:        committed\n")
	}
	if r.BuildContext != nil {
		fmt.Printf("Build Env:   %s\n", r.BuildContext.SummaryLine())
		fmt.Printf("Reproduce:   %s\n", r.BuildContext.ReproduceCommand)
	}
	if r.EnvironmentDrift != nil && r.EnvironmentDrift.DriftDetected {
		fmt.Printf("Env Drift:   ⚠️  %s\n", r.EnvironmentDrift.Summary)
	}
	fmt.Printf("Duration:    %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("Policies:    %d evaluated\n", len(r.Evaluations))
	fmt.Println()

	if len(r.Evaluations) == 0 {
		fmt.Println("No policies fired. All checks passed or none applied.")
		return
	}

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
	Short: "Show a specific policy evaluation with evidence and build-context receipt",
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

		ev, err := policy.LoadEvaluation(context.Background(), pool, getTenantID(), evalID)
		if err != nil {
			return fmt.Errorf("evaluation not found: %w", err)
		}

		fmt.Println()
		fmt.Println("POLICY EVALUATION RECEIPT")
		fmt.Println(strings.Repeat("═", 65))
		fmt.Printf("  ID:           %s\n", ev.ID)
		fmt.Printf("  Policy:       %s (v%s)\n", ev.PolicyID, ev.PolicyVersion)
		fmt.Printf("  Workspace:    %s\n", ev.WorkspaceID)
		fmt.Printf("  Subject:      %s/%s\n", ev.SubjectKind, ev.SubjectID)
		fmt.Printf("  Actor:        %s\n", ev.Actor)
		fmt.Printf("  Evaluated:    %s\n", ev.EvaluatedAt.Format(time.RFC3339))
		if ev.Evidence.BuildContext != nil {
			bc := ev.Evidence.BuildContext
			fmt.Println()
			fmt.Println("  Build & Graph Provenance Receipt:")
			fmt.Printf("    Language:       %s\n", bc.Language)
			fmt.Printf("    Target Platform:%s/%s (CGO=%t)\n", bc.TargetOS, bc.TargetArch, bc.CGOEnabled)
			fmt.Printf("    Build Tags:     %v\n", bc.BuildTags)
			fmt.Printf("    Exact Commit:   %s (dirty=%t)\n", bc.CommitSHA, bc.WorktreeDirty)
			fmt.Printf("    Graph SHA-256:  %s\n", bc.GraphFingerprint)
			fmt.Printf("    Receipt SHA-256:%s\n", bc.ReceiptHash)
			fmt.Printf("    Reproduce Cmd:  %s\n", bc.ReproduceCommand)
		}
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

		ev, err := policy.LoadEvaluation(context.Background(), pool, getTenantID(), evalID)
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
			if ev.Evidence.BuildContext != nil {
				fmt.Printf("   Build Receipt: %s\n", ev.Evidence.BuildContext.SummaryLine())
				fmt.Printf("   Reproduce:     %s\n", ev.Evidence.BuildContext.ReproduceCommand)
			}
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

// ─────────────────────────────────────────────────────────────────────────────
// Registration
// ─────────────────────────────────────────────────────────────────────────────

func init() {
	policyEvaluateCmd.Flags().StringVarP(&policyWorkspaceFlag, "workspace", "w", "", "Workspace name (default: GARUDA_WORKSPACE env)")
	policyEvaluateCmd.Flags().StringVar(&policySubjectKind, "subject-kind", "manual", "Subject kind (commit | pr | agent_action | manual)")
	policyEvaluateCmd.Flags().StringVar(&policySubjectID, "subject-id", "", "Subject identifier (commit SHA, PR number, etc.)")
	policyEvaluateCmd.Flags().StringVar(&policyActorFlag, "actor", "", "Actor (default: GARUDA_AGENT env)")
	policyEvaluateCmd.Flags().StringVar(&policyTagsFlag, "tags", "", "Comma-separated build tags recorded in the policy receipt")
	policyEvaluateCmd.Flags().StringVar(&policyTargetFlag, "target", "", "Target platform (os/arch) recorded in the policy receipt")
	policyEvaluateCmd.Flags().BoolVar(&policyCGOFlag, "cgo", false, "CGO state recorded in the policy receipt")
	policyEvaluateCmd.Flags().BoolVar(&policyJSONFlag, "json", false, "Output JSON")
	policyEvaluateCmd.Flags().BoolVar(&policyFailOnBlock, "fail-on-block", false, "Exit non-zero if any policy returns BLOCK")
	policyEvaluateCmd.Flags().BoolVar(&policyReconcileFlag, "reconcile", false,
		"Mark active policies not present in <dir> as superseded. Off by default; a test directory evaluated against a production workspace would otherwise supersede production policies. Requires --save.")
	policyEvaluateCmd.Flags().BoolVarP(&policySaveFlag, "save", "s", false,
		"Persist evaluations and anchor decisions to the Merkle ledger. Off by default; without --save, policy evaluate is a preview.")

	policyCmd.AddCommand(policyListCmd)
	policyCmd.AddCommand(policyValidateCmd)
	policyCmd.AddCommand(policyEvaluateCmd)
	policyCmd.AddCommand(policyShowCmd)
	policyCmd.AddCommand(policyVerifyCmd)

	rootCmd.AddCommand(policyCmd)
}
