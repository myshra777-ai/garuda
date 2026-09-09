// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/analyzer"
)

type EntityInspection struct {
	ID             string
	Name           string
	Kind           string
	Package        string
	FilePath       string
	Signature      string
	IsExported     bool
	Fields         []analyzer.Field
	Methods        []analyzer.Method
	OutgoingClaims []string
	IncomingClaims []string
}

type EntityListing struct {
	Name       string
	Kind       string
	Package    string
	FilePath   string
	IsExported bool
}

type GraphEntityNode struct {
	ID         string
	Name       string
	Kind       string
	Package    string
	FilePath   string
	IsExported bool
}

type GraphClaimEdge struct {
	From string
	To   string
	Type string
}

// InspectEntityDetails retrieves full entity metadata and its incoming/outgoing claims.
func (s *PostgresStore) InspectEntityDetails(ctx context.Context, tenantID uuid.UUID, entityName string) (*EntityInspection, error) {
	var insp EntityInspection
	var fieldsJSON, methodsJSON []byte

	err := s.pool.QueryRow(ctx, `
		SELECT id, name, kind, package, file_path, fields, methods, signature, is_exported
		FROM entities
		WHERE tenant_id = $1 AND name = $2
		LIMIT 1
	`, tenantID, entityName).Scan(
		&insp.ID, &insp.Name, &insp.Kind, &insp.Package,
		&insp.FilePath, &fieldsJSON, &methodsJSON, &insp.Signature, &insp.IsExported,
	)
	if err != nil {
		return nil, fmt.Errorf("entity '%s' not found: %w", entityName, err)
	}

	if len(fieldsJSON) > 0 {
		_ = json.Unmarshal(fieldsJSON, &insp.Fields)
	}
	if len(methodsJSON) > 0 {
		_ = json.Unmarshal(methodsJSON, &insp.Methods)
	}

	rowsOut, err := s.pool.Query(ctx, `
		SELECT claim_type, to_entity_id FROM claims
		WHERE tenant_id = $1 AND from_entity_id = $2
	`, tenantID, insp.ID)
	if err == nil {
		defer rowsOut.Close()
		for rowsOut.Next() {
			var typ, toID string
			if err := rowsOut.Scan(&typ, &toID); err == nil {
				insp.OutgoingClaims = append(insp.OutgoingClaims, fmt.Sprintf("      • %s -> %s", typ, toID))
			}
		}
	}

	rowsIn, err := s.pool.Query(ctx, `
		SELECT claim_type, from_entity_id FROM claims
		WHERE tenant_id = $1 AND to_entity_id = $2
	`, tenantID, insp.ID)
	if err == nil {
		defer rowsIn.Close()
		for rowsIn.Next() {
			var typ, fromID string
			if err := rowsIn.Scan(&typ, &fromID); err == nil {
				insp.IncomingClaims = append(insp.IncomingClaims, fmt.Sprintf("      • %s <- %s", typ, fromID))
			}
		}
	}

	return &insp, nil
}

// ListWorkspaceEntities lists all entities in a given workspace.
func (s *PostgresStore) ListWorkspaceEntities(ctx context.Context, tenantID, workspaceID uuid.UUID) ([]EntityListing, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT name, kind, package, file_path, is_exported
		FROM entities
		WHERE tenant_id = $1 AND workspace_id = $2
		ORDER BY package, name
	`, tenantID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query workspace entities: %w", err)
	}
	defer rows.Close()

	var entities []EntityListing
	for rows.Next() {
		var e EntityListing
		if err := rows.Scan(&e.Name, &e.Kind, &e.Package, &e.FilePath, &e.IsExported); err != nil {
			return nil, err
		}
		entities = append(entities, e)
	}
	return entities, rows.Err()
}

// FindRepositoryByFilter resolves a repository in a workspace by UUID, URL pattern, or module path.
func (s *PostgresStore) FindRepositoryByFilter(ctx context.Context, workspaceID uuid.UUID, filter string) (*Repository, error) {
	if filter == "" {
		return nil, nil
	}
	var r Repository
	if repoUUID, err := uuid.Parse(filter); err == nil {
		err := s.pool.QueryRow(ctx, `
			SELECT id, url, default_branch, language, module_path, enabled, analysis_status
			FROM repositories
			WHERE workspace_id = $1 AND id = $2
		`, workspaceID, repoUUID).Scan(&r.ID, &r.URL, &r.DefaultBranch, &r.Language, &r.ModulePath, &r.Enabled, &r.AnalysisStatus)
		if err != nil {
			return nil, err
		}
		return &r, nil
	}

	err := s.pool.QueryRow(ctx, `
		SELECT id, url, default_branch, language, module_path, enabled, analysis_status
		FROM repositories
		WHERE workspace_id = $1 AND (url LIKE $2 OR module_path = $3)
		LIMIT 1
	`, workspaceID, "%"+filter+"%", filter).Scan(&r.ID, &r.URL, &r.DefaultBranch, &r.Language, &r.ModulePath, &r.Enabled, &r.AnalysisStatus)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// GetWorkspaceGraphElements retrieves nodes and edges for graph generation.
func (s *PostgresStore) GetWorkspaceGraphElements(ctx context.Context, tenantID, workspaceID uuid.UUID, repoID *uuid.UUID) ([]GraphEntityNode, []GraphClaimEdge, error) {
	query := `
		SELECT
			id,
			COALESCE(name, 'unknown') as name,
			COALESCE(kind, 'unknown') as kind,
			COALESCE(package, '') as package,
			COALESCE(file_path, '') as file_path,
			COALESCE(is_exported, false) as is_exported
		FROM entities
		WHERE tenant_id = $1 AND workspace_id = $2
	`
	args := []interface{}{tenantID, workspaceID}
	if repoID != nil {
		query += ` AND repository_id = $3`
		args = append(args, *repoID)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query entities for graph: %w", err)
	}
	defer rows.Close()

	var nodes []GraphEntityNode
	for rows.Next() {
		var n GraphEntityNode
		if err := rows.Scan(&n.ID, &n.Name, &n.Kind, &n.Package, &n.FilePath, &n.IsExported); err != nil {
			return nil, nil, err
		}
		nodes = append(nodes, n)
	}

	claimRows, err := s.pool.Query(ctx, `
		SELECT from_entity_id, to_entity_id, claim_type
		FROM claims
		WHERE tenant_id = $1 AND workspace_id = $2
	`, tenantID, workspaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query claims for graph: %w", err)
	}
	defer claimRows.Close()

	var edges []GraphClaimEdge
	for claimRows.Next() {
		var fromID, toID uuid.UUID
		var claimType string
		if err := claimRows.Scan(&fromID, &toID, &claimType); err != nil {
			continue
		}
		if fromID == uuid.Nil || toID == uuid.Nil {
			continue
		}
		edges = append(edges, GraphClaimEdge{
			From: fromID.String(),
			To:   toID.String(),
			Type: claimType,
		})
	}

	return nodes, edges, nil
}
