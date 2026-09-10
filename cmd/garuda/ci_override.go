// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

// This file ensures ciCheckCmd is attached to the existing ciCmd safely without redeclaration.
import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/knowledge"
	"github.com/spf13/cobra"
)

var ciCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Evaluate workspace knowledge integrity and fail if architectural contradictions exist",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			dbURL = "postgres://garuda:garudapassword@localhost:5432/garuda?sslmode=disable"
		}

		failOnContradiction, _ := cmd.Flags().GetBool("fail-on-contradiction")

		ctx := context.Background()
		pool, err := pgxpool.New(ctx, dbURL)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer pool.Close()

		tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		workspace := "default"

		fmt.Println("🛡️ Running Garuda Software Knowledge Integrity Gate...")
		evaluator := knowledge.NewEvaluator(pool)
		stats, _, err := evaluator.EvaluateWorkspace(ctx, tenantID, workspace)
		if err != nil {
			return fmt.Errorf("evaluation failed: %w", err)
		}

		fmt.Printf("\nCI GATE EVALUATION SUMMARY\n")
		fmt.Printf("==========================\n")
		fmt.Printf("Total Document Claims : %d\n", stats.TotalClaims)
		fmt.Printf("Supported (Passing)   : %d\n", stats.Supported)
		fmt.Printf("Unimplemented (Drift) : %d\n", stats.Unimplemented)
		fmt.Printf("Contradictions        : %d\n", stats.Contradicted)

		if stats.Contradicted > 0 && failOnContradiction {
			fmt.Printf("\n❌ BUILD FAILED: Garuda detected %d explicit architectural contradictions.\n", stats.Contradicted)
			fmt.Println("Please resolve code violations against your architectural decision records (ADRs) before merging.")
			os.Exit(1)
		}

		fmt.Printf("\n🎉 CI Gate passed successfully!\n")
		return nil
	},
}

func init() {
	ciCheckCmd.Flags().Bool("fail-on-contradiction", true, "Exit with non-zero code if architectural contradictions are found")
	// Attach to existing ciCmd from ci.go
	if ciCmd != nil {
		ciCmd.AddCommand(ciCheckCmd)
	}
}
