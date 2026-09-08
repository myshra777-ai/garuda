// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/myshra777-ai/garuda/internal/store"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage workspaces (logical groups of repositories)",
}

var workspaceCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new workspace",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleWorkspaceCreate(args[0])
	},
}

var workspaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workspaces",
	Run: func(cmd *cobra.Command, args []string) {
		handleWorkspaceList()
	},
}

var workspaceDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a workspace (and all its repositories)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleWorkspaceDelete(args[0])
	},
}

var workspaceSyncCmd = &cobra.Command{
	Use:   "sync [workspace-name]",
	Short: "Analyze all enabled repositories in a workspace (parallel) and update the Company Graph",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleWorkspaceSync(args[0])
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceCreateCmd)
	workspaceCmd.AddCommand(workspaceListCmd)
	workspaceCmd.AddCommand(workspaceDeleteCmd)
	workspaceCmd.AddCommand(workspaceSyncCmd)
	rootCmd.AddCommand(workspaceCmd)
}

func handleWorkspaceCreate(name string) {
	dbURL := getDBURL()
	tenantID := getTenantIDString()
	ctx := context.Background()

	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	ws, err := st.CreateWorkspace(ctx, tenantID, name, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to create workspace: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Workspace '%s' created (ID: %s)\n", ws.Name, ws.ID)
}

func handleWorkspaceList() {
	dbURL := getDBURL()
	tenantID := getTenantIDString()
	ctx := context.Background()

	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	workspaces, err := st.ListWorkspaces(ctx, tenantID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to list: %v\n", err)
		os.Exit(1)
	}
	if len(workspaces) == 0 {
		fmt.Println("📂 No workspaces found. Create one with: garuda workspace create <name>")
		return
	}

	fmt.Printf("📂 Workspaces (%d):\n", len(workspaces))
	for _, w := range workspaces {
		fmt.Printf("  • %s (ID: %s)\n", w.Name, w.ID)
	}
}

func handleWorkspaceDelete(name string) {
	dbURL := getDBURL()
	tenantID := getTenantIDString()
	ctx := context.Background()

	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	var wsID string
	err = st.Pool().QueryRow(ctx, `SELECT id FROM workspaces WHERE tenant_id = $1 AND name = $2`, tenantID, name).Scan(&wsID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Workspace '%s' not found\n", name)
		os.Exit(1)
	}

	_, err = st.Pool().Exec(ctx, `DELETE FROM workspaces WHERE id = $1`, wsID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to delete: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Workspace '%s' deleted.\n", name)
}

func handleWorkspaceSync(workspaceName string) {
	tenantID := getTenantIDString()
	dbURL := getDBURL()
	ctx := context.Background()

	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	var wsID uuid.UUID
	err = st.Pool().QueryRow(ctx, `SELECT id FROM workspaces WHERE tenant_id = $1 AND name = $2`, tenantID, workspaceName).Scan(&wsID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Workspace '%s' not found\n", workspaceName)
		os.Exit(1)
	}

	repos, err := st.ListRepositories(ctx, wsID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to list repositories: %v\n", err)
		os.Exit(1)
	}
	if len(repos) == 0 {
		fmt.Printf("⚠️ No repositories in workspace '%s'.\n", workspaceName)
		return
	}

	fmt.Printf("🔄 Syncing workspace '%s' (%d repos)...\n", workspaceName, len(repos))
	for i, repo := range repos {
		fmt.Printf("[%d/%d] %s\n", i+1, len(repos), repo.URL)
		if !repo.Enabled {
			fmt.Printf("  ⏭️ Skipping disabled repository\n")
			continue
		}

		tempDir := filepath.Join(os.TempDir(), "garuda-sync", workspaceName, repo.ID.String())
		if _, err := os.Stat(tempDir); os.IsNotExist(err) {
			if strings.HasPrefix(repo.URL, "file://") {
				localPath := strings.TrimPrefix(repo.URL, "file://")
				if !filepath.IsAbs(localPath) {
					if absPath, err := filepath.Abs(localPath); err == nil {
						localPath = absPath
					}
				}
				if _, err := os.Stat(localPath); err == nil {
					fmt.Printf("  📂 Copying local repo from %s\n", localPath)
					if err := os.MkdirAll(filepath.Dir(tempDir), 0755); err != nil {
						fmt.Printf("  ❌ Failed to create directory: %v\n", err)
						continue
					}
					cmd := exec.Command("cp", "-r", localPath, tempDir)
					if err := cmd.Run(); err != nil {
						fmt.Printf("  ❌ Failed to copy local repo: %v\n", err)
						continue
					}
				} else {
					fmt.Printf("  ❌ Local path not found: %s\n", localPath)
					continue
				}
			} else {
				cmd := exec.Command("git", "clone", repo.URL, tempDir)
				if err := cmd.Run(); err != nil {
					fmt.Printf("  ❌ Failed to clone: %v\n", err)
					continue
				}
			}
		} else {
			if _, err := os.Stat(filepath.Join(tempDir, ".git")); err == nil {
				cmd := exec.Command("git", "-C", tempDir, "pull")
				if err := cmd.Run(); err != nil {
					fmt.Printf("  ⚠️ Failed to pull (continuing with existing): %v\n", err)
				}
			} else {
				fmt.Printf("  ⚠️ Not a git repository, skipping pull\n")
			}
		}

		commitCmd := exec.Command("git", "-C", tempDir, "rev-parse", "HEAD")
		commitOutput, err := commitCmd.Output()
		if err != nil {
			fmt.Printf("  ❌ Failed to get commit: %v\n", err)
			continue
		}
		commitSHA := strings.TrimSpace(string(commitOutput))

		modulePath, err := store.DetectModulePath(tempDir)
		if err != nil {
			fmt.Printf("  ⚠️ Could not detect module path: %v\n", err)
			modulePath = ""
		} else {
			fmt.Printf("  📦 Module path: %s\n", modulePath)
		}

		fmt.Printf("  🔍 Analysing...\n")
		analyzeArgs := []string{"analyze", tempDir, "--save", "--workspace", workspaceName, "--repo", repo.URL, "--commit", commitSHA}
		if modulePath != "" {
			analyzeArgs = append(analyzeArgs, "--module-path", modulePath)
		}

		analyzeCmd := exec.Command("./garuda", analyzeArgs...)
		analyzeCmd.Env = append(os.Environ(), "DATABASE_URL="+os.Getenv("DATABASE_URL"), "GARUDA_TENANT_ID="+tenantID)

		output, err := analyzeCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("  ❌ Failed: %v\n", err)
			fmt.Printf("  Output: %s\n", string(output))
			continue
		}
		fmt.Printf("  ✅ Done\n")

		err = st.UpdateRepositorySyncStatus(ctx, tenantID, repo.ID, commitSHA, "synced")
		if err != nil {
			fmt.Printf("  ⚠️ Failed to update status: %v\n", err)
		}
	}
	fmt.Println("🎉 Sync completed.")
}
