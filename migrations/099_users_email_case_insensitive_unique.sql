-- 099: enforce case-insensitive uniqueness of user emails.
-- users.email is UNIQUE (case-sensitive), so "Rohit@x.com" and "rohit@x.com" could both
-- sign up as separate accounts. This adds a unique index on lower(email).
-- Safe on dirty data: if case-insensitive duplicates already exist, the index is skipped
-- with a NOTICE (resolve the duplicates, then re-run this file; it is idempotent).
DO $$
BEGIN
    IF to_regclass('public.users') IS NULL THEN
        RAISE NOTICE '099: users table missing, skipping';
    ELSIF EXISTS (SELECT 1 FROM users GROUP BY lower(email) HAVING count(*) > 1) THEN
        RAISE NOTICE '099: skipped users_email_lower_key - duplicate emails differing only by case exist; merge them and re-run';
    ELSE
        CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_key ON users (lower(email));
    END IF;
END $$;
