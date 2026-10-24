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
	"time"

	"github.com/myshra777-ai/garuda/internal/impact"
	"github.com/myshra777-ai/garuda/internal/policy"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/spf13/cobra"
)

var (
	checkSymbolFlag      string
	checkPoliciesDirFlag string
	checkWorkspaceFlag   string
	checkSaveFlag        bool
	checkFailOnBlockFlag bool
	checkJSONFlag        bool
)

// ChangeControlDecision represents the unified governance decision across
// software state, blast radius, runtime/intent evidence, and policy evaluation.
type ChangeControlDecision struct {
	Workspace         string            `json:"workspace"`
	TargetSymbol      string            `json:"target_symbol"`
	AffectedEntities  int               `json:"affected_entities"`
	BySeverity        map[string]int    `json:"by_severity"`
	EvidenceState     map[string]int64  `json:"evidence_state"`
	PoliciesEvaluated int               `json:"policies_evaluated"`
	BlockedBy         *string           `json:"blocked_by,omitempty"`
	GovernanceOutcome string            `json:"governance_outcome"`
	MerkleBlockHeight *int64            `json:"merkle_block_height,omitempty"`
	DurationMs        int64             `json:"duration_ms"`
}

var checkCmd = &cobra.Command{
	Use:   "check [dir]",
	Short: "Turn a proposed software change into an evidence-backed governance decision",
	Long: `Connects all five Garuda planes in a single change-control decision:
  01 ANALYZE    — Resolve software state and transitive blast radius
  02 CORRELATE  — Link intent, static claims, and runtime evidence (SUPPORTED / UNVERIFIED / CONTRADICTED)
  03 POLICY     — Evaluate declarative and transitive architectural policies
  04 GOVERN     — Determine final outcome (ALLOW, WARN, REVIEW, or BLOCK)
  05 VERIFY     — Anchor the decision into the RFC 6962 SHA-256 Merkle ledger`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return err
		}

		ctx := context.Background()
		dbURL := getDBURL()
		_ = store.Migrate(dbURL, "migrations")
		st, err := store.NewPostgresStore(dbURL)
		if err != nil {
			return fmt.Errorf("connect store: %w", err)
		}
		defer st.Close()

		tenantID := getTenantID()
		wsName := checkWorkspaceFlag
		if wsName == "" {
			wsName = os.Getenv("GARUDA_WORKSPACE")
		}
		wsID, err := store.ResolveWorkspaceID(ctx, st.Pool(), tenantID, wsName)
		if err != nil {
			return fmt.Errorf("resolve workspace %q: %w", wsName, err)
		}
		if wsName == "" {
			wsName = wsID.String()
		}

		// 1. Blast Radius for target symbol (if specified or auto-selected)
		affectedCount := 0
		sevMap := map[string]int{"CRITICAL": 0, "HIGH": 0, "MEDIUM": 0, "LOW": 0}
		targetDisplay := checkSymbolFlag
		if targetDisplay == "" {
			targetDisplay = filepath.Base(absDir)
		}

		if idx, idxErr := st.BuildImpactIndex(ctx, wsID); idxErr == nil && checkSymbolFlag != "" {
			var entityIDStr string
			_ = st.Pool().QueryRow(ctx, `
				SELECT id::text FROM entities
				WHERE workspace_id = $1 AND name = $2 AND kind != 'external'
				ORDER BY is_exported DESC LIMIT 1
			`, wsID, checkSymbolFlag).Scan(&entityIDStr)

			if entityIDStr != "" {
				br := impact.ComputeBlastRadius(idx, entityIDStr, impact.BlastRadiusConfig{
					MaxDepth:        5,
					MinConfidence:   0.50,
					IncludeInferred: true,
				})
				affectedCount = br.TotalAffected
				sevMap["CRITICAL"] = len(br.BySeverity[impact.SeverityCritical])
				sevMap["HIGH"] = len(br.BySeverity[impact.SeverityHigh])
				sevMap["MEDIUM"] = len(br.BySeverity[impact.SeverityMedium])
				sevMap["LOW"] = len(br.BySeverity[impact.SeverityLow])
			}
		} else {
			_ = st.Pool().QueryRow(ctx, `
				SELECT COUNT(*) FROM entities WHERE workspace_id = $1 AND kind != 'external'
			`, wsID).Scan(&affectedCount)
		}

		// 2. Correlate 3-State Evidence (SUPPORTED / UNVERIFIED / CONTRADICTED)
		evState := map[string]int64{"SUPPORTED": 0, "UNVERIFIED": 0, "CONTRADICTED": 0}
		rows, err := st.Pool().Query(ctx, `
			SELECT status, COUNT(*)
			FROM claim_verifications
			WHERE workspace_id = $1
			GROUP BY status
		`, wsID)
		if err == nil {
			for rows.Next() {
				var stName string
				var cnt int64
				if rows.Scan(&stName, &cnt) == nil {
					evState[stName] = cnt
				}
			}
			rows.Close()
		}

		// 3. Evaluate Policies & Seal Merkle Receipt
		policyDir := checkPoliciesDirFlag
		if policyDir == "" {
			policyDir = filepath.Join(absDir, "policies")
		}
		eng := policy.NewEngine(st.Pool())
		runRes, err := eng.Run(
			ctx,
			tenantID,
			wsID,
			policyDir,
			"workspace",
			wsName,
			"garuda-check-cli",
			policy.RunOptions{DryRun: !checkSaveFlag},
		)
		if err != nil {
			return fmt.Errorf("policy evaluation failed: %w", err)
		}

		outcome := string(runRes.FinalDecision)
		if outcome == "ALLOW" && (evState["CONTRADICTED"] > 0 || sevMap["CRITICAL"] > 5) {
			outcome = "REVIEW"
		}

		var latestBlock *int64
		for _, ev := range runRes.Evaluations {
			if ev.MerkleBlockHeight != nil {
				latestBlock = ev.MerkleBlockHeight
			}
		}

		var blockedByStr *string
		if runRes.BlockedBy != nil {
			s := runRes.BlockedBy.String()
			blockedByStr = &s
		}

		decision := ChangeControlDecision{
			Workspace:         wsName,
			TargetSymbol:      targetDisplay,
			AffectedEntities:  affectedCount,
			BySeverity:        sevMap,
			EvidenceState:     evState,
			PoliciesEvaluated: len(runRes.Evaluations),
			BlockedBy:         blockedByStr,
			GovernanceOutcome: outcome,
			MerkleBlockHeight: latestBlock,
			DurationMs:        time.Since(start).Milliseconds(),
		}

		if checkJSONFlag {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(decision)
		} else {
			fmt.Println()
			fmt.Println("╔═══════════════════════════════════════════════════════════════════╗")
			fmt.Println("║               GARUDA CHANGE-CONTROL DECISION                      ║")
			fmt.Println("╚═══════════════════════════════════════════════════════════════════╝")
			fmt.Printf("  Change:      %s (workspace: %s)\n", decision.TargetSymbol, decision.Workspace)
			fmt.Printf("  Affected:    %d entities (CRITICAL=%d, HIGH=%d)\n",
				decision.AffectedEntities, sevMap["CRITICAL"], sevMap["HIGH"])
			fmt.Printf("  Evidence:    SUPPORTED=%d  |  UNVERIFIED=%d  |  CONTRADICTED=%d\n",
				evState["SUPPORTED"], evState["UNVERIFIED"], evState["CONTRADICTED"])
			fmt.Printf("  Policies:    %d evaluated in %dms\n", decision.PoliciesEvaluated, decision.DurationMs)
			if decision.MerkleBlockHeight != nil {
				fmt.Printf("  Merkle Seal: RFC 6962 block #%d\n", *decision.MerkleBlockHeight)
			}
			fmt.Println("───")
			fmt.Printf("  GOVERNANCE OUTCOME: %s\n", decision.GovernanceOutcome)
			if decision.BlockedBy != nil {
				fmt.Printf("  Blocked By:         %s\n", *decision.BlockedBy)
			}
			fmt.Println()
		}

		if checkFailOnBlockFlag && decision.GovernanceOutcome == string(policy.DecisionBlock) {
			os.Exit(2)
		}
		return nil
	},
}

func init() {
	checkCmd.Flags().StringVar(&checkSymbolFlag, "symbol", "", "Target symbol or service name being modified")
	checkCmd.Flags().StringVar(&checkSymbolFlag, "target", "", "Alias for --symbol")
	checkCmd.Flags().StringVar(&checkPoliciesDirFlag, "policies", "./policies", "Directory containing policy YAML files")
	checkCmd.Flags().StringVar(&checkPoliciesDirFlag, "policy", "./policies", "Alias for --policies")
	checkCmd.Flags().StringVarP(&checkWorkspaceFlag, "workspace", "w", "", "Workspace name")
	checkCmd.Flags().StringVar(&checkWorkspaceFlag, "project", "", "Alias for --workspace")
	checkCmd.Flags().BoolVarP(&checkSaveFlag, "save", "s", true, "Persist evaluations and seal RFC 6962 Merkle receipt")
	checkCmd.Flags().BoolVar(&checkFailOnBlockFlag, "fail-on-block", false, "Exit with code 2 when outcome is BLOCK")
	checkCmd.Flags().BoolVar(&checkJSONFlag, "json", false, "Emit structured JSON decision card")
	rootCmd.AddCommand(checkCmd)
}
