-- ==========================================================================
-- 097_multi_tenant_rbac_isolation.sql
-- Garuda v1.9.9:
-- 1. Adds multi-tenant RBAC & onboarding columns to users and tenants:
--    - users: username, account_type ('tenant_owner' | 'team_member' | 'individual_dev'),
--             primary_tenant_id, job_title
--    - tenants: slug, org_domain, invite_code ('GRD-XXXXXX'), owner_user_id, plan_tier
-- 2. Cleans up any historical cross-tenant membership rows where non-owner
--    signups were erroneously linked to the canonical tenant (00000000-0000-0000-0000-000000000001).
-- ==========================================================================

ALTER TABLE users ADD COLUMN IF NOT EXISTS username TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS account_type TEXT NOT NULL DEFAULT 'tenant_owner';
ALTER TABLE users ADD COLUMN IF NOT EXISTS primary_tenant_id UUID;
ALTER TABLE users ADD COLUMN IF NOT EXISTS job_title TEXT NOT NULL DEFAULT '';

ALTER TABLE tenants ADD COLUMN IF NOT EXISTS slug TEXT NOT NULL DEFAULT '';
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS org_domain TEXT NOT NULL DEFAULT '';
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS invite_code TEXT NOT NULL DEFAULT '';
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS owner_user_id UUID;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS plan_tier TEXT NOT NULL DEFAULT 'enterprise';

-- Populate deterministic invite codes (e.g. GRD-000001 or GRD-8A3F1C) for all tenants
UPDATE tenants
   SET invite_code = CASE
           WHEN id = '00000000-0000-0000-0000-000000000001' THEN 'GRD-GARUDA1'
           ELSE 'GRD-' || UPPER(SUBSTRING(REPLACE(id::text, '-', ''), 1, 6))
       END
 WHERE COALESCE(invite_code, '') = '';

UPDATE tenants
   SET slug = LOWER(REGEXP_REPLACE(name, '[^a-zA-Z0-9]+', '-', 'g'))
 WHERE COALESCE(slug, '') = '';

-- Populate username from email prefix where empty
UPDATE users
   SET username = LOWER(REGEXP_REPLACE(SPLIT_PART(email, '@', 1), '[^a-zA-Z0-9_]+', '_', 'g'))
 WHERE COALESCE(username, '') = '';

-- Identify users who have their own dedicated personal/org tenant (created via /signup)
-- and remove the erroneous cross-tenant link to canonical tenant 00000000-0000-0000-0000-000000000001
DELETE FROM workspace_members wm
 USING workspaces w, users u
 WHERE wm.workspace_id = w.id
   AND wm.user_id = u.id
   AND w.tenant_id = '00000000-0000-0000-0000-000000000001'
   AND u.email NOT IN ('admin@local', 'myshra777@gmail.com')
   AND EXISTS (
       SELECT 1 FROM tenant_members tm2
        WHERE tm2.user_id = u.id
          AND tm2.tenant_id != '00000000-0000-0000-0000-000000000001'
   );

DELETE FROM tenant_members tm
 USING users u
 WHERE tm.user_id = u.id
   AND tm.tenant_id = '00000000-0000-0000-0000-000000000001'
   AND u.email NOT IN ('admin@local', 'myshra777@gmail.com')
   AND EXISTS (
       SELECT 1 FROM tenant_members tm2
        WHERE tm2.user_id = u.id
          AND tm2.tenant_id != '00000000-0000-0000-0000-000000000001'
   );

-- Set primary_tenant_id for canonical owner(s)
UPDATE users
   SET primary_tenant_id = '00000000-0000-0000-0000-000000000001',
       account_type = 'tenant_owner',
       role = 'owner'
 WHERE email IN ('myshra777@gmail.com', 'admin@local');

UPDATE tenants
   SET name = 'Garuda Headquarters (Primary Tenant)',
       slug = 'garuda-hq',
       invite_code = 'GRD-GARUDA1'
 WHERE id = '00000000-0000-0000-0000-000000000001'
   AND (name = 'default' OR name LIKE 'tenant-%');

-- Backfill primary_tenant_id for all other users from their non-canonical tenant membership first
UPDATE users u
   SET primary_tenant_id = sub.tenant_id
  FROM (
       SELECT DISTINCT ON (user_id) user_id, tenant_id
         FROM tenant_members
        ORDER BY user_id, (tenant_id != '00000000-0000-0000-0000-000000000001') DESC, created_at DESC
  ) sub
 WHERE u.id = sub.user_id
   AND u.primary_tenant_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_tenants_invite_code ON tenants (invite_code);
CREATE INDEX IF NOT EXISTS idx_users_primary_tenant ON users (primary_tenant_id);
