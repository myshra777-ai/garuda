package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/techtaytor/garuda/internal/types"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(connString string) (*PostgresStore, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}
	return &PostgresStore{pool: pool}, nil
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}

func (s *PostgresStore) IngestEvidence(ctx context.Context, tenantID uuid.UUID, evidence []types.Evidence) error {
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
