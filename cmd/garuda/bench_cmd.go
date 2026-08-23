// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

//

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/benchmark"
	"github.com/spf13/cobra"
)

var benchCmd = &cobra.Command{
	Use:   "bench",
	Short: "Execute GAP-20 grounding benchmark harness (Naive vs Garuda MCP Grounded)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			dbURL = "postgres://test:test@localhost:5433/garuda_test?sslmode=disable"
		}

		pool, err := pgxpool.New(ctx, dbURL)
		if err != nil {
			return fmt.Errorf("database connection failed: %w", err)
		}
		defer pool.Close()

		tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		var workspaceID uuid.UUID
		err = pool.QueryRow(ctx, `SELECT id FROM workspaces WHERE name = 'uuid-ws' LIMIT 1`).Scan(&workspaceID)
		if err != nil {
			_ = pool.QueryRow(ctx, `SELECT id FROM workspaces LIMIT 1`).Scan(&workspaceID)
		}

		fmt.Println("🦅 Running GAP-20 Epistemic Grounding Benchmark Suite...")
		fmt.Printf("📦 Workspace: uuid-ws (%s)\n\n", workspaceID)

		runner := benchmark.NewRunner(pool, tenantID, workspaceID)
		report, err := runner.RunSuite(ctx, "uuid-ws")
		if err != nil {
			return fmt.Errorf("benchmark run failed: %w", err)
		}

		// Print Comparative Metric Table
		fmt.Println("==========================================================================================")
		fmt.Println("                       GARUDA GAP-20 GROUNDING BENCHMARK REPORT                          ")
		fmt.Println("==========================================================================================")
		fmt.Printf("%-32s | %-20s | %-20s | %-12s\n", "Metric Dimension", "Naive (Unassisted)", "Garuda MCP Grounded", "Delta / Gain")
		fmt.Println("------------------------------------------------------------------------------------------")
		fmt.Printf("%-32s | %-20.1f%% | %-20.1f%% | +%.1f%%\n", "Average Symbol Precision", report.NaiveMetrics.AvgPrecision*100, report.GarudaMCPMetrics.AvgPrecision*100, report.ImprovementFactor["precision_gain_pct"])
		fmt.Printf("%-32s | %-20.1f%% | %-20.1f%% | +%.1f%%\n", "Upstream Caller Recall", report.NaiveMetrics.AvgUpstreamRecall*100, report.GarudaMCPMetrics.AvgUpstreamRecall*100, (report.GarudaMCPMetrics.AvgUpstreamRecall-report.NaiveMetrics.AvgUpstreamRecall)*100)
		fmt.Printf("%-32s | %-20.1f%% | %-20.1f%% | +%.1f%%\n", "Downstream Dep Recall", report.NaiveMetrics.AvgDownstreamRecall*100, report.GarudaMCPMetrics.AvgDownstreamRecall*100, (report.GarudaMCPMetrics.AvgDownstreamRecall-report.NaiveMetrics.AvgDownstreamRecall)*100)
		fmt.Printf("%-32s | %-20.1f%% | %-20.1f%% | -%.1f%%\n", "Hallucination / Error Rate", report.NaiveMetrics.HallucinationRate, report.GarudaMCPMetrics.HallucinationRate, report.ImprovementFactor["hallucination_reduction_pct"])
		fmt.Printf("%-32s | %-20.1f%% | %-20.1f%% | +%.1f%%\n", "Violation Quarantine Rate", report.NaiveMetrics.ViolationCatchRate, report.GarudaMCPMetrics.ViolationCatchRate, report.ImprovementFactor["violation_catch_gain_pct"])
		fmt.Printf("%-32s | %-20d | %-20d | -%.1f%%\n", "Context Overhead (Tokens)", report.NaiveMetrics.AvgContextTokens, report.GarudaMCPMetrics.AvgContextTokens, report.ImprovementFactor["token_efficiency_gain_pct"])
		fmt.Println("==========================================================================================")

		// Save JSON Report
		outPath := "benchmark_report.json"
		data, _ := json.MarshalIndent(report, "", "  ")
		_ = os.WriteFile(outPath, data, 0644)
		fmt.Printf("\n📄 Full JSON benchmark dataset written to: %s\n", outPath)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(benchCmd)
}
