-- ==========================================================================
-- 098_strict_workspace_and_tenant_isolation.sql
-- Garuda v1.9.9 Strict Multi-Tenant & Per-Workspace Isolation:
-- 1. Adds workspace_id and created_by_user_id to decisions so ADR & decision
--    lineage is strictly isolated per (tenant_id, workspace_id) and never
--    bleeds across workspaces or tenants.
-- 2. Repairs any historical non-bootstrap user accounts that were linked to
--    the canonical tenant (00000000-0000-0000-0000-000000000001) prior to 097.
-- 3. Purges auto-seeded demo document_claims from empty workspaces (workspaces
--    with 0 indexed repositories).
-- ==========================================================================

ALTER TABLE decisions ADD COLUMN IF NOT EXISTS workspace_id UUID;
ALTER TABLE decisions ADD COLUMN IF NOT EXISTS created_by_user_id UUID;
CREATE INDEX IF NOT EXISTS idx_decisions_tenant_workspace ON decisions (tenant_id, workspace_id);

-- Backfill existing decisions in canonical tenant to its 'default' workspace only
UPDATE decisions d
   SET workspace_id = w.id
  FROM workspaces w
 WHERE d.tenant_id = w.tenant_id
   AND w.name = 'default'
   AND d.workspace_id IS NULL;

-- Ensure every non-bootstrap tenant_owner / individual_dev user has their own isolated tenant
DO $$
DECLARE
    rec RECORD;
    new_tid UUID;
    new_wid UUID;
    inv_code TEXT;
    ws_name TEXT;
BEGIN
    FOR rec IN
        SELECT id, email, COALESCE(NULLIF(full_name, ''), email) AS fname,
               COALESCE(NULLIF(username, ''), SPLIT_PART(email, '@', 1)) AS uname,
               COALESCE(NULLIF(account_type, ''), 'tenant_owner') AS atype,
               primary_tenant_id
          FROM users
         WHERE LOWER(email) NOT IN ('admin@local', 'myshra777@gmail.com')
    LOOP
        IF rec.atype != 'team_member' AND (rec.primary_tenant_id IS NULL OR rec.primary_tenant_id = '00000000-0000-0000-0000-000000000001') THEN
            -- Check if user already owns a non-canonical tenant
            SELECT id INTO new_tid FROM tenants WHERE owner_user_id = rec.id AND id != '00000000-0000-0000-0000-000000000001' LIMIT 1;
            IF new_tid IS NULL THEN
                new_tid := gen_random_uuid();
                inv_code := 'GRD-' || UPPER(SUBSTRING(REPLACE(new_tid::text, '-', ''), 1, 6));
                INSERT INTO tenants (id, name, slug, invite_code, owner_user_id, created_at, updated_at)
                VALUES (new_tid, rec.fname || '''s Organization', LOWER(REGEXP_REPLACE(rec.uname, '[^a-zA-Z0-9]+', '-', 'g')), inv_code, rec.id, NOW(), NOW())
                ON CONFLICT (id) DO NOTHING;
            END IF;

            UPDATE users SET primary_tenant_id = new_tid WHERE id = rec.id;

            INSERT INTO tenant_members (user_id, tenant_id, role, created_at)
            VALUES (rec.id, new_tid, 'owner', NOW())
            ON CONFLICT (user_id, tenant_id) DO UPDATE SET role = 'owner';

            -- Remove erroneous membership in canonical tenant
            DELETE FROM workspace_members wm
             USING workspaces w
             WHERE wm.workspace_id = w.id
               AND wm.user_id = rec.id
               AND w.tenant_id = '00000000-0000-0000-0000-000000000001';

            DELETE FROM tenant_members
             WHERE user_id = rec.id
               AND tenant_id = '00000000-0000-0000-0000-000000000001';

            -- Ensure user has an initial workspace in their isolated tenant
            IF rec.atype = 'individual_dev' THEN
                ws_name := rec.uname || '-workflow';
            ELSE
                ws_name := 'default';
            END IF;

            SELECT id INTO new_wid FROM workspaces WHERE tenant_id = new_tid LIMIT 1;
            IF new_wid IS NULL THEN
                new_wid := gen_random_uuid();
                INSERT INTO workspaces (id, tenant_id, name, root_path, is_go_work, description, created_at, updated_at)
                VALUES (new_wid, new_tid, ws_name, '.', false, 'Isolated workspace', NOW(), NOW())
                ON CONFLICT (tenant_id, name) DO NOTHING;
                SELECT id INTO new_wid FROM workspaces WHERE tenant_id = new_tid AND name = ws_name LIMIT 1;
            END IF;

            IF new_wid IS NOT NULL THEN
                INSERT INTO workspace_members (user_id, workspace_id, role, created_at)
                VALUES (rec.id, new_wid, 'owner', NOW())
                ON CONFLICT (user_id, workspace_id) DO NOTHING;
            END IF;
        END IF;
    END LOOP;
END $$;

-- Purge auto-seeded demo document_claims from workspaces that have 0 indexed repositories
DELETE FROM document_claims dc
 WHERE NOT EXISTS (
     SELECT 1 FROM repositories r WHERE r.workspace_id = dc.workspace_id
 )
   AND (
     dc.document_path LIKE '%0001-payment-idempotency%'
     OR dc.document_path LIKE '%0002-merkle-integrity-layer%'
     OR dc.document_path LIKE '%0003-workspace-state-snapshots%'
     OR dc.document_path LIKE '%0004-zero-trust-agent-gate%'
   );

CREATE TABLE IF NOT EXISTS workspace_documents (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    workspace_id UUID NOT NULL,
    workspace_name TEXT NOT NULL DEFAULT 'default',
    title TEXT NOT NULL,
    doc_type TEXT NOT NULL DEFAULT 'ADR',
    path TEXT NOT NULL,
    content TEXT NOT NULL,
    sha256_hash TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'ENFORCED',
    created_by_user_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, workspace_id, path)
);
