# 🦅 Garuda — Evidence-Backed Software Intelligence

[![Release](https://img.shields.io/github/v/release/myshra777-ai/garuda?style=flat-square&color=blue)](https://github.com/myshra777-ai/garuda/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/myshra777-ai/garuda/garuda-ci.yml?branch=main&style=flat-square&label=CI)](https://github.com/myshra777-ai/garuda/actions)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg?style=flat-square)](LICENSE)

**Understand your codebase as a connected, inspectable, and verifiable system.**

Garuda is a deterministic code intelligence and governance platform. It analyzes Go repositories to build a structured semantic graph of your software, preserves cryptographically hashed evidence behind every claim, and provides high-speed CLI tools for blast-radius impact analysis, architectural validation, and agentic code quality.

> **Code → Semantics → Relationships → Evidence → Understanding**

> 🧠 **Building the Company Brain**: Garuda is the foundation for a continuously updated, cryptographically verifiable semantic model of your entire software ecosystem.

---

## 📋 Table of Contents

- [The Core Philosophy](#the-core-philosophy)
- [Company Brain — The North Star](#company-brain--the-north-star)
- [Capabilities Matrix](#capabilities-matrix)
- [AST Semantic Benchmark Suite](#ast-semantic-benchmark-suite)
- [System Architecture](#system-architecture)
- [The Epistemic Model](#the-epistemic-model)
- [The 10 Immutable Engineering Laws](#the-10-immutable-engineering-laws)
- [Installation & Quick Start](#installation--quick-start)
- [Core Capabilities & CLI](#core-capabilities--cli)
- [Multi-Repository Scaling Model](#multi-repository-scaling-model)
- [Roadmap & What's Next](#roadmap--whats-next)
- [Security & Trust](#security--trust)
- [Development & Contributing](#development--contributing)
- [License](#license)

---

## The Core Philosophy

Modern codebases are not just collections of files. They are interconnected webs of packages, structs, interfaces, methods, and API contracts.

**Garuda is built on a single premise: Build structured truth first. Put generative intelligence on top of it later.**

Instead of relying on an LLM to guess what your codebase does via probabilistic search (RAG), Garuda parses the actual AST. It extracts deterministic facts, maps adjacency relationships with O(1) lookup, and anchors every claim to an exact line of code and commit hash.

### What Garuda Is — and Isn't

| Garuda Is | Garuda Is Not |
|-----------|---------------|
| 🧠 A deterministic semantic graph of your code | ❌ A generic LLM wrapper or chat agent |
| 🔍 Evidence-aware, linking claims to exact source lines | ❌ A probabilistic vector-search tool |
| 🔐 A cryptographically verified analysis ledger | ❌ A disposable, memoryless linter |
| 🚧 A foundation for cross-repo governance | ❌ An autonomous code-rewriting bot |
| 📊 Blast-radius impact analysis | ❌ A code search tool |
| 🧠 A foundation for the Company Brain | ❌ A complete company-wide brain today |

---

## 🧠 Company Brain — The North Star

Garuda's long-term destination is the **Company Brain**: a continuously updated, cryptographically verifiable semantic model of your entire software ecosystem. It's the foundation for answering questions like:

- *"What services depend on this API?"*
- *"Which repositories implement this business capability?"*
- *"What breaks if we change this schema?"*
- *"Does our code match our architecture decisions?"*

### Evolution to Company Brain

```mermaid
flowchart LR
    subgraph Current["✅ Current: Single Repo Intelligence"]
        C1[Semantic Graph] --> C2[Blast Radius] --> C3[Evidence Backing]
    end
    
    subgraph Growing["🧪 Growing: Multi-Repo Foundation"]
        G1[Workspace Sync] --> G2[Cross-Repo Detection] --> G3[Company Graph Alpha]
    end
    
    subgraph Future["📋 Future: Company Brain"]
        F1[Cross-Repo Impact] --> F2[Policy Enforcement] --> F3[AI Reasoning]
    end
    
    Current --> Growing --> Future
```

### What's Already Built (The Foundation)

| Component | Status | What It Does |
|-----------|--------|--------------|
| **Semantic Graph** | ✅ Stable | Entities (structs, functions, APIs) + relationships (calls, imports, dependencies) |
| **Canonical UUIDv5 Identity** | ✅ Stable | Deterministic identity across commits, renames, and package moves |
| **Blast Radius Impact** | ✅ Stable | BFS traversal to find who depends on what |
| **Multi-Repository Sync** | ✅ Stable | Analyze 10-25 repositories in one workspace |
| **Evidence Backing** | ✅ Stable | Every claim traces to source lines + commit SHA |
| **Truth Benchmark Suite** | ✅ Stable | 10-fixture semantic extraction verification |
| **Cross-Repo Detection** | 🧪 Alpha | Detect imports across repository boundaries |

### What's Growing (Ongoing)

| Component | Status | What It's Becoming |
|-----------|--------|-------------------|
| **Cross-Repo Impact** | 🧪 Alpha | Impact analysis across repository boundaries |
| **Contract Extraction** | 🧪 Alpha | HTTP routes, SQL migrations, API schemas |
| **Company Graph** | 📋 Planned | Unified graph of all repositories in an organization |

### What Garuda Can Answer Today

| Question | Status |
|----------|--------|
| What entities exist in this repository? | ✅ Yes |
| What calls/depends on entity X? | ✅ Yes |
| What breaks if I change entity X? | ✅ Yes |
| Which repositories import package Y? | 🧪 Alpha |
| How does service A depend on service B? | 📋 Planned |
| Does this change violate our architecture policy? | 📋 Planned |

> **The Company Brain is being built incrementally. We're starting with one repository at a time, proving correctness before scaling.**

---

## Capabilities Matrix

Garuda's capabilities follow the V4 roadmap doctrine: **evidence before confidence, stable before experimental**.

> 📄 **Complete Capabilities Reference**: See [`docs/CAPABILITIES.md`](docs/CAPABILITIES.md) for the full machine-generated matrix of analyzers, invariants, and CLI commands.

### ✅ Stable — Production Ready

| Capability | Description | Command |
|------------|-------------|---------|
| **Semantic Analysis** | Extracts entities (structs, functions, interfaces, packages) with line-level evidence | `garuda analyze` |
| **Canonical Identity (UUIDv5)** | Deterministic identity hashing across structs, methods, and functions | Core AST Engine |
| **Truth Benchmark Verification** | Automated validation against 10 semantic truth fixtures | `go test ./...` |
| **Entity Inspection** | View entity details: package, file, kind, fields, methods, relationships | `garuda inspect` |
| **Graph Visualization** | Interactive HTML graph of workspace architecture with D3.js | `garuda graph` |
| **Snapshot Diff** | Compare two semantic snapshots with breaking change detection | `garuda diff` |
| **Blast Radius Impact** | BFS traversal to find all consumers of an entity with severity scoring | `garuda impact` |
| **Diff Impact** | Compare impact between two snapshots | `garuda impact-diff` |
| **Immutable Ledger** | Merkle tree-backed audit trail with cryptographic integrity | `garuda verify` |
| **Workspace Management** | Create, list, delete workspaces | `garuda workspace` |
| **Repository Management** | Add, list, enable, disable repositories | `garuda repo` |
| **Multi-Repo Sync** | Analyze all enabled repositories in a workspace | `garuda workspace sync` |

### 🧪 Ongoing — Active Development

| Capability | Description | Status |
|------------|-------------|--------|
| **Cross-Repo Impact** | Impact analysis across repository boundaries | Alpha |
| **Contract Extraction** | HTTP routes, API schemas, SQL migrations | Alpha |
| **Schema Discovery** | Unify Go structs, JSON tags, OpenAPI, SQL | In Progress |
| **Code Quality (Ponytail)** | Dead code detection, duplicates, standard-library alternatives | Alpha |
| **Governance Judge** | Breaking change detection with blocking decisions | Alpha |
| **Topology Management** | Multi-agent task execution and handoff lineage | Alpha |

### 📋 Planned — Future

| Capability | Description | Phase |
|------------|-------------|-------|
| **CI Integration** | GitHub Action for PR impact/quality comments | Phase 5 |
| **Company Graph** | Cross-repository semantic graph | Phase 4 |
| **Policy Enforcement** | Evaluate code against organizational policies | Phase 6 |
| **Runtime Intelligence** | Combine static analysis with runtime evidence | Phase 6 |
| **Business State Integrity** | Idempotency keys, business rule enforcement | Phase 7 |

---

## 🔬 AST Semantic Benchmark Suite

Garuda's semantic extraction is evaluated against a **10-fixture synthetic and real-world ground-truth test suite** (`test/benchmark/truth_fixtures`). Every release is gated on 100% deterministic precision across semantic AST extraction invariants.

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
    end
    
    V --> R[Report]
    R --> G[Gate: 100% Precision]
```

### Snapshot Extraction Metrics

| Metric | Count |
| :--- | :--- |
| **Parsed Files** | `177` |
| **Packages** | `32` |
| **Discovered Structs** | `279` |
| **Discovered Interfaces** | `20` |
| **Functions & Methods** | `327` |
| **Total Struct Fields** | `1401` |

### Feature Verification & Status

| Benchmark Target | Fixtures | Invariant Verified | Metric |
|---|---|---|---|
| **Package Resolution** | 10 | Fully qualified package path resolution | ✅ 100% Accuracy |
| **Entity Fingerprinting** | 10 | Receiver types, structs, interfaces, and methods | ✅ 100% Match |
| **Call-Graph Edges** | 10 | Deterministic `CALLS` relationship extraction | ✅ 0 False Edges |
| **Import Graph** | 10 | Cross-package and internal module import tracing | ✅ 100% Recall |
| **Canonical UUIDv5** | 10 | Stable identity across file moves and renames | ✅ Deterministic |

### Truth Fixture Coverage

| Fixture | Target Scenario | Key Invariant Tested |
|---|---|---|
| `001-basic` | Standard struct & function declarations | Canonical UUIDv5 hashing |
| `002-method-identity` | Value vs. pointer receiver methods | Disambiguated method identity |
| `003-interface` | Explicit & implicit interface implementation | `IMPLEMENTS` edge emission |
| `004-generics` | Parameterized types & generic constraints | Type parameter erasure consistency |
| `005-alias` | Type definitions vs. type aliases (`type A = B`) | Target identity resolution |
| `006-embedding` | Struct & interface field embedding | Promoted field/method resolution |
| `007-cross-package` | Internal multi-package dependencies | Fully qualified package path lookup |
| `008-cross-module` | External Go module dependencies | Module boundary edge classification |
| `009-same-symbol-names` | Identical identifier names across packages | Zero-collision UUIDv5 guarantee |
| `010-api-contract` | HTTP handler routes & SQL query strings | Contract edge extraction |

### Run the Benchmark Suite

```bash
go test -v -run TestTruthBenchmarks ./test/benchmark/...
```

---

## System Architecture

### Complete Pipeline

```mermaid
graph TD
    subgraph Source["Source Layer"]
        Git[Git Repository] --> Checkout[Checkout Commit]
        Checkout --> AST[Go AST Analyzer]
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
        Hash --> |Anchors to| Ev[Evidence Contract<br/>Line, File, Commit]
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

### Data Flow

```mermaid
flowchart LR
    subgraph Input
        Code[Source Code]
    end
    
    subgraph Analysis
        AST[AST Parser]
        Types[Type Resolution]
        Graph[Graph Builder]
    end
    
    subgraph Storage
        Entities[Entities Table]
        Claims[Claims Table]
        Evidence[Evidence Table]
    end
    
    subgraph Output
        Inspect[garuda inspect]
        Impact[garuda impact]
        GraphViz[garuda graph]
        Diff[garuda diff]
    end
    
    Code --> AST --> Types --> Graph
    Graph --> Entities --> Inspect
    Graph --> Claims --> Impact
    Graph --> Evidence --> GraphViz
    Entities --> Diff
    Claims --> Diff
```

### Blast Radius Impact Engine

```mermaid
flowchart TD
    Start[Entity Change] --> BFS[BFS Traversal]
    
    BFS --> Depth1[Depth 1: Direct Consumers]
    BFS --> Depth2[Depth 2: Indirect Consumers]
    BFS --> DepthN[Depth N: Transitive Consumers]
    
    Depth1 --> Classify[Severity Classification]
    Depth2 --> Classify
    DepthN --> Classify
    
    Classify --> Critical[CRITICAL<br/>Public API/Contract]
    Classify --> High[HIGH<br/>Depth 1 Non-Test]
    Classify --> Medium[MEDIUM<br/>Depth 2 Public Route]
    Classify --> Low[LOW<br/>Test/Confidence<0.7]
    
    Critical --> Report[Impact Report]
    High --> Report
    Medium --> Report
    Low --> Report
    
    Report --> JSON[JSON Output]
    Report --> Human[Human-Readable]
```

### Future Company Brain Architecture

```mermaid
flowchart LR
    subgraph Today["Today: Single Repo"]
        R1[Repo 1] --> G1[Semantic Graph]
    end
    
    subgraph Tomorrow["Tomorrow: Multi-Repo"]
        R1[Repo 1] --> WS[Workspace]
        R2[Repo 2] --> WS
        R3[Repo 3] --> WS
        WS --> G2[Company Graph Alpha]
    end
    
    subgraph Future["Future: Company Brain"]
        R1[Repo 1] --> CB[Company Brain]
        R2[Repo 2] --> CB
        R3[Repo 3] --> CB
        R4[Repo 4] --> CB
        R5[Repo 5] --> CB
        R25[Repo 25] --> CB
        
        CB --> Policy[Policy Enforcement]
        CB --> Impact[Cross-Repo Impact]
        CB --> AI[AI Reasoning]
        CB --> Governance[Governance]
    end
    
    Today --> Tomorrow --> Future
```

---

## The Epistemic Model

To prevent AI hallucination and architecture drift, Garuda rigidly separates different classes of knowledge. **Observations are never collapsed into Decisions.**

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

### The Three Kinds of Truth

| Category | Definition | Example | Source |
|----------|------------|---------|--------|
| **Observation** | Directly extracted from source | `PaymentService IMPORTS Postgres` | AST/Type Checker |
| **Inference** | Derived by analysis/model logic | `CheckoutService CALLS PaymentService (conf: 0.87)` | Graph Traversal |
| **Decision** | Intentional organizational choice | `Production DB MUST BE Postgres` | Human/Policy |

### The Rule

> **Observation ≠ Inference ≠ Decision**
>
> A code analyzer observes that a service uses PostgreSQL. A human or policy may decide that production must use PostgreSQL. Those are different epistemic objects and must remain distinct.

---

## The 10 Immutable Engineering Laws

Garuda's development and execution are governed by strict V4 architectural doctrine:

| # | Law | Meaning |
|---|-----|---------|
| 1 | **Evidence before confidence** | No claim exists without a source pointer |
| 2 | **Epistemic separation** | Observations, inferences, decisions, and policies remain type-distinct |
| 3 | **Immutability** | Commits are immutable; analysis artifacts are versioned |
| 4 | **Stable Identity** | Entities require deterministic canonical IDs before cross-repo reasoning |
| 5 | **Least Privilege Traversal** | Authorization applies to graph traversal, not just storage reads |
| 6 | **Explicit Uncertainty** | Partial extraction must be visible, never silently upgraded to truth |
| 7 | **No AI Overwrites** | LLMs may summarize, but they cannot overwrite deterministic facts |
| 8 | **Verification First** | An unverified claim cannot be promoted to an enforced invariant |
| 9 | **Safe Failures** | A failed analysis never destroys the last known-good snapshot |
| 10 | **Evaluation Gates** | Production capabilities require strict benchmark evaluation |

---

## Installation & Quick Start

### Requirements

- Go 1.22+
- PostgreSQL 16+
- Docker & Docker Compose (optional, for local testing)

### Download Pre-built Binary

Download the pre-compiled binary for your architecture from [GitHub Releases](https://github.com/myshra777-ai/garuda/releases):

```bash
# Example: Linux AMD64
curl -sL https://github.com/myshra777-ai/garuda/releases/latest/download/garuda-linux-amd64 -o garuda
chmod +x garuda
sudo mv garuda /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda
go build -o bin/garuda cmd/garuda/main.go

# Optional: add to PATH
export PATH=$PATH:$(pwd)/bin
```

### Quick Workflow

```bash
# 1. Start PostgreSQL (Docker)
docker run --name garuda-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=garuda -p 5432:5432 -d postgres:16-alpine

# 2. Run idempotent database migrations (000-041)
DATABASE_URL="postgres://postgres:postgres@localhost:5432/garuda?sslmode=disable" go run cmd/migrate/main.go

# 3. Set environment variables
export GARUDA_TENANT_ID=$(uuidgen)
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/garuda?sslmode=disable"

# 4. Create a workspace
bin/garuda workspace create my-workspace
export GARUDA_WORKSPACE=my-workspace

# 5. Analyze and persist
bin/garuda analyze . --save

# 6. Explore
bin/garuda inspect User
bin/garuda graph my-workspace --open

# 7. Make a change, then compare
bin/garuda analyze . -o v1.json
# ... make code changes ...
bin/garuda analyze . -o v2.json
bin/garuda diff v1.json v2.json
```

### Configuration Variables

#### Core Runtime Variables

| Variable | Description | Default |
|---|---|---|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://test:test@localhost:5433/garuda_test?sslmode=disable` |
| `GARUDA_TENANT_ID` | Active tenant UUID for multi-tenant isolation | *Required for DB writes* |
| `GARUDA_WORKSPACE` | Active workspace name or UUID | `default` |

#### Developer Tooling Variables

| Variable | Description | Default |
|---|---|---|
| `GEMINI_API_KEY` | Key for AI-assisted changelog and docgen scripts (`scripts/`) | *Optional; only needed for doc generation* |

---

## Core Capabilities & CLI

### 🔍 Analysis & Diffing

Extract deterministic facts and compare architectural drift between commits.

```bash
# Generate snapshot
garuda analyze . -o v1.json

# Compare snapshots
garuda diff v1.json v2.json

# JSON output for CI
garuda diff v1.json v2.json --json -o diff-report.json
```

### 💥 Blast Radius Impact (`garuda impact`)

Perform a BFS graph traversal to identify downstream consumers affected by a code change. Includes confidence weighting and severity scoring (CRITICAL → HIGH → MEDIUM → LOW).

```bash
# Direct trace of an entity's blast radius
garuda impact --workspace <uuid> --target <entity-id>

# With custom depth and confidence
garuda impact --workspace <uuid> --target <entity-id> --depth 5 --min-confidence 0.7

# CI-ready diff impact analysis
garuda impact-diff v1.json v2.json --json
```

**Sample Output:**

```text
$ garuda impact --target e8b9c2a1-0f4e-4b2a-9e12-4c9a8f23b123 --depth 3

Target: internal/store.Migrate (Function)
Blast Radius: 14 impacted entities across 3 packages (Max Depth: 3)

[CRITICAL] (Depth 1) cmd/garuda/main.go:42 - main()
[HIGH]     (Depth 1) internal/server/handler.go:108 - HandleWorkspaceSync()
[MEDIUM]   (Depth 2) internal/router/routes.go:54 - RegisterAdminRoutes()
[LOW]      (Depth 3) internal/store/migrate_test.go:19 - TestMigrateIdempotency()

Confidence Score: 0.98 | Evidence Hash: 4f8b9e... (SHA-256)
```

**Severity Classification:**
- **CRITICAL**: Depth 1 public contract breach (API endpoint, SQL schema)
- **HIGH**: Depth 1 non-test consumer
- **MEDIUM**: Depth 2 public route consumer
- **LOW**: Test files, low confidence (<0.7), or depth ≥3

### 🧹 Code Quality (`garuda ponytail`)

Evaluate code minimalism using the semantic graph (not just file-local linting). Detects dead endpoints, orphaned DTOs, over-abstractions, and dependency obesity.

```bash
# Run quality checks
garuda ponytail .

# JSON output for CI
garuda ponytail . --json -o ponytail-report.json
```

**What it checks:**
- 🔴 **Dead Code**: Entities with zero incoming references
- 🟡 **Duplications**: Same name in different packages
- 📚 **Stdlib Alternatives**: Suggests `slices.Contains`, `slices.Sort`, etc.

### ⚖️ Governance (`garuda judge`)

Evaluate a proposed code snapshot against organizational decisions and policies, flagging architectural contradictions.

```bash
# Compare snapshots with governance judgement
garuda judge v1.json v2.json

# Block on breaking changes (CI)
garuda judge v1.json v2.json --block
```

### 🏗️ Workspace & Repository Management

```bash
# Workspace operations
garuda workspace create my-workspace
garuda workspace list
garuda workspace delete my-workspace

# Repository operations
garuda repo add my-workspace https://github.com/org/repo
garuda repo list my-workspace
garuda repo enable my-workspace https://github.com/org/repo

# Sync all repositories
garuda workspace sync my-workspace
```

### 🌐 Topology & Handoff Management (`garuda topology`)

Manage multi-agent task execution, topology sequences, and verifiable handoff lineages.

```bash
# List active topologies
garuda topology list

# Inspect topology task execution status
garuda topology inspect <topology-id>
```

### 📜 Policy Engine (`garuda policy`)

Validate codebase architecture against declarative constraints and detect violations.

```bash
# Check policy violations across workspace
garuda policy check --workspace <workspace-id>
```

### 🔒 Trust & Integrity

```bash
# Verify ledger integrity
garuda verify

# Explain a decision with evidence
garuda explain <decision-id>

# List entities
garuda entities
```

### 🛠️ Developer Tooling & Verification

```bash
# Run the 10-fixture ground truth AST benchmark suite
go test -v -race ./...

# Regenerate capabilities matrix from YAML schema
python3 scripts/generate_capabilities.py

# Generate release changelog (requires GEMINI_API_KEY)
python3 scripts/generate_changelog.py --context docs/generated/docs-context.json --output CHANGELOG.md
```

---

## Multi-Repository Scaling Model

Garuda explicitly follows an incremental 1→10→25 scaling model. Each gate must pass before advancing.

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

### Cross-Repository Resolution

Never merge entities only because names match. Resolution combines:
- Qualified names (strong)
- API contracts (strong)
- Import/reference graph (strong)
- Schema compatibility (strong)
- Naming similarity (weak — never enough alone)

---

## Security & Trust

### Security Guarantees

| Boundary | Required Guarantee |
|----------|-------------------|
| **Tenant** | No cross-tenant reads or writes |
| **Workspace** | Only authorized workspace members |
| **Repository** | Repository-level access policy enforced |
| **Evidence** | Unauthorized graph paths cannot reveal protected evidence |
| **MCP/API** | Tool calls inherit authorization context |
| **Agents** | Agent identity cannot bypass policy |
| **Audit** | Important mutations are traceable and integrity-verifiable |

### Evidence Contract

Every claim in Garuda carries:

```json
{
  "repository_id": "uuid",
  "commit_sha": "string",
  "file_path": "string",
  "symbol": "string",
  "line_start": 42,
  "line_end": 68,
  "content_snippet": "...",
  "content_hash": "sha256",
  "analyzer": "garuda",
  "analyzer_version": "v1.0"
}
```

### Integrity Chain

```text
Source file/commit → Analyzer run → Snapshot artifact → Entity/relationship extraction
→ Claim/observation → Evidence reference → User/API/CI answer
```

Every important answer can walk this chain backward.

---

## Roadmap & What's Next

### Current Status

| Phase | Focus | Status |
|-------|-------|--------|
| **P0** | Trust Foundation & Merkle Verification | ✅ Complete |
| **P1** | Semantic Core & Canonical UUIDv5 Identity | ✅ Complete |
| **P2** | Go AST Analyzer & Truth Benchmarks (10 Fixtures) | ✅ Complete |
| **P3** | CLI, Artifacts & Automated Docgen Pipeline | ✅ Complete |
| **P4** | Multi-Repo Sync & Cross-Repo Edges | 🧪 Alpha |
| **P5** | CI Integration (GitHub Action PR Impact) | 📋 Planned |
| **P6** | Governance & Policy Enforcement | 📋 Planned |
| **P7** | Business State Integrity | 📋 Planned |

### What's Next (Planned)

1. **Cross-Repo Impact Radius** — Expand BFS traversal across cross-repository boundary edges.
2. **Contract Extraction** — Deep schema extraction for HTTP routes, SQL migrations, and OpenAPI definitions.
3. **CI Action Wrapper** — GitHub Action for automated PR impact and architecture drift comments.
4. **Interactive Mind Map Explorer** — Enhanced UI disclosure for complex dependency paths.

---

## Development & Contributing

Contributions are strictly evaluated against the **10 Immutable Engineering Laws**.

### Adding a New Analyzer Feature?

1. ✅ Must be **deterministic** — avoid regex where AST parsing works
2. ✅ Must retain **evidence** — never emit an edge without line-level source tracking
3. ✅ Must pass the **benchmark corpus** — no regression in precision/recall

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
├── cmd/garuda/          # CLI entry points and command definitions
├── internal/
│   ├── analyzer/        # Go AST parsing and UUIDv5 canonical identity
│   ├── store/           # PostgreSQL migrations and atomic transactions
│   ├── impact/          # BFS blast-radius engine
│   ├── graph/           # Visualization and mind maps
│   └── engine/          # Governance judge and epistemic classification
├── test/
│   └── benchmark/       # 10-fixture truth benchmark suite
├── migrations/          # Idempotent PostgreSQL migrations (000-041)
├── scripts/             # Documentation generation scripts
├── docs/
│   └── CAPABILITIES.md  # Generated capabilities matrix
└── go.mod
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
