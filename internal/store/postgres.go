// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myshra777-ai/garuda/internal/types"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore initializes the connection pool against the database target.
func NewPostgresStore(connString string) (*PostgresStore, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}
	return &PostgresStore{pool: pool}, nil
}

// Pool exposes the underlying connection pool for ledger queries and transactions.
func (s *PostgresStore) Pool() *pgxpool.Pool {
	return s.pool
}

// Close gracefully terminates active database connections.
func (s *PostgresStore) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// IngestEvidence commits structural governance evidence payloads inside an atomic transaction.
func (s *PostgresStore) IngestEvidence(ctx context.Context, tenantID uuid.UUID, evidence []types.Evidence) error {
	if len(evidence) == 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
        INSERT INTO evidence_store (tenant_id, block_hash, content, ref_count, created_at)
        VALUES ($1, $2, $3, 1, NOW())
        ON CONFLICT (tenant_id, block_hash) DO NOTHING;
    `
	for _, e := range evidence {
		if _, err := tx.Exec(ctx, query, tenantID, e.Hash[:], e.Content); err != nil {
			return fmt.Errorf("failed to ingest evidence %x: %w", e.Hash, err)
		}
	}
	return tx.Commit(ctx)
}

// IngestBlocks performs a content-addressable upsert for agent evidence blocks.
func (s *PostgresStore) IngestBlocks(ctx context.Context, blocks []types.Block) error {
	if len(blocks) == 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for blocks: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
        INSERT INTO blocks (hash, content, ref_count, created_at)
        VALUES ($1, $2, $3, NOW())
        ON CONFLICT (hash) DO UPDATE 
        SET ref_count = blocks.ref_count + EXCLUDED.ref_count;
    `

	batch := &pgx.Batch{}
	for _, b := range blocks {
		batch.Queue(query, b.Hash[:], b.Content, b.RefCount)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(blocks); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed executing block ingestion batch at index %d: %w", i, err)
		}
	}

	return tx.Commit(ctx)
}

// SaveTaskManifest handles full upsert transitions for agent executions and universal context states.
func (s *PostgresStore) SaveTaskManifest(ctx context.Context, m *types.TaskManifest) error {
	query := `
        INSERT INTO task_manifests (
            task_id, customer_id, credential_ref, title, scope_domain, scope_system, 
            status, manifest_blocks, normalized_ir, ir_version, decision_ids, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
        ON CONFLICT (task_id) DO UPDATE SET
            credential_ref  = EXCLUDED.credential_ref,
            title           = EXCLUDED.title,
            scope_domain    = EXCLUDED.scope_domain,
            scope_system    = EXCLUDED.scope_system,
            status          = EXCLUDED.status,
            manifest_blocks = EXCLUDED.manifest_blocks,
            normalized_ir   = EXCLUDED.normalized_ir,
            ir_version      = EXCLUDED.ir_version,
            decision_ids    = EXCLUDED.decision_ids,
            updated_at      = NOW();
    `

	blockSlices := make([][]byte, len(m.ManifestBlocks))
	for i, hash := range m.ManifestBlocks {
		blockSlices[i] = hash[:]
	}

	_, err := s.pool.Exec(ctx, query,
		m.TaskID,
		m.CustomerID,
		m.CredentialRef,
		m.Title,
		m.ScopeDomain,
		m.ScopeSystem,
		string(m.Status),
		blockSlices,
		m.NormalizedIR,
		m.IRVersion,
		m.DecisionIDs,
	)

	if err != nil {
		return fmt.Errorf("failed to save task manifest %s: %w", m.TaskID, err)
	}
	return nil
}

// ListDecisions queries decisions filtered by scope parameters and allowable status states.
func (s *PostgresStore) ListDecisions(ctx context.Context, tenantID uuid.UUID, scope types.Scope, statuses []types.DecisionStatus) ([]*types.Decision, error) {
	if s == nil || s.pool == nil {
		return nil, fmt.Errorf("postgres store is not initialized")
	}

	statusStrs := make([]string, len(statuses))
	for i, st := range statuses {
		statusStrs[i] = string(st)
	}

	query := `
        SELECT id, tenant_id, title, statement, status, domain, system, team, env,
               owner, confidence, fingerprint, parent_id, temporal_metadata,
               created_at, updated_at, approved_at
        FROM decisions
        WHERE `
	args := []any{}
	paramIndex := 1
	if tenantID != uuid.Nil {
		query += fmt.Sprintf(`tenant_id = $%d AND `, paramIndex)
		args = append(args, tenantID)
		paramIndex++
	}
	query += fmt.Sprintf(`($%d = '' OR domain = $%d)`, paramIndex, paramIndex)
	args = append(args, scope.Domain)
	paramIndex++
	query += fmt.Sprintf(` AND ($%d = '' OR system = $%d)`, paramIndex, paramIndex)
	args = append(args, scope.System)
	paramIndex++
	query += fmt.Sprintf(` AND ($%d = '' OR team = $%d)`, paramIndex, paramIndex)
	args = append(args, scope.Team)
	paramIndex++
	query += fmt.Sprintf(` AND ($%d = '' OR env = $%d)`, paramIndex, paramIndex)
	args = append(args, scope.Env)
	paramIndex++
	query += fmt.Sprintf(` AND (cardinality($%d::text[]) = 0 OR status = ANY($%d))`, paramIndex, paramIndex)
	args = append(args, statusStrs)
	query += `
        ORDER BY created_at DESC;
    `

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list decisions: %w", err)
	}
	defer rows.Close()

	var results []*types.Decision
	for rows.Next() {
		var d types.Decision
		var statusStr string
		err := rows.Scan(
			&d.ID, &d.TenantID, &d.Title, &d.Statement, &statusStr,
			&d.Scope.Domain, &d.Scope.System, &d.Scope.Team, &d.Scope.Env,
			&d.Owner, &d.Confidence, &d.Fingerprint, &d.ParentID,
			&d.TemporalMetadata, &d.CreatedAt, &d.UpdatedAt, &d.ApprovedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning decision row: %w", err)
		}
		d.Status = types.DecisionStatus(statusStr)
		results = append(results, &d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading decisions query results: %w", err)
	}

	return results, nil
}

// ListByScope returns decisions filtered by scope and status.
func (s *PostgresStore) ListByScope(ctx context.Context, tenantID uuid.UUID, scope types.Scope, statuses []types.DecisionStatus) ([]*types.Decision, error) {
	return s.ListDecisions(ctx, tenantID, scope, statuses)
}

// ═══════════════════════════════════════════════════════════════════════════════════
//  IMPORTANT: SaveAnalysisDecision has been removed from this file.
//  It is now located in internal/store/analysis.go to maintain a clean separation
//  of concerns and avoid duplicate method definitions.
//  The implementation there matches the signature required by main.go:
//     func (s *PostgresStore) SaveAnalysisDecision(ctx context.Context, tenantID uuid.UUID, result *analyzer.Result) (uuid.UUID, int, error)
// ═══════════════════════════════════════════════════════════════════════════════════
