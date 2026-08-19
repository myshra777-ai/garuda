// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/merkle"
	"github.com/myshra777-ai/garuda/internal/store"
	"github.com/myshra777-ai/garuda/internal/types"
)

type WorkerConfig struct {
	Interval time.Duration
}

func main() {
	slog.Info("[GARUDA-WORKER] Initializing Merkle Root Snapshot Worker...")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/garuda?sslmode=disable"
	}

	interval := 30 * time.Second
	if envInterval := os.Getenv("SNAPSHOT_INTERVAL"); envInterval != "" {
		if d, err := time.ParseDuration(envInterval); err == nil {
			interval = d
		}
	}

	cfg := WorkerConfig{Interval: interval}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbStore, err := store.NewPostgresStore(dbURL)
	if err != nil {
		slog.Error("[GARUDA-WORKER] Database connection failure", "error", err)
		os.Exit(1)
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-stopChan
		slog.Info("[GARUDA-WORKER] Graceful shutdown initiated", "signal", sig)
		cancel()
	}()

	slog.Info("[GARUDA-WORKER] Daemon online", "interval", cfg.Interval)
	runWorker(ctx, dbStore, cfg)
}

func runWorker(ctx context.Context, dbStore *store.PostgresStore, cfg WorkerConfig) {
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	// Run initial snapshot cycle immediately upon startup
	if err := snapshotAllTenants(ctx, dbStore); err != nil {
		slog.Error("[GARUDA-WORKER] Initial snapshot cycle error", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("[GARUDA-WORKER] Worker loop terminated.")
			return
		case <-ticker.C:
			if err := snapshotAllTenants(ctx, dbStore); err != nil {
				slog.Error("[GARUDA-WORKER] Scheduled snapshot cycle error", "error", err)
			}
		}
	}
}

func snapshotAllTenants(ctx context.Context, dbStore *store.PostgresStore) error {
	slog.Info("[GARUDA-WORKER] Executing snapshot cycle...")

	tenantIDs, err := dbStore.ListAllTenants(ctx)
	if err != nil {
		return err
	}
	if len(tenantIDs) == 0 {
		slog.Info("[GARUDA-WORKER] No active tenant Merkle roots found.")
		return nil
	}

	successCount := 0
	for _, tenantID := range tenantIDs {
		if err := snapshotTenant(ctx, dbStore, tenantID); err != nil {
			slog.Error("[GARUDA-WORKER] Tenant snapshot failed", "tenant_id", tenantID, "error", err)
			continue
		}
		successCount++
	}

	slog.Info("[GARUDA-WORKER] Snapshot cycle complete", "processed_tenants", len(tenantIDs), "snapshots_saved", successCount)
	return nil
}

func snapshotTenant(ctx context.Context, dbStore *store.PostgresStore, tenantID uuid.UUID) error {
	root, err := dbStore.GetMerkleRoot(ctx, tenantID)
	if err != nil {
		return err
	}

	latest, _ := dbStore.GetLatestMerkleSnapshot(ctx, tenantID)
	var parentID *uuid.UUID
	parentHashStr := ""
	if latest != nil {
		parentID = &latest.ID
		parentHashStr = latest.SnapshotHash
	}

	now := time.Now().UTC()
	epochUnix := now.Unix()
	snapshotHash := merkle.SnapshotHash(tenantID, root.RootHash, root.BlockHeight, parentHashStr, epochUnix)

	snapshot := &types.MerkleSnapshot{
		ID:               uuid.New(),
		TenantID:         tenantID,
		RootHash:         root.RootHash,
		BlockHeight:      root.BlockHeight,
		ParentSnapshotID: parentID,
		SnapshotHash:     snapshotHash,
		EpochTimestamp:   now, // Pass time.Time (TIMESTAMPTZ) to match struct and DB schema
		CreatedAt:        now,
	}

	return dbStore.SaveMerkleSnapshot(ctx, snapshot)
}
