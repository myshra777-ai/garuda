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
	"strings"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/spf13/cobra"
)

var (
	ponytailOutput string
	ponytailJSON   bool
)

func init() {
	ponytailCmd.Flags().StringVarP(&ponytailOutput, "output", "o", "", "Write report to file")
	ponytailCmd.Flags().BoolVar(&ponytailJSON, "json", false, "Output JSON")
	rootCmd.AddCommand(ponytailCmd)
}

var ponytailCmd = &cobra.Command{
	Use:   "ponytail [path]",
	Short: "Detect dead code, duplications, and standard-lib alternatives",
	Long: `Scans the codebase and reports:
  - Dead code (zero incoming references)
  - Duplications (similar structural entities)
  - Standard library alternatives

Uses the semantic graph for evidence-backed findings.

Examples:
  garuda ponytail .
  garuda ponytail . --json -o ponytail.json`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		handlePonytail(path)
	},
}

func handlePonytail(path string) {
	dbURL := getDBURL()
	tenantID := getTenantIDString()
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

	tenantUUID := uuid.MustParse(tenantID)
	var wsID uuid.UUID
	err = st.Pool().QueryRow(ctx, `
		SELECT id FROM workspaces WHERE tenant_id = $1 AND name = $2
	`, tenantUUID, workspaceName).Scan(&wsID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Workspace '%s' not found\n", workspaceName)
		os.Exit(1)
	}

	// Fetch all entities and relationships
	entities, err := st.ListEntities(ctx, tenantUUID, wsID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to list entities: %v\n", err)
		os.Exit(1)
	}
	_, edges, err := st.GetGraphData(ctx, tenantUUID, wsID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to get graph data: %v\n", err)
		os.Exit(1)
	}

	// Build incoming references map
	incomingCount := make(map[string]int)
	for _, e := range edges {
		target := e["to"].(string)
		incomingCount[target]++
	}

	// Build report
	report := struct {
		DeadCode      []string       `json:"dead_code"`
		Duplications  []string       `json:"duplications"`
		StdLibAlts    []string       `json:"stdlib_alternatives"`
		Summary       map[string]int `json:"summary"`
		Entities      int            `json:"total_entities"`
		Relationships int            `json:"total_relationships"`
	}{}

	deadCode := []string{}
	for _, e := range entities {
		if incomingCount[e.ID] == 0 && e.Kind != "package" && e.Kind != "file" {
			deadCode = append(deadCode, fmt.Sprintf("%s (%s) in %s", e.Name, e.Kind, e.File))
		}
	}
	report.DeadCode = deadCode

	// Simple duplications: same name in different packages (simplistic)
	dupMap := make(map[string][]string)
	for _, e := range entities {
		dupMap[e.Name] = append(dupMap[e.Name], e.Package)
	}
	duplications := []string{}
	for name, pkgs := range dupMap {
		if len(pkgs) > 1 && name != "" {
			duplications = append(duplications, fmt.Sprintf("%s appears in %s", name, strings.Join(pkgs, ", ")))
		}
	}
	report.Duplications = duplications

	// Stdlib alternatives: contains, sort, etc.
	stdlib := []string{}
	for _, e := range entities {
		if strings.Contains(e.Name, "Contains") || strings.Contains(e.Name, "Sort") {
			alt := "slices.Contains or slices.Sort"
			if strings.Contains(e.Name, "Contains") {
				alt = "slices.Contains"
			}
			if strings.Contains(e.Name, "Sort") {
				alt = "slices.Sort"
			}
			stdlib = append(stdlib, fmt.Sprintf("%s → use %s", e.Name, alt))
		}
	}
	report.StdLibAlts = stdlib

	report.Summary = map[string]int{
		"dead_code":    len(deadCode),
		"duplications": len(duplications),
		"stdlib_alts":  len(stdlib),
	}
	report.Entities = len(entities)
	report.Relationships = len(edges)

	if ponytailJSON {
		data, _ := json.MarshalIndent(report, "", "  ")
		if ponytailOutput != "" {
			os.WriteFile(ponytailOutput, data, 0644)
			fmt.Printf("📄 JSON written to %s\n", ponytailOutput)
		} else {
			fmt.Println(string(data))
		}
		return
	}

	// Human output
	fmt.Println("🔍 PONYTAIL REPORT")
	fmt.Println("📊 Ponytail report generated.")
	fmt.Printf("Entities: %d, Relationships: %d\n\n", report.Entities, report.Relationships)

	if len(deadCode) > 0 {
		fmt.Printf("🔴 DEAD CODE (%d):\n", len(deadCode))
		for _, dc := range deadCode {
			fmt.Printf("  • %s\n", dc)
		}
		fmt.Println()
	}
	if len(duplications) > 0 {
		fmt.Printf("🟡 DUPLICATIONS (%d):\n", len(duplications))
		for _, dup := range duplications {
			fmt.Printf("  • %s\n", dup)
		}
		fmt.Println()
	}
	if len(stdlib) > 0 {
		fmt.Printf("📚 STANDARD LIBRARY ALTERNATIVES (%d):\n", len(stdlib))
		for _, alt := range stdlib {
			fmt.Printf("  • %s\n", alt)
		}
		fmt.Println()
	}
	if len(deadCode) == 0 && len(duplications) == 0 && len(stdlib) == 0 {
		fmt.Println("✅ No issues found – your code is lean and clean.")
	}
}
