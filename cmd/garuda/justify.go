package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/justify"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/spf13/cobra"
)

var (
	justifyOutput string
	justifyJSON   bool
)

func init() {
	justifyCmd.Flags().StringVarP(&justifyOutput, "output", "o", "", "Write output to file")
	justifyCmd.Flags().BoolVar(&justifyJSON, "json", false, "Output JSON")
	rootCmd.AddCommand(justifyCmd)
}

var justifyCmd = &cobra.Command{
	Use:   "justify <entity-name>",
	Short: "Justify why an entity exists with evidence and confidence",
	Long: `Provides a detailed justification for any code entity (struct, interface, function, API).
Uses Ponytail principles to evaluate:
  - Necessity (incoming calls)
  - Simplicity (complexity score)
  - Standard library alternatives
  - Duplication
  - Contract impact (breaking changes)

Output includes evidence (file:line references) and a confidence score.

Examples:
  garuda justify PaymentService
  garuda justify PaymentService --json -o justification.json`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleJustify(args[0])
	},
}

func handleJustify(entityName string) {
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

	// Get entity details
	entity, err := st.GetEntity(ctx, tenantUUID, wsID, entityName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Entity '%s' not found: %v\n", entityName, err)
		os.Exit(1)
	}

	// Get relationships
	incoming, outgoing, err := st.GetEntityRelationships(ctx, tenantUUID, wsID, entity.Name, entity.Package)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️ Failed to get relationships: %v\n", err)
	}

	// Build justification
	just := justify.Justify(entity, incoming, outgoing)

	if justifyJSON {
		data, _ := json.MarshalIndent(just, "", "  ")
		if justifyOutput != "" {
			os.WriteFile(justifyOutput, data, 0644)
			fmt.Printf("📄 JSON written to %s\n", justifyOutput)
		} else {
			fmt.Println(string(data))
		}
		return
	}

	printJustification(just)
}

func printJustification(j *justify.Justification) {
	fmt.Printf("\n🧠 JUSTIFICATION: %s (%s)\n", j.EntityName, j.File)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	fmt.Printf("📌 Necessity: %s\n", statusText(j.Necessity.Verified))
	fmt.Printf("  • Called by: %d consumers\n", j.Necessity.Consumers)
	if j.Necessity.LastChanged != "" {
		fmt.Printf("  • Last changed: %s\n", j.Necessity.LastChanged)
	}
	fmt.Printf("  • Confidence: %.2f\n\n", j.Necessity.Confidence)

	fmt.Printf("🔍 Simplicity: %s\n", statusText(j.Simplicity.Pass))
	if j.Simplicity.Suggestion != "" {
		fmt.Printf("  • Suggestion: %s\n", j.Simplicity.Suggestion)
	}
	fmt.Printf("  • Lines: %d\n\n", j.Simplicity.Lines)

	fmt.Printf("📚 Standard Library: %s\n", statusText(j.StdLib.Pass))
	if j.StdLib.Alternative != "" {
		fmt.Printf("  • Alternative: %s\n", j.StdLib.Alternative)
	}
	fmt.Printf("  • Confidence: %.2f\n\n", j.StdLib.Confidence)

	fmt.Printf("🔄 Duplication: %s\n", statusText(j.Duplication.Pass))
	if j.Duplication.DuplicateOf != "" {
		fmt.Printf("  • Duplicate of: %s\n", j.Duplication.DuplicateOf)
	}
	fmt.Printf("  • Confidence: %.2f\n\n", j.Duplication.Confidence)

	fmt.Printf("🔗 Contract Impact: %d consumers affected\n", j.ContractImpact.Consumers)
	if j.ContractImpact.IsBreaking {
		fmt.Printf("  ⚠️ BREAKING CHANGE – %s\n", j.ContractImpact.Details)
	} else {
		fmt.Printf("  ✅ Backwards compatible\n")
	}
	fmt.Printf("  • Evidence: %v\n\n", strings.Join(j.ContractImpact.Evidence, ", "))

	fmt.Printf("🔒 Evidence:\n")
	for _, ev := range j.Evidence {
		fmt.Printf("  • %s\n", ev)
	}

	fmt.Printf("\n✅ %s\n", j.Conclusion)
	fmt.Printf("   Confidence: %.2f\n", j.OverallConfidence)
}
