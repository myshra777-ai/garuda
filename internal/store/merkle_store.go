// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/myshra777-ai/garuda/internal/merkle"
	"github.com/myshra777-ai/garuda/internal/types"
)

// GetMerkleRoot retrieves current root hash state or initializes a genesis root if missing.
func (s *PostgresStore) GetMerkleRoot(ctx context.Context, tenantID uuid.UUID) (*types.MerkleRoot, error) {
	query := `
		SELECT 
			tenant_id, 
			encode(root_hash, 'hex') AS root_hash, 
			block_height, 
			created_at, 
			updated_at
		FROM merkle_roots
		WHERE tenant_id = $1
		ORDER BY block_height DESC
		LIMIT 1;
	`

	var mr types.MerkleRoot
	err := s.pool.QueryRow(ctx, query, tenantID).Scan(
		&mr.TenantID,
		&mr.RootHash,
		&mr.BlockHeight,
		&mr.CreatedAt,
		&mr.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) || strings.Contains(err.Error(), "no rows") {
			return s.createGenesisMerkleRoot(ctx, tenantID)
		}
		return nil, fmt.Errorf("failed to fetch Merkle root: %w", err)
	}
	return &mr, nil
}

func (s *PostgresStore) createGenesisMerkleRoot(ctx context.Context, tenantID uuid.UUID) (*types.MerkleRoot, error) {
	genesisHash := merkle.HashDecision(uuid.Nil, "GENESIS_ROOT", "active", "system", "core", "garuda", nil)
	query := `
		INSERT INTO merkle_roots (tenant_id, root_hash, block_height, created_at, updated_at)
		VALUES ($1, $2, 0, NOW(), NOW())
		ON CONFLICT (tenant_id) DO NOTHING;
	`
	_, _ = s.pool.Exec(ctx, query, tenantID, genesisHash)

	return &types.MerkleRoot{
		TenantID:    tenantID,
		RootHash:    genesisHash,
		BlockHeight: 0,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}, nil
}

// AppendMerkleChain locks root row and extends the chain atomically.
func (s *PostgresStore) AppendMerkleChain(ctx context.Context, tenantID uuid.UUID, decisionHash string) (*types.MerkleRoot, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin Merkle transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var currentRoot string
	var currentHeight int64
	lockQuery := `
		SELECT root_hash, block_height FROM merkle_roots
		WHERE tenant_id = $1 FOR UPDATE;
	`
	err = tx.QueryRow(ctx, lockQuery, tenantID).Scan(&currentRoot, &currentHeight)
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) || (err != nil && strings.Contains(err.Error(), "no rows")) {
		genHash := merkle.HashDecision(uuid.Nil, "GENESIS_ROOT", "active", "system", "core", "garuda", nil)
		_, err = tx.Exec(ctx, `INSERT INTO merkle_roots (tenant_id, root_hash, block_height) VALUES ($1, $2, 0);`, tenantID, genHash)
		if err != nil {
			return nil, fmt.Errorf("failed to insert genesis root: %w", err)
		}
		currentRoot = genHash
		currentHeight = 0
	} else if err != nil {
		return nil, fmt.Errorf("failed to lock Merkle root row: %w", err)
	}

	newHash := merkle.ChainHash(currentRoot, decisionHash)
	newHeight := currentHeight + 1

	updateQuery := `
		UPDATE merkle_roots
		SET root_hash = $1, block_height = $2, updated_at = NOW()
		WHERE tenant_id = $3;
	`
	if _, err := tx.Exec(ctx, updateQuery, newHash, newHeight, tenantID); err != nil {
		return nil, fmt.Errorf("failed to update Merkle root: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit Merkle chain update: %w", err)
	}

	return &types.MerkleRoot{
		TenantID:    tenantID,
		RootHash:    newHash,
		BlockHeight: newHeight,
		UpdatedAt:   time.Now().UTC(),
	}, nil
}

// AddEvidenceBlock appends a new block to the immutable evidence chain.
func (s *PostgresStore) AddEvidenceBlock(ctx context.Context, tenantID, decisionID uuid.UUID, payload any) (*types.EvidenceBlock, error) {
	var prevHash string
	queryPrev := `
		SELECT evidence_hash FROM evidence_blocks
		WHERE tenant_id = $1 AND decision_id = $2
		ORDER BY created_at DESC LIMIT 1;
	`
	err := s.pool.QueryRow(ctx, queryPrev, tenantID, decisionID).Scan(&prevHash)
	if err != nil {
		prevHash = "0000000000000000000000000000000000000000000000000000000000000000"
	}

	bytes, _ := json.Marshal(payload)
	evidenceHash := fmt.Sprintf("%x", sha256.Sum256(bytes))

	block := &types.EvidenceBlock{
		ID:           uuid.New(),
		TenantID:     tenantID,
		DecisionID:   decisionID,
		PrevHash:     prevHash,
		EvidenceHash: evidenceHash,
		Payload:      payload,
		CreatedAt:    time.Now().UTC(),
	}

	insertQuery := `
		INSERT INTO evidence_blocks (id, tenant_id, decision_id, prev_hash, evidence_hash, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`
	_, err = s.pool.Exec(ctx, insertQuery,
		block.ID, block.TenantID, block.DecisionID, block.PrevHash, block.EvidenceHash, block.Payload, block.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert evidence block: %w", err)
	}

	return block, nil
}

// ListAllTenants returns all distinct tenant IDs from both merkle_roots and workspaces.
func (s *PostgresStore) ListAllTenants(ctx context.Context) ([]uuid.UUID, error) {
	query := `
		SELECT DISTINCT tenant_id FROM (
			SELECT tenant_id FROM merkle_roots WHERE tenant_id IS NOT NULL
			UNION
			SELECT tenant_id FROM workspaces WHERE tenant_id IS NOT NULL
		) t;
	`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenants: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		ids = append(ids, uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	}
	return ids, nil
}

// GetLatestMerkleSnapshot retrieves the most recent snapshot for a tenant.
func (s *PostgresStore) GetLatestMerkleSnapshot(ctx context.Context, tenantID uuid.UUID) (*types.MerkleSnapshot, error) {
	var snap types.MerkleSnapshot
	err := s.pool.QueryRow(ctx, `
		SELECT 
			id, tenant_id, parent_snapshot_id, block_height, snapshot_hash,
			COALESCE(static_root_hash, ''), COALESCE(runtime_root_hash, ''),
			COALESCE(runtime_leaf_count, 0), COALESCE(verified_claims_count, 0),
			COALESCE(contradicted_claims_count, 0), created_at
		FROM merkle_snapshots
		WHERE tenant_id = $1
		ORDER BY block_height DESC
		LIMIT 1
	`, tenantID).Scan(
		&snap.ID,
		&snap.TenantID,
		&snap.ParentSnapshotID,
		&snap.BlockHeight,
		&snap.SnapshotHash,
		&snap.StaticRootHash,
		&snap.RuntimeRootHash,
		&snap.RuntimeLeafCount,
		&snap.VerifiedClaimsCount,
		&snap.ContradictedClaimsCount,
		&snap.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Populate backward-compatible fields
	snap.RootHash = snap.SnapshotHash
	snap.EpochTimestamp = snap.CreatedAt.Unix()

	return &snap, nil
}

// SaveMerkleSnapshot persists a new cryptographically chained Merkle snapshot.
func (s *PostgresStore) SaveMerkleSnapshot(ctx context.Context, snapshot *types.MerkleSnapshot) error {
	query := `
		INSERT INTO merkle_snapshots (
			id,
			tenant_id,
			snapshot_hash,
			block_height,
			parent_snapshot_id,
			static_root_hash,
			runtime_root_hash,
			runtime_leaf_count,
			verified_claims_count,
			contradicted_claims_count,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (tenant_id, snapshot_hash) DO NOTHING;
	`

	createdAt := snapshot.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	hash := snapshot.SnapshotHash
	if hash == "" {
		hash = snapshot.RootHash
	}

	_, err := s.pool.Exec(
		ctx,
		query,
		snapshot.ID,
		snapshot.TenantID,
		hash,
		snapshot.BlockHeight,
		snapshot.ParentSnapshotID,
		snapshot.StaticRootHash,
		snapshot.RuntimeRootHash,
		snapshot.RuntimeLeafCount,
		snapshot.VerifiedClaimsCount,
		snapshot.ContradictedClaimsCount,
		createdAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save merkle snapshot: %w", err)
	}

	return nil
}

// ListMerkleSnapshots returns the snapshot history for a tenant up to the specified limit.
func (s *PostgresStore) ListMerkleSnapshots(ctx context.Context, tenantID uuid.UUID, limit int) ([]types.MerkleSnapshot, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	query := `
		SELECT id, tenant_id, parent_snapshot_id, block_height, snapshot_hash,
		       COALESCE(static_root_hash, ''), COALESCE(runtime_root_hash, ''),
		       COALESCE(runtime_leaf_count, 0), COALESCE(verified_claims_count, 0),
		       COALESCE(contradicted_claims_count, 0), created_at
		FROM merkle_snapshots
		WHERE tenant_id = $1
		ORDER BY block_height DESC
		LIMIT $2;
	`
	rows, err := s.pool.Query(ctx, query, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []types.MerkleSnapshot
	for rows.Next() {
		var snap types.MerkleSnapshot
		var parentID *uuid.UUID

		err := rows.Scan(
			&snap.ID, &snap.TenantID, &parentID, &snap.BlockHeight, &snap.SnapshotHash,
			&snap.StaticRootHash, &snap.RuntimeRootHash, &snap.RuntimeLeafCount,
			&snap.VerifiedClaimsCount, &snap.ContradictedClaimsCount, &snap.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan merkle snapshot: %w", err)
		}
		snap.ParentSnapshotID = parentID
		snap.RootHash = snap.SnapshotHash
		snap.EpochTimestamp = snap.CreatedAt.Unix()
		results = append(results, snap)
	}
	return results, nil
}

// CreateUnifiedMerkleSnapshot generates a dual-rooted Merkle snapshot covering static claims and runtime verifications.
func (s *PostgresStore) CreateUnifiedMerkleSnapshot(ctx context.Context, tenantID uuid.UUID) (*types.MerkleSnapshot, error) {
	// 1. Get latest snapshot for block chain continuity
	var prevHash string = "0000000000000000000000000000000000000000000000000000000000000000"
	var prevHeight int64 = 0
	var parentID *uuid.UUID

	latest, err := s.GetLatestMerkleSnapshot(ctx, tenantID)
	if err == nil && latest != nil {
		prevHash = latest.SnapshotHash
		prevHeight = latest.BlockHeight
		parentID = &latest.ID
	}
	currentHeight := prevHeight + 1

	// 2. Compute Static AST Root from Entities and Claims
	var staticRoot string
	var staticLeafCount int64
	err = s.pool.QueryRow(ctx, `
		SELECT 
			COALESCE(encode(sha256(string_agg(id::text, '|' ORDER BY id)::bytea), 'hex'), 'GARUDA_EMPTY_STATIC_TREE'),
			COUNT(*)
		FROM entities
		WHERE tenant_id = $1 OR workspace_id IN (SELECT id FROM workspaces WHERE tenant_id = $1)
	`, tenantID).Scan(&staticRoot, &staticLeafCount)
	if err != nil || staticRoot == "" {
		staticRoot = "GARUDA_EMPTY_STATIC_TREE"
	}

	// 3. Query Claim Verifications for Runtime Evidence Root
	rows, err := s.pool.Query(ctx, `
		SELECT 
			source_entity_id::text,
			target_entity_id::text,
			status,
			runtime_observed_count,
			reason,
			COALESCE(evidence_payload->>'raw_target', ''),
			COALESCE(evidence_payload->>'last_trace_id', '')
		FROM claim_verifications
		WHERE tenant_id = $1 OR workspace_id IN (SELECT id FROM workspaces WHERE tenant_id = $1)
		ORDER BY source_entity_id, target_entity_id
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch claim verifications: %w", err)
	}
	defer rows.Close()

	var runtimeLeaves []merkle.RuntimeLeaf
	var supportedCount, contradictedCount int64

	for rows.Next() {
		var l merkle.RuntimeLeaf
		if err := rows.Scan(&l.SourceID, &l.TargetID, &l.Status, &l.ObservedCount, &l.Reason, &l.RawTarget, &l.LastTraceID); err == nil {
			runtimeLeaves = append(runtimeLeaves, l)
			if l.Status == "SUPPORTED" {
				supportedCount++
			} else if l.Status == "CONTRADICTED" {
				contradictedCount++
			}
		}
	}

	runtimeRoot, runtimeLeafCount := merkle.BuildRuntimeMerkleRoot(runtimeLeaves)

	// 4. Compute Unified Epoch State Root
	unifiedHash := merkle.ComputeUnifiedEpochRoot(staticRoot, runtimeRoot, prevHash, currentHeight)
	now := time.Now().UTC()

	// 5. Persist Snapshot Record with full legacy & dual-root columns
	snapshotID := uuid.New()
	_, err = s.pool.Exec(ctx, `
		INSERT INTO merkle_snapshots (
			id, tenant_id, parent_snapshot_id, block_height,
			snapshot_hash, root_hash, epoch_timestamp,
			static_root_hash, runtime_root_hash, runtime_leaf_count,
			verified_claims_count, contradicted_claims_count, created_at
		) VALUES ($1, $2, $3, $4, $5, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (tenant_id, snapshot_hash) DO NOTHING;
	`, snapshotID, tenantID, parentID, currentHeight,
		unifiedHash, now.Unix(), staticRoot, runtimeRoot,
		runtimeLeafCount, supportedCount, contradictedCount, now)
	if err != nil {
		return nil, fmt.Errorf("failed to insert unified merkle snapshot: %w", err)
	}

	return &types.MerkleSnapshot{
		ID:                      snapshotID,
		TenantID:                tenantID,
		ParentSnapshotID:        parentID,
		BlockHeight:             currentHeight,
		SnapshotHash:            unifiedHash,
		RootHash:                unifiedHash,
		StaticRootHash:          staticRoot,
		RuntimeRootHash:         runtimeRoot,
		RuntimeLeafCount:        int64(runtimeLeafCount),
		VerifiedClaimsCount:     supportedCount,
		ContradictedClaimsCount: contradictedCount,
		EpochTimestamp:          now.Unix(),
		CreatedAt:               now,
	}, nil
}
