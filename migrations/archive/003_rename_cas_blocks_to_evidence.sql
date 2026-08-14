ALTER TABLE cas_blocks RENAME TO evidence_store;
ALTER INDEX idx_cas_created RENAME TO idx_evidence_created;