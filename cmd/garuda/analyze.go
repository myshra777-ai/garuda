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
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/store"
)

var (
	outputFlag         string
	outputFileFlag     string
	jsonFlag           bool
	jsonOutputFlag     bool
	saveFlag           bool
	workspaceFlag      string
	repoFlag           string
	modulePathFlag     string
	failOnBreakingFlag bool
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [path...]",
	Short: "Analyze a Go codebase, extract semantic schema, and optionally persist to ledger",
	Long:  `Analyzes a single repository or a local path. Use --save to persist the analysis into the cryptographic ledger.`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) == 1 {
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

func init() {
	analyzeCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Write JSON report to file")
	analyzeCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output report in JSON format")
	analyzeCmd.Flags().BoolVarP(&saveFlag, "save", "s", false, "Save analysis snapshot into PostgreSQL ledger")
	analyzeCmd.Flags().StringVar(&workspaceFlag, "workspace", "", "Workspace name (for provenance)")
	analyzeCmd.Flags().StringVar(&repoFlag, "repo", "", "Repository URL (for provenance)")
	analyzeCmd.Flags().String("commit", "", "Commit SHA (auto-detected if not provided)")
	analyzeCmd.Flags().StringVar(&modulePathFlag, "module-path", "", "Go module path (auto-detected if not provided)")

	diffCmd.Flags().BoolVar(&jsonOutputFlag, "json", false, "Output diff in JSON format")
	diffCmd.Flags().StringVarP(&outputFileFlag, "output", "o", "", "Write diff to file")
	diffCmd.Flags().BoolVar(&failOnBreakingFlag, "fail-on-breaking", false, "Exit with code 1 if breaking changes are detected (CI Gate)")

	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(diffCmd)
}

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

	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Invalid path '%s': %v\n", path, err)
		os.Exit(1)
	}

	ctx := context.Background()

	// ─────────────────────────────────────────────────────────────
	// Language dispatch: Python vs Go
	// ─────────────────────────────────────────────────────────────
	var result *analyzer.Result
	var language string

	if analyzer.IsPythonProject(absPath) {
		fmt.Printf("🐍 Analyzing Python project %s...\n", absPath)
		pyResult, err := analyzer.AnalyzePythonWorkspace(ctx, absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Python analysis failed: %v\n", err)
			os.Exit(1)
		}
		result = pyResult
		language = "python"
		fmt.Printf("✓ %d entities, %d relationships extracted\n",
			len(result.Entities), len(result.Relationships))
	} else {
		fmt.Printf("🔍 Analyzing Go workspace %s...\n", absPath)
		wsMeta, err := analyzer.DiscoverWorkspace(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Workspace discovery failed: %v\n", err)
			os.Exit(1)
		}

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

		goResult, err := analyzer.AnalyzeWorkspaceWithOptions(ctx, wsMeta, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Analysis failed: %v\n", err)
			os.Exit(1)
		}
		if goResult.Stats.Files == 0 && len(goResult.Entities) == 0 {
			fmt.Fprintf(os.Stderr, "❌ No Go files found in '%s'. Aborting.\n", path)
			os.Exit(1)
		}
		result = goResult
		language = "go"
	}

	// ─────────────────────────────────────────────────────────────
	// JSON output (both languages)
	// ─────────────────────────────────────────────────────────────
	isJSONRequested := jsonFlag || jsonOutputFlag
	if outputFlag != "" || isJSONRequested {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️ Failed to marshal JSON snapshot: %v\n", err)
		} else {
			if outputFlag != "" {
				if err := os.WriteFile(outputFlag, data, 0644); err != nil {
					fmt.Fprintf(os.Stderr, "⚠️ Failed to write JSON: %v\n", err)
				} else {
					fmt.Printf("📄 JSON report saved to %s (Fingerprint: %s)\n", outputFlag, result.Fingerprint)
				}
			}
			if isJSONRequested && outputFlag == "" {
				fmt.Println(string(data))
			}
		}
	}

	// ─────────────────────────────────────────────────────────────
	// Persist to ledger (both languages)
	// ─────────────────────────────────────────────────────────────
	if saveFlag || workspaceFlag != "" || repoFlag != "" {
		dbURL := getDBURL()
		tenantIDStr := getTenantIDString()
		tenantUUID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Invalid tenant UUID '%s': %v\n", tenantIDStr, err)
			os.Exit(1)
		}

		persistCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		st, err := store.NewPostgresStore(dbURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to connect to DB: %v\n", err)
			os.Exit(1)
		}
		defer st.Close()

		decisionID, revisionID, rev, err := st.SaveAnalysisDecision(persistCtx, tenantIDStr, result)
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
				repoURL = "file://" + absPath
			}
		}
		repoURL = sanitizeGitURL(repoURL)

		// Module path detection: Go reads go.mod, Python reads pyproject.toml
		modulePath := modulePathFlag
		if modulePath == "" {
			switch language {
			case "go":
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
			case "python":
				pyprojectPath := filepath.Join(path, "pyproject.toml")
				if data, err := os.ReadFile(pyprojectPath); err == nil {
					for _, line := range strings.Split(string(data), "\n") {
						line = strings.TrimSpace(line)
						if strings.HasPrefix(line, "name") && strings.Contains(line, "=") {
							parts := strings.SplitN(line, "=", 2)
							if len(parts) == 2 {
								v := strings.TrimSpace(parts[1])
								v = strings.Trim(v, `"'`)
								if v != "" {
									modulePath = v
									break
								}
							}
						}
					}
				}
				if modulePath == "" {
					// Fall back to directory name
					modulePath = filepath.Base(absPath)
				}
			}
		}

		var workspaceID uuid.UUID
		wsRecord, err := st.GetWorkspaceByName(persistCtx, tenantIDStr, wsName)
		if err != nil {
			newWs, err := st.CreateWorkspace(persistCtx, tenantIDStr, wsName, "")
			if err != nil {
				fmt.Fprintf(os.Stderr, "⚠️ Failed to create workspace '%s': %v\n", wsName, err)
				return
			}
			workspaceID = newWs.ID
			fmt.Printf("   Created workspace: %s\n", wsName)
		} else {
			workspaceID = wsRecord.ID
		}

		var repoID uuid.UUID
		repoRecord, err := st.GetRepositoryByURL(persistCtx, workspaceID, repoURL)
		if err != nil {
			provider := "local"
			if strings.Contains(repoURL, "github.com") {
				provider = "github"
			} else if strings.Contains(repoURL, "gitlab.com") {
				provider = "gitlab"
			} else if strings.Contains(repoURL, "bitbucket") {
				provider = "bitbucket"
			}
			newRepo, err := st.AddRepository(persistCtx, workspaceID, provider, repoURL, "main", language, modulePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "⚠️ Failed to create repository: %v\n", err)
				return
			}
			repoID = newRepo.ID
			fmt.Printf("   Created repository: %s (%s)\n", repoURL, language)
		} else {
			repoID = repoRecord.ID
		}

		if modulePath != "" {
			_ = st.UpdateRepositoryModulePath(persistCtx, repoID, modulePath)
			fmt.Printf("   Module path: %s\n", modulePath)
		}

		commitSHA := "local-uncommitted"
		gitCmd := exec.Command("git", "-C", path, "rev-parse", "HEAD")
		if out, err := gitCmd.Output(); err == nil {
			commitSHA = strings.TrimSpace(string(out))
		}

		err = st.SaveSemanticGraph(persistCtx, tenantUUID, workspaceID, repoID, revisionID, result, commitSHA)
		if err != nil {
			fmt.Printf("⚠️ Failed to save semantic graph: %v\n", err)
		} else {
			fmt.Printf("   🧠 Semantic graph saved (%d entities, %d relationships, commit: %s)\n",
				len(result.Entities), len(result.Relationships), commitSHA[:min(8, len(commitSHA))])
			_ = st.UpdateRepositorySyncStatus(persistCtx, tenantIDStr, repoID, commitSHA, "synced")
		}
	}
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
