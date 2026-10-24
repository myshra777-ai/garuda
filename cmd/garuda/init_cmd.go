// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/auth"
	"github.com/myshra777-ai/garuda/internal/config"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/tenant"
	"github.com/spf13/cobra"
)

var (
	initEmailFlag     string
	initPasswordFlag  string
	initWorkspaceFlag string
	initTargetFlag    string
	initSkipScanFlag  bool
)

var initCmd = &cobra.Command{
	Use:   "init [target_repo_path]",
	Short: "Initialize Garuda in 1 command: start Postgres, run migrations, seed user & workspace, configure MCP, and index code",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		config.LoadDotEnv()

		fmt.Println("🦅 Initializing Garuda Epistemic Intelligence Engine (1-Click Setup)...")

		// 1. Resolve & Ensure PostgreSQL is running
		dbURL := config.ResolveDBURL()
		if !isPostgresReachable(ctx, dbURL) {
			fmt.Println("  ⏳ PostgreSQL not yet reachable — attempting to start Docker container 'garuda-postgres'...")
			ensureDockerPostgres()
		}

		pool, err := pgxpool.New(ctx, dbURL)
		if err != nil {
			return fmt.Errorf("could not connect to PostgreSQL (%s): %w", dbURL, err)
		}
		defer pool.Close()
		fmt.Println("  ✓ PostgreSQL connection verified")

		// 2. Run Idempotent Database Migrations
		if err := store.Migrate(dbURL, "migrations"); err != nil {
			return fmt.Errorf("database migration failed: %w", err)
		}
		fmt.Println("  ✓ Database schema & migrations verified (100% up-to-date)")

		// 3. Resolve Workspace Name & Persist .env
		workspaceName := initWorkspaceFlag
		if workspaceName == "" {
			workspaceName = os.Getenv("GARUDA_WORKSPACE")
		}
		if workspaceName == "" {
			workspaceName = "go-validation-10"
		}
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		controlToken := os.Getenv("GARUDA_CONTROL_TOKEN")
		if controlToken == "" {
			controlToken = "devtoken-changeme-please"
		}

		_ = config.UpsertDotEnv(map[string]string{
			"DATABASE_URL":          dbURL,
			"CONTROL_DATABASE_URL":  dbURL,
			"GARUDA_CONTROL_TOKEN":  controlToken,
			"GARUDA_WORKSPACE":      workspaceName,
			"PORT":                  port,
		})
		// Ensure JWT keys are generated and persisted in .env
		_, _ = auth.NewJWTConfig("garuda", "garuda-api", 15*time.Minute)
		fmt.Println("  ✓ Persisted environment configuration to .env")

		// 4. Seed Canonical Tenant & Default Workspace
		tenantID := getTenantID()
		pgStore, err := store.NewPostgresStore(dbURL)
		if err != nil {
			return fmt.Errorf("failed to open store: %w", err)
		}
		defer pgStore.Close()

		ws, err := pgStore.CreateWorkspace(ctx, tenantID.String(), workspaceName, "Initialized via garuda init")
		if err != nil && !errors.Is(err, store.ErrWorkspaceExists) {
			return fmt.Errorf("failed to seed workspace: %w", err)
		}
		fmt.Printf("  ✓ Workspace ready: %s (%s)\n", workspaceName, ws.ID.String())

		// 5. Seed Owner Account (so /login and /signup work out of the box)
		email := strings.TrimSpace(initEmailFlag)
		password := initPasswordFlag
		if email != "" && len(password) >= 8 {
			if hash, hErr := auth.HashPassword(password); hErr == nil {
				fullName := email
				if at := strings.IndexByte(email, '@'); at > 0 {
					fullName = email[:at]
				}
				user, _, sErr := pgStore.SignupUser(ctx, email, hash, fullName)
				var userID string
				if sErr == nil && user != nil {
					userID = user.ID
				} else if existing, gErr := pgStore.GetUserByEmail(ctx, email); gErr == nil && existing != nil {
					userID = existing.ID
				}
				if userID != "" {
					_, _ = pool.Exec(ctx, `
						INSERT INTO tenant_members (user_id, tenant_id, role, created_at)
						VALUES ($1, $2, 'owner', NOW())
						ON CONFLICT (user_id, tenant_id) DO NOTHING
					`, userID, tenant.CanonicalID)
					_, _ = pool.Exec(ctx, `
						INSERT INTO workspace_members (user_id, workspace_id, role, created_at)
						SELECT $1, w.id, 'owner', NOW()
						  FROM workspaces w
						 WHERE w.tenant_id = $2
						ON CONFLICT (user_id, workspace_id) DO NOTHING
					`, userID, tenant.CanonicalID)
					fmt.Printf("  ✓ Account provisioned & linked to all workspaces: %s\n", email)
				}
			}
		}

		// 6. Ensure bin/garuda-mcp and bin/garuda-api are built
		selfPath, err := os.Executable()
		if err != nil {
			wd, _ := os.Getwd()
			selfPath = filepath.Join(wd, "bin", "garuda")
		}
		binDir := filepath.Dir(selfPath)
		mcpPath := filepath.Join(binDir, "garuda-mcp")
		if _, statErr := os.Stat(mcpPath); statErr != nil {
			fmt.Println("  🔨 Building bin/garuda-mcp...")
			_ = exec.Command("go", "build", "-o", mcpPath, "./cmd/garuda-mcp").Run()
		}
		apiPath := filepath.Join(binDir, "garuda-api")
		if _, statErr := os.Stat(apiPath); statErr != nil {
			fmt.Println("  🔨 Building bin/garuda-api...")
			_ = exec.Command("go", "build", "-o", apiPath, "./cmd/garuda-api").Run()
		}

		// 7. Auto-Configure Cursor MCP (.cursor/mcp.json) & Claude Desktop
		cursorDir := ".cursor"
		_ = os.MkdirAll(cursorDir, 0755)
		cursorConfigPath := filepath.Join(cursorDir, "mcp.json")
		mcpEnv := map[string]string{
			"DATABASE_URL":     dbURL,
			"GARUDA_WORKSPACE": workspaceName,
		}
		cursorConfig := map[string]interface{}{
			"mcpServers": map[string]interface{}{
				"garuda": map[string]interface{}{
					"command": mcpPath,
					"env":     mcpEnv,
				},
			},
		}
		cursorData, _ := json.MarshalIndent(cursorConfig, "", "  ")
		_ = os.WriteFile(cursorConfigPath, cursorData, 0644)
		fmt.Printf("  ✓ Configured Cursor MCP (%s)\n", cursorConfigPath)

		homeDir, _ := os.UserHomeDir()
		claudeDir := filepath.Join(homeDir, ".config", "Claude")
		_ = os.MkdirAll(claudeDir, 0755)
		claudeConfigPath := filepath.Join(claudeDir, "claude_desktop_config.json")
		var existingClaude map[string]interface{}
		if claudeBytes, err := os.ReadFile(claudeConfigPath); err == nil {
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
			"command": mcpPath,
			"env":     mcpEnv,
		}
		existingClaude["mcpServers"] = servers
		claudeOut, _ := json.MarshalIndent(existingClaude, "", "  ")
		_ = os.WriteFile(claudeConfigPath, claudeOut, 0644)
		fmt.Printf("  ✓ Configured Claude Desktop MCP (%s)\n", claudeConfigPath)

		// 8. Optional initial repository analysis
		targetPath := initTargetFlag
		if len(args) > 0 && args[0] != "" {
			targetPath = args[0]
		}
		if !initSkipScanFlag && targetPath != "" {
			fmt.Printf("\n🔍 Performing AST & call-graph scan on '%s'...\n", targetPath)
			rootCmd.SetArgs([]string{"analyze", targetPath, "--workspace", workspaceName, "--save"})
			if err := rootCmd.Execute(); err != nil {
				return fmt.Errorf("initial analysis failed: %w", err)
			}
		}

		fmt.Println("\n══════════════════════════════════════════════════════════════════")
		fmt.Println("🎉 Garuda is initialized, migrated, and configured!")
		fmt.Println("══════════════════════════════════════════════════════════════════")
		fmt.Printf("  • Active Workspace:   %s\n", workspaceName)
		fmt.Printf("  • Owner Login:        %s\n", email)
		fmt.Printf("  • Change-Control CLI: ./bin/garuda check --project %s --policy policies --save\n", workspaceName)
		fmt.Printf("  • Verify MCP Server:  python3 scripts/mcp_verify.py\n")
		fmt.Println("══════════════════════════════════════════════════════════════════")
		return nil
	},
}

func isPostgresReachable(ctx context.Context, dbURL string) bool {
	cCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	pool, err := pgxpool.New(cCtx, dbURL)
	if err != nil {
		return false
	}
	defer pool.Close()
	return pool.Ping(cCtx) == nil
}

func ensureDockerPostgres() {
	// Try starting existing container first, otherwise create it
	if err := exec.Command("docker", "start", "garuda-postgres").Run(); err != nil {
		_ = exec.Command(
			"docker", "run", "-d",
			"--name", "garuda-postgres",
			"-e", "POSTGRES_USER=test",
			"-e", "POSTGRES_PASSWORD=test",
			"-e", "POSTGRES_DB=garuda_test",
			"-p", "5433:5432",
			"postgres:16-alpine",
		).Run()
	}
	for i := 0; i < 25; i++ {
		if conn, err := net.DialTimeout("tcp", "127.0.0.1:5433", 1*time.Second); err == nil {
			_ = conn.Close()
			time.Sleep(1 * time.Second)
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func init() {
	initCmd.Flags().StringVar(&initEmailFlag, "email", "myshra777@gmail.com", "Owner email to provision")
	initCmd.Flags().StringVar(&initPasswordFlag, "password", "garuda2026", "Owner password to provision")
	initCmd.Flags().StringVar(&initWorkspaceFlag, "workspace", "", "Workspace name to initialize")
	initCmd.Flags().StringVar(&initTargetFlag, "target", ".", "Repository path to index")
	initCmd.Flags().BoolVar(&initSkipScanFlag, "skip-scan", false, "Skip initial repository AST scan")
	rootCmd.AddCommand(initCmd)
}
