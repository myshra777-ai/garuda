-- ==========================================================================
-- 099_reconcile_developer_workspaces.sql
-- Garuda v1.9.9 CLI & Multi-Tenant Workspace Synchronization:
-- 1. Reassigns orphan workspaces created via CLI under the canonical bootstrap
--    tenant (00000000-0000-0000-0000-000000000001) such as 'formance-ledger'
--    into the active developer/owner tenant so they are instantly visible in
--    the web dashboard.
-- 2. Aligns child tables (entities, claims, repositories, decisions) to match
--    the workspace's active tenant.
-- 3. Grants 'owner' role in workspace_members to the developer user for all
--    their workspaces.
-- ==========================================================================

DO $$
DECLARE
    owner_rec RECORD;
BEGIN
    -- Discover the primary developer/owner on this instance
    SELECT id, primary_tenant_id INTO owner_rec
      FROM users
     WHERE primary_tenant_id IS NOT NULL AND primary_tenant_id != '00000000-0000-0000-0000-000000000001'
     ORDER BY (role = 'owner' OR account_type IN ('tenant_owner', 'individual_dev')) DESC,
              (email NOT IN ('admin@local', 'myshra777@gmail.com')) DESC,
              updated_at DESC, created_at DESC
     LIMIT 1;

    IF owner_rec.primary_tenant_id IS NOT NULL THEN
        -- Adopt orphan workspaces created via CLI under canonical tenant
        UPDATE workspaces
           SET tenant_id = owner_rec.primary_tenant_id
         WHERE tenant_id = '00000000-0000-0000-0000-000000000001'
           AND name NOT IN ('default', 'audit-sandbox');

        -- Align child tables
        UPDATE entities e
           SET tenant_id = w.tenant_id
          FROM workspaces w
         WHERE e.workspace_id = w.id AND e.tenant_id != w.tenant_id;

        UPDATE claims c
           SET tenant_id = w.tenant_id
          FROM workspaces w
         WHERE c.workspace_id = w.id AND c.tenant_id != w.tenant_id;

        UPDATE repositories r
           SET tenant_id = w.tenant_id
          FROM workspaces w
         WHERE r.workspace_id = w.id AND r.tenant_id != w.tenant_id;

        UPDATE decisions d
           SET tenant_id = w.tenant_id
          FROM workspaces w
         WHERE d.workspace_id = w.id AND d.tenant_id != w.tenant_id;

        -- Ensure developer has owner access in workspace_members for all workspaces in their tenant
        INSERT INTO workspace_members (user_id, workspace_id, role, created_at)
        SELECT owner_rec.id, w.id, 'owner', NOW()
          FROM workspaces w
         WHERE w.tenant_id = owner_rec.primary_tenant_id
        ON CONFLICT (user_id, workspace_id) DO UPDATE SET role = 'owner';
    END IF;
END $$;
