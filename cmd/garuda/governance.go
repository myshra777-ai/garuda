// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/ingest"
	"github.com/myshra777-ai/garuda/internal/store"
)

type ProposeRequest struct {
	Title       string `json:"title"`
	ScopeDomain string `json:"scope_domain"`
	ScopeSystem string `json:"scope_system"`
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

var mcpInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Auto-inject Garuda MCP server into Claude Desktop & Cursor",
	Run: func(cmd *cobra.Command, args []string) {
		handleMCPInstall()
	},
}

func init() {
	rootCmd.AddCommand(proposeCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(explainCmd)
	rootCmd.AddCommand(ingestCmd)

	if mcpCmd != nil {
		mcpCmd.AddCommand(mcpInstallCmd)
	}
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

	// The Merkle revision chain is per-decision. Fetch every decision
	// that has at least one revision and verify each chain
	// independently.
	decisionIDs, err := st.ListDecisionsWithRevisions(ctx, tid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Query failed: %v\n", err)
		os.Exit(1)
	}

	if len(decisionIDs) == 0 {
		fmt.Println("ℹ️ No revisions in ledger.")
		return
	}

	totalRevisions := 0
	brokenDecisions := 0

	for _, decisionID := range decisionIDs {
		chain, err := st.GetRevisionChain(ctx, tid, decisionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to fetch chain for %s: %v\n", decisionID, err)
			brokenDecisions++
			continue
		}

		var prevHash []byte
		chainValid := true

		for i, entry := range chain {
			totalRevisions++
			// Revision 1's predecessor is the genesis root; its
			// previous_revision_hash is the zero hash, which is
			// expected. Every subsequent revision must point at its
			// predecessor.
			if i == 0 {
				prevHash = entry.DecisionHash
				continue
			}
			if string(entry.PreviousRevisionHash) != string(prevHash) {
				fmt.Printf("❌ Chain broken at revision %s (decision %s, revision #%d)\n",
					entry.ID, decisionID, i+1)
				fmt.Printf("   Expected prev: %x\n", prevHash)
				fmt.Printf("   Actual prev:   %x\n", entry.PreviousRevisionHash)
				chainValid = false
				brokenDecisions++
				break
			}
			prevHash = entry.DecisionHash
		}

		if chainValid {
			fmt.Printf("✅ Decision %s: %d revision(s) verified.\n", decisionID, len(chain))
		}
	}

	fmt.Println()
	if brokenDecisions > 0 {
		fmt.Printf("❌ %d decision chain(s) broken out of %d. %d revision(s) examined.\n",
			brokenDecisions, len(decisionIDs), totalRevisions)
		os.Exit(1)
	}
	fmt.Printf("✅ All %d decision chain(s) intact. %d revision(s) verified.\n",
		len(decisionIDs), totalRevisions)
}

func statusText(valid bool) string {
	if valid {
		return "VALID "
	}
	return "INVALID"
}

func handleExplain(decisionIDStr string) {
	decisionID, err := uuid.Parse(decisionIDStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Invalid decision ID '%s': %v\n", decisionIDStr, err)
		os.Exit(1)
	}

	dbURL := getDBURL()
	st, err := store.NewPostgresStore(dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	ctx := context.Background()
	tid := getTenantID()

	exp, err := st.GetDecisionExplanation(ctx, tid, decisionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Decision not found: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📄 DECISION EXPLANATION\n")
	fmt.Printf("Decision ID:       %s\n", exp.ID)
	fmt.Printf("Status:            %s\n", exp.Status)
	fmt.Printf("Title:             %s\n", exp.Title)
	fmt.Printf("Statement:         %s\n", exp.Statement)
	fmt.Printf("Scope:             %s\n", string(exp.ScopeJSON))
	fmt.Printf("Owner:             %s\n", exp.Owner)
	fmt.Printf("Confidence:        %.2f\n", exp.Confidence)
	fmt.Printf("Revision:          #%d\n", exp.RevisionNumber)
	fmt.Printf("Created:           %s\n", exp.RevisionCreatedAt.Format(time.RFC3339))
	fmt.Printf("Content Hash:      %x\n", exp.DecisionHash)
	fmt.Printf("Previous Revision: %x\n", exp.PreviousRevisionHash)
	fmt.Printf("Merkle Root:       %x\n", exp.MerkleRoot)

	var summary analyzer.RevisionSummary
	if err := json.Unmarshal(exp.CanonicalJSON, &summary); err == nil && summary.Fingerprint != "" {
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
