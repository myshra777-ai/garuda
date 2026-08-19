// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package selfdescribe

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/store"
)

// Options for self-describe
type Options struct {
	Path        string
	Workspace   string
	OutputFile  string
	Markdown    bool
	TenantID    string
	DatabaseURL string
}

// Run executes the self-describe command
func Run(ctx context.Context, opts *Options) error {
	// 1. Run analysis on the path
	result, err := analyzer.Analyze(opts.Path)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// 2. Gather Git metadata
	repoURL, commit, branch := getGitMetadata(opts.Path)

	// 3. Connect to DB if workspace is provided
	var semantic SemanticInfo
	var trust TrustInfo
	var workspaceID uuid.UUID

	if opts.Workspace != "" && opts.TenantID != "" && opts.DatabaseURL != "" {
		db, err := store.NewPostgresStore(opts.DatabaseURL)
		if err != nil {
			return fmt.Errorf("failed to connect to DB: %w", err)
		}
		defer db.Close()

		// Get workspace ID
		err = db.Pool().QueryRow(ctx, `
			SELECT id FROM workspaces
			WHERE tenant_id = $1 AND name = $2
		`, opts.TenantID, opts.Workspace).Scan(&workspaceID)
		if err != nil {
			return fmt.Errorf("workspace '%s' not found: %w", opts.Workspace, err)
		}

		// Get semantic stats
		semantic, err = getSemanticStats(ctx, db, opts.TenantID, workspaceID)
		if err != nil {
			return fmt.Errorf("failed to get semantic stats: %w", err)
		}

		// Get trust evidence
		trust, err = getTrustEvidence(ctx, db, opts.TenantID)
		if err != nil {
			return fmt.Errorf("failed to get trust evidence: %w", err)
		}
	} else {
		// Fallback: use analyzer stats
		semantic = SemanticInfo{
			Entities:      len(result.Entities),
			Relationships: len(result.Relationships),
			Evidence:      0,
			Lineage:       false,
		}
		trust = TrustInfo{
			ImmutableLedger: false,
			MerkleRoot:      "not_available",
			RevisionCount:   0,
			AuditTrail:      false,
		}
	}

	// 4. Extract capabilities from the analyzer result
	capabilities := ExtractCapabilities(result)

	// 5. Parse CLI commands (if this is a Go project with Cobra)
	cliInfo := ParseCLICommands(opts.Path)

	// 6. Try to read product metadata
	product := ReadProductMetadata(opts.Path)

	// 7. Try to read roadmap (if exists)
	roadmap := ReadRoadmap(opts.Path)

	// 8. Build the final description
	desc := &SelfDescription{
		SchemaVersion: "1.0",
		GeneratedAt:   time.Now().UTC(),
		Source: SourceInfo{
			Repository:    repoURL,
			Commit:        commit,
			Branch:        branch,
			Language:      "Go",
			LanguageScope: []string{"Go"},
		},
		Product:      product,
		Capabilities: capabilities,
		CLI:          cliInfo,
		Semantic:     semantic,
		Benchmarks: BenchmarkInfo{
			Available: false,
			Metrics:   map[string]interface{}{},
		},
		Trust:   trust,
		Roadmap: roadmap,
	}

	// 9. Output
	if opts.Markdown {
		md := GenerateMarkdown(desc)
		if opts.OutputFile != "" {
			return os.WriteFile(opts.OutputFile, []byte(md), 0644)
		}
		fmt.Println(md)
		return nil
	}

	data, err := json.MarshalIndent(desc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if opts.OutputFile != "" {
		return os.WriteFile(opts.OutputFile, data, 0644)
	}

	fmt.Println(string(data))
	return nil
}

// getGitMetadata extracts repository URL, commit, and branch
func getGitMetadata(path string) (string, string, string) {
	repoURL := ""
	commit := ""
	branch := ""

	// Get remote URL
	cmd := exec.Command("git", "-C", path, "remote", "get-url", "origin")
	if out, err := cmd.Output(); err == nil {
		repoURL = strings.TrimSpace(string(out))
	}

	// Get commit SHA
	cmd = exec.Command("git", "-C", path, "rev-parse", "HEAD")
	if out, err := cmd.Output(); err == nil {
		commit = strings.TrimSpace(string(out))
	}

	// Get branch name
	cmd = exec.Command("git", "-C", path, "branch", "--show-current")
	if out, err := cmd.Output(); err == nil {
		branch = strings.TrimSpace(string(out))
	}

	return repoURL, commit, branch
}
