-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- 077_tenants.sql
--
-- Introduces the tenants table. Before this migration, tenant identity
-- existed only as a UUID on workspaces.tenant_id and on other content
-- tables. There was no name, no creation timestamp, no way to list or
-- manage tenants.
--
-- This migration creates the table and seeds one row: the canonical
-- tenant that owns every workspace currently in the database. The
-- canonical UUID matches the constant that has been hardcoded across
-- the codebase (internal/tenant/tenant.go) since before this session.
-- After Session A, that constant becomes a backfill marker rather
-- than the runtime source of truth.

CREATE TABLE IF NOT EXISTS tenants (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed the canonical tenant. ON CONFLICT DO NOTHING so re-running on a
-- database that already has the row is a no-op.
INSERT INTO tenants (id, name, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'canonical',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO NOTHING;

-- Verify every workspace's tenant_id corresponds to a row in tenants.
-- Any orphan is a data-integrity problem: it means a workspace was
-- created under a tenant that does not exist. Abort rather than leave
-- the database in a state where the FK cannot be added cleanly.
DO $$
DECLARE
  orphan_count INT;
BEGIN
  SELECT COUNT(*) INTO orphan_count
    FROM workspaces w
   WHERE NOT EXISTS (SELECT 1 FROM tenants t WHERE t.id = w.tenant_id);
  IF orphan_count > 0 THEN
    RAISE EXCEPTION '077: % workspaces reference a tenant_id with no tenants row', orphan_count;
  END IF;
END $$;