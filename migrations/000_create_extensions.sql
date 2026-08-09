-- Ensure cryptographic UUID generator functions are available
-- Creates `gen_random_uuid()` used by later migrations
CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;
