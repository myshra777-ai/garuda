-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- 081_telemetry_aggregates.sql
--
-- Daily aggregates of operational counters. Read by the admin
-- dashboard. Writes only from RefreshAggregates.
--
-- Why this table exists. The admin dashboard needs counts of signups,
-- active users, workspaces, and telemetry totals. Computing those by
-- scanning the content tables on every page load would:
--   - scale linearly with the number of events (telemetry_events has
--     unbounded growth)
--   - require the admin dashboard to read tables that carry user or
--     workspace content, which is the boundary this schema exists to
--     enforce.
--
-- The privacy boundary is structural: this table has no user_id,
-- repository_id, entity_id, file_path, or content column. A query
-- against it cannot answer "which user did X" because the schema has
-- no way to express it.
--
-- Two scopes coexist:
--   tenant_id IS NOT NULL  per-tenant metric (workspaces_created,
--                          active_users)
--   tenant_id IS NULL      global metric (signups, tenants_created,
--                          telemetry_event_count, tokens_saved,
--                          cost_saved_usd, agents_active)
--
-- Two partial unique indexes because Postgres treats NULLs as
-- distinct in a unique index. A global row and a tenant row for the
-- same (bucket_date, metric) cannot collide.

CREATE TABLE IF NOT EXISTS telemetry_aggregates (
    bucket_date  DATE NOT NULL,
    tenant_id    UUID REFERENCES tenants(id) ON DELETE CASCADE,
    metric       TEXT NOT NULL,
    value        BIGINT NOT NULL DEFAULT 0,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_telemetry_aggregates_scoped
    ON telemetry_aggregates (bucket_date, tenant_id, metric)
    WHERE tenant_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_telemetry_aggregates_global
    ON telemetry_aggregates (bucket_date, metric)
    WHERE tenant_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_telemetry_aggregates_bucket
    ON telemetry_aggregates (bucket_date DESC);