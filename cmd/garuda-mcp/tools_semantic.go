// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/analyzer"
)

func sanitizeMCPFilePath(p string) string {
	clean := filepath.ToSlash(strings.TrimSpace(p))
	if clean == "" {
		return ""
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		slashHome := filepath.ToSlash(home)
		for _, sub := range []string{"/garuda-validation/", "/garuda/"} {
			if strings.HasPrefix(clean, slashHome+sub) {
				return strings.TrimPrefix(clean, slashHome+sub)
			}
		}
		if strings.HasPrefix(clean, slashHome+"/") {
			return strings.TrimPrefix(clean, slashHome+"/")
		}
	}
	for _, prefix := range []string{"/home/rohit/garuda-validation/", "/home/rohit/garuda/", "/tmp/garuda/"} {
		if strings.HasPrefix(clean, prefix) {
			return strings.TrimPrefix(clean, prefix)
		}
	}
	return clean
}

// handleEntities lists entities in a workspace, optionally filtered by
// package or kind.
func (s *MCPServer) handleEntities(args map[string]interface{}) (interface{}, error) {
	ctx := context.Background()
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	workspaceName := s.resolveMCPWorkspaceName(args, workspaceID)

	pkgFilter, _ := args["package"].(string)
	kindFilter, _ := args["kind"].(string)
	limit := 100
	if v, ok := args["limit"].(float64); ok && v > 0 && v <= 1000 {
		limit = int(v)
	}

	whereClause := " WHERE workspace_id = $1 AND kind != 'external'"
	dbArgs := []interface{}{workspaceID}
	argPos := 2

	if pkgFilter != "" {
		whereClause += fmt.Sprintf(" AND package LIKE '%%' || $%d || '%%'", argPos)
		dbArgs = append(dbArgs, pkgFilter)
		argPos++
	}
	if kindFilter != "" {
		whereClause += fmt.Sprintf(" AND kind = $%d", argPos)
		dbArgs = append(dbArgs, kindFilter)
		argPos++
	}

	if s.store == nil || s.store.Pool() == nil {
		return map[string]interface{}{
			"tenant_id":      tenantID.String(),
			"workspace_id":   workspaceID.String(),
			"workspace_name": workspaceName,
			"filters":        map[string]any{"package": pkgFilter, "kind": kindFilter},
			"count":          0,
			"total_count":    0,
			"truncated":      false,
			"entities":       []interface{}{},
			"limit_used":     limit,
		}, nil
	}

	var totalCount int
	_ = s.store.Pool().QueryRow(ctx, "SELECT COUNT(*)::int FROM entities"+whereClause, dbArgs...).Scan(&totalCount)

	q := `
		SELECT id::text, name, kind, COALESCE(package, ''), COALESCE(file_path, ''),
		       COALESCE(line_start, 0), COALESCE(signature, ''), is_exported
		FROM entities
	` + whereClause + fmt.Sprintf(" ORDER BY package, name LIMIT $%d", argPos)
	queryArgs := append(append([]interface{}{}, dbArgs...), limit)

	rows, err := s.store.Pool().Query(ctx, q, queryArgs...)
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
		Signature string `json:"signature,omitempty"`
		Exported  bool   `json:"exported"`
	}

	entities := make([]entityRow, 0, limit)
	for rows.Next() {
		var e entityRow
		if err := rows.Scan(&e.ID, &e.Name, &e.Kind, &e.Package, &e.File, &e.LineStart, &e.Signature, &e.Exported); err != nil {
			continue
		}
		e.File = sanitizeMCPFilePath(e.File)
		entities = append(entities, e)
	}

	if totalCount < len(entities) {
		totalCount = len(entities)
	}

	return map[string]interface{}{
		"tenant_id":      tenantID.String(),
		"workspace_id":   workspaceID.String(),
		"workspace_name": workspaceName,
		"filters":        map[string]any{"package": pkgFilter, "kind": kindFilter},
		"count":          len(entities),
		"total_count":    totalCount,
		"truncated":      totalCount > len(entities),
		"entities":       entities,
		"limit_used":     limit,
	}, nil
}

// handleInspect returns one entity and its incoming and outgoing
// relationships, including its canonical signature.
func (s *MCPServer) handleInspect(args map[string]interface{}) (interface{}, error) {
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	workspaceName := s.resolveMCPWorkspaceName(args, workspaceID)

	entityIDStr, _ := args["entity_id"].(string)
	nameStr, _ := args["name"].(string)
	if entityIDStr == "" && nameStr == "" {
		return nil, fmt.Errorf("entity_id is required")
	}

	ctx := context.Background()
	lookupTarget := entityIDStr
	if lookupTarget == "" {
		lookupTarget = nameStr
	}
	if _, parseErr := uuid.Parse(lookupTarget); parseErr != nil {
		resolvedID, _, rErr := s.resolveSubjectEntity(ctx, workspaceID, "", lookupTarget)
		if rErr != nil {
			return nil, fmt.Errorf("invalid entity_id %q: %w", lookupTarget, parseErr)
		}
		entityIDStr = resolvedID.String()
	} else {
		entityIDStr = lookupTarget
	}

	var entity struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Kind         string `json:"kind"`
		Package      string `json:"package"`
		ReceiverType string `json:"receiver_type,omitempty"`
		Language     string `json:"language"`
		File         string `json:"file"`
		LineStart    int    `json:"line_start"`
		LineEnd      int    `json:"line_end,omitempty"`
		Signature    string `json:"signature"`
		Exported     bool   `json:"exported"`
	}
	if s.store == nil || s.store.Pool() == nil {
		return nil, fmt.Errorf("database not connected")
	}

	err = s.store.Pool().QueryRow(ctx, `
		SELECT id::text, name, kind, COALESCE(package, ''), COALESCE(receiver_type, ''),
		       COALESCE(language, 'go'), COALESCE(file_path, ''),
		       COALESCE(line_start, 0), COALESCE(line_end, 0),
		       COALESCE(signature, ''), COALESCE(is_exported, false)
		FROM entities
		WHERE workspace_id = $1 AND id = $2
	`, workspaceID, entityIDStr).Scan(
		&entity.ID, &entity.Name, &entity.Kind, &entity.Package, &entity.ReceiverType,
		&entity.Language, &entity.File, &entity.LineStart, &entity.LineEnd,
		&entity.Signature, &entity.Exported,
	)
	if err != nil {
		return nil, fmt.Errorf("entity %q not found in workspace %q", entityIDStr, workspaceName)
	}

	entity.File = sanitizeMCPFilePath(entity.File)
	if strings.TrimSpace(entity.Signature) == "" {
		switch entity.Kind {
		case "method":
			if entity.ReceiverType != "" {
				entity.Signature = fmt.Sprintf("func (%s) %s()", entity.ReceiverType, entity.Name)
			} else {
				entity.Signature = fmt.Sprintf("func %s.%s()", entity.Package, entity.Name)
			}
		case "function":
			if entity.Package != "" {
				entity.Signature = fmt.Sprintf("func %s.%s()", entity.Package, entity.Name)
			} else {
				entity.Signature = fmt.Sprintf("func %s()", entity.Name)
			}
		case "struct":
			entity.Signature = fmt.Sprintf("type %s struct", entity.Name)
		case "interface":
			entity.Signature = fmt.Sprintf("type %s interface", entity.Name)
		case "class":
			entity.Signature = fmt.Sprintf("class %s", entity.Name)
		default:
			entity.Signature = fmt.Sprintf("%s %s.%s", entity.Kind, entity.Package, entity.Name)
		}
	}

	incoming, outgoing, err := s.store.GetEntityRelationships(
		ctx, tenantID, workspaceID, entityIDStr,
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

	convert := func(rs []analyzer.Relationship) []relRow {
		out := make([]relRow, 0, len(rs))
		for _, r := range rs {
			out = append(out, relRow{
				From: r.From,
				To:   r.To,
				Type: r.Type,
				File: sanitizeMCPFilePath(r.Evidence.File),
				Line: r.Evidence.LineStart,
			})
		}
		return out
	}

	return map[string]interface{}{
		"tenant_id":      tenantID.String(),
		"workspace_id":   workspaceID.String(),
		"workspace_name": workspaceName,
		"entity":         entity,
		"signature":      entity.Signature,
		"incoming_count": len(incoming),
		"outgoing_count": len(outgoing),
		"incoming":       convert(incoming),
		"outgoing":       convert(outgoing),
	}, nil
}

// handleBriefing returns a session-start briefing (BUG-03 & BUG-08).
// Uses LoadWorkspaceCounts as the single source of truth for scale and epistemic counts.
func (s *MCPServer) handleBriefing(args map[string]interface{}) (interface{}, error) {
	ctx := context.Background()
	tenantID, workspaceID, err := s.resolveTenantAndWorkspace(args)
	if err != nil {
		return nil, err
	}
	workspaceName := s.resolveMCPWorkspaceName(args, workspaceID)

	agentID, _ := args["agent_id"].(string)
	if agentID == "" {
		agentID = s.clientName
	}
	if agentID == "" {
		agentID = "default"
	}

	sessionID := uuid.New().String()

	if s.store == nil || s.store.Pool() == nil {
		counts := LoadWorkspaceCounts(ctx, nil, tenantID, workspaceID)
		return map[string]interface{}{
			"tenant_id":              tenantID.String(),
			"workspace":              workspaceName,
			"workspace_id":           workspaceID.String(),
			"entity_count":           counts.Entities,
			"exported_entity_count":  counts.ExportedEntities,
			"external_stubs_count":   counts.ExternalStubs,
			"ast_relationship_count": counts.ASTRelationships,
			"document_claim_count":   counts.DocumentClaims,
			"policy_verdict":         "NO_DATA",
			"epistemic": map[string]int{
				"supported":           0,
				"unverified":          0,
				"contradicted":        0,
				"open_contradictions": 0,
			},
			"epistemic_counts": map[string]int{
				"SUPPORTED":           0,
				"UNVERIFIED":          0,
				"CONTRADICTED":        0,
				"OPEN_CONTRADICTIONS": 0,
			},
			"briefing": map[string]interface{}{},
		}, nil
	}

	briefing, err := s.store.GetBriefing(
		ctx,
		tenantID, workspaceID,
		workspaceName, agentID, sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("build briefing: %w", err)
	}

	counts := LoadWorkspaceCounts(ctx, s.store.Pool(), tenantID, workspaceID)
	briefing.Scale.Repositories = counts.Repositories
	briefing.Scale.Entities = counts.Entities
	briefing.Scale.ExportedEntities = counts.ExportedEntities
	briefing.Scale.ExternalStubs = counts.ExternalStubs
	briefing.Scale.Claims = counts.ASTRelationships
	briefing.Scale.ASTRelationships = counts.ASTRelationships
	briefing.Scale.DocumentClaims = counts.DocumentClaims
	briefing.Attention.OpenContradictions = counts.OpenContradictions

	epistemic := map[string]int{
		"supported":             counts.SupportedClaims,
		"unverified":            counts.UnverifiedClaims,
		"contradicted":          counts.ContradictedClaims,
		"open_contradictions":   counts.OpenContradictions,
		"total_document_claims": counts.DocumentClaims,
		"total_ast_claims":      counts.ASTRelationships,
	}
	briefing.EpistemicCounts = map[string]int{
		"SUPPORTED":             counts.SupportedClaims,
		"UNVERIFIED":            counts.UnverifiedClaims,
		"CONTRADICTED":          counts.ContradictedClaims,
		"OPEN_CONTRADICTIONS":   counts.OpenContradictions,
		"TOTAL_DOCUMENT_CLAIMS": counts.DocumentClaims,
		"TOTAL_AST_CLAIMS":      counts.ASTRelationships,
	}

	return map[string]interface{}{
		"tenant_id":              tenantID.String(),
		"workspace":              workspaceName,
		"workspace_id":           workspaceID.String(),
		"entity_count":           counts.Entities,
		"exported_entity_count":  counts.ExportedEntities,
		"external_stubs_count":   counts.ExternalStubs,
		"ast_relationship_count": counts.ASTRelationships,
		"document_claim_count":   counts.DocumentClaims,
		"policy_verdict":         briefing.PolicyVerdict,
		"epistemic":              epistemic,
		"epistemic_counts":       briefing.EpistemicCounts,
		"briefing":               briefing,
	}, nil
}
