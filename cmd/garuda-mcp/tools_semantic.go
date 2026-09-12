// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/analyzer"
)

// handleEntities lists entities in a workspace, optionally filtered by
// package or kind.
//
// Read-only. Uses the same store API the dashboard's semantic panels
// use. Does not modify the graph.
func (s *MCPServer) handleEntities(args map[string]interface{}) (interface{}, error) {
	tenantID, workspace, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}

	var workspaceID uuid.UUID
	if err := s.store.Pool().QueryRow(context.Background(), `
		SELECT id FROM workspaces WHERE name = $1 LIMIT 1
	`, workspace).Scan(&workspaceID); err != nil {
		return nil, fmt.Errorf("workspace %q not found: %w", workspace, err)
	}

	pkgFilter, _ := args["package"].(string)
	kindFilter, _ := args["kind"].(string)
	limit := 100
	if v, ok := args["limit"].(float64); ok && v > 0 && v <= 1000 {
		limit = int(v)
	}

	q := `
		SELECT id::text, name, kind, COALESCE(package, ''), COALESCE(file_path, ''),
		       COALESCE(line_start, 0), is_exported
		FROM entities
		WHERE workspace_id = $1 AND kind != 'external'
	`
	dbArgs := []interface{}{workspaceID}
	argPos := 2

	if pkgFilter != "" {
		q += fmt.Sprintf(" AND package LIKE '%%' || $%d || '%%'", argPos)
		dbArgs = append(dbArgs, pkgFilter)
		argPos++
	}
	if kindFilter != "" {
		q += fmt.Sprintf(" AND kind = $%d", argPos)
		dbArgs = append(dbArgs, kindFilter)
		argPos++
	}
	q += fmt.Sprintf(" ORDER BY package, name LIMIT $%d", argPos)
	dbArgs = append(dbArgs, limit)

	rows, err := s.store.Pool().Query(context.Background(), q, dbArgs...)
	if err != nil {
		return nil, fmt.Errorf("query entities: %w", err)
	}
	defer rows.Close()

	type entityRow struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Kind      string `json:"kind"`
		Package   string `json:"package"`
		File      string `json:"file"`
		LineStart int    `json:"line_start"`
		Exported  bool   `json:"exported"`
	}

	entities := make([]entityRow, 0, limit)
	for rows.Next() {
		var e entityRow
		if err := rows.Scan(&e.ID, &e.Name, &e.Kind, &e.Package, &e.File, &e.LineStart, &e.Exported); err != nil {
			continue
		}
		entities = append(entities, e)
	}

	return map[string]interface{}{
		"tenant_id":  tenantID.String(),
		"workspace":  workspace,
		"filters":    map[string]any{"package": pkgFilter, "kind": kindFilter},
		"count":      len(entities),
		"entities":   entities,
		"limit_used": limit,
	}, nil
}

// handleInspect returns one entity and its incoming and outgoing
// relationships.
//
// Read-only. Uses GetEntityRelationships, which reads from claims.
func (s *MCPServer) handleInspect(args map[string]interface{}) (interface{}, error) {
	tenantID, workspace, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}

	entityIDStr, _ := args["entity_id"].(string)
	if entityIDStr == "" {
		return nil, fmt.Errorf("entity_id is required")
	}
	if _, err := uuid.Parse(entityIDStr); err != nil {
		return nil, fmt.Errorf("invalid entity_id %q: %w", entityIDStr, err)
	}

	var workspaceID uuid.UUID
	if err := s.store.Pool().QueryRow(context.Background(), `
		SELECT id FROM workspaces WHERE name = $1 LIMIT 1
	`, workspace).Scan(&workspaceID); err != nil {
		return nil, fmt.Errorf("workspace %q not found: %w", workspace, err)
	}

	// Verify the entity exists and fetch its basic fields.
	var entity struct {
		ID        string
		Name      string
		Kind      string
		Package   string
		File      string
		LineStart int
		Exported  bool
	}
	err = s.store.Pool().QueryRow(context.Background(), `
		SELECT id::text, name, kind, COALESCE(package, ''), COALESCE(file_path, ''),
		       COALESCE(line_start, 0), is_exported
		FROM entities
		WHERE workspace_id = $1 AND id = $2
	`, workspaceID, entityIDStr).Scan(
		&entity.ID, &entity.Name, &entity.Kind, &entity.Package,
		&entity.File, &entity.LineStart, &entity.Exported,
	)
	if err != nil {
		return nil, fmt.Errorf("entity %q not found in workspace %q", entityIDStr, workspace)
	}

	incoming, outgoing, err := s.store.GetEntityRelationships(
		context.Background(), tenantID, workspaceID, entityIDStr,
	)
	if err != nil {
		return nil, fmt.Errorf("fetch relationships: %w", err)
	}

	type relRow struct {
		From string `json:"from"`
		To   string `json:"to"`
		Type string `json:"type"`
		File string `json:"file,omitempty"`
		Line int    `json:"line,omitempty"`
	}

	toRows := func(rels interface{ Len() int }) []relRow { return nil }
	_ = toRows // placeholder to keep imports clean; real conversion below

	convert := func(rs []analyzer.Relationship) []relRow {
		out := make([]relRow, 0, len(rs))
		for _, r := range rs {
			out = append(out, relRow{
				From: r.From,
				To:   r.To,
				Type: r.Type,
				File: r.Evidence.File,
				Line: r.Evidence.LineStart,
			})
		}
		return out
	}

	return map[string]interface{}{
		"tenant_id":      tenantID.String(),
		"workspace":      workspace,
		"entity":         entity,
		"incoming_count": len(incoming),
		"outgoing_count": len(outgoing),
		"incoming":       convert(incoming),
		"outgoing":       convert(outgoing),
	}, nil
}
