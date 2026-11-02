# ADR-0003: Workspace State Snapshots — Design Stub

**Status:** Proposed
**Date:** 2026-09-11
**Supersedes:** None
**Depends on:** ADR-0002 (Merkle Integrity Layer)

## Context

Garuda has two independent Merkle constructs:

1. **The decision log.** One leaf per decision or policy evaluation.
   Lives in `merkle_roots` and `merkle_epoch_leaves`. Every row is
   anchored via `store.AppendLeafAndSealTx`. Described by ADR-0002.

2. **The workspace state snapshot.** A periodic commitment to the set
   of entities and claim verifications that existed in a workspace at
   a given moment. Lives in `merkle_snapshots`. Produced by the worker
   loop every N minutes and by `CreateUnifiedMerkleSnapshot`.

These serve different purposes:

- The **decision log** answers "was this specific decision anchored at
  block height H?" It supports O(log n) inclusion proofs for individual
  decisions. Its audience is a verifier who wants to prove a change was
  governed.

- The **workspace snapshot** answers "what did the workspace look like
  at time T?" It supports a single hash that summarises the entire
  state. Its audience is a compliance reviewer who wants to freeze a
  point in time.

ADR-0002 specifies the v1 model for the decision log. It does not
specify a v1 model for the workspace snapshot. Migration 069 added
format CHECK constraints to `merkle_snapshots` that the current worker
code violated for empty tenants; the fix (Commit 5b.3b.3) uses the
defined empty-tree hash from `tree.go` and labels all worker-produced
rows `verification_version = 0` to distinguish them from the v1
decision log.

## Problem Statement

The current workspace snapshot has three shortcomings that this ADR
must resolve:

1. **The static tier is not a Merkle tree.** It concatenates entity
   IDs with a pipe and hashes the result. That is a single leaf, not a
   tree. It has no inclusion proof.

2. **The two tiers have no defined leaves.** The runtime tier uses
   `RuntimeLeaf` values from `claim_verifications`. The static tier
   uses raw entity IDs. Neither has a canonical byte encoding.

3. **There is no verification path.** No code can take a claimed
   entity or claim verification and prove that it appears in a
   specific snapshot. The snapshot hash is a commitment, not a
   queryable proof.

## Open Questions

Before this ADR can move from Proposed to Accepted, the following
must be decided:

### Q1. What is the leaf for an entity?

Candidate answers:
- The entity's canonical ID (UUID).
- A canonical encoding of the entity's full content (name, kind,
  package, fields, methods, hashes).
- The `(repository_id, canonical_name)` pair.

The choice determines what a proof demonstrates. UUID-only proves
membership; full-content proves integrity of the entity's current
form.

### Q2. What is the leaf for a claim verification?

Same question, same tradeoff. `RuntimeLeaf.ComputeLeafHash` already
exists and produces a deterministic value, but it depends on a
specific field set that may evolve.

### Q3. Does the workspace snapshot need to be a two-tier tree?

The current design has a static tier and a runtime tier, combined by
`ComputeUnifiedEpochRoot`. That mirrors the decision log's structure,
but the decision log's tiers correspond to "things a human decided"
versus "things observed running." For a workspace snapshot, the
analogous split might be "structure that exists" versus "behaviour
that occurred." Is that the right axis?

### Q4. How is the snapshot consumed?

Three candidate consumers:
- Compliance export: one hash that says "this state existed."
- IDE query: "is this entity present in the snapshot at time T?"
- Drift detection: "what changed between snapshot A and B?"

The third consumer is the most demanding, because it requires
diffing two trees. That may push the design toward a
persistent-merkle structure.

### Q5. Storage location.

Options:
- Continue storing snapshots in `merkle_snapshots` with v1 labels.
- A new `merkle_snapshot_leaves` table parallel to
  `merkle_epoch_leaves`.
- Reuse `merkle_epoch_leaves` with a distinct tier namespace.

### Q6. Re-anchoring.

Should existing v0 snapshots be re-anchored under the v1 model? If
yes, migration cost and audit trail concerns apply. If no, we carry
two models indefinitely.

## Out of Scope

- The decision log's v1 model. That is ADR-0002 and is complete.
- Business-state integrity. Tracked separately.

## Decision

Pending. This document records the design work required. No
implementation follows until Q1–Q6 are resolved and this ADR is
marked Accepted.

## Consequences Of The Interim State

Until this ADR is resolved:

- Worker-produced snapshots carry `verification_version = 0`.
- They pass migration 069's format CHECK constraints because empty
  tiers now produce the defined empty-tree hash from `tree.go`.
- They are not included in any v1 proof path.
- The `CreateUnifiedMerkleSnapshot` and `snapshotTenant` functions
  are marked Deprecated.
- New work on the decision log does not touch the worker.