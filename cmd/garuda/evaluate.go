package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/evaluation"
)

var evaluateCmd = &cobra.Command{
	Use:   "evaluate [baseline.json] [proposed.json]",
	Short: "Evaluate a change between two snapshots (semantic operationality spike)",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		baseline, err := analyzer.LoadResult(args[0])
		if err != nil {
			fmt.Printf("❌ Failed to load baseline: %v\n", err)
			os.Exit(1)
		}
		proposed, err := analyzer.LoadResult(args[1])
		if err != nil {
			fmt.Printf("❌ Failed to load proposed: %v\n", err)
			os.Exit(1)
		}
		assessment, err := evaluation.AssessChange(baseline, proposed)
		if err != nil {
			fmt.Printf("❌ Evaluation failed: %v\n", err)
			os.Exit(1)
		}
		data, err := json.MarshalIndent(assessment, "", "  ")
		if err != nil {
			fmt.Printf("❌ Failed to marshal assessment: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	},
}

func init() {
	rootCmd.AddCommand(evaluateCmd)
}
