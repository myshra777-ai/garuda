// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type WorkspaceCounts struct {
	Repositories int
	Entities     int
	Claims       int
}

type ArchitecturalHub struct {
	Name    string
	Kind    string
	Package string
	Callers int
}

type CrossRepoBridge struct {
	FromModule string
	ToModule   string
	CallCount  int
}

type RepoSummaryMetrics struct {
	EntityCount  int
	PackageCount int
}

type ExportedSymbol struct {
	Name         string
	Kind         string
	ReceiverType string
}

type DependencyRef struct {
	Package    string
	References int
}

type SymbolSummaryDetail struct {
	ID            uuid.UUID
	Name          string
	Kind          string
	Package       string
	ReceiverType  string
	FilePath      string
	Signature     string
	IsExported    bool
	Line          int
	InboundCount  int
	OutboundCount int
}

// ResolveWorkspaceTarget resolves workspace by name or falls back to the first available workspace.
func (s *PostgresStore) ResolveWorkspaceTarget(ctx context.Context, name string) (uuid.UUID, string, error) {
	var wsID uuid.UUID
	err := s.pool.QueryRow(ctx, "SELECT id FROM workspaces WHERE name = $1 LIMIT 1", name).Scan(&wsID)
	if err == nil {
		return wsID, name, nil
	}

	var fallbackName string
	err = s.pool.QueryRow(ctx, "SELECT id, name FROM workspaces LIMIT 1").Scan(&wsID, &fallbackName)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("no workspaces found: %w", err)
	}
	return wsID, fallbackName, nil
}

// FindRepoByTarget searches for a repository by module path or URL match.
func (s *PostgresStore) FindRepoByTarget(ctx context.Context, workspaceID uuid.UUID, target string) (uuid.UUID, string, string, error) {
	var repoID uuid.UUID
	var repoURL, modPath string
	err := s.pool.QueryRow(ctx, `
		SELECT id, url, module_path FROM repositories 
		WHERE workspace_id = $1 AND (module_path ILIKE '%' || $2 || '%' OR url ILIKE '%' || $2 || '%')
		LIMIT 1
	`, workspaceID, target).Scan(&repoID, &repoURL, &modPath)
	if err != nil {
		return uuid.Nil, "", "", err
	}
	return repoID, repoURL, modPath, nil
}

// GetWorkspaceCounts retrieves counts of repositories, internal entities, and claims.
func (s *PostgresStore) GetWorkspaceCounts(ctx context.Context, workspaceID uuid.UUID) (WorkspaceCounts, error) {
	var c WorkspaceCounts
	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM repositories WHERE workspace_id = $1`, workspaceID).Scan(&c.Repositories)
	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM entities WHERE workspace_id = $1 AND kind != 'external'`, workspaceID).Scan(&c.Entities)
	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM claims WHERE workspace_id = $1`, workspaceID).Scan(&c.Claims)
	return c, nil
}

// GetTopArchitecturalHubs returns entities with the most inbound callers.
func (s *PostgresStore) GetTopArchitecturalHubs(ctx context.Context, workspaceID uuid.UUID, limit int) ([]ArchitecturalHub, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT e.name, e.kind, e.package, count(c.id) as callers
		FROM entities e
		JOIN claims c ON c.to_entity_id = e.id
		WHERE e.workspace_id = $1 AND e.kind != 'external'
		GROUP BY e.id, e.name, e.kind, e.package
		ORDER BY callers DESC
		LIMIT $2;
	`, workspaceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hubs []ArchitecturalHub
	for rows.Next() {
		var h ArchitecturalHub
		if err := rows.Scan(&h.Name, &h.Kind, &h.Package, &h.Callers); err == nil {
			hubs = append(hubs, h)
		}
	}
	return hubs, nil
}

// GetCrossRepoBridges returns cross-repository calls in a workspace.
func (s *PostgresStore) GetCrossRepoBridges(ctx context.Context, workspaceID uuid.UUID, limit int) ([]CrossRepoBridge, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT r1.module_path, r2.module_path, count(*) 
		FROM claims c
		JOIN entities e1 ON c.from_entity_id = e1.id
		JOIN entities e2 ON c.to_entity_id = e2.id
		JOIN repositories r1 ON e1.repository_id = r1.id
		JOIN repositories r2 ON e2.repository_id = r2.id
		WHERE c.workspace_id = $1 AND r1.id != r2.id
		GROUP BY r1.module_path, r2.module_path
		LIMIT $2;
	`, workspaceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bridges []CrossRepoBridge
	for rows.Next() {
		var b CrossRepoBridge
		if err := rows.Scan(&b.FromModule, &b.ToModule, &b.CallCount); err == nil {
			bridges = append(bridges, b)
		}
	}
	return bridges, nil
}

// GetLatestMerkleRoot retrieves the most recent Merkle root across decision revisions.
func (s *PostgresStore) GetLatestMerkleRoot(ctx context.Context) string {
	var root string
	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(merkle_root, '<pending>')
		FROM decision_revisions
		GROUP BY merkle_root, created_at
		ORDER BY created_at DESC LIMIT 1
	`).Scan(&root)
	if root == "" {
		return "<pending>"
	}
	return root
}

// GetRepoMetrics returns entity and distinct package counts for a repo.
func (s *PostgresStore) GetRepoMetrics(ctx context.Context, repoID uuid.UUID) (RepoSummaryMetrics, error) {
	var m RepoSummaryMetrics
	err := s.pool.QueryRow(ctx, `SELECT count(*), count(DISTINCT package) FROM entities WHERE repository_id = $1 AND kind != 'external'`, repoID).Scan(&m.EntityCount, &m.PackageCount)
	return m, err
}

// GetExportedSymbols returns top exported symbols for a repository.
func (s *PostgresStore) GetExportedSymbols(ctx context.Context, repoID uuid.UUID, limit int) ([]ExportedSymbol, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT name, kind, COALESCE(receiver_type, '')
		FROM entities 
		WHERE repository_id = $1 AND is_exported = true AND kind IN ('function', 'struct', 'interface')
		ORDER BY kind DESC, name ASC
		LIMIT $2;
	`, repoID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exports []ExportedSymbol
	for rows.Next() {
		var sym ExportedSymbol
		if err := rows.Scan(&sym.Name, &sym.Kind, &sym.ReceiverType); err == nil {
			exports = append(exports, sym)
		}
	}
	return exports, nil
}

// GetExternalDependencies returns external or cross-repo package references.
func (s *PostgresStore) GetExternalDependencies(ctx context.Context, repoID uuid.UUID, limit int) ([]DependencyRef, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT e2.package, count(*) as count
		FROM claims c
		JOIN entities e1 ON c.from_entity_id = e1.id
		JOIN entities e2 ON c.to_entity_id = e2.id
		WHERE e1.repository_id = $1 AND (e2.repository_id != $1 OR e2.kind = 'external')
		GROUP BY e2.package
		ORDER BY count DESC
		LIMIT $2;
	`, repoID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []DependencyRef
	for rows.Next() {
		var d DependencyRef
		if err := rows.Scan(&d.Package, &d.References); err == nil {
			deps = append(deps, d)
		}
	}
	return deps, nil
}

// GetSymbolSummaryDetail resolves an entity by ID or name and counts inbound/outbound claims.
func (s *PostgresStore) GetSymbolSummaryDetail(ctx context.Context, workspaceID uuid.UUID, target string) (*SymbolSummaryDetail, error) {
	var det SymbolSummaryDetail
	var err error

	if parsedID, parseErr := uuid.Parse(target); parseErr == nil {
		err = s.pool.QueryRow(ctx, `
			SELECT id, name, kind, package, receiver_type, file_path, signature, is_exported, line
			FROM entities WHERE workspace_id = $1 AND id = $2 LIMIT 1
		`, workspaceID, parsedID).Scan(&det.ID, &det.Name, &det.Kind, &det.Package, &det.ReceiverType, &det.FilePath, &det.Signature, &det.IsExported, &det.Line)
	} else {
		sym := target
		if dot := strings.LastIndex(target, "."); dot != -1 {
			sym = target[dot+1:]
		}
		err = s.pool.QueryRow(ctx, `
			SELECT id, name, kind, package, receiver_type, file_path, signature, is_exported, line
			FROM entities 
			WHERE workspace_id = $1 AND (name = $2 OR name ILIKE $2)
			ORDER BY (kind != 'external') DESC, is_exported DESC
			LIMIT 1
		`, workspaceID, sym).Scan(&det.ID, &det.Name, &det.Kind, &det.Package, &det.ReceiverType, &det.FilePath, &det.Signature, &det.IsExported, &det.Line)
	}

	if err != nil {
		return nil, fmt.Errorf("could not find symbol matching '%s': %w", target, err)
	}

	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM claims WHERE workspace_id = $1 AND to_entity_id = $2`, workspaceID, det.ID).Scan(&det.InboundCount)
	_ = s.pool.QueryRow(ctx, `SELECT count(*) FROM claims WHERE workspace_id = $1 AND from_entity_id = $2`, workspaceID, det.ID).Scan(&det.OutboundCount)

	return &det, nil
}
