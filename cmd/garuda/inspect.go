// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

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
	tenantID := getTenantID()
	ctx := context.Background()

	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect to DB: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	insp, err := st.InspectEntityDetails(ctx, tenantID, entityName)
	if err != nil {
		fmt.Printf("❌ Entity '%s' not found.\n", entityName)
		os.Exit(1)
	}

	fmt.Printf("🔍 Entity: %s\n", insp.Name)
	fmt.Printf("   Package:  %s\n", insp.Package)
	fmt.Printf("   Kind:     %s\n", insp.Kind)
	fmt.Printf("   File:     %s\n", insp.FilePath)
	fmt.Printf("   Exported: %v\n", insp.IsExported)

	if insp.Signature != "" {
		fmt.Printf("   Signature: %s\n", insp.Signature)
	}

	if len(insp.Fields) > 0 {
		fmt.Printf("\n   Fields:\n")
		for _, f := range insp.Fields {
			fmt.Printf("      • %s: %s\n", f.Name, f.Type)
		}
	}

	if len(insp.Methods) > 0 {
		fmt.Printf("\n   Methods:\n")
		for _, m := range insp.Methods {
			fmt.Printf("      • %s %s\n", m.Name, m.Signature)
		}
	}

	if len(insp.OutgoingClaims) > 0 {
		fmt.Printf("\n   Claims (outgoing):\n")
		for _, c := range insp.OutgoingClaims {
			fmt.Println(c)
		}
	}

	if len(insp.IncomingClaims) > 0 {
		fmt.Printf("\n   Claims (incoming):\n")
		for _, c := range insp.IncomingClaims {
			fmt.Println(c)
		}
	}
}

func handleListEntities() {
	dbURL := getDBURL()
	tenantID := getTenantID()
	tenantStr := getTenantIDString()
	ctx := context.Background()

	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	workspaceName := os.Getenv("GARUDA_WORKSPACE")
	if workspaceName == "" {
		workspaceName = "default"
	}

	ws, err := st.GetWorkspaceByName(ctx, tenantStr, workspaceName)
	if err != nil {
		fmt.Printf("❌ Workspace '%s' not found.\n", workspaceName)
		fmt.Println("   Create one with: garuda workspace create " + workspaceName)
		os.Exit(1)
	}

	entities, err := st.ListWorkspaceEntities(ctx, tenantID, ws.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to list entities: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📋 Entities in workspace '%s' (%d):\n\n", workspaceName, len(entities))
	currentPkg := ""
	for _, e := range entities {
		if e.Package != currentPkg {
			currentPkg = e.Package
			fmt.Printf("📦 Package %s:\n", currentPkg)
		}
		exported := ""
		if !e.IsExported {
			exported = " (unexported)"
		}
		fmt.Printf("   • %s [%s]%s -> %s\n", e.Name, e.Kind, exported, e.FilePath)
	}
}

func handleGraph(workspaceName string) {
	dbURL := getDBURL()
	tenantUUID := getTenantID()
	tenantStr := getTenantIDString()
	ctx := context.Background()

	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	ws, err := st.GetWorkspaceByName(ctx, tenantStr, workspaceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Workspace '%s' not found in database\n", workspaceName)
		os.Exit(1)
	}

	var repoFilterID *uuid.UUID
	if graphRepoFlag != "" {
		repo, err := st.FindRepositoryByFilter(ctx, ws.ID, graphRepoFlag)
		if err == nil && repo != nil {
			repoFilterID = &repo.ID
		} else {
			fmt.Printf("⚠️ Repository '%s' not found, ignoring filter\n", graphRepoFlag)
		}
	}

	rawNodes, rawEdges, err := st.GetWorkspaceGraphElements(ctx, tenantUUID, ws.ID, repoFilterID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to query graph elements: %v\n", err)
		os.Exit(1)
	}

	if len(rawNodes) == 0 {
		fmt.Printf("⚠️ No entities found in workspace '%s'", workspaceName)
		if graphRepoFlag != "" {
			fmt.Printf(" for repo '%s'", graphRepoFlag)
		}
		fmt.Println(".")
		fmt.Println("   Run 'garuda workspace sync <workspace>' to populate entities.")
		return
	}

	var nodes []graph.Node
	for _, rn := range rawNodes {
		nodes = append(nodes, graph.Node{
			ID:       rn.ID,
			Label:    rn.Name,
			Kind:     rn.Kind,
			Package:  rn.Package,
			File:     rn.FilePath,
			Exported: rn.IsExported,
		})
	}

	var edges []graph.Edge
	for _, re := range rawEdges {
		edges = append(edges, graph.Edge{
			From: re.From,
			To:   re.To,
			Type: re.Type,
		})
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
