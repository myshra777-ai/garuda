// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Metric keys written by RefreshAggregates. Named constants so a grep
// for a metric finds every reader and writer. Adding a metric is a
// code change here, not a migration: the metric column is TEXT.
const (
	MetricSignups             = "signups"               // global
	MetricTenantsCreated      = "tenants_created"       // global
	MetricTelemetryEventCount = "telemetry_event_count" // global
	MetricTokensSaved         = "tokens_saved"          // global
	MetricCostSavedUSDCents   = "cost_saved_usd_cents"  // global
	MetricAgentsActive        = "agents_active"         // global
	MetricActiveUsersGlobal   = "active_users_global"   // global — distinct users with a login that day
	MetricWorkspacesCreated   = "workspaces_created"    // per-tenant
	MetricActiveUsers         = "active_users"          // per-tenant

	// Current-snapshot metrics. These are written with today's
	// bucket_date on every refresh and overwritten each cycle. They
	// are what the admin dashboard shows as "total users" and "total
	// workspaces" without reading a table that carries user content.
	// Yesterday's snapshot is frozen: it records what the total was
	// when the last refresh of that day ran.
	MetricUsersTotalCurrent      = "users_total_current"
	MetricTenantsTotalCurrent    = "tenants_total_current"
	MetricWorkspacesTotalCurrent = "workspaces_total_current"
)

// RefreshAggregates computes every metric for the given date and
// upserts the result.
//
// Idempotent. Running it twice for the same date produces the same
// rows because every write is an ON CONFLICT DO UPDATE keyed on
// (bucket_date, tenant_id, metric).
//
// Safe to run concurrently with itself and with other daemon
// instances: the unique indexes and the upsert guarantee convergence.
//
// The date is interpreted in UTC. A "day" is 00:00:00Z to 23:59:59Z.
// Cross-timezone correctness is a later concern; the aggregate is a
// product metric, not a billing record.
func (s *PostgresStore) RefreshAggregates(ctx context.Context, date time.Time) error {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)

	// 1. Global metrics — one row each, tenant_id = NULL.
	globals := []struct {
		metric string
		query  string
	}{
		{
			MetricSignups,
			`SELECT COUNT(*)::bigint FROM users
			  WHERE created_at >= $1 AND created_at < $2`,
		},
		{
			MetricTenantsCreated,
			`SELECT COUNT(*)::bigint FROM tenants
			  WHERE created_at >= $1 AND created_at < $2`,
		},
		{
			MetricTelemetryEventCount,
			`SELECT COUNT(*)::bigint FROM telemetry_events
			  WHERE created_at >= $1 AND created_at < $2`,
		},
		{
			MetricTokensSaved,
			`SELECT COALESCE(SUM(tokens_saved), 0)::bigint FROM telemetry_events
			  WHERE created_at >= $1 AND created_at < $2`,
		},
		{
			MetricCostSavedUSDCents,
			`SELECT COALESCE(ROUND(SUM(cost_saved_usd) * 100), 0)::bigint FROM telemetry_events
			  WHERE created_at >= $1 AND created_at < $2`,
		},
		{
			MetricAgentsActive,
			`SELECT COUNT(DISTINCT agent_runtime)::bigint FROM telemetry_events
			  WHERE created_at >= $1 AND created_at < $2
			    AND agent_runtime IS NOT NULL`,
		},

		{
			MetricSignups,
			`SELECT COUNT(*)::bigint FROM users
			  WHERE created_at >= $1 AND created_at < $2`,
		},
		{
			MetricTenantsCreated,
			`SELECT COUNT(*)::bigint FROM tenants
			  WHERE created_at >= $1 AND created_at < $2`,
		},
		{
			MetricTelemetryEventCount,
			`SELECT COUNT(*)::bigint FROM telemetry_events
			  WHERE created_at >= $1 AND created_at < $2`,
		},
		{
			MetricTokensSaved,
			`SELECT COALESCE(SUM(tokens_saved), 0)::bigint FROM telemetry_events
			  WHERE created_at >= $1 AND created_at < $2`,
		},
		{
			MetricCostSavedUSDCents,
			`SELECT COALESCE(ROUND(SUM(cost_saved_usd) * 100), 0)::bigint FROM telemetry_events
			  WHERE created_at >= $1 AND created_at < $2`,
		},
		{
			MetricAgentsActive,
			`SELECT COUNT(DISTINCT agent_runtime)::bigint FROM telemetry_events
			  WHERE created_at >= $1 AND created_at < $2
			    AND agent_runtime IS NOT NULL`,
		},
		{
			MetricActiveUsersGlobal,
			`SELECT COUNT(*)::bigint FROM users
			  WHERE last_login_at >= $1 AND last_login_at < $2`,
		},
	}

	// Current-snapshot metrics. Same bucket_date (today) on every
	// refresh; each write overwrites the previous value for that
	// date. Yesterday's row is not touched, so it holds what the
	// total was at the last refresh of that day.
	snapshots := []struct {
		metric string
		query  string
	}{
		{MetricUsersTotalCurrent, `SELECT COUNT(*)::bigint FROM users`},
		{MetricTenantsTotalCurrent, `SELECT COUNT(*)::bigint FROM tenants`},
		{MetricWorkspacesTotalCurrent, `SELECT COUNT(*)::bigint FROM workspaces`},
	}
	for _, snap := range snapshots {
		var v int64
		if err := s.pool.QueryRow(ctx, snap.query).Scan(&v); err != nil {
			return fmt.Errorf("snapshot %s: %w", snap.metric, err)
		}
		if err := s.upsertAggregate(ctx, dayStart, nil, snap.metric, v); err != nil {
			return err
		}
	}

	for _, g := range globals {
		var v int64
		if err := s.pool.QueryRow(ctx, g.query, dayStart, dayEnd).Scan(&v); err != nil {
			return fmt.Errorf("aggregate %s: %w", g.metric, err)
		}
		if err := s.upsertAggregate(ctx, dayStart, nil, g.metric, v); err != nil {
			return err
		}
	}

	// 2. Per-tenant metrics — one row per tenant per metric.
	// Iterate every tenant. For the beta there is one; the loop is
	// here because the schema supports many, and skipping it would
	// require a rewrite the moment a second tenant exists.
	tenantRows, err := s.pool.Query(ctx, `SELECT id FROM tenants`)
	if err != nil {
		return fmt.Errorf("list tenants for aggregates: %w", err)
	}
	var tenantIDs []uuid.UUID
	for tenantRows.Next() {
		var id uuid.UUID
		if err := tenantRows.Scan(&id); err == nil {
			tenantIDs = append(tenantIDs, id)
		}
	}
	tenantRows.Close()

	for _, tid := range tenantIDs {
		tenantID := tid

		// workspaces_created: workspaces in this tenant, created that day.
		var wc int64
		if err := s.pool.QueryRow(ctx, `
			SELECT COUNT(*)::bigint FROM workspaces
			 WHERE tenant_id = $1
			   AND created_at >= $2 AND created_at < $3
		`, tenantID, dayStart, dayEnd).Scan(&wc); err != nil {
			return fmt.Errorf("aggregate workspaces_created for %s: %w", tenantID, err)
		}
		if err := s.upsertAggregate(ctx, dayStart, &tenantID, MetricWorkspacesCreated, wc); err != nil {
			return err
		}

		// active_users: distinct users who are members of this tenant
		// and who logged in during the day.
		var au int64
		if err := s.pool.QueryRow(ctx, `
			SELECT COUNT(DISTINCT u.id)::bigint
			  FROM users u
			  JOIN tenant_members tm ON tm.user_id = u.id
			 WHERE tm.tenant_id = $1
			   AND u.last_login_at >= $2 AND u.last_login_at < $3
		`, tenantID, dayStart, dayEnd).Scan(&au); err != nil {
			return fmt.Errorf("aggregate active_users for %s: %w", tenantID, err)
		}
		if err := s.upsertAggregate(ctx, dayStart, &tenantID, MetricActiveUsers, au); err != nil {
			return err
		}
	}

	return nil
}

// upsertAggregate writes one row. tenantID nil means a global row.
//
// The ON CONFLICT target differs between the two cases because the
// unique indexes are partial. Postgres cannot bind a bare target to
// a partial index; the WHERE clause on the conflict target must
// mirror the index predicate.
func (s *PostgresStore) upsertAggregate(
	ctx context.Context,
	bucketDate time.Time,
	tenantID *uuid.UUID,
	metric string,
	value int64,
) error {
	if tenantID == nil {
		_, err := s.pool.Exec(ctx, `
			INSERT INTO telemetry_aggregates (bucket_date, tenant_id, metric, value, updated_at)
			VALUES ($1::date, NULL, $2, $3, NOW())
			ON CONFLICT (bucket_date, metric)
				WHERE tenant_id IS NULL
			DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
		`, bucketDate, metric, value)
		if err != nil {
			return fmt.Errorf("upsert global aggregate %s: %w", metric, err)
		}
		return nil
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO telemetry_aggregates (bucket_date, tenant_id, metric, value, updated_at)
		VALUES ($1::date, $2, $3, $4, NOW())
		ON CONFLICT (bucket_date, tenant_id, metric)
			WHERE tenant_id IS NOT NULL
		DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
	`, bucketDate, *tenantID, metric, value)
	if err != nil {
		return fmt.Errorf("upsert tenant aggregate %s for %s: %w", metric, *tenantID, err)
	}
	return nil
}
