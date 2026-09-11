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
  <a href="https://github.com/myshra777-ai/garuda/releases"><img src="https://img.shields.io/badge/version-v0.1.x-blue.svg?style=flat-square" alt="Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache_2.0-green.svg?style=flat-square" alt="License"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.25-00ADD8.svg?style=flat-square" alt="Go Version"></a>
  <a href="internal/merkle/"><img src="https://img.shields.io/badge/trust_layer-RFC_6962-blueviolet.svg?style=flat-square" alt="Merkle"></a>
  <a href="docs/adr/0002-merkle-integrity-layer.md"><img src="https://img.shields.io/badge/ADR--0002-complete-brightgreen.svg?style=flat-square" alt="ADR-0002"></a>
</p>

<p align="center">
  Garuda gives teams a shared, evidence-backed understanding of their software systems — with every governance decision anchored to a cryptographic ledger any third party can independently verify.
</p>

---

## What Garuda Is

Modern software organizations run on distributed knowledge. Architecture decisions live in pull requests, interfaces live in source code, runtime behavior lives in telemetry, and the reasoning behind all of it lives in people's heads.

Garuda brings these layers into one structured, verifiable model. It connects source code, runtime observations, and governance decisions — and it anchors each decision to a cryptographic log that a third party can independently confirm.

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

Garuda does three things:

1. **Analyze.** Extract entities and relationships from source code into a structured semantic model. Every relationship links back to the evidence that justifies it.
2. **Verify.** Correlate static analysis with runtime observations. Distinguish what is supported by evidence from what remains unverified.
3. **Govern.** Evaluate declarative policies against the semantic model and anchor every decision — allow, warn, review, or block — to a cryptographic ledger.

Garuda is not another text search engine, another generic dependency graph, another documentation generator, or an LLM that guesses how a repository works.

---

## Capabilities

### Semantic Analysis

Garuda analyzes Go, Python, and TypeScript source code into one unified semantic graph. Language is a property of the entity, not the workspace — the pipeline is shared across all supported languages.

| Language | Extraction |
| :--- | :--- |
| Go | Compiler-backed. Packages, structs, interfaces, functions, methods, fields, imports, calls, implements, embeds. |
| Python | Structural. Classes, base classes, methods, functions, and imports. |
| TypeScript | tree-sitter based. Classes, interfaces, functions, methods, decorators, extends, implements, and imports. |

### Cryptographic Trust Layer

Every governance decision — from a policy evaluation or an architectural choice — is anchored to a versioned Merkle log. Decisions are cryptographically committed and independently verifiable.

### Policy Engine

Declarative YAML policies evaluated against the semantic graph. Each evaluation produces one of four outcomes — allow, warn, review, or block — anchored to the same ledger as the evidence it cited.

### Evidence Ledger

Every relationship in the semantic graph carries the source file and line that justifies it. Nothing is asserted without a traceable origin.

### Integration Surface

Garuda exposes its state through a CLI, an HTTP API, a dashboard, an interactive graph visualizer, a Model Context Protocol server, and an IDE extension.

---

## How Verification Works

Garuda distinguishes three verification states for every claim in the graph.

| State | Meaning |
| :--- | :--- |
| Supported | Static declarations and runtime telemetry agree. |
| Unverified | Static structure exists but lacks runtime evidence. Does not imply dead or broken code. |
| Contradicted | Observed runtime behavior explicitly violates a recorded architectural policy or static expectation. |

```mermaid
flowchart TD
    STATIC["Static Analysis"] --> SEMANTIC["Semantic State"]
    RUNTIME["Runtime Telemetry"] --> ENGINE["Verification Engine"]
    SEMANTIC --> ENGINE

    ENGINE --> SUPPORTED["Supported"]
    ENGINE --> UNVERIFIED["Unverified"]
    ENGINE --> CONTRADICTED["Contradicted"]

    CONTRADICTED --> QUARANTINE["Quarantine"]
```

> Absence of evidence is not evidence of absence. Static structure and runtime behavior answer different questions.

---

## The Trust Layer

The trust layer lives in [`internal/merkle/`](internal/merkle/) and is specified in full in [ADR-0002](docs/adr/0002-merkle-integrity-layer.md).

```mermaid
flowchart TD
    DECISION["Policy Decision"] --> CANON["Canonical Encoder"]
    CANON --> LEAF["Leaf Hash"]
    LEAF --> TREE["RFC 6962 Merkle Tree"]
    TREE --> ROOT["Epoch Root"]
    ROOT --> PROOF["Inclusion Proof"]
    PROOF --> VERIFY["Independent Verification"]
```

### Canonical Encoding

Every hash input is produced by a versioned byte encoder that emits a length-prefixed, field-ordered sequence. There is no dependency on language-native serializers, so two independent implementations in different languages produce byte-identical output for the same input.

Golden vectors for the encoder are frozen in [`testdata/vectors_v1.json`](internal/merkle/testdata/vectors_v1.json) and covered by tests. Any change requires a version bump and a new design record.

### Merkle Tree

The tree is RFC 6962 compliant with explicit domain separation:

```text
leaf_hash     = SHA256(0x00 || leaf_data)
internal_hash = SHA256(0x01 || left || right)
```

Domain separation prevents second-preimage attacks. Odd-node handling uses promotion, closing a known Merkle tree collision vector.

### Inclusion Proofs

An inclusion proof is a list of position-and-hash nodes. Verification is a pure function:

```go
func VerifyInclusion(leafData []byte, proof []ProofNode, root []byte) bool
```

It takes three inputs, touches no storage, and performs O(log n) hashing operations. For a tree of 100,000 leaves, a proof is approximately 17 nodes.

### Versioning

Two Merkle schemes coexist for backward compatibility:

| Version | Scheme | Verification |
| :--- | :--- | :--- |
| 0 | Linear hash chain | Legacy path |
| 1 | RFC 6962 tree with canonical leaves | `VerifyInclusion` |

Every record carries `verification_version`, so verification selects the correct path automatically.

### Independent Verification

When a decision is anchored, the response includes the canonical leaf hash, the epoch root, the block height, and an inclusion proof. A third party with the source of the Merkle package — but no access to Garuda's infrastructure — can recompute the leaf, verify the proof, and confirm the epoch root. No trust in Garuda's database or processes is required.

---

## The Analyzer

The Go analyzer produces a typed semantic snapshot from a Go module or workspace.

### Relationship Types

| Relationship | Source Signal | Resolution |
| :--- | :--- | :--- |
| Imports | Import specs | Import resolution |
| Calls | Type-checked selector expressions | Compiler type information |
| Defines | Named type method sets | Exact AST structure |
| Embeds | Struct field inspection | Exact AST structure |
| Implements | Interface satisfaction | Compiler type information |
| References | Function signature parameters and results | Compiler type information |

Every relationship carries a confidence score, a resolution status, the method used to derive it, an epistemic classification, and the source file and line that justifies it.

### Validation

The analyzer is validated against a test corpus covering interfaces, generics, embedding, type aliases, polymorphism, and cross-module resolution. Precision and recall are measured for entities and relationships separately.

---

## The Policy Engine

Policies are declarative YAML. A pure-function evaluator checks each policy against the semantic graph and produces one of four decisions.

```mermaid
flowchart TD
    YAML["Policy YAML"] --> VALIDATE["Validate"]
    VALIDATE --> EVAL["Evaluate"]
    EVAL --> DECISION["Decision"]
    DECISION --> ANCHOR["Anchor to Ledger"]
    ANCHOR --> MERKLE["Merkle Trust Layer"]
    MERKLE --> PROOF["Inclusion Proof"]
    PROOF --> VERIFY["Independent Verify"]
```

### Example Policy

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

### Supported Predicates

- `entity_exists` — matching entities exist
- `claim_exists` — matching claims exist
- `contradiction_exists` — a contradicted verification exists
- `verification_missing` — an entity has no runtime verification
- `language_matches` — the scope contains entities of a given language

---

## Architecture

```mermaid
flowchart TB
    SOURCES["Sources"] --> ANALYZER["Analyzer"]
    RUNTIME["Runtime"] --> OTEL["Telemetry Ingestion"]

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

    VERIFY --> SUP["Supported"]
    VERIFY --> UNV["Unverified"]
    VERIFY --> CON["Contradicted"]

    SUP --> POLICY["Policy Engine"]
    UNV --> POLICY
    CON --> POLICY

    POLICY --> OUTCOMES["Allow | Warn | Review | Block"]

    OUTCOMES --> TRUST["Merkle Trust Layer"]
    TRUST --> SURFACES["CLI | HTTP | MCP | IDE"]
```

Every layer stores only what it can justify.

- Unknown stays unknown. Unverified is a first-class state, not a fallback.
- Observations are not decisions. The two are type-distinct at the schema level.
- Analysis runs preserve the last known-good state. A new run never destroys an old one until the new one commits.

---

## Quick Start

### Build

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda
go build -o bin/garuda ./cmd/garuda
```

Go 1.25 or later. CGO is required for the tree-sitter TypeScript parser.

### Apply Migrations

```bash
export DATABASE_URL="postgres://user:pass@localhost:5432/garuda?sslmode=disable"
psql "$DATABASE_URL" -f migrations/063_claims_resolution.sql
```

Apply all migrations in order.

### Analyze a Repository

```bash
./bin/garuda analyze /path/to/repo --save --workspace my-workspace
```

Language is detected automatically from project markers.

### Evaluate Policies

```bash
./bin/garuda policy evaluate ./policies --workspace my-workspace
```

### Verify a Decision

```bash
./bin/garuda policy verify <evaluation-id>
```

Re-derives the inclusion proof and confirms the decision was committed at a specific block height.

### Start the Daemon

```bash
./bin/garuda dev
```

The daemon serves the API on `http://localhost:8080` and exposes the graph visualizer at `/graph`.

---

## CLI Reference

| Command | Purpose |
| :--- | :--- |
| `garuda analyze` | Analyze a repository. Language is auto-detected. |
| `garuda diff` | Semantic diff between two snapshots. |
| `garuda inspect` | Inspect a semantic entity. |
| `garuda entities` | List all entities in the workspace. |
| `garuda graph` | Generate an interactive HTML graph. |
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
2. **Unknown stays unknown.** Unverified is a state, not a failure mode.
3. **Deterministic identity.** The same entity remains identifiable across analysis runs.
4. **Immutable history.** Correct the present with new records; do not rewrite the past.
5. **Observations are not decisions.** The two are type-distinct at every layer.
6. **Enforcement must be deterministic.** Policy decisions are computed, not inferred. Every decision is anchored.
7. **AI consumes structured state.** Agents query reusable semantic state instead of reconstructing a repository from raw text.
8. **Correctness before scale.** Validated semantics come before expanding language and deployment coverage.

The full set of invariants is documented in [ADR-0002](docs/adr/0002-merkle-integrity-layer.md) and encoded in every source file header.

---

## Scope

Garuda's Go analyzer is compiler-backed and validated against a controlled test corpus. The Python and TypeScript analyzers provide structural extraction across the same semantic pipeline.

Runtime verification correlates static analysis with OpenTelemetry spans. Unobserved runtime paths remain in the unverified state and are never interpreted as dead or incorrect.

Merkle roots are anchored in the local persistence layer. Independent verification requires only the source of the Merkle package, not an external timestamp authority or a public ledger.

Garuda is not a substitute for legal assessment or compliance certification. It provides technical capabilities that can support a governance program.

---

## Security

- **Authentication.** Mutation endpoints are guarded by Ed25519 JWTs.
- **Integrity.** State changes are anchored to a verifiable Merkle root.
- **Idempotency.** Decision commits pass through `idempotency_keys` to prevent duplicate execution.
- **Runtime.** The API runs in a non-root scratch container.

For vulnerability reporting, see [SECURITY.md](SECURITY.md).

Cryptographic mechanisms provide tamper-evident state and verification. They do not replace credential security, database security, access controls, key management, or operational security practices.

---

## Documentation

| Document | Purpose |
| :--- | :--- |
| [ADR-0002](docs/adr/0002-merkle-integrity-layer.md) | Merkle trust layer design |
| [ADR-0003](docs/adr/0003-workspace-state-snapshots.md) | Workspace snapshot design |
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

Garuda is developed in the open. Contributions are welcome across compiler tooling, language parsers, developer infrastructure, semantic graphs, runtime verification, cryptographic state, and developer experience.

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
