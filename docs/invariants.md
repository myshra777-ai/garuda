# Garuda Invariants

These are the unbreakable rules of the Garuda truth substrate. Every feature, refactor, or bug fix must preserve these invariants.

Each invariant carries a verification status. `Enforced` means the rule holds today and the enforcing code is cited. `Partial` means the rule holds in some paths but not all, or holds by application logic without a database-level guard — the gap is named. `Planned` means the rule is intended but not yet built.

An invariant that is not yet enforced is still an invariant. It is a claim about the shape of the system, and the system is expected to converge on it. What it is not is a description of the current state. The status line is what makes the difference.

---

## 1. Revisions are append-only

`decision_revisions` rows are never updated or deleted by application code. Obsolete decisions are marked `superseded`, not removed.

**Status:** Partial. The revision write path in `internal/store/revision_store.go` inserts new rows only — no `UPDATE` or `DELETE` against `decision_revisions` is issued anywhere in the codebase. The table is created in `migrations/005_decision_revisions.sql` and refined in `046_align_decision_revisions_schema.sql`. There is no database-level trigger or rule that blocks an accidental `UPDATE` or `DELETE`. Adding one is tracked as a hardening item.

---

## 2. Content hash excludes metadata

`content_hash` = SHA-256 of the canonical JSON encoding of `DecisionContent`. Metadata — actor, timestamp, revision_id — is not part of the hash. The same content produces the same hash across submissions.

**Status:** Enforced. Computed in `internal/store/revision_store.go` before the row is inserted. The hash is verified on every read.

---

## 3. Actor comes from authentication

`actor` is always derived from the authenticated context. A client-provided actor is ignored, and may be stored separately as `requested_by`.

**Status:** Enforced. `internal/auth/jwt.go` extracts the actor from validated claims. Handlers pass it through to the store layer; the store never accepts an actor from the request body.

---

## 4. Tenant isolation is enforced

Every scoped query includes `tenant_id`. A workspace, decision, entity, or relationship belonging to one tenant is not visible to another. Cross-tenant reads return empty, not an error, so tenant existence is not leaked.

**Status:** Partial. Application-level enforcement is complete — every query in `internal/store/` and `internal/api/` filters by `tenant_id`. Database-level enforcement via Row Level Security is currently on `harvested_decisions` only (`migrations/023_harvested_decisions.sql`). Extending RLS to the remaining tenant-scoped tables is tracked as a hardening item.

---

## 5. Every mutation is atomic

Decision revision, Merkle update, and audit event are committed in one transaction. There are no partial states.

**Status:** Enforced. `SaveDecision` in `internal/store/postgres.go` opens a transaction, anchors the Merkle leaf, writes the decision row, and commits — or rolls back everything. The same pattern is used by the revision write path and by the policy evaluation write path.

---

## 6. The hash structure is verifiable

Every decision's inclusion in the ledger can be verified independently, without access to Garuda's database, servers, or keys.

Two schemes coexist.

**v1 — RFC 6962 Merkle tree with canonical byte encoding.** The current write path. Every leaf is produced by the versioned canonical encoder in `internal/merkle/decision.go`. Leaves are combined with domain separation: `leaf_hash = SHA256(0x00 || canonical_leaf_data)` and `internal_hash = SHA256(0x01 || left || right)`. The `0x00` and `0x01` prefixes prevent a second-preimage attack. Odd-node handling uses promotion, closing CVE-2012-2459. Introduced in:

- `migrations/067_merkle_verification_version.sql`
- `migrations/068_merkle_epoch_leaves.sql`
- `migrations/069_decisions_v1_merkle.sql`
- `migrations/070_policy_evaluations_v1.sql`

**v0 — linear hash chain.** `root_hash = SHA256(prev_root || decision_hash)`. The implementation in `internal/merkle/hash.go` is marked deprecated. It remains for historical rows only.

Every row in `merkle_roots`, `merkle_snapshots`, `decisions`, and `policy_evaluations` carries `verification_version`. The verifier selects the correct path automatically. `garuda policy verify` re-derives the inclusion proof from the stored data and the current epoch root.

**Status:** Enforced for v1 rows. v0 rows verify through the legacy path, preserved for backward compatibility.

---

## 7. Safe errors

Internal errors never leak to clients. Every client error includes a `request_id` for correlation.

**Status:** Partial. The handoff and resume paths in `internal/api/handoff_handlers.go` generate a request ID via `getRequestID(r)` and include it in every error response. The dashboard handlers use a `writeJSONError` helper that does not yet thread the request ID. Closing this gap is tracked as a follow-up.

---

## 8. Idempotency

Mutations support idempotency keys for safe retries. Same key produces the same result; no duplicate state is created.

**Status:** Enforced. The `idempotency_keys` table is defined in `migrations/058_idempotency_keys.sql` and queried before every mutating write. A repeat submission with the same key returns the previously computed result rather than creating a second row.

---

## 9. Observability

Every operation has a `request_id`. Logs, audit events, and errors share the same ID.

**Status:** Partial. Same coverage as invariant 7 — enforced in the handoff and resume paths, not yet threaded through the dashboard paths. When the gap in invariant 7 is closed, this one closes with it.

---

## 10. No external dependencies for core integrity

The truth substrate does not depend on LLMs, APIs, or external services. It is deterministic and self-contained.

**Status:** Enforced. The Merkle layer in `internal/merkle/` uses only the Go standard library. The canonical encoder has frozen golden vectors in `testdata/` and produces byte-identical output across implementations. No network calls are made during anchoring, verification, or proof derivation.

---

## What "Enforced" means here

A status of `Enforced` means three things:

1. The rule holds in the current code.
2. The enforcing code is cited, so a reader can check.
3. There is at least one test that would fail if the rule were violated, or a database constraint that makes the violation impossible. Where neither exists, the invariant is listed as `Partial` even if the code currently satisfies it.

A rule that holds by convention but not by test is a rule that will break on the next refactor. The distinction matters because the entire point of this document is that the invariants are permanent.