<p align="center">
  <img src="assets/garuda-logo.png" width="128" alt="Garuda">
</p>

<h1 align="center">Garuda</h1>

<p align="center">
  <strong>Analyze | Verify | Govern</strong>
</p>

<p align="center">
  <em>One verified model of your software. Cryptographically anchored. Shared by engineers and their AI agents.</em>
</p>

<p align="center">
  <a href="https://github.com/myshra777-ai/garuda/releases"><img src="https://img.shields.io/badge/version-v0.1.x-blue.svg?style=flat-square" alt="Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache_2.0-green.svg?style=flat-square" alt="License"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.26-00ADD8.svg?style=flat-square" alt="Go Version"></a>
  <a href="internal/merkle/"><img src="https://img.shields.io/badge/trust_layer-RFC_6962-blueviolet.svg?style=flat-square" alt="Merkle"></a>
  <a href="docs/adr/0002-merkle-integrity-layer.md"><img src="https://img.shields.io/badge/ADR--0002-complete-brightgreen.svg?style=flat-square" alt="ADR-0002"></a>
</p>

---

## What makes Garuda different

Every software intelligence tool today can produce a graph. None of them can prove the graph is right.

Garuda treats every architectural decision, every policy evaluation, and every verification state as a cryptographically anchored fact. A decision is hashed into an RFC 6962 Merkle tree using a canonical, language-independent byte encoding. An auditor — with the source of the Merkle package and no access to Garuda's database, servers, or keys — can recompute the hash and confirm the decision existed at a specific block height.

That is the moat. Everything else in this README describes capabilities that other tools also provide. The trust layer is what none of them do.

```mermaid
flowchart LR
    DECISION["Architectural Decision"] --> HASH["Canonical Hash"]
    EVIDENCE["Evidence File & Line"] --> HASH
    POLICY["Policy Evaluation"] --> HASH
    HASH --> TREE["RFC 6962 Merkle Tree"]
    TREE --> ROOT["Epoch Root"]
    ROOT --> PROOF["Inclusion Proof"]
    PROOF --> AUDITOR["Independent Auditor"]
    ROOT --> AGENT["AI Agent"]
    ROOT --> CI["CI / Governance"]
```

Three properties fall out of this design, and no competitor ships all three:

**Verification is a state, not a score.** Every claim in the graph is `SUPPORTED`, `UNVERIFIED`, or `CONTRADICTED`. A number like "87% confidence" tells you nothing. A state tells you what to do next.

**Evidence is a first-class citizen.** Every relationship carries the source file and line that justifies it, plus a classification of how it was derived — compiler type information, import resolution, exact AST structure, or heuristic. Nothing is asserted without a traceable origin.

**Governance decisions live in the same ledger as the evidence they cited.** When a policy blocks a change, the block is anchored to the exact block that referenced the entities and claims involved. An auditor can walk the chain.

---

## The problem

Software teams lose time to three related failures.

**Onboarding is slow because structure is invisible.** A new engineer joins a repository with 500 files. Nothing tells them which service calls which, which package is safe to change, or which function has 400 callers. They read code for weeks before they can make a first commit.

**AI agents see files, not systems.** A developer asks Cursor to refactor a helper. The agent has the file open, sees the function signature, and proposes a change. It does not see the eleven other services that call the function across module boundaries. The change compiles. A week later, a downstream service starts returning 500s. The agent did nothing wrong — it was given a file and asked a filesystem question.

**Decisions evaporate.** Why is this service using PostgreSQL and not Redis? Was that a deliberate choice last year, or an accident from a refactor nobody reviewed? The answer lives in an archived Slack thread, or in a person who left. There is no shared record.

These are not tooling problems. They are the same problem: **no single, verified model of what the software actually is.**

---

## What changes for you

| Outcome | What changes | Mechanism |
|---|---|---|
| **Save time** | Onboard a new engineer in days, not weeks | `garuda graph` opens the entire workspace in the browser, drillable from repository to package to symbol to caller |
| **Save time** | Debug faster with an accurate caller/callee map | The graph is compiler-built, not regex-matched. `garuda impact <symbol>` reports the actual blast radius |
| **Save effort** | Refactor shared helpers without breaking downstream services | Every edge is backed by evidence; the affected callers are named, not guessed |
| **Save effort** | AI agents stop re-exploring the codebase from scratch | The MCP server answers "what calls this" in one query instead of a five-file read |
| **Save money** | Cut AI agent token spend by grounding in a real graph | Prompts shrink from whole-repository context to targeted structural queries |
| **Save money** | Catch architectural drift before it merges | Policies run in CI. Violations are anchored to the Merkle ledger before the PR merges |
| **Save money** | Compress incident response | Blast-radius analysis on the affected symbol returns in sub-second time |
| **Reduce hallucination** | AI agents stop inventing functions, callers, and signatures | The agent queries a verified model instead of guessing from a file |
| **Audit trail** | Every governance decision is independently verifiable | The Merkle inclusion proof is a pure function of `leaf`, `proof`, and `root` |

---

## How verification works

Garuda distinguishes three states for every claim in the graph.

| State | Meaning |
| :--- | :--- |
| **Supported** | Static declarations and runtime telemetry agree. |
| **Unverified** | Static structure exists but lacks runtime evidence. Does not imply dead or broken code. |
| **Contradicted** | Observed runtime behavior explicitly violates a recorded architectural policy or static expectation. |

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

> Absence of evidence is not evidence of absence. Static structure and runtime behavior answer different questions. A path that has not been observed stays `UNVERIFIED`; it is never silently promoted to "dead" or "broken."

---

## The trust layer, in detail

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

**Canonical encoding.** Every hash input is produced by a versioned byte encoder that emits a length-prefixed, field-ordered sequence. There is no dependency on language-native serializers, so two independent implementations in different languages produce byte-identical output for the same input. Golden vectors are frozen in [`testdata/vectors_v1.json`](internal/merkle/testdata/vectors_v1.json) and covered by tests. Any change to the encoding requires a version bump and a new design record.

**RFC 6962 tree with domain separation.**

```text
leaf_hash     = SHA256(0x00 || leaf_data)
internal_hash = SHA256(0x01 || left || right)
```

The `0x00` and `0x01` prefixes prevent a second-preimage attack where a leaf's bytes could equal an internal node's bytes. Odd-node handling uses promotion, closing CVE-2012-2459.

**Inclusion proofs are pure functions.**

```go
func VerifyInclusion(leafData []byte, proof []ProofNode, root []byte) bool
```

Three inputs. No storage access. No Garuda code in the call path. O(log n) hashing operations. For a tree of 100,000 leaves, a proof is approximately 17 nodes.

**Independent verification.** When a decision is anchored, the response includes the canonical leaf hash, the epoch root, the block height, and an inclusion proof. A third party with the source of the Merkle package — but no access to Garuda's infrastructure — can recompute the leaf, verify the proof, and confirm the epoch root. No trust in Garuda's database or processes is required.

---

## Garuda also carries these capabilities

The trust layer is the moat. Everything in this section is table stakes for a modern software intelligence platform — and Garuda carries it too.

### Semantic analysis across languages

Go, Python, and TypeScript are extracted into one unified semantic graph. Language is a property of each entity, not a property of the workspace, so the pipeline is shared across all three.

| Language | Extraction | Status |
| :--- | :--- | :--- |
| **Go** | Compiler-backed: packages, structs, interfaces, functions, methods, fields, imports, calls, implements, embeds | Stable |
| **Python** | Structural: classes, base classes, methods, functions, imports | Beta |
| **TypeScript** | tree-sitter: classes, interfaces, functions, methods, decorators, extends, implements, imports | Beta |

Rust, Java, and Elixir analyzers are in development, targeting the same pipeline.

### Semantic graph exploration

Interactive dashboard with repository, package, entity, and neighborhood views. Force-directed topology with community filtering. Every node drillable to its evidence.

### Impact analysis and blast radius

`garuda impact <symbol>` reports direct callers, transitive callers, and cross-repository edges. Sub-second response on workspaces with tens of thousands of edges.

### Global search

Fuzzy search across every symbol, package, file, and repository in the workspace.

### CLI and HTTP APIs

Every capability is exposed through both surfaces. Scripting and CI integration are first-class, not afterthoughts.

### IDE and agent integration

Garuda is a native Model Context Protocol server. Cursor, Claude Desktop, and any MCP-compatible client can query the semantic graph directly. AI coding agents stop reconstructing the repository from raw text and start querying a verified model.

### CI / CD governance

`garuda ci` and `garuda judge` evaluate proposed changes against a baseline, run the policy engine, and report contract breakage before merge.

### Policy engine

Declarative YAML policies evaluated against the semantic graph.

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

Five predicates — `entity_exists`, `claim_exists`, `contradiction_exists`, `verification_missing`, `language_matches`. Four decision outcomes — `ALLOW`, `WARN`, `REVIEW`, `BLOCK`. Every evaluation is anchored to the same ledger as the evidence it cited.

**Example policy:**

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

- Unknown stays unknown. `UNVERIFIED` is a first-class state, not a fallback.
- Observations are not decisions. The two are type-distinct at the schema level.
- Analysis runs preserve the last known-good state. A new run never destroys an old one until the new one commits.

---

## Relationship types

Every relationship in the graph carries a confidence score, a resolution status, the method used to derive it, an epistemic classification, and the source file and line that justifies it.

| Relationship | Source Signal | Resolution |
| :--- | :--- | :--- |
| Imports | Import specs | Import resolution |
| Calls | Type-checked selector expressions | Compiler type information |
| Defines | Named type method sets | Exact AST structure |
| Embeds | Struct field inspection | Exact AST structure |
| Implements | Interface satisfaction | Compiler type information |
| References | Function signature parameters and results | Compiler type information |

---

## Quick start

### Build from source

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda
go build -o bin/garuda ./cmd/garuda
```

Requires **Go 1.26** or later. CGO is required for the tree-sitter TypeScript parser.

### One-click stack

```bash
docker compose -f deploy/compose/docker-compose.prod.yml up -d
```

Brings up Postgres, the API, the worker, and the telemetry collector.

### Manual setup

```bash
export DATABASE_URL="postgres://user:pass@localhost:5432/garuda?sslmode=disable"

# Apply every migration in order.
for f in migrations/[0-9]*.sql; do
  psql "$DATABASE_URL" -f "$f"
done
```

### Analyze a repository

```bash
./bin/garuda analyze /path/to/repo --save --workspace my-workspace
```

Language is detected automatically from project markers (`go.mod`, `pyproject.toml`, `tsconfig.json`).

### Evaluate policies

```bash
./bin/garuda policy evaluate ./policies --workspace my-workspace
```

### Verify a decision

```bash
./bin/garuda policy verify <evaluation-id>
```

Re-derives the inclusion proof and confirms the decision was committed at a specific block height.

### Start the daemon

```bash
./bin/garuda dev
```

Serves the API on `http://localhost:8080` and exposes the graph visualizer at `/graph`.

---

## CLI reference

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

## Design principles

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

## What is available today, and what is launching

### Available now

- **Go analyzer.** Compiler-backed, validated against a controlled test corpus.
- **Python and TypeScript analyzers.** Structural extraction feeding the same semantic pipeline.
- **Merkle v1 trust layer.** RFC 6962, canonical encoding, golden vectors, independent verification.
- **Policy engine.** Five predicates, four decision outcomes, anchored evaluations.
- **Semantic graph explorer.** Interactive dashboard with repository, package, entity, and neighborhood views.
- **MCP server for AI agents.** Query the semantic graph over stdio from Cursor, Claude Desktop, or any MCP-compatible client.
- **CLI and HTTP APIs** for every capability.
- **CI/CD integration** via `garuda ci` and `garuda judge`.

### In beta with design partners

- **Multi-tenant identity. Workspace membership and per-workspace access control. Design partners are running early builds today. Enterprise SSO is on the roadmap.
- **Runtime verification at scale.** Coverage semantics, cross-service correlation, and a broadened contradiction test suite.
- **Multi-repository benchmarks.** Quality gates for workspaces of 10, 25, and 100+ repositories.

### Launching soon

- **Rust, Java, and Elixir analyzers** on the same pipeline. No schema changes required to add them.
- **Extended MCP tool surface** for deeper agent workflows.
- **Broader evidence exploration** in the dashboard.

The roadmap is tracked in [`docs/ROADMAP.md`](docs/ROADMAP.md) and `docs/ROADMAP_INTERNAL.md`.

---

## Scope

Garuda's Go analyzer is compiler-backed and validated against a controlled test corpus. The Python and TypeScript analyzers provide structural extraction across the same semantic pipeline.

Runtime verification correlates static analysis with OpenTelemetry spans. The correlation path is implemented and exercised against synthetic spans in tests. Production correlation is in beta with design partners. Unobserved runtime paths remain in the unverified state and are never interpreted as dead or incorrect.

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

## Repository structure

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
