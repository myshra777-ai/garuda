// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/myshra777-ai/garuda/internal/store"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage repositories within a workspace",
}

var repoAddCmd = &cobra.Command{
	Use:   "add [workspace-name] [repo-url]",
	Short: "Add a repository to a workspace",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		workspaceName := args[0]
		repoURL := args[1]
		modulePath, _ := cmd.Flags().GetString("module-path")
		handleRepoAdd(workspaceName, repoURL, modulePath)
	},
}

var repoListCmd = &cobra.Command{
	Use:   "list [workspace-name]",
	Short: "List repositories in a workspace",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleRepoList(args[0])
	},
}

var repoRemoveCmd = &cobra.Command{
	Use:   "remove [workspace-name] [repo-url]",
	Short: "Remove a repository from a workspace",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		handleRepoRemove(args[0], args[1])
	},
}

var repoEnableCmd = &cobra.Command{
	Use:   "enable [workspace-name] [repo-url]",
	Short: "Enable analysis for a repository",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		handleRepoEnable(args[0], args[1], true)
	},
}

var repoDisableCmd = &cobra.Command{
	Use:   "disable [workspace-name] [repo-url]",
	Short: "Disable analysis for a repository",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		handleRepoEnable(args[0], args[1], false)
	},
}

func init() {
	repoAddCmd.Flags().String("module-path", "", "Go module path (e.g., github.com/org/repo)")

	repoCmd.AddCommand(repoAddCmd)
	repoCmd.AddCommand(repoListCmd)
	repoCmd.AddCommand(repoRemoveCmd)
	repoCmd.AddCommand(repoEnableCmd)
	repoCmd.AddCommand(repoDisableCmd)

	rootCmd.AddCommand(repoCmd)
}

func handleRepoAdd(workspaceName, repoURL, modulePath string) {
	dbURL := getDBURL()
	tenantID := getTenantIDString()
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

	provider := "github"
	if strings.Contains(repoURL, "gitlab") {
		provider = "gitlab"
	} else if strings.Contains(repoURL, "bitbucket") {
		provider = "bitbucket"
	}

	repo, err := st.AddRepository(ctx, wsID, provider, sanitizeGitURL(repoURL), "main", "go", modulePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to add repository: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Repository '%s' added to workspace '%s' (ID: %s)\n", repo.URL, workspaceName, repo.ID)
	if modulePath != "" {
		fmt.Printf("  📦 Module path: %s\n", modulePath)
	}
}

func handleRepoList(workspaceName string) {
	dbURL := getDBURL()
	tenantID := getTenantIDString()
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
		fmt.Fprintf(os.Stderr, "❌ Failed to list: %v\n", err)
		os.Exit(1)
	}
	if len(repos) == 0 {
		fmt.Printf("⚠️ No repositories in workspace '%s'.\n", workspaceName)
		fmt.Printf("  Add one with: garuda repo add %s <url>\n", workspaceName)
		return
	}

	fmt.Printf("📂 Repositories in workspace '%s':\n", workspaceName)
	for _, r := range repos {
		commit := "N/A"
		if r.CurrentCommit != nil {
			commit = *r.CurrentCommit
		}
		fmt.Printf("  • %s (branch: %s, commit: %s, status: %s)\n", r.URL, r.DefaultBranch, commit, r.AnalysisStatus)
	}
}

func handleRepoRemove(workspaceName, repoURL string) {
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
	err = st.Pool().QueryRow(ctx, `SELECT id FROM workspaces WHERE tenant_id = $1 AND name = $2`, tenantID, workspaceName).Scan(&wsID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Workspace '%s' not found\n", workspaceName)
		os.Exit(1)
	}

	_, err = st.Pool().Exec(ctx, `DELETE FROM repositories WHERE workspace_id = $1 AND url = $2`, wsID, repoURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to remove: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Repository '%s' removed from workspace '%s'.\n", repoURL, workspaceName)
}

func handleRepoEnable(workspaceName, repoURL string, enable bool) {
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
	err = st.Pool().QueryRow(ctx, `SELECT id FROM workspaces WHERE tenant_id = $1 AND name = $2`, tenantID, workspaceName).Scan(&wsID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Workspace '%s' not found\n", workspaceName)
		os.Exit(1)
	}

	_, err = st.Pool().Exec(ctx, `UPDATE repositories SET enabled = $1 WHERE workspace_id = $2 AND url = $3`, enable, wsID, repoURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to update: %v\n", err)
		os.Exit(1)
	}

	state := "enabled"
	if !enable {
		state = "disabled"
	}
	fmt.Printf("✓ Repository '%s' %s.\n", repoURL, state)
}
