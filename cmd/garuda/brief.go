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

	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/spf13/cobra"
)

var briefJSONFlag bool

func init() {
	briefCmd.Flags().BoolVarP(&briefJSONFlag, "json", "j", false, "Output briefing in structured JSON format")
}

var briefCmd = &cobra.Command{
	Use:   "brief",
	Short: "Session-start briefing: workspace state, trust anchor, scale, hubs, policies, attention, and what changed",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		dbURL := getDBURL()
		st, err := store.NewPostgresStore(dbURL)
		if err != nil {
			return fmt.Errorf("failed to connect to store: %w", err)
		}
		defer st.Close()

		tenantID := getTenantID()
		workspaceName := getWorkspaceName()

		workspaceID, resolvedName, err := st.ResolveWorkspaceTarget(ctx, tenantID, workspaceName)
		if err != nil {
			return err
		}
		workspaceName = resolvedName

		// The CLI brief is a single-shot call: no per-agent watermark
		// is kept, so the diff section always reports "no prior
		// briefing". The MCP tool is where the per-agent watermark
		// matters; the CLI is for a one-time operator view.
		agentID := "cli-operator"
		sessionID := "cli-brief"

		briefing, err := st.GetBriefing(ctx, tenantID, workspaceID, workspaceName, agentID, sessionID)
		if err != nil {
			return fmt.Errorf("build briefing: %w", err)
		}

		if briefJSONFlag {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(briefing)
		}

		fmt.Printf("\n🌐 GARUDA BRIEFING: %s\n", briefing.Workspace)
		fmt.Println(strings.Repeat("━", 65))
		fmt.Printf("Session:  %s\n", briefing.SessionID)
		fmt.Printf("Agent:    %s\n\n", briefing.AgentID)

		fmt.Println("🔒 Trust")
		fmt.Printf("   • Status:      %s\n", briefing.Trust.Status)
		fmt.Printf("   • Block:       #%d\n", briefing.Trust.BlockHeight)
		if briefing.Trust.MerkleRoot != "" {
			root := briefing.Trust.MerkleRoot
			if len(root) > 16 {
				root = root[:16] + "..."
			}
			fmt.Printf("   • Merkle root: %s\n", root)
		}

		fmt.Println("\n📊 Scale")
		fmt.Printf("   • Repositories:        %d\n", briefing.Scale.Repositories)
		fmt.Printf("   • Entities:            %d\n", briefing.Scale.Entities)
		fmt.Printf("   • Claims:              %d\n", briefing.Scale.Claims)
		fmt.Printf("   • Cross-repo bridges:  %d\n", briefing.Scale.CrossRepoBridges)

		if len(briefing.Hubs) > 0 {
			fmt.Println("\n🏛️  Architectural hubs")
			for _, h := range briefing.Hubs {
				shortPkg := h.Package
				if parts := strings.Split(h.Package, "/"); len(parts) > 0 {
					shortPkg = parts[len(parts)-1]
				}
				fmt.Printf("   • %s.%s (%s) ── %d inbound edges\n", shortPkg, h.Name, h.Kind, h.Callers)
			}
		}

		fmt.Println("\n🛡️  Policies")
		fmt.Printf("   • Active: %d\n", briefing.Policies.ActiveCount)
		for _, t := range briefing.Policies.Titles {
			fmt.Printf("     - %s\n", t)
		}

		fmt.Println("\n⚠️  Attention")
		fmt.Printf("   • Open contradictions:   %d\n", briefing.Attention.OpenContradictions)
		fmt.Printf("   • Undocumented entities: %d\n", briefing.Attention.UndocumentedEntities)
		fmt.Printf("   • Drift findings:        %d\n", briefing.Attention.DriftFindings)

		fmt.Println("\n🆕 New since last briefing")
		if !briefing.NewSince.HasPriorBriefing {
			fmt.Println("   • No prior briefing on record.")
		} else {
			fmt.Printf("   • Claims added:        %d\n", briefing.NewSince.ClaimsAdded)
			fmt.Printf("   • Verifications since: %d\n", briefing.NewSince.VerificationsSince)
		}
		fmt.Println(strings.Repeat("━", 65))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(briefCmd)
}
