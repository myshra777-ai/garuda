-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- 090_delete_test_residue.sql
--
-- Removes tenants, users, and their cascaded rows that were created
-- by TestSignupUser_CreatesFullScope and
-- TestSignupUser_RejectsDuplicateEmail between 2026-09-13 and
-- 2026-09-19. Before this migration the cleanup hook deleted the
-- user but not the tenant, so every test run left one tenant and
-- one workspace behind. Sixty runs over six days is the count we
-- see in the tenants table.
--
-- The hook is fixed in the same commit. This migration removes the
-- residue the old hook could not.
--
-- Idempotent: re-applying matches nothing on a clean database.

BEGIN;

-- Test tenants. The naming convention is baked into the test source:
--   signup-test-<hex>@local's tenant
--   signup-dup-<hex>@local's tenant
--   verify-<n>@local's tenant
-- Canonical tenant is excluded (name = 'canonical').
-- Real user tenants end in @gmail.com or another domain, not @local.
DELETE FROM tenants
 WHERE name LIKE '%@local''s tenant';

-- Test users that signed up for the tenants above. The tenant delete
-- above cascades tenant_members. The users rows have no FK to the
-- tenant and would survive.
DELETE FROM users
 WHERE email LIKE '%@local'
   AND role = 'user'
   AND email NOT IN ('admin@local');

COMMIT;
