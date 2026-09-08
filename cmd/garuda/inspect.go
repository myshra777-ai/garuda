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
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/graph"
	"github.com/myshra777-ai/garuda/internal/store"
)

var (
	graphOpenFlag bool
	graphRepoFlag string
)

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

func init() {
	graphCmd.Flags().BoolVar(&graphOpenFlag, "open", false, "Open the graph in browser automatically")
	graphCmd.Flags().StringVar(&graphRepoFlag, "repo", "", "Filter graph to a specific repository (URL or ID)")

	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(graphCmd)
	rootCmd.AddCommand(entitiesCmd)
}

func handleInspect(entityName string) {
	dbURL := getDBURL()
	tenantID := getTenantIDString()
	ctx := context.Background()

	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect to DB: %v\n", err)
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

func handleGraph(target string) {
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
			html, err := graph.Generate(graphTitle, nodes, edges)
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

	html, err := graph.Generate(workspaceName, nodes, edges)
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
