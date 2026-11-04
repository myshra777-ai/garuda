// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"context"
	"fmt"

	"github.com/myshra777-ai/garuda/internal/types"
	"github.com/redis/go-redis/v9"
)

type RefCountManager struct {
	rdb *redis.Client
}

func NewRefCountManager(rdb *redis.Client) *RefCountManager {
	return &RefCountManager{rdb: rdb}
}

func (rm *RefCountManager) EmitReferenceChange(ctx context.Context, evidenceHash types.EvidenceHash, delta int) error {
	hashBytes := evidenceHash[:]
	err := rm.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "garuda:ref_changes",
		Values: map[string]interface{}{
			"hash":  hashBytes,
			"delta": delta,
		},
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to emit reference change to Redis stream: %w", err)
	}
	return nil
}
