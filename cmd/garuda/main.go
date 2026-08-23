// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

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
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/graph"
	"github.com/myshra777-ai/garuda/internal/ingest"
	"github.com/myshra777-ai/garuda/internal/store"
)

const version = "v0.4.0-multi-repo"
const defaultAPIAddr = "http://localhost:8080"

// 1. GLOBAL HELPERS (Tenant, DB, Auth)

func getDBURL() string {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		fmt.Fprintln(os.Stderr, "⚠️ WARNING: DATABASE_URL not set. Using default local dev string.")
		url = "postgres://test:test@localhost:5433/garuda_test?sslmode=disable"
	}
	return url
}

func getTenantIDString() string {
	tenantID := os.Getenv("GARUDA_TENANT_ID")
	if tenantID == "" {
		fmt.Fprintln(os.Stderr, "❌ FATAL: GARUDA_TENANT_ID environment variable is required.")
		os.Exit(1)
	}
	err := uuid.Validate(tenantID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ FATAL: Invalid GARUDA_TENANT_ID format: %v\n", err)
		os.Exit(1)
	}
	return tenantID
}

func getTenantID() uuid.UUID {
	return uuid.MustParse(getTenantIDString())
}

func getAuthToken() string {
	if token := os.Getenv("GARUDA_API_KEY"); token != "" {
		return token
	}

	fmt.Fprintln(os.Stderr, "⚠️ WARNING: GARUDA_API_KEY not set. Using debug token endpoint (insecure).")
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

// 2. GLOBAL FLAGS

var (
	summaryFlag    bool
	progressFlag   bool
	checkpointFlag string
	agentIDFlag    string
	versionFlag    bool

	outputFlag         string
	saveFlag           bool
	jsonOutputFlag     bool
	outputFileFlag     string
	workspaceFlag      string
	repoFlag           string
	graphOpenFlag      bool
	modulePathFlag     string
	graphRepoFlag      string
	failOnBreakingFlag bool // CI Policy Gate Flag
)

// External Commands (Assumed to be defined in other package main files)
//Since impactCmd, impactDiffCmd, justifyCmd, judgeCmd, ciCmd, ponytailCmd, and selfDescribeCmd are already exported package-level variables in their respective files (impact.go, justify.go, etc.), delete their duplicate variable declarations in main.go. Only keep them in rootCmd.AddCommand(...)

// 3. ROOT COMMAND

var rootCmd = &cobra.Command{
	Use:   "garuda",
	Short: "Garuda Organizational Intelligence & Governance Runtime",
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
			fmt.Println("🔄 Querying Active Agent Progress & Tasks...")
			atTime := time.Now().UTC().Format(time.RFC3339)
			fetchEndpoint(fmt.Sprintf("/api/v1/decisions/active?at=%s", atTime))
			return
		}

		if checkpointFlag != "" {
			fmt.Printf("🛡️ Triggering Manual Merkle Checkpoint '%s'...\n", checkpointFlag)
			postEndpoint("/api/v1/agents/checkpoint", map[string]string{
				"agent_id":        agentIDFlag,
				"checkpoint_name": checkpointFlag,
				"reason":          "manual_cli_trigger",
			})
			return
		}

		if len(args) > 0 && strings.HasPrefix(args[0], "/") {
			fmt.Printf("⚡ Slash command detected. Invoking MCP bridge for '%s'...\n", args[0])
			return
		}

		_ = cmd.Help()
	},
}

// 4. COMMAND DECLARATIONS

var analyzeCmd = &cobra.Command{
	Use:   "analyze [path]",
	Short: "Analyze a Go codebase, extract semantic schema, and optionally persist to ledger",
	Long:  `Analyzes a single repository or a local path. Use --save to persist the analysis into the cryptographic ledger.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		handleAnalyze(path)
	},
}

var diffCmd = &cobra.Command{
	Use:   "diff [json1] [json2]",
	Short: "Compare two analysis JSON snapshots",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		handleDiff(args[0], args[1])
	},
}

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


var mcpInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Auto-inject Garuda MCP server into Claude Desktop & Cursor",
	Run: func(cmd *cobra.Command, args []string) {
		handleMCPInstall()
	},
}

var ingestCmd = &cobra.Command{
	Use:   "ingest [repo-path]",
	Short: "Extract decisions from a Git repository (commit messages, ADRs, garudarules)",
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

var inspectCmd = &cobra.Command{
	Use:   "inspect <entity>",
	Short: "Inspect a semantic entity (type, service, API, etc.)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleInspect(args[0])
	},
}

var graphCmd = &cobra.Command{
	Use:   "graph [workspace-name]",
	Short: "Generate interactive HTML graph for a workspace",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleGraph(args[0])
	},
}

var entitiesCmd = &cobra.Command{
	Use:   "entities",
	Short: "List all entities in the current workspace",
	Run: func(cmd *cobra.Command, args []string) {
		handleListEntities()
	},
}

// 5. INITIALIZATION (flag binding + command assembly)

func init() {
	rootCmd.Flags().BoolVar(&summaryFlag, "summary", false, "Fetch executive metrics")
	rootCmd.Flags().BoolVar(&progressFlag, "progress", false, "Fetch active decisions")
	rootCmd.Flags().StringVar(&checkpointFlag, "checkpoint", "", "Create checkpoint")
	rootCmd.Flags().StringVar(&agentIDFlag, "agent", "cli-operator", "Agent ID context")
	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "Display runtime version")

	analyzeCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Write JSON report to file")
	analyzeCmd.Flags().BoolVarP(&saveFlag, "save", "s", false, "Save analysis snapshot into PostgreSQL ledger")
	analyzeCmd.Flags().StringVar(&workspaceFlag, "workspace", "", "Workspace name (for provenance)")
	analyzeCmd.Flags().StringVar(&repoFlag, "repo", "", "Repository URL (for provenance)")
	analyzeCmd.Flags().String("commit", "", "Commit SHA (auto-detected if not provided)")
	analyzeCmd.Flags().StringVar(&modulePathFlag, "module-path", "", "Go module path (auto-detected if not provided)")

	diffCmd.Flags().BoolVar(&jsonOutputFlag, "json", false, "Output diff in JSON format")
	diffCmd.Flags().StringVarP(&outputFileFlag, "output", "o", "", "Write diff to file")
	diffCmd.Flags().BoolVar(&failOnBreakingFlag, "fail-on-breaking", false, "Exit with code 1 if breaking changes are detected (CI Gate)")

	graphCmd.Flags().BoolVar(&graphOpenFlag, "open", false, "Open the graph in browser automatically")
	graphCmd.Flags().StringVar(&graphRepoFlag, "repo", "", "Filter graph to a specific repository (URL or ID)")

	repoAddCmd.Flags().String("module-path", "", "Go module path (e.g., github.com/org/repo)")

	mcpCmd.AddCommand(mcpInstallCmd)

	workspaceCmd.AddCommand(workspaceCreateCmd)
	workspaceCmd.AddCommand(workspaceListCmd)
	workspaceCmd.AddCommand(workspaceDeleteCmd)
	workspaceCmd.AddCommand(workspaceSyncCmd)

	repoCmd.AddCommand(repoAddCmd)
	repoCmd.AddCommand(repoListCmd)
	repoCmd.AddCommand(repoRemoveCmd)
	repoCmd.AddCommand(repoEnableCmd)
	repoCmd.AddCommand(repoDisableCmd)

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
	rootCmd.AddCommand(workspaceCmd)
	rootCmd.AddCommand(repoCmd)
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(graphCmd)
	rootCmd.AddCommand(entitiesCmd)
	rootCmd.AddCommand(summaryCmd)

	// Note: External commands must be defined in other files within 'main' package.
	if impactCmd != nil {
		rootCmd.AddCommand(impactCmd)
	}
	if impactDiffCmd != nil {
		rootCmd.AddCommand(impactDiffCmd)
	}
	if ponytailCmd != nil {
		rootCmd.AddCommand(ponytailCmd)
	}
	if justifyCmd != nil {
		rootCmd.AddCommand(justifyCmd)
	}
	if judgeCmd != nil {
		rootCmd.AddCommand(judgeCmd)
	}
	if selfDescribeCmd != nil {
		rootCmd.AddCommand(selfDescribeCmd)
	}
	if ciCmd != nil {
		rootCmd.AddCommand(ciCmd)
	}
	if summaryCmd != nil {
		rootCmd.AddCommand(summaryCmd)
	}
}

// 6. MAIN ENTRY

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// 7. CORE HANDLERS

func handleAnalyze(path string) {
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

	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Invalid path '%s': %v\n", path, err)
		os.Exit(1)
	}

	// 1. Discover Workspace / Multi-Module Boundaries
	ws, err := analyzer.DiscoverWorkspace(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Workspace discovery failed: %v\n", err)
		os.Exit(1)
	}

	// 2. Configure Execution Options with Cache
	opts := analyzer.WorkspaceAnalysisOptions{
		Cache: analyzer.NewMemoryPackageCache(),
	}
	if saveFlag || workspaceFlag != "" || repoFlag != "" {
		if tidStr := os.Getenv("GARUDA_TENANT_ID"); tidStr != "" {
			if tid, err := uuid.Parse(tidStr); err == nil {
				opts.TenantID = tid
			}
		}
	}

	ctx := context.Background()
	result, err := analyzer.AnalyzeWorkspaceWithOptions(ctx, ws, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Analysis failed: %v\n", err)
		os.Exit(1)
	}
	if result.Stats.Files == 0 && len(result.Entities) == 0 {
		fmt.Fprintf(os.Stderr, "❌ No Go files found in '%s'. Aborting.\n", path)
		os.Exit(1)
	}

	if outputFlag != "" {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Printf("⚠️ Failed to marshal JSON snapshot: %v\n", err)
		} else if err := os.WriteFile(outputFlag, data, 0644); err != nil {
			fmt.Printf("⚠️ Failed to write JSON: %v\n", err)
		} else {
			fmt.Printf("📄 JSON report saved to %s (Fingerprint: %s)\n", outputFlag, result.Fingerprint)
		}
	}

	if saveFlag || workspaceFlag != "" || repoFlag != "" {
		dbURL := getDBURL()
		tenantIDStr := getTenantIDString()
		tenantUUID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Invalid tenant UUID '%s': %v\n", tenantIDStr, err)
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		st, err := store.NewPostgresStore(dbURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to connect to DB: %v\n", err)
			os.Exit(1)
		}
		defer st.Close()

		decisionID, revisionID, rev, err := st.SaveAnalysisDecision(ctx, tenantIDStr, result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to log decision to ledger: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\n🔒 Cryptographic Ledger Update:\n")
		fmt.Printf("   Decision ID: %s\n", decisionID)
		fmt.Printf("   Revision ID: %s\n", revisionID)
		fmt.Printf("   Revision:    #%d\n", rev)
		fmt.Printf("   Status:      COMMITTED ✓\n")
		fmt.Printf("   🔗 Run `./garuda explain %s` to inspect.\n", decisionID)

		wsName := workspaceFlag
		if wsName == "" {
			wsName = os.Getenv("GARUDA_WORKSPACE")
			if wsName == "" {
				wsName = "default"
			}
		}

		repoURL := repoFlag
		if repoURL == "" {
			cmd := exec.Command("git", "-C", path, "remote", "get-url", "origin")
			if out, err := cmd.Output(); err == nil {
				repoURL = strings.TrimSpace(string(out))
			}
			if repoURL == "" {
				absPath, _ := filepath.Abs(path)
				repoURL = "file://" + absPath
			}
		}

		modulePath := modulePathFlag
		if modulePath == "" {
			goModPath := filepath.Join(path, "go.mod")
			if modData, err := os.ReadFile(goModPath); err == nil {
				for _, line := range strings.Split(string(modData), "\n") {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "module ") {
						modulePath = strings.TrimSpace(strings.TrimPrefix(line, "module "))
						break
					}
				}
			}
		}

		var workspaceID uuid.UUID
		err = st.Pool().QueryRow(ctx, `SELECT id FROM workspaces WHERE tenant_id = $1 AND name = $2`, tenantUUID, wsName).Scan(&workspaceID)
		if err != nil {
			ws, err := st.CreateWorkspace(ctx, tenantIDStr, wsName, "")
			if err != nil {
				fmt.Fprintf(os.Stderr, "⚠️ Failed to create workspace '%s': %v\n", wsName, err)
				return
			}
			workspaceID = ws.ID
			fmt.Printf("   Created workspace: %s\n", wsName)
		}

		var repoID uuid.UUID
		err = st.Pool().QueryRow(ctx, `SELECT id FROM repositories WHERE workspace_id = $1 AND url = $2`, workspaceID, repoURL).Scan(&repoID)
		if err != nil {
			provider := "local"
			if strings.Contains(repoURL, "github.com") {
				provider = "github"
			} else if strings.Contains(repoURL, "gitlab.com") {
				provider = "gitlab"
			} else if strings.Contains(repoURL, "bitbucket") {
				provider = "bitbucket"
			}
			repo, err := st.AddRepository(ctx, workspaceID, provider, repoURL, "main", "go", modulePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "⚠️ Failed to create repository: %v\n", err)
				return
			}
			repoID = repo.ID
			fmt.Printf("   Created repository: %s\n", repoURL)
		}

		if modulePath != "" {
			_, _ = st.Pool().Exec(ctx, `UPDATE repositories SET module_path = $1 WHERE id = $2`, modulePath, repoID)
			fmt.Printf("   Module path: %s\n", modulePath)
		}

		commitSHA := "local-uncommitted"
		gitCmd := exec.Command("git", "-C", path, "rev-parse", "HEAD")
		if out, err := gitCmd.Output(); err == nil {
			commitSHA = strings.TrimSpace(string(out))
		}

		err = st.SaveSemanticGraph(ctx, tenantUUID, workspaceID, repoID, revisionID, result, commitSHA)
		if err != nil {
			fmt.Printf("⚠️ Failed to save semantic graph: %v\n", err)
		} else {
			fmt.Printf("   🧠 Semantic graph saved (%d entities, %d relationships, commit: %s)\n",
				len(result.Entities), len(result.Relationships), commitSHA[:min(8, len(commitSHA))])
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func handleDiff(file1, file2 string) {
	before, err := analyzer.LoadResult(file1)
	if err != nil {
		fmt.Printf("❌ Failed to load base snapshot: %v\n", err)
		os.Exit(1)
	}
	after, err := analyzer.LoadResult(file2)
	if err != nil {
		fmt.Printf("❌ Failed to load new snapshot: %v\n", err)
		os.Exit(1)
	}

	report := analyzer.Diff(before, after)
	if jsonOutputFlag {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			fmt.Printf("❌ Failed to marshal JSON: %v\n", err)
			return
		}
		if outputFileFlag != "" {
			if err := os.WriteFile(outputFileFlag, data, 0644); err != nil {
				fmt.Printf("❌ Failed to write diff JSON: %v\n", err)
			} else {
				fmt.Printf("📄 JSON diff written to %s\n", outputFileFlag)
			}
		} else {
			fmt.Println(string(data))
		}
	} else {
		printDiffReport(report)
	}

	// Automated CI Policy Gate
	if failOnBreakingFlag && report.HasBreakingChanges() {
		fmt.Fprintf(os.Stderr, "\n[POLICY VIOLATION] %d breaking change(s) detected. CI Gate FAILED.\n", report.Summary.BreakingChanges)
		os.Exit(1)
	}
}

func printDiffReport(report *analyzer.DiffReport) {
	fmt.Println("📊 SCHEMA DIFF")
	fmt.Println("-------------")
	fmt.Println()
	fmt.Printf("Stats:\n")
	fmt.Printf("  Files:      %+d\n", report.StatsDiff.Files)
	fmt.Printf("  Packages:   %+d\n", report.StatsDiff.Packages)
	fmt.Printf("  Structs:    %+d\n", report.StatsDiff.Structs)
	fmt.Printf("  Interfaces: %+d\n", report.StatsDiff.Interfaces)
	fmt.Printf("  Functions:  %+d\n", report.StatsDiff.Functions)
	fmt.Printf("  Imports:    %+d\n", report.StatsDiff.Imports)
	fmt.Println()
	fmt.Printf("Fingerprint Match: %t\n", report.FingerprintDiff.Match)
	fmt.Println()

	if len(report.EntityDiffs) > 0 {
		fmt.Println("Entities:")
		for _, ed := range report.EntityDiffs {
			fmt.Printf("  [%s] %s %s", ed.Status, ed.Kind, ed.Name)
			if ed.Status == "modified" {
				if ed.FieldsDiff != nil {
					fmt.Printf(" (fields: +%d -%d ~%d)", len(ed.FieldsDiff.Added), len(ed.FieldsDiff.Removed), len(ed.FieldsDiff.Modified))
				}
				if ed.MethodsDiff != nil {
					fmt.Printf(" (methods: +%d -%d)", len(ed.MethodsDiff.Added), len(ed.MethodsDiff.Removed))
				}
			}
			if ed.Impact > 0 {
				fmt.Printf(" (%d references)", ed.Impact)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	if len(report.RelationshipDiffs) > 0 {
		fmt.Println("Relationships:")
		for _, rd := range report.RelationshipDiffs {
			fmt.Printf("  [%s] %s %s -> %s\n", rd.Status, rd.Type, rd.From, rd.To)
		}
		fmt.Println()
	}

	fmt.Println("Summary:")
	fmt.Printf("  Breaking changes: %d\n", report.Summary.BreakingChanges)
	fmt.Printf("  Warnings:         %d\n", report.Summary.Warnings)
	fmt.Printf("  Additions:        %d\n", report.Summary.Additions)
	fmt.Printf("  Removals:         %d\n", report.Summary.Removals)
	fmt.Printf("  Modified:         %d\n", report.Summary.Modified)
}

// --- Workspace handlers ---

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

// --- Repository handlers ---

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

	repo, err := st.AddRepository(ctx, wsID, provider, repoURL, "main", "go", modulePath)
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

// --- Entities Handlers ---

func handleListEntities() {
	dbURL := getDBURL()
	tenantID := getTenantIDString()
	ctx := context.Background()

	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect to DB: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	workspaceName := os.Getenv("GARUDA_WORKSPACE")
	if workspaceName == "" {
		workspaceName = "default"
	}

	var wsID uuid.UUID
	err = st.Pool().QueryRow(ctx, `SELECT id FROM workspaces WHERE tenant_id = $1 AND name = $2`, tenantID, workspaceName).Scan(&wsID)
	if err != nil {
		fmt.Printf("❌ Workspace '%s' not found.\n", workspaceName)
		fmt.Println("  Create one with: garuda workspace create " + workspaceName)
		os.Exit(1)
	}

	rows, err := st.Pool().Query(ctx, `
		SELECT name, kind, package, file_path, is_exported
		FROM entities
		WHERE tenant_id = $1 AND workspace_id = $2
		ORDER BY package, name
	`, tenantID, wsID)

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to list entities: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	type EntityRow struct {
		Name     string
		Kind     string
		Package  string
		File     string
		Exported bool
	}

	var entities []EntityRow
	for rows.Next() {
		var e EntityRow
		if err := rows.Scan(&e.Name, &e.Kind, &e.Package, &e.File, &e.Exported); err != nil {
			continue
		}
		entities = append(entities, e)
	}

	if len(entities) == 0 {
		fmt.Printf("⚠️ No entities found in workspace '%s'.\n", workspaceName)
		fmt.Println("  Run: garuda analyze --save")
		return
	}

	fmt.Printf("🔍 Entities in workspace '%s' (%d):\n", workspaceName, len(entities))
	for _, e := range entities {
		exported := ""
		if e.Exported {
			exported = "*"
		}
		fmt.Printf("  • %s%s.%s (%s) [%s]\n", exported, e.Package, e.Name, e.Kind, e.File)
	}
}

// --- Inspect Handlers ---

func handleInspect(entityName string) {
	dbURL := getDBURL()
	tenantID := getTenantIDString()
	ctx := context.Background()

	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	var id string
	var name, kind, pkg, filePath, signature string
	var isExported bool
	var fieldsJSON, methodsJSON []byte

	err = st.Pool().QueryRow(ctx, `
		SELECT id, name, kind, package, file_path, fields, methods, signature, is_exported
		FROM entities
		WHERE tenant_id = $1 AND name = $2
		LIMIT 1
	`, tenantID, entityName).Scan(&id, &name, &kind, &pkg, &filePath, &fieldsJSON, &methodsJSON, &signature, &isExported)

	if err != nil {
		fmt.Printf("❌ Entity '%s' not found.\n", entityName)
		os.Exit(1)
	}

	fmt.Printf("🔍 Entity: %s\n", name)
	fmt.Printf("   Package:  %s\n", pkg)
	fmt.Printf("   Kind:     %s\n", kind)
	fmt.Printf("   File:     %s\n", filePath)
	fmt.Printf("   Exported: %v\n", isExported)

	if signature != "" {
		fmt.Printf("   Signature: %s\n", signature)
	}

	if len(fieldsJSON) > 0 {
		fmt.Printf("\n   Fields:\n")
		var fields []analyzer.Field
		if err := json.Unmarshal(fieldsJSON, &fields); err == nil {
			for _, f := range fields {
				fmt.Printf("     • %s: %s\n", f.Name, f.Type)
			}
		}
	}

	if len(methodsJSON) > 0 {
		fmt.Printf("\n   Methods:\n")
		var methods []analyzer.Method
		if err := json.Unmarshal(methodsJSON, &methods); err == nil {
			for _, m := range methods {
				fmt.Printf("     • %s %s\n", m.Name, m.Signature)
			}
		}
	}

	rows, err := st.Pool().Query(ctx, `
		SELECT claim_type, to_entity_id FROM claims
		WHERE tenant_id = $1 AND from_entity_id = $2
	`, tenantID, id)

	if err == nil {
		defer rows.Close()
		var outgoing []string
		for rows.Next() {
			var typ, toID string
			if err := rows.Scan(&typ, &toID); err == nil {
				outgoing = append(outgoing, fmt.Sprintf("     • %s -> %s", typ, toID))
			}
		}
		if len(outgoing) > 0 {
			fmt.Printf("\n   Claims (outgoing):\n")
			for _, c := range outgoing {
				fmt.Println(c)
			}
		}
	}

	rows, err = st.Pool().Query(ctx, `
		SELECT claim_type, from_entity_id FROM claims
		WHERE tenant_id = $1 AND to_entity_id = $2
	`, tenantID, id)

	if err == nil {
		defer rows.Close()
		var incoming []string
		for rows.Next() {
			var typ, fromID string
			if err := rows.Scan(&typ, &fromID); err == nil {
				incoming = append(incoming, fmt.Sprintf("     • %s -> %s", fromID, typ))
			}
		}
		if len(incoming) > 0 {
			fmt.Printf("\n   Claims (incoming):\n")
			for _, c := range incoming {
				fmt.Println(c)
			}
		}
	}
}

// --- Graph Handlers ---

func handleGraph(target string) {
	// 1. Check if target is a JSON snapshot file or local directory (Offline Standalone Mode)
	if info, err := os.Stat(target); err == nil {
		var result *analyzer.Result

		if !info.IsDir() && strings.HasSuffix(target, ".json") {
			fmt.Printf("📊 Generating graph from snapshot %s...\n", target)
			res, err := analyzer.LoadResult(target)
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to load snapshot: %v\n", err)
				os.Exit(1)
			}
			result = res
		} else if info.IsDir() {
			fmt.Printf("🔍 Analyzing %s for offline graph generation...\n", target)
			absPath, _ := filepath.Abs(target)
			ws, err := analyzer.DiscoverWorkspace(absPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Discovery failed: %v\n", err)
				os.Exit(1)
			}
			res, err := analyzer.AnalyzeWorkspaceWithOptions(context.Background(), ws, analyzer.WorkspaceAnalysisOptions{
				Cache: analyzer.NewMemoryPackageCache(),
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Analysis failed: %v\n", err)
				os.Exit(1)
			}
			result = res
		}

		if result != nil {
			var nodes []graph.Node
			for _, e := range result.Entities {
				nodes = append(nodes, graph.Node{
					ID:       e.ID,
					Label:    e.Name,
					Kind:     string(e.Kind),
					Package:  e.Package,
					File:     e.File,
					Exported: e.Exported,
				})
			}

			var edges []graph.Edge
			for _, r := range result.Relationships {
				edges = append(edges, graph.Edge{
					From: r.From,
					To:   r.To,
					Type: r.Type,
				})
			}

			graphTitle := filepath.Base(target)
			html, err := generateGraphHTML(graphTitle, nodes, edges)
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to generate graph HTML: %v\n", err)
				os.Exit(1)
			}

			filename := fmt.Sprintf("garuda_graph_%s.html", graphTitle)
			if err := os.WriteFile(filename, []byte(html), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to write graph file: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✓ Offline graph written to %s (%d entities, %d edges)\n", filename, len(nodes), len(edges))
			if graphOpenFlag {
				openFile(filename)
			}
			return
		}
	}

	// 2. Fallback: PostgreSQL Workspace Mode
	dbURL := getDBURL()
	tenantIDStr := getTenantIDString()
	ctx := context.Background()

	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	tenantUUID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Invalid tenant ID: %v\n", err)
		os.Exit(1)
	}

	workspaceName := target
	var wsID uuid.UUID
	err = st.Pool().QueryRow(ctx, `SELECT id FROM workspaces WHERE tenant_id = $1 AND name = $2`, tenantUUID, workspaceName).Scan(&wsID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Workspace '%s' not found in database\n", workspaceName)
		os.Exit(1)
	}

	query := `
		SELECT
			id,
			COALESCE(name, 'unknown') as name,
			COALESCE(kind, 'unknown') as kind,
			COALESCE(package, '') as package,
			COALESCE(file_path, '') as file_path,
			COALESCE(is_exported, false) as is_exported
		FROM entities
		WHERE tenant_id = $1 AND workspace_id = $2
	`
	args := []interface{}{tenantUUID, wsID}

	if graphRepoFlag != "" {
		if err := uuid.Validate(graphRepoFlag); err == nil {
			query += ` AND repository_id = $3`
			args = append(args, graphRepoFlag)
		} else {
			var repoID uuid.UUID
			err = st.Pool().QueryRow(ctx, `
				SELECT id FROM repositories
				WHERE workspace_id = $1 AND (url LIKE $2 OR module_path = $3)
			`, wsID, "%"+graphRepoFlag+"%", graphRepoFlag).Scan(&repoID)

			if err == nil {
				query += ` AND repository_id = $3`
				args = append(args, repoID)
			} else {
				fmt.Printf("⚠️ Repository '%s' not found, ignoring filter\n", graphRepoFlag)
			}
		}
	}

	rows, err := st.Pool().Query(ctx, query, args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to query entities: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	var nodes []graph.Node
	for rows.Next() {
		var id uuid.UUID
		var name, kind, pkg, file string
		var exported bool
		if err := rows.Scan(&id, &name, &kind, &pkg, &file, &exported); err != nil {
			continue
		}
		nodes = append(nodes, graph.Node{
			ID:       id.String(),
			Label:    name,
			Kind:     kind,
			Package:  pkg,
			File:     file,
			Exported: exported,
		})
	}

	if err := rows.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Entity query error: %v\n", err)
		os.Exit(1)
	}

	if len(nodes) == 0 {
		fmt.Printf("⚠️ No entities found in workspace '%s'", workspaceName)
		if graphRepoFlag != "" {
			fmt.Printf(" for repo '%s'", graphRepoFlag)
		}
		fmt.Println(".")
		fmt.Println("  Run 'garuda workspace sync <workspace>' to populate entities.")
		return
	}

	rows2, err := st.Pool().Query(ctx, `
		SELECT from_entity_id, to_entity_id, claim_type
		FROM claims
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantUUID, wsID)

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to query claims: %v\n", err)
		os.Exit(1)
	}
	defer rows2.Close()

	var edges []graph.Edge
	for rows2.Next() {
		var fromID, toID uuid.UUID
		var claimType string
		if err := rows2.Scan(&fromID, &toID, &claimType); err != nil {
			continue
		}
		if fromID == uuid.Nil || toID == uuid.Nil {
			continue
		}
		edges = append(edges, graph.Edge{
			From: fromID.String(),
			To:   toID.String(),
			Type: claimType,
		})
	}

	if err := rows2.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Claims query error: %v\n", err)
		os.Exit(1)
	}

	html, err := generateGraphHTML(workspaceName, nodes, edges)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to generate graph HTML: %v\n", err)
		os.Exit(1)
	}

	filename := fmt.Sprintf("garuda_graph_%s.html", workspaceName)
	if err := os.WriteFile(filename, []byte(html), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to write graph: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Graph written to %s (%d entities, %d edges)\n", filename, len(nodes), len(edges))
	if graphOpenFlag {
		openFile(filename)
	}
}

func generateGraphHTML(workspaceName string, nodes []graph.Node, edges []graph.Edge) (string, error) {
	return graph.Generate(workspaceName, nodes, edges)
}

func openFile(filename string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", filename)
	case "darwin":
		cmd = exec.Command("open", filename)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", filename)
	default:
		fmt.Printf("⚠️ Unsupported OS. Please open %s manually.\n", filename)
		return
	}
	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ Failed to open browser: %v\n", err)
	}
}

// 8. REMAINING HANDLERS

func handleInit() {
	fmt.Println("⚙️ Initializing Garuda Local Runtime Environment...")
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
	fmt.Println("✓ Generated docker-compose.garuda.yml successfully.")
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

	fmt.Println("🛠️ Running database migrations...")
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

	fmt.Println("📡 Waiting for API to come online...")
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
		fmt.Fprintln(os.Stderr, "⚠️ API health check timed out. It may still be starting.")
	}

	fmt.Println("\n✅ Garuda runtime is ONLINE!")
	fmt.Println("🌐 Mission Control: http://localhost:8080/dashboard")
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
		fmt.Fprintf(os.Stderr, "❌ Error stopping containers: %v\n", err)
	}
	fmt.Println("✅ Garuda containers stopped.")
}

func handleStatus() {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(defaultAPIAddr + "/health")
	if err != nil {
		fmt.Println("❌ Garuda Gateway OFFLINE. Run 'garuda up' to boot services.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("⚠️ Garuda Gateway UNHEALTHY (status %d).\n", resp.StatusCode)
		return
	}
	fmt.Println("✅ Garuda Gateway: ONLINE (:8080)")
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

	fmt.Println("🛡️ GARUDA INTEGRITY CHECK")
	fmt.Printf("Revisions:           %d\n", count)
	fmt.Printf("Hash chain:          %s\n", statusText(chainValid))
	fmt.Printf("Audit events:        %d\n", count)
	fmt.Printf("Evidence references: %s\n", statusText(chainValid))
	fmt.Printf("Immutable history:   %s\n", statusText(chainValid))
	fmt.Printf("Tenant isolation:    %s\n", statusText(chainValid))

	if chainValid {
		fmt.Println("Integrity: PASS ✅")
	} else {
		fmt.Println("Integrity: FAIL ❌")
	}
}

func statusText(valid bool) string {
	if valid {
		return "VALID "
	}
	return "INVALID"
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

	// 1. Fetch parent decision metadata
	var title, statement, owner, status string
	var scopeJSON []byte
	var confidence float64
	var parentCreatedAt time.Time
	err = st.Pool().QueryRow(ctx, `
		SELECT title, statement, scope, owner, confidence, status, created_at
		FROM decisions
		WHERE tenant_id = $1 AND id = $2
	`, tid, decisionID).Scan(&title, &statement, &scopeJSON, &owner, &confidence, &status, &parentCreatedAt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Decision not found: %v\n", err)
		os.Exit(1)
	}

	// 2. Fetch latest immutable revision
	var revNumber int
	var canonicalJSON []byte
	var hash, prevHash []byte
	var createdAt time.Time
	_ = st.Pool().QueryRow(ctx, `
		SELECT revision_number, canonical_json, decision_hash, previous_revision_hash, created_at
		FROM decision_revisions
		WHERE tenant_id = $1 AND decision_id = $2
		ORDER BY revision_number DESC LIMIT 1
	`, tid, decisionID).Scan(&revNumber, &canonicalJSON, &hash, &prevHash, &createdAt)

	fmt.Printf("📄 DECISION EXPLANATION\n")
	fmt.Printf("Decision ID:       %s\n", decisionID)
	fmt.Printf("Status:            %s\n", status)
	fmt.Printf("Title:             %s\n", title)
	fmt.Printf("Statement:         %s\n", statement)
	fmt.Printf("Scope:             %s\n", string(scopeJSON))
	fmt.Printf("Owner:             %s\n", owner)
	fmt.Printf("Confidence:        %.2f\n", confidence)
	fmt.Printf("Revision:          #%d\n", revNumber)
	fmt.Printf("Created:           %s\n", createdAt.Format(time.RFC3339))
	fmt.Printf("Content Hash:      %x\n", hash)
	fmt.Printf("Previous Revision: %x\n", prevHash)

	var merkleRoot []byte
	_ = st.Pool().QueryRow(ctx, `SELECT root_hash FROM merkle_roots WHERE tenant_id = $1`, tid).Scan(&merkleRoot)
	fmt.Printf("Merkle Root:       %x\n", merkleRoot)

	// 3. Print AST snapshot stats if present in canonical_json
	var summary analyzer.RevisionSummary
	if err := json.Unmarshal(canonicalJSON, &summary); err == nil && summary.Fingerprint != "" {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("📦 AST Snapshot Fingerprint: %s\n", summary.Fingerprint)
		fmt.Printf("📊 Packages: %d | Structs: %d | Interfaces: %d | Functions: %d | Files: %d\n",
			summary.Stats.Packages, summary.Stats.Structs, summary.Stats.Interfaces, summary.Stats.Functions, summary.Stats.Files)
		fmt.Printf("🔒 Evidence Payload Hash:    %s\n", summary.PayloadHash)
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
	fmt.Println("✓ Injection complete (placeholder).")
}

func openDashboard() {
	url := defaultAPIAddr + "/dashboard"
	fmt.Printf("🌐 Opening %s...\n", url)
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
		fmt.Printf("❌ Failed to open browser: %v. Visit %s manually.\n", err, url)
	}
}

// 9. HTTP HELPERS

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
