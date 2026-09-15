// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

// tools_graph.go contains the four tools that operate on graph
// structure: inheritance, implementation, one-hop neighborhood, and
// name-based lookup.
//
// The four tools are read-only. Their queries are the ones the
// dashboard's semantic panels run; the MCP surface just exposes them
// without the HTTP wrapper.
//
// A note on what these tools can and cannot return:
//
//   - INHERITS and IMPLEMENTS are dense in this codebase. A query
//     for subclasses of SQLModel returns 171 rows; a query for
//     implementers of CmdTyper returns 97. These tools produce
//     useful answers on day one.
//
//   - CALLS is sparse at the leaf. The median inbound CALLS count
//     across the workspace is 1, and 90% of targets have 4 or fewer.
//     A dedicated "callers" tool would return a single edge for
//     most entities and produce misleading results. Neighbors is
//     the tool that covers calls, imports, references, and defines
//     together, and is honest about the shape of the data.
//
// See docs/ROADMAP.md for the CALLS density audit that informs this.

// resolveSubjectEntity turns a caller's (entity_id | name) into a
// resolved UUID and canonical name. Either input may be provided; if
// both are provided, entity_id wins.
//
// Returns an error if neither is provided, or if the entity cannot
// be found by the given identifier.
func (s *MCPServer) resolveSubjectEntity(
	ctx context.Context,
	workspaceID uuid.UUID,
	entityIDStr, nameStr string,
) (uuid.UUID, string, error) {
	if entityIDStr != "" {
		parsed, err := uuid.Parse(entityIDStr)
		if err != nil {
			return uuid.Nil, "", fmt.Errorf("invalid entity_id %q: %w", entityIDStr, err)
		}
		var name string
		err = s.store.Pool().QueryRow(ctx, `
			SELECT name FROM entities WHERE workspace_id = $1 AND id = $2
		`, workspaceID, parsed).Scan(&name)
		if err != nil {
			return uuid.Nil, "", fmt.Errorf("entity_id %q not found in this workspace", entityIDStr)
		}
		return parsed, name, nil
	}
	if nameStr != "" {
		var id uuid.UUID
		var name string
		err := s.store.Pool().QueryRow(ctx, `
			SELECT id, name FROM entities
			 WHERE workspace_id = $1 AND name = $2 AND kind != 'external'
			 ORDER BY is_exported DESC
			 LIMIT 1
		`, workspaceID, nameStr).Scan(&id, &name)
		if err != nil {
			return uuid.Nil, "", fmt.Errorf("no entity named %q in this workspace", nameStr)
		}
		return id, name, nil
	}
	return uuid.Nil, "", fmt.Errorf("one of entity_id or name is required")
}

// handleSubclasses returns entities that inherit from or embed the
// subject.
//
// Uses claims where claim_type is INHERITS (Go struct embedding,
// Python class inheritance, TypeScript extends) or EMBEDS (Go
// interface embedding). The subject is the target of the claim.
//
// Read-only.
func (s *MCPServer) handleSubclasses(args map[string]interface{}) (interface{}, error) {
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	workspaceName, _ := args["workspace"].(string)

	entityIDStr, _ := args["entity_id"].(string)
	nameStr, _ := args["name"].(string)
	subjectID, subjectName, err := s.resolveSubjectEntity(
		context.Background(), workspaceID, entityIDStr, nameStr)
	if err != nil {
		return nil, err
	}

	languageFilter, _ := args["language"].(string)
	limit := 50
	if v, ok := args["limit"].(float64); ok && v > 0 && v <= 500 {
		limit = int(v)
	}

	q := `
		SELECT e.id::text, e.name, e.kind, COALESCE(e.package, ''), COALESCE(e.language, ''),
		       COALESCE(e.file_path, ''), COALESCE(e.line_start, 0),
		       COALESCE(r.name, 'unknown') AS repo,
		       c.claim_type, c.confidence,
		       e.is_exported
		  FROM claims c
		  JOIN entities e ON e.id = c.from_entity_id
		  LEFT JOIN repositories r ON r.id = e.repository_id
		 WHERE c.workspace_id = $1
		   AND c.claim_type IN ('INHERITS', 'EMBEDS')
		   AND c.to_entity_id = $2
		   AND e.kind != 'external'
	`
	dbArgs := []interface{}{workspaceID, subjectID}
	argPos := 3

	if languageFilter != "" {
		q += fmt.Sprintf(" AND e.language = $%d", argPos)
		dbArgs = append(dbArgs, languageFilter)
		argPos++
	}
	q += fmt.Sprintf(" ORDER BY e.name LIMIT $%d", argPos)
	dbArgs = append(dbArgs, limit)

	rows, err := s.store.Pool().Query(context.Background(), q, dbArgs...)
	if err != nil {
		return nil, fmt.Errorf("query subclasses: %w", err)
	}
	defer rows.Close()

	type subclassRow struct {
		ID         string  `json:"id"`
		Name       string  `json:"name"`
		Kind       string  `json:"kind"`
		Package    string  `json:"package"`
		Language   string  `json:"language"`
		File       string  `json:"file"`
		LineStart  int     `json:"line_start"`
		Repo       string  `json:"repo"`
		ClaimType  string  `json:"claim_type"`
		Confidence float64 `json:"confidence"`
		Exported   bool    `json:"exported"`
	}

	results := make([]subclassRow, 0, limit)
	for rows.Next() {
		var r subclassRow
		if err := rows.Scan(
			&r.ID, &r.Name, &r.Kind, &r.Package, &r.Language,
			&r.File, &r.LineStart, &r.Repo,
			&r.ClaimType, &r.Confidence, &r.Exported,
		); err != nil {
			continue
		}
		results = append(results, r)
	}

	return map[string]interface{}{
		"tenant_id":      tenantID.String(),
		"workspace_id":   workspaceID.String(),
		"workspace_name": workspaceName,
		"subject":        map[string]string{"id": subjectID.String(), "name": subjectName},
		"filters":        map[string]any{"language": languageFilter},
		"count":          len(results),
		"results":        results,
		"limit_used":     limit,
	}, nil
}

// handleImplementers returns entities that implement the subject
// interface.
//
// Uses claims where claim_type is IMPLEMENTS. The subject is the
// target of the claim.
//
// Read-only.
func (s *MCPServer) handleImplementers(args map[string]interface{}) (interface{}, error) {
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	workspaceName, _ := args["workspace"].(string)

	entityIDStr, _ := args["entity_id"].(string)
	nameStr, _ := args["name"].(string)
	subjectID, subjectName, err := s.resolveSubjectEntity(
		context.Background(), workspaceID, entityIDStr, nameStr)
	if err != nil {
		return nil, err
	}

	limit := 50
	if v, ok := args["limit"].(float64); ok && v > 0 && v <= 500 {
		limit = int(v)
	}

	rows, err := s.store.Pool().Query(context.Background(), `
		SELECT e.id::text, e.name, e.kind, COALESCE(e.package, ''), COALESCE(e.language, ''),
		       COALESCE(e.file_path, ''), COALESCE(e.line_start, 0),
		       COALESCE(r.name, 'unknown') AS repo,
		       c.confidence,
		       e.is_exported
		  FROM claims c
		  JOIN entities e ON e.id = c.from_entity_id
		  LEFT JOIN repositories r ON r.id = e.repository_id
		 WHERE c.workspace_id = $1
		   AND c.claim_type = 'IMPLEMENTS'
		   AND c.to_entity_id = $2
		   AND e.kind != 'external'
		 ORDER BY e.name
		 LIMIT $3
	`, workspaceID, subjectID, limit)
	if err != nil {
		return nil, fmt.Errorf("query implementers: %w", err)
	}
	defer rows.Close()

	type implRow struct {
		ID         string  `json:"id"`
		Name       string  `json:"name"`
		Kind       string  `json:"kind"`
		Package    string  `json:"package"`
		Language   string  `json:"language"`
		File       string  `json:"file"`
		LineStart  int     `json:"line_start"`
		Repo       string  `json:"repo"`
		Confidence float64 `json:"confidence"`
		Exported   bool    `json:"exported"`
	}

	results := make([]implRow, 0, limit)
	for rows.Next() {
		var r implRow
		if err := rows.Scan(
			&r.ID, &r.Name, &r.Kind, &r.Package, &r.Language,
			&r.File, &r.LineStart, &r.Repo,
			&r.Confidence, &r.Exported,
		); err != nil {
			continue
		}
		results = append(results, r)
	}

	return map[string]interface{}{
		"tenant_id":      tenantID.String(),
		"workspace_id":   workspaceID.String(),
		"workspace_name": workspaceName,
		"subject":        map[string]string{"id": subjectID.String(), "name": subjectName},
		"count":          len(results),
		"results":        results,
		"limit_used":     limit,
	}, nil
}

// handleNeighbors returns every entity connected to the subject in
// either direction by a single edge.
//
// This is the tool that replaces the pattern "call inspect ten times
// and merge in context". It returns both directions in one query,
// with names and packages resolved, and edge kinds filterable.
//
// External stubs are excluded by default. An external stub is an
// unresolved reference the analyzer emitted when it could not map a
// call or import to a real entity. Including them is useful for
// investigating what an entity depends on outside the workspace, but
// they dominate the row count and hide the real neighbors.
//
// Read-only.
func (s *MCPServer) handleNeighbors(args map[string]interface{}) (interface{}, error) {
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	workspaceName, _ := args["workspace"].(string)

	entityIDStr, _ := args["entity_id"].(string)
	if entityIDStr == "" {
		return nil, fmt.Errorf("entity_id is required")
	}
	subjectID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid entity_id %q: %w", entityIDStr, err)
	}

	var subjectName string
	err = s.store.Pool().QueryRow(context.Background(), `
		SELECT name FROM entities WHERE workspace_id = $1 AND id = $2
	`, workspaceID, subjectID).Scan(&subjectName)
	if err != nil {
		return nil, fmt.Errorf("entity_id %q not found in this workspace", entityIDStr)
	}

	includeExternal := false
	if v, ok := args["include_external"].(bool); ok {
		includeExternal = v
	}

	var edgeKinds []string
	if raw, ok := args["edge_kinds"].([]interface{}); ok {
		for _, k := range raw {
			if sv, ok := k.(string); ok && sv != "" {
				edgeKinds = append(edgeKinds, sv)
			}
		}
	}

	limit := 100
	if v, ok := args["limit"].(float64); ok && v > 0 && v <= 500 {
		limit = int(v)
	}

	// UNION ALL of outgoing and incoming edges. The result carries a
	// direction column so the caller can tell "what this calls" from
	// "what calls this" without a second query.
	q := `
		WITH edges AS (
			SELECT c.to_entity_id AS other_id,
			       c.claim_type, c.confidence, 'outgoing' AS direction
			  FROM claims c
			 WHERE c.workspace_id = $1 AND c.from_entity_id = $2
			UNION ALL
			SELECT c.from_entity_id AS other_id,
			       c.claim_type, c.confidence, 'incoming' AS direction
			  FROM claims c
			 WHERE c.workspace_id = $1 AND c.to_entity_id = $2
		)
		SELECT e.id::text, e.name, e.kind, COALESCE(e.package, ''), COALESCE(e.language, ''),
		       COALESCE(e.file_path, ''), COALESCE(e.line_start, 0),
		       COALESCE(r.name, 'unknown') AS repo,
		       ed.claim_type, ed.confidence, ed.direction,
		       e.is_exported
		  FROM edges ed
		  JOIN entities e ON e.id = ed.other_id
		  LEFT JOIN repositories r ON r.id = e.repository_id
		 WHERE 1=1
	`
	dbArgs := []interface{}{workspaceID, subjectID}
	argPos := 3

	if !includeExternal {
		q += " AND e.kind != 'external'"
	}
	if len(edgeKinds) > 0 {
		placeholders := ""
		for i, k := range edgeKinds {
			if i > 0 {
				placeholders += ","
			}
			placeholders += fmt.Sprintf("$%d", argPos)
			dbArgs = append(dbArgs, k)
			argPos++
		}
		q += fmt.Sprintf(" AND ed.claim_type IN (%s)", placeholders)
	}
	q += fmt.Sprintf(" ORDER BY ed.direction, ed.claim_type, e.name LIMIT $%d", argPos)
	dbArgs = append(dbArgs, limit)

	rows, err := s.store.Pool().Query(context.Background(), q, dbArgs...)
	if err != nil {
		return nil, fmt.Errorf("query neighbors: %w", err)
	}
	defer rows.Close()

	type neighborRow struct {
		ID         string  `json:"id"`
		Name       string  `json:"name"`
		Kind       string  `json:"kind"`
		Package    string  `json:"package"`
		Language   string  `json:"language"`
		File       string  `json:"file"`
		LineStart  int     `json:"line_start"`
		Repo       string  `json:"repo"`
		ClaimType  string  `json:"claim_type"`
		Confidence float64 `json:"confidence"`
		Direction  string  `json:"direction"`
		Exported   bool    `json:"exported"`
	}

	results := make([]neighborRow, 0, limit)
	for rows.Next() {
		var r neighborRow
		if err := rows.Scan(
			&r.ID, &r.Name, &r.Kind, &r.Package, &r.Language,
			&r.File, &r.LineStart, &r.Repo,
			&r.ClaimType, &r.Confidence, &r.Direction, &r.Exported,
		); err != nil {
			continue
		}
		results = append(results, r)
	}

	return map[string]interface{}{
		"tenant_id":      tenantID.String(),
		"workspace_id":   workspaceID.String(),
		"workspace_name": workspaceName,
		"subject":        map[string]string{"id": subjectID.String(), "name": subjectName},
		"filters": map[string]any{
			"edge_kinds":       edgeKinds,
			"include_external": includeExternal,
		},
		"count":      len(results),
		"results":    results,
		"limit_used": limit,
	}, nil
}

// handleFindEntity returns entities matching the given filters.
//
// Results are ranked by inbound edge count when more matches exist
// than the limit. An entity with 397 inbound edges is more likely to
// be what the caller wants than an entity with 1, when both match
// the same name pattern.
//
// The rank query counts every claim that targets the entity,
// regardless of claim_type. That is intentional: an interface with
// 97 inbound IMPLEMENTS is as important to surface as a method with
// 397 inbound CALLS.
//
// Read-only.
func (s *MCPServer) handleFindEntity(args map[string]interface{}) (interface{}, error) {
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	workspaceName, _ := args["workspace"].(string)

	namePattern, _ := args["name_pattern"].(string)
	kindFilter, _ := args["kind"].(string)
	packageFilter, _ := args["package"].(string)
	fileFilter, _ := args["file_path_contains"].(string)
	exportedOnly, _ := args["exported_only"].(bool)

	if namePattern != "" {
		if _, err := regexp.Compile(namePattern); err != nil {
			return nil, fmt.Errorf("invalid name_pattern %q: %w", namePattern, err)
		}
	}

	limit := 50
	if v, ok := args["limit"].(float64); ok && v > 0 && v <= 500 {
		limit = int(v)
	}

	// Rank by inbound edge count. The subquery groups claims by
	// target once, then the outer join brings the count in.
	q := `
		SELECT e.id::text, e.name, e.kind, COALESCE(e.package, ''), COALESCE(e.language, ''),
		       COALESCE(e.file_path, ''), COALESCE(e.line_start, 0),
		       COALESCE(r.name, 'unknown') AS repo,
		       e.is_exported,
		       COALESCE(inbound.cnt, 0) AS inbound_edges
		  FROM entities e
		  LEFT JOIN repositories r ON r.id = e.repository_id
		  LEFT JOIN (
		      SELECT to_entity_id, COUNT(*)::int AS cnt
		        FROM claims
		       WHERE workspace_id = $1
		       GROUP BY to_entity_id
		  ) inbound ON inbound.to_entity_id = e.id
		 WHERE e.workspace_id = $1
		   AND e.kind != 'external'
	`
	dbArgs := []interface{}{workspaceID}
	argPos := 2

	if namePattern != "" {
		q += fmt.Sprintf(" AND e.name ~ $%d", argPos)
		dbArgs = append(dbArgs, namePattern)
		argPos++
	}
	if kindFilter != "" {
		q += fmt.Sprintf(" AND e.kind = $%d", argPos)
		dbArgs = append(dbArgs, kindFilter)
		argPos++
	}
	if packageFilter != "" {
		q += fmt.Sprintf(" AND e.package LIKE '%%' || $%d || '%%'", argPos)
		dbArgs = append(dbArgs, packageFilter)
		argPos++
	}
	if fileFilter != "" {
		q += fmt.Sprintf(" AND e.file_path LIKE '%%' || $%d || '%%'", argPos)
		dbArgs = append(dbArgs, fileFilter)
		argPos++
	}
	if exportedOnly {
		q += " AND e.is_exported = true"
	}
	q += fmt.Sprintf(" ORDER BY inbound_edges DESC, e.name LIMIT $%d", argPos)
	dbArgs = append(dbArgs, limit)

	rows, err := s.store.Pool().Query(context.Background(), q, dbArgs...)
	if err != nil {
		return nil, fmt.Errorf("query entities: %w", err)
	}
	defer rows.Close()

	type entityRow struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Kind         string `json:"kind"`
		Package      string `json:"package"`
		Language     string `json:"language"`
		File         string `json:"file"`
		LineStart    int    `json:"line_start"`
		Repo         string `json:"repo"`
		Exported     bool   `json:"exported"`
		InboundEdges int    `json:"inbound_edges"`
	}

	results := make([]entityRow, 0, limit)
	for rows.Next() {
		var r entityRow
		if err := rows.Scan(
			&r.ID, &r.Name, &r.Kind, &r.Package, &r.Language,
			&r.File, &r.LineStart, &r.Repo,
			&r.Exported, &r.InboundEdges,
		); err != nil {
			continue
		}
		results = append(results, r)
	}

	return map[string]interface{}{
		"tenant_id":      tenantID.String(),
		"workspace_id":   workspaceID.String(),
		"workspace_name": workspaceName,
		"filters": map[string]any{
			"name_pattern":       namePattern,
			"kind":               kindFilter,
			"package":            packageFilter,
			"file_path_contains": fileFilter,
			"exported_only":      exportedOnly,
		},
		"count":      len(results),
		"results":    results,
		"limit_used": limit,
	}, nil
}
