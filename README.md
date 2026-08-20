# 🦅 Garuda — Evidence-Backed Software Intelligence

[![Release](https://img.shields.io/github/v/release/myshra777-ai/garuda?style=flat-square&color=blue)](https://github.com/myshra777-ai/garuda/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/myshra777-ai/garuda/garuda-ci.yml?branch=main&style=flat-square&label=CI)](https://github.com/myshra777-ai/garuda/actions)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg?style=flat-square)](LICENSE)
[![Benchmarks](https://img.shields.io/badge/Benchmarks-100%25-brightgreen?style=flat-square)](https://github.com/myshra777-ai/garuda#-ast-semantic-benchmark-suite)

**Understand your codebase as a connected, inspectable, and verifiable system.**

Garuda is a **deterministic** code intelligence and governance platform. It parses Go source code into a **canonical AST**, builds a **structured semantic graph**, anchors every relationship to **cryptographic evidence** (commit SHA, file, line range, content hash), and provides high‑speed CLI tools for **blast‑radius impact analysis**, **architectural validation**, and **agentic code quality**.

> **Code → Semantics → Relationships → Evidence → Understanding**

> 🧠 **Building the Company Brain** – a continuously updated, cryptographically verifiable semantic model of your entire software ecosystem.

---

## 📋 Table of Contents

- [The Core Philosophy](#the-core-philosophy)
- [The Epistemic Triad](#the-epistemic-triad)
- [Capabilities Matrix](#capabilities-matrix)
- [AST Semantic Benchmark Suite](#ast-semantic-benchmark-suite)
- [System Architecture](#system-architecture)
- [Blast Radius Impact Engine](#blast-radius-impact-engine)
- [The 10 Immutable Engineering Laws](#the-10-immutable-engineering-laws)
- [Installation & Quick Start](#installation--quick-start)
- [Core Capabilities & CLI](#core-capabilities--cli)
- [Cryptographic Evidence Contract](#cryptographic-evidence-contract)
- [Multi-Repository Scaling Model](#multi-repository-scaling-model)
- [Roadmap & What's Next](#roadmap--whats-next)
- [Security & Trust](#security--trust)
- [Development & Contributing](#development--contributing)
- [License](#license)

---

## The Core Philosophy

Modern codebases are not collections of files – they are interconnected webs of packages, structs, interfaces, methods, and contracts.

**Garuda is built on a single premise: Build structured truth first. Put generative intelligence on top of it later.**

Instead of relying on an LLM to guess what your codebase does via probabilistic RAG, Garuda:

- Parses the actual **Go AST** and resolves symbols with `go/types`.
- Extracts **deterministic facts** – entities, relationships, and evidence.
- Anchors every claim to a **commit SHA, file path, line range, and content hash**.
- Persists the graph in PostgreSQL with **Merkle‑backed integrity**.

### What Garuda Is — and Isn't

| Garuda Is | Garuda Is Not |
|-----------|---------------|
| 🧠 A deterministic semantic graph of your code | ❌ A generic LLM wrapper or chat agent |
| 🔍 Evidence‑aware, linking claims to exact source lines | ❌ A probabilistic vector‑search tool |
| 🔐 A cryptographically verified analysis ledger | ❌ A disposable, memoryless linter |
| 🚧 A foundation for cross‑repo governance | ❌ An autonomous code‑rewriting bot |
| 📊 Blast‑radius impact analysis | ❌ A code search tool |
| 🧠 A foundation for the Company Brain | ❌ A complete company‑wide brain today |

---

## The Epistemic Triad

Garuda rigidly separates three classes of knowledge. **Observations are never collapsed into Decisions.**

```mermaid
flowchart LR
    Obs[fa:fa-eye Observation<br/>'A calls B']:::obs --> Claim[fa:fa-link Claim<br/>Backed by Evidence]:::claim
    Inf[fa:fa-lightbulb Inference<br/>'A likely depends on B']:::inf --> Claim
    Claim --> Conflict{fa:fa-bolt Contradiction Engine}
    Dec[fa:fa-gavel Decision<br/>'A must use C']:::gov --> Conflict
    Pol[fa:fa-shield Policy<br/>'No external DBs']:::gov --> Conflict
    Conflict --> |Contradiction| Alert[fa:fa-bell Quarantine/Alert]
    Conflict --> |Resolution| NewDec[fa:fa-pen New Decision Revision]
    classDef obs fill:#1e40af,color:#fff;
    classDef inf fill:#6b21a8,color:#fff;
    classDef gov fill:#991b1b,color:#fff;
    classDef claim fill:#065f46,color:#fff;
```

| Category | Definition | Example | Source |
|----------|------------|---------|--------|
| **Observation** | Directly extracted from source | `PaymentService IMPORTS Postgres` | AST / Type Checker |
| **Inference** | Derived by analysis/model logic | `CheckoutService CALLS PaymentService (conf: 0.87)` | Graph Traversal |
| **Decision** | Intentional organizational choice | `Production DB MUST BE Postgres` | Human / Policy |

> **Observation ≠ Inference ≠ Decision**. A code analyzer observes that a service uses PostgreSQL. A human or policy may decide that production must use PostgreSQL. Those are distinct epistemic objects and must remain separate.

---

## Capabilities Matrix

Garuda’s capabilities follow the V5 roadmap doctrine: **evidence before confidence, stable before experimental**.

> 📄 **Complete Capabilities Reference**: See [`docs/CAPABILITIES.md`](docs/CAPABILITIES.md) for the full machine‑generated matrix.

### ✅ Stable — Production Ready

| Capability | Description | Command |
|------------|-------------|---------|
| **Semantic Analysis** | Extracts entities (structs, functions, interfaces, packages) with line‑level evidence | `garuda analyze` |
| **Canonical UUIDv5 Identity** | Deterministic identity hashing across structs, methods, and functions | Core AST Engine |
| **Truth Benchmark Verification** | Automated validation against 20 truth fixtures | `go test ./...` |
| **Entity Inspection** | View entity details: package, file, kind, fields, methods, relationships | `garuda inspect` |
| **Graph Visualization** | Interactive HTML graph with D3.js | `garuda graph` |
| **Snapshot Diff** | Compare two semantic snapshots with breaking change detection | `garuda diff` |
| **Blast Radius Impact** | BFS traversal to find all consumers with severity scoring | `garuda impact` |
| **Diff Impact** | Compare impact between two snapshots | `garuda impact-diff` |
| **Immutable Ledger** | Merkle‑backed audit trail with cryptographic integrity | `garuda verify` |
| **Workspace Management** | Create, list, delete workspaces | `garuda workspace` |
| **Repository Management** | Add, list, enable, disable repositories | `garuda repo` |
| **Multi-Repo Sync** | Analyze all enabled repositories in a workspace | `garuda workspace sync` |

### 🧪 Ongoing — Active Development

| Capability | Description | Status |
|------------|-------------|--------|
| **Cross-Repo Impact** | Impact analysis across repository boundaries | Alpha |
| **Contract Extraction** | HTTP routes, SQL migrations, API schemas | Alpha |
| **Schema Discovery** | Unify Go structs, JSON tags, OpenAPI, SQL | In Progress |
| **Code Quality (Ponytail)** | Dead code, duplicates, stdlib alternatives | Alpha |
| **Governance Judge** | Breaking change detection with blocking decisions | Alpha |

### 📋 Planned — Future

| Capability | Description | Phase |
|------------|-------------|-------|
| **CI Integration** | GitHub Action for PR impact/quality comments | Phase 5 |
| **Company Graph** | Cross‑repository semantic graph | Phase 4 |
| **Policy Enforcement** | Evaluate code against organizational policies | Phase 6 |
| **Runtime Intelligence** | Combine static analysis with runtime evidence | Phase 6 |
| **Business State Integrity** | Idempotency keys, business rule enforcement | Phase 7 |

---

## 🔬 AST Semantic Benchmark Suite

Garuda’s semantic extraction is evaluated against a **20‑fixture synthetic and real‑world ground‑truth test suite** (`test/benchmark/truth_fixtures`). Every release is gated on deterministic precision and recall.

```mermaid
flowchart LR
    subgraph Suite["Truth Benchmark Suite"]
        F1[001-basic] --> V[Verification Engine]
        F2[002-method-identity] --> V
        F3[003-interface] --> V
        F4[004-generics] --> V
        F5[005-alias] --> V
        F6[006-embedding] --> V
        F7[007-cross-package] --> V
        F8[008-cross-module] --> V
        F9[009-same-symbol-names] --> V
        F10[010-api-contract] --> V
        F11[011-breaking-change] --> V
        F12[012-non-breaking-change] --> V
        F13[013-consumer-impact] --> V
        F14[014-rename] --> V
        F15[015-delete] --> V
        F16[016-multi-repo] --> V
        F17[017-evidence] --> V
        F18[018-determinism] --> V
        F19[019-noise] --> V
        F20[020-large-ish] --> V
    end
    V --> R[Report]
    R --> G[Gate: 100% Precision/Recall]
```

### Snapshot Extraction Metrics

| Metric | Count |
| :--- | :--- |
| **Parsed Files** | `214` |
| **Packages** | `43` |
| **Discovered Structs** | `322` |
| **Discovered Interfaces** | `25` |
| **Functions & Methods** | `389` |
| **Total Struct Fields** | `1,581` |

### Truth Fixture Coverage & Results

| Benchmark Target | Fixtures | Invariant Verified | Result |
|---|---|---|---|
| **Package Resolution** | 20 | Fully qualified package path resolution | ✅ 100% Accuracy |
| **Entity Fingerprinting** | 20 | Receiver types, structs, interfaces, methods | ✅ 100% Match |
| **Call‑Graph Edges** | 20 | Deterministic `CALLS` relationship extraction | ✅ 100% Precision/Recall |
| **Import Graph** | 20 | Cross‑package and internal module import tracing | ✅ 100% Recall |
| **Canonical UUIDv5** | 20 | Stable identity across file moves and renames | ✅ Deterministic |
| **Breaking Change Detection** | 20 | Semantic diff classification | ✅ 100% Accuracy |
| **Impact Analysis** | 20 | Consumer tracing & severity | ✅ 100% Precision |

**Aggregate Benchmark Scorecard (20 fixtures):**

```
Overall Entity Precision:       100.0%
Overall Entity Recall:          100.0%
Overall Relationship Precision: 100.0%
Overall Relationship Recall:    100.0%
```

### Run the Benchmark Suite

```bash
go test -v -run TestTruthBenchmarks ./test/benchmark/...
```

---

## System Architecture

```mermaid
graph TD
    subgraph Source["Source Layer"]
        Git[Git Repository] --> Checkout[Checkout Commit]
        Checkout --> AST[Go AST + go/types Analyzer]
    end
    subgraph Semantic["Semantic Core"]
        AST --> |Extracts| E[Entities<br/>Packages, Structs, APIs]
        AST --> |Maps| R[Relationships<br/>Calls, Imports, Implements]
        R --> |BFS Traversal| Impact[Blast Radius Engine]
        E --> Graph[Semantic Graph]
        R --> Graph
    end
    subgraph Trust["Trust Layer"]
        E --> Hash[SHA-256 Hashing]
        R --> Hash
        Hash --> |Anchors to| Ev[Evidence Contract<br/>Line, File, Commit, Hash]
        Ev --> Postgres[(PostgreSQL Graph)]
        Ev --> JSONB[(Immutable Artifacts)]
    end
    subgraph Experience["Experience Layer"]
        Postgres --> CLI{Garuda CLI}
        Postgres --> CI[CI/PR Bots]
        Postgres --> GUI[Web Graph]
        Impact --> CLI
        Impact --> CI
        Graph --> GUI
    end
```

---

## Blast Radius Impact Engine

When you change an entity, Garuda performs a **BFS traversal** of the dependency graph to find all downstream consumers. Each impacted entity is classified by severity:

| Severity | Condition |
|----------|-----------|
| **CRITICAL** | Depth 1 public contract breach (API endpoint, SQL schema) |
| **HIGH** | Depth 1 non‑test consumer |
| **MEDIUM** | Depth 2 public route consumer |
| **LOW** | Test files, confidence < 0.7, or depth ≥ 3 |

```bash
$ garuda impact --target e8b9c2a1-0f4e-4b2a-9e12-4c9a8f23b123 --depth 3

Target: internal/store.Migrate (Function)
Blast Radius: 14 impacted entities across 3 packages (Max Depth: 3)

[CRITICAL] (Depth 1) cmd/garuda/main.go:42 - main()
[HIGH]     (Depth 1) internal/server/handler.go:108 - HandleWorkspaceSync()
[MEDIUM]   (Depth 2) internal/router/routes.go:54 - RegisterAdminRoutes()
[LOW]      (Depth 3) internal/store/migrate_test.go:19 - TestMigrateIdempotency()

Confidence Score: 0.98 | Evidence Hash: 4f8b9e... (SHA-256)
```

---

## The 10 Immutable Engineering Laws

| # | Law | Meaning |
|---|-----|---------|
| 1 | **Evidence before confidence** | No claim exists without a source pointer |
| 2 | **Epistemic separation** | Observations, inferences, decisions, and policies remain type‑distinct |
| 3 | **Immutability** | Commits are immutable; analysis artifacts are versioned |
| 4 | **Stable Identity** | Entities require deterministic canonical IDs before cross‑repo reasoning |
| 5 | **Least Privilege Traversal** | Authorization applies to graph traversal, not just storage reads |
| 6 | **Explicit Uncertainty** | Partial extraction must be visible, never silently upgraded to truth |
| 7 | **No AI Overwrites** | LLMs may summarize, but they cannot overwrite deterministic facts |
| 8 | **Verification First** | An unverified claim cannot be promoted to an enforced invariant |
| 9 | **Safe Failures** | A failed analysis never destroys the last known‑good snapshot |
| 10 | **Evaluation Gates** | Production capabilities require strict benchmark evaluation |

---

## Installation & Quick Start

### Requirements

- Go 1.22+
- PostgreSQL 16+
- Docker & Docker Compose (optional, for local testing)

### Download Pre‑built Binary

```bash
# Linux AMD64
curl -sL https://github.com/myshra777-ai/garuda/releases/latest/download/garuda-linux-amd64 -o garuda
chmod +x garuda
sudo mv garuda /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda
go build -o bin/garuda cmd/garuda/main.go
export PATH=$PATH:$(pwd)/bin
```

### Quick Workflow

```bash
# 1. Start PostgreSQL (Docker)
docker run --name garuda-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=garuda -p 5432:5432 -d postgres:16-alpine

# 2. Run migrations (000-041)
DATABASE_URL="postgres://postgres:postgres@localhost:5432/garuda?sslmode=disable" go run cmd/migrate/main.go

# 3. Set environment
export GARUDA_TENANT_ID=$(uuidgen)
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/garuda?sslmode=disable"
export GARUDA_WORKSPACE=my-workspace

# 4. Create workspace & analyze
bin/garuda workspace create my-workspace
bin/garuda analyze . --save

# 5. Explore
bin/garuda inspect User
bin/garuda graph my-workspace --open
```

### Configuration Variables

| Variable | Description | Default |
|---|---|---|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://test:test@localhost:5433/garuda_test?sslmode=disable` |
| `GARUDA_TENANT_ID` | Tenant UUID for multi‑tenant isolation | *Required for DB writes* |
| `GARUDA_WORKSPACE` | Active workspace name | `default` |

---

## Core Capabilities & CLI

```bash
# Analysis & diffing
garuda analyze . -o v1.json
garuda diff v1.json v2.json --json

# Blast radius impact
garuda impact --workspace <uuid> --target <entity-id> --depth 5

# Code quality (Ponytail)
garuda ponytail . --json

# Governance (Judge)
garuda judge v1.json v2.json --block

# Workspace & repository management
garuda workspace create my-workspace
garuda repo add my-workspace https://github.com/org/repo
garuda workspace sync my-workspace

# Trust & integrity
garuda verify
garuda explain <decision-id>
garuda entities
```

---

## Cryptographic Evidence Contract

Every claim in Garuda carries the following evidence contract:

```json
{
  "repository_id": "uuid",
  "commit_sha": "4f7c8a9b...",
  "file_path": "internal/payment/service.go",
  "symbol": "ProcessPayment",
  "line_start": 42,
  "line_end": 68,
  "content_snippet": "func (s *Service) ProcessPayment(...)",
  "content_hash": "sha256:4f8b9e2a...",
  "analyzer": "garuda",
  "analyzer_version": "v1.0"
}
```

You can walk the integrity chain from a claim back to the exact source line, commit, and artifact:

```text
Source file/commit → Analyzer run → Snapshot artifact → Entity/relationship extraction
→ Claim/observation → Evidence reference → User/API/CI answer
```

---

## Multi-Repository Scaling Model

Garuda follows a disciplined **1 → 10 → 25** scaling model.

```mermaid
flowchart LR
    subgraph Gate1["Gate 1: Single Repository"]
        G1_1[Deterministic Extraction] --> G1_2[Evidence Accuracy >95%]
        G1_2 --> G1_3[Developer AHA Moment]
    end
    subgraph Gate2["Gate 2: 10 Repositories"]
        G2_1[Workspace Sync Durable] --> G2_2[Cross-Repo Edges Precise]
        G2_2 --> G2_3[Permissions Enforced]
    end
    subgraph Gate3["Gate 3: 25 Repositories"]
        G3_1[Query Latency Practical] --> G3_2[CI Impact Actionable]
        G3_2 --> G3_3[Storage Economically Viable]
    end
    Gate1 --> Gate2 --> Gate3
```

### Repository Lifecycle

| State | Meaning |
|-------|---------|
| **Registered** | Metadata exists; no analysis yet |
| **Connected** | Garuda can access the repository |
| **Analyzing** | Versioned analysis job is running |
| **Analyzed** | Valid snapshot exists for a commit |
| **Stale** | Repository changed since last successful analysis |
| **Failed** | Latest analysis failed; last good snapshot remains |

---

## Roadmap & What's Next

| Phase | Focus | Status |
|-------|-------|--------|
| **P0** | Trust Foundation & Merkle Verification | ✅ Complete |
| **P1** | Semantic Core & Canonical UUIDv5 Identity | ✅ Complete |
| **P2** | Go AST Analyzer & Truth Benchmarks (20 Fixtures) | ✅ Complete |
| **P3** | CLI, Artifacts & Automated Docgen Pipeline | ✅ Complete |
| **P4** | Multi-Repo Sync & Cross-Repo Edges | 🧪 Alpha |
| **P5** | CI Integration (GitHub Action PR Impact) | 📋 Planned |
| **P6** | Governance & Policy Enforcement | 📋 Planned |
| **P7** | Business State Integrity | 📋 Planned |

**Next deliverables:**
- Expand BFS blast radius across repository boundaries.
- Deep contract extraction (HTTP, SQL, OpenAPI).
- GitHub Action for automated PR impact comments.
- Enhanced interactive graph explorer.

---

## Security & Trust

| Boundary | Required Guarantee |
|----------|-------------------|
| **Tenant** | No cross‑tenant reads or writes |
| **Workspace** | Only authorized workspace members |
| **Repository** | Repository‑level access policy enforced |
| **Evidence** | Unauthorized graph paths cannot reveal protected evidence |
| **MCP/API** | Tool calls inherit authorization context |
| **Agents** | Agent identity cannot bypass policy |
| **Audit** | Important mutations are traceable and integrity‑verifiable |

---

## Development & Contributing

Contributions are evaluated against the **10 Immutable Engineering Laws**.

### Adding a New Analyzer Feature?

1. ✅ Must be **deterministic** – avoid regex where AST parsing works.
2. ✅ Must retain **evidence** – never emit an edge without line‑level source tracking.
3. ✅ Must pass the **benchmark corpus** – no regression in precision/recall.

### Local Development

```bash
# Spin up testing DB
docker-compose up -d postgres

# Run tests
DATABASE_URL="postgres://test:test@localhost:5433/garuda_test" go test ./...

# Build
go build -o bin/garuda cmd/garuda/main.go
```

### Project Structure

```
garuda/
├── cmd/garuda/          # CLI entry points
├── internal/
│   ├── analyzer/        # Go AST parsing + UUIDv5 canonical identity
│   ├── store/           # PostgreSQL migrations and atomic transactions
│   ├── impact/          # BFS blast‑radius engine
│   ├── graph/           # Visualization
│   └── engine/          # Governance judge and epistemic classification
├── test/benchmark/      # 20‑fixture truth benchmark suite
├── migrations/          # Idempotent PostgreSQL migrations (000-041)
├── scripts/             # Documentation generation scripts
└── docs/                # Generated capabilities matrix
```

---

## License

Copyright 2026 Rohit Mishra.

Garuda is licensed under the [Apache License 2.0](LICENSE).

---

## 🦅 Garuda

**Understand the code.**  
**See the connections.**  
**Trace the evidence.**  
**Build the Company Brain.**

---

<p align="center">
  <strong>Code → Semantics → Graph → Evidence → Intelligence</strong>
</p>