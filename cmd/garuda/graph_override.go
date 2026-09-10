package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/knowledge"
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

		tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		workspace := "default"

		svc := knowledge.NewGraphService(pool)
		graph, err := svc.BuildUnifiedGraph(ctx, tenantID, workspace)
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
