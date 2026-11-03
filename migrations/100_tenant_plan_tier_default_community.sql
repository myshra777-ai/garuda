-- 100: new tenants start on the community plan, not enterprise.
-- plan_tier was added in 097 with DEFAULT 'enterprise', which would have granted every
-- new account the top tier once plan entitlements exist. Nothing assigns or enforces a
-- plan today (there is no billing), so existing 'enterprise' values were only ever the
-- default, never a purchase; they are reset to 'community'.
-- Idempotent: safe to re-run.
DO $$
BEGIN
    IF to_regclass('public.tenants') IS NULL THEN
        RAISE NOTICE '100: tenants table missing, skipping';
    ELSE
        ALTER TABLE tenants ALTER COLUMN plan_tier SET DEFAULT 'community';
        UPDATE tenants SET plan_tier = 'community' WHERE plan_tier = 'enterprise';
    END IF;
END $$;
