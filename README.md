# 🦅 Garuda

<p align="center">
  <img src="assets/garuda-logo.png" width="128" alt="Garuda — Evidence-backed software intelligence">
</p>

<h1 align="center">Garuda</h1>

<p align="center">
  <strong>Analyze | Verify | Collaborate</strong>
</p>

<p align="center">
  <em>Built as a shared intelligence workspace.</em>
</p>

<p align="center">
  <a href="https://github.com/myshra777-ai/garuda/releases"><img src="https://img.shields.io/badge/version-v0.2.0-blue.svg?style=flat-square" alt="Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-green.svg?style=flat-square" alt="License"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.22+-00ADD8.svg?style=flat-square&logo=go" alt="Go Version"></a>
  <a href="EVIDENCE.md"><img src="https://img.shields.io/badge/verification-14%20repos%20passed-success.svg?style=flat-square" alt="Verification Status"></a>
  <a href="EVIDENCE.md"><img src="https://img.shields.io/badge/GAP--20-87.2%25%20token%20reduction-purple.svg?style=flat-square" alt="GAP-20 Benchmark"></a>
</p>

<p align="center">
  Garuda gives humans and AI agents a shared, evidence-backed understanding of a software system.
</p>

<p align="center">
  <a href="PLAYBOOK.md">Playbook</a> ·
  <a href="EVIDENCE.md">Evidence</a> ·
  <a href="docs/">Architecture</a> ·
  <a href="SECURITY.md">Security</a> ·
  <a href="CHANGELOG.md">Changelog</a>
</p>

---

## Overview

Modern software teams are increasingly distributed across humans, AI coding agents, repositories, services, runtime systems, and constantly changing architectural decisions.

Garuda provides a persistent semantic workspace that connects:

* source code
* semantic entities
* relationships and dependencies
* evidence and provenance
* revisions and lineage
* decisions and policies
* runtime observations
* verification state
* AI and developer workflows

The central idea is:

> **Humans and AI agents should not have to rediscover the software system independently every time they work on it.**

```mermaid
flowchart LR
    CODE["Source Code"] --> ANALYZE["Analyze"]
    RUNTIME["Runtime"] --> VERIFY["Verify"]
    DECISIONS["Decisions"] --> WORKSPACE["Garuda Workspace"]

    ANALYZE --> SEMANTIC["Semantic State"]
    SEMANTIC --> EVIDENCE["Evidence"]
    EVIDENCE --> VERIFY

    VERIFY --> WORKSPACE

    WORKSPACE --> HUMAN["Developers"]
    WORKSPACE --> IDE["IDE"]
    WORKSPACE --> AI["AI Agents"]
    WORKSPACE --> CI["CI / Governance"]
```

---

# The Problem

Software organizations increasingly operate with multiple developers and heterogeneous AI coding agents working on the same system.

Without a shared semantic state, each participant repeatedly reconstructs architecture from source text, local context, previous conversations, or incomplete documentation.

## 1. Context Fragmentation

A developer changes a subsystem. Another developer or AI agent later works on a related component with a separate context window.

The second agent may need to rediscover:

* existing symbols
* dependency paths
* interfaces
* architectural decisions
* prior investigations
* previous changes
* runtime observations

The result is duplicated discovery and increased opportunity for incorrect assumptions.

```mermaid
flowchart LR
    A["Developer A + AI"] --> REBUILD1["Reconstruct System Context"]
    B["Developer B + AI"] --> REBUILD2["Reconstruct System Context"]
    C["Developer C + AI"] --> REBUILD3["Reconstruct System Context"]

    REBUILD1 --> DUP["Repeated Discovery"]
    REBUILD2 --> DUP
    REBUILD3 --> DUP

    DUP --> COST["Context + Time + Token Overhead"]
```

## 2. Duplicate AI Investigation

Large repositories often require substantial context exploration before an agent can safely answer a question.

Without persistent semantic state:

```mermaid
flowchart TD
    TASK1["Agent Task A"] --> EXPLORE1["Explore Repository"]
    EXPLORE1 --> MODEL1["Build Mental Model"]

    TASK2["Agent Task B"] --> EXPLORE2["Explore Repository"]
    EXPLORE2 --> MODEL2["Build Mental Model"]

    TASK3["Agent Task C"] --> EXPLORE3["Explore Repository"]
    EXPLORE3 --> MODEL3["Build Mental Model"]

    MODEL1 --> DUPLICATE["Repeated Reconstruction"]
    MODEL2 --> DUPLICATE
    MODEL3 --> DUPLICATE
```

Garuda persists the reusable part of that system understanding so later workflows can query structured state instead of reconstructing everything from raw source again.

## 3. Engineering Knowledge Decays

Important knowledge is distributed across:

* source code
* pull requests
* code reviews
* architecture discussions
* incident investigations
* human memory
* AI conversations
* discarded approaches
* previous implementations

When those contexts disappear, the reasoning behind the current architecture becomes difficult to recover.

Garuda treats machine-relevant engineering state as structured data that can be inspected, queried, revised, verified, and shared.

---

# What Is Garuda?

Garuda is an **evidence-backed software intelligence and shared intelligence workspace**.

It builds a structured representation of a software system containing:

```text
Entities
   │
   ├── Relationships
   ├── Dependencies
   ├── Claims
   ├── Evidence
   ├── Observations
   ├── Verification State
   ├── Decisions
   ├── Lineage
   ├── Revisions
   └── Topology
```

That state can be consumed through:

```mermaid
flowchart TB
    WS["GARUDA WORKSPACE"]

    WS --> CLI["CLI"]
    WS --> API["API"]
    WS --> DASH["Dashboard"]
    WS --> IDE["IDE"]
    WS --> MCP["MCP"]
    WS --> CI["CI / Governance"]

    CLI --> USERS["Developers"]
    DASH --> USERS
    IDE --> USERS

    MCP --> AGENTS["AI Agents"]
    API --> SYSTEMS["Systems / Integrations"]
    CI --> DELIVERY["Engineering Workflow"]
```

Garuda is therefore **not simply a graph generator**.

The graph is one representation of a broader semantic state.

---

# Analyze | Verify | Collaborate

Garuda is organized around three core operations.

```mermaid
flowchart LR
    ANALYZE["ANALYZE<br/>Understand the system"] -->
    VERIFY["VERIFY<br/>Compare claims with evidence"] -->
    COLLABORATE["COLLABORATE<br/>Share the resulting state"]
```

| Capability      | What Garuda does                                                                 | Why it matters                                      |
| --------------- | -------------------------------------------------------------------------------- | --------------------------------------------------- |
| **Analyze**     | Extracts entities, relationships, dependencies and evidence from source          | Creates structured software understanding           |
| **Verify**      | Correlates static structure with runtime observations and detects contradictions | Separates supported knowledge from assumptions      |
| **Collaborate** | Exposes semantic state to developers, IDEs, CI, dashboards and AI agents         | Gives multiple participants a shared system context |

---

# The Garuda Semantic Model

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

## 1. Observation

What does the system contain or what has been observed?

Examples include:

* AST structures
* resolved symbols
* source declarations
* runtime telemetry
* runtime spans

## 2. Claim

What does the available information imply?

Examples include:

* calls relationships
* interface relationships
* dependency relationships
* repository topology

## 3. Evidence

Why should a claim be trusted?

Evidence can include:

* source files
* line locations
* revisions
* runtime observations
* provenance

## 4. Verification

What can currently be established from available evidence?

```text
SUPPORTED
UNVERIFIED
CONTRADICTED
```

## 5. Decision

What did the team decide?

Examples include:

* architectural choices
* accepted approaches
* rejected approaches
* policies
* unresolved decisions

## 6. Lineage

How did the current state evolve?

Lineage connects changes, revisions, decisions, supersessions, and previous states.

---

# Evidence and Verification

A core Garuda principle is:

> **Absence of evidence is not evidence of absence.**

Static structure and runtime behaviour answer different questions.

```mermaid
flowchart TB
    SOURCE["SOURCE CODE"] --> STATIC["STATIC ANALYSIS"]
    STATIC --> SEMANTIC["SEMANTIC STATE"]

    RUNTIME["RUNTIME TELEMETRY"] --> OBS["RUNTIME OBSERVATIONS"]

    SEMANTIC --> CORRELATE["CORRELATE"]
    OBS --> CORRELATE

    CORRELATE --> SUPPORTED["SUPPORTED"]
    CORRELATE --> UNVERIFIED["UNVERIFIED"]
    CORRELATE --> CONTRADICTED["CONTRADICTED"]
```

## Verification States

| State          | Meaning                                                                                                   |
| -------------- | --------------------------------------------------------------------------------------------------------- |
| `SUPPORTED`    | Static declarations and observed runtime behaviour agree for the evaluated relationship                   |
| `UNVERIFIED`   | Static structure exists, but available runtime evidence is insufficient to establish runtime verification |
| `CONTRADICTED` | Observed runtime behaviour conflicts with the evaluated static expectation or policy                      |

### Important distinction

`UNVERIFIED` does **not** mean:

* dead code
* broken code
* unused code
* incorrect code

It means only:

> **There is not enough runtime evidence to claim verification.**

`CONTRADICTED` is different:

> **Observed runtime behaviour conflicts with the expected static or policy state.**

---

# From Source to Shared Engineering State

```mermaid
flowchart LR
    SOURCE["Source Code"] -->
    AST["Go AST + go/types"] -->
    ENTITIES["Canonical Entities"] -->
    REL["Semantic Relationships"] -->
    EVIDENCE["Evidence + Revision"]

    EVIDENCE --> STATIC["Static State"]
    EVIDENCE --> RUNTIME["Runtime Observations"]

    STATIC --> VERIFY["Verification"]
    RUNTIME --> VERIFY

    VERIFY --> SUPPORTED["SUPPORTED"]
    VERIFY --> UNVERIFIED["UNVERIFIED"]
    VERIFY --> CONTRADICTED["CONTRADICTED"]

    VERIFY --> WORKSPACE["Shared Workspace"]

    WORKSPACE --> HUMAN["Human"]
    WORKSPACE --> IDE["IDE"]
    WORKSPACE --> AI["AI"]
```

---

# Verification Flow

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

## Runtime Verification Pipeline

Current validation measured the following behaviour:

| Stage                                  | Current validated behaviour                       |
| -------------------------------------- | ------------------------------------------------- |
| Telemetry admission                    | HTTP `202 Accepted`                               |
| Measured ingestion response            | ~`1.8 ms p95` under the tested workload           |
| Verification cycle                     | Asynchronous, using 10-second epochs              |
| Full dashboard / consensus propagation | Typically `1–2 minutes` in the tested environment |

These values describe observed validation runs, not universal production SLAs.

---

# IDE Integration

Garuda integrates with editor workflows through Language Server Protocol and Model Context Protocol components.

## Inline Contradictions

When an evaluated runtime observation conflicts with the expected architecture, the relevant code location can be surfaced in the editor.

<p align="center">
  <img src="assets/screenshots/ide-contradictions.png" width="850" alt="Garuda inline contradiction detection">
</p>

## Symbol Context and Blast Radius

Symbols can expose their surrounding dependency context and recursive relationships.

<p align="center">
  <img src="assets/screenshots/ide-graph-view.png" width="850" alt="Garuda symbol blast radius and graph view">
</p>

Example:

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

## Interactive Topology

The development daemon can host the topology visualizer at:

```text
http://localhost:8080/graph
```

A complete UI walkthrough is available in:

[docs/WALKTHROUGH.md](docs/WALKTHROUGH.md)

---

# One Shared Workspace

The shared-workspace model is intentionally simple.

```mermaid
flowchart TD
    TEAM["ENGINEERING TEAM"]

    TEAM --> DEV["Developers"]
    TEAM --> AGENTS["AI Agents"]

    DEV --> WS["GARUDA WORKSPACE"]
    AGENTS --> WS

    WS --> ENTITIES["Entities"]
    WS --> REL["Relationships"]
    WS --> EVIDENCE["Evidence"]
    WS --> DECISIONS["Decisions"]
    WS --> CONTRA["Contradictions"]
    WS --> RUNTIME["Runtime Observations"]

    ENTITIES --> SHARED["Shared Engineering State"]
    REL --> SHARED
    EVIDENCE --> SHARED
    DECISIONS --> SHARED
    CONTRA --> SHARED
    RUNTIME --> SHARED
```

Humans and AI agents can consume the same underlying semantic state when they are configured against the same Garuda workspace and database.

> **Shared state requires shared infrastructure.** Separate local Garuda databases are not automatically synchronized.

---

# Why This Matters for AI Agents

AI agents can reason over code, but repeated repository exploration is expensive.

Without shared semantic state:

```mermaid
flowchart LR
    AGENTA["Agent A"] --> READ1["Read Source"]
    READ1 --> MODEL1["Build Context"]

    AGENTB["Agent B"] --> READ2["Read Source"]
    READ2 --> MODEL2["Build Context"]

    AGENTC["Agent C"] --> READ3["Read Source"]
    READ3 --> MODEL3["Build Context"]

    MODEL1 --> DUP["Duplicate Discovery"]
    MODEL2 --> DUP
    MODEL3 --> DUP
```

With Garuda:

```mermaid
flowchart LR
    SOURCE["Software System"] --> GARUDA["Garuda Workspace"]

    GARUDA --> STRUCTURE["Structure"]
    GARUDA --> EVIDENCE["Evidence"]
    GARUDA --> STATE["Current State"]

    STRUCTURE --> CONTEXT["Grounded Context"]
    EVIDENCE --> CONTEXT
    STATE --> CONTEXT

    CONTEXT --> CLAUDE["Claude"]
    CONTEXT --> GPT["GPT"]
    CONTEXT --> GEMINI["Gemini"]
    CONTEXT --> CURSOR["Cursor"]
```

The objective is not to make an AI omniscient.

The objective is to give agents a reusable, inspectable, structured representation of the system they are operating on.

---

# Decisions Are Software State

Architecture is shaped by decisions as much as by code.

Garuda treats decisions and their lineage as part of the broader semantic workspace.

```mermaid
flowchart TD
    PROPOSAL["Proposal"] --> DECISION["Decision"]

    DECISION --> EVID["Evidence"]
    DECISION --> CONTEXT["Context"]
    DECISION --> REV["Revision"]
    DECISION --> ACTOR["Author / Actor"]

    DECISION --> STATUS{"Superseded?"}

    STATUS -->|No| CURRENT["Current Decision"]
    STATUS -->|Yes| NEXT["New Decision"]

    NEXT --> LINEAGE["Decision Lineage"]
```

This makes it possible to recover machine-relevant context such as:

* why a dependency exists
* why an implementation was rejected
* why a service boundary was introduced
* what policy applies
* what approach replaced an earlier one
* what remains unresolved

For multi-agent workflows, this helps avoid repeatedly proposing approaches that have already been evaluated and rejected.

---

# MCP and AI Workflows

Garuda exposes its semantic state through Model Context Protocol workflows.

```mermaid
flowchart LR
    AGENT["AI Agent"] --> MCP["MCP"]
    MCP --> GARUDA["Garuda"]
    
    GARUDA --> SEMANTICS["Semantics"]
    GARUDA --> EVIDENCE["Evidence"]
    GARUDA --> STATE["Current State"]
    GARUDA --> DECISIONS["Decisions"]

    SEMANTICS --> CONTEXT["Grounded Context"]
    EVIDENCE --> CONTEXT
    STATE --> CONTEXT
    DECISIONS --> CONTEXT

    CONTEXT --> AGENT
```

Examples of structured questions include:

```text
What calls this function?

What is the blast radius of this change?

Where is this type implemented?

What evidence supports this relationship?

Has this path been observed at runtime?

Are there contradictions involving this package?

What architectural decisions affect this service?

Was this approach previously rejected?

What changed since the previous revision?
```

The goal is to make software state **machine-readable as well as human-readable**.

---

# CI/CD Governance

Garuda can participate in CI workflows through `garuda ci` and `garuda judge`.

```bash
garuda judge baseline.json proposed.json
```

The workflow evaluates a proposed state against a baseline.

```mermaid
flowchart TD
    PR["Pull Request"] --> CI["garuda ci"]
    CI --> DIFF["Semantic / AST Diff"]
    DIFF --> BASELINE["Baseline State"]
    BASELINE --> EVAL["Policy + Contract Evaluation"]
    DIFF --> EVAL

    EVAL --> PASS["PASS"]
    EVAL --> BLOCK["BLOCK"]
    EVAL --> REPORT["Audit Report"]

    BLOCK --> REPORT
```

Validated CI-oriented capabilities include:

* contract-breakage isolation
* architectural policy evaluation
* dependency and topology evaluation
* blast-radius reporting
* automated PR reporting

These workflows are currently categorized as **Early Testing** rather than universal production guarantees.

---

# Business and Engineering Impact

The current validation corpus includes controlled workflow measurements covering onboarding, catch-up, architectural review, dependency lookup and AI context usage.

| Workflow                       | Traditional Approach                        | Garuda Workflow                                  | Observed Result                                        |
| ------------------------------ | ------------------------------------------- | ------------------------------------------------ | ------------------------------------------------------ |
| New engineer onboarding        | Manual documentation and repository tracing | Interactive topology and blast-radius navigation | ~80% faster time-to-first-PR in the evaluated workflow |
| Post-leave catch-up            | Meetings and manual PR archaeology          | Structured workspace summary and topology        | ~15 minutes in the evaluated workflow                  |
| PR architectural review        | Manual contract inspection                  | Automated CI evaluation                          | Automated detection in the evaluated benchmark         |
| Incident blast-radius analysis | Manual search and dependency tracing        | Symbol-level dependency lookup                   | Sub-second lookup in the evaluated workflow            |
| AI context overhead            | 4,850 tokens                                | 620 tokens                                       | ~87.2% reduction in GAP-20 benchmark                   |

> These are benchmarked observations from controlled environments and specific workflows. They are not universal guarantees.

---

# Validation Evidence

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

For full methodology, screenshots, validation artifacts and historical runs:

**[→ Open the Garuda Evidence Center](EVIDENCE.md)**

---

# GAP-20 Grounding Benchmark

The GAP-20 benchmark compares unassisted LLM exploration with Garuda-grounded workflows.

| Metric                        | Naive LLM | Garuda-Grounded |     Observed Difference |
| ----------------------------- | --------: | --------------: | ----------------------: |
| Symbol precision              |     40.0% |          100.0% |   +60 percentage points |
| Structural hallucination rate |     66.7% |            0.0% | −66.7 percentage points |
| Prompt token overhead         |     4,850 |             620 |        ~87.2% reduction |
| Upstream caller recall        |     20.0% |          100.0% |   +80 percentage points |
| Downstream dependency recall  |     33.0% |          100.0% |   +67 percentage points |
| Violation quarantine rate     |      0.0% |          100.0% |  +100 percentage points |

### What the benchmark demonstrates

In the evaluated benchmark tasks:

* Garuda-grounded workflows reached **100% symbol precision**
* structural hallucination rate was **0%**
* context token overhead was reduced by approximately **87.2%**
* tested upstream and downstream dependency recall improved to **100%**
* evaluated violations were quarantined at **100%**

The structural hallucination result refers specifically to fabricated functions, receivers, and symbols in the evaluated benchmark tasks.

It does **not** claim elimination of all forms of AI hallucination.

---

# Capability Status

## Stable

The following form the current semantic foundation:

* Go semantic analysis
* compiler-backed type resolution
* deterministic entity identity
* semantic relationships
* repository and workspace modelling
* evidence provenance
* semantic snapshots
* semantic diff
* impact analysis
* impact diff
* lineage
* cryptographic state verification
* global workspace search
* progressive architecture exploration
* CLI workflows
* dashboard exploration

## Early Testing

Implemented capabilities currently undergoing active validation:

* multi-repository intelligence
* OpenTelemetry ingestion
* runtime observations
* static ↔ runtime correlation
* contradiction verification
* runtime evidence views
* MCP integration
* Garuda IDE
* grounding benchmark harness
* agent-oriented workflows
* `garuda ci`
* `garuda judge`
* PR governance and policy automation

These capabilities should be treated as **early testing rather than universal production guarantees**.

## Launching Soon

The next product-facing layer focuses on continuous operational use:

* richer runtime verification
* deeper MCP workflows
* expanded evidence exploration
* broader grounding benchmarks
* production-oriented telemetry workflows
* additional developer workflow integrations

## Longer-term Direction

Garuda is being developed toward a broader software intelligence substrate connecting source, runtime, evidence, decisions, policies, AI agents, operational state and governance.

Longer-term directions include:

* company-scale software graphs
* multi-language semantic intelligence
* richer runtime correlation
* stronger architectural governance
* AI-grounded software operations
* business-state integrity
* autonomous verification and remediation

These are development directions, not current production guarantees.

---

# Design Principles

### 1. Evidence over inference

Prefer inspectable evidence over opaque conclusions.

### 2. Unknown stays unknown

Missing evidence must not silently become certainty.

### 3. Identity is deterministic

The same semantic entity should remain identifiable across analysis runs.

### 4. State is auditable

Important state transitions should leave an inspectable trail.

### 5. Architecture should be navigable

Users should be able to move from system-level topology to specific entities and evidence.

### 6. Static and runtime knowledge remain distinct

Static analysis describes what the source contains.

Runtime observation describes what has been observed executing.

These are complementary signals, not interchangeable facts.

### 7. AI should consume structured state

Agents should not have to reconstruct an entire repository from raw text whenever reusable semantic state already exists.

### 8. Correctness before scale

Garuda prioritizes validated semantics and evidence before expanding language and deployment scope.

---

# What Garuda Is Not

Garuda is not:

* another text search engine
* another code linter
* another generic dependency graph
* another observability platform
* another documentation generator
* another generic vector database
* an LLM that guesses how a repository works

Garuda is intended to connect these layers through a structured semantic and evidence substrate.

---

# A Graph Is Only One View

A conventional dependency graph answers:

> **What connects to what?**

Garuda is designed to answer broader questions:

```text
What exists?

What depends on it?

What evidence supports that relationship?

What has actually been observed?

What contradicts the expected architecture?

What decision led to the current design?

What changed?

What superseded the previous decision?

What should an AI agent know before changing this component?
```

```mermaid
flowchart TB
    GRAPH["Dependency / Topology Graph"] --> SYSTEM["System Understanding"]

    SYSTEM --> STRUCTURE["Structure"]
    SYSTEM --> EVIDENCE["Evidence"]
    SYSTEM --> RUNTIME["Runtime Observations"]
    SYSTEM --> DECISIONS["Decisions"]
    SYSTEM --> LINEAGE["Lineage"]
    SYSTEM --> STATE["Verification State"]
```

---

# Quick Start

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | sh
```

## Initialize

```bash
garuda init
```

## Start the stack

```bash
garuda up
```

## Analyze a repository

```bash
garuda analyze .
```

## Start the development daemon

```bash
garuda dev
```

The unified daemon provides the local API, telemetry collector, Merkle worker and UI components.

## Open the topology visualizer

```bash
open http://localhost:8080/graph
```

---

# MCP Configuration

For Cursor or Claude Desktop, configure the Garuda MCP server against the same workspace and PostgreSQL database used by the team.

Example:

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

For detailed installation, operational procedures and workflow examples:

**[→ Read the Garuda Playbook](PLAYBOOK.md)**

---

# Documentation

| Document                           | Purpose                                                         |
| ---------------------------------- | --------------------------------------------------------------- |
| [Playbook](PLAYBOOK.md)            | Installation, commands, workflows, telemetry, MCP and IDE usage |
| [Evidence](EVIDENCE.md)            | Validation results, benchmarks, screenshots and methodology     |
| [Architecture](docs/)              | Technical architecture and design                               |
| [Walkthrough](docs/WALKTHROUGH.md) | Dashboard and IDE workflow tour                                 |
| [Security](SECURITY.md)            | Security model and vulnerability reporting                      |
| [Contributing](CONTRIBUTING.md)    | Contribution and development workflow                           |
| [Changelog](CHANGELOG.md)          | Project history and implementation changes                      |

---

# Repository Structure

```text
garuda/
│
├── README.md
├── PLAYBOOK.md
├── EVIDENCE.md
├── SECURITY.md
├── CONTRIBUTING.md
├── CHANGELOG.md
├── LICENSE
│
├── assets/
│   ├── garuda-logo.png
│   └── screenshots/
│       ├── Dashboard_overview_workspace_active14repo_and_10_contradictions.png
│       ├── global_search.png
│       ├── repository-topology.png
│       ├── top_level_repo_architecture.png
│       ├── cross-repo-graph.png
│       ├── dashboard-14repo-contradictions.png
│       ├── evidence-cryptographic-trust.png
│       ├── ide-contradictions.png
│       └── ide-graph-view.png
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

# Early Adopters and Design Partners

Garuda is onboarding engineering teams running multi-service Go architectures with active AI coding workflows.

For pilots, experimentation, or design-partner discussions:

[GitHub Discussions](https://github.com/myshra777-ai/garuda/discussions)

---

# Contributing

Garuda is being built in the open.

Relevant areas include:

* compiler tooling
* Go analysis
* developer infrastructure
* semantic graphs
* runtime verification
* OpenTelemetry
* MCP
* AI coding agents
* software architecture
* provenance
* cryptographic state
* developer experience

Start with:

**[CONTRIBUTING.md](CONTRIBUTING.md)**

---

# Security

Security issues should be reported according to:

**[SECURITY.md](SECURITY.md)**

Garuda treats the following as first-class engineering concerns:

* provenance
* state integrity
* authentication boundaries
* evidence integrity
* auditability

---

# License

Garuda is released under the Apache License 2.0.

See [LICENSE](LICENSE).

---

# The Direction

Software systems are becoming increasingly distributed and dynamic.

Repositories multiply. Dependencies cross boundaries. Runtime behaviour can diverge from static expectations. AI agents increasingly modify code. Engineering knowledge is distributed across people, tools, conversations and time.

Garuda is being built around a different model:

```mermaid
flowchart LR
    SOFTWARE["SOFTWARE"]

    SOFTWARE --> CODE["CODE"]
    SOFTWARE --> RUNTIME["RUNTIME"]
    SOFTWARE --> INTENT["INTENT"]

    CODE --> GARUDA["GARUDA"]
    RUNTIME --> GARUDA
    INTENT --> GARUDA

    GARUDA --> EVIDENCE["EVIDENCE"]
    EVIDENCE --> VERIFY["VERIFICATION"]

    VERIFY --> HUMAN["HUMANS"]
    VERIFY --> AI["AI AGENTS"]
    VERIFY --> SYSTEMS["ENGINEERING SYSTEMS"]
```

---

# Garuda as Engineering Memory

Every engineering organization accumulates machine-relevant knowledge across:

```text
CODE
RUNTIME
DECISIONS
EVIDENCE
PULL REQUESTS
CHANGES
POLICIES
OBSERVATIONS
```

Garuda's long-term purpose is to preserve and connect that state in a structured, inspectable form.

```mermaid
flowchart TB
    CODE["Code"]
    RUNTIME["Runtime"]
    DECISIONS["Decisions"]
    EVIDENCE["Evidence"]
    PRS["Pull Requests"]
    CHANGES["Changes"]
    POLICIES["Policies"]
    OBS["Observations"]

    CODE --> GARUDA["GARUDA"]
    RUNTIME --> GARUDA
    DECISIONS --> GARUDA
    EVIDENCE --> GARUDA
    PRS --> GARUDA
    CHANGES --> GARUDA
    POLICIES --> GARUDA
    OBS --> GARUDA

    GARUDA --> HUMAN["Human"]
    GARUDA --> IDE["IDE"]
    GARUDA --> AI["AI"]
    GARUDA --> CI["CI / Governance"]
```

The long-term goal is not simply to make software easier to search.

It is to make software **easier to understand, verify, change, and reason about without repeatedly reconstructing the same system reality from scratch.**

---

# Current Scope and Evidence Boundary

Garuda's current semantic engine is **Go-specific**.

The implemented static analysis path uses:

* Go AST analysis
* `go/types`
* deterministic symbol identity
* Go-oriented semantic resolution

Multi-language support such as TypeScript, Rust, Python and Java is future direction, not current v0.2.0 functionality.

Runtime verification is also bounded by telemetry coverage and the observations available to Garuda.

A runtime path that has not been observed should remain `UNVERIFIED`, not be interpreted as dead or incorrect.

---

# Technical Disclaimer

This README describes Garuda according to the current implementation and documented validation runs.

### Shared Workspace

Multiple developers and AI agents share the same semantic state only when they are configured against the same Garuda workspace and PostgreSQL database.

Separate local databases are not automatically synchronized.

### Near-real-time behaviour

Telemetry admission is asynchronous and was measured at approximately `1.8 ms p95` under the tested workload.

Verification runs on asynchronous 10-second epochs.

Full dashboard and consensus propagation typically completed within `1–2 minutes` in the tested environment.

These are observed validation characteristics, not universal production SLAs.

### Performance and benchmark results

Precision, token reduction and workflow measurements derive from specific controlled environments.

Results can vary with:

* repository structure
* hardware
* infrastructure
* telemetry coverage
* network conditions
* workload
* AI model
* agent behaviour

### AI hallucination boundary

Garuda is designed to reduce unsupported reasoning and repeated context reconstruction.

The GAP-20 benchmark demonstrated **0% structural hallucination** for the evaluated benchmark tasks involving fabricated functions, receivers and symbols.

This does not claim elimination of all forms of AI hallucination.

### Cryptographic integrity

Cryptographic mechanisms provide tamper-evident state and verification.

They do not replace:

* credential security
* database security
* access controls
* key management
* operational security practices

### Regulatory references

Garuda may provide technical capabilities relevant to governance and compliance workflows.

Whether a deployment satisfies the EU AI Act or another legal or regulatory requirement depends on the deployment, organizational processes, risk classification and applicable law.

Garuda is not legal advice or a legal certification.

### Reproducibility

For reproducible validation, run the included benchmark and review the evidence artifacts in:

**[EVIDENCE.md](EVIDENCE.md)**

---

<p align="center">
  <strong>Garuda</strong><br>
  Analyze | Verify | Collaborate
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
    <strong>Version:</strong> v0.2.0 ·
    <strong>License:</strong> Apache 2.0
  </sub>
</p>
