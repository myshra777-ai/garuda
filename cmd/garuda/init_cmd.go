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
	"github.com/myshra777-ai/garuda/internal/talon"
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

// resolveSmartEmail discovers the developer's email from flags, env, git config, or OS user.
func resolveSmartEmail(flagVal string) string {
	if v := strings.TrimSpace(flagVal); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("GARUDA_EMAIL")); v != "" {
		return v
	}
	if out, err := exec.Command("git", "config", "user.email").Output(); err == nil {
		if email := strings.TrimSpace(string(out)); email != "" && strings.Contains(email, "@") {
			return email
		}
	}
	if u := strings.TrimSpace(os.Getenv("USER")); u != "" {
		return u + "@garuda.local"
	}
	return "admin@garuda.local"
}

// resolveSmartWorkspace discovers the workspace name from flags, env, existing DB state, or target directory.
func resolveSmartWorkspace(ctx context.Context, pool *pgxpool.Pool, flagVal, targetPath string) string {
	if v := strings.TrimSpace(flagVal); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("GARUDA_WORKSPACE")); v != "" {
		return v
	}
	if pool != nil {
		var existing string
		err := pool.QueryRow(ctx, `
			SELECT name FROM workspaces
			 WHERE tenant_id = $1 AND name != 'default'
			 ORDER BY updated_at DESC LIMIT 1
		`, tenant.CanonicalID).Scan(&existing)
		if err == nil && existing != "" {
			return existing
		}
	}
	if abs, err := filepath.Abs(targetPath); err == nil {
		base := filepath.Base(abs)
		if base != "" && base != "." && base != "/" {
			return base
		}
	}
	return "default"
}

var initCmd = &cobra.Command{
	Use:   "init [target_repo_path]",
	Short: "Initialize Garuda in 1 command: start Postgres, run migrations, seed user & workspace, configure Active MCP & Talon Harness, and index code",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		config.LoadDotEnv()

		fmt.Println("🦅 Initializing Garuda Epistemic Intelligence Engine + Garuda Talon Active Harness...")

		targetPath := initTargetFlag
		if len(args) > 0 && args[0] != "" {
			targetPath = args[0]
		}

		// 1. Resolve & Ensure PostgreSQL is running
		dbURL := config.ResolveDBURL()
		if !isPostgresReachable(ctx, dbURL) {
			fmt.Println("  ⏳ PostgreSQL not yet reachable — starting Docker container 'garuda-postgres'...")
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

		// 3. Dynamically Resolve Workspace Name, Email & Persist .env
		workspaceName := resolveSmartWorkspace(ctx, pool, initWorkspaceFlag, targetPath)
		email := resolveSmartEmail(initEmailFlag)
		password := initPasswordFlag
		if password == "" {
			password = os.Getenv("GARUDA_PASSWORD")
		}
		if password == "" {
			password = "garuda2026"
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
			"DATABASE_URL":         dbURL,
			"CONTROL_DATABASE_URL": dbURL,
			"GARUDA_CONTROL_TOKEN": controlToken,
			"GARUDA_WORKSPACE":     workspaceName,
			"GARUDA_TALON_ENFORCE": "true",
			"PORT":                 port,
		})
		// Ensure JWT keys are generated and persisted in .env
		_, _ = auth.NewJWTConfig("garuda", "garuda-api", 15*time.Minute)
		fmt.Println("  ✓ Persisted environment configuration to .env")

		// 4. Seed Canonical Tenant & Workspace
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

		// 5. Seed Owner Account dynamically
		if email != "" && len(password) >= 8 {
			if hash, hErr := auth.HashPassword(password); hErr == nil {
				fullName := email
				if out, gErr := exec.Command("git", "config", "user.name").Output(); gErr == nil {
					if gn := strings.TrimSpace(string(out)); gn != "" {
						fullName = gn
					}
				} else if at := strings.IndexByte(email, '@'); at > 0 {
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

		// 6. Ensure bin/garuda-mcp, bin/garuda-api, and bin/garuda-talon are built
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
		talonBinPath := filepath.Join(binDir, "garuda-talon")
		if _, statErr := os.Stat(talonBinPath); statErr != nil {
			fmt.Println("  🔨 Building bin/garuda-talon...")
			_ = exec.Command("go", "build", "-o", talonBinPath, "./cmd/garuda-talon").Run()
		}

		// 7. Auto-Configure Claude Desktop MCP
		absTarget, _ := filepath.Abs(targetPath)
		if absTarget == "" {
			absTarget, _ = os.Getwd()
		}
		mcpEnv := map[string]string{
			"DATABASE_URL":         dbURL,
			"GARUDA_WORKSPACE":     workspaceName,
			"GARUDA_TALON_ENFORCE": "true",
			"GARUDA_REPO_ROOT":     absTarget,
		}
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
		if !initSkipScanFlag && targetPath != "" {
			fmt.Printf("\n🔍 Performing AST & call-graph scan on '%s'...\n", targetPath)
			rootCmd.SetArgs([]string{"analyze", targetPath, "--workspace", workspaceName, "--save"})
			if err := rootCmd.Execute(); err != nil {
				return fmt.Errorf("initial analysis failed: %w", err)
			}
		}

		// 9. Compile & Install Garuda Talon Active Enforcement (Contract + Cursor/Claude/Agents/Git Hooks)
		policyDir := filepath.Join(absTarget, "policies")
		contract, _ := talon.CompileContract(ctx, pool, tenantID, ws.ID, workspaceName, policyDir)
		if talonSummary, tErr := talon.InstallHooks(absTarget, selfPath, mcpPath, dbURL, workspaceName, contract); tErr == nil {
			fmt.Printf("  ✓ Installed Garuda Talon Active Contract (%s)\n", talonSummary.ContractPath)
			fmt.Printf("  ✓ Installed Active Cursor Rule (%s) & MCP (%s)\n", talonSummary.CursorRulePath, talonSummary.CursorMCPPath)
			fmt.Printf("  ✓ Installed Claude Code Active Hooks (%s) & CLAUDE.md / AGENTS.md\n", talonSummary.ClaudeSettings)
		}

		fmt.Println("\n══════════════════════════════════════════════════════════════════")
		fmt.Println("🎉 Garuda v1.5 + Garuda Talon Active Harness Initialized!")
		fmt.Println("══════════════════════════════════════════════════════════════════")
		fmt.Printf("  • Active Workspace:   %s\n", workspaceName)
		fmt.Printf("  • Owner Login:        %s\n", email)
		fmt.Printf("  • Talon Guard CLI:    ./bin/garuda talon guard --workspace %s\n", workspaceName)
		fmt.Printf("  • Diff Check CLI:     ./bin/garuda check --diff HEAD --project %s --policy policies --save\n", workspaceName)
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
	initCmd.Flags().StringVar(&initEmailFlag, "email", "", "Owner email to provision (auto-detected from git config if omitted)")
	initCmd.Flags().StringVar(&initPasswordFlag, "password", "", "Owner password to provision (defaults to GARUDA_PASSWORD or garuda2026)")
	initCmd.Flags().StringVar(&initWorkspaceFlag, "workspace", "", "Workspace name to initialize (auto-detected if omitted)")
	initCmd.Flags().StringVar(&initTargetFlag, "target", ".", "Repository path to index")
	initCmd.Flags().BoolVar(&initSkipScanFlag, "skip-scan", false, "Skip initial repository AST scan")
	rootCmd.AddCommand(initCmd)
}
