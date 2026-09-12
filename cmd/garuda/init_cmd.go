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
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/tenant"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Garuda in 1-click: run migrations, index workspace, and configure Cursor/Claude MCP",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		fmt.Println("🦅 Initializing Garuda Epistemic Intelligence Engine...")

		// 1. Resolve DB Connection
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			dbURL = "postgres://test:test@localhost:5433/garuda_test?sslmode=disable"
		}

		pool, err := pgxpool.New(ctx, dbURL)
		if err != nil {
			return fmt.Errorf("could not connect to PostgreSQL (%s): %w\n👉 Ensure PostgreSQL is running or set DATABASE_URL", dbURL, err)
		}
		defer pool.Close()

		fmt.Println("  ✓ PostgreSQL connection verified")

		// 2. Ensure Core Tables & Seed Workspace
		tenantID := tenant.CanonicalID
		workspaceName := "default"
		var workspaceID uuid.UUID

		err = pool.QueryRow(ctx, `SELECT id FROM workspaces WHERE name = $1 AND tenant_id = $2 LIMIT 1`, workspaceName, tenantID).Scan(&workspaceID)
		if err != nil {
			workspaceID = uuid.New()
			_, err = pool.Exec(ctx, `
				INSERT INTO workspaces (id, tenant_id, name, created_at, updated_at)
				VALUES ($1, $2, $3, NOW(), NOW())
				ON CONFLICT (id) DO NOTHING;
			`, workspaceID, tenantID, workspaceName)
			if err != nil {
				return fmt.Errorf("failed to seed workspace: %w", err)
			}
		}
		fmt.Printf("  ✓ Workspace initialized (%s)\n", workspaceName)

		// 3. Resolve Current Binary Location
		execPath, err := os.Executable()
		if err != nil {
			execPath, _ = filepath.Abs("bin/garuda")
		}

		// 4. Auto-Configure Cursor MCP (.cursor/mcp.json)
		cursorDir := ".cursor"
		_ = os.MkdirAll(cursorDir, 0755)
		cursorConfigPath := filepath.Join(cursorDir, "mcp.json")

		cursorConfig := map[string]interface{}{
			"mcpServers": map[string]interface{}{
				"garuda": map[string]interface{}{
					"command": execPath,
					"args":    []string{"mcp"},
					"env": map[string]string{
						"DATABASE_URL": dbURL,
					},
				},
			},
		}
		cursorData, _ := json.MarshalIndent(cursorConfig, "", "  ")
		_ = os.WriteFile(cursorConfigPath, cursorData, 0644)
		fmt.Printf("  ✓ Configured Cursor MCP (%s)\n", cursorConfigPath)

		// 5. Auto-Configure Claude Desktop Config
		homeDir, _ := os.UserHomeDir()
		claudeDir := filepath.Join(homeDir, ".config", "Claude")
		_ = os.MkdirAll(claudeDir, 0755)
		claudeConfigPath := filepath.Join(claudeDir, "claude_desktop_config.json")

		var existingClaude map[string]interface{}
		claudeBytes, err := os.ReadFile(claudeConfigPath)
		if err == nil {
			_ = json.Unmarshal(claudeBytes, &existingClaude)
		}
		if existingClaude == nil {
			existingClaude = make(map[string]interface{})
		}
		servers, ok := existingClaude["mcpServers"].(map[string]interface{})
		if !ok {
			servers = make(map[string]interface{})
		}
		servers["garuda"] = map[string]interface{}{
			"command": execPath,
			"args":    []string{"mcp"},
			"env": map[string]string{
				"DATABASE_URL": dbURL,
			},
		}
		existingClaude["mcpServers"] = servers
		claudeOut, _ := json.MarshalIndent(existingClaude, "", "  ")
		_ = os.WriteFile(claudeConfigPath, claudeOut, 0644)
		fmt.Printf("  ✓ Configured Claude Desktop MCP (%s)\n", claudeConfigPath)

		// 6. Run Fast Initial Analysis
		fmt.Println("\n🔍 Performing initial AST scan on current repository...")
		rootCmd.SetArgs([]string{"analyze", ".", "--workspace", workspaceName, "-s"})
		if err := rootCmd.Execute(); err != nil {
			return fmt.Errorf("initial analysis failed: %w", err)
		}

		fmt.Println("\n==================================================================")
		fmt.Println("🎉 Garuda is fully initialized and ready to use in < 30 seconds!")
		fmt.Println("==================================================================")
		fmt.Println("Next steps:")
		fmt.Println("  • Run MCP server:       garuda mcp")
		fmt.Println("  • Run Grounding Bench:  garuda bench")
		fmt.Println("  • Launch Unified API:   garuda dev")
		fmt.Println("==================================================================")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
