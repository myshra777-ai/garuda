// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/knowledge"
)

type ClaimStore struct {
	pool *pgxpool.Pool
}

func NewClaimStore(pool *pgxpool.Pool) *ClaimStore {
	return &ClaimStore{pool: pool}
}

// ResolveWorkspaceID returns the UUID of the named workspace, or the
// most recently updated workspace for the tenant if name is empty.
//
// The empty-name case is the "no explicit choice" path: a caller that
// does not know which workspace to use gets the most recently active
// one. This matches the empty-name branch of the dashboard's
// resolveWorkspaceID. It replaces the previous pattern where callers
// defaulted to the literal string "default", which does not correspond
// to any workspace row unless `garuda init` has seeded it.
//
// A non-empty name that does not match any workspace is an error. The
// caller asked for something specific; silently returning zero rows
// would be the same class of lie the workspace_id migration was
// introduced to end.
//
// Every lookup is tenant-scoped. A fallback that drops the tenant
// filter would allow a caller in tenant A to resolve a workspace
// belonging to tenant B, which is the class of bug fixed in
// resolveWorkspaceID at the api layer and in the tenant package.
func ResolveWorkspaceID(ctx context.Context, pool *pgxpool.Pool, tenantID uuid.UUID, name string) (uuid.UUID, error) {
	var id uuid.UUID

	if name == "" {
		err := pool.QueryRow(ctx, `
			SELECT id FROM workspaces
			 WHERE tenant_id = $1
			 ORDER BY updated_at DESC, created_at DESC, id DESC
			 LIMIT 1
		`, tenantID).Scan(&id)
		if err != nil {
			return uuid.Nil, fmt.Errorf("no workspaces exist for tenant %s", tenantID)
		}
		return id, nil
	}

	err := pool.QueryRow(ctx, `
		SELECT id FROM workspaces
		 WHERE tenant_id = $1 AND name = $2
		 ORDER BY created_at ASC, id ASC
		 LIMIT 1
	`, tenantID, name).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("workspace %q not found for tenant %s", name, tenantID)
	}
	return id, nil
}

// SaveClaims clears existing claims for the target document path and persists the fresh batch atomically.
func (s *ClaimStore) SaveClaims(ctx context.Context, claims []knowledge.ClaimIR) error {
	if len(claims) == 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Group by document path and extract scope metadata
	docPaths := make(map[string]bool)
	var tenantID uuid.UUID
	var workspace string

	for _, c := range claims {
		docPaths[c.DocumentPath] = true
		tenantID = c.TenantID
		workspace = c.Workspace
	}

	// Resolve the workspace UUID once. Migration 075 made
	// document_claims.workspace_id NOT NULL; every INSERT must carry
	// it. The name is retained in the workspace column for display and
	// for the transition period until every reader uses workspace_id.
	//
	// A workspace name that does not resolve is an error, not a silent
	// default. The previous code would have inserted under a stale
	// string and later returned zero rows to every reader.
	workspaceID, err := ResolveWorkspaceID(ctx, s.pool, tenantID, workspace)
	if err != nil {
		return fmt.Errorf("resolve workspace %q: %w", workspace, err)
	}

	for path := range docPaths {
		// Scoped by workspace_id. Migration 075 backfilled every row and
		// set NOT NULL, so no NULL rows remain to fall back to.
		_, err := tx.Exec(ctx, `
			DELETE FROM document_claims 
			WHERE tenant_id = $1
			  AND workspace_id = $2
			  AND document_path = $3
		`, tenantID, workspaceID, path)
		if err != nil {
			return fmt.Errorf("failed to clear old claims for %s: %w", path, err)
		}
	}

	stmt := `
		INSERT INTO document_claims (
			id, workspace, workspace_id, tenant_id, document_path, document_title, document_type,
			section_title, line_start, line_end, raw_statement, subject,
			modality, predicate, object, provenance_class, confidence, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
	`

	for _, c := range claims {
		_, err := tx.Exec(ctx, stmt,
			c.ID,
			c.Workspace,
			workspaceID,
			c.TenantID,
			c.DocumentPath,
			c.DocumentTitle,
			c.DocumentType,
			c.SectionTitle,
			c.LineStart,
			c.LineEnd,
			c.RawStatement,
			c.Subject,
			string(c.Modality),
			string(c.Predicate),
			c.Object,
			string(c.Provenance),
			c.Confidence,
			string(c.Status),
		)
		if err != nil {
			return fmt.Errorf("failed to insert claim %s: %w", c.ID, err)
		}
	}

	return tx.Commit(ctx)
}

// GetClaimsByWorkspace resolves the workspace name to its UUID, then
// returns every document claim scoped to that workspace.
//
// This function has no callers in the current tree. It is retained
// because the resolution path it demonstrates is the pattern every
// future reader should follow. Delete it if a future commit removes
// it for hygiene.
func (s *ClaimStore) GetClaimsByWorkspace(ctx context.Context, tenantID uuid.UUID, workspace string) ([]knowledge.ClaimIR, error) {
	workspaceID, err := ResolveWorkspaceID(ctx, s.pool, tenantID, workspace)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace %q: %w", workspace, err)
	}

	query := `
		SELECT 
			id, tenant_id, document_path, document_title, document_type,
			section_title, line_start, line_end, raw_statement, subject,
			modality, predicate, object, provenance_class, confidence, status,
			matched_entity_id, contradiction_reason, created_at
		FROM document_claims
		WHERE tenant_id = $1 AND workspace_id = $2
		ORDER BY document_path, line_start ASC
	`

	rows, err := s.pool.Query(ctx, query, tenantID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query document claims: %w", err)
	}
	defer rows.Close()

	var results []knowledge.ClaimIR
	for rows.Next() {
		var c knowledge.ClaimIR
		var mod, pred, prov, stat string
		err := rows.Scan(
			&c.ID, &c.TenantID, &c.DocumentPath, &c.DocumentTitle, &c.DocumentType,
			&c.SectionTitle, &c.LineStart, &c.LineEnd, &c.RawStatement, &c.Subject,
			&mod, &pred, &c.Object, &prov, &c.Confidence, &stat,
			&c.MatchedEntityID, &c.ContradictionMsg, &c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan claim row: %w", err)
		}
		c.Modality = knowledge.Modality(mod)
		c.Predicate = knowledge.Predicate(pred)
		c.Provenance = knowledge.ProvenanceClass(prov)
		c.Status = knowledge.VerificationStatus(stat)
		results = append(results, c)
	}

	return results, rows.Err()
}
