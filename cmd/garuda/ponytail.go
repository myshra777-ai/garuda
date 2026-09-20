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

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/hygiene"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/spf13/cobra"
)

func newHygieneCommand(use string, deprecated string) *cobra.Command {
	var outputJSON bool
	var outputPath string

	cmd := &cobra.Command{
		Use:        use,
		Short:      "Report static hygiene candidates and standard-library alternatives",
		Deprecated: deprecated,
		Long: `Scans the workspace semantic graph and reports advisory findings:
  - Static unreferenced candidates (zero incoming graph references)
  - Duplicate symbol-name candidates across packages
  - Standard-library alternative suggestions

These findings are advisory. A zero-incoming entity is not proof of dead code,
and a repeated symbol name is not proof of duplicated implementation.

Examples:
  garuda hygiene .
  garuda hygiene . --json -o hygiene.json`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}
			handleHygiene(path, outputJSON, outputPath)
		},
	}

	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output JSON")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Write report to file")

	return cmd
}

var (
	hygieneCmd  = newHygieneCommand("hygiene [path]", "")
	ponytailCmd = newHygieneCommand("ponytail [path]", "use `garuda hygiene` instead")
)

func init() {
	rootCmd.AddCommand(hygieneCmd)
	rootCmd.AddCommand(ponytailCmd)
}

func hygieneRelationships(edges []map[string]interface{}) []analyzer.Relationship {
	relationships := make([]analyzer.Relationship, 0, len(edges))
	for _, edge := range edges {
		from, fromOK := edge["from"].(string)
		to, toOK := edge["to"].(string)
		typ, typeOK := edge["type"].(string)
		if !fromOK || !toOK || !typeOK || from == "" || to == "" || typ == "" {
			continue
		}
		relationships = append(relationships, analyzer.Relationship{
			From: from,
			To:   to,
			Type: typ,
		})
	}
	return relationships
}

func handleHygiene(path string, outputJSON bool, outputPath string) {
	dbURL := getDBURL()
	tenantID := getTenantIDString()
	ctx := context.Background()

	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	workspaceName := getWorkspaceName()

	tenantUUID := uuid.MustParse(tenantID)
	ws, err := st.GetWorkspaceByName(ctx, tenantID, workspaceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Workspace '%s' not found\n", workspaceName)
		os.Exit(1)
	}
	wsID := ws.ID

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

	relationships := hygieneRelationships(edges)
	hygieneReport := hygiene.Analyze(entities, relationships)

	report := struct {
		DeadCode      []string       `json:"dead_code"`
		Duplications  []string       `json:"duplications"`
		StdLibAlts    []string       `json:"stdlib_alternatives"`
		Summary       map[string]int `json:"summary"`
		Entities      int            `json:"total_entities"`
		Relationships int            `json:"total_relationships"`
	}{}

	for _, finding := range hygieneReport.Findings {
		switch finding.Kind {
		case hygiene.FindingStaticUnreferencedCandidate:
			report.DeadCode = append(report.DeadCode, finding.Message)
		case hygiene.FindingDuplicateSymbolName:
			report.Duplications = append(report.Duplications, finding.Message)
		case hygiene.FindingStandardLibraryAlternative:
			report.StdLibAlts = append(report.StdLibAlts, finding.Message)
		}
	}

	report.Summary = map[string]int{
		"dead_code":    len(report.DeadCode),
		"duplications": len(report.Duplications),
		"stdlib_alts":  len(report.StdLibAlts),
	}
	report.Entities = hygieneReport.EntityCount
	report.Relationships = hygieneReport.RelationCount

	if outputJSON {
		data, _ := json.MarshalIndent(report, "", "  ")
		if outputPath != "" {
			os.WriteFile(outputPath, data, 0644)
			fmt.Printf("📄 JSON written to %s\n", outputPath)
		} else {
			fmt.Println(string(data))
		}
		return
	}

	// Human output
	fmt.Println("🔍 PONYTAIL REPORT")
	fmt.Println("📊 Ponytail report generated.")
	fmt.Printf("Entities: %d, Relationships: %d\n\n", report.Entities, report.Relationships)

	if len(report.DeadCode) > 0 {
		fmt.Printf("STATIC UNREFERENCED CANDIDATES %d\n", len(report.DeadCode))
		for _, dc := range report.DeadCode {
			fmt.Printf("  • %s\n", dc)
		}
		fmt.Println()
	}
	if len(report.Duplications) > 0 {
		fmt.Printf("DUPLICATE SYMBOL-NAME CANDIDATES %d\n", len(report.Duplications))
		for _, dup := range report.Duplications {
			fmt.Printf("  • %s\n", dup)
		}
		fmt.Println()
	}
	if len(report.StdLibAlts) > 0 {
		fmt.Printf("STANDARD-LIBRARY ALTERNATIVE SUGGESTIONS %d\n", len(report.StdLibAlts))
		for _, alt := range report.StdLibAlts {
			fmt.Printf("  • %s\n", alt)
		}
		fmt.Println()
	}
	if len(report.DeadCode) == 0 && len(report.Duplications) == 0 && len(report.StdLibAlts) == 0 {
		fmt.Println("No static hygiene candidates or suggestions found.")
	}
}
