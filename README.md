

 🦅 Garuda — Evidence‑Backed Software Intelligence

[![Release](https://img.shields.io/github/v/release/myshra777-ai/garuda?style=flat-square&color=blue)](https://github.com/myshra777-ai/garuda/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/myshra777-ai/garuda/garuda-ci.yml?branch=main&style=flat-square&label=CI)](https://github.com/myshra777-ai/garuda/actions)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg?style=flat-square)](LICENSE)
[![Benchmarks](https://img.shields.io/badge/Benchmarks-100%25-brightgreen?style=flat-square)](#-ast-semantic-benchmark-suite)

---

**Understand your codebase as a connected, inspectable, and verifiable system.**

Garuda is a **deterministic** code intelligence and governance platform. It parses Go source code into a canonical AST, resolves symbols with **compiler‑grade `go/types`**, builds a structured semantic graph, anchors every relationship to cryptographic evidence (commit SHA, file path, line range, content hash), and provides high‑speed CLI tools for **blast‑radius impact analysis**, **architectural validation**, and **agentic code quality**.

> **Code → Semantics → Relationships → Evidence → Understanding**

> 🧠 **Building the Company Brain** – a continuously updated, cryptographically verifiable semantic model of your entire software ecosystem.

---

## 🚀 Why Garuda?

| Problem | Garuda's Answer |
|---------|----------------|
| **“What does this code actually do?”** | AST‑level extraction with `go/types` resolution yields a precise, machine‑readable semantic graph. |
| **“How does this change affect everything?”** | BFS blast‑radius impact analysis with depth‑based severity scoring (CRITICAL → LOW). |
| **“Can I trust this answer?”** | Every claim is backed by source evidence and a Merkle‑verified ledger. |
| **“Is this code minimal?”** | Ponytail checks for dead code, duplications, and stdlib alternatives. |
| **“Will this break downstream services?”** | Semantic diff and impact‑diff catch breaking changes before merge. |
| **“Does this interface implementation hold?”** | `go/types` resolves concrete receivers, embedded methods, and interface satisfaction. |

---

## 🔬 The Epistemic Triad

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

---

## 🧩 Capabilities Matrix

> 📄 **Full matrix**: [`docs/CAPABILITIES.md`](docs/CAPABILITIES.md)

### ✅ Stable — Production Ready

| Capability | Description | Command |
|------------|-------------|---------|
| **Semantic Analysis** | AST‑level extraction with `go/types` resolution | `garuda analyze` |
| **Canonical UUIDv5 Identity** | Deterministic identity across commits & renames | Core AST Engine |
| **Truth Benchmark Suite** | 23‑fixture verification – 100% precision/recall | `go test ./...` |
| **Entity Inspection** | View fields, methods, relationships | `garuda inspect` |
| **Graph Visualization** | Interactive D3.js HTML graph | `garuda graph` |
| **Snapshot Diff** | Semantic diff with breaking‑change detection | `garuda diff` |
| **Blast Radius Impact** | BFS traversal with severity (Critical → Low) | `garuda impact` |
| **Impact‑Diff** | Compare impact between two snapshots | `garuda impact-diff` |
| **Immutable Ledger** | Merkle‑verified audit trail | `garuda verify` |
| **Workspace/Repo Mgmt** | Create, list, sync, enable/disable | `garuda workspace`, `garuda repo` |

### 🧪 Alpha — Active Development

| Capability | Status |
|------------|--------|
| Cross‑Repo Impact | Alpha |
| Contract Extraction (HTTP, SQL, OpenAPI) | Alpha |
| Schema Discovery | In Progress |
| Code Quality (Ponytail) | Alpha |
| Governance Judge | Alpha |

### 📋 Planned

| Capability | Phase |
|------------|-------|
| CI Integration (GitHub Action) | Phase 5 |
| Company Graph (cross‑repo) | Phase 4 |
| Policy Enforcement | Phase 6 |
| Runtime Intelligence | Phase 6 |
| Business State Integrity | Phase 7 |

---

## 🔬 AST Semantic Benchmark Suite

Garuda's semantic extraction is evaluated against a **23‑fixture** synthetic and real‑world ground‑truth test suite. Every release is gated on deterministic precision and recall.

```mermaid
flowchart LR
    subgraph Suite1["AST Core Determinism (001–010)"]
        F1[001-basic] --> V[Verification Engine]
        F2[002-method-identity] --> V
        F3[003-interface] --> V
        F4[004-generics] --> V
        F5[005-alias] --> V
        F6[006-embedding] --> V
        F7[007-variadics] --> V
        F8[008-closures] --> V
        F9[009-interface-polymorphism] --> V
        F10[010-multi-method-struct] --> V
    end

    subgraph Suite2["Governance & System Truth (011–020)"]
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

    subgraph Suite3["Generic & Type System (021–023)"]
        F21[021-generic-instantiation] --> V
        F22[022-alias-vs-defined] --> V
        F23[023-embedded-promotion] --> V
    end
    
    V --> R[Report Engine]
    R --> G[Gate: 100% Precision / Recall]
```

### 📊 Snapshot Extraction Metrics

| Metric | Count | Status |
| :--- | :--- | :--- |
| **Parsed Files** | `214` | Complete Source Traversal |
| **Packages** | `43` | Fully Qualified Resolution |
| **Discovered Structs** | `322` | Tag & Line Span Verification |
| **Discovered Interfaces** | `25` | Dynamic Method Set Matching |
| **Functions & Methods** | `389` | Receiver Disambiguation Verified |
| **Total Struct Fields** | `1,581` | Direct & Embedded Fields |

### 🏆 Benchmark Results (23 Fixtures)

| Metric | Result | Gate |
|--------|--------|------|
| **Entity Precision** | **100.0%** | ≥ 98% |
| **Entity Recall** | **100.0%** | ≥ 98% |
| **Relationship Precision** | **100.0%** | ≥ 95% |
| **Relationship Recall** | **100.0%** | ≥ 90% |
| **Resolution Coverage** | **100.0%** | ≥ 95% |
| **Identity Stability** | **100.0%** | ≥ 99% |
| **Diff Accuracy** | **100.0%** | ≥ 95% |
| **Impact Precision** | **100.0%** | ≥ 95% |
| **Evidence Correctness** | **100.0%** | 100% |

### 📋 Truth Fixture Coverage (23 Fixtures)

| Fixture | Target Scenario | Key Invariant Tested |
|---|---|---|
| `001-basic` | Standard struct & function declarations | Canonical UUIDv5 hashing |
| `002-method-identity` | Value vs. pointer receiver methods | Disambiguated receiver method identity |
| `003-interface` | Explicit & implicit interface implementation | `IMPLEMENTS` edge emission |
| `004-generics` | Parameterized types & generic constraints | Type parameter erasure consistency |
| `005-alias` | Type definitions vs. type aliases (`type A = B`) | Target identity resolution |
| `006-embedding` | Struct & interface field embedding | Promoted field/method resolution |
| `007-variadics` | Ellipsis arguments & variadic signatures | Variadic call signature preservation |
| `008-closures` | Anonymous functions & closures | Nested lexical scope extraction |
| `009-interface-polymorphism` | Multi-interface satisfaction | Polymorphic dynamic method sets |
| `010-multi-method-struct` | Structs implementing overlapping interfaces | Accurate interface set mapping |
| `011-breaking-change` | Contract & signature modifications | Breaking change detection |
| `012-non-breaking-change` | Additive non-breaking changes | Zero false breaking alarms |
| `013-consumer-impact` | Multi-tier call hierarchy (api → service → store) | BFS blast radius propagation |
| `014-rename` | Struct & method identifier renames | UUIDv5 identity migration tracking |
| `015-delete` | Entity removal across versions | Orphaned edge elimination |
| `016-multi-repo` | Cross-module boundaries | Cross-repository edge classification |
| `017-evidence` | Exact line-range snippet hashing | SHA-256 evidence integrity |
| `018-determinism` | Repeated AST runs on identical source | Byte-for-byte snapshot replay identity |
| `019-noise` | Deep syntax noise and formatting variations | Robust parser stability (zero panics) |
| `020-large-ish` | Multi-package domain models | Scaled BFS impact traversal |
| `021-generic-instantiation` | Type parameter erasure & concrete method signatures | Generic method set resolution |
| `022-alias-vs-defined` | `type A = B` vs `type A B` | Distinct UUIDv5 identity |
| `023-embedded-promotion` | Promoted field/method resolution via `types.Selection` | Explicit `EMBEDS` edge emission |

### Run the Benchmark Suite

```bash
# Run all 23 truth benchmarks & invariant regression tests
go test -v -count=1 ./test/benchmark/...

# Run standalone benchmark reporter scorecard
go run garuda-bench/runner/main.go
```

---

## 🏗️ System Architecture

```mermaid
graph TD
    subgraph Source["Source Layer"]
        Git[Git Repository] --> Checkout[Checkout Commit]
        Checkout --> AST[Go AST + go/types Analyzer]
    end
    
    subgraph Semantic["Semantic Core"]
        AST --> |Extracts| E[Entities<br/>Packages, Structs, APIs]
        AST --> |Resolves| T[go/types<br/>Selections & Uses]
        T --> |Maps| R[Relationships<br/>Calls, Imports, Implements, Embeds]
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

### Compiler‑Grade Resolution Pipeline

Garuda now uses `go/types` to resolve symbols, receivers, and interfaces with compiler‑grade accuracy. The `workspaceImporter` resolves intra‑workspace dependencies on‑demand.

```mermaid
flowchart LR
    subgraph Input["Source Input"]
        S1[Go Files] --> P[Parser]
        S2[go.mod] --> P
    end
    
    subgraph Resolution["go/types Resolution"]
        P --> |AST| C[types.Config]
        C --> |Check| Info[types.Info<br/>Uses, Selections]
        Info --> R1[Resolve Receiver]
        Info --> R2[Resolve Interface]
        Info --> R3[Resolve Embeds]
        Info --> R4[Resolve Generics]
    end
    
    subgraph Output["Semantic Output"]
        R1 --> E1[Method Identity]
        R2 --> E2[IMPLEMENTS Edge]
        R3 --> E3[EMBEDS Edge]
        R4 --> E4[Generic Instantiation]
    end
```

### Workspace Importer (On‑Demand Package Resolution)

```mermaid
flowchart TD
    W[workspaceImporter] --> |Import| P[types.Package]
    P --> |Check| C[types.Config]
    C --> |Uses| I[types.Info]
    I --> |Selections| S[Resolved Receiver]
    I --> |Implements| I2[Interface Satisfaction]
    I --> |Uses| U[Concrete Symbol]
```

---

## 🏗️ Semantic Resolution Pipeline Architecture (Garuda V5)

Garuda processes Go source code through a **deterministic, four‑stage compilation and graph analysis pipeline**, mapping AST declarations to an authoritative semantic relationship graph and persisting cryptographic proofs into the immutable ledger.

```text
Source Files (*.go)
       │
       ▼
┌──────────────────────────────────────────────────────────────────┐
│ Stage T2: AST Parsing & Canonical Entity Extraction              │
│ • go/parser & token.FileSet tokenization                         │
│ • Metadata capture: struct fields, interface methods, signatures │
│ • Deterministic Canonical UUIDv5 & SHA-256 source hashing        │
└──────────────────────────────┬───────────────────────────────────┘
                               │
                               ▼
┌──────────────────────────────────────────────────────────────────┐
│ Stage T3: Authoritative Type Resolution (go/types)               │
│ • On-demand workspaceImporter for cross-package resolution       │
│ • Dynamic method-set evaluation (IMPLEMENTS relations)           │
│ • Disambiguation: Generics, Type Aliases, Embedded Promotions    │
└──────────────────────────────┬───────────────────────────────────┘
                               │
                               ▼
┌──────────────────────────────────────────────────────────────────┐
│ Stage T4: Semantic Diff & Graph Blast Radius Engine              │
│ • Signature & contract diffing between snapshots ($v_1 \to v_2$) │
│ • Breaking vs. Non-Breaking classification                       │
│ • Multi-hop reverse BFS graph traversal with severity decay      │
└──────────────────────────────┬───────────────────────────────────┘
                               │
                               ▼
┌──────────────────────────────────────────────────────────────────┐
│ Cryptographic Ledger Persistence & Attestation                   │
│ • Content-Addressed CAS: evidence_store (Full AST JSON)          │
│ • Bitemporal decisions & immutable decision_revisions            │
│ • Incremental tenant Merkle root attestation (SHA-256 chain)     │
└──────────────────────────────────────────────────────────────────┘
```

### Pipeline Stages

#### 1. Stage T2 — AST Corpus Ingestion & Canonical Identity
- Traverses the target repository and groups AST files by package path.
- Extracts high‑fidelity structural declarations: structs with field listings, interfaces with required method signatures, and package‑level functions.
- Assigns **RFC 4122 UUIDv5** canonical identifiers derived from `(PackagePath, ReceiverType, Name)` to eliminate identifier collision across branches.

#### 2. Stage T3 — Authoritative Type & Relationship Resolution
- Uses an integrated `workspaceImporter` to feed multi‑package workspace trees into `go/types.Config.Check` on demand.
- Disambiguates complex Go semantics without runtime reflection:
  - **Interface Satisfaction:** Verifies receiver method sets against interface contracts to emit typed `IMPLEMENTS` edges with 1.0 confidence.
  - **Generic Types & Instantiations:** Unpacks generic index expressions (`*Container[T]`) to isolate underlying types.
  - **Embedded Promotion:** Resolves inner struct method promotions directly to their canonical owners.
  - **Call Graph Tracing:** Inspects AST selector expressions and `types.Info.Uses` to construct `CALLS` and `CALLS_INTERFACE` edge sets while pruning language built‑ins and type casts.

#### 3. Stage T4 — Semantic Differentiation & BFS Blast Radius Engine
- **AST Contract Diffing:** Performs declarative entity comparisons across snapshots (\(v_1 \to v_2\)):
  - **Breaking:** Struct field deletions, interface method removals, signature mutation.
  - **Non‑Breaking:** Optional field additions, struct method extensions, internal AST hash mutations (`IMPLEMENTATION_CHANGE`).
- **Graph Blast Radius Traversal:** Executes a breadth‑first search (BFS) over incoming consumer edges from a mutated root:
  - **Depth 1 (Direct Implementation):** `CRITICAL` severity for concrete types satisfying broken interfaces.
  - **Depth 1 (Direct Caller):** `HIGH` severity for call sites directly invoking the target entity.
  - **Depth 2+ (Transitive Callers):** `MEDIUM` (depth 2) and `LOW` (depth ≥ 3) severity decay across upstream consumers.

#### 4. Cryptographic Storage & Ledger Guarantees
- **Content‑Addressed Storage (`evidence_store`):** Stores complete AST snapshots keyed by SHA‑256 block hash for immutable provenance replay.
- **Bitemporal Decision Ledger (`decisions`):** Tracks system state validity (`valid_from`, `approved_at`, `scope_domain`, `scope_system`).
- **Merkle Attestation (`merkle_roots`):** Computes \(\text{Root}_n = \text{SHA256}(\text{Root}_{n-1} \| \text{RevisionHash})\) within an atomic transaction to ensure zero state tampering.

### Architecture Invariants

| Phase | Invariant | Verification Mechanism |
|-------|-----------|------------------------|
| **AST Extraction** | Zero Identifier Collision | Canonical UUIDv5 derived from fully qualified AST signatures. |
| **Type Resolution** | Closed‑World Type Soundness | `workspaceImporter` fallback resolving intra‑workspace imports. |
| **Semantic Diff** | Deterministic Impact Classification | AST signature parity checking independent of whitespace/formatting. |
| **Blast Radius** | Strict Severity Monotonicity | BFS reverse traversal with bounded depth decay (C → H → M → L). |
| **Ledger Storage** | Non‑Repudiation & Immutability | Dual‑layer CAS storage with SHA‑256 Merkle chain attestation. |

---

## 💥 Blast Radius Impact Engine

When you change an entity, Garuda performs a **BFS traversal** of the dependency graph to find all downstream consumers. Severity classification decays with depth:

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

### Severity Classification

| Severity | Condition |
|----------|-----------|
| **CRITICAL** | Depth 1 public contract breach (API endpoint, SQL schema, interface) |
| **HIGH** | Depth 1 non‑test consumer |
| **MEDIUM** | Depth 2 public route consumer |
| **LOW** | Test files, confidence < 0.7, or depth ≥ 3 |

### 📟 Sample Output

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

## ⚖️ The 10 Immutable Engineering Laws

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

## 🚀 Installation & Quick Start

### Requirements

- Go 1.22+
- PostgreSQL 16+
- Docker (optional)

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

### 5‑Minute Workflow

```bash
# 1. Start PostgreSQL
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
| `GARUDA_TENANT_ID` | Tenant UUID | *Required* |
| `GARUDA_WORKSPACE` | Active workspace | `default` |

---

## ⌨️ CLI Reference

```bash
garuda analyze [path]           # Extract semantic snapshot
garuda diff <a> <b>             # Compare two snapshots
garuda impact --workspace <id> --target <entity-id>  # Blast radius
garuda impact-diff v1.json v2.json                   # Diff impact
garuda inspect <entity>         # View entity details
garuda graph <workspace>        # Generate interactive graph
garuda verify                   # Verify ledger integrity
garuda explain <decision-id>    # Explain a decision with evidence
garuda workspace create/list/delete/sync
garuda repo add/list/enable/disable
garuda ponytail .               # Code quality (dead code, duplicates, stdlib)
garuda judge v1.json v2.json    # Governance judgement
```

---

## 🔒 Cryptographic Evidence Contract

Every claim carries:

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

You can walk the integrity chain:

```text
Source → Analyzer → Snapshot → Entity → Claim → Evidence → Answer
```

---

## 📈 Multi-Repository Scaling

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

## 🗺️ Roadmap

| Phase | Focus | Status |
|-------|-------|--------|
| **P0** | Trust Foundation & Merkle | ✅ |
| **P1** | Semantic Core & UUIDv5 | ✅ |
| **P2** | Go Analyzer & 23 Benchmarks | ✅ |
| **P3** | CLI, Artifacts & Docgen | ✅ |
| **P4** | Multi‑Repo Sync | 🧪 Alpha |
| **P5** | CI Integration | 📋 Planned |
| **P6** | Governance & Policy | 📋 Planned |
| **P7** | Business State Integrity | 📋 Planned |

**Next up:**
- Cross‑repo impact radius expansion.
- Contract extraction (HTTP, SQL, OpenAPI).
- GitHub Action for PR comments.
- Enhanced graph explorer.

---

## 🔐 Security & Trust

| Boundary | Guarantee |
|----------|-----------|
| **Tenant** | No cross‑tenant access |
| **Workspace** | Authorized members only |
| **Repository** | Policy enforced |
| **Evidence** | Protected from unauthorized traversal |
| **MCP/API** | Authorization inherited |
| **Agents** | Identity enforced |
| **Audit** | Traceable & verifiable |

---

## 🤝 Contributing

Contributions must uphold the **10 Immutable Engineering Laws**.

When adding a new analyzer feature:

1. ✅ Be **deterministic** – AST over regex.
2. ✅ Retain **evidence** – line‑level source tracking.
3. ✅ Pass **benchmark corpus** – no precision/recall regressions.

### Local Development

```bash
docker-compose up -d postgres
DATABASE_URL="postgres://test:test@localhost:5433/garuda_test" go test ./...
go build -o bin/garuda cmd/garuda/main.go
```

---

## 📄 License

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
