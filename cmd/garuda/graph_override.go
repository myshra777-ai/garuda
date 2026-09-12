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
	"github.com/myshra777-ai/garuda/internal/knowledge"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/tenant"
	"github.com/spf13/cobra"
)

var graphExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export the unified Docs-to-Code knowledge graph as JSON for UI dashboards",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			dbURL = "postgres://garuda:garudapassword@localhost:5432/garuda?sslmode=disable"
		}

		ctx := context.Background()
		pool, err := pgxpool.New(ctx, dbURL)
		if err != nil {
			return fmt.Errorf("database connection failed: %w", err)
		}
		defer pool.Close()

		tenantID := tenant.CanonicalID

		workspaceID, err := store.ResolveWorkspaceID(ctx, pool, tenantID, "")
		if err != nil {
			return fmt.Errorf("resolve workspace: %w", err)
		}

		svc := knowledge.NewGraphService(pool)
		graph, err := svc.BuildUnifiedGraph(ctx, tenantID, workspaceID)
		if err != nil {
			return fmt.Errorf("failed to build graph: %w", err)
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(graph)
	},
}

func init() {
	// Attach export to the existing graphCmd from inspect.go if it exists
	if graphCmd != nil {
		graphCmd.AddCommand(graphExportCmd)
	}
}
