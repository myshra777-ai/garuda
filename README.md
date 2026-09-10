```markdown
# 🦅 Garuda

<p align="center">
  <img src="assets/garuda-logo.png" width="128" alt="Garuda — Evidence-backed software intelligence">
</p>

<h1 align="center">Garuda</h1>

<p align="center">
  <strong>Analyze | Verify | Govern | Collaborate</strong>
</p>

<p align="center">
  <em>A deterministic data governance plane for humans and AI agents.</em>
</p>

<p align="center">
  <a href="https://github.com/myshra777-ai/garuda/releases"><img src="https://img.shields.io/badge/version-v1.0.0-blue.svg?style=flat-square" alt="Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-green.svg?style=flat-square" alt="License"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.25-00ADD8.svg?style=flat-square&logo=go" alt="Go Version"></a>
  <a href="EVIDENCE.md"><img src="https://img.shields.io/badge/verification-14%20repos%20passed-success.svg?style=flat-square" alt="Verification Status"></a>
  <a href="EVIDENCE.md"><img src="https://img.shields.io/badge/GAP--20-~87.2%25%20token%20reduction-purple.svg?style=flat-square" alt="GAP-20 Benchmark"></a>
</p>

<p align="center">
  Garuda gives humans and AI agents a shared, evidence-backed understanding of a software system.
</p>

<p align="center">
  <a href="PLAYBOOK.md">Playbook</a> ·
  <a href="EVIDENCE.md">Evidence</a> ·
  <a href="BENCHMARKS_AND_VERIFICATION_v1.0.0.md">Benchmarks</a> ·
  <a href="docs/">Architecture</a> ·
  <a href="SECURITY.md">Security</a> ·
  <a href="CHANGELOG.md">Changelog</a>
</p>

---

## Overview

Modern software teams are increasingly distributed across humans, heterogeneous AI coding agents, repositories, services, runtime systems, and constantly changing architectural decisions.

Without a shared semantic state, each participant repeatedly reconstructs the architecture from raw source text, isolated conversations, or fragmented documentation.

Garuda is a **deterministic data governance plane** that persists system understanding as structured, verifiable data. It connects:

- Canonical semantic entities and dependencies
- Bitemporal architectural decisions and policies
- Cryptographic state verification (Merkle ledgers)
- Agent checkpoints and CAS (Compare-and-Swap) handoffs
- Runtime observations and telemetry

The central idea:

> **Humans and AI agents should not have to rediscover the software system independently every time they work on it, and state transitions must be mathematically verifiable.**

```mermaid
flowchart LR
    CODE["Source & Runtime"] --> GARUDA["Garuda Workspace"]
    GARUDA --> MERKLE["Merkle Ledger"]
    GARUDA --> MCP["MCP Bridge"]
    GARUDA --> CAS["Agent CAS Checkpoints"]

    MERKLE --> AI["AI Agents"]
    MCP --> AI
    CAS --> AI

    MERKLE --> HUMANS["Developers & CI"]
```

---

## What Is Garuda?

Garuda treats machine-relevant engineering state as a structured, inspectable, immutable ledger that can be inspected, queried, revised, verified, and shared.

```text
Garuda Engine
   │
   ├── Semantic Graph (Entities, Topology, Lineage)
   ├── Decision Ledger (Bitemporal Policies, Quarantines)
   ├── Agent Handoffs (CAS Checkpoints, State Resumption)
   ├── Cryptographic State (Merkle Roots, Block Heights)
   ├── Telemetry Pipeline (Observations, Span Ingestion)
   └── Secure Gateway (Ed25519 JWT Auth, MCP Bridge)
```

Garuda is therefore **not simply a graph generator**. The graph is one representation of a broader semantic state.

---

## The Core Loop: Analyze | Verify | Govern | Collaborate

| Capability      | Engine Function                                                                              | Why it matters                                                        |
| --------------- | -------------------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| **Analyze**     | Extracts topology, entities, and dependencies from source and telemetry.                      | Creates a structured, machine-readable software understanding.         |
| **Verify**      | Hashes state into an immutable Merkle tree ledger (`X-Garuda-Block-Height`).                  | Ensures deterministic integrity of all architectural claims.           |
| **Govern**      | Enforces bitemporal decision policies and contradictions.                                     | Separates verified knowledge from assumptions and flags architectural drift. |
| **Collaborate** | Manages CAS-based agent handoffs and context resumption.                                      | Allows multiple AI models (e.g., Claude, GPT-4o) to safely pause and resume tasks. |

```mermaid
flowchart LR
    ANALYZE["ANALYZE<br/>Understand the system"] -->
    VERIFY["VERIFY<br/>Compare claims with evidence"] -->
    GOVERN["GOVERN<br/>Enforce policies & quarantine drift"] -->
    COLLABORATE["COLLABORATE<br/>Share state across agents"]
```

---

## The Garuda Semantic Model

Garuda separates different kinds of software knowledge instead of collapsing them into a single graph or confidence score.

```mermaid
flowchart TB
    OBS["OBSERVATIONS<br/>What was observed"]
    CLAIM["CLAIMS<br/>What the observations imply"]
    EVID["EVIDENCE<br/>Why the claim is supported"]
    VERIFY["VERIFICATION<br/>What has been confirmed"]
    DECISION["DECISIONS<br/>What the team agreed"]
    LINEAGE["LINEAGE<br/>How the state evolved"]

    OBS --> CLAIM
    CLAIM --> EVID
    EVID --> VERIFY
    VERIFY --> DECISION
    DECISION --> LINEAGE
```

1. **Observation** — AST structures, resolved symbols, source declarations, runtime telemetry, runtime spans.
2. **Claim** — calls, interface, dependency relationships, repository topology.
3. **Evidence** — source files, line locations, revisions, runtime observations, provenance.
4. **Verification** — `SUPPORTED`, `UNVERIFIED`, or `CONTRADICTED`.
5. **Decision** — architectural choices, accepted / rejected approaches, policies, unresolved decisions.
6. **Lineage** — how the current state evolved.

---

## Agent Checkpoints & CAS Handoffs

Garuda v1.0.0 introduces **Compare-and-Swap (CAS) state handoffs**, allowing long-running AI workflows to span multiple agent executions without losing context.

Instead of passing massive context windows between sessions, agents checkpoint their state into Garuda and hand off a UUID to another agent:

- **Agent A** (e.g., Claude-3.5) completes a system analysis and commits a `checkpoint_id`.
- **Agent B** (e.g., GPT-4o) resumes the exact execution state via `POST /api/v1/agents/resume`.
- Garuda **invalidates the checkpoint** upon resumption to prevent replay attacks and split-brain scenarios. Re-resume attempts return `404 Checkpoint Not Found`.

```mermaid
flowchart LR
    A["Agent A (Claude-3.5)"] -->|checkpoint| GARUDA["Garuda"]
    GARUDA -->|checkpoint_id| B["Agent B (GPT-4o)"]
    B -->|POST /agents/resume| GARUDA
    GARUDA -->|invalidated on resume| B
```

Checkpoint and resume state transitions are guaranteed via **atomic Compare-And-Swap**, executed under PostgreSQL `SERIALIZABLE` isolation.

---

## Cryptographic State & The Merkle Ledger

Garuda guarantees that architectural invariants are tamper-evident. Every decision, policy change, and semantic update increments the engine's block height and produces a new Merkle root.

```mermaid
flowchart TD
    DECISION["Architectural Decision"] --> HASH["Hash Invariant"]
    HASH --> TREE["Merkle Tree"]
    TREE --> ROOT["Active Merkle Root"]
    ROOT --> HEADER["X-Garuda-Merkle-Root"]
```

All protected API responses return the `X-Garuda-Merkle-Root` and `X-Garuda-Block-Height` headers, allowing consumers and AI agents to cryptographically verify the current state of the workspace.

### Decision Commit Pipeline

- **Contract**: `POST /api/v1/decisions`
- **Guarantees**:
  - Mandatory UUID generation (`decision_id`)
  - Actor derivation exclusively via validated cryptographic context (`r.Context()`), preventing client-side actor impersonation
  - Structured domain-scoped taxonomy (`Domain`, `System`, `Team`, `Env`, `Region`)
  - Idempotency verification via `idempotency_keys` preventing duplicate execution
  - Atomic calculation of SHA-256 content hashes and Merkle root updates within a single serializable transaction

---

## Model Context Protocol (MCP) Integration

Garuda acts as a native MCP server, exposing its deterministic state directly to IDEs (Cursor) and AI clients (Claude Desktop).

**Exposed MCP tools:**

- `garuda_propose_decision` — Propose governance decisions with automatic contradiction evaluation.
- `garuda_handoff_task` — Execute atomic, crash-safe task handoffs to another agent.
- `garuda_resume_agent` — Resume an agent's execution context from a specific checkpoint.
- `garuda_get_lineage` — Retrieve the full DAG lineage for a task or semantic entity.

---

## Verification Flow & Contradiction Quarantines

Garuda correlates static source code with runtime telemetry. When an evaluated runtime observation conflicts with the expected architecture, the engine flags it as a **contradiction** and places it into quarantine.

| State          | Meaning                                                                                                                       |
| -------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| `SUPPORTED`    | Static declarations and runtime telemetry agree.                                                                              |
| `UNVERIFIED`   | Static structure exists, but lacks runtime evidence. (This does **not** imply dead or broken code.)                           |
| `CONTRADICTED` | Observed runtime behavior explicitly violates a recorded architectural policy or static expectation. Enters **Quarantine**.   |

> **Absence of evidence is not evidence of absence.** Static structure and runtime behavior answer different questions.

```mermaid
flowchart TD
    SOURCE["SOURCE CODE"] --> STATIC["STATIC ANALYSIS"]
    STATIC --> GRAPH["SEMANTIC STATE"]

    GRAPH --> ENTITIES["ENTITIES"]
    GRAPH --> RELATIONSHIPS["RELATIONSHIPS"]
    GRAPH --> EVIDENCE["EVIDENCE"]

    RUNTIME["RUNTIME TELEMETRY"] --> VERIFY["VERIFICATION ENGINE"]
    ENTITIES --> VERIFY
    RELATIONSHIPS --> VERIFY
    EVIDENCE --> VERIFY

    VERIFY --> SUPPORTED["SUPPORTED"]
    VERIFY --> UNVERIFIED["UNVERIFIED"]
    VERIFY --> CONTRADICTED["CONTRADICTED"]

    CONTRADICTED --> QUARANTINE["QUARANTINED / ARCH_DRIFT_001"]
```

Status updates propagate immediately to the Mission Control analytics stream without requiring process restarts.

---

## Zero-Turn Agent Discovery & Context Warmup

- **Contract**: `GET /system/bootstrap` & `POST /api/v1/agents/warmup`
- **Guarantees**:
  - Machine-readable bootstrap manifest providing the active Merkle root, block height, budget constraints, registered MCP tools, and system routes in sub-5ms.
  - Conserves context tokens by pre-computing topology and governance boundaries.

---

## One-Click Deployment

Garuda v1.0.0 ships as a production-ready, zero-friction Docker Compose stack. The API gateway is compiled as a fully static, non-root scratch container (Go 1.25) running behind an Ed25519 JWT security boundary.

### 1. Start the Engine

```bash
docker compose up -d
```

This automatically provisions the isolated `garuda-net` bridge, boots PostgreSQL, and executes all 58+ schema migrations.

### 2. Verify Bootstrap State

```bash
curl -s http://localhost:8080/system/bootstrap | jq .
```

Expected output (fresh deployment):

```json
{
  "system": "Garuda Engine",
  "version": "v1.0.0",
  "active_merkle_root": "genesis_root_00000000000000000000000000000000",
  "block_height": 0
}
```

### 3. Generate Dev Token & Access MCP

```bash
garuda token --actor=cli-operator
garuda mcp
```

### Build from Source

```bash
docker build \
  --build-arg VERSION=v1.0.0 \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  -t garuda-api:v1.0.0 .
```

---

## Empirical Benchmark Results (v1.0.0)

> **Disclaimer**: The benchmark metrics, latencies, and transaction throughput reported below were measured in an isolated testing environment running locally on Linux x86_64 (`garuda_test` schema, Docker 26.x, PostgreSQL 16 on local NVMe storage). Real-world production results will vary depending on network topology, database connection pool sizing, disk IOPS, hardware specs, and agent swarm concurrency.

### 1. Decision Ledger Write Throughput

The decision submission engine was subjected to high-concurrency load testing to evaluate lock contention under PostgreSQL serializable isolation and synchronous cryptographic hashing.

- **Concurrency**: 20 parallel workers
- **Total Commits**: 500 decisions
- **Payload**: Structured `SubmitDecisionRequest` with dynamic UUIDs, scoped domains, and unique idempotency keys
- **Hash Calculation**: Synchronous SHA-256 leaf and root Merkle calculation per commit

| Metric                                | Result           |
| :------------------------------------ | :--------------- |
| **Total Requests**                    | 500              |
| **Successful Commits (HTTP 201)**     | 500              |
| **Failures / Aborts (HTTP 4xx/5xx)**  | 0                |
| **Success Rate**                      | **100.0%**       |
| **Total Wall-Clock Time**             | 6.535 seconds    |
| **Sustained Throughput**              | **76.50 req/sec**|
| **Transaction Latency (Average)**     | ~13.07 ms        |

### 2. Micro-Benchmark Operation Latencies

Measured via end-to-end HTTP traces against the local Unix socket / TCP loopback interface:

| Operation                          | Route                              | Status        | Duration     |
| :--------------------------------- | :--------------------------------- | :------------ | :----------- |
| **System Bootstrap Handshake**     | `GET /system/bootstrap`            | `200 OK`      | **3 – 4 ms** |
| **Telemetry Span Ingestion**       | `POST /api/v1/telemetry/spans`     | `200 OK`      | **4 ms**     |
| **Agent Checkpoint Creation**      | `POST /api/v1/agents/checkpoint`   | `201 Created` | **4 ms**     |
| **Agent State Resume (CAS Lock)**  | `POST /api/v1/agents/resume`       | `200 OK`      | **7 ms**     |
| **Single Decision Hash Commit**    | `POST /api/v1/decisions`           | `201 Created` | **10 – 34 ms** |
| **Dashboard Analytics Rollup**     | `GET /api/v1/dashboard/stats`      | `200 OK`      | **69 – 97 ms** |

> These values describe observed validation runs, not universal production SLAs.

### 3. State Verification Telemetry Snapshot

The following state proof reflects the live database ledger status upon completion of the verification suite:

```json
{
  "system": "Garuda Engine",
  "version": "v1.0.0",
  "block_height": 38,
  "active_merkle_root": "e4361a7582ef01547f74cd39d23b8ac2c805db15d18c1692e6016a14c6f2182a",
  "trust_status": "Verified",
  "total_claims": 8515,
  "contradicted": 1,
  "quarantined_count": 1,
  "tokens_saved": 1428500,
  "estimated_cost_saved_usd": 28.57,
  "active_agents_count": 4,
  "cold_start_latency_ms": 42.5
}
```

---

## Production Artifact Specifications

The distribution pipeline compiles a hermetic, statically linked ELF executable bundled into a zero-dependency scratch container.

**Binary Footprint**

- **Binary File**: `bin/garuda-api-linux-amd64`
- **Format**: ELF 64-bit LSB executable, x86-64, statically linked, stripped
- **Binary Size**: 21 MB
- **Build Flags**: `-trimpath -ldflags="-s -w -extldflags '-static'"`
- **CGO Dependencies**: `CGO_ENABLED=0` (Pure Go runtime)

**Container Footprint**

- **Base Image**: `scratch` (unprivileged non-root user `65532:65532`)
- **Included Bundles**: Upstream Mozilla CA certificates (`ca-certificates.crt`), IANA time zone data (`zoneinfo`)
- **Container Image Size**: 30.5 MB
- **Exposed Ports**: `8080/TCP`

---

## GAP-20 Grounding Benchmark

The GAP-20 benchmark compares unassisted LLM exploration against Garuda-grounded workflows.

| Metric                        | Naive LLM | Garuda-Grounded |     Observed Difference |
| ----------------------------- | --------: | --------------: | ----------------------: |
| Symbol precision              |     40.0% |          100.0% |   +60 percentage points |
| Structural hallucination rate |     66.7% |            0.0% | −66.7 percentage points |
| Prompt token overhead         |     4,850 |             620 |        ~87.2% reduction |
| Cross-Agent Handoff Success   |      0.0% |          100.0% | CAS checkpoint state restored |
| Upstream caller recall        |     20.0% |          100.0% |   +80 percentage points |
| Downstream dependency recall  |     33.0% |          100.0% |   +67 percentage points |
| Violation quarantine rate     |      0.0% |          100.0% |  +100 percentage points |

### What the benchmark demonstrates

In the evaluated benchmark tasks:

- Garuda-grounded workflows reached **100% symbol precision**
- Structural hallucination rate was **0%**
- Context token overhead was reduced by approximately **87.2%**
- Tested upstream and downstream dependency recall improved to **100%**
- Evaluated violations were quarantined at **100%**

The structural hallucination result refers specifically to **fabricated functions, receivers, and symbols** in the evaluated benchmark tasks. It does **not** claim elimination of all forms of AI hallucination.

---

## Validation Evidence

Garuda has been evaluated against **14 heterogeneous Go repositories**, including Garuda itself, in controlled test environments.

| Metric                   |      Result |
| ------------------------ | ----------: |
| Repositories             |          14 |
| Packages                 |         143 |
| Entities                 |       3,675 |
| Relationships            |       5,679 |
| Cross-repository bridges |          55 |
| Contradictions injected  |          10 |
| Contradictions detected  |     10 / 10 |
| Entity precision         |        100% |
| Relationship precision   |       99.9% |
| GAP-20 token reduction   |      ~87.2% |
| Merkle ledger height     | Block #898+ |

For full methodology, screenshots, validation artifacts, and historical runs, see [EVIDENCE.md](EVIDENCE.md).

---

## Business and Engineering Impact

The current validation corpus includes controlled workflow measurements covering onboarding, catch-up, architectural review, dependency lookup, and AI context usage.

| Workflow                       | Traditional Approach                        | Garuda Workflow                                  | Observed Result                                        |
| ------------------------------ | ------------------------------------------- | ------------------------------------------------ | ------------------------------------------------------ |
| New engineer onboarding        | Manual documentation and repository tracing | Interactive topology and blast-radius navigation | ~80% faster time-to-first-PR in the evaluated workflow |
| Post-leave catch-up            | Meetings and manual PR archaeology          | Structured workspace summary and topology        | ~15 minutes in the evaluated workflow                  |
| PR architectural review        | Manual contract inspection                  | Automated CI evaluation                          | Automated detection in the evaluated benchmark         |
| Incident blast-radius analysis | Manual search and dependency tracing        | Symbol-level dependency lookup                   | Sub-second lookup in the evaluated workflow            |
| AI context overhead            | 4,850 tokens                                | 620 tokens                                       | ~87.2% reduction in GAP-20 benchmark                   |

> These are benchmarked observations from controlled environments and specific workflows. They are **not universal guarantees**.

---

## IDE Integration

Garuda integrates with editor workflows through Language Server Protocol and Model Context Protocol components.

### Inline Contradictions

When an evaluated runtime observation conflicts with the expected architecture, the relevant code location can be surfaced in the editor.

### Symbol Context and Blast Radius

Symbols can expose their surrounding dependency context and recursive relationships.

```text
🦅 Garuda Architectural Context: HarvestedDecision

Blast Radius:
4 Upstream Callers | 0 Downstream Dependencies

Direct Callers:

harvester (external)
github.com/myshra777-ai/garuda/internal/harvester

garuda (external)
github.com/myshra777-ai/garuda/cmd/garuda

[ Open in Visualizer ]
```

### Interactive Topology

The development daemon can host the topology visualizer at:

```text
http://localhost:8080/graph
```

A complete UI walkthrough is available in [docs/WALKTHROUGH.md](docs/WALKTHROUGH.md).

---

## Quick Start

### Install

```bash
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | sh
```

### Initialize

```bash
garuda init
```

### Start the stack

```bash
garuda up
```

### Analyze a repository

```bash
garuda analyze .
```

### Start the development daemon

```bash
garuda dev
```

The unified daemon provides the local API, telemetry collector, Merkle worker, and UI components.

### Open the topology visualizer

```bash
open http://localhost:8080/graph
```

---

## MCP Configuration

For Cursor or Claude Desktop, configure the Garuda MCP server against the same workspace and PostgreSQL database used by the team.

```json
{
  "mcpServers": {
    "garuda": {
      "command": "/usr/local/bin/garuda",
      "args": ["mcp"],
      "env": {
        "DATABASE_URL": "postgres://postgres:postgres@localhost:5432/garuda?sslmode=disable"
      }
    }
  }
}
```

For detailed installation, operational procedures, and workflow examples, see [PLAYBOOK.md](PLAYBOOK.md).

---

## Capability Status

### Stable

- Go semantic analysis
- Compiler-backed type resolution
- Deterministic entity identity
- Semantic relationships
- Repository and workspace modelling
- Evidence provenance
- Semantic snapshots
- Semantic diff
- Impact analysis
- Impact diff
- Lineage
- Cryptographic state verification (Merkle ledger)
- Agent checkpoints and CAS handoffs
- Global workspace search
- Progressive architecture exploration
- CLI workflows
- Dashboard exploration

### Early Testing

- Multi-repository intelligence
- OpenTelemetry ingestion
- Runtime observations
- Static ↔ runtime correlation
- Contradiction verification and quarantine
- Runtime evidence views
- MCP integration
- Garuda IDE
- Grounding benchmark harness
- Agent-oriented workflows
- `garuda ci`
- `garuda judge`
- PR governance and policy automation

These capabilities should be treated as **early testing rather than universal production guarantees**.

### Launching Soon

- Richer runtime verification
- Deeper MCP workflows
- Expanded evidence exploration
- Broader grounding benchmarks
- Production-oriented telemetry workflows
- Additional developer workflow integrations

### Longer-term Direction

- Company-scale software graphs
- Multi-language semantic intelligence
- Richer runtime correlation
- Stronger architectural governance
- AI-grounded software operations
- Business-state integrity
- Autonomous verification and remediation

These are development directions, not current production guarantees.

---

## Design Principles

1. **Evidence over inference** — Prefer inspectable evidence over opaque conclusions.
2. **Unknown stays unknown** — Missing evidence must not silently become certainty.
3. **Identity is deterministic** — The same semantic entity should remain identifiable across analysis runs.
4. **State is auditable** — Important state transitions should leave an inspectable trail.
5. **Architecture should be navigable** — Users should move from system-level topology to specific entities and evidence.
6. **Static and runtime knowledge remain distinct** — Complementary signals, not interchangeable facts.
7. **AI should consume structured state** — Agents should not reconstruct an entire repository from raw text whenever reusable semantic state already exists.
8. **Correctness before scale** — Validated semantics and evidence before expanding language and deployment scope.

---

## What Garuda Is Not

Garuda is not:

- another text search engine
- another code linter
- another generic dependency graph
- another observability platform
- another documentation generator
- another generic vector database
- an LLM that guesses how a repository works

Garuda is intended to connect these layers through a structured semantic and evidence substrate.

---

## Security

Security, provenance, and deterministic state are first-class engineering concerns.

- **Authentication**: All mutation endpoints are guarded by Ed25519 JWTs.
- **Integrity**: State changes are anchored to a verifiable Merkle root.
- **Runtime**: The API executes in a non-root `65532:65532` minimal scratch container.
- **Idempotency**: All decision commits pass through `idempotency_keys` to prevent duplicate execution.

For vulnerability reporting, see [SECURITY.md](SECURITY.md).

Cryptographic mechanisms provide tamper-evident state and verification. They do **not** replace credential security, database security, access controls, key management, or operational security practices.

---

## Documentation

| Document                                                            | Purpose                                                         |
| ------------------------------------------------------------------- | --------------------------------------------------------------- |
| [Playbook](PLAYBOOK.md)                                             | Installation, commands, workflows, telemetry, MCP and IDE usage |
| [Evidence](EVIDENCE.md)                                             | Validation results, benchmarks, screenshots and methodology     |
| [Benchmarks (v1.0.0)](BENCHMARKS_AND_VERIFICATION_v1.0.0.md)        | System verification & benchmark report                          |
| [Architecture](docs/)                                               | Technical architecture and design                               |
| [Walkthrough](docs/WALKTHROUGH.md)                                  | Dashboard and IDE workflow tour                                 |
| [Security](SECURITY.md)                                             | Security model and vulnerability reporting                      |
| [Contributing](CONTRIBUTING.md)                                     | Contribution and development workflow                           |
| [Changelog](CHANGELOG.md)                                           | Project history and implementation changes                      |

---

## Repository Structure

```text
garuda/
│
├── README.md
├── PLAYBOOK.md
├── EVIDENCE.md
├── BENCHMARKS_AND_VERIFICATION_v1.0.0.md
├── SECURITY.md
├── CONTRIBUTING.md
├── CHANGELOG.md
├── LICENSE
│
├── assets/
│   ├── garuda-logo.png
│   └── screenshots/
│
├── cmd/
├── internal/
├── migrations/
├── garuda-bench/
├── test/
├── scripts/
└── docs/
```

---

## Contributing

Garuda is being built in the open. Relevant areas include compiler tooling, Go analysis, developer infrastructure, semantic graphs, runtime verification, OpenTelemetry, MCP, AI coding agents, software architecture, provenance, cryptographic state, and developer experience.

Start with [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

Garuda is released under the Apache License 2.0. See [LICENSE](LICENSE).

---

## Current Scope and Evidence Boundary

Garuda's current semantic engine is **Go-specific**. The implemented static analysis path uses Go AST analysis, `go/types`, deterministic symbol identity, and Go-oriented semantic resolution.

Multi-language support such as TypeScript, Rust, Python, and Java is future direction — **not current v1.0.0 functionality**.

Runtime verification is also bounded by telemetry coverage and the observations available to Garuda. A runtime path that has not been observed should remain `UNVERIFIED`, not be interpreted as dead or incorrect.

---

## Technical Disclaimer

This README describes Garuda according to the current implementation and documented validation runs.

**Shared Workspace** — Multiple developers and AI agents share the same semantic state only when they are configured against the same Garuda workspace and PostgreSQL database. Separate local databases are not automatically synchronized.

**Near-real-time behaviour** — Telemetry admission is asynchronous and was measured at approximately `1.8 ms p95` under the tested workload. Verification runs on asynchronous 10-second epochs. Full dashboard and consensus propagation typically completed within `1–2 minutes` in the tested environment. These are observed validation characteristics, not universal production SLAs.

**Performance and benchmark results** — Precision, token reduction, throughput, and latency measurements derive from specific controlled environments. Results can vary with repository structure, hardware, infrastructure, telemetry coverage, network conditions, workload, AI model, and agent behaviour.

**AI hallucination boundary** — Garuda is designed to reduce unsupported reasoning and repeated context reconstruction. The GAP-20 benchmark demonstrated **0% structural hallucination** for the evaluated benchmark tasks involving fabricated functions, receivers, and symbols. This does not claim elimination of all forms of AI hallucination.

**Cryptographic integrity** — Cryptographic mechanisms provide tamper-evident state and verification. They do not replace credential security, database security, access controls, key management, or operational security practices.

**Regulatory references** — Garuda may provide technical capabilities relevant to governance and compliance workflows. Whether a deployment satisfies the EU AI Act or another legal or regulatory requirement depends on the deployment, organizational processes, risk classification, and applicable law. Garuda is not legal advice or a legal certification.

**Reproducibility** — For reproducible validation, run the included benchmark and review the evidence artifacts in [EVIDENCE.md](EVIDENCE.md) and [BENCHMARKS_AND_VERIFICATION_v1.0.0.md](BENCHMARKS_AND_VERIFICATION_v1.0.0.md).

---

<p align="center">
  <strong>Garuda</strong><br>
  Analyze | Verify | Govern | Collaborate
</p>

<p align="center">
  <sub>
    Built as a shared intelligence workspace.<br>
    If the system can provide evidence, don't replace it with a guess.<br>
    If the system can verify, don't rely on assumption alone.<br>
    If the context already exists, don't make every human or AI agent rediscover it.
  </sub>
</p>

<p align="center">
  <sub>
    <strong>Version:</strong> v1.0.0 ·
    <strong>License:</strong> Apache 2.0
  </sub>
</p>
```

