<p align="center">
  <img src="assets/garuda-logo.png" width="128" alt="Garuda">
</p>

<h1 align="center">Garuda</h1>

<p align="center">
  <strong>Analyze | Verify | Govern</strong>
</p>

<p align="center">
  <em>Evidence-backed software intelligence with a cryptographic trust layer.</em>
</p>

<p align="center">
  <a href="https://github.com/myshra777-ai/garuda/releases"><img src="https://img.shields.io/badge/version-v0.1.x-orange.svg?style=flat-square" alt="Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache_2.0-green.svg?style=flat-square" alt="License"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.25-00ADD8.svg?style=flat-square" alt="Go Version"></a>
  <a href="internal/merkle/"><img src="https://img.shields.io/badge/trust_layer-RFC_6962-blueviolet.svg?style=flat-square" alt="Merkle"></a>
  <a href="docs/adr/0002-merkle-integrity-layer.md"><img src="https://img.shields.io/badge/ADR--0002-complete-brightgreen.svg?style=flat-square" alt="ADR-0002"></a>
</p>

<p align="center">
  <sub>
    <strong>Status:</strong> alpha. The Merkle trust layer (ADR-0002) is complete for the decision log.
    Multi-language analysis is partial. The MCP server is under rework.
    See <a href="#current-state">Current State</a> for the honest breakdown.
  </sub>
</p>

---

## What Garuda Is

Garuda builds a structured, verifiable semantic model of a software system and anchors every governance decision to a cryptographic log that a third party can independently verify.

```mermaid
flowchart LR
    SOURCE["Source Code"] --> ANALYZE["Analyze"]
    RUNTIME["Runtime"] --> VERIFY["Verify"]
    ANALYZE --> SEMANTIC["Semantic Model"]
    SEMANTIC --> CLAIMS["Claims and Evidence"]
    CLAIMS --> VERIFY
    VERIFY --> POLICY["Policy Engine"]
    POLICY --> MERKLE["Merkle Trust Layer"]
    MERKLE --> AUDIT["Independent Auditors"]
    MERKLE --> AGENTS["AI Agents"]
    MERKLE --> CI["CI and Governance"]
```

Three capabilities, in the order they are real:

1. **Analyze.** Extract entities and relationships from source code. Every edge carries a real resolution classification — `GO_TYPES`, `IMPORT_RESOLUTION`, `AST_EXACT`, or `HEURISTIC` — and evidence pointing to the source line that justifies it.
2. **Verify.** Anchor decisions and policy evaluations in a versioned, tenant-scoped Merkle log. Every leaf is produced by a canonical byte encoder with frozen golden vectors, so independent verifiers in any language can reproduce the hash. Inclusion proofs are pure functions of `leaf`, `proof`, and `root`.
3. **Govern.** Evaluate declarative YAML policies against the workspace. Every ALLOW, WARN, REVIEW, or BLOCK decision is anchored to the same Merkle ledger as the evidence it cited. An agent or auditor can independently confirm the decision happened at a specific block height.

Garuda is not:

- another text search engine
- another generic dependency graph
- another documentation generator
- an LLM that guesses how a repository works

---

## Current State

This table is authoritative. When a feature's status is unclear, trust this table over any other part of the repository.

| Capability | Status | Evidence |
| :--- | :--- | :--- |
| Go semantic analysis | **Stable** | 20/20 fixtures at 100% precision, 100% recall. All six edge types with real resolution classification. |
| Python semantic analysis | **Beta** | Structural scanner, not type-checked. Entities and imports only. No calls or implements. |
| TypeScript semantic analysis | **Beta** | tree-sitter parser. Has never written entities to a workspace in a validated run. Requires CGO. |
| Canonical byte encoder | **Stable** | `internal/merkle/canonical.go`. Frozen golden vectors in `testdata/vectors_v1.json`. |
| RFC 6962 Merkle tree | **Stable** | `internal/merkle/tree.go`. Domain separation, CVE-2012-2459 defense. Golden vectors frozen. |
| Decision write path | **Stable** | `SaveDecision` writes a canonical leaf and a self-contained JSON proof. Integration-tested. |
| Policy evaluation write path | **Stable** | `Anchor.AppendEvaluation` writes a canonical leaf. Integration-tested. |
| Workspace snapshot | **Deprecated** | Produces valid v0 snapshots. A v1 design is tracked in ADR-0003. |
| Policy engine | **Stable** | YAML DSL, four-outcome decisions, Merkle-anchored. Every decision exposes its evidence chain. |
| Evidence ledger | **Stable** | Every relationship emitted by the Go analyzer carries `File` and `LineStart`. Enforced by `TestEvidencePopulation_NoEmptyEdges`. |
| Multi-repository intelligence | **Alpha** | Workspace sync works for same-language repos. Cross-language resolution is unproven. |
| Documentation ingestion | **Alpha** | Parser exists. `document_claims` table has rows. Zero claims have been verified against code. |
| Runtime verification | **Not yet active** | Ingestion endpoint exists. `runtime_observations` has zero rows. |
| MCP server | **Under rework** | Current implementation reads one request from stdin and exits. Not a long-lived MCP server. |
| IDE extension | **Unvalidated** | Extension source exists. Not tested against a real workspace. |
| Multi-language support | **Overstated** | Only Go entities are present in any validated workspace. |

### On the multi-language claim

The analyzers for Python and TypeScript exist. They compile and produce output on synthetic inputs. But no validated workspace contains Python or TypeScript entities. The accurate reading is:

> Garuda's language pipeline is designed to be language-agnostic. Go is stable. Python and TypeScript analyzers exist but have not been validated against a real cross-language workspace.

### On precision numbers

The 100/100/100/100 aggregate is from the 20-fixture synthetic corpus in `garuda-bench/`. Real-repository precision on gin, cobra, testify, and chi has not been measured. Any claim of high precision on real repositories without a named corpus, a stated methodology, and a reproducible result is unsupported.

---

## The Trust Layer

The trust layer is `internal/merkle/`. Its design is specified in full in [ADR-0002](docs/adr/0002-merkle-integrity-layer.md).

```mermaid
flowchart TD
    DECISION["Policy Decision"] --> CANON["Canonical Encoder"]
    CANON --> LEAF["Leaf Hash SHA256 0x00"]
    LEAF --> TREE["RFC 6962 Tree"]
    TREE --> ROOT["Epoch Root"]
    ROOT --> PROOF["Inclusion Proof"]
    PROOF --> VERIFY["VerifyInclusion"]
    VERIFY --> AUDITOR["Independent Auditor"]
```

### Canonical encoding

Every hash input is produced by a versioned byte encoder in [`canonical.go`](internal/merkle/canonical.go) that emits a length-prefixed, field-ID-ordered sequence. There is no dependency on `encoding/json`, map iteration order, or any other language-native serializer. Two independent implementations in different languages will produce byte-identical output for the same logical input.

Golden vectors for the encoder are frozen in [`testdata/vectors_v1.json`](internal/merkle/testdata/vectors_v1.json). A test asserts byte-for-byte reproduction. Any change to the encoding requires a version byte bump and a new ADR.

### Merkle tree

The tree is RFC 6962-compliant with explicit domain separation:

```text
leaf_hash     = SHA256(0x00 || leaf_data)
internal_hash = SHA256(0x01 || left || right)
```

The `0x00` and `0x01` prefixes prevent a second-preimage attack where a leaf's bytes could equal an internal node's bytes. Odd-node handling uses promotion, which closes CVE-2012-2459: `[A, B, C]` and `[A, B, C, C]` produce different roots.

### Verifiable proofs

An inclusion proof is a list of `Position` and `Hash` nodes. Verification is a pure function:

```go
func VerifyInclusion(leafData []byte, proof []ProofNode, root []byte) bool
```

It takes three inputs, touches no storage, calls no Garuda code, and performs `O(log n)` hashing operations. For a tree of 100,000 leaves, a proof is approximately 17 nodes.

### Versioning

Two Merkle schemes coexist:

| Version | Scheme | Verified by |
| :--- | :--- | :--- |
| 0 | linear hash chain `SHA256(parent || child)` | deprecated path, kept for historical rows |
| 1 | RFC 6962 tree with canonical leaves | `VerifyInclusion`, pure function |

Every row in `merkle_roots`, `merkle_snapshots`, `decisions`, and `policy_evaluations` carries `verification_version`. Verification selects the path automatically.

### What this means in practice

When a policy decision is anchored via the v1 path, the response includes:

- The leaf hash computed by the canonical encoder.
- The tier roots and parent epoch root.
- The epoch height at which the leaf was committed.
- An inclusion proof path.

A third party with the source of `internal/merkle/tree.go` and `canonical.go` — but no access to Garuda's infrastructure — can recompute the leaf hash, verify the inclusion proof, and confirm the epoch root. No trust in Garuda's database or processes is required.

---

## The Analyzer

The Go analyzer produces a typed semantic snapshot from a Go module or workspace. It is the only analyzer with a complete and measured implementation.

### Edge types

| Edge | Source signal | Resolution |
| :--- | :--- | :--- |
| IMPORTS | `*ast.ImportSpec` | IMPORT_RESOLUTION |
| CALLS | `info.Uses` on selector expressions | GO_TYPES |
| DEFINES | `types.Named.NumMethods` | AST_EXACT |
| EMBEDS | `types.Struct.NumFields` with `f.Embedded()` | AST_EXACT |
| IMPLEMENTS | `types.Implements` against full method set | GO_TYPES |
| REFERENCES | Function signature params and results via `go/types` | GO_TYPES / AST_EXACT |

Every edge carries:

- `Confidence` — a float from 0.40 to 1.0, not a fixed constant
- `ResolutionStatus` — `RESOLVED`, `AMBIGUOUS`, `UNRESOLVED`, or `INFERRED`
- `ResolutionMethod` — `GO_TYPES`, `IMPORT_RESOLUTION`, `AST_EXACT`, or `HEURISTIC`
- `EpistemicClass` — `OBSERVATION`, `INFERENCE`, `DECISION`, or `POLICY`
- `Evidence` — the source file and line that justifies the edge

### Corpus results

The `garuda-bench` suite runs 20 synthetic fixtures covering interfaces, generics, embedding, aliases, polymorphism, and cross-module resolution. Latest run:

```text
Entity Precision:        100.0%
Entity Recall:           100.0%
Relationship Precision:  100.0%
Relationship Recall:     100.0%
```

These are synthetic fixtures. They validate the analyzer's correctness against known inputs. They do not constitute a real-repository precision measurement.

---

## The Policy Engine

Policies are declarative YAML. A pure-function evaluator checks each policy against the semantic graph. The result is one of four decisions — ALLOW, WARN, REVIEW, or BLOCK — and every decision is Merkle-anchored.

```mermaid
flowchart TD
    YAML["Policy YAML"] --> VALIDATE["Validate"]
    VALIDATE --> EVAL["Evaluate against Graph"]
    EVAL --> DECISION["Decision"]
    DECISION --> ANCHOR["Anchor Leaf"]
    ANCHOR --> MERKLE["Merkle Ledger"]
    MERKLE --> PROOF["Inclusion Proof"]
    PROOF --> VERIFY["Policy Verify"]
```

### Policy shape

```yaml
id: payment-no-direct-db
version: v1
title: "Payment code must not import database/sql directly"
priority: 200
language: go
authority: "team-platform@company.com"
scope:
  domain: payments
when:
  - type: claim_exists
    params:
      claim_type: IMPORTS
      from_name_pattern: ".*Payment.*"
      to_name_pattern: ".*sql.*"
then:
  decision: BLOCK
  reason: "Direct database/sql import from payment code violates hexagonal boundary."
```

Supported predicates:

- `entity_exists` — matching entities exist
- `claim_exists` — matching claims exist
- `contradiction_exists` — a `CONTRADICTED` claim verification exists
- `verification_missing` — an entity has no runtime verification
- `language_matches` — the scope contains entities of a given language

### CLI

```bash
garuda policy validate ./policies
garuda policy evaluate ./policies --workspace prod
garuda policy list
garuda policy show <evaluation-id>
garuda policy verify <evaluation-id>
```

`garuda policy verify` re-derives the inclusion proof and confirms the decision was committed at a specific block height. The verifier does not need to trust the graph — it recomputes the proof itself.

---

## Architecture

```mermaid
flowchart TB
    SOURCES["Sources"] --> ANALYZER["Analyzer"]
    RUNTIME["Runtime"] --> OTEL["OTel Ingestion"]

    ANALYZER --> CORE["Semantic Core"]
    OTEL --> CORE

    CORE --> ENTITIES["Entities"]
    CORE --> RELATIONSHIPS["Relationships"]
    CORE --> CLAIMS["Claims"]
    CORE --> EVIDENCE["Evidence"]

    ENTITIES --> VERIFY["Verification"]
    RELATIONSHIPS --> VERIFY
    CLAIMS --> VERIFY
    EVIDENCE --> VERIFY

    VERIFY --> SUP["SUPPORTED"]
    VERIFY --> UNV["UNVERIFIED"]
    VERIFY --> CON["CONTRADICTED"]

    SUP --> POLICY["Policy Engine"]
    UNV --> POLICY
    CON --> POLICY

    POLICY --> ALLOW["ALLOW"]
    POLICY --> WARN["WARN"]
    POLICY --> REVIEW["REVIEW"]
    POLICY --> BLOCK["BLOCK"]

    ALLOW --> TRUST["Merkle Trust Layer"]
    WARN --> TRUST
    REVIEW --> TRUST
    BLOCK --> TRUST

    TRUST --> SURFACES["CLI / HTTP / MCP / IDE"]
```

Every layer stores only what it can justify. Nothing fabricates certainty.

- Unknown stays unknown. `UNVERIFIED` is a first-class state, not a fallback.
- Observations are not decisions. The two are type-distinct at the schema level.
- Failed analysis preserves the last known-good snapshot. A new run never destroys an old one until the new one commits.

---

## Verification States

| State | Meaning |
| :--- | :--- |
| SUPPORTED | Static declarations and runtime telemetry agree. |
| UNVERIFIED | Static structure exists but lacks runtime evidence. Does not imply dead or broken code. |
| CONTRADICTED | Observed runtime behavior explicitly violates a recorded architectural policy or static expectation. Enters quarantine. |

```mermaid
flowchart TD
    STATIC["Static Analysis"] --> SEMANTIC["Semantic State"]
    RUNTIME["Runtime Telemetry"] --> ENGINE["Verification Engine"]
    SEMANTIC --> ENGINE

    ENGINE --> SUPPORTED["SUPPORTED"]
    ENGINE --> UNVERIFIED["UNVERIFIED"]
    ENGINE --> CONTRADICTED["CONTRADICTED"]

    CONTRADICTED --> QUARANTINE["Quarantine ARCH_DRIFT_001"]
```

> Absence of evidence is not evidence of absence. Static structure and runtime behavior answer different questions.

---

## Quick Start

### Build

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda
go build -o bin/garuda ./cmd/garuda
```

Go 1.25 or later. `CGO_ENABLED=1` is required for the tree-sitter TypeScript parser.

### Apply migrations

```bash
export DATABASE_URL="postgres://user:pass@localhost:5432/garuda?sslmode=disable"
psql "$DATABASE_URL" -f migrations/063_claims_resolution.sql
```

Apply all migrations in order.

### Analyze a repository

```bash
./bin/garuda analyze /path/to/go/repo --save --workspace my-workspace
```

### Evaluate policies

```bash
./bin/garuda policy evaluate ./policies --workspace my-workspace
```

### Verify a decision

```bash
./bin/garuda policy verify <evaluation-id>
```

### Start the daemon

```bash
./bin/garuda dev
```

Daemon listens on `http://localhost:8080` and exposes the graph visualizer at `/graph`.

---

## CLI Reference

| Command | Purpose |
| :--- | :--- |
| `garuda analyze` | Analyze a repository. Language is auto-detected. |
| `garuda diff` | Semantic diff between two snapshots. |
| `garuda inspect` | Inspect a semantic entity. |
| `garuda entities` | List all entities in the workspace. |
| `garuda graph` | Generate interactive HTML graph. |
| `garuda impact` | Blast-radius analysis for a symbol. |
| `garuda policy list` | List active policies. |
| `garuda policy validate` | Parse and validate policy YAML. |
| `garuda policy evaluate` | Run policies and anchor decisions. |
| `garuda policy show` | Show one evaluation with evidence and Merkle proof. |
| `garuda policy verify` | Re-derive the Merkle inclusion proof for a decision. |
| `garuda workspace` | Manage workspaces. |
| `garuda repo` | Manage repositories within a workspace. |
| `garuda dev` | Start the unified daemon. |
| `garuda mcp` | Run the MCP server over stdio. |
| `garuda verify` | Verify ledger integrity. |
| `garuda status` | Inspect Merkle root and daemon status. |

---

## Design Principles

1. **Evidence before confidence.** Every material assertion links to evidence.
2. **Unknown stays unknown.** `UNVERIFIED` is a state, not a failure mode.
3. **Deterministic identity.** The same semantic entity remains identifiable across analysis runs.
4. **Immutable history.** Correct the present with new records; do not rewrite the past.
5. **Observations are not decisions.** The two are type-distinct at every layer.
6. **Enforcement must be deterministic.** Policy decisions are computed, not inferred. Every decision is anchored.
7. **AI should consume structured state.** Agents should not reconstruct an entire repository from raw text when a reusable semantic state exists.
8. **Correctness before scale.** Validated semantics come before expanding language and deployment scope.

The full set of invariants is documented in [ADR-0002](docs/adr/0002-merkle-integrity-layer.md) and the laws encoded in every source file header.

---

## What Is Not Done

Explicit list. Do not claim any of these as current capability.

- **MCP server is not a long-lived server.** The current implementation reads one request from stdin and exits. Rebuild is scheduled as Phase 2.
- **TypeScript has never written to a validated workspace.** The parser exists but has not produced rows in `claims` for any real repository.
- **Real-repo relationship precision is unmeasured.** The 100% figure is synthetic-fixture only.
- **Runtime verification has zero observations.** The OTel ingestion endpoint exists. No agent has sent a span.
- **Documentation ingestion has zero verified claims.** The parser produces claims. Zero have been matched against code entities.
- **IDE extension is unvalidated.** Source exists. Not tested end-to-end.
- **Workspace snapshot is v0.** The decision log is v1. The workspace snapshot is a different construct. ADR-0003 will define its v1 model.
- **Merkle batching is not implemented.** Each write is currently its own single-leaf epoch. Batched multi-leaf epochs are a schema-compatible upgrade tracked for a follow-up commit.
- **External anchoring is not implemented.** Merkle roots live only in the local Postgres instance. Independent verification requires the source code of `internal/merkle/`, not an external timestamp or public ledger.
- **EU AI Act compliance is not a claim.** Garuda provides technical capabilities that may support a compliance program. It is not certified and does not substitute for legal assessment.

---

## Security

- Authentication: mutation endpoints are guarded by Ed25519 JWTs.
- Integrity: state changes are anchored to a verifiable Merkle root.
- Idempotency: decision commits pass through `idempotency_keys` to prevent duplicate execution.
- Runtime: the API runs in a non-root scratch container.

For vulnerability reporting, see [SECURITY.md](SECURITY.md).

Cryptographic mechanisms provide tamper-evident state and verification. They do not replace credential security, database security, access controls, key management, or operational security practices.

---

## Documentation

| Document | Purpose |
| :--- | :--- |
| [ADR-0002](docs/adr/0002-merkle-integrity-layer.md) | Merkle trust layer design |
| [ADR-0003](docs/adr/0003-workspace-state-snapshots.md) | Workspace snapshot design (proposed) |
| [PLAYBOOK.md](PLAYBOOK.md) | Installation, operational procedures, workflow examples |
| [EVIDENCE.md](EVIDENCE.md) | Validation artifacts and methodology |
| [SECURITY.md](SECURITY.md) | Security model and vulnerability reporting |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contribution workflow |

---

## Repository Structure

```text
garuda/
│
├── README.md
├── PLAYBOOK.md
├── EVIDENCE.md
├── SECURITY.md
├── CONTRIBUTING.md
├── LICENSE
│
├── assets/
│   └── garuda-logo.png
│
├── cmd/
├── internal/
│   ├── analyzer/
│   ├── ast/
│   ├── merkle/
│   ├── policy/
│   ├── store/
│   └── api/
│
├── migrations/
├── garuda-bench/
├── docs/
│   └── adr/
└── test/
```

---

## Contributing

Garuda is being built in the open. Relevant areas include compiler tooling, multi-language parsers, developer infrastructure, semantic graphs, runtime verification, MCP, AI coding agents, provenance, cryptographic state, and developer experience.

Start with [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

Apache License 2.0. See [LICENSE](LICENSE).

---

<p align="center">
  <strong>Garuda</strong><br>
  Analyze | Verify | Govern
</p>

<p align="center">
  <sub>
    If the system can provide evidence, don't replace it with a guess.<br>
    If the system can verify, don't rely on assumption.<br>
    If the context already exists, don't reconstruct it.
  </sub>
</p>
