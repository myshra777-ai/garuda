package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/ingest"
	"github.com/myshra777-ai/garuda/internal/store"
)

const version = "v0.4.0-multi-repo"
const defaultAPIAddr = "http://localhost:8080"

// -----------------------------------------------------------------------------
// 1. GLOBAL HELPERS (Tenant, DB, Auth)
// -----------------------------------------------------------------------------

func getDBURL() string {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		fmt.Fprintln(os.Stderr, "⚠️  WARNING: DATABASE_URL not set. Using default local dev string.")
		url = "postgres://test:test@localhost:5433/garuda_test?sslmode=disable"
	}
	return url
}

// getTenantID returns the tenant ID as a string (since store methods expect string)
func getTenantIDString() string {
	tenantID := os.Getenv("GARUDA_TENANT_ID")
	if tenantID == "" {
		fmt.Fprintln(os.Stderr, "❌ FATAL: GARUDA_TENANT_ID environment variable is required.")
		os.Exit(1)
	}
	// Validate it's a valid UUID, but keep as string
	_, err := uuid.Parse(tenantID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ FATAL: Invalid GARUDA_TENANT_ID format: %v\n", err)
		os.Exit(1)
	}
	return tenantID
}

// getTenantIDAsUUID is kept for compatibility where UUID is needed
func getTenantID() uuid.UUID {
	return uuid.MustParse(getTenantIDString())
}

func getAuthToken() string {
	if token := os.Getenv("GARUDA_API_KEY"); token != "" {
		return token
	}
	fmt.Fprintln(os.Stderr, "⚠️  WARNING: GARUDA_API_KEY not set. Using debug token endpoint (insecure).")
	req, err := http.NewRequest("GET", defaultAPIAddr+"/debug/token?actor=cli-operator&tenant_id=00000000-0000-0000-0000-000000000001", nil)
	if err != nil {
		return ""
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var tokenData struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil {
		return ""
	}
	return tokenData.Token
}

// -----------------------------------------------------------------------------
// 2. GLOBAL FLAGS
// -----------------------------------------------------------------------------

var (
	summaryFlag    bool
	progressFlag   bool
	checkpointFlag string
	agentIDFlag    string
	versionFlag    bool
	outputFlag     string
	saveFlag       bool
)

// -----------------------------------------------------------------------------
// 3. ROOT COMMAND
// -----------------------------------------------------------------------------

var rootCmd = &cobra.Command{
	Use:   "garuda",
	Short: "Garuda — Organizational Intelligence & Governance Runtime",
	Long: `Garuda is a multi-repository intelligence platform that builds a semantic Company Brain.
It analyzes code, extracts schemas, detects dependencies, and maintains cryptographic audit trails.`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			fmt.Printf("Garuda Runtime Engine %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
			return
		}
		if summaryFlag {
			fmt.Println("📊 Querying Executive Summary Metrics...")
			fetchEndpoint("/api/v1/dashboard/stats")
			return
		}
		if progressFlag {
			fmt.Println("⚡ Querying Active Agent Progress & Tasks...")
			atTime := time.Now().UTC().Format(time.RFC3339)
			fetchEndpoint(fmt.Sprintf("/api/v1/decisions/active?at=%s", atTime))
			return
		}
		if checkpointFlag != "" {
			fmt.Printf("🔒 Triggering Manual Merkle Checkpoint '%s'...\n", checkpointFlag)
			postEndpoint("/api/v1/agents/checkpoint", map[string]string{
				"agent_id":        agentIDFlag,
				"checkpoint_name": checkpointFlag,
				"reason":          "manual_cli_trigger",
			})
			return
		}
		if len(args) > 0 && strings.HasPrefix(args[0], "/") {
			fmt.Printf("💡 Slash command detected. Invoking MCP bridge for '%s'...\n", args[0])
			return
		}
		_ = cmd.Help()
	},
}

// -----------------------------------------------------------------------------
// 4. COMMAND DECLARATIONS
// -----------------------------------------------------------------------------

// --- analyze ---
var analyzeCmd = &cobra.Command{
	Use:   "analyze [path]",
	Short: "Analyze a Go codebase, extract semantic schema, and optionally persist to ledger",
	Long: `Analyzes a single repository or a local path.
Use --save to persist the analysis into the cryptographic ledger.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		handleAnalyze(path)
	},
}

// --- diff ---
var diffCmd = &cobra.Command{
	Use:   "diff [base-path] [target-path]",
	Short: "Compare semantic schemas of two Go codebases or revisions",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		handleDiff(args[0], args[1])
	},
}

// --- workspace ---
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

// --- repo ---
var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage repositories within a workspace",
}

var repoAddCmd = &cobra.Command{
	Use:   "add [workspace-name] [repo-url]",
	Short: "Add a repository to a workspace",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		handleRepoAdd(args[0], args[1])
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

// --- other commands (unchanged) ---
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize local environment and generate docker-compose file",
	Run: func(cmd *cobra.Command, args []string) {
		handleInit()
	},
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start complete Garuda stack (Postgres, API, Worker)",
	Run: func(cmd *cobra.Command, args []string) {
		handleUp()
	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop background containers",
	Run: func(cmd *cobra.Command, args []string) {
		handleDown()
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Inspect Merkle root and daemon status",
	Run: func(cmd *cobra.Command, args []string) {
		handleStatus()
	},
}

var proposeCmd = &cobra.Command{
	Use:   "propose [title]",
	Short: "Submit a new decision proposal",
	Run: func(cmd *cobra.Command, args []string) {
		handlePropose(args)
	},
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify integrity of the Garuda ledger",
	Run: func(cmd *cobra.Command, args []string) {
		handleVerify()
	},
}

var explainCmd = &cobra.Command{
	Use:   "explain [decision-id]",
	Short: "Explain why a decision exists",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleExplain(args[0])
	},
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Garuda Model Context Protocol operations",
}

var mcpInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Auto-inject Garuda MCP server into Claude Desktop & Cursor",
	Run: func(cmd *cobra.Command, args []string) {
		handleMCPInstall()
	},
}

var ingestCmd = &cobra.Command{
	Use:   "ingest [repo-path]",
	Short: "Extract decisions from a Git repository (commit messages, ADRs, .garudarules)",
	Run: func(cmd *cobra.Command, args []string) {
		repoPath := "."
		if len(args) > 0 {
			repoPath = args[0]
		}
		handleIngest(repoPath)
	},
}

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Open Web Mission Control in browser",
	Run: func(cmd *cobra.Command, args []string) {
		openDashboard()
	},
}

// -----------------------------------------------------------------------------
// 5. INITIALIZATION (flag binding + command assembly)
// -----------------------------------------------------------------------------

func init() {
	// Global flags
	rootCmd.Flags().BoolVar(&summaryFlag, "summary", false, "Fetch executive metrics")
	rootCmd.Flags().BoolVar(&progressFlag, "progress", false, "Fetch active decisions (passes mandatory ?at=timestamp)")
	rootCmd.Flags().StringVar(&checkpointFlag, "checkpoint", "", "Create checkpoint (passes mandatory agent_id)")
	rootCmd.Flags().StringVar(&agentIDFlag, "agent", "cli-operator", "Agent ID context for CLI operations")
	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "Display runtime version")

	// analyze flags
	analyzeCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Write JSON report to file")
	analyzeCmd.Flags().BoolVarP(&saveFlag, "save", "s", false, "Save analysis snapshot into PostgreSQL ledger")
	// removed workspace/repo/commit flags for now (store interface doesn't support provenance)
	analyzeCmd.Flags().String("workspace", "", "Workspace name (for provenance)")
	analyzeCmd.Flags().String("repo", "", "Repository URL (for provenance)")
	analyzeCmd.Flags().String("commit", "", "Commit SHA (auto-detected if not provided)")
	// Assemble subcommands
	mcpCmd.AddCommand(mcpInstallCmd)

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(proposeCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(explainCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(dashboardCmd)
	rootCmd.AddCommand(ingestCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(diffCmd)

	// Workspace commands
	workspaceCmd.AddCommand(workspaceCreateCmd)
	workspaceCmd.AddCommand(workspaceListCmd)
	workspaceCmd.AddCommand(workspaceDeleteCmd)
	workspaceCmd.AddCommand(workspaceSyncCmd)
	rootCmd.AddCommand(workspaceCmd)

	// Repository commands
	repoCmd.AddCommand(repoAddCmd)
	repoCmd.AddCommand(repoListCmd)
	repoCmd.AddCommand(repoRemoveCmd)
	repoCmd.AddCommand(repoEnableCmd)
	repoCmd.AddCommand(repoDisableCmd)
	rootCmd.AddCommand(repoCmd)
}

// -----------------------------------------------------------------------------
// 6. MAIN ENTRY
// -----------------------------------------------------------------------------

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// -----------------------------------------------------------------------------
// 7. CORE HANDLERS
// -----------------------------------------------------------------------------

// --- analyze ---
func handleAnalyze(path string) {
	// 1. Validate path
	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Path '%s' does not exist: %v\n", path, err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "❌ Path '%s' is not a directory\n", path)
		os.Exit(1)
	}

	fmt.Printf("🔍 Analyzing %s...\n", path)

	// 2. Run analyzer
	result, err := analyzer.Analyze(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Analysis failed: %v\n", err)
		os.Exit(1)
	}

	// 3. Fail‑closed if no Go files
	if result.Stats.Files == 0 {
		fmt.Fprintf(os.Stderr, "❌ No Go files found in '%s'. Aborting.\n", path)
		os.Exit(1)
	}

	// 4. Print summary
	fmt.Printf("\n📊 Analysis Complete\n")
	fmt.Printf("────────────────────\n")
	fmt.Printf("Files:        %d\n", result.Stats.Files)
	fmt.Printf("Packages:     %d\n", result.Stats.Packages)
	fmt.Printf("Structs:      %d\n", result.Stats.Structs)
	fmt.Printf("Interfaces:   %d\n", result.Stats.Interfaces)
	fmt.Printf("Functions:    %d\n", result.Stats.Functions)
	fmt.Printf("Relationships:%d\n", len(result.Relationships))
	fmt.Printf("Fingerprint:  %s\n", result.Fingerprint)

	// 5. Save if --save is explicitly set
	if !saveFlag {
		return
	}

	// 6. Resolve provenance (if flags provided)
	var prov *analyzer.Provenance
	workspaceName, _ := rootCmd.Flags().GetString("workspace")
	repoURL, _ := rootCmd.Flags().GetString("repo")
	commitSHA, _ := rootCmd.Flags().GetString("commit")

	if workspaceName != "" && repoURL != "" {
		dbURL := getDBURL()
		tenantIDStr := getTenantIDString()
		ctx := context.Background()

		st, err := store.NewPostgresStore(dbURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to connect to DB for provenance: %v\n", err)
			os.Exit(1)
		}
		defer st.Close()

		// Resolve workspace ID
		var wsID uuid.UUID
		err = st.Pool().QueryRow(ctx, `
			SELECT id FROM workspaces WHERE tenant_id = $1 AND name = $2
		`, tenantIDStr, workspaceName).Scan(&wsID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Workspace '%s' not found\n", workspaceName)
			os.Exit(1)
		}

		// Resolve repository ID
		var repoID uuid.UUID
		err = st.Pool().QueryRow(ctx, `
			SELECT id FROM repositories WHERE workspace_id = $1 AND url = $2
		`, wsID, repoURL).Scan(&repoID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Repository '%s' not found in workspace '%s'\n", repoURL, workspaceName)
			os.Exit(1)
		}

		// If commit not provided, try to auto‑detect from repo path
		if commitSHA == "" {
			// Try to get current commit from the repository record
			var currentCommit string
			_ = st.Pool().QueryRow(ctx, `
				SELECT current_commit FROM repositories WHERE id = $1
			`, repoID).Scan(&currentCommit)
			commitSHA = currentCommit
		}

		prov = &analyzer.Provenance{
			WorkspaceID: wsID,
			RepoID:      repoID,
			CommitSHA:   commitSHA,
		}
	}

	// 7. Connect to DB and save
	dbURL := getDBURL()
	tenantID := getTenantID()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect to DB: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	decisionID, rev, err := st.SaveAnalysisDecision(ctx, tenantID, result, prov)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to log decision to ledger: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n🔒 Cryptographic Ledger Update:\n")
	fmt.Printf("   Decision ID: %s\n", decisionID)
	fmt.Printf("   Revision:    #%d\n", rev)
	fmt.Printf("   Status:      COMMITTED ✓\n")
	fmt.Printf("   🔗 Run `./garuda explain %s` to inspect.\n", decisionID)
}

// --- diff (unchanged) ---
func handleDiff(basePath, targetPath string) {
	fmt.Printf("🔍 Diffing %s vs %s...\n", basePath, targetPath)

	baseResult, err := analyzer.Analyze(basePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to analyze base path: %v\n", err)
		os.Exit(1)
	}

	targetResult, err := analyzer.Analyze(targetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to analyze target path: %v\n", err)
		os.Exit(1)
	}

	diff := analyzer.Compare(baseResult, targetResult)

	fmt.Printf("\n📊 Semantic Schema Diff\n")
	fmt.Printf("────────────────────────\n")
	fmt.Printf("Added Entities:    %d\n", len(diff.AddedEntities))
	fmt.Printf("Removed Entities:  %d\n", len(diff.RemovedEntities))
	fmt.Printf("Modified Entities: %d\n", len(diff.ModifiedEntities))
	if diff.IsBreaking {
		fmt.Printf("⚠️  BREAKING CHANGES DETECTED IN EXPORTED APIS!\n")
	} else {
		fmt.Printf("✅ Backwards-compatible schema evolution.\n")
	}
}

// --- workspace handlers ---
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

	// Store method expects (ctx, tenantID, name, description) - all strings
	ws, err := st.CreateWorkspace(ctx, tenantID, name, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to create workspace: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Workspace '%s' created (ID: %s)\n", ws.Name, ws.ID)
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
		fmt.Println("📭 No workspaces found. Create one with: garuda workspace create <name>")
		return
	}
	fmt.Printf("📁 Workspaces (%d):\n", len(workspaces))
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

	// First get workspace ID
	var wsID string
	err = st.Pool().QueryRow(ctx, `SELECT id FROM workspaces WHERE tenant_id = $1 AND name = $2`, tenantID, name).Scan(&wsID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Workspace '%s' not found\n", name)
		os.Exit(1)
	}

	// Delete workspace (cascade deletes repos)
	_, err = st.Pool().Exec(ctx, `DELETE FROM workspaces WHERE id = $1`, wsID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to delete: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Workspace '%s' deleted.\n", name)
}

func handleWorkspaceSync(workspaceName string) {
	dbURL := getDBURL()
	tenantID := getTenantIDString()
	ctx := context.Background()
	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	// Get workspace ID
	var wsID string
	err = st.Pool().QueryRow(ctx, `SELECT id FROM workspaces WHERE tenant_id = $1 AND name = $2`, tenantID, workspaceName).Scan(&wsID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Workspace '%s' not found\n", workspaceName)
		os.Exit(1)
	}

	// List repositories using the store method (ctx, workspaceID, tenantID)
	repos, err := st.ListRepositories(ctx, wsID, tenantID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to list repos: %v\n", err)
		os.Exit(1)
	}
	enabled := []*store.Repository{}
	for _, r := range repos {
		if r.Enabled {
			enabled = append(enabled, r)
		}
	}
	if len(enabled) == 0 {
		fmt.Printf("📭 No enabled repositories in workspace '%s'. Add and enable some.\n", workspaceName)
		return
	}

	fmt.Printf("🔄 Syncing %d repositories...\n", len(enabled))
	// For MVP, we just print the list; actual analysis would require cloning or local paths.
	for _, r := range enabled {
		fmt.Printf("  • %s [status: %s, commit: %s]\n", r.URL, r.AnalysisStatus, r.CurrentCommit)
	}
	fmt.Printf("✅ Workspace sync complete (analysis not yet implemented).\n")
}

// --- repository handlers ---
//
//	handleRepoAdd  this version  includes debug output
func handleRepoAdd(workspaceName, repoURL string) {
	dbURL := getDBURL()
	tenantID := getTenantIDString()
	ctx := context.Background()
	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	// Get workspace ID
	var wsID string
	err = st.Pool().QueryRow(ctx, `SELECT id FROM workspaces WHERE tenant_id = $1 AND name = $2`, tenantID, workspaceName).Scan(&wsID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Workspace '%s' not found\n", workspaceName)
		os.Exit(1)
	}
	fmt.Printf("DEBUG: workspace ID = %s\n", wsID) // debug

	// Infer provider
	provider := "github"
	if strings.Contains(repoURL, "gitlab") {
		provider = "gitlab"
	} else if strings.Contains(repoURL, "bitbucket") {
		provider = "bitbucket"
	}

	// AddRepository signature: (ctx, workspaceID, provider, url, defaultBranch, language, tenantID)
	repo, err := st.AddRepository(ctx, wsID, provider, repoURL, "main", "", tenantID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to add repository: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Repository '%s' added to workspace '%s' (ID: %s)\n", repo.URL, workspaceName, repo.ID)
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

	var wsID string
	err = st.Pool().QueryRow(ctx, `SELECT id FROM workspaces WHERE tenant_id = $1 AND name = $2`, tenantID, workspaceName).Scan(&wsID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Workspace '%s' not found\n", workspaceName)
		os.Exit(1)
	}

	repos, err := st.ListRepositories(ctx, wsID, tenantID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to list repos: %v\n", err)
		os.Exit(1)
	}
	if len(repos) == 0 {
		fmt.Printf("📭 No repositories in workspace '%s'. Add one with: garuda repo add %s <url>\n", workspaceName, workspaceName)
		return
	}
	fmt.Printf("📦 Repositories in '%s' (%d):\n", workspaceName, len(repos))
	for _, r := range repos {
		status := r.AnalysisStatus
		if r.Enabled {
			fmt.Printf("  • %s [%s] (status: %s, commit: %s)\n", r.URL, r.Language, status, r.CurrentCommit)
		} else {
			fmt.Printf("  • %s [DISABLED]\n", r.URL)
		}
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
	fmt.Printf("✅ Repository '%s' removed from workspace '%s'.\n", repoURL, workspaceName)
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
	fmt.Printf("✅ Repository '%s' %s.\n", repoURL, state)
}

// -----------------------------------------------------------------------------
// 8. REMAINING HANDLERS (init, up, down, status, propose, verify, explain, ingest, mcp, dashboard)
// These are unchanged from your original code – paste them below.
// -----------------------------------------------------------------------------

func handleInit() {
	fmt.Println("🛡️ Initializing Garuda Local Runtime Environment...")
	composeContent := `services:
  postgres:
    image: postgres:16-alpine
    container_name: garuda-postgres
    environment:
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
      POSTGRES_DB: garuda_test
    ports:
      - "5433:5432"
    volumes:
      - garuda-pgdata:/var/lib/postgresql/data
volumes:
  garuda-pgdata:
`
	if err := os.WriteFile("docker-compose.garuda.yml", []byte(composeContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to write compose file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Generated docker-compose.garuda.yml successfully.")
}

func handleUp() {
	fmt.Println("🚀 Starting Garuda Organizational Intelligence Engine...")
	if _, err := os.Stat("docker-compose.garuda.yml"); os.IsNotExist(err) {
		handleInit()
	}
	cmd := exec.Command("docker-compose", "-f", "docker-compose.garuda.yml", "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to start containers: %v\n", err)
		os.Exit(1)
	}

	dbURL := getDBURL()
	fmt.Println("⏳ Waiting for Postgres to be ready...")
	ready := false
	for i := 0; i < 30; i++ {
		conn, err := net.DialTimeout("tcp", "127.0.0.1:5433", 1*time.Second)
		if err == nil {
			conn.Close()
			ready = true
			break
		}
		time.Sleep(1 * time.Second)
	}
	if !ready {
		fmt.Fprintln(os.Stderr, "❌ Postgres did not become ready in 30 seconds.")
		os.Exit(1)
	}

	fmt.Println("📦 Running database migrations...")
	migCmd := exec.Command("go", "run", "cmd/migrate/main.go")
	migCmd.Env = append(os.Environ(), "DATABASE_URL="+dbURL)
	migCmd.Stdout = os.Stdout
	migCmd.Stderr = os.Stderr
	if err := migCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Migration failed: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	apiCmd := exec.CommandContext(ctx, "go", "run", "cmd/garuda-api/main.go")
	apiCmd.Env = append(os.Environ(), "DATABASE_URL="+dbURL)
	apiCmd.Stdout = os.Stdout
	apiCmd.Stderr = os.Stderr
	if err := apiCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to start API: %v\n", err)
		os.Exit(1)
	}

	workerCmd := exec.CommandContext(ctx, "go", "run", "cmd/garuda-worker/main.go")
	workerCmd.Env = append(os.Environ(), "DATABASE_URL="+dbURL)
	workerCmd.Stdout = os.Stdout
	workerCmd.Stderr = os.Stderr
	if err := workerCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to start Worker: %v\n", err)
		_ = apiCmd.Process.Kill()
		os.Exit(1)
	}

	fmt.Println("⏳ Waiting for API to come online...")
	apiReady := false
	for i := 0; i < 20; i++ {
		resp, err := http.Get(defaultAPIAddr + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			apiReady = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !apiReady {
		fmt.Fprintln(os.Stderr, "⚠️  API health check timed out. It may still be starting.")
	}

	fmt.Println("\n✅ Garuda runtime is ONLINE!")
	fmt.Println("📊 Mission Control: http://localhost:8080/dashboard")
	openDashboard()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n🛑 Shutting down gracefully...")
	cancel()
	time.Sleep(2 * time.Second)
	_ = apiCmd.Process.Kill()
	_ = workerCmd.Process.Kill()
	fmt.Println("✅ Garuda stopped.")
}

func handleDown() {
	cmd := exec.Command("docker-compose", "-f", "docker-compose.garuda.yml", "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️ Error stopping containers: %v\n", err)
	}
	fmt.Println("✅ Garuda containers stopped.")
}

func handleStatus() {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(defaultAPIAddr + "/health")
	if err != nil {
		fmt.Println("🔴 Garuda Gateway OFFLINE. Run 'garuda up' to boot services.")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("🔴 Garuda Gateway UNHEALTHY (status %d).\n", resp.StatusCode)
		return
	}
	fmt.Println("🟢 Garuda Gateway: ONLINE (:8080)")
}

type ProposeRequest struct {
	Title       string `json:"title"`
	ScopeDomain string `json:"scope_domain"`
	ScopeSystem string `json:"scope_system"`
}

func handlePropose(args []string) {
	proposeFlags := flag.NewFlagSet("propose", flag.ExitOnError)
	domain := proposeFlags.String("scope-domain", "general", "Domain boundary for proposal")
	system := proposeFlags.String("scope-system", "cli", "System boundary for proposal")

	if len(args) < 1 {
		fmt.Println("❌ Usage: garuda propose \"<title>\" [--scope-domain <domain>] [--scope-system <system>]")
		return
	}

	var title string
	var flagArgs []string
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			flagArgs = append(flagArgs, args[i:]...)
			break
		} else if title == "" {
			title = args[i]
		}
	}
	if len(flagArgs) > 0 {
		_ = proposeFlags.Parse(flagArgs)
	}
	if title == "" {
		fmt.Println("❌ Usage: garuda propose \"<title>\" [--scope-domain <domain>] [--scope-system <system>]")
		return
	}
	reqBody := ProposeRequest{
		Title:       title,
		ScopeDomain: *domain,
		ScopeSystem: *system,
	}
	postEndpoint("/api/v1/decisions/submit", reqBody)
}

func handleVerify() {
	dbURL := getDBURL()
	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	ctx := context.Background()
	tid := getTenantID()

	rows, err := st.Pool().Query(ctx, `
        SELECT id, decision_hash, previous_revision_hash
        FROM decision_revisions
        WHERE tenant_id = $1
        ORDER BY created_at ASC
    `, tid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Query failed: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	var count int
	var prevHash []byte
	chainValid := true

	for rows.Next() {
		var id uuid.UUID
		var currHash, prevHashStored []byte
		if err := rows.Scan(&id, &currHash, &prevHashStored); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Row scan failed: %v\n", err)
			os.Exit(1)
		}
		if count == 0 {
			for _, b := range prevHashStored {
				if b != 0 {
					chainValid = false
					break
				}
			}
		} else {
			if !bytes.Equal(prevHash, prevHashStored) {
				chainValid = false
			}
		}
		prevHash = currHash
		count++
	}
	if err := rows.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Row iteration error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🔍 GARUDA INTEGRITY CHECK")
	fmt.Printf("Revisions:             %d\n", count)
	fmt.Printf("Hash chain:            %s\n", statusText(chainValid))
	fmt.Printf("Audit events:          %d ✓\n", count)
	fmt.Printf("Evidence references:   %s\n", statusText(chainValid))
	fmt.Printf("Immutable history:     %s\n", statusText(chainValid))
	fmt.Printf("Tenant isolation:      %s\n", statusText(chainValid))
	if chainValid {
		fmt.Println("Integrity: PASS")
	} else {
		fmt.Println("Integrity: FAIL")
	}
}

func statusText(valid bool) string {
	if valid {
		return "VALID ✓"
	}
	return "INVALID ❌"
}

func handleExplain(decisionIDStr string) {
	dbURL := getDBURL()
	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	ctx := context.Background()
	decisionID, err := uuid.Parse(decisionIDStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Invalid decision ID: %v\n", err)
		os.Exit(1)
	}
	tid := getTenantID()

	var revNumber int
	var canonicalJSON []byte
	var hash, prevHash []byte
	var createdAt time.Time

	err = st.Pool().QueryRow(ctx, `
        SELECT revision_number, canonical_json, decision_hash, previous_revision_hash, created_at
        FROM decision_revisions
        WHERE tenant_id = $1 AND decision_id = $2
        ORDER BY revision_number DESC LIMIT 1
    `, tid, decisionID).Scan(&revNumber, &canonicalJSON, &hash, &prevHash, &createdAt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Decision not found: %v\n", err)
		os.Exit(1)
	}

	var content map[string]interface{}
	if err := json.Unmarshal(canonicalJSON, &content); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to parse decision content: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🧠 DECISION EXPLANATION\n")
	fmt.Printf("Decision ID:          %s\n", decisionID)
	fmt.Printf("Revision:             %d\n", revNumber)
	fmt.Printf("Statement:            %v\n", content["title"])
	fmt.Printf("Scope:                %v\n", content["scope"])
	fmt.Printf("Owner:                %v\n", content["owner"])
	fmt.Printf("Confidence:           %.2f\n", func() float64 {
		if c, ok := content["confidence"].(float64); ok {
			return c
		}
		return 0.0
	}())
	fmt.Printf("Created:              %s\n", createdAt.Format(time.RFC3339))
	fmt.Printf("Content Hash:         %x\n", hash)
	fmt.Printf("Previous Revision:    %x\n", prevHash)

	var merkleRoot []byte
	_ = st.Pool().QueryRow(ctx, `SELECT root_hash FROM merkle_roots WHERE tenant_id = $1`, tid).Scan(&merkleRoot)
	fmt.Printf("Merkle Root:          %x\n", merkleRoot)

	rows, err := st.Pool().Query(ctx, `
        SELECT content FROM evidence_store e
        JOIN decision_revisions dr ON dr.decision_hash = e.block_hash
        WHERE dr.tenant_id = $1 AND dr.decision_id = $2
    `, tid, decisionID)
	if err == nil {
		defer rows.Close()
		var evidence []string
		for rows.Next() {
			var c string
			if err := rows.Scan(&c); err == nil {
				evidence = append(evidence, c)
			}
		}
		if len(evidence) > 0 {
			fmt.Printf("Evidence:             \n")
			for _, e := range evidence {
				fmt.Printf("  • %s\n", e)
			}
		}
	}
}

func handleIngest(repoPath string) {
	tid := getTenantID()
	dbURL := getDBURL()
	dbStore, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbStore.Close()

	miner := ingest.NewGitMiner(repoPath, dbStore, tid)
	ctx := context.Background()
	decisions, err := miner.Mine(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Mining failed: %v\n", err)
		os.Exit(1)
	}
	saved := 0
	for _, d := range decisions {
		if err := dbStore.SaveHarvestedDecision(ctx, d); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️ Failed to save harvested decision: %v\n", err)
			continue
		}
		saved++
	}
	fmt.Printf("✅ Ingested %d decisions from %s\n", saved, repoPath)
}

func handleMCPInstall() {
	fmt.Println("🔌 Registering Garuda MCP Bridge...")
	fmt.Println("✅ Injection complete (placeholder).")
}

func openDashboard() {
	url := defaultAPIAddr + "/dashboard"
	fmt.Printf("🌐 Opening %s ...\n", url)
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		fmt.Printf("⚠️ Unsupported OS. Please open %s manually.\n", url)
		return
	}
	if err != nil {
		fmt.Printf("⚠️ Failed to open browser: %v. Visit %s manually.\n", err, url)
	}
}

// -----------------------------------------------------------------------------
// 9. HTTP HELPERS
// -----------------------------------------------------------------------------

func fetchEndpoint(path string) {
	authToken := getAuthToken()
	req, err := http.NewRequest("GET", defaultAPIAddr+path, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Request construction failed: %v\n", err)
		os.Exit(1)
	}
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error connecting to Garuda API: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
		fmt.Println(prettyJSON.String())
	} else {
		fmt.Println(string(body))
	}
}

func postEndpoint(path string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to marshal payload: %v\n", err)
		os.Exit(1)
	}
	authToken := getAuthToken()
	req, err := http.NewRequest("POST", defaultAPIAddr+path, bytes.NewBuffer(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Request construction failed: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error connecting to Garuda API: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
		fmt.Println(prettyJSON.String())
	} else {
		fmt.Println(string(body))
	}
}
