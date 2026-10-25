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
	"strings"

	"github.com/myshra777-ai/garuda/internal/diff"
	"github.com/myshra777-ai/garuda/internal/policy"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/talon"
	"github.com/spf13/cobra"
)

var (
	talonWorkspaceFlag   string
	talonPoliciesDirFlag string
	talonTargetFlag      string
	talonStagedFlag      bool
	talonDiffFlag        string
	talonFilesFlag       string
	talonHookFlag        bool
	talonFailOnBlockFlag bool
	talonJSONFlag        bool
	talonMaxRetriesFlag  int
	talonAutoIndexFlag   bool
)

var talonCmd = &cobra.Command{
	Use:   "talon",
	Short: "Garuda Talon — Active AI Agent Execution Harness & Enforcement Gate",
	Long: `Garuda Talon transforms Garuda from a passive graph into an active, fail-closed
execution harness for AI coding agents (Cursor, Claude Code, Codex, Devin, Aider, Windsurf).

Subcommands:
  init      Compile .garuda/TALON_CONTRACT.md and install active IDE & Git hooks
  contract  Print or regenerate the active workspace architectural & ADR contract
  brief     Generate deterministic pre-edit briefing for a file or symbol
  guard     Run incremental AST sync + diff-scoped policy & ADR check + self-repair output
  run       Supervisor loop: run an agent command and auto-loop self-repair until ALLOW`,
}

var talonInitCmd = &cobra.Command{
	Use:   "init [repo_dir]",
	Short: "Compile active ADR/policy contract and install Cursor, Claude Code, Codex, Windsurf & Git hooks",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoDir := "."
		if len(args) > 0 {
			repoDir = args[0]
		}
		absRepo, err := filepath.Abs(repoDir)
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
		wsName := talonWorkspaceFlag
		if wsName == "" {
			wsName = getWorkspaceName()
		}
		wsID, err := store.ResolveWorkspaceID(ctx, st.Pool(), tenantID, wsName)
		if err != nil {
			ws, cErr := st.CreateWorkspace(ctx, tenantID.String(), wsName, "Initialized via garuda talon init")
			if cErr == nil && ws != nil {
				wsID = ws.ID
			}
		}
		if wsName == "" {
			wsName = filepath.Base(absRepo)
		}

		policyDir := talonPoliciesDirFlag
		if policyDir == "" {
			policyDir = filepath.Join(absRepo, "policies")
		}

		contract, err := talon.CompileContract(ctx, st.Pool(), tenantID, wsID, wsName, policyDir)
		if err != nil {
			return fmt.Errorf("compile contract: %w", err)
		}

		selfPath, err := os.Executable()
		if err != nil {
			selfPath = filepath.Join(absRepo, "bin", "garuda")
		}
		binDir := filepath.Dir(selfPath)
		mcpPath := filepath.Join(binDir, "garuda-mcp")

		summary, err := talon.InstallHooks(absRepo, selfPath, mcpPath, dbURL, wsName, contract)
		if err != nil {
			return fmt.Errorf("install talon hooks: %w", err)
		}

		if talonJSONFlag {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(summary)
		}

		fmt.Println("🦅 Garuda Talon Active Enforcement Installed:")
		fmt.Printf("  • Contract:        %s\n", summary.ContractPath)
		fmt.Printf("  • Cursor Rule:     %s (alwaysApply: true)\n", summary.CursorRulePath)
		fmt.Printf("  • Cursor MCP:      %s (GARUDA_TALON_ENFORCE=true)\n", summary.CursorMCPPath)
		fmt.Printf("  • Claude Contract: %s\n", summary.ClaudeMDPath)
		fmt.Printf("  • Claude Hooks:    %s (PostToolUse -> garuda talon guard --hook)\n", summary.ClaudeSettings)
		fmt.Printf("  • Agents Contract: %s & %s\n", summary.AgentsMDPath, summary.WindsurfRulesPath)
		if summary.GitPreCommitPath != "" {
			fmt.Printf("  • Git Pre-Commit:  %s\n", summary.GitPreCommitPath)
		}
		if summary.GitCommitMsgPath != "" {
			fmt.Printf("  • Git Commit-Msg:  %s\n", summary.GitCommitMsgPath)
		}
		return nil
	},
}

var talonContractCmd = &cobra.Command{
	Use:   "contract [repo_dir]",
	Short: "Compile and print the active workspace ADR + policy architectural contract",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoDir := "."
		if len(args) > 0 {
			repoDir = args[0]
		}
		absRepo, _ := filepath.Abs(repoDir)

		ctx := context.Background()
		st, err := store.NewPostgresStore(getDBURL())
		if err != nil {
			return err
		}
		defer st.Close()

		tenantID := getTenantID()
		wsName := talonWorkspaceFlag
		if wsName == "" {
			wsName = getWorkspaceName()
		}
		wsID, err := store.ResolveWorkspaceID(ctx, st.Pool(), tenantID, wsName)
		if err != nil {
			return err
		}
		if wsName == "" {
			wsName = wsID.String()
		}

		policyDir := talonPoliciesDirFlag
		if policyDir == "" {
			policyDir = filepath.Join(absRepo, "policies")
		}

		contract, err := talon.CompileContract(ctx, st.Pool(), tenantID, wsID, wsName, policyDir)
		if err != nil {
			return err
		}

		if talonJSONFlag {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(contract)
		}
		fmt.Print(contract.Markdown)
		return nil
	},
}

var talonBriefCmd = &cobra.Command{
	Use:   "brief [target_symbol_or_file]",
	Short: "Emit deterministic pre-flight context, blast radius, and ADR rules before editing",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := talonTargetFlag
		if len(args) > 0 && target == "" {
			target = args[0]
		}

		ctx := context.Background()
		st, err := store.NewPostgresStore(getDBURL())
		if err != nil {
			return err
		}
		defer st.Close()

		tenantID := getTenantID()
		wsName := talonWorkspaceFlag
		if wsName == "" {
			wsName = getWorkspaceName()
		}
		wsID, err := store.ResolveWorkspaceID(ctx, st.Pool(), tenantID, wsName)
		if err != nil {
			return err
		}
		if wsName == "" {
			wsName = wsID.String()
		}

		brief, err := talon.BuildPreEditBriefing(ctx, st, tenantID, wsID, wsName, target, talonPoliciesDirFlag)
		if err != nil {
			return err
		}

		if talonJSONFlag {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(brief)
		}

		fmt.Printf("🦅 GARUDA TALON PRE-EDIT BRIEFING (workspace: %s, target: %s)\n", brief.Workspace, brief.Target)
		fmt.Printf("  • Transitive Blast Radius: %d entities (CRITICAL=%d, HIGH=%d)\n",
			brief.BlastRadiusTotal, brief.BySeverity["CRITICAL"], brief.BySeverity["HIGH"])
		for _, me := range brief.MatchedEntities {
			fmt.Printf("  • Symbol: %s.%s (%s, %d inbound edges) at %s\n",
				me.Package, me.Name, me.Kind, me.InboundEdges, me.FilePath)
		}
		for _, adr := range brief.ApplicableADRs {
			fmt.Printf("  • ADR Rule [%s]: %s %s %s %s (%s)\n",
				adr.Modality, adr.Subject, adr.Modality, adr.Predicate, adr.Object, adr.Document)
		}
		for _, pol := range brief.ActivePolicies {
			fmt.Printf("  • Policy [%s] %s (%s): %s\n",
				pol.Decision, pol.Title, pol.ID, pol.Reason)
		}
		return nil
	},
}

var talonGuardCmd = &cobra.Command{
	Use:   "guard [repo_dir]",
	Short: "Run incremental AST re-indexing + diff-scoped policy & ADR verification + self-repair output",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoDir := "."
		if len(args) > 0 {
			repoDir = args[0]
		}
		absRepo, err := filepath.Abs(repoDir)
		if err != nil {
			return err
		}

		ctx := context.Background()
		dbURL := getDBURL()
		_ = store.Migrate(dbURL, "migrations")
		st, err := store.NewPostgresStore(dbURL)
		if err != nil {
			return err
		}
		defer st.Close()

		tenantID := getTenantID()
		wsName := talonWorkspaceFlag
		if wsName == "" {
			wsName = getWorkspaceName()
		}
		wsID, err := store.ResolveWorkspaceID(ctx, st.Pool(), tenantID, wsName)
		if err != nil {
			return err
		}
		if wsName == "" {
			wsName = wsID.String()
		}

		var explicit []string
		if talonFilesFlag != "" {
			for _, f := range strings.Split(talonFilesFlag, ",") {
				if trimmed := strings.TrimSpace(f); trimmed != "" {
					explicit = append(explicit, trimmed)
				}
			}
		}

		// Incrementally sync working tree if dirty files exist
		if talonAutoIndexFlag {
			_, deltas, dErr := diff.CollectGitDeltas(absRepo, talonStagedFlag, talonDiffFlag, explicit)
			if dErr == nil && len(deltas) > 0 {
				workspaceFlag = wsName
				incrementalFlag = true
				saveFlag = true
				if !talonJSONFlag && !talonHookFlag {
					handleAnalyze(absRepo)
				}
				incrementalFlag = false
			}
		}

		verdict, err := talon.RunGuard(ctx, st, tenantID, wsID, talon.GuardOptions{
			RepoDir:       absRepo,
			WorkspaceName: wsName,
			PolicyDir:     talonPoliciesDirFlag,
			Staged:        talonStagedFlag,
			DiffRev:       talonDiffFlag,
			ExplicitFiles: explicit,
			NetNewOnly:    true,
			SaveLedger:    true,
			HookMode:      talonHookFlag,
		})
		if err != nil {
			return err
		}

		if talonJSONFlag {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(verdict)
		} else if talonHookFlag {
			if verdict.Outcome == string(policy.DecisionBlock) {
				fmt.Fprint(os.Stderr, verdict.SelfRepairPrompt)
				os.Exit(2)
			}
			fmt.Printf("✓ Garuda Talon Guard: %s (Net-New Violations: %d, Baseline: %d)\n",
				verdict.Outcome, verdict.NetNewViolationsCount, verdict.BaselineViolationsCount)
		} else {
			fmt.Println("╔═══════════════════════════════════════════════════════════════════╗")
			fmt.Println("║                  GARUDA TALON ACTIVE GUARD                        ║")
			fmt.Println("╚═══════════════════════════════════════════════════════════════════╝")
			fmt.Printf("  Workspace:   %s\n", verdict.Workspace)
			fmt.Printf("  Diff Scope:  %d file(s), %d touched symbol(s)\n",
				len(verdict.TouchedFiles), len(verdict.TouchedSymbols))
			fmt.Printf("  Blast Radius:%d affected entities (CRITICAL=%d, HIGH=%d)\n",
				verdict.BlastRadiusAffected, verdict.BySeverity["CRITICAL"], verdict.BySeverity["HIGH"])
			fmt.Printf("  Violations:  Net-New=%d  |  Baseline=%d  |  ADR Contradictions=%d\n",
				verdict.NetNewViolationsCount, verdict.BaselineViolationsCount, verdict.ADRContradictionsCount)
			if verdict.MerkleBlockHeight != nil {
				fmt.Printf("  Merkle Seal: RFC 6962 block #%d\n", *verdict.MerkleBlockHeight)
			}
			fmt.Printf("  OUTCOME:     %s\n", verdict.Outcome)
			if verdict.SelfRepairPrompt != "" {
				fmt.Println("\n" + verdict.SelfRepairPrompt)
			}
		}

		if (talonFailOnBlockFlag || talonHookFlag) && verdict.Outcome == string(policy.DecisionBlock) {
			os.Exit(2)
		}
		return nil
	},
}

var talonRunCmd = &cobra.Command{
	Use:   "run -- <agent-command> [args...]",
	Short: "Supervisor harness: run a coding agent command and auto-loop self-repair until ALLOW",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		absRepo, _ := filepath.Abs(".")
		ctx := context.Background()
		dbURL := getDBURL()
		_ = store.Migrate(dbURL, "migrations")
		st, err := store.NewPostgresStore(dbURL)
		if err != nil {
			return err
		}
		defer st.Close()

		tenantID := getTenantID()
		wsName := talonWorkspaceFlag
		if wsName == "" {
			wsName = getWorkspaceName()
		}
		wsID, err := store.ResolveWorkspaceID(ctx, st.Pool(), tenantID, wsName)
		if err != nil {
			return err
		}
		if wsName == "" {
			wsName = wsID.String()
		}

		reIndex := func(dir string) {
			workspaceFlag = wsName
			incrementalFlag = true
			saveFlag = true
			handleAnalyze(dir)
			incrementalFlag = false
		}

		verdict, attempts, err := talon.RunHarnessLoop(
			ctx,
			st,
			tenantID,
			wsID,
			talon.GuardOptions{
				RepoDir:       absRepo,
				WorkspaceName: wsName,
				PolicyDir:     talonPoliciesDirFlag,
				NetNewOnly:    true,
				SaveLedger:    true,
			},
			args,
			talonMaxRetriesFlag,
			reIndex,
		)
		if err != nil {
			return err
		}

		fmt.Printf("\n🦅 Garuda Talon Supervisor finished after %d attempt(s): OUTCOME = %s\n",
			attempts, verdict.Outcome)
		if verdict.Outcome == string(policy.DecisionBlock) {
			fmt.Fprintln(os.Stderr, verdict.SelfRepairPrompt)
			os.Exit(2)
		}
		return nil
	},
}

func init() {
	for _, sub := range []*cobra.Command{talonInitCmd, talonContractCmd, talonBriefCmd, talonGuardCmd, talonRunCmd} {
		sub.Flags().StringVarP(&talonWorkspaceFlag, "workspace", "w", "", "Workspace name")
		sub.Flags().StringVar(&talonPoliciesDirFlag, "policies", "./policies", "Directory containing policy YAML files")
		sub.Flags().BoolVar(&talonJSONFlag, "json", false, "Emit JSON output")
	}

	talonBriefCmd.Flags().StringVar(&talonTargetFlag, "target", "", "Target symbol or file path")

	talonGuardCmd.Flags().BoolVar(&talonStagedFlag, "staged", false, "Check staged git index changes")
	talonGuardCmd.Flags().StringVar(&talonDiffFlag, "diff", "", "Git revision or range to check")
	talonGuardCmd.Flags().StringVar(&talonFilesFlag, "files", "", "Comma-separated list of modified files")
	talonGuardCmd.Flags().BoolVar(&talonHookFlag, "hook", false, "Run in IDE/Claude hook mode (exits 2 with self-repair prompt on BLOCK)")
	talonGuardCmd.Flags().BoolVar(&talonFailOnBlockFlag, "fail-on-block", false, "Exit with code 2 when outcome is BLOCK")
	talonGuardCmd.Flags().BoolVar(&talonAutoIndexFlag, "auto-index", true, "Incrementally re-index dirty files before evaluating")

	talonRunCmd.Flags().IntVar(&talonMaxRetriesFlag, "max-retries", 3, "Maximum self-repair attempts before aborting")

	talonCmd.AddCommand(talonInitCmd, talonContractCmd, talonBriefCmd, talonGuardCmd, talonRunCmd)
	rootCmd.AddCommand(talonCmd)
}
