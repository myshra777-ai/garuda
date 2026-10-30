// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// resolveSubjectEntity turns a caller's (entity_id | name) into a
// resolved UUID and canonical name. Either input may be provided; if
// both are provided, entity_id wins.
func (s *MCPServer) resolveSubjectEntity(
	ctx context.Context,
	workspaceID uuid.UUID,
	entityIDStr, nameStr string,
) (uuid.UUID, string, error) {
	if s.store == nil || s.store.Pool() == nil {
		return uuid.Nil, nameStr, fmt.Errorf("database not connected")
	}
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
		if err == nil {
			return id, name, nil
		}
		// Fallback to case-insensitive or qualified suffix match (e.g. "Router" -> "chi.Router")
		err = s.store.Pool().QueryRow(ctx, `
			SELECT id, name FROM entities
			 WHERE workspace_id = $1
			   AND (name ILIKE $2 OR name ILIKE '%.' || $2)
			   AND kind != 'external'
			 ORDER BY is_exported DESC
			 LIMIT 1
		`, workspaceID, nameStr).Scan(&id, &name)
		if err == nil {
			return id, name, nil
		}
		return uuid.Nil, "", fmt.Errorf("no entity named %q in this workspace", nameStr)
	}
	return uuid.Nil, "", fmt.Errorf("one of entity_id or name is required")
}

// resolveTopClaimTarget finds the entity in workspaceID with the most incoming
// claims of the given claimTypes. Used ONLY when a caller invokes subclasses or
// implementers without specifying any entity_id or name.
func (s *MCPServer) resolveTopClaimTarget(
	ctx context.Context,
	workspaceID uuid.UUID,
	claimTypes []string,
) (uuid.UUID, string, error) {
	var id uuid.UUID
	var name string
	err := s.store.Pool().QueryRow(ctx, `
		SELECT e.id, e.name
		  FROM claims c
		  JOIN entities e ON e.id = c.to_entity_id
		 WHERE c.workspace_id = $1
		   AND c.claim_type = ANY($2)
		   AND e.kind != 'external'
		 GROUP BY e.id, e.name
		 ORDER BY COUNT(*) DESC, e.name ASC
		 LIMIT 1
	`, workspaceID, claimTypes).Scan(&id, &name)
	if err != nil {
		return uuid.Nil, "", err
	}
	return id, name, nil
}

// handleSubclasses returns entities that inherit from or embed the subject (BUG-18).
// Never silently substitutes an unrelated hub when a specific entity_id or name is requested,
// and always reports total_count and truncated.
func (s *MCPServer) handleSubclasses(args map[string]interface{}) (interface{}, error) {
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	workspaceName := s.resolveMCPWorkspaceName(args, workspaceID)

	entityIDStr, _ := args["entity_id"].(string)
	nameStr, _ := args["name"].(string)
	entityIDStr = strings.TrimSpace(entityIDStr)
	nameStr = strings.TrimSpace(nameStr)
	explicitSubject := entityIDStr != "" || nameStr != ""
	ctx := context.Background()

	limit := 50
	if v, ok := args["limit"].(float64); ok && v > 0 && v <= 500 {
		limit = int(v)
	}
	languageFilter, _ := args["language"].(string)

	claimTypes := []string{"INHERITS", "EMBEDS"}
	var subjectID uuid.UUID
	var subjectName string

	if explicitSubject {
		subjectID, subjectName, err = s.resolveSubjectEntity(ctx, workspaceID, entityIDStr, nameStr)
		if err != nil {
			reqTarget := nameStr
			if reqTarget == "" {
				reqTarget = entityIDStr
			}
			return map[string]interface{}{
				"tenant_id":      tenantID.String(),
				"workspace_id":   workspaceID.String(),
				"workspace_name": workspaceName,
				"status":         "not_found",
				"not_found":      true,
				"subject":        map[string]string{"id": "", "name": reqTarget},
				"count":          0,
				"total_count":    0,
				"truncated":      false,
				"results":        []interface{}{},
				"limit_used":     limit,
				"hint":           fmt.Sprintf("no entity matching %q found in workspace %s", reqTarget, workspaceName),
			}, nil
		}
	} else {
		if fbID, fbName, fbErr := s.resolveTopClaimTarget(ctx, workspaceID, []string{"INHERITS", "EMBEDS"}); fbErr == nil {
			subjectID = fbID
			subjectName = fbName
		} else if fbID2, fbName2, fbErr2 := s.resolveTopClaimTarget(ctx, workspaceID, []string{"IMPLEMENTS"}); fbErr2 == nil {
			subjectID = fbID2
			subjectName = fbName2
			claimTypes = []string{"INHERITS", "EMBEDS", "IMPLEMENTS"}
		} else {
			return map[string]interface{}{
				"tenant_id":      tenantID.String(),
				"workspace_id":   workspaceID.String(),
				"workspace_name": workspaceName,
				"subject":        map[string]string{"id": "", "name": ""},
				"count":          0,
				"total_count":    0,
				"truncated":      false,
				"results":        []interface{}{},
				"limit_used":     limit,
			}, nil
		}
	}

	countQ := `
		SELECT COUNT(*)::int
		  FROM claims c
		  JOIN entities e ON e.id = c.from_entity_id
		 WHERE c.workspace_id = $1
		   AND c.claim_type = ANY($3)
		   AND c.to_entity_id = $2
		   AND e.kind != 'external'
	`
	countArgs := []interface{}{workspaceID, subjectID, claimTypes}
	if languageFilter != "" {
		countQ += " AND e.language = $4"
		countArgs = append(countArgs, languageFilter)
	}
	if s.store == nil || s.store.Pool() == nil {
		return map[string]interface{}{
			"tenant_id":      tenantID.String(),
			"workspace_id":   workspaceID.String(),
			"workspace_name": workspaceName,
			"status":         "not_found",
			"not_found":      true,
			"subject":        map[string]string{"id": subjectID.String(), "name": subjectName},
			"count":          0,
			"total_count":    0,
			"truncated":      false,
			"results":        []interface{}{},
			"limit_used":     limit,
			"hint":           "database not connected",
		}, nil
	}

	var totalCount int
	_ = s.store.Pool().QueryRow(ctx, countQ, countArgs...).Scan(&totalCount)

	q := `
		SELECT e.id::text, e.name, e.kind, COALESCE(e.package, ''), COALESCE(e.language, ''),
		       COALESCE(e.file_path, ''), COALESCE(e.line_start, 0),
		       COALESCE(NULLIF(r.name, ''), NULLIF(r.module_path, ''), r.url, 'unknown') AS repo,
		       c.claim_type, c.confidence,
		       e.is_exported
		  FROM claims c
		  JOIN entities e ON e.id = c.from_entity_id
		  LEFT JOIN repositories r ON r.id = e.repository_id
		 WHERE c.workspace_id = $1
		   AND c.claim_type = ANY($3)
		   AND c.to_entity_id = $2
		   AND e.kind != 'external'
	`
	dbArgs := []interface{}{workspaceID, subjectID, claimTypes}
	argPos := 4

	if languageFilter != "" {
		q += fmt.Sprintf(" AND e.language = $%d", argPos)
		dbArgs = append(dbArgs, languageFilter)
		argPos++
	}
	q += fmt.Sprintf(" ORDER BY e.name LIMIT $%d", argPos)
	dbArgs = append(dbArgs, limit)

	rows, err := s.store.Pool().Query(ctx, q, dbArgs...)
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
		r.File = sanitizeMCPFilePath(r.File)
		results = append(results, r)
	}

	if totalCount < len(results) {
		totalCount = len(results)
	}

	return map[string]interface{}{
		"tenant_id":      tenantID.String(),
		"workspace_id":   workspaceID.String(),
		"workspace_name": workspaceName,
		"subject":        map[string]string{"id": subjectID.String(), "name": subjectName},
		"filters":        map[string]any{"language": languageFilter},
		"count":          len(results),
		"total_count":    totalCount,
		"truncated":      totalCount > len(results),
		"results":        results,
		"limit_used":     limit,
	}, nil
}

// handleImplementers returns entities that implement the subject interface (BUG-18).
// Never silently substitutes an unrelated hub when a specific entity_id or name is requested,
// and always reports total_count and truncated.
func (s *MCPServer) handleImplementers(args map[string]interface{}) (interface{}, error) {
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	workspaceName := s.resolveMCPWorkspaceName(args, workspaceID)

	entityIDStr, _ := args["entity_id"].(string)
	nameStr, _ := args["name"].(string)
	entityIDStr = strings.TrimSpace(entityIDStr)
	nameStr = strings.TrimSpace(nameStr)
	explicitSubject := entityIDStr != "" || nameStr != ""
	ctx := context.Background()

	limit := 50
	if v, ok := args["limit"].(float64); ok && v > 0 && v <= 500 {
		limit = int(v)
	}

	if s.store == nil || s.store.Pool() == nil {
		return map[string]interface{}{
			"tenant_id":      tenantID.String(),
			"workspace_id":   workspaceID.String(),
			"workspace_name": workspaceName,
			"subject":        map[string]string{"id": "", "name": nameStr},
			"count":          0,
			"total_count":    0,
			"truncated":      false,
			"results":        []interface{}{},
			"limit_used":     limit,
		}, nil
	}
	claimFilter := []string{"IMPLEMENTS"}
	var subjectID uuid.UUID
	var subjectName string

	if explicitSubject {
		subjectID, subjectName, err = s.resolveSubjectEntity(ctx, workspaceID, entityIDStr, nameStr)
		if err != nil {
			reqTarget := nameStr
			if reqTarget == "" {
				reqTarget = entityIDStr
			}
			return map[string]interface{}{
				"tenant_id":      tenantID.String(),
				"workspace_id":   workspaceID.String(),
				"workspace_name": workspaceName,
				"status":         "not_found",
				"not_found":      true,
				"subject":        map[string]string{"id": "", "name": reqTarget},
				"count":          0,
				"total_count":    0,
				"truncated":      false,
				"results":        []interface{}{},
				"limit_used":     limit,
				"hint":           fmt.Sprintf("no interface or entity matching %q found in workspace %s", reqTarget, workspaceName),
			}, nil
		}
	} else {
		if fbID, fbName, fbErr := s.resolveTopClaimTarget(ctx, workspaceID, []string{"IMPLEMENTS"}); fbErr == nil {
			subjectID = fbID
			subjectName = fbName
		} else if fbID2, fbName2, fbErr2 := s.resolveTopClaimTarget(ctx, workspaceID, []string{"IMPLEMENTS", "EMBEDS", "INHERITS"}); fbErr2 == nil {
			subjectID = fbID2
			subjectName = fbName2
			claimFilter = []string{"IMPLEMENTS", "EMBEDS", "INHERITS"}
		} else {
			return map[string]interface{}{
				"tenant_id":      tenantID.String(),
				"workspace_id":   workspaceID.String(),
				"workspace_name": workspaceName,
				"subject":        map[string]string{"id": "", "name": ""},
				"count":          0,
				"total_count":    0,
				"truncated":      false,
				"results":        []interface{}{},
				"limit_used":     limit,
			}, nil
		}
	}

	var totalCount int
	_ = s.store.Pool().QueryRow(ctx, `
		SELECT COUNT(*)::int
		  FROM claims c
		  JOIN entities e ON e.id = c.from_entity_id
		 WHERE c.workspace_id = $1
		   AND c.claim_type = ANY($3)
		   AND c.to_entity_id = $2
		   AND e.kind != 'external'
	`, workspaceID, subjectID, claimFilter).Scan(&totalCount)

	rows, err := s.store.Pool().Query(ctx, `
		SELECT e.id::text, e.name, e.kind, COALESCE(e.package, ''), COALESCE(e.language, ''),
		       COALESCE(e.file_path, ''), COALESCE(e.line_start, 0),
		       COALESCE(NULLIF(r.name, ''), NULLIF(r.module_path, ''), r.url, 'unknown') AS repo,
		       c.confidence,
		       e.is_exported
		  FROM claims c
		  JOIN entities e ON e.id = c.from_entity_id
		  LEFT JOIN repositories r ON r.id = e.repository_id
		 WHERE c.workspace_id = $1
		   AND c.claim_type = ANY($4)
		   AND c.to_entity_id = $2
		   AND e.kind != 'external'
		 ORDER BY e.name
		 LIMIT $3
	`, workspaceID, subjectID, limit, claimFilter)
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
		r.File = sanitizeMCPFilePath(r.File)
		results = append(results, r)
	}

	if totalCount < len(results) {
		totalCount = len(results)
	}

	return map[string]interface{}{
		"tenant_id":      tenantID.String(),
		"workspace_id":   workspaceID.String(),
		"workspace_name": workspaceName,
		"subject":        map[string]string{"id": subjectID.String(), "name": subjectName},
		"count":          len(results),
		"total_count":    totalCount,
		"truncated":      totalCount > len(results),
		"results":        results,
		"limit_used":     limit,
	}, nil
}

// handleNeighbors returns every entity connected to the subject in
// either direction by a single edge, ensuring both incoming and outgoing
// edges are represented even when incoming count exceeds limit.
func (s *MCPServer) handleNeighbors(args map[string]interface{}) (interface{}, error) {
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	workspaceName := s.resolveMCPWorkspaceName(args, workspaceID)
	ctx := context.Background()

	entityIDStr, _ := args["entity_id"].(string)
	nameStr, _ := args["name"].(string)
	if entityIDStr == "" && nameStr == "" {
		return nil, fmt.Errorf("entity_id is required")
	}

	var subjectID uuid.UUID
	var subjectName string
	if parsed, pErr := uuid.Parse(entityIDStr); pErr == nil {
		subjectID = parsed
		err = s.store.Pool().QueryRow(ctx, `
			SELECT name FROM entities WHERE workspace_id = $1 AND id = $2
		`, workspaceID, subjectID).Scan(&subjectName)
		if err != nil {
			return nil, fmt.Errorf("entity_id %q not found in this workspace", entityIDStr)
		}
	} else {
		lookup := entityIDStr
		if lookup == "" {
			lookup = nameStr
		}
		subjectID, subjectName, err = s.resolveSubjectEntity(ctx, workspaceID, "", lookup)
		if err != nil {
			return nil, fmt.Errorf("invalid entity_id %q: %w", lookup, err)
		}
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

	if s.store == nil || s.store.Pool() == nil {
		return map[string]interface{}{
			"tenant_id":            tenantID.String(),
			"workspace_id":         workspaceID.String(),
			"workspace_name":       workspaceName,
			"subject":              map[string]string{"id": entityIDStr, "name": nameStr},
			"filters":              map[string]any{"edge_kinds": edgeKinds, "include_external": includeExternal},
			"total_incoming_count": 0,
			"total_outgoing_count": 0,
			"incoming_count":       0,
			"outgoing_count":       0,
			"incoming":             []interface{}{},
			"outgoing":             []interface{}{},
			"count":                0,
			"total_count":          0,
			"truncated":            false,
			"results":              []interface{}{},
			"limit_used":           limit,
		}, nil
	}

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

	queryDirection := func(direction string, allowExt bool, dirLimit int) ([]neighborRow, int, error) {
		colMatch := "c.to_entity_id = $2"
		otherCol := "c.from_entity_id"
		if direction == "outgoing" {
			colMatch = "c.from_entity_id = $2"
			otherCol = "c.to_entity_id"
		}

		extFilter := ""
		if !allowExt {
			extFilter = " AND e.kind != 'external'"
		}

		countQ := fmt.Sprintf(`
			SELECT COUNT(*)::int
			  FROM claims c
			  JOIN entities e ON e.id = %s
			 WHERE c.workspace_id = $1 AND %s%s
		`, otherCol, colMatch, extFilter)
		var totalDir int
		_ = s.store.Pool().QueryRow(ctx, countQ, workspaceID, subjectID).Scan(&totalDir)

		q := fmt.Sprintf(`
			SELECT e.id::text, e.name, e.kind, COALESCE(e.package, ''), COALESCE(e.language, 'go'),
			       COALESCE(e.file_path, ''), COALESCE(e.line_start, 0),
			       COALESCE(NULLIF(r.name, ''), NULLIF(r.module_path, ''), r.url, 'unknown') AS repo,
			       c.claim_type, c.confidence, $3::text AS direction,
			       COALESCE(e.is_exported, false)
			  FROM claims c
			  JOIN entities e ON e.id = %s
			  LEFT JOIN repositories r ON r.id = e.repository_id
			 WHERE c.workspace_id = $1 AND %s%s
		`, otherCol, colMatch, extFilter)

		dbArgs := []interface{}{workspaceID, subjectID, direction}
		argPos := 4
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
			q += fmt.Sprintf(" AND c.claim_type IN (%s)", placeholders)
		}
		nbOffset := ParseCursorOffset(args)
		q += fmt.Sprintf(" ORDER BY c.claim_type, e.name LIMIT $%d OFFSET $%d", argPos, argPos+1)
		dbArgs = append(dbArgs, dirLimit, nbOffset)

		rows, qErr := s.store.Pool().Query(ctx, q, dbArgs...)
		if qErr != nil {
			return nil, totalDir, qErr
		}
		defer rows.Close()

		out := make([]neighborRow, 0, dirLimit)
		for rows.Next() {
			var r neighborRow
			if err := rows.Scan(
				&r.ID, &r.Name, &r.Kind, &r.Package, &r.Language,
				&r.File, &r.LineStart, &r.Repo,
				&r.ClaimType, &r.Confidence, &r.Direction, &r.Exported,
			); err == nil {
				r.File = sanitizeMCPFilePath(r.File)
				out = append(out, r)
			}
		}
		return out, totalDir, nil
	}

	incomingRows, totalIncoming, err := queryDirection("incoming", includeExternal, limit)
	if err != nil {
		return nil, fmt.Errorf("query incoming neighbors: %w", err)
	}
	outgoingRows, totalOutgoing, err := queryDirection("outgoing", includeExternal, limit)
	if err != nil {
		return nil, fmt.Errorf("query outgoing neighbors: %w", err)
	}
	if len(outgoingRows) == 0 && !includeExternal {
		extOut, extTotal, extErr := queryDirection("outgoing", true, limit)
		if extErr == nil && len(extOut) > 0 {
			outgoingRows = extOut
			totalOutgoing = extTotal
		}
	}

	results := make([]neighborRow, 0, limit)
	halfLimit := limit / 2
	if halfLimit < 1 {
		halfLimit = 1
	}
	outTake := len(outgoingRows)
	if outTake > halfLimit && len(incomingRows) > (limit-halfLimit) {
		outTake = halfLimit
	}
	inTake := limit - outTake
	if inTake > len(incomingRows) {
		inTake = len(incomingRows)
		if outTake < len(outgoingRows) {
			rem := limit - inTake
			if rem > len(outgoingRows) {
				rem = len(outgoingRows)
			}
			outTake = rem
		}
	}
	results = append(results, incomingRows[:inTake]...)
	results = append(results, outgoingRows[:outTake]...)

	totalBoth := totalIncoming + totalOutgoing

	nbOffset := ParseCursorOffset(args)
	isTruncated := (nbOffset + len(results)) < totalBoth
	nextCursor := ""
	if isTruncated {
		nextCursor = fmt.Sprintf("%d", nbOffset+len(results))
	}
	return map[string]interface{}{
		"tenant_id":            tenantID.String(),
		"workspace_id":         workspaceID.String(),
		"workspace_name":       workspaceName,
		"subject":              map[string]string{"id": subjectID.String(), "name": subjectName},
		"filters":              map[string]any{"edge_kinds": edgeKinds, "include_external": includeExternal},
		"total_incoming_count": totalIncoming,
		"total_outgoing_count": totalOutgoing,
		"incoming_count":       len(incomingRows),
		"outgoing_count":       len(outgoingRows),
		"incoming":             incomingRows,
		"outgoing":             outgoingRows,
		"count":                len(results),
		"returned_count":       len(results),
		"total_count":          totalBoth,
		"offset":               nbOffset,
		"next_cursor":          nextCursor,
		"has_more":             isTruncated,
		"truncated":            isTruncated,
		"results":              results,
		"limit_used":           limit,
	}, nil
}

// handleFindEntity returns entities matching the given filters (BUG-05, BUG-18, CLASS-04, CLASS-06).
// Accepts both `name` and `name_pattern`. When any filter is provided and 0 rows match,
// returns an explicit not_found response with empty results instead of top hubs.
func (s *MCPServer) handleFindEntity(args map[string]interface{}) (interface{}, error) {
	ctx := context.Background()
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	workspaceName := s.resolveMCPWorkspaceName(args, workspaceID)

	exactName, _ := args["name"].(string)
	exactName = strings.TrimSpace(exactName)
	namePattern, _ := args["name_pattern"].(string)
	namePattern = strings.TrimSpace(namePattern)
	kindFilter, _ := args["kind"].(string)
	kindFilter = strings.TrimSpace(kindFilter)
	packageFilter, _ := args["package"].(string)
	packageFilter = strings.TrimSpace(packageFilter)
	fileFilter, _ := args["file_path_contains"].(string)
	fileFilter = strings.TrimSpace(fileFilter)
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

	whereClause := " WHERE e.workspace_id = $1 AND e.kind != 'external'"
	dbArgs := []interface{}{workspaceID}
	argPos := 2

	if exactName != "" {
		whereClause += fmt.Sprintf(" AND (e.name = $%d OR e.name ILIKE $%d OR e.name ILIKE '%%.' || $%d)", argPos, argPos, argPos)
		dbArgs = append(dbArgs, exactName)
		argPos++
	}
	if namePattern != "" {
		whereClause += fmt.Sprintf(" AND e.name ~ $%d", argPos)
		dbArgs = append(dbArgs, namePattern)
		argPos++
	}
	if kindFilter != "" {
		whereClause += fmt.Sprintf(" AND e.kind = $%d", argPos)
		dbArgs = append(dbArgs, kindFilter)
		argPos++
	}
	if packageFilter != "" {
		whereClause += fmt.Sprintf(" AND e.package LIKE '%%' || $%d || '%%'", argPos)
		dbArgs = append(dbArgs, packageFilter)
		argPos++
	}
	if fileFilter != "" {
		whereClause += fmt.Sprintf(" AND e.file_path LIKE '%%' || $%d || '%%'", argPos)
		dbArgs = append(dbArgs, fileFilter)
		argPos++
	}
	if exportedOnly {
		whereClause += " AND e.is_exported = true"
	}

	var totalCount int
	_ = s.store.Pool().QueryRow(ctx, "SELECT COUNT(*)::int FROM entities e"+whereClause, dbArgs...).Scan(&totalCount)

	q := `
		SELECT e.id::text, e.name, e.kind, COALESCE(e.package, ''), COALESCE(e.language, ''),
		       COALESCE(e.file_path, ''), COALESCE(e.line_start, 0),
		       COALESCE(NULLIF(r.name, ''), NULLIF(r.module_path, ''), r.url, 'unknown') AS repo,
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
	` + whereClause + fmt.Sprintf(" ORDER BY inbound_edges DESC, e.name LIMIT $%d", argPos)
	queryArgs := append(append([]interface{}{}, dbArgs...), limit)

	rows, err := s.store.Pool().Query(ctx, q, queryArgs...)
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
		r.File = sanitizeMCPFilePath(r.File)
		results = append(results, r)
	}

	if totalCount < len(results) {
		totalCount = len(results)
	}

	hasFilter := exactName != "" || namePattern != "" || kindFilter != "" || packageFilter != "" || fileFilter != ""
	filtersMap := map[string]any{
		"name":               exactName,
		"name_pattern":       namePattern,
		"kind":               kindFilter,
		"package":            packageFilter,
		"file_path_contains": fileFilter,
		"exported_only":      exportedOnly,
	}

	if len(results) == 0 && hasFilter {
		filterDesc := exactName
		if filterDesc == "" {
			filterDesc = namePattern
		}
		if filterDesc == "" {
			filterDesc = packageFilter
		}
		return map[string]interface{}{
			"tenant_id":      tenantID.String(),
			"workspace_id":   workspaceID.String(),
			"workspace_name": workspaceName,
			"status":         "not_found",
			"not_found":      true,
			"filters":        filtersMap,
			"count":          0,
			"total_count":    0,
			"truncated":      false,
			"entities":       []entityRow{},
			"results":        []entityRow{},
			"limit_used":     limit,
			"hint":           fmt.Sprintf("no symbol matching %q found in workspace %s", filterDesc, workspaceName),
		}, nil
	}

	return map[string]interface{}{
		"tenant_id":      tenantID.String(),
		"workspace_id":   workspaceID.String(),
		"workspace_name": workspaceName,
		"status":         "ok",
		"not_found":      false,
		"filters":        filtersMap,
		"count":          len(results),
		"total_count":    totalCount,
		"truncated":      totalCount > len(results),
		"entities":       results,
		"results":        results,
		"limit_used":     limit,
	}, nil
}
