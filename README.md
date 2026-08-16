```markdown
# Garuda

### Evidence-backed software intelligence for Go codebases.

Garuda analyzes a Go repository, builds a structured semantic model of its entities and relationships, preserves the evidence behind that model, and lets developers explore what exists, how components connect, and what changed.

> **Understand your codebase. Follow its relationships. Trace the evidence.**

Garuda starts with a single repository and is being built toward a multi‑repository Company Graph.

---

## Why Garuda?

Modern codebases are too large to understand from files and folders alone.

A repository contains:

- packages
- files
- structs
- interfaces
- functions
- methods
- APIs
- dependencies
- calls
- references
- implementation relationships

But traditional code browsing forces developers to reconstruct these relationships manually.

Garuda creates a **machine-readable semantic model** of the repository so that the structure of the system can be explored directly.

Instead of asking:

> "Where is this implemented?"

you can ask:

> "What is this entity connected to?"

> "What depends on it?"

> "What changed?"

> "Where did Garuda get this information?"

The goal is not simply to search code. The goal is to build a **continuously verifiable understanding of software**.

---

## The Core Idea

```
                         SOURCE CODE
                              │
                              ▼
                       ┌─────────────┐
                       │ Go Analyzer │
                       │ AST / Types │
                       └──────┬──────┘
                              │
                              ▼
                     ┌─────────────────┐
                     │  Semantic Core  │
                     │                 │
                     │ Entities        │
                     │ Relationships   │
                     │ Claims          │
                     │ Observations    │
                     └────────┬────────┘
                              │
                              ▼
                         ┌──────────┐
                         │ Evidence │
                         │          │
                         │ File     │
                         │ Location │
                         │ Commit   │
                         │ Hash     │
                         └────┬─────┘
                              │
                              ▼
                    ┌───────────────────┐
                    │ Immutable Snapshot│
                    │ + Integrity Layer │
                    └─────────┬─────────┘
                              │
                ┌─────────────┼──────────────┐
                ▼             ▼              ▼
             CLI          Graph         Artifacts
                │             │              │
                └─────────────┼──────────────┘
                              ▼
                     Developers + AI
```

Garuda deliberately follows a simple principle:

> **Build structured truth first. Put generative intelligence on top of it later.**

---

## What Garuda Understands

Garuda's semantic model is designed around explicit entities, relationships, claims, and evidence.

### Entities

Depending on analyzer coverage, Garuda can represent entities such as:

- Packages
- Files
- Structs
- Interfaces
- Types
- Functions
- Methods
- APIs
- Dependencies
- External packages

Every entity has a stable identity within the semantic model.

### Relationships

The semantic graph represents relationships between entities.

Examples include:

```
CALLS
IMPORTS
REFERENCES
CONTAINS
DEFINES
IMPLEMENTS
EMBEDS
DEPENDS_ON
```

This turns a repository from a collection of files into an interconnected system.

For example:

```
PaymentService
      │
      ├── CONTAINS ──► ProcessPayment()
      │
      ├── DEPENDS_ON ──► PaymentRepository
      │
      ├── CALLS ──► ValidatePayment()
      │
      └── REFERENCES ──► PaymentRequest
```

The graph is designed to make these relationships directly explorable.

---

## Evidence Is Part of the Model

Garuda is not an unexplained graph.

Important semantic information is traceable back to source evidence.

Conceptually:

```
Entity / Relationship
        │
        ▼
      Claim
        │
        ▼
     Evidence
        │
        ├── Repository
        ├── Commit
        ├── File
        ├── Location
        └── Content / Hash
```

This allows Garuda to answer not only:

> "What does the system contain?"

but also:

> "Why does Garuda believe this?"

and:

> "What source evidence supports this relationship?"

This evidence‑first architecture is a core design principle.

---

## Immutable Analysis

Garuda treats analysis results as versioned artifacts.

A simplified lifecycle is:

```
Repository
    │
    ▼
Commit
    │
    ▼
Analysis
    │
    ▼
Snapshot
    │
    ▼
Entities + Relationships
    │
    ▼
Claims + Evidence
```

The trust layer includes:

- immutable revisions
- cryptographic integrity
- Merkle‑backed verification
- provenance
- reproducible analysis artifacts

This allows historical semantic state to be preserved instead of continuously overwriting the latest result.

---

## Explore Your Codebase as a Graph

Garuda's graph experience is designed around the idea that software should be explored as a connected system rather than as isolated files.

```
                    ┌─────────────┐
                    │   Package   │
                    └──────┬──────┘
                           │
             ┌─────────────┼─────────────┐
             ▼             ▼             ▼
         ┌──────┐      ┌────────┐    ┌────────┐
         │ File │      │ Struct │    │ Func   │
         └──┬───┘      └───┬────┘    └───┬────┘
            │              │             │
            └──────────────┼─────────────┘
                           ▼
                    ┌────────────┐
                    │ Dependency │
                    └─────┬──────┘
                          │
                          ▼
                    ┌────────────┐
                    │  External  │
                    └────────────┘
```

Progressive exploration:

```
Repository
    ↓
Package
    ↓
File
    ↓
Entity
    ↓
Relationship
    ↓
Evidence
    ↓
Source
```

The objective is to let a developer select an entity and progressively understand its surrounding system without losing the underlying provenance.

---

## Current MVP Focus

Garuda is currently focused on **one repository** – getting it genuinely understandable before scaling outward.

The initial implementation focuses on:

- Go source analysis
- deterministic semantic extraction
- repository snapshots
- entity modeling
- relationship modeling
- semantic inspection
- semantic diffing
- evidence and provenance
- immutable analysis history
- cryptographic integrity
- interactive graph exploration

Go is the first analyzer target.

Additional languages are intentionally deferred until the Go semantic model and evaluation process are sufficiently mature.

---

## CLI Commands

Garuda is primarily designed to be used from the command line.

### Production‑Ready Commands

| Command | Purpose | Status |
|---------|---------|--------|
| `garuda analyze . --save` | Analyze and persist semantic state | ✅ Stable |
| `garuda analyze . -o v1.json` | Create machine‑readable snapshot | ✅ Stable |
| `garuda inspect <entity>` | Inspect a semantic entity | ✅ Stable |
| `garuda diff v1.json v2.json` | Compare semantic snapshots | ✅ Stable |
| `garuda graph <workspace> --open` | Generate interactive graph | ✅ Stable |
| `garuda verify` | Verify cryptographic integrity | ✅ Stable |
| `garuda explain <id>` | Trace evidence for a claim | ✅ Stable |
| `garuda self-describe` | Generate product description | ✅ Stable |

### Commands in Active Development

| Command | Purpose | Status |
|---------|---------|--------|
| `garuda justify <entity>` | Explain why code exists | 🧪 Active |
| `garuda judge v1.json v2.json` | Governance report | 🧪 Active |
| `garuda ponytail .` | Code hygiene detection | 🧪 Active |

### Planned / Future Commands

| Command | Purpose | Status |
|---------|---------|--------|
| `garuda graph <workspace> --cross-repo` | Cross‑repository graph | 🔜 Planned |
| `garuda ci` | CI integration | 🔜 Planned |
| `garuda policy` | Policy evaluation | 🔜 Planned |

---

## Product Self‑Description

Garuda contains a self‑description capability designed to keep product documentation aligned with the actual product state.

```bash
# Generate a machine‑readable product description
garuda self-describe --workspace my-workspace

# Generate a Markdown README skeleton
garuda self-describe \
  --workspace my-workspace \
  --markdown \
  --output README.md
```

The documentation workflow is designed so that product documentation is driven by explicit product and capability contracts rather than allowing an LLM to guess what the software does.

---

## Architecture

Garuda is intentionally layered.

```
┌──────────────────────────────────────────────┐
│                  SOURCES                     │
│       Git / Code / Commits / Documents      │
└──────────────────────┬───────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────┐
│             HARVEST & ANALYSIS              │
│       AST / Go Types / Dependencies         │
└──────────────────────┬───────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────┐
│               SEMANTIC CORE                 │
│                                              │
│ Entities │ Relationships │ Claims │ Evidence │
└──────────────────────┬───────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────┐
│              TRUST / INTEGRITY              │
│                                              │
│ Immutable revisions │ Merkle │ Provenance   │
└──────────────────────┬───────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────┐
│                 EXPERIENCE                  │
│                                              │
│       CLI │ API │ Graph │ Artifacts         │
└──────────────────────┬───────────────────────┘
                       │
                       ▼
              DEVELOPERS + AI
```

A key architectural rule is that Garuda keeps different kinds of knowledge distinct.

For example:

```
OBSERVATION
"Service A imports package B"

        ≠

INFERENCE
"Service A probably depends on Service B"

        ≠

DECISION
"Service A must use Service B"
```

Keeping these epistemic categories separate prevents observed implementation details from being incorrectly presented as organizational decisions or policies.

---

## Data Model

At the conceptual level:

```
Workspace
    │
    └── Repository
           │
           └── Commit
                  │
                  └── Analysis
                         │
                         └── Snapshot
                                │
                ┌───────────────┼───────────────┐
                ▼               ▼               ▼
             Entities      Relationships      Claims
                                                │
                                                ▼
                                             Evidence
```

Core concepts include:

| Concept | Purpose |
|---|---|
| Workspace | Company or project boundary |
| Repository | Source system being analyzed |
| Commit | Immutable source state |
| Analysis | One analyzer execution |
| Snapshot | Normalized semantic output |
| Entity | Identifiable software element |
| Relationship | Typed connection between entities |
| Claim | Statement represented by Garuda |
| Evidence | Source information supporting a claim |

---

## Storage

Garuda uses **PostgreSQL** as the primary structured persistence layer.

Conceptually:

```
PostgreSQL
├── Tenants
├── Workspaces
├── Repositories
├── Commits
├── Analyses
├── Entities
├── Relationships
├── Claims
├── Evidence metadata
└── Governance data
```

Immutable analysis artifacts are represented as structured snapshots. Larger evidence and reproducible artifacts use content‑addressed storage.

The graph is a semantic experience and projection over this structured information – not an isolated visualization disconnected from the underlying evidence.

---

## Trust Model

Garuda is built around a simple principle:

> **Evidence before confidence.**

The system distinguishes:

```
Observed
   ↓
Supported by evidence
   ↓
Represented as semantic information
   ↓
Versioned
   ↓
Cryptographically verifiable
```

This makes the graph more than a visualization. It becomes an inspectable representation of what Garuda knows about the software system.

---

## Development

### Requirements

- Go 1.21+
- PostgreSQL 16+
- Git

### Build

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda
go build -o garuda cmd/garuda/*.go
```

### Run tests

```bash
go test ./...
```

### Quick Start

```bash
# Start the stack
garuda up

# Analyze your codebase
garuda analyze . --save

# Generate interactive graph
garuda graph my-workspace --open

# Inspect an entity
garuda inspect PaymentService
```

### Example Workflow

```bash
# 1. Analyze the repository
garuda analyze . --save

# 2. Inspect the semantic model
garuda inspect PaymentService

# 3. Generate the graph
garuda graph my-workspace --open

# 4. Create a snapshot
garuda analyze . -o before.json

# 5. Make code changes...

# 6. Analyze again
garuda analyze . -o after.json

# 7. Compare semantic changes
garuda diff before.json after.json

# 8. Verify Garuda's integrity
garuda verify
```

The important workflow is:

```
Analyze
   ↓
Understand
   ↓
Explore
   ↓
Change
   ↓
Re-analyze
   ↓
Diff
   ↓
Verify
```

---

## Roadmap

Garuda follows an incremental strategy:

```
1 Repository
      ↓
10 Repositories
      ↓
25 Repositories
      ↓
Company Graph
      ↓
Trusted State & Governance
```

### Current Direction

| Phase | Focus | Status |
|---|---|---|
| P0 | Trust Foundation | ✅ Complete |
| P1 | Semantic Core | ✅ Complete |
| P2 | Go Analyzer | ✅ Complete |
| P3 | Artifact + CLI | ✅ Complete |
| P4 | Governance (justify, judge, ponytail) | 🧪 Active |
| P5 | Multi‑Repository | 🔜 Planned |
| P6 | Company Brain | 🔜 Planned |
| P7 | CI / PR Integration | 🔜 Planned |
| P8 | Governance Policies | 🔜 Planned |
| P9 | Business Integrity | 📋 Long‑term |

The important milestone before multi‑repository expansion is **confidence in single‑repository semantic analysis**.

The goal is not to build a large graph quickly. The goal is to build a graph that can be trusted.

---

## Design Principles

Garuda follows several principles:

### Deterministic first
Prefer deterministic analysis for facts that can be extracted from source code.

### Evidence before confidence
Important claims should have traceable evidence.

### Structured truth before generative intelligence
Build the semantic model before asking an LLM to reason over it.

### Observations are not decisions
What the code does and what an organization has decided should remain separate concepts.

### Immutable history
Analysis results should be reproducible and historically inspectable.

### Progressive disclosure
Start with the repository and progressively drill down:

```
Repository
 → Package
 → File
 → Entity
 → Relationship
 → Evidence
 → Source
```

### Scale only after correctness
The 1 → 10 → 25 repository strategy prevents a weak single‑repository analyzer from becoming a weak Company Graph.

---

## What Garuda Is Becoming

Garuda is being developed in three broad layers.

### 1. Software Intelligence
Understand one repository.

> "What is in this codebase and how does it work?"

### 2. Company Brain
Connect repositories and services.

> "How do all our systems connect?"

### 3. Trusted State & Governance
Connect software knowledge with organizational decisions, policies, and business state.

> "Can this change or state transition happen safely?"

These layers are intentionally developed in sequence. Garuda is currently focused on the first layer.

---

## Not the MVP Yet

The following areas are intentionally being developed later:

- Full multi‑language analysis
- Large‑scale multi‑repository deployments
- Complete cross‑repository dependency resolution
- Company‑wide semantic overview
- Deep CI / PR integration
- Advanced policy governance
- Runtime‑informed semantic analysis
- Business‑state integrity
- Full natural‑language graph querying

These are part of the longer‑term direction rather than claims about the current MVP.

---

## Contributing

Contributions are welcome.

A good contribution should:

1. Have a clear purpose.
2. Preserve deterministic behavior where possible.
3. Include tests for new behavior.
4. Preserve evidence and provenance semantics.
5. Avoid silently changing the meaning of existing semantic entities or relationships.
6. Keep documentation aligned with actual capabilities.

For larger changes, open an issue first to discuss the proposed design.

---

## License

Apache License 2.0

---

## Status

Garuda is an actively developed project.

The current MVP is focused on **Go‑based software intelligence and evidence‑backed semantic exploration**.

The architecture is intentionally being developed incrementally, with correctness and evidence preceding multi‑repository scale and governance.

---

<p align="center">
  <strong>Garuda</strong><br>
  Understand your codebase. Follow its relationships. Trace the evidence.
</p>
```
