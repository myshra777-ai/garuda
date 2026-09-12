// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/tenant"
	"github.com/spf13/cobra"
)

var justifyCmd = &cobra.Command{
	Use:   "justify [entity-id-or-name]",
	Short: "Justify semantic relationships and provenance for a given entity",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		target := strings.TrimSpace(args[0])

		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			dbURL = "postgres://test:test@localhost:5433/garuda_test?sslmode=disable"
		}

		st, err := store.NewPostgresStore(dbURL)
		if err != nil {
			return fmt.Errorf("failed to connect to store: %w", err)
		}
		defer st.Close()

		tenantStr := os.Getenv("GARUDA_TENANT_ID")
		if tenantStr == "" {
			tenantStr = tenant.CanonicalIDStr
		}
		tenantID, err := uuid.Parse(tenantStr)
		if err != nil {
			return fmt.Errorf("invalid tenant ID: %w", err)
		}

		workspaceName := os.Getenv("GARUDA_WORKSPACE")
		if workspaceName == "" {
			workspaceName = "default"
		}

		ws, err := st.GetWorkspaceByName(ctx, tenantID.String(), workspaceName)
		if err != nil {
			return fmt.Errorf("failed to resolve workspace '%s': %w", workspaceName, err)
		}
		workspaceID := ws.ID

		var entityID string
		var entityName string

		if targetUUID, err := uuid.Parse(target); err == nil {
			entity, err := st.GetEntityByID(ctx, tenantID, workspaceID, targetUUID)
			if err != nil {
				return fmt.Errorf("entity ID %s not found: %w", targetUUID, err)
			}
			entityID = entity.ID
			entityName = entity.Name
		} else {
			parts := strings.Split(target, ".")
			if len(parts) < 2 {
				return fmt.Errorf("invalid entity format: must be UUID or 'pkg.Name'")
			}
			pkg := strings.Join(parts[:len(parts)-1], ".")
			name := parts[len(parts)-1]

			entity, err := st.GetEntity(ctx, tenantID, workspaceID, pkg, name)
			if err != nil {
				return fmt.Errorf("failed to find entity '%s.%s': %w", pkg, name, err)
			}
			entityID = entity.ID
			entityName = entity.Name
		}

		incoming, outgoing, err := st.GetEntityRelationships(ctx, tenantID, workspaceID, entityID)
		if err != nil {
			return fmt.Errorf("failed to fetch relationships: %w", err)
		}

		fmt.Printf("⚖️  JUSTIFICATION REPORT FOR: %s (%s)\n", entityName, entityID)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("📥 Incoming Claims (%d):\n", len(incoming))
		for _, in := range incoming {
			fmt.Printf("   • From: %-36s | Type: %-12s | Conf: %.2f\n", in.From, in.Type, in.Confidence)
		}

		fmt.Printf("\n📤 Outgoing Claims (%d):\n", len(outgoing))
		for _, out := range outgoing {
			fmt.Printf("   • To:   %-36s | Type: %-12s | Conf: %.2f\n", out.To, out.Type, out.Confidence)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(justifyCmd)
}
