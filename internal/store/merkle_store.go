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
        SELECT tenant_id, root_hash, block_height, created_at, updated_at
        FROM merkle_roots
        WHERE tenant_id = $1;
    `
	var mr types.MerkleRoot
	err := s.pool.QueryRow(ctx, query, tenantID).Scan(
		&mr.TenantID, &mr.RootHash, &mr.BlockHeight, &mr.CreatedAt, &mr.UpdatedAt,
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

// ListAllTenants returns all distinct tenant IDs registered in merkle_roots (fast indexed lookup).
func (s *PostgresStore) ListAllTenants(ctx context.Context) ([]uuid.UUID, error) {
	query := `SELECT tenant_id FROM merkle_roots;`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenants from merkle_roots: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan tenant_id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// GetLatestMerkleSnapshot retrieves the most recent snapshot for a tenant matching Schema 016.
func (s *PostgresStore) GetLatestMerkleSnapshot(ctx context.Context, tenantID uuid.UUID) (*types.MerkleSnapshot, error) {
	query := `
		SELECT id, tenant_id, root_hash, block_height,
		       parent_snapshot_id, snapshot_hash, epoch_timestamp, created_at
		FROM merkle_snapshots
		WHERE tenant_id = $1
		ORDER BY epoch_timestamp DESC
		LIMIT 1;
	`
	var snap types.MerkleSnapshot
	var parentID *uuid.UUID
	err := s.pool.QueryRow(ctx, query, tenantID).Scan(
		&snap.ID, &snap.TenantID, &snap.RootHash, &snap.BlockHeight,
		&parentID, &snap.SnapshotHash, &snap.EpochTimestamp, &snap.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	snap.ParentSnapshotID = parentID
	return &snap, nil
}

// SaveMerkleSnapshot persists a new cryptographically chained Merkle snapshot.
func (s *PostgresStore) SaveMerkleSnapshot(ctx context.Context, snapshot *types.MerkleSnapshot) error {
	query := `
		INSERT INTO merkle_snapshots (
			id,
			tenant_id,
			root_hash,
			block_height,
			parent_snapshot_id,
			snapshot_hash,
			epoch_timestamp,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (tenant_id, snapshot_hash) DO NOTHING;
	`

	// Pass EpochTimestamp directly as time.Time (TIMESTAMPTZ column)
	epochTime := snapshot.EpochTimestamp
	if epochTime.IsZero() {
		epochTime = time.Now().UTC()
	}

	createdAt := snapshot.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	_, err := s.pool.Exec(
		ctx,
		query,
		snapshot.ID,
		snapshot.TenantID,
		snapshot.RootHash,
		snapshot.BlockHeight,
		snapshot.ParentSnapshotID,
		snapshot.SnapshotHash,
		epochTime,
		createdAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save merkle snapshot: %w", err)
	}

	return nil
}

// ListMerkleSnapshots returns the snapshot history for a tenant matching Schema 016.
func (s *PostgresStore) ListMerkleSnapshots(ctx context.Context, tenantID uuid.UUID, limit int) ([]types.MerkleSnapshot, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	query := `
		SELECT id, tenant_id, root_hash, block_height,
		       parent_snapshot_id, snapshot_hash, epoch_timestamp, created_at
		FROM merkle_snapshots
		WHERE tenant_id = $1
		ORDER BY epoch_timestamp DESC
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
			&snap.ID, &snap.TenantID, &snap.RootHash, &snap.BlockHeight,
			&parentID, &snap.SnapshotHash, &snap.EpochTimestamp, &snap.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan merkle snapshot: %w", err)
		}
		snap.ParentSnapshotID = parentID
		results = append(results, snap)
	}
	return results, nil
}
