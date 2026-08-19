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
	"github.com/myshra777-ai/garuda/internal/analyzer"
	"github.com/myshra777-ai/garuda/internal/store"
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
			tenantStr = "00000000-0000-0000-0000-000000000001"
		}
		tenantID, err := uuid.Parse(tenantStr)
		if err != nil {
			return fmt.Errorf("invalid tenant ID: %w", err)
		}

		workspaceName := os.Getenv("GARUDA_WORKSPACE")
		if workspaceName == "" {
			workspaceName = "default"
		}

		var workspaceID uuid.UUID
		err = st.Pool().QueryRow(ctx, "SELECT id FROM workspaces WHERE name = $1 LIMIT 1", workspaceName).Scan(&workspaceID)
		if err != nil {
			return fmt.Errorf("failed to resolve workspace '%s': %w", workspaceName, err)
		}

		var entityID string
		var entityName string

		if targetUUID, err := uuid.Parse(target); err == nil {
			var entity analyzer.Entity
			var id, kind, pkgName, pkgPath, modPath, recType, file, signature string
			var exported bool
			var fieldsJSON, methodsJSON []byte
			var line, lineStart, lineEnd int

			err := st.Pool().QueryRow(ctx, `
				SELECT id, name, kind, package, package_path, module_path, receiver_type,
				       file_path, signature, fields, methods, is_exported,
				       line, line_start, line_end
				FROM entities
				WHERE tenant_id = $1 AND workspace_id = $2 AND id = $3
				LIMIT 1
			`, tenantID, workspaceID, targetUUID).Scan(
				&id, &entity.Name, &kind, &pkgName, &pkgPath, &modPath, &recType,
				&file, &signature, &fieldsJSON, &methodsJSON, &exported,
				&line, &lineStart, &lineEnd,
			)
			if err != nil {
				return fmt.Errorf("entity ID %s not found: %w", targetUUID, err)
			}
			entityID = id
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
