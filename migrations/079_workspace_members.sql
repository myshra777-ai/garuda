-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- 079_workspace_members.sql
--
-- Introduces workspace_members. One row per (user, workspace).
-- Enforces the team-level isolation described in the multi-tenant
-- roadmap: two teams under one tenant, each with a different set of
-- members, each seeing only their workspace.
--
-- Backfill: every existing user is added as an 'owner' of every
-- existing workspace in the canonical tenant. At this point in the
-- sequence there is exactly one user (admin@local) and two
-- workspaces (garuda-validation, go-validation-10). After Session B,
-- new users get one workspace_members row at signup, and further
-- rows only through invitation.

CREATE TABLE IF NOT EXISTS workspace_members (
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    role            TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'member')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, workspace_id)
);

CREATE INDEX IF NOT EXISTS idx_workspace_members_workspace
    ON workspace_members (workspace_id);

-- Backfill: every existing user is owner of every existing workspace
-- in the canonical tenant. This is correct for the current single-user
-- deployment. After Session C, membership is checked on every request
-- and this broad grant is narrowed by invitation logic.
INSERT INTO workspace_members (user_id, workspace_id, role, created_at)
SELECT u.id, w.id, 'owner', NOW()
  FROM users u
  CROSS JOIN workspaces w
 WHERE w.tenant_id = '00000000-0000-0000-0000-000000000001'
ON CONFLICT (user_id, workspace_id) DO NOTHING;

-- Verify: every workspace has at least one owner. A workspace with no
-- owner cannot be managed and cannot be deleted. Fail loudly.
DO $$
DECLARE
  orphan_count INT;
BEGIN
  SELECT COUNT(*) INTO orphan_count
    FROM workspaces w
   WHERE NOT EXISTS (
     SELECT 1 FROM workspace_members wm
      WHERE wm.workspace_id = w.id AND wm.role = 'owner'
   );
  IF orphan_count > 0 THEN
    RAISE EXCEPTION '079: % workspaces have no owner after backfill', orphan_count;
  END IF;
END $$;