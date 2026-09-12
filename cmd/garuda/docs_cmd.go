// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/knowledge"
	"github.com/myshra777-ai/garuda/internal/knowledge/markdown"

	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Document ingestion, intent extraction, and verification commands",
	Long: `Document ingestion and verification for Garuda's Knowledge Layer.

Supported formats: Markdown (.md, .markdown), PDF (.pdf), DOCX (.docx),
Plain text (.txt), reStructuredText (.rst).

Commands:
  ingest <path>   Extract claims from a file or directory
  sync            Ingest all sources declared in .garuda/workspace.yaml
  status          Show freshness of configured document sources
  verify          Correlate claims against AST entities and detect drift`,
}

// -----------------------------------------------------------------------------
// docs ingest <path>
// -----------------------------------------------------------------------------

var docsIngestCmd = &cobra.Command{
	Use:   "ingest [file or directory]",
	Short: "Extract verifiable claims from documents (Markdown, PDF, DOCX, TXT)",
	Args:  cobra.ExactArgs(1),
	RunE:  runDocsIngest,
}

var supportedExtensions = map[string]bool{
	".md": true, ".markdown": true, ".pdf": true,
	".docx": true, ".txt": true, ".rst": true,
}

func runDocsIngest(cmd *cobra.Command, args []string) error {
	targetPath := knowledge.ExpandHome(args[0])
	dbURL := getDocsDBURL()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	claimStore := store.NewClaimStore(pool)
	tenantID := getDocsTenantID()
	workspace := getDocsWorkspace()

	files, err := collectFiles(targetPath)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		fmt.Println("⚠️  No supported documents found.")
		return nil
	}

	fmt.Printf("📚 Discovered %d document(s) to process\n\n", len(files))

	totalExtracted := 0
	totalFiles := 0
	skippedFiles := 0

	for _, file := range files {
		count, err := ingestOneFile(ctx, file, tenantID, workspace, claimStore)
		if err != nil {
			fmt.Printf("⚠️  Skipped %s: %v\n", file, err)
			skippedFiles++
			continue
		}
		if count > 0 {
			totalFiles++
			totalExtracted += count
		}
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("🎉 Ingestion complete\n")
	fmt.Printf("   Files processed:   %d\n", totalFiles)
	fmt.Printf("   Files skipped:     %d\n", skippedFiles)
	fmt.Printf("   Claims extracted:  %d\n", totalExtracted)
	fmt.Println("═══════════════════════════════════════════════════════════")

	return nil
}

// ingestOneFile normalizes a single file, extracts claims, persists them.
// Returns the number of claims extracted.
func ingestOneFile(
	ctx context.Context,
	file string,
	tenantID uuid.UUID,
	workspace string,
	claimStore *store.ClaimStore,
) (int, error) {
	normalized, err := knowledge.NormalizeDocument(file)
	if err != nil {
		return 0, err
	}

	if strings.TrimSpace(normalized) == "" {
		return 0, fmt.Errorf("empty after normalization")
	}

	// Write normalized text to a temp .md so existing ADR parser can consume it
	tmp, err := os.CreateTemp("", "garuda-norm-*.md")
	if err != nil {
		return 0, err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(normalized); err != nil {
		tmp.Close()
		return 0, err
	}
	tmp.Close()

	claims, err := markdown.ParseADRPublic(tmp.Name(), file, tenantID, workspace)

	if err != nil {
		return 0, err
	}

	if len(claims) == 0 {
		return 0, nil
	}

	if err := claimStore.SaveClaims(ctx, claims); err != nil {
		return 0, fmt.Errorf("persist: %w", err)
	}

	fmt.Printf("✓ %s\n", file)
	fmt.Printf("   Extracted %d claim(s)\n", len(claims))
	for _, c := range claims {
		preview := c.Object
		if len(preview) > 60 {
			preview = preview[:57] + "..."
		}
		fmt.Printf("   ├─ [%s] %s %s %s (line %d)\n",
			c.Modality, c.Subject, c.Predicate, preview, c.LineStart)
	}
	fmt.Println()

	return len(claims), nil
}

func collectFiles(targetPath string) ([]string, error) {
	info, err := os.Stat(targetPath)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	var files []string
	if info.IsDir() {
		err := filepath.Walk(targetPath, func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if f.IsDir() {
				base := filepath.Base(path)
				if strings.HasPrefix(base, ".") || base == "node_modules" || base == "vendor" {
					return filepath.SkipDir
				}
				return nil
			}
			ext := strings.ToLower(filepath.Ext(f.Name()))
			if supportedExtensions[ext] {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		files = append(files, targetPath)
	}
	return files, nil
}

// -----------------------------------------------------------------------------
// docs sync (reads .garuda/workspace.yaml)
// -----------------------------------------------------------------------------

var docsSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Ingest all document sources declared in .garuda/workspace.yaml",
	RunE:  runDocsSync,
}

func runDocsSync(cmd *cobra.Command, args []string) error {
	configPath, err := knowledge.FindWorkspaceConfig()
	if err != nil {
		return fmt.Errorf("%w\n\nCreate .garuda/workspace.yaml first, or use 'garuda docs ingest <path>'", err)
	}

	fmt.Printf("📖 Reading config: %s\n\n", configPath)
	cfg, err := knowledge.LoadWorkspaceConfig(configPath)
	if err != nil {
		return err
	}

	fmt.Printf("Workspace: %s\n", cfg.Workspace)
	fmt.Printf("Document sources: %d\n\n", len(cfg.Documents))

	files, err := cfg.ResolveDocPaths()
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Println("⚠️  No documents found in configured sources.")
		return nil
	}

	fmt.Printf("📚 Discovered %d file(s) across all sources\n\n", len(files))

	dbURL := getDocsDBURL()
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer pool.Close()

	claimStore := store.NewClaimStore(pool)
	tenantID := getDocsTenantID()

	totalClaims := 0
	for _, file := range files {
		count, err := ingestOneFile(ctx, file, tenantID, cfg.Workspace, claimStore)
		if err != nil {
			fmt.Printf("⚠️  Skipped %s: %v\n", file, err)
			continue
		}
		totalClaims += count
	}

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("🎉 Sync complete: %d claims across %d files\n", totalClaims, len(files))
	fmt.Println("═══════════════════════════════════════════════════════════")

	return nil
}

// -----------------------------------------------------------------------------
// docs status (reads .garuda/workspace.yaml + DB)
// -----------------------------------------------------------------------------

var docsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show freshness of configured document sources",
	RunE:  runDocsStatus,
}

func runDocsStatus(cmd *cobra.Command, args []string) error {
	configPath, err := knowledge.FindWorkspaceConfig()
	if err != nil {
		return err
	}

	cfg, err := knowledge.LoadWorkspaceConfig(configPath)
	if err != nil {
		return err
	}

	dbURL := getDocsDBURL()
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer pool.Close()

	fmt.Printf("📖 Config:      %s\n", configPath)
	fmt.Printf("   Workspace:   %s\n", cfg.Workspace)
	fmt.Printf("   Tenant:      %s\n\n", cfg.TenantID)

	fmt.Println("DOCUMENT SOURCES")
	fmt.Println("═══════════════════════════════════════════════════════════")

	for _, doc := range cfg.Documents {
		status := "✓"
		note := ""

		info, err := os.Stat(doc.Path)
		if err != nil {
			status = "✗"
			note = "MISSING"
		} else if info.IsDir() {
			count := 0
			filepath.Walk(doc.Path, func(p string, f os.FileInfo, err error) error {
				if err == nil && !f.IsDir() {
					ext := strings.ToLower(filepath.Ext(f.Name()))
					if supportedExtensions[ext] {
						count++
					}
				}
				return nil
			})
			note = fmt.Sprintf("%d file(s)", count)
		} else {
			note = fmt.Sprintf("%d bytes", info.Size())
		}

		fmt.Printf("  %s  %s\n", status, doc.Path)
		if note != "" {
			fmt.Printf("       %s\n", note)
		}
	}

	// Claims in DB for this workspace
	var claimCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM document_claims WHERE workspace = $1
	`, cfg.Workspace).Scan(&claimCount)
	if err == nil {
		fmt.Println()
		fmt.Printf("  Claims in DB for '%s': %d\n", cfg.Workspace, claimCount)
	}

	fmt.Println("═══════════════════════════════════════════════════════════")

	return nil
}

// -----------------------------------------------------------------------------
// docs verify (unchanged)
// -----------------------------------------------------------------------------

var docsVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Correlate document claims against AST entities and detect Knowledge Drift",
	RunE:  runDocsVerify,
}

func runDocsVerify(cmd *cobra.Command, args []string) error {
	dbURL := getDocsDBURL()
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	tenantID := getDocsTenantID()
	workspace := getDocsWorkspace()

	fmt.Println("🦅 Correlating Document Intent → Implementation AST → Runtime...")
	fmt.Println()

	evaluator := knowledge.NewEvaluator(pool)
	stats, undocumented, err := evaluator.EvaluateWorkspace(ctx, tenantID, workspace)
	if err != nil {
		return fmt.Errorf("evaluation failed: %w", err)
	}

	fmt.Println("DOCUMENTATION HEALTH")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("  %d documented capabilities\n", stats.TotalClaims)
	fmt.Printf("  %d backed by code                       ✓\n", stats.Supported)
	if stats.Unverified > 0 {
		fmt.Printf("  %d documented but not yet verified       ?\n", stats.Unverified)
	}
	if stats.Contradicted > 0 {
		fmt.Printf("  %d contradicted by code                 ✗\n", stats.Contradicted)
	}
	if stats.UndocumentedCode > 0 {
		fmt.Printf("  %d implemented but undocumented         ⚠\n", stats.UndocumentedCode)
	}
	fmt.Println("═══════════════════════════════════════════════════════════")

	// Document → Code drift
	rows, err := pool.Query(ctx, `
		SELECT subject, modality, predicate, object, status, contradiction_reason
		FROM document_claims
		WHERE tenant_id = $1 AND workspace = $2 AND status != 'SUPPORTED'
		ORDER BY line_start ASC
	`, tenantID, workspace)
	if err == nil {
		defer rows.Close()
		fmt.Println("\nDocument → Code Drift:")
		found := false
		for rows.Next() {
			var subj, mod, pred, obj, stat, reason string
			if err := rows.Scan(&subj, &mod, &pred, &obj, &stat, &reason); err != nil {
				continue
			}
			badge := "⚠"
			if stat == "CONTRADICTED" {
				badge = "❌"
			}
			fmt.Printf("  %s [%s] `%s` %s `%s`\n      → %s\n",
				badge, mod, subj, pred, obj, reason)
			found = true
		}
		if !found {
			fmt.Println("  ✓ No drift detected")
		}
	}

	// Code → Document drift
	if len(undocumented) > 0 {
		fmt.Printf("\nCode → Document Drift (%d undocumented):\n", len(undocumented))
		for _, u := range undocumented {
			fmt.Printf("  ⚠ [%s] `%s` in %s\n", u.Kind, u.SymbolName, u.Path)
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func getDocsDBURL() string {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://test:test@localhost:5433/garuda_test?sslmode=disable"
	}
	return url
}

func getDocsTenantID() uuid.UUID {
	t := os.Getenv("GARUDA_TENANT_ID")
	if t == "" {
		return uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}
	return uuid.MustParse(t)
}

func getDocsWorkspace() string {
	w := os.Getenv("GARUDA_WORKSPACE")
	if w == "" {
		return "default"
	}
	return w
}

// -----------------------------------------------------------------------------
// Registration
// -----------------------------------------------------------------------------

func init() {
	docsCmd.AddCommand(docsIngestCmd)
	docsCmd.AddCommand(docsSyncCmd)
	docsCmd.AddCommand(docsStatusCmd)
	docsCmd.AddCommand(docsVerifyCmd)
	rootCmd.AddCommand(docsCmd)
}
