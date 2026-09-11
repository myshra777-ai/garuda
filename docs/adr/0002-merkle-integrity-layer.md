# ADR-0002: Merkle Integrity Layer — Canonical Design

**Status:** Accepted
**Date:** 2026-09-11
**Supersedes:** None (first Merkle design lock)
**Deciders:** Rohit Mishra

## Context

The Garuda trust narrative depends on cryptographic auditability. An auditor,
investor, or security reviewer must be able to independently verify that:

1. A given decision existed at a given block height in a given tenant's log.
2. No decision has been altered or removed after inclusion.
3. The root hash published at any block height commits to all decisions
   before it.

The current implementation (v0) does not satisfy (1) because it uses a linear
hash chain rather than a Merkle tree, provides no inclusion proof, and
contains a genesis that is not tenant-scoped or versioned. Additionally, the
`HashDecision` function relies on JSON map serialization whose ordering is an
implementation detail of `encoding/json`, not a specification.

This ADR defines the v1 integrity layer. It is the authoritative reference
for hashing, tree construction, inclusion proofs, and verification. No code
in `internal/merkle/` may deviate from it.

## Decisions

### D1. Canonical serialization is versioned and layout-fixed.

Every hash input is produced by a canonical encoder that emits a
length-prefixed byte sequence:

GenesisV1Content = "GARUDA_CHAIN_GENESIS_V1"
GenesisV1Date = "2026-09-11T00:00:00Z"

text

The per-tenant genesis root is:
genesisRoot(tenantID) = SHA256(version_byte || genesisContent || genesisDate || tenantID)

text

Two tenants have different genesis roots. The content and date are literal
strings in source code. An auditor can recompute any tenant's genesis root
from this ADR alone.

### D3. The tree is a binary Merkle tree with documented odd-leaf handling.

Leaves are hashed in a canonical order (defined per-tree — see D5).
Internal nodes are `SHA256(left_hash || right_hash)` where `left_hash` and
`right_hash` are 32 raw bytes (not hex). When a level has an odd number of
nodes, the last node is duplicated (`H(leaf || leaf)`).

The tree is rebuilt deterministically from the ordered leaf set. No
incremental state is persisted — the leaf set is the source of truth.

### D4. Inclusion proofs are O(log n) and self-contained.

A proof is a sequence of `ProofNode`:
type ProofNode struct {
Position string // "L" or "R" — where the sibling sits relative to the current hash
Hash []byte // 32 bytes
}

text

Verification is a pure function:
func VerifyInclusion(leaf, proof []ProofNode, root []byte) bool

text

It takes exactly three inputs, accesses no storage, calls no other Garuda
code, and performs `O(log n)` hashing operations. A proof for a tree of
100,000 leaves is approximately 17 nodes ≈ 544 bytes.

### D5. Two tiers, one root.

Garuda maintains two independent Merkle trees per tenant, per block:

- **Static tree** — leaves are canonical decision hashes (D6). Ordered by
  insertion sequence within the block.
- **Runtime tree** — leaves are runtime verification hashes (existing
  `RuntimeLeaf.ComputeLeafHash`). Ordered lexicographically by hash.

The **epoch root** for a block is:
epochRoot = SHA256(version_byte || static_root || runtime_root || parent_epoch_root || block_height_u64_be)

text

`parent_epoch_root` is the epoch root of the previous block, or the tenant's
genesis root for block 1. This is the only place where epochs are chained.
Within an epoch, the trees are true Merkle trees.

### D6. Decision hashing is canonical and evidence-aware.
canonicalDecision(decisionID, title, status, scopeDomain, scopeSystem, owner, evidenceIDs) =
version_byte ||
u32(decisionID_raw_16_bytes) ||
u32(len(title)) || title_bytes ||
u32(len(status)) || status_bytes ||
u32(len(scopeDomain)) || scopeDomain_bytes ||
u32(len(scopeSystem)) || scopeSystem_bytes ||
u32(len(owner)) || owner_bytes ||
u32(len(evidenceIDs)) ||
for each id in evidenceIDs sorted lexicographically:
u32(len(id)) || id_bytes

text

Evidence IDs are sorted before encoding. Duplicate evidence IDs are
preserved (they carry information). The result is hashed with SHA-256.

### D7. Roots are stored as BYTEA; hex is a presentation concern only.

The database stores raw 32-byte hashes in `BYTEA` columns. Hex encoding
happens at the API boundary. Check constraints enforce `octet_length = 32`
on every hash column. This eliminates the current `TEXT` vs `BYTEA` class of
bugs.

### D8. Verification never re-derives from the same code path.

`VerifyInclusion` is the only verification path for tree membership. It
takes `(leaf, proof, root)` and nothing else. It does not call the code that
produced the leaf, the proof, or the root. It is tested against golden
vectors (see D10).

`VerifyEvaluation` at the API level calls `VerifyInclusion`. It does not call
`HashDecision` or `ChainHash` to check consistency — it only confirms that
the presented proof is valid against the presented root.

### D9. External anchoring is part of the design, not an add-on.

Every N blocks (configurable, default 100), Garuda emits a signed snapshot:
{
"tenant_id": "...",
"block_height": 100,
"epoch_root_hex": "...",
"previous_anchor_hash": "...",
"signed_at": "RFC3339",
"signature": "<ed25519 sig over canonical serialization>"
}

text

The snapshot is written to a configurable sink (local directory in v1;
S3/git/public timestamping in v2). External anchoring is what makes the
chain independent of Garuda's own storage.

### D10. Test vectors are committed and frozen.

`internal/merkle/testdata/vectors_v1.json` contains:

- Known leaf inputs and their expected hashes
- Known leaf sets, expected roots, and inclusion proofs for each leaf
- Tamper cases (modified leaf, modified proof, modified root) that MUST fail
- Genesis root for at least three different tenant IDs

These vectors are generated once and committed. The test suite asserts that
the v1 implementation reproduces them byte-for-byte. If a code change alters
any hash, the test fails.

## Migration from v0

The existing `merkle_roots` and `merkle_snapshots` tables contain data
produced under v0 rules (JSON-serialized decisions, linear chain). Two
options:

**Option 1 (selected):** Declare v0 historical. Preserve the data in place.
Existing blocks remain queryable under `verification_version = 0`. New
blocks are written with `verification_version = 1`. The `garuda verify`
command reports which version a proof used. No v0 data is re-hashed.

**Option 2 (rejected):** Re-hash all historical data under v1. Rejected
because it destroys the historical integrity property — a v0 root was true
under v0 rules, and re-signing it under v1 rules fabricates a new claim
about old data.

The migration is:

```sql
ALTER TABLE merkle_roots ADD COLUMN verification_version SMALLINT NOT NULL DEFAULT 0;
ALTER TABLE merkle_roots ALTER COLUMN root_hash TYPE BYTEA USING decode(root_hash, 'hex');
-- ... with a DEFAULT 1 for new rows
Existing rows are labeled v0. New rows default to v1. Both have check
constraints.

Consequences
Positive: Independent verifiability. Standard Merkle construction. No
JSON dependency. Proofs are O(log n). Roots are typographically guaranteed
to be 32 bytes. Historical data is preserved but distinguished.

Negative: The v0 verification code path must be preserved for
historical proofs — it cannot be deleted. This adds ~150 lines of dead
code that must not be removed until the last v0 block is 10 years old.

Negative: Every existing root hash must be migrated from TEXT to
BYTEA. This is a one-way migration and requires a database backup.

Non-negotiable: No code outside internal/merkle/ may compute a hash
for storage. All hashing goes through the canonical encoders defined in
this ADR.

References
RFC 6962 (Certificate Transparency) — Merkle tree construction

RFC 8032 (Ed25519) — signature scheme for external anchoring

Bitcoin's genesis block — example of a verifiable, dated genesis

text

---

OOdd-node handling: when a level has an odd number of nodes, the final
node is promoted unchanged to the next level. It is NOT duplicated.

Rationale: RFC 6962 uses split-at-largest-power-of-two, which avoids
the odd-node case entirely. Bitcoin's historical implementation used
duplication H(node || node) but was vulnerable to CVE-2012-2459, where
[A, B, C] and [A, B, C, C] produced identical Merkle roots. Garuda
chooses promotion: the odd node passes through to the next level without
a fabricated sibling. This produces a tree shape that depends on the
exact leaf count, so [A, B, C] and [A, B, C, C] cannot collide. The
leaf and internal domain-separation prefixes (first paragraph of D3)
prevent leaf-vs-internal confusion.