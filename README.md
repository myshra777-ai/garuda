# 🦅 Garuda — Evidence-Backed Software Intelligence

**Understand your codebase as a connected, inspectable, and verifiable system.**

Garuda is a deterministic code intelligence and governance platform. It analyzes Go repositories to build a structured semantic graph of your software, preserves cryptographically hashed evidence behind every claim, and provides high-speed CLI tools for blast-radius impact analysis, architectural validation, and agentic code quality.

> **Code → Semantics → Relationships → Evidence → Understanding**

> 🧠 **Building the Company Brain**: Garuda is the foundation for a continuously updated, cryptographically verifiable semantic model of your entire software ecosystem.

---

## 📋 Table of Contents

- [The Core Philosophy](#the-core-philosophy)
- [Company Brain — The North Star](#company-brain--the-north-star)
- [Capabilities Matrix](#capabilities-matrix)
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
    subgraph Current[✅ Current: Single Repo Intelligence]
        C1[Semantic Graph] --> C2[Blast Radius] --> C3[Evidence Backing]
    end
    
    subgraph Growing[🧪 Growing: Multi-Repo Foundation]
        G1[Workspace Sync] --> G2[Cross-Repo Detection] --> G3[Company Graph Alpha]
    end
    
    subgraph Future[📋 Future: Company Brain]
        F1[Cross-Repo Impact] --> F2[Policy Enforcement] --> F3[AI Reasoning]
    end
    
    Current --> Growing --> Future
```

### What's Already Built (The Foundation)

| Component | Status | What It Does |
|-----------|--------|--------------|
| **Semantic Graph** | ✅ Stable | Entities (structs, functions, APIs) + relationships (calls, imports, dependencies) |
| **Blast Radius Impact** | ✅ Stable | BFS traversal to find who depends on what |
| **Multi-Repository Sync** | ✅ Stable | Analyze 10-25 repositories in one workspace |
| **Evidence Backing** | ✅ Stable | Every claim traces to source lines + commit SHA |
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

Garuda's capabilities are explicitly categorized by maturity. This reflects the V4 roadmap doctrine: **evidence before confidence, stable before experimental.**

### ✅ Stable — Production Ready

| Capability | Description | Command |
|------------|-------------|---------|
| **Semantic Analysis** | Extracts entities (structs, functions, interfaces, packages) with line-level evidence | `garuda analyze` |
| **Entity Inspection** | View entity details: package, file, kind, fields, methods, incoming/outgoing relationships | `garuda inspect` |
| **Graph Visualization** | Interactive HTML graph of workspace architecture with D3.js | `garuda graph` |
| **Snapshot Diff** | Compare two semantic snapshots with breaking change detection | `garuda diff` |
| **Blast Radius Impact** | BFS traversal to find all consumers of an entity with severity scoring | `garuda impact` |
| **Diff Impact** | Compare impact between two snapshots | `garuda impact-diff` |
| **Code Quality (Ponytail)** | Dead code detection, duplicates, standard-library alternatives | `garuda ponytail` |
| **Governance Judge** | Breaking change detection with blocking decisions for CI | `garuda judge` |
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
| **Line Number Extraction** | Precise line ranges for evidence | ✅ Complete |
| **AI Guardrails** | Ponytail principles for AI code generation | Planned |

### 📋 Planned — Future

| Capability | Description | Phase |
|------------|-------------|-------|
| **CI Integration** | GitHub Action for PR impact/quality comments | Phase 5 |
| **Company Graph** | Cross-repository semantic graph | Phase 4 |
| **Policy Enforcement** | Evaluate code against organizational policies | Phase 6 |
| **Runtime Intelligence** | Combine static analysis with runtime evidence | Phase 6 |
| **Business State Integrity** | Idempotency keys, business rule enforcement | Phase 7 |

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

- Go 1.21+
- PostgreSQL 16+
- Docker (optional, for `garuda up`)

### Build from Source

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda
go build -o garuda cmd/garuda/*.go
```

### Quick Workflow

```bash
# 1. Set tenant ID
export GARUDA_TENANT_ID=$(uuidgen)

# 2. Start the PostgreSQL stack
./garuda up

# 3. Create a workspace
./garuda workspace create my-workspace
export GARUDA_WORKSPACE=my-workspace

# 4. Analyze and persist
./garuda analyze . --save

# 5. Explore
./garuda inspect User
./garuda graph my-workspace --open

# 6. Make a change, then compare
./garuda analyze . -o v1.json
# ... make code changes ...
./garuda analyze . -o v2.json
./garuda diff v1.json v2.json
```

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

### 🔒 Trust & Integrity

```bash
# Verify ledger integrity
garuda verify

# Explain a decision with evidence
garuda explain <decision-id>

# List entities
garuda entities
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
| **P0** | Trust Foundation | ✅ Complete |
| **P1** | Semantic Core | ✅ Complete |
| **P2** | Go Analyzer | ✅ Complete |
| **P3** | CLI & Artifacts | ✅ Complete |
| **P4** | Multi-Repo Sync | 🧪 Alpha |
| **P5** | CI Integration | 📋 Planned |
| **P6** | Governance | 📋 Planned |
| **P7** | Business Integrity | 📋 Planned |

### What's Next (Planned)

1. **Cross-Repo Impact** — Enhance blast radius across repository boundaries
2. **Contract Extraction** — HTTP routes, SQL migrations, API schemas
3. **CI Integration** — GitHub Action wrapper for `garuda diff` and `garuda impact`
4. **Benchmark Suite** — Precision/recall metrics for semantic extraction

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
go build -o garuda cmd/garuda/*.go
```

### Project Structure

```
garuda/
├── cmd/garuda/          # CLI (30+ commands)
├── internal/
│   ├── analyzer/        # Go AST parsing
│   ├── store/           # PostgreSQL operations
│   ├── impact/          # Blast radius engine
│   ├── graph/           # Visualization
│   ├── engine/          # Governance
│   └── ...              # 25+ internal packages
├── migrations/          # 70+ SQL migrations
├── docs/                # Documentation
├── .github/workflows/   # CI/CD
└── go.mod
```

---

## License

Apache License 2.0

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
