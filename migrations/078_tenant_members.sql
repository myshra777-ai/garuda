-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- 078_tenant_members.sql
--
-- Introduces tenant_members. One row per (user, tenant). Replaces the
-- practice of inferring tenant membership from the fact that a user
-- is logged in.
--
-- Backfill: every existing user is added as an 'owner' of the
-- canonical tenant. At this point in the sequence there is exactly
-- one user (admin@local) and exactly one tenant (canonical). After
-- Session B adds /signup, new users will create their own tenants
-- and get an 'owner' row there instead.

CREATE TABLE IF NOT EXISTS tenant_members (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role        TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'member')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, tenant_id)
);

CREATE INDEX IF NOT EXISTS idx_tenant_members_tenant
    ON tenant_members (tenant_id);

-- Backfill: every existing user becomes owner of the canonical tenant.
-- ON CONFLICT DO NOTHING so a re-run is a no-op.
INSERT INTO tenant_members (user_id, tenant_id, role, created_at)
SELECT u.id, '00000000-0000-0000-0000-000000000001', 'owner', NOW()
  FROM users u
ON CONFLICT (user_id, tenant_id) DO NOTHING;

-- Verify: every user has at least one tenant_members row after the
-- backfill. A user without a tenant cannot log in through the new
-- middleware (Session C), so the backfill must be complete before
-- that session ships. Fail loudly here rather than discover the gap
-- in Session C.
DO $$
DECLARE
  orphan_count INT;
BEGIN
  SELECT COUNT(*) INTO orphan_count
    FROM users u
   WHERE NOT EXISTS (SELECT 1 FROM tenant_members tm WHERE tm.user_id = u.id);
  IF orphan_count > 0 THEN
    RAISE EXCEPTION '078: % users have no tenant_members row after backfill', orphan_count;
  END IF;
END $$;