// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CachedSymbol represents a persisted AST symbol in PostgreSQL.
type CachedSymbol struct {
	ID            uuid.UUID       `json:"id"`
	TenantID      uuid.UUID       `json:"tenant_id"`
	PackagePath   string          `json:"package_path"`
	SymbolName    string          `json:"symbol_name"`
	Kind          string          `json:"kind"`
	Receiver      string          `json:"receiver"`
	SignatureHash []byte          `json:"signature_hash"`
	ASTHash       []byte          `json:"ast_hash"`
	Payload       json.RawMessage `json:"payload"`
}

// CrossModuleEdge represents a resolved dependency edge across module boundaries.
type CrossModuleEdge struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	FromModule  string    `json:"from_module"`
	FromPackage string    `json:"from_package"`
	FromSymbol  string    `json:"from_symbol"`
	ToModule    string    `json:"to_module"`
	ToPackage   string    `json:"to_package"`
	ToSymbol    string    `json:"to_symbol"`
	EdgeType    string    `json:"edge_type"`
	Confidence  float64   `json:"confidence"`
}

// RegisterWorkspace upserts a workspace entry for the tenant.
func (s *PostgresStore) RegisterWorkspace(ctx context.Context, tenantID uuid.UUID, name, rootPath string, isGoWork bool) (uuid.UUID, error) {
	var workspaceID uuid.UUID
	err := s.pool.QueryRow(ctx, `
		INSERT INTO workspaces (tenant_id, name, root_path, is_go_work, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (tenant_id, name) DO UPDATE SET
			root_path = EXCLUDED.root_path,
			is_go_work = EXCLUDED.is_go_work,
			updated_at = NOW()
		RETURNING id
	`, tenantID, name, rootPath, isGoWork).Scan(&workspaceID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to register workspace: %w", err)
	}
	return workspaceID, nil
}

// GetModuleTreeHash retrieves the cached SHA-256 tree hash for a module.
func (s *PostgresStore) GetModuleTreeHash(ctx context.Context, tenantID, workspaceID uuid.UUID, modulePath string) ([]byte, error) {
	var treeHash []byte
	err := s.pool.QueryRow(ctx, `
		SELECT tree_hash FROM workspace_modules
		WHERE tenant_id = $1 AND workspace_id = $2 AND module_path = $3
	`, tenantID, workspaceID, modulePath).Scan(&treeHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get module tree hash: %w", err)
	}
	return treeHash, nil
}

// UpsertWorkspaceModule updates or records a module's tree hash.
func (s *PostgresStore) UpsertWorkspaceModule(ctx context.Context, tenantID, workspaceID uuid.UUID, modulePath, relPath, commitSHA string, treeHash []byte) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO workspace_modules (tenant_id, workspace_id, module_path, relative_path, commit_sha, tree_hash, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (workspace_id, module_path) DO UPDATE SET
			relative_path = EXCLUDED.relative_path,
			commit_sha = EXCLUDED.commit_sha,
			tree_hash = EXCLUDED.tree_hash,
			updated_at = NOW()
	`, tenantID, workspaceID, modulePath, relPath, commitSHA, treeHash)
	if err != nil {
		return fmt.Errorf("failed to upsert workspace module: %w", err)
	}
	return nil
}

// BatchUpsertSymbols stores analyzed symbols into symbol_cache.
func (s *PostgresStore) BatchUpsertSymbols(ctx context.Context, tenantID uuid.UUID, symbols []CachedSymbol) error {
	if len(symbols) == 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin symbol upsert transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, sym := range symbols {
		_, err := tx.Exec(ctx, `
			INSERT INTO symbol_cache (
				tenant_id, package_path, symbol_name, kind, receiver,
				signature_hash, ast_hash, payload, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
			ON CONFLICT (tenant_id, package_path, receiver, symbol_name) DO UPDATE SET
				kind = EXCLUDED.kind,
				signature_hash = EXCLUDED.signature_hash,
				ast_hash = EXCLUDED.ast_hash,
				payload = EXCLUDED.payload,
				updated_at = NOW()
		`, tenantID, sym.PackagePath, sym.SymbolName, sym.Kind, sym.Receiver,
			sym.SignatureHash, sym.ASTHash, sym.Payload)
		if err != nil {
			return fmt.Errorf("failed to upsert symbol %s.%s: %w", sym.PackagePath, sym.SymbolName, err)
		}
	}

	return tx.Commit(ctx)
}

// GetSymbolsByPackage fetches all cached symbols for a package.
func (s *PostgresStore) GetSymbolsByPackage(ctx context.Context, tenantID uuid.UUID, packagePath string) ([]CachedSymbol, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, tenant_id, package_path, symbol_name, kind, receiver, signature_hash, ast_hash, payload
		FROM symbol_cache
		WHERE tenant_id = $1 AND package_path = $2
	`, tenantID, packagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to query symbol cache: %w", err)
	}
	defer rows.Close()

	var symbols []CachedSymbol
	for rows.Next() {
		var sym CachedSymbol
		if err := rows.Scan(
			&sym.ID, &sym.TenantID, &sym.PackagePath, &sym.SymbolName,
			&sym.Kind, &sym.Receiver, &sym.SignatureHash, &sym.ASTHash, &sym.Payload,
		); err != nil {
			return nil, fmt.Errorf("failed to scan cached symbol: %w", err)
		}
		symbols = append(symbols, sym)
	}
	return symbols, rows.Err()
}

// BatchUpsertCrossModuleEdges records dependency relationships between modules.
func (s *PostgresStore) BatchUpsertCrossModuleEdges(ctx context.Context, tenantID uuid.UUID, edges []CrossModuleEdge) error {
	if len(edges) == 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin cross-module edge transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, e := range edges {
		_, err := tx.Exec(ctx, `
			INSERT INTO cross_module_edges (
				tenant_id, workspace_id, from_module, from_package, from_symbol,
				to_module, to_package, to_symbol, edge_type, confidence, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
			ON CONFLICT (workspace_id, from_package, from_symbol, to_package, to_symbol, edge_type) DO UPDATE SET
				confidence = EXCLUDED.confidence
		`, tenantID, e.WorkspaceID, e.FromModule, e.FromPackage, e.FromSymbol,
			e.ToModule, e.ToPackage, e.ToSymbol, e.EdgeType, e.Confidence)
		if err != nil {
			return fmt.Errorf("failed to insert cross-module edge: %w", err)
		}
	}

	return tx.Commit(ctx)
}
