-- Copyright 2026 Rohit Mishra
-- SPDX-License-Identifier: Apache-2.0
--
-- 074_cross_repo_edges_unresolved_unique.sql
--
-- Adds the partial unique index required by the ON CONFLICT clause
-- used in the unresolved branch of detectCrossRepoImports. Without
-- this index, that branch's ON CONFLICT would also throw SQLSTATE
-- 42P10.
--
-- Note: to_entity_id is NULL for these rows, and PostgreSQL treats
-- NULL as distinct in a unique index by default. The index therefore
-- omits to_entity_id from its key columns and includes it only as a
-- predicate filter (WHERE to_entity_id IS NULL). Two unresolved edges
-- with identical (tenant, from_repo, to_repo, from_entity, rel_type)
-- then collide as intended.

CREATE UNIQUE INDEX IF NOT EXISTS idx_cross_repo_edges_unique_unresolved
  ON cross_repo_edges (
    tenant_id, from_repo_id, to_repo_id, from_entity_id, relationship_type
  )
  WHERE to_entity_id IS NULL;