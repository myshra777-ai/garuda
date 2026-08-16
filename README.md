# 🛡️ Garuda — Evidence-Backed Software Intelligence

**Understand your codebase as a connected, inspectable, and verifiable system.**

Garuda analyzes Go repositories, builds a structured semantic model of the software inside them, preserves the evidence behind that model, and provides interactive tools for exploring how the codebase is connected.

> **Code → Semantics → Relationships → Evidence → Understanding**

Garuda is being built incrementally: **single-repository correctness first, multi-repository intelligence next, Company Graph later.**

---

## 🧭 What Garuda Does

Modern codebases are not just collections of files.

A single repository can contain thousands of interconnected:

* 📦 Packages
* 📄 Files
* 🧱 Structs
* 🔌 Interfaces
* ⚙️ Functions
* 🧩 Methods
* 🔗 Dependencies
* 📞 Calls
* 🔍 References
* 🧬 Implementations

The difficult part is not finding a file.

The difficult part is understanding **how everything connects**.

Garuda builds a semantic representation of those connections so developers can move from:

```text
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

Instead of manually reconstructing the architecture, you can explore it.

---

# 🧠 The Core Idea

```text
                         SOURCE CODE
                              │
                              ▼
                    ┌──────────────────┐
                    │   Go Analyzer    │
                    │                  │
                    │ AST / Types      │
                    │ Imports / Calls  │
                    │ Dependencies     │
                    └────────┬─────────┘
                             │
                             ▼
                  ┌───────────────────────┐
                  │     SEMANTIC CORE     │
                  │                       │
                  │ Entities              │
                  │ Relationships         │
                  │ Claims                │
                  │ Observations          │
                  └───────────┬───────────┘
                              │
                              ▼
                     ┌────────────────┐
                     │    EVIDENCE    │
                     │                │
                     │ File           │
                     │ Location       │
                     │ Commit         │
                     │ Content / Hash │
                     └───────┬────────┘
                             │
                             ▼
                  ┌──────────────────────┐
                  │ IMMUTABLE ANALYSIS   │
                  │       HISTORY        │
                  │                      │
                  │ Revisions            │
                  │ Merkle Integrity     │
                  │ Provenance           │
                  └──────────┬───────────┘
                             │
                ┌────────────┼────────────┐
                ▼            ▼            ▼
             CLI          GRAPH       ARTIFACTS
                │            │            │
                └────────────┼────────────┘
                             ▼
                    DEVELOPERS + AI
```

### The principle

> **Build structured truth first. Put generative intelligence on top of it later.**

Garuda does not start with an LLM guessing what your repository means.

It starts with the source code.

---

# 🔗 A Codebase as a Web

Garuda's graph is designed to represent software as an interconnected system.

For example:

```text
                         ┌───────────────┐
                         │   Package     │
                         └───────┬───────┘
                                 │
              ┌──────────────────┼──────────────────┐
              │                  │                  │
              ▼                  ▼                  ▼
        ┌──────────┐       ┌──────────┐       ┌──────────┐
        │   File   │       │  Struct  │       │ Function │
        └────┬─────┘       └────┬─────┘       └────┬─────┘
             │                  │                  │
             │             ┌────┴────┐             │
             │             │         │             │
             ▼             ▼         ▼             ▼
        ┌─────────┐   ┌─────────┐ ┌────────┐ ┌────────────┐
        │ Method  │   │Interface│ │ Type   │ │ Dependency │
        └────┬────┘   └────┬────┘ └────────┘ └─────┬──────┘
             │             │                       │
             └─────────────┼───────────────────────┘
                           ▼
                     ┌────────────┐
                     │  External  │
                     │  Package   │
                     └────────────┘
```

Relationships are represented explicitly rather than being inferred only by the visualization layer.

Examples include:

```text
CALLS
IMPORTS
REFERENCES
CONTAINS
DEFINES
IMPLEMENTS
EMBEDS
DEPENDS_ON
```

A relationship can therefore be explored as:

```text
PaymentService
      │
      ├── CONTAINS ──────► ProcessPayment()
      │
      ├── CALLS ─────────► ValidatePayment()
      │
      ├── DEPENDS_ON ────► PaymentRepository
      │
      └── REFERENCES ────► PaymentRequest
```

The objective is simple:

> **Turn the codebase into something developers can navigate as a system, not merely as a directory tree.**

---

# 🔍 Evidence Is Part of the Graph

Garuda is not intended to be a graph of unexplained assertions.

Semantic information is tied back to evidence.

```text
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

This creates a critical distinction.

Garuda should be able to represent:

> **What exists**

and, where supported:

> **Why Garuda believes it exists**

and:

> **Where the supporting evidence came from**

This is the foundation for trustworthy software intelligence.

---

# 🔐 Immutable Analysis & Integrity

Garuda preserves analysis history instead of treating every analysis as disposable output.

```text
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
    ├────────► Entities
    │
    ├────────► Relationships
    │
    ├────────► Claims
    │
    └────────► Evidence
```

The trust layer is designed around:

* 🔒 Immutable revisions
* 🔗 Cryptographic integrity
* 🌳 Merkle-backed verification
* 📍 Provenance
* 📦 Reproducible analysis artifacts

The result is an analysis history that can be inspected and verified rather than silently overwritten.

---

# 🧩 Garuda's Semantic Model

Garuda deliberately separates different kinds of information.

For example:

```text
OBSERVATION

"Service A imports package B"


        ≠


INFERENCE

"Service A probably depends on Service B"


        ≠


DECISION

"Service A must use Service B"
```

This distinction matters.

A code analyzer can observe relationships in source code.

It should not automatically turn those observations into organizational policies or decisions.

> **Observed software state is not the same thing as organizational truth.**

---

# 🗂️ What Garuda Understands

The semantic model is designed around explicit entities, relationships, claims, observations, and evidence.

### Entities

Depending on analyzer coverage:

| Entity       | Description                     |
| ------------ | ------------------------------- |
| 📦 Package   | Go package                      |
| 📄 File      | Source file                     |
| 🧱 Struct    | Struct/type definition          |
| 🔌 Interface | Interface definition            |
| ⚙️ Function  | Function declaration            |
| 🧩 Method    | Method associated with a type   |
| 🏷️ Type     | Other semantic type definitions |
| 🌐 External  | External dependency or package  |

### Relationships

| Relationship | Meaning                                |
| ------------ | -------------------------------------- |
| `CALLS`      | One function or method invokes another |
| `IMPORTS`    | A package/file imports another package |
| `REFERENCES` | An entity references another entity    |
| `CONTAINS`   | One entity contains another            |
| `DEFINES`    | A file/package defines an entity       |
| `IMPLEMENTS` | A type implements an interface         |
| `EMBEDS`     | A type embeds another type             |
| `DEPENDS_ON` | A semantic dependency exists           |

The exact set of relationships grows with analyzer coverage and validation.

---

# 🖥️ Interactive Graph

Garuda's graph interface is designed around **progressive exploration**.

A user should be able to move from:

```text
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

The intended interaction model includes:

* 🔎 Search
* 🖱️ Click-to-select
* 🧭 Progressive exploration
* 🔗 Relationship traversal
* 🔍 Zoom and pan
* 🧩 Entity filtering
* 📋 Detailed entity information
* 🧾 Relationship information
* 📍 Evidence/provenance inspection

The graph is not meant to replace the underlying semantic model.

It is a **visual interface over it**.

---

# ⚙️ CLI

Garuda is designed to be useful directly from the command line.

## Stable Core Commands

| Command                           | Purpose                                 |
| --------------------------------- | --------------------------------------- |
| `garuda analyze . --save`         | Analyze and persist semantic state      |
| `garuda analyze . -o v1.json`     | Generate a semantic snapshot            |
| `garuda inspect <entity>`         | Inspect a semantic entity               |
| `garuda diff v1.json v2.json`     | Compare semantic snapshots              |
| `garuda graph <workspace> --open` | Open the interactive graph              |
| `garuda verify`                   | Verify analysis integrity               |
| `garuda explain <id>`             | Trace available evidence                |
| `garuda self-describe`            | Generate structured product information |

## Active Development

These capabilities are being developed and validated separately from the stable semantic-analysis core:

| Command                        | Purpose                                |
| ------------------------------ | -------------------------------------- |
| `garuda justify <entity>`      | Explain why an entity/code path exists |
| `garuda judge v1.json v2.json` | Governance-oriented comparison         |
| `garuda ponytail .`            | Code-quality and hygiene analysis      |

## Planned

| Capability             | Direction                                          |
| ---------------------- | -------------------------------------------------- |
| Cross-repository graph | Connect semantic models across repositories        |
| CI integration         | Bring semantic analysis into development workflows |
| Policy evaluation      | Evaluate software state against explicit policies  |

> Command maturity is intentionally documented separately so experimental capabilities are not presented as production guarantees.

---

# 🚀 Quick Start

## 1. Clone Garuda

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda
```

## 2. Build

```bash
go build -o garuda cmd/garuda/*.go
```

## 3. Start the stack

```bash
./garuda up
```

## 4. Analyze a repository

```bash
./garuda analyze . --save
```

## 5. Inspect an entity

```bash
./garuda inspect PaymentService
```

## 6. Generate the graph

```bash
./garuda graph my-workspace --open
```

## 7. Generate a snapshot

```bash
./garuda analyze . -o v1.json
```

## 8. Change the code and analyze again

```bash
./garuda analyze . -o v2.json
```

## 9. Compare the semantic state

```bash
./garuda diff v1.json v2.json
```

## 10. Verify the analysis history

```bash
./garuda verify
```

---

# 🔄 The Garuda Workflow

```text
             ┌───────────────┐
             │  SOURCE CODE  │
             └───────┬───────┘
                     │
                     ▼
              ┌─────────────┐
              │   ANALYZE   │
              └──────┬──────┘
                     │
                     ▼
             ┌───────────────┐
             │   SEMANTIC    │
             │     MODEL     │
             └───────┬───────┘
                     │
          ┌──────────┼──────────┐
          ▼          ▼          ▼
       INSPECT      GRAPH      SNAPSHOT
          │          │          │
          └──────────┼──────────┘
                     ▼
                  CHANGE
                     │
                     ▼
                  ANALYZE
                     │
                     ▼
                   DIFF
                     │
                     ▼
                  VERIFY
```

The central loop is:

> **Analyze → Understand → Explore → Change → Re-analyze → Diff → Verify**

---

# 📦 Data Model

At a high level:

```text
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
                    ┌───────────┼───────────┐
                    ▼           ▼           ▼
                 Entities  Relationships  Claims
                                             │
                                             ▼
                                          Evidence
```

Core concepts:

| Concept      | Purpose                                      |
| ------------ | -------------------------------------------- |
| Workspace    | Logical analysis boundary                    |
| Repository   | Source system being analyzed                 |
| Commit       | Source state associated with analysis        |
| Analysis     | One execution of the analyzer                |
| Snapshot     | Structured representation of analysis output |
| Entity       | Identifiable software element                |
| Relationship | Typed connection between entities            |
| Claim        | Semantic statement represented by Garuda     |
| Evidence     | Source information supporting that statement |

---

# 🗄️ Storage Architecture

Garuda uses PostgreSQL for structured persistence.

Conceptually:

```text
PostgreSQL
│
├── Workspaces
├── Repositories
├── Commits
├── Analyses
├── Entities
├── Relationships
├── Claims
├── Observations
└── Evidence Metadata
```

The graph experience is a semantic projection of this underlying information.

It is not intended to become a disconnected visualization layer.

---

# 🏗️ Architecture

```text
┌──────────────────────────────────────────────┐
│                  SOURCES                     │
│                                              │
│       Git • Code • Commits • Documents       │
└──────────────────────┬───────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────┐
│              ANALYSIS LAYER                  │
│                                              │
│        Go AST • Types • Dependencies         │
└──────────────────────┬───────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────┐
│               SEMANTIC CORE                  │
│                                              │
│ Entities │ Relationships │ Claims │ Evidence │
└──────────────────────┬───────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────┐
│               TRUST LAYER                    │
│                                              │
│ Immutable Revisions • Merkle • Provenance    │
└──────────────────────┬───────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────┐
│              EXPERIENCE LAYER                │
│                                              │
│        CLI • Graph • API • Artifacts         │
└──────────────────────┬───────────────────────┘
                       │
                       ▼
                 DEVELOPERS + AI
```

---

# 🎯 Current MVP

Garuda is currently focused on **single-repository Go software intelligence**.

The current MVP prioritizes:

* ✅ Go source analysis
* ✅ Semantic entity extraction
* ✅ Relationship modeling
* ✅ Repository snapshots
* ✅ Semantic inspection
* ✅ Semantic diffing
* ✅ Evidence and provenance
* ✅ Immutable analysis history
* ✅ Cryptographic integrity
* ✅ Interactive graph exploration

Go is the first language target.

Additional languages are intentionally deferred until the Go semantic model and evaluation process are mature enough to justify expansion.

---

# 🧪 Why Single Repository First?

Garuda is deliberately following:

```text
        1 Repository
             │
             ▼
       Validate Semantics
             │
             ▼
        Benchmark
             │
             ▼
       10 Repositories
             │
             ▼
       25 Repositories
             │
             ▼
      Company Graph
```

A weak semantic analyzer becomes a much bigger problem when multiplied across hundreds of repositories.

Garuda therefore follows a simple rule:

> **Correctness before scale.**

The objective of the MVP is not to claim that Garuda already understands an entire company.

The objective is to make Garuda understand **one repository extremely well**.

---

# 🧭 Roadmap

## Phase 0 — Trust Foundation

**Status: ✅ Complete**

* Immutable revisions
* Cryptographic integrity
* Merkle-backed history
* Provenance
* Audit foundations

---

## Phase 1 — Semantic Core

**Status: ✅ Complete**

* Entity model
* Relationship model
* Claims
* Observations
* Evidence model
* Workspace/repository foundations

---

## Phase 2 — Go Analyzer

**Status: 🧪 MVP / Active Refinement**

* Go AST analysis
* Semantic entity extraction
* Relationship extraction
* Repository-level semantic model
* Analysis snapshots
* Accuracy validation

---

## Phase 3 — CLI & Interactive Exploration

**Status: 🧪 MVP / Active Refinement**

* `analyze`
* `inspect`
* `diff`
* `verify`
* `explain`
* `graph`
* Interactive semantic exploration

---

## Phase 4 — Multi-Repository

**Status: 🔜 Next Major Expansion**

```text
Repository A ─────┐
Repository B ─────┤
Repository C ─────┼────► Company Semantic Graph
Repository D ─────┤
Repository E ─────┘
```

Focus:

* Repository synchronization
* Cross-repository entities
* Cross-repository relationships
* Dependency mapping
* Company-level graph exploration

---

## Phase 5 — Contract Intelligence

**Status: 📋 Planned**

Move from understanding code structure toward understanding contracts between systems.

Potential areas include:

* API relationships
* producer/consumer relationships
* missing dependencies
* contractual gaps
* impact analysis

---

## Phase 6 — Runtime Intelligence

**Status: 📋 Planned**

Combine static semantic understanding with runtime evidence.

```text
STATIC CODE
     │
     ├──────────┐
     │          │
     ▼          ▼
Semantic     Runtime
Graph       Evidence
     │          │
     └────┬─────┘
          ▼
     Richer System
     Understanding
```

---

## Phase 7+ — Governance & Trusted Organizational Intelligence

**Status: 📋 Future**

Long-term areas include:

* policy evaluation
* agent governance
* decision intelligence
* business-state integrity
* temporal analysis
* AI-assisted reasoning over evidence-backed software state

These are future layers, not claims about the current MVP.

---

# 🧠 Garuda's Design Principles

### 01 — Deterministic First

If a fact can be reliably extracted from source code, prefer deterministic analysis over probabilistic interpretation.

### 02 — Evidence Before Confidence

A claim should be traceable to the evidence supporting it.

### 03 — Structured Truth Before Generative Intelligence

Build the semantic model first.

Use AI on top of structured information rather than asking AI to reconstruct everything from raw code.

### 04 — Observation ≠ Inference ≠ Decision

Different epistemic categories must remain distinguishable.

### 05 — Immutable History

Changes in semantic state should be inspectable over time.

### 06 — Progressive Disclosure

Do not overwhelm the developer with the entire graph.

Move naturally from:

```text
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

### 07 — Scale Only After Correctness

The Company Graph is the destination.

A trustworthy single-repository graph is the foundation.

---

# 🤖 AI + Garuda

Garuda's long-term AI architecture is intentionally different from:

```text
Code → LLM → Answer
```

The intended model is:

```text
                   SOURCE CODE
                       │
                       ▼
              DETERMINISTIC ANALYSIS
                       │
                       ▼
               SEMANTIC KNOWLEDGE
                       │
                       ▼
                    EVIDENCE
                       │
                       ▼
                      AI
                       │
                       ▼
              REASONING / UX
```

This allows AI systems to reason over structured and traceable software information.

The AI should not become the source of truth.

> **Garuda provides the substrate. Models provide reasoning.**

---

# 📚 Documentation as Code

Garuda is also building a documentation pipeline around explicit product contracts.

```text
product.yaml
      │
      ├──────────────┐
      ▼              ▼
capabilities.yaml  roadmap.yaml
      │              │
      └──────┬───────┘
             ▼
      docs-context.json
             │
             ▼
       Documentation AI
             │
       ┌─────┼─────┐
       ▼     ▼     ▼
    README  API  CHANGELOG
```

The important rule is:

> **AI generates documentation. AI does not define product reality.**

Product positioning, capability maturity, and roadmap state remain explicitly controlled by project contracts.

---

# 🔎 Product Self-Description

Garuda can generate a machine-readable description of the product:

```bash
./garuda self-describe --workspace my-workspace
```

And a Markdown representation:

```bash
./garuda self-describe \
  --workspace my-workspace \
  --markdown \
  --output README.md
```

This is intended to become part of an automated documentation workflow so that README, API documentation, changelogs, and related material remain synchronized with the project.

---

# 📊 What Garuda Is — and Isn't

| Garuda is                             | Garuda is not                                            |
| ------------------------------------- | -------------------------------------------------------- |
| 🧠 Semantic software intelligence     | ❌ Merely a code search tool                              |
| 🕸️ An interconnected code graph      | ❌ Just a static visualization                            |
| 🔍 Evidence-aware analysis            | ❌ An unexplained AI answer engine                        |
| 🔐 Integrity-aware analysis history   | ❌ Disposable analysis output                             |
| 🧩 A structured semantic substrate    | ❌ A replacement for your source code                     |
| 🚧 Building toward a Company Graph    | ❌ Already a complete company-wide brain                  |
| 🤖 Designed for AI-assisted reasoning | ❌ Dependent on an LLM to understand basic code structure |

---

# 🚧 What Is Not the MVP

Garuda intentionally does **not** treat the following as completed capabilities:

* Full multi-language support
* Large-scale multi-repository analysis
* Complete cross-repository dependency resolution
* Company-wide semantic intelligence
* Runtime-informed analysis
* Full natural-language graph querying
* Advanced policy governance
* Business-state integrity
* Mature autonomous agent governance

These are part of the longer-term architecture.

The MVP stays focused.

---

# 🛠️ Development

## Requirements

* Go 1.21+
* PostgreSQL 16+
* Git

## Build

```bash
go build -o garuda cmd/garuda/*.go
```

## Test

```bash
go test ./...
```

---

# 🤝 Contributing

Contributions are welcome.

When contributing to Garuda:

1. Keep semantic behavior deterministic where possible.
2. Add tests for new behavior.
3. Preserve evidence and provenance.
4. Avoid silently changing existing semantic meanings.
5. Keep documentation aligned with implemented capabilities.
6. Clearly distinguish experimental functionality from stable functionality.

For larger architectural changes, open an issue or discussion before implementation.

---

# 📜 License

Apache License 2.0

---

# 🦅 Garuda

**Understand the code.**

**See the connections.**

**Trace the evidence.**

**Build the Company Graph.**

<p align="center">

### `Code → Semantics → Graph → Evidence → Intelligence`

</p>
