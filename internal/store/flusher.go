// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

// StartRefCountFlusher kicks off an unblocked polling routine that aggregates reference updates out of Redis Streams.
func (s *PostgresStore) StartRefCountFlusher(ctx context.Context, rdb *redis.Client, stream, group string) error {
	// Ensure the stream and consumer group exist
	if err := EnsureStreamGroup(ctx, rdb, stream, group); err != nil {
		return fmt.Errorf("failed to verify structural stream parameters: %w", err)
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Halting reference flusher engine loop via context cancel")
			return nil
		case <-ticker.C:
			if err := s.flush(ctx, rdb, stream, group); err != nil {
				slog.Error("Failed to flush aggregated reference counts down to storage layer", "error", err)
			}
		}
	}
}

func (s *PostgresStore) flush(ctx context.Context, rdb *redis.Client, stream, group string) error {
	// 1. Read pending entries from the stream (consuming unacknowledged nodes)
	entries, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: "garuda_worker_1",
		Streams:  []string{stream, ">"},
		Count:    1000,
		Block:    1 * time.Second,
	}).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("redis stream ingestion bottleneck encountered: %w", err)
	}

	// 2. Aggregate deltas cleanly in memory
	deltas := make(map[string]int)
	msgIDs := []string{}

	for _, streamEntry := range entries {
		for _, msg := range streamEntry.Messages {
			hashStr, ok := msg.Values["hash"].(string)
			if !ok {
				slog.Warn("Skipping stream frame due to missing or invalid hash type mapping", "msg_id", msg.ID)
				continue
			}

			// Redis stream numbers cross wire vectors as strings; parse explicitly to prevent type assertion crashes
			deltaStr, ok := msg.Values["delta"].(string)
			if !ok {
				slog.Warn("Skipping stream frame due to missing delta payload attributes", "msg_id", msg.ID)
				continue
			}

			delta, err := strconv.Atoi(deltaStr)
			if err != nil {
				slog.Error("Malformed delta digit encoding discovered inside message packet", "msg_id", msg.ID, "raw", deltaStr)
				continue
			}

			deltas[hashStr] += delta
			msgIDs = append(msgIDs, msg.ID)
		}
	}

	if len(deltas) == 0 {
		return nil
	}

	// 3. Batch commit delta updates to Postgres
	if err := s.commitEvidenceDeltas(ctx, deltas); err != nil {
		return fmt.Errorf("failed to commit structural reference update pass: %w", err)
	}

	// 4. Acknowledge processed messages back to Redis broker
	if err := rdb.XAck(ctx, stream, group, msgIDs...).Err(); err != nil {
		return fmt.Errorf("failed to acknowledge aggregated stream messages: %w", err)
	}

	return nil
}

func (s *PostgresStore) commitEvidenceDeltas(ctx context.Context, deltas map[string]int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start database delta transaction context: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
        UPDATE evidence_store SET ref_count = ref_count + $1 WHERE block_hash = $2;
    `

	// Use pipelined batches to maximize query performance on database connections
	batch := &pgx.Batch{}
	activeOperationsCount := 0

	for hashStr, delta := range deltas {
		if delta == 0 {
			continue
		}

		hashBytes, err := hex.DecodeString(hashStr)
		if err != nil {
			return fmt.Errorf("corrupted hexadecimal evidence key hash parsed: %w", err)
		}

		batch.Queue(query, delta, hashBytes)
		activeOperationsCount++
	}

	if activeOperationsCount == 0 {
		return nil
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < activeOperationsCount; i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("pipelined execution failed at position index %d: %w", i, err)
		}
	}

	return tx.Commit(ctx)
}

// EnsureStreamGroup handles structural safety initializations for Kafka-style streams inside Redis.
func EnsureStreamGroup(ctx context.Context, rdb *redis.Client, stream, group string) error {
	err := rdb.XGroupCreateMkStream(ctx, stream, group, "$").Err()
	if err != nil {
		// Suppress errors stating that the consumer group already exists in the metadata register
		if err.Error() == "BUSYGROUP Consumer Group name already exists" {
			return nil
		}
		return err
	}
	return nil
}
