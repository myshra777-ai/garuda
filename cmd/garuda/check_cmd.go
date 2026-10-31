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
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/diff"
	"github.com/myshra777-ai/garuda/internal/impact"
	"github.com/myshra777-ai/garuda/internal/knowledge"
	"github.com/myshra777-ai/garuda/internal/policy"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/talon"
	"github.com/spf13/cobra"
)

var (
	checkSymbolFlag        string
	checkPoliciesDirFlag   string
	checkWorkspaceFlag     string
	checkSaveFlag          bool
	checkFailOnBlockFlag   bool
	checkJSONFlag          bool
	checkDiffFlag          string
	checkStagedFlag        bool
	checkFilesFlag         string
	checkNetNewOnlyFlag    bool
	checkAutoIndexFlag     bool
	checkTagsFlag          string
	checkPlatformFlag      string
	checkCGOFlag           bool
	checkCommitFlag        string
	checkMatrixFlag        bool
	checkWorktreeDirtyFlag bool
	checkDirtyTreeHashFlag string
)

// ChangeControlDecision represents the unified governance decision across
// software state, blast radius, runtime/intent evidence, build-context receipt,
// and policy evaluation.
type ChangeControlDecision struct {
	Workspace              string                           `json:"workspace"`
	TargetSymbol           string                           `json:"target_symbol"`
	DiffMode               string                           `json:"diff_mode,omitempty"`
	TouchedFiles           []string                         `json:"touched_files,omitempty"`
	TouchedSymbols         []string                         `json:"touched_symbols,omitempty"`
	TouchedPackages        []string                         `json:"touched_packages,omitempty"`
	AffectedEntities       int                              `json:"affected_entities"`
	BySeverity             map[string]int                   `json:"by_severity"`
	EvidenceState          map[string]int64                 `json:"evidence_state"`
	ADRContradictions      int                              `json:"adr_contradictions"`
	PoliciesEvaluated      int                              `json:"policies_evaluated"`
	NetNewViolations       int                              `json:"net_new_violations"`
	BaselineViolations     int                              `json:"baseline_violations"`
	Violations             []string                         `json:"violations,omitempty"`
	SelfRepairInstructions []string                         `json:"self_repair_instructions,omitempty"`
	BlockedBy              *string                          `json:"blocked_by,omitempty"`
	BlockedByPolicy        string                           `json:"blocked_by_policy,omitempty"`
	GovernanceOutcome      string                           `json:"governance_outcome"`
	MerkleBlockHeight      *int64                           `json:"merkle_block_height,omitempty"`
	BuildContext           *analyzer.BuildContext           `json:"build_context,omitempty"`
	EnvironmentDrift       *analyzer.EnvironmentDriftReport `json:"environment_drift,omitempty"`
	MatrixResults          []talon.MatrixTargetVerdict      `json:"matrix_results,omitempty"`
	ReproduceCommand       string                           `json:"reproduce_command,omitempty"`
	DurationMs             int64                            `json:"duration_ms"`
}

var checkCmd = &cobra.Command{
	Use:   "check [dir]",
	Short: "Turn a proposed software change or git diff into an evidence-backed, build-reproducible governance decision",
	Long: `Connects all five Garuda planes in a single change-control decision with deterministic BuildContext provenance:
  01 ANALYZE    — Resolve software state, build constraints (GOOS/GOARCH/CGO/BuildTags), git-diff touched symbols, and blast radius
  02 CORRELATE  — Link ADR intent (document_claims), static claims, and runtime evidence
  03 POLICY     — Evaluate declarative and transitive policies (with net-new vs baseline filtering)
  04 GOVERN     — Determine final outcome (ALLOW, WARN, REVIEW, or BLOCK) + CI-vs-Editor environment drift + self-repair plan
  05 VERIFY     — Anchor the decision + BuildContext receipt into the RFC 6962 SHA-256 Merkle ledger`,
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

		// Handle --target when passed as os/arch (e.g. --target linux/amd64) vs symbol name
		platformStr := checkPlatformFlag
		symbolStr := checkSymbolFlag
		if platformStr == "" && strings.Contains(symbolStr, "/") &&
			(strings.HasPrefix(symbolStr, "linux/") || strings.HasPrefix(symbolStr, "darwin/") || strings.HasPrefix(symbolStr, "windows/")) {
			platformStr = symbolStr
			symbolStr = ""
		}

		cgoChanged := cmd.Flags().Changed("cgo") || os.Getenv("CGO_ENABLED") != ""
		cgoVal := checkCGOFlag
		if !cmd.Flags().Changed("cgo") && os.Getenv("CGO_ENABLED") != "" {
			cgoVal = os.Getenv("CGO_ENABLED") == "1" || strings.EqualFold(os.Getenv("CGO_ENABLED"), "true")
		}
		targetSpec := buildTargetFromFlags(platformStr, checkTagsFlag, cgoChanged, cgoVal)

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

		// Optional git-diff / staged / explicit files targeting
		isDiffMode := checkStagedFlag || checkDiffFlag != "" || checkFilesFlag != ""
		var explicitFiles []string
		if checkFilesFlag != "" {
			for _, f := range strings.Split(checkFilesFlag, ",") {
				if trimmed := strings.TrimSpace(f); trimmed != "" {
					explicitFiles = append(explicitFiles, trimmed)
				}
			}
		}

		var diffRes *diff.TargetResolution
		if isDiffMode {
			mode, deltas, dErr := diff.CollectGitDeltas(absDir, checkStagedFlag, checkDiffFlag, explicitFiles)
			if dErr == nil && diff.HasSourceFileDeltas(deltas) && checkAutoIndexFlag {
				workspaceFlag = wsName
				incrementalFlag = true
				saveFlag = true
				analyzeTargetFlag = platformStr
				analyzeTagsFlag = checkTagsFlag
				analyzeCGOFlag = cgoVal
				if !checkJSONFlag {
					handleAnalyze(absDir)
				}
				incrementalFlag = false
			}
			if dErr == nil {
				diffRes, _ = diff.ResolveTouchedEntities(ctx, st.Pool(), wsID, absDir, mode, checkDiffFlag, deltas)
			}
		}

		// 1. Blast Radius for target symbol(s) or union of diff-touched symbols
		affectedCount := 0
		sevMap := map[string]int{"CRITICAL": 0, "HIGH": 0, "MEDIUM": 0, "LOW": 0}
		targetDisplay := symbolStr
		if targetDisplay == "" {
			if diffRes != nil && len(diffRes.TouchedEntities) > 0 {
				names := make([]string, 0, min(4, len(diffRes.TouchedEntities)))
				for i, te := range diffRes.TouchedEntities {
					if i >= 4 {
						break
					}
					names = append(names, te.Name)
				}
				targetDisplay = fmt.Sprintf("diff:%s (%d symbols: %s)", diffRes.Mode, len(diffRes.TouchedEntities), strings.Join(names, ", "))
			} else {
				targetDisplay = filepath.Base(absDir)
			}
		}

		idx, idxErr := st.BuildImpactIndex(ctx, wsID)
		if idxErr == nil && diffRes != nil && len(diffRes.TouchedEntities) > 0 {
			seenAffected := make(map[string]impact.SeverityLevel)
			for _, te := range diffRes.TouchedEntities {
				br := impact.ComputeBlastRadius(idx, te.ID.String(), impact.BlastRadiusConfig{
					MaxDepth:        5,
					MinConfidence:   0.50,
					IncludeInferred: true,
				})
				for sev, list := range br.BySeverity {
					for _, item := range list {
						if prev, exists := seenAffected[item.EntityID]; !exists || sevRank(sev) > sevRank(prev) {
							seenAffected[item.EntityID] = sev
						}
					}
				}
			}
			affectedCount = len(seenAffected)
			for _, sev := range seenAffected {
				sevMap[string(sev)]++
			}
		} else if idxErr == nil && symbolStr != "" {
			var entityIDStr string
			_ = st.Pool().QueryRow(ctx, `
				SELECT id::text FROM entities
				WHERE workspace_id = $1 AND name = $2 AND kind != 'external'
				ORDER BY is_exported DESC LIMIT 1
			`, wsID, symbolStr).Scan(&entityIDStr)

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

		// 2. Correlate 3-State Evidence (SUPPORTED / UNVERIFIED / CONTRADICTED) + ADR claims
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

		// Verify ADR document_claims against current AST
		adrContradictions := 0
		var adrViolations []string
		ke := knowledge.NewEvaluator(st.Pool())
		if kStats, _, kErr := ke.EvaluateWorkspace(ctx, tenantID, wsID); kErr == nil && kStats != nil {
			adrContradictions = kStats.Contradicted
			evState["CONTRADICTED"] += int64(kStats.Contradicted)
			evState["SUPPORTED"] += int64(kStats.Supported)
		}
		if adrContradictions > 0 {
			cRows, cErr := st.Pool().Query(ctx, `
				SELECT COALESCE(document_title, 'ADR'), subject, modality, predicate, object, COALESCE(contradiction_reason, '')
				FROM document_claims
				WHERE tenant_id = $1 AND workspace_id = $2 AND status = 'CONTRADICTED'
			`, tenantID, wsID)
			if cErr == nil {
				for cRows.Next() {
					var doc, subj, mod, pred, obj, rsn string
					if cRows.Scan(&doc, &subj, &mod, &pred, &obj, &rsn) == nil {
						adrViolations = append(adrViolations,
							fmt.Sprintf("[ADR:%s] %s %s %s %s — %s", doc, subj, mod, pred, obj, rsn))
					}
				}
				cRows.Close()
			}
		}

		// 3. Build DiffScopeFilter if diff mode or net-new-only is requested
		var scopeFilter *policy.DiffScopeFilter
		if checkNetNewOnlyFlag || isDiffMode {
			entMap := make(map[uuid.UUID]bool)
			pkgMap := make(map[string]bool)
			if diffRes != nil {
				for _, te := range diffRes.TouchedEntities {
					entMap[te.ID] = true
				}
				for _, p := range diffRes.TouchedPackages {
					pkgMap[p] = true
				}
			}
			scopeFilter = &policy.DiffScopeFilter{
				TouchedEntityIDs: entMap,
				TouchedPackages:  pkgMap,
				OnlyNetNew:       true,
			}
		}

		// 4. Evaluate Policies & Seal Merkle Receipt (with BuildContext)
		activeBuildCtx := analyzer.ResolveBuildContext(absDir, "", wsName, targetSpec)
		if checkCommitFlag != "" {
			activeBuildCtx.CommitSHA = checkCommitFlag
			activeBuildCtx.CommitShort = checkCommitFlag[:min(12, len(checkCommitFlag))]
			activeBuildCtx.ReproduceCommand = activeBuildCtx.FormatReproduceCommand(wsName)
			activeBuildCtx.ReceiptHash = activeBuildCtx.ComputeReceiptHash()
		}
		if checkWorktreeDirtyFlag {
			activeBuildCtx.WorktreeDirty = true
			activeBuildCtx.ReproduceCommand = activeBuildCtx.FormatReproduceCommand(wsName)
			activeBuildCtx.ReceiptHash = activeBuildCtx.ComputeReceiptHash()
		}
		if checkDirtyTreeHashFlag != "" {
			activeBuildCtx.DirtyFilesHash = checkDirtyTreeHashFlag
			activeBuildCtx.ReproduceCommand = activeBuildCtx.FormatReproduceCommand(wsName)
			activeBuildCtx.ReceiptHash = activeBuildCtx.ComputeReceiptHash()
		}

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
			policy.RunOptions{
				DryRun:       !checkSaveFlag,
				DiffScope:    scopeFilter,
				RepoDir:      absDir,
				BuildTarget:  targetSpec,
				BuildContext: activeBuildCtx,
			},
		)
		if err != nil {
			return fmt.Errorf("policy evaluation failed: %w", err)
		}

		outcome := string(runRes.FinalDecision)
		if adrContradictions > 0 {
			outcome = string(policy.DecisionBlock)
		} else if outcome == "ALLOW" && (evState["CONTRADICTED"] > 0 || sevMap["CRITICAL"] > 5) {
			outcome = "REVIEW"
		}

		var latestBlock *int64
		var violations []string
		var repairs []string
		violations = append(violations, adrViolations...)
		for _, v := range adrViolations {
			repairs = append(repairs, "Refactor code to satisfy ADR rule: "+v)
		}

		for _, ev := range runRes.Evaluations {
			if ev.MerkleBlockHeight != nil {
				latestBlock = ev.MerkleBlockHeight
			}
			if ev.Decision != policy.DecisionAllow {
				for _, note := range ev.Evidence.ReasoningNotes {
					violations = append(violations, fmt.Sprintf("[%s:%s] %s", ev.PolicyExternalID, ev.Decision, note))
				}
				repairs = append(repairs, fmt.Sprintf("Policy %s (%s): %s", ev.PolicyExternalID, ev.PolicyTitle, ev.Reason))
			}
		}

		var blockedByStr *string
		if runRes.BlockedBy != nil {
			s := runRes.BlockedBy.String()
			blockedByStr = &s
		}

		var matrixResults []talon.MatrixTargetVerdict
		if checkMatrixFlag {
			matrixResults, _ = talon.RunMatrixVerification(ctx, absDir, wsName, nil)
		}

		decision := ChangeControlDecision{
			Workspace:              wsName,
			TargetSymbol:           targetDisplay,
			AffectedEntities:       affectedCount,
			BySeverity:             sevMap,
			EvidenceState:          evState,
			ADRContradictions:      adrContradictions,
			PoliciesEvaluated:      len(runRes.Evaluations),
			NetNewViolations:       runRes.NetNewViolationsCount + adrContradictions,
			BaselineViolations:     runRes.BaselineViolationsCount,
			Violations:             violations,
			SelfRepairInstructions: repairs,
			BlockedBy:              blockedByStr,
			BlockedByPolicy:        runRes.BlockedByPolicyID,
			GovernanceOutcome:      outcome,
			MerkleBlockHeight:      latestBlock,
			BuildContext:           runRes.BuildContext,
			EnvironmentDrift:       runRes.EnvironmentDrift,
			MatrixResults:          matrixResults,
			DurationMs:             time.Since(start).Milliseconds(),
		}
		if runRes.BuildContext != nil {
			decision.ReproduceCommand = runRes.BuildContext.ReproduceCommand
		}
		if diffRes != nil {
			decision.DiffMode = diffRes.Mode
			decision.TouchedFiles = diffRes.TouchedFiles
			decision.TouchedPackages = diffRes.TouchedPackages
			for _, te := range diffRes.TouchedEntities {
				decision.TouchedSymbols = append(decision.TouchedSymbols, fmt.Sprintf("%s.%s (%s:%d)", te.Package, te.Name, filepath.Base(te.FilePath), te.LineStart))
			}
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
			if decision.BuildContext != nil {
				bc := decision.BuildContext
				fmt.Printf("  Build Env:   %s\n", bc.SummaryLine())
				if bc.ReceiptHash != "" {
					fmt.Printf("  Receipt SHA: %s\n", bc.ReceiptHash[:min(16, len(bc.ReceiptHash))])
				}
				fmt.Printf("  Reproduce:   %s\n", bc.ReproduceCommand)
			}
			if decision.EnvironmentDrift != nil && decision.EnvironmentDrift.DriftDetected {
				fmt.Printf("  Env Drift:   ⚠️  %s\n", decision.EnvironmentDrift.Summary)
				for _, d := range decision.EnvironmentDrift.Discrepancies {
					fmt.Printf("               • %s\n", d)
				}
			}
			if len(decision.TouchedFiles) > 0 {
				fmt.Printf("  Diff Scope:  %d file(s), %d touched symbol(s)\n",
					len(decision.TouchedFiles), len(decision.TouchedSymbols))
			}
			fmt.Printf("  Affected:    %d entities (CRITICAL=%d, HIGH=%d)\n",
				decision.AffectedEntities, sevMap["CRITICAL"], sevMap["HIGH"])
			fmt.Printf("  Evidence:    SUPPORTED=%d  |  UNVERIFIED=%d  |  CONTRADICTED=%d\n",
				evState["SUPPORTED"], evState["UNVERIFIED"], evState["CONTRADICTED"])
			fmt.Printf("  Policies:    %d evaluated in %dms (Net-New=%d, Baseline=%d)\n",
				decision.PoliciesEvaluated, decision.DurationMs, decision.NetNewViolations, decision.BaselineViolations)
			if decision.MerkleBlockHeight != nil {
				fmt.Printf("  Merkle Seal: RFC 6962 block #%d\n", *decision.MerkleBlockHeight)
			}
			if len(decision.MatrixResults) > 0 {
				fmt.Println("  Build Matrix:")
				for _, mr := range decision.MatrixResults {
					fpShort := mr.GraphFingerprint
					if len(fpShort) > 10 {
						fpShort = fpShort[:10]
					}
					fmt.Printf("    • %-14s cgo=%-5t tags=%-8v files=%d (skipped=%d) edges=%d fp=%s [%s]\n",
						mr.TargetPlatform, mr.CGOEnabled, mr.BuildTags, mr.ActiveFilesCount, mr.SkippedByBuildTagsCount, mr.EdgesCount, fpShort, mr.Outcome)
				}
			}
			fmt.Println("───")
			fmt.Printf("  GOVERNANCE OUTCOME: %s\n", decision.GovernanceOutcome)
			if decision.BlockedBy != nil {
				fmt.Printf("  Blocked By:         %s (%s)\n", *decision.BlockedBy, decision.BlockedByPolicy)
			}
			for _, v := range decision.Violations {
				fmt.Printf("  ✗ Violation:        %s\n", v)
			}
			for _, r := range decision.SelfRepairInstructions {
				fmt.Printf("  → Self-Repair:      %s\n", r)
			}
			fmt.Println()
		}

		if checkFailOnBlockFlag && decision.GovernanceOutcome == string(policy.DecisionBlock) {
			os.Exit(2)
		}
		return nil
	},
}

func sevRank(s impact.SeverityLevel) int {
	switch s {
	case impact.SeverityCritical:
		return 4
	case impact.SeverityHigh:
		return 3
	case impact.SeverityMedium:
		return 2
	case impact.SeverityLow:
		return 1
	default:
		return 0
	}
}

func init() {
	checkCmd.Flags().StringVar(&checkSymbolFlag, "symbol", "", "Target symbol or service name being modified")
	checkCmd.Flags().StringVar(&checkSymbolFlag, "target", "", "Target symbol OR target platform (e.g. linux/amd64)")
	checkCmd.Flags().StringVar(&checkPlatformFlag, "platform", "", "Target platform in os/arch format (e.g. linux/amd64, darwin/arm64, windows/amd64)")
	checkCmd.Flags().StringVar(&checkTagsFlag, "tags", "", "Comma-separated build tags / environment markers (e.g. integration,purego)")
	checkCmd.Flags().BoolVar(&checkCGOFlag, "cgo", false, "Enable CGO when evaluating Go //go:build cgo constraints")
	checkCmd.Flags().StringVar(&checkCommitFlag, "commit", "", "Exact git commit SHA for build-context receipt verification")
	checkCmd.Flags().BoolVar(&checkWorktreeDirtyFlag, "worktree-dirty", false, "Assert working tree dirty state when checking build-context receipt")
	checkCmd.Flags().StringVar(&checkDirtyTreeHashFlag, "dirty-tree-hash", "", "Dirty working tree files hash for receipt verification")
	checkCmd.Flags().BoolVar(&checkMatrixFlag, "matrix", false, "Verify graph across cross-platform build matrix (linux/amd64, darwin/arm64, windows/amd64, cgo/purego)")
	checkCmd.Flags().StringVar(&checkDiffFlag, "diff", "", "Git revision or range to target (e.g. HEAD~1..HEAD or main...HEAD)")
	checkCmd.Flags().BoolVar(&checkStagedFlag, "staged", false, "Target staged git index changes (git diff --cached)")
	checkCmd.Flags().StringVar(&checkFilesFlag, "files", "", "Comma-separated list of modified files to target")
	checkCmd.Flags().BoolVar(&checkNetNewOnlyFlag, "net-new-only", false, "Only block on net-new violations touching the active diff")
	checkCmd.Flags().BoolVar(&checkAutoIndexFlag, "auto-index", true, "Automatically run incremental AST sync on dirty files before checking")
	checkCmd.Flags().StringVar(&checkPoliciesDirFlag, "policies", "./policies", "Directory containing policy YAML files")
	checkCmd.Flags().StringVar(&checkPoliciesDirFlag, "policy", "./policies", "Alias for --policies")
	checkCmd.Flags().StringVarP(&checkWorkspaceFlag, "workspace", "w", "", "Workspace name")
	checkCmd.Flags().StringVar(&checkWorkspaceFlag, "project", "", "Alias for --workspace")
	checkCmd.Flags().BoolVarP(&checkSaveFlag, "save", "s", true, "Persist evaluations and seal RFC 6962 Merkle receipt")
	checkCmd.Flags().BoolVar(&checkFailOnBlockFlag, "fail-on-block", false, "Exit with code 2 when outcome is BLOCK")
	checkCmd.Flags().BoolVar(&checkJSONFlag, "json", false, "Emit structured JSON decision card")
	rootCmd.AddCommand(checkCmd)
}
