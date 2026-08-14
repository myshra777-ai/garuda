# Garuda Invariants

These are the unbreakable rules of the Garuda truth substrate. Every feature, refactor, or bug fix must preserve these invariants.

## 1. Revisions are append-only
- `decision_revisions` rows are never updated or deleted.
- Obsolete decisions are marked `superseded`, not removed.
- `ON DELETE CASCADE` is prohibited.

## 2. Content hash excludes metadata
- `content_hash` = SHA-256(canonical JSON of `DecisionContent`).
- Metadata (actor, timestamp, revision_id) are not part of the hash.
- Same content → same hash across submissions.

## 3. Actor comes from authentication
- `actor` is always derived from the authenticated context.
- Client-provided actor is ignored (may be stored as `requested_by`).

## 4. Tenant isolation is enforced
- All queries include `tenant_id`.
- RLS (Row Level Security) is enabled.
- Cross-tenant access is impossible at the SQL level.

## 5. Every mutation is atomic
- Decision revision, Merkle update, and audit event are committed in one transaction.
- No partial states.

## 6. Hash chain is verifiable
- `root_hash` = SHA-256(prev_root || decision_hash).
- `garuda verify` can validate the entire chain.

## 7. Safe errors
- Internal errors never leak to clients.
- Every client error includes a `request_id` for correlation.

## 8. Idempotency
- Mutations support idempotency keys for safe retries.
- Same key → same result; no duplicate state.

## 9. Observability
- Every operation has a `request_id`.
- Logs, audit events, and errors share the same ID.

## 10. No external dependencies for core integrity
- The truth substrate does not depend on LLMs, APIs, or external services.
- It is deterministic and self-contained.