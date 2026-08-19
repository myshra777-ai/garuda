// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/judge"
	"github.com/spf13/cobra"
)

var (
	judgeOutput string
	judgeJSON   bool
	judgeBlock  bool
)

func init() {
	judgeCmd.Flags().StringVarP(&judgeOutput, "output", "o", "", "Write report to file")
	judgeCmd.Flags().BoolVar(&judgeJSON, "json", false, "Output JSON")
	judgeCmd.Flags().BoolVar(&judgeBlock, "block", false, "Exit with non-zero code if blocking issues found")
	rootCmd.AddCommand(judgeCmd)
}

var judgeCmd = &cobra.Command{
	Use:   "judge <baseline.json> <proposed.json>",
	Short: "Compare two snapshots and produce a governance judgement",
	Long: `Evaluates changes between two snapshots using Ponytail principles:
  - Breaking changes (signature/contract changes)
  - Dead code introduced (new entities with zero incoming calls)
  - Duplications introduced
  - Standard library alternatives

Produces a judgement report with recommendations and a blocking decision.

Examples:
  garuda judge v1.json v2.json
  garuda judge v1.json v2.json --json -o report.json
  garuda judge v1.json v2.json --block`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		handleJudge(args[0], args[1])
	},
}

func handleJudge(baselineFile, proposedFile string) {
	before, err := analyzer.LoadResult(baselineFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to load baseline: %v\n", err)
		os.Exit(1)
	}
	after, err := analyzer.LoadResult(proposedFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to load proposed: %v\n", err)
		os.Exit(1)
	}

	// Enrich with graph data (incoming references) to detect dead code
	// For simplicity we use the diff report which already has impact counts
	report := judge.Judge(before, after)

	if judgeJSON {
		data, _ := json.MarshalIndent(report, "", "  ")
		if judgeOutput != "" {
			os.WriteFile(judgeOutput, data, 0644)
			fmt.Printf("📄 JSON report written to %s\n", judgeOutput)
		} else {
			fmt.Println(string(data))
		}
	} else {
		printJudgeReport(report)
	}

	if judgeBlock && report.Block {
		os.Exit(1)
	}
}

func printJudgeReport(r *judge.Report) {
	fmt.Println("⚖️ GOVERNANCE JUDGEMENT")
	fmt.Println("🧠 Judge analysis complete.")
	fmt.Printf("📊 Summary:\n")
	fmt.Printf("  • Total changes: %d\n", r.TotalChanges)
	fmt.Printf("  • Breaking: %d ⚠️\n", r.BreakingCount)
	fmt.Printf("  • Duplications: %d\n", r.DuplicationCount)
	fmt.Printf("  • Dead code introduced: %d\n\n", r.DeadCodeCount)

	if len(r.BreakingChanges) > 0 {
		fmt.Printf("🚫 BREAKING CHANGES (%d):\n", len(r.BreakingChanges))
		for _, bc := range r.BreakingChanges {
			fmt.Printf("  │ %s\n", bc.Entity)
			fmt.Printf("  │   • Consumers: %d services will break\n", bc.ImpactCount)
			fmt.Printf("  │   • Mitigation: %s\n", bc.Mitigation)
			fmt.Printf("  │\n")
			fmt.Printf("  │ └── Evidence: %v\n", bc.Evidence)
			fmt.Printf("  │\n")
		}
		fmt.Println()
	}

	if len(r.Duplications) > 0 {
		fmt.Printf("🟡 DUPLICATIONS (%d):\n", len(r.Duplications))
		for _, dup := range r.Duplications {
			fmt.Printf("  │ %s duplicates %s\n", dup.NewEntity, dup.ExistingEntity)
			fmt.Printf("  │   → %s\n", dup.Recommendation)
			fmt.Printf("  │\n")
		}
		fmt.Println()
	}

	if len(r.DeadCode) > 0 {
		fmt.Printf("🔴 DEAD CODE INTRODUCED (%d):\n", len(r.DeadCode))
		for _, dc := range r.DeadCode {
			fmt.Printf("  │ %s in %s\n", dc.Entity, dc.File)
			fmt.Printf("  │   → No incoming calls found\n")
			fmt.Printf("  │\n")
		}
		fmt.Println()
	}

	if len(r.StdLibAlternatives) > 0 {
		fmt.Printf("📚 STANDARD LIBRARY ALTERNATIVES:\n")
		for _, alt := range r.StdLibAlternatives {
			fmt.Printf("  │ %s → use %s\n", alt.Entity, alt.Alternative)
			fmt.Printf("  │\n")
		}
		fmt.Println()
	}

	fmt.Printf("💡 RECOMMENDATIONS:\n")
	for _, rec := range r.Recommendations {
		fmt.Printf("  %s\n", rec)
	}
	fmt.Println()

	if r.Block {
		fmt.Printf("❌ JUDGEMENT: BLOCK – %s\n", r.BlockReason)
	} else {
		fmt.Printf("✅ JUDGEMENT: PASS – %s\n", r.PassReason)
	}
}
