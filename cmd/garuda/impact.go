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
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/impact"
	"github.com/myshra777-ai/garuda/internal/store"
)

// -----------------------------------------------------------------------------
// IMPACT COMMAND (Workspace-based blast radius)
// -----------------------------------------------------------------------------

var (
	flagWorkspaceID string
	flagTargetID    string
	flagMaxDepth    int
	flagMinConf     float64
	flagJSON        bool
	flagOutput      string
)

var impactCmd = &cobra.Command{
	Use:   "impact",
	Short: "Analyze blast radius of a specific entity in a workspace",
	Long: `Computes the blast radius of changes to an entity.
It performs a BFS traversal of the dependency graph to find all consumers.
Evidence-backed and confidence-weighted.

Examples:
  garuda impact --workspace <uuid> --target <entity-id>
  garuda impact --workspace <uuid> --target <entity-id> --json
  garuda impact --workspace <uuid> --target <entity-id> --depth 5 --min-confidence 0.7`,
	Run: func(cmd *cobra.Command, args []string) {
		handleImpact()
	},
}

func init() {
	impactCmd.Flags().StringVar(&flagWorkspaceID, "workspace", "", "Workspace UUID (Required)")
	impactCmd.Flags().StringVar(&flagTargetID, "target", "", "Target Entity ID to trace (Required)")
	impactCmd.Flags().IntVar(&flagMaxDepth, "depth", 3, "Maximum traversal depth")
	impactCmd.Flags().Float64Var(&flagMinConf, "min-confidence", 0.50, "Minimum edge confidence threshold")
	impactCmd.Flags().BoolVar(&flagJSON, "json", false, "Output machine-readable JSON")
	impactCmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Write report to file")
	_ = impactCmd.MarkFlagRequired("workspace")
	_ = impactCmd.MarkFlagRequired("target")
}

func handleImpact() {
	ctx := context.Background()
	dbURL := getDBURL()

	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect to DB: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	workspaceID, err := uuid.Parse(flagWorkspaceID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Invalid workspace UUID: %v\n", err)
		os.Exit(1)
	}

	// Build impact index (lazy in-memory projection)
	idx, err := st.BuildImpactIndex(ctx, workspaceID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to build impact index: %v\n", err)
		os.Exit(1)
	}

	// Run BFS
	cfg := impact.BlastRadiusConfig{
		MaxDepth:        flagMaxDepth,
		MinConfidence:   flagMinConf,
		IncludeInferred: true,
	}

	result := impact.ComputeBlastRadius(idx, flagTargetID, cfg)

	if flagJSON {
		report := struct {
			Version       string                  `json:"version"`
			WorkspaceID   string                  `json:"workspace_id"`
			TargetEntity  string                  `json:"target_entity"`
			AnalyzedAt    time.Time               `json:"analyzed_at"`
			TotalAffected int                     `json:"total_affected"`
			BlastRadius   []impact.ImpactedEntity `json:"blast_radius"`
		}{
			Version:       "1.0",
			WorkspaceID:   flagWorkspaceID,
			TargetEntity:  flagTargetID,
			AnalyzedAt:    time.Now().UTC(),
			TotalAffected: result.TotalAffected,
			BlastRadius:   []impact.ImpactedEntity{},
		}

		// Flatten all impacted entities
		for _, entities := range result.BySeverity {
			report.BlastRadius = append(report.BlastRadius, entities...)
		}

		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to marshal JSON: %v\n", err)
			os.Exit(1)
		}

		if flagOutput != "" {
			if err := os.WriteFile(flagOutput, data, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to write output: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("📄 JSON impact report written to %s\n", flagOutput)
		} else {
			fmt.Println(string(data))
		}
		return
	}

	// Human-readable output
	printImpactReport(result)
}

func printImpactReport(result *impact.BlastRadiusResult) {
	fmt.Printf("\n🔍 GARUDA IMPACT BLAST RADIUS REPORT\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Target:      %s\n", flagTargetID[:8]+"...")
	fmt.Printf("Workspace:   %s\n", flagWorkspaceID)
	fmt.Printf("Total affected: %d\n", result.TotalAffected)

	if result.TotalAffected == 0 {
		fmt.Println("\n✅ No consumers found. Entity appears safe to change.")
		return
	}

	severityCounts := map[impact.SeverityLevel]int{
		impact.SeverityCritical: 0,
		impact.SeverityHigh:     0,
		impact.SeverityMedium:   0,
		impact.SeverityLow:      0,
	}
	for sev, list := range result.BySeverity {
		severityCounts[sev] = len(list)
	}

	fmt.Printf("  Critical:  %d\n", severityCounts[impact.SeverityCritical])
	fmt.Printf("  High:      %d\n", severityCounts[impact.SeverityHigh])
	fmt.Printf("  Medium:    %d\n", severityCounts[impact.SeverityMedium])
	fmt.Printf("  Low:       %d\n", severityCounts[impact.SeverityLow])
	fmt.Println()

	fmt.Println("📋 Impact details:")

	for _, sev := range []impact.SeverityLevel{
		impact.SeverityCritical,
		impact.SeverityHigh,
		impact.SeverityMedium,
		impact.SeverityLow,
	} {
		entities := result.BySeverity[sev]
		if len(entities) == 0 {
			continue
		}
		entities = impact.SortImpactedEntities(entities)

		fmt.Printf("\n  [%s] %s (%d):\n", severityEmoji(sev), string(sev), len(entities))
		for _, item := range entities {
			fmt.Printf("    • %s.%s (%s) — Depth: %d (Conf: %.2f)\n",
				item.Package, item.Name, item.Kind, item.Depth, item.CumulativeConf)
			if len(item.TraversalChain) > 1 {
				fmt.Printf("      Chain: %s\n", strings.Join(item.TraversalChain, " → "))
			}
			if item.Evidence.FilePath != "" {
				fmt.Printf("      Evidence: %s:%d-%d\n",
					item.Evidence.FilePath, item.Evidence.LineStart, item.Evidence.LineEnd)
			}
		}
	}

	fmt.Println()
	if severityCounts[impact.SeverityCritical] > 0 {
		fmt.Println("🚫 Critical changes detected. Review before merging.")
	} else if severityCounts[impact.SeverityHigh] > 0 {
		fmt.Println("⚠️ High-impact changes detected. Consider additional review.")
	} else if severityCounts[impact.SeverityMedium] > 0 {
		fmt.Println("🟡 Medium-impact changes. Review recommended.")
	} else {
		fmt.Println("🟢 Low-impact changes. Safe to proceed.")
	}
}

func severityEmoji(sev impact.SeverityLevel) string {
	switch sev {
	case impact.SeverityCritical:
		return "🚫"
	case impact.SeverityHigh:
		return "🔴"
	case impact.SeverityMedium:
		return "🟡"
	case impact.SeverityLow:
		return "🟢"
	default:
		return "⚪"
	}
}

// -----------------------------------------------------------------------------
// IMPACT-DIFF COMMAND (Compare two snapshots)
// -----------------------------------------------------------------------------

var (
	impactDiffOutput     string
	impactDiffJSON       bool
	impactDiffDepth      int
	impactDiffConfidence float64
)

var impactDiffCmd = &cobra.Command{
	Use:   "impact-diff <baseline.json> <proposed.json>",
	Short: "Analyze impact between two snapshots",
	Long: `Compares two snapshots and identifies breaking changes and affected consumers.

Examples:
  garuda impact-diff v1.json v2.json
  garuda impact-diff v1.json v2.json --json`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		handleImpactDiff(args[0], args[1])
	},
}

func init() {
	impactDiffCmd.Flags().StringVarP(&impactDiffOutput, "output", "o", "", "Write report to file")
	impactDiffCmd.Flags().BoolVar(&impactDiffJSON, "json", false, "Output structured JSON for CI")
	impactDiffCmd.Flags().IntVar(&impactDiffDepth, "depth", 3, "Max graph traversal depth")
	impactDiffCmd.Flags().Float64Var(&impactDiffConfidence, "min-confidence", 0.50, "Min confidence threshold")
}

func handleImpactDiff(baselineFile, proposedFile string) {
	baseline, err := analyzer.LoadResult(baselineFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to load baseline snapshot: %v\n", err)
		os.Exit(1)
	}
	proposed, err := analyzer.LoadResult(proposedFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to load proposed snapshot: %v\n", err)
		os.Exit(1)
	}

	dbURL := getDBURL()
	tenantIDStr := getTenantIDString()
	ctx := context.Background()

	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Store connection failed: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	tenantUUID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Invalid tenant ID: %v\n", err)
		os.Exit(1)
	}

	workspaceName := os.Getenv("GARUDA_WORKSPACE")
	if workspaceName == "" {
		workspaceName = "default"
	}

	var workspaceID uuid.UUID
	err = st.Pool().QueryRow(ctx, `
		SELECT id FROM workspaces WHERE tenant_id = $1 AND name = $2
	`, tenantUUID, workspaceName).Scan(&workspaceID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Workspace '%s' not found. Run 'garuda workspace create %s' first.\n", workspaceName, workspaceName)
		os.Exit(1)
	}

	diff := analyzer.Diff(baseline, proposed)

	impactIndex, err := st.BuildImpactIndex(ctx, workspaceID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed building in-memory impact index: %v\n", err)
		os.Exit(1)
	}

	cfg := impact.BlastRadiusConfig{
		MaxDepth:        impactDiffDepth,
		MinConfidence:   impactDiffConfidence,
		IncludeInferred: true,
	}

	allResults := make([]*impact.BlastRadiusResult, 0)

	for _, ed := range diff.EntityDiffs {
		var entityID uuid.UUID
		err := st.Pool().QueryRow(ctx, `
			SELECT id FROM entities 
			WHERE workspace_id = $1 AND name = $2 AND package = $3
			LIMIT 1
		`, workspaceID, ed.Name, extractPackageFromEntity(ed.Name, ed.EntityID)).Scan(&entityID)
		if err != nil {
			continue
		}

		res := impact.ComputeBlastRadius(impactIndex, entityID.String(), cfg)
		if res.TotalAffected > 0 {
			allResults = append(allResults, res)
		}
	}

	if impactDiffJSON {
		output := struct {
			Workspace     string                      `json:"workspace"`
			AnalyzedAt    time.Time                   `json:"analyzed_at"`
			Baseline      string                      `json:"baseline"`
			Proposed      string                      `json:"proposed"`
			TotalAffected int                         `json:"total_affected"`
			Impacted      []*impact.BlastRadiusResult `json:"impacted"`
		}{
			Workspace:     workspaceName,
			AnalyzedAt:    time.Now().UTC(),
			Baseline:      baselineFile,
			Proposed:      proposedFile,
			TotalAffected: 0,
			Impacted:      allResults,
		}

		for _, r := range allResults {
			output.TotalAffected += r.TotalAffected
		}

		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed marshaling impact JSON: %v\n", err)
			os.Exit(1)
		}

		if impactDiffOutput != "" {
			if err := os.WriteFile(impactDiffOutput, data, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to write file: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Println(string(data))
		}
		return
	}

	printImpactDiffReport(allResults)
}

func printImpactDiffReport(results []*impact.BlastRadiusResult) {
	if len(results) == 0 {
		fmt.Println("✅ No downstream impact detected. Safe to merge.")
		return
	}

	totalAffected := 0
	counts := map[impact.SeverityLevel]int{}

	for _, r := range results {
		totalAffected += r.TotalAffected
		for sev, list := range r.BySeverity {
			counts[sev] += len(list)
		}
	}

	fmt.Println("\n🔍 GARUDA IMPACT DIFF REPORT")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Total Downstream Entities Affected: %d\n", totalAffected)
	fmt.Printf("Breakdown: [CRITICAL: %d] [HIGH: %d] [MEDIUM: %d] [LOW: %d]\n\n",
		counts[impact.SeverityCritical], counts[impact.SeverityHigh],
		counts[impact.SeverityMedium], counts[impact.SeverityLow])

	for _, sev := range []impact.SeverityLevel{
		impact.SeverityCritical,
		impact.SeverityHigh,
		impact.SeverityMedium,
		impact.SeverityLow,
	} {
		var list []impact.ImpactedEntity
		for _, r := range results {
			list = append(list, r.BySeverity[sev]...)
		}
		if len(list) == 0 {
			continue
		}

		list = impact.SortImpactedEntities(list)
		fmt.Printf("[%s] (%d findings):\n", string(sev), len(list))

		for _, e := range list {
			fmt.Printf("  • %s.%s (%s) — Depth: %d, Conf: %.2f [%s]\n",
				e.Package, e.Name, e.Kind, e.Depth, e.CumulativeConf, e.EpistemicClass)
			fmt.Printf("    Chain:    %s\n", strings.Join(e.TraversalChain, " → "))
			if e.Evidence.FilePath != "" {
				fmt.Printf("    Evidence: %s:%d-%d\n", e.Evidence.FilePath, e.Evidence.LineStart, e.Evidence.LineEnd)
			}
			fmt.Println()
		}
	}
}

func extractPackageFromEntity(name, entityID string) string {
	for _, candidate := range []string{name, entityID} {
		if candidate == "" {
			continue
		}
		if idx := strings.LastIndex(candidate, "."); idx > 0 && idx < len(candidate)-1 {
			return candidate[:idx]
		}
	}
	return ""
}
