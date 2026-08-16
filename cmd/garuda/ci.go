// cmd/garuda/ci.go

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/evaluation"
)

var (
	baselineFile string
	outputFile   string
	blockOnBreak bool
)

var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "Run Garuda in CI mode – analyze and compare with baseline",
	Long: `Garuda CI runs the analysis on the current codebase and compares it with
a baseline snapshot. It outputs a report and exits with a non-zero code
if breaking changes are detected (when --block-on-break is set).`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Run analysis on current code
		fmt.Println("🔍 Analyzing current codebase...")
		currentResult, err := runAnalysis(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Analysis failed: %v\n", err)
			os.Exit(1)
		}

		// 2. Load baseline
		baselineResult, err := loadBaseline(baselineFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to load baseline: %v\n", err)
			os.Exit(1)
		}

		// 3. Diff and evaluate
		report := analyzer.Diff(baselineResult, currentResult)
		assessment, err := evaluation.AssessChange(baselineResult, currentResult)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Evaluation failed: %v\n", err)
			os.Exit(1)
		}

		// 4. Output report to file (if requested)
		if outputFile != "" {
			data, _ := json.MarshalIndent(report, "", "  ")
			if err := os.WriteFile(outputFile, data, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "⚠️ Failed to write output: %v\n", err)
			}
		}

		// 5. Print human‑readable summary
		printCIReport(report, assessment)

		// 6. Exit with error if blocking is enabled and breaking changes exist
		if blockOnBreak && assessment.Status == evaluation.StatusBreaking {
			fmt.Println("\n❌ Breaking changes detected. CI blocked.")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(ciCmd)
	ciCmd.Flags().StringVar(&baselineFile, "baseline", "garuda-baseline.json", "Path to baseline snapshot")
	ciCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write JSON report to file")
	ciCmd.Flags().BoolVar(&blockOnBreak, "block-on-break", false, "Exit with error if breaking changes are detected")
}

func runAnalysis(path string) (*analyzer.Result, error) {
	return analyzer.Analyze(path)
}

func loadBaseline(path string) (*analyzer.Result, error) {
	return analyzer.LoadResult(path)
}

func printCIReport(report *analyzer.DiffReport, assessment *evaluation.Assessment) {
	fmt.Println("\n📊 GARUDA CI REPORT")
	fmt.Println("───────────────────")
	fmt.Printf("  Status:        %s\n", assessment.Status)
	fmt.Printf("  Confidence:    %.2f\n", assessment.Confidence)
	fmt.Printf("  Breaking:      %d\n", report.Summary.BreakingChanges)
	fmt.Printf("  Warnings:      %d\n", report.Summary.Warnings)
	fmt.Printf("  Additions:     %d\n", report.Summary.Additions)
	fmt.Printf("  Removals:      %d\n", report.Summary.Removals)
	fmt.Printf("  Modified:      %d\n", report.Summary.Modified)

	if len(report.EntityDiffs) > 0 {
		fmt.Println("\n  Entity Changes:")
		for _, ed := range report.EntityDiffs {
			fmt.Printf("    %s %s %s\n", ed.Status, ed.Kind, ed.Name)
		}
	}
	if len(report.RelationshipDiffs) > 0 {
		fmt.Println("\n  Relationship Changes:")
		for _, rd := range report.RelationshipDiffs {
			fmt.Printf("    %s %s %s -> %s\n", rd.Status, rd.Type, rd.From, rd.To)
		}
	}

	if assessment.Status == evaluation.StatusBreaking {
		fmt.Println("\n  🔴 Breaking changes detected. Review before merging.")
	} else if assessment.Status == evaluation.StatusRedundant {
		fmt.Println("\n  🟡 Redundant code detected. Consider consolidating.")
	} else {
		fmt.Println("\n  ✅ No critical issues detected.")
	}
}
