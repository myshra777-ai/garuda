-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.
--
-- 091: Remove workspaces whose tenant no longer exists.
--
-- Migration 090 removed 21 test tenants but left their workspaces
-- behind. A workspace is a child of a tenant; deleting the parent
-- should delete the children. It did not.
--
-- Effect: 115 workspaces total, 6 belong to live tenants. Every
-- "count all workspaces" metric is inflated by ~109x. The Business
-- tab is honest about the number but the number is misleading.
--
-- This migration deletes workspaces whose tenant_id does not resolve
-- to a row in tenants. It is idempotent: a second run finds zero
-- orphans and deletes zero rows.
--
-- Every child table that references workspace_id must already have
-- had its rows cleaned up (verified before applying). If a foreign
-- key blocks the delete, the migration will fail loudly and no
-- partial delete occurs.

BEGIN;

-- Guard: fail if any orphan workspace still has child rows. This
-- prevents a partial delete that would leave dangling references.
DO $$
DECLARE
    orphan_count int;
    child_count bigint;
BEGIN
    SELECT COUNT(*) INTO orphan_count
      FROM workspaces w
     WHERE NOT EXISTS (SELECT 1 FROM tenants t WHERE t.id = w.tenant_id);

    IF orphan_count = 0 THEN
        RAISE NOTICE '091: no orphan workspaces; nothing to do';
        RETURN;
    END IF;

    SELECT
      (SELECT COUNT(*) FROM repositories r
         WHERE r.workspace_id IN (
           SELECT id FROM workspaces w
            WHERE NOT EXISTS (SELECT 1 FROM tenants t WHERE t.id = w.tenant_id)))
    + (SELECT COUNT(*) FROM entities e
         WHERE e.workspace_id IN (
           SELECT id FROM workspaces w
            WHERE NOT EXISTS (SELECT 1 FROM tenants t WHERE t.id = w.tenant_id)))
    + (SELECT COUNT(*) FROM mcp_sessions s
         WHERE s.workspace_id IN (
           SELECT id FROM workspaces w
            WHERE NOT EXISTS (SELECT 1 FROM tenants t WHERE t.id = w.tenant_id)))
    + (SELECT COUNT(*) FROM policy_evaluations p
         WHERE p.workspace_id IN (
           SELECT id FROM workspaces w
            WHERE NOT EXISTS (SELECT 1 FROM tenants t WHERE t.id = w.tenant_id)))
    INTO child_count;

    IF child_count > 0 THEN
        RAISE EXCEPTION '091: % orphan workspaces still have % child rows; clean up children first',
            orphan_count, child_count;
    END IF;

    RAISE NOTICE '091: % orphan workspaces, zero child rows — safe to delete', orphan_count;
END $$;

DELETE FROM workspaces w
 WHERE NOT EXISTS (SELECT 1 FROM tenants t WHERE t.id = w.tenant_id);

COMMIT;