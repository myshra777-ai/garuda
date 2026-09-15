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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/benchmark"
	"github.com/myshra777-ai/garuda/internal/store"
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

		tenantID := getTenantID()
		workspaceName := getWorkspaceName()

		// Resolve via the same tenant-scoped helper the rest of the CLI
		// uses. The previous query was `SELECT id, name FROM workspaces
		// ORDER BY updated_at DESC LIMIT 1` with no filter at all: it
		// returned the most recently updated workspace in the entire
		// deployment, regardless of tenant, regardless of the caller's
		// GARUDA_WORKSPACE. The resolved workspace and the resolved
		// tenant could then belong to different tenants, and the runner
		// downstream was given that inconsistent pair.
		//
		// Verified 2026-09-15: with GARUDA_WORKSPACE=go-validation-10 and
		// the canonical tenant active, the command resolved to
		// 08545e15-... (name "default", tenant e2f82aa6-...), a workspace
		// in a different tenant entirely.
		workspaceID, err := store.ResolveWorkspaceID(ctx, pool, tenantID, workspaceName)
		if err != nil {
			return fmt.Errorf("resolve workspace %q for tenant %s: %w", workspaceName, tenantID, err)
		}
		var resolvedName string
		_ = pool.QueryRow(ctx, `SELECT name FROM workspaces WHERE id = $1`, workspaceID).Scan(&resolvedName)
		if resolvedName != "" {
			workspaceName = resolvedName
		}

		fmt.Println("🦅 Running GAP-20 Epistemic Grounding Benchmark Suite...")
		fmt.Printf("📦 Workspace: %s (%s)\n\n", workspaceName, workspaceID)

		runner := benchmark.NewRunner(pool, tenantID, workspaceID)
		report, err := runner.RunSuite(ctx, workspaceName)
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
