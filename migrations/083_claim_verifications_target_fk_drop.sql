-- migrations/083_claim_verifications_target_fk_drop.sql
--
-- The target_entity_id column may hold either a real entity UUID
-- (when the verifier resolved the target) or a hash-based UUID
-- derived from the raw target string (when it did not). The
-- verifier's own CONTRADICTED query at internal/runtime/verifier.go
-- writes COALESCE(target_entity_id, md5(raw_target)::uuid) — a
-- hash, not an entity.
--
-- The FK to entities(id) is incompatible with that design. It has
-- never fired because no CONTRADICTED row has ever landed (verified
-- 2026-09-15: 0 rows), but it would fire the moment one did.
--
-- Dropping the FK makes the schema match the code. The
-- evidence_payload->>'raw_target' column remains the human-readable
-- identifier of the unresolved target; target_entity_id is a stable
-- lookup key.
--
-- The FK on source_entity_id is unaffected: the source is always a
-- real entity, resolved by the correlator.

BEGIN;

ALTER TABLE claim_verifications
  DROP CONSTRAINT IF EXISTS claim_verifications_target_entity_id_fkey;

COMMIT;