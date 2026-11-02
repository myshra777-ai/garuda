DO $$
BEGIN
  IF to_regclass('evidence_store') IS NULL AND to_regclass('cas_blocks') IS NOT NULL THEN
    ALTER TABLE cas_blocks RENAME TO evidence_store;
  END IF;
  IF to_regclass('idx_evidence_created') IS NULL AND to_regclass('idx_cas_created') IS NOT NULL THEN
    ALTER INDEX idx_cas_created RENAME TO idx_evidence_created;
  END IF;
END $$;
