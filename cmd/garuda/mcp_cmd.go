// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

//

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run the Garuda Model Context Protocol (MCP) server over standard I/O",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			dbURL = "postgres://test:test@localhost:5433/garuda_test?sslmode=disable"
		}

		pool, err := pgxpool.New(ctx, dbURL)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer pool.Close()

		tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		var workspaceID uuid.UUID
		err = pool.QueryRow(ctx, `SELECT id FROM workspaces WHERE name = 'uuid-ws' LIMIT 1`).Scan(&workspaceID)
		if err != nil {
			_ = pool.QueryRow(ctx, `SELECT id FROM workspaces LIMIT 1`).Scan(&workspaceID)
		}

		server := mcp.NewServer(pool, tenantID, workspaceID)
		return server.ServeStdio(ctx, os.Stdin, os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
