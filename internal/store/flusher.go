package store

import (
	"context"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

func (s *PostgresStore) StartRefCountFlusher(ctx context.Context, rdb *redis.Client, stream, group string) error {
	// Ensure the stream and consumer group exist
	if err := EnsureStreamGroup(ctx, rdb, stream, group); err != nil {
		return err
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.flush(ctx, rdb, stream, group); err != nil {
				slog.Error("failed to flush reference counts", "error", err)
			}
		}
	}
}

func (s *PostgresStore) flush(ctx context.Context, rdb *redis.Client, stream, group string) error {
	// 1. Read pending entries from the stream
	entries, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: "worker_1",
		Streams:  []string{stream, ">"},
		Count:    1000,
		Block:    1 * time.Second,
	}).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}

	// 2. Aggregate deltas in memory
	deltas := make(map[string]int)
	msgIDs := []string{}
	for _, streamEntry := range entries {
		for _, msg := range streamEntry.Messages {
			hashStr, ok := msg.Values["hash"].(string)
			if !ok {
				continue
			}
			delta, ok := msg.Values["delta"].(int)
			if !ok {
				continue
			}
			deltas[hashStr] += delta
			msgIDs = append(msgIDs, msg.ID)
		}
	}

	if len(deltas) == 0 {
		return nil
	}

	// 3. Batch commit to Postgres
	if err := s.commitEvidenceDeltas(ctx, deltas); err != nil {
		return err
	}

	// 4. Acknowledge processed messages
	return rdb.XAck(ctx, stream, group, msgIDs...).Err()
}

func (s *PostgresStore) commitEvidenceDeltas(ctx context.Context, deltas map[string]int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE evidence_store SET ref_count = ref_count + $1 WHERE block_hash = $2
	`
	for hashStr, delta := range deltas {
		if delta == 0 {
			continue
		}
		hashBytes, err := hex.DecodeString(hashStr)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, query, delta, hashBytes)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
