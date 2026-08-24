```markdown
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
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.22+-00ADD8.svg?style=flat-square&logo=go" alt="Go Version"></a>
  <a href="EVIDENCE.md"><img src="https://img.shields.io/badge/verification-14%20repos%20passed-success.svg?style=flat-square" alt="Verification Status"></a>
  <a href="EVIDENCE.md"><img src="https://img.shields.io/badge/hallucinations-0.0%25%20(GAP--20)-purple.svg?style=flat-square" alt="GAP-20 Grounding"></a>
</p>

<p align="center">
  Garuda aligns humans and AI agents around a shared, evidence-backed understanding of a software system.
</p>

<p align="center">
  <a href="PLAYBOOK.md">Playbook</a> ·
  <a href="EVIDENCE.md">Evidence</a> ·
  <a href="docs/">Architecture</a> ·
  <a href="SECURITY.md">Security</a> ·
  <a href="CHANGELOG.md">Changelog</a>
</p>

---

## The Reality of Modern Engineering

Software development is no longer human-only. Engineering teams run heterogeneous fleets of AI coding agents alongside human developers across dozens of microservices.

```text
                      THE MULTI-AGENT CONTEXT VACUUM
┌────────────────────────────────────────────────────────────────────────┐
│  Dev A (Claude Code)    Dev B (Cursor / GPT-4)    Dev C (Gemini Code)  │
│         │                         │                        │           │
│  Refactors Auth Core     Adds Webhook Handler     Touches Database API │
│         └───────────────┬─────────┴────────────────────────┘           │
│                         ▼                                              │
│        ❌ Probabilistic LLM Assumptions & Context Drift                │
│                         │                                              │
│                         ▼ (With Garuda)                                │
│        🦅 GARUDA EPISTEMIC TRUTH GRAPH & MERKLE LEDGER                 │
│        • Compiler-verified symbols                                     │
│        • Evidence-backed context                                       │
│        • Transitive Blast-Radius Resolution across 14+ Repos           │
│        • Near-real-time OpenTelemetry verification & contradiction quarantine │
└────────────────────────────────────────────────────────────────────────┘
```

---

## The Problem: The Multi-Agent Tower of Babel

When a team of 20 engineers deploys heterogeneous AI coding fleets (Claude Code, Cursor, GPT, Gemini), software architecture fractures across three distinct failure modes.

### 1. The Multi-Agent Context Vacuum

**The Scenario:** Dev A uses Claude to refactor the payment gateway. Ten minutes later, Dev B asks Cursor to add a webhook handler. Because Dev B's AI runs in a stateless silo without compiler awareness of Dev A's local refactor, it generates code referencing deprecated structs and deleted database handles.

**The Cost:** Silent merge conflicts, broken cross-package assumptions, and failed integration builds.

### 2. The Runaway AI Token Invoicing Crisis

**The Scenario:** Dev A spends 25,000 prompt tokens having Claude map the dependency tree of an un-instrumented auth service. Two days later, Dev C tasks GPT with investigating a bug in the same service—burning another 25,000 tokens rediscovering the exact same call paths.

**The Cost:** Thousands of dollars in duplicate LLM API bills rediscovering static AST structures that should be computed once and persisted.

### 3. The "14-Day Sick Leave" & Onboarding Catch-up Tax

**The Scenario:** A senior engineer returns from two weeks of leave to 50 merged PRs and outdated Notion documentation. An intern spends two months reading code before making their first safe commit, only to leave the team shortly after.

**The Cost:** Days lost to manual code archaeology, repetitive architecture sync meetings, and high developer friction.

---

## What Is Garuda?

Garuda is an **evidence-backed software intelligence and shared intelligence workspace**.

It analyzes software repositories and builds a structured representation of:

- entities
- relationships
- dependencies
- source evidence
- semantic state
- runtime observations
- contradictions
- decisions
- lineage
- revisions
- topology
- verification state

That state can then be consumed by:

- developers
- IDEs
- CI
- dashboards
- APIs
- MCP
- AI agents

The goal is simple:

> **Humans and AI agents should not have to rediscover the software system independently every time they work on it.**

Instead:

```text
                         TEAM
                          │
            ┌─────────────┼─────────────┐
            │             │             │
         Developer     Developer      Developer
            │             │             │
          Claude         GPT          Gemini
            │             │             │
            └─────────────┼─────────────┘
                          │
                          ▼
                   GARUDA WORKSPACE
                          │
             ┌────────────┼────────────┐
             │            │            │
         Semantic       Evidence     Runtime
           State                     Observations
             │            │            │
             └────────────┼────────────┘
                          │
                          ▼
                    Verification
                          │
             ┌────────────┼────────────┐
             │            │            │
          Supported    Unverified   Contradicted
             │            │            │
             └────────────┼────────────┘
                          ▼
                 Shared Understanding
```

Garuda is therefore **not simply a graph generator**.

The graph is one representation of the underlying semantic state.

The larger objective is to create a persistent, inspectable and verifiable understanding of the software system that both humans and machines can use.

---

## The Core Triad: Analyze · Verify · Collaborate

Garuda is organized around three fundamental operations.

```text
            GARUDA: FROM COMPILER TO RUNTIME PRODUCTION
┌──────────────────┐      ┌──────────────────┐      ┌──────────────────┐
│     ANALYZE      │      │      VERIFY      │      │   COLLABORATE    │
│ Compiler-Grade   │ ──▶  │ Continuous Dual  │ ──▶  │ Shared Machine   │
│ Go AST Parsing   │      │ Merkle Ledger    │      │ Memory via MCP   │
└──────────────────┘      └──────────────────┘      └──────────────────┘
```

| Capability | What Garuda does | Why it matters |
| ---------- | ---------------- | -------------- |
| **Analyze** | Extracts entities, relationships, dependencies and evidence from source | Builds a structured understanding of the codebase |
| **Verify** | Correlates static structure with observations and identifies contradictions | Distinguishes what is known from what is merely assumed |
| **Collaborate** | Makes the resulting state available to developers and AI agents | Reduces repeated discovery and context fragmentation |

---

## How It Works: The 6 Epistemic Layers

Garuda partitions software knowledge into 6 strict epistemological layers to prevent guesswork.

```text
┌─────────────────────────────────────────────────────────────────────────┐
│ LAYER 6: LINEAGE      ── How the system arrived at its current state    │
├─────────────────────────────────────────────────────────────────────────┤
│ LAYER 5: DECISIONS    ── Recorded architectural rationale & policies    │
├─────────────────────────────────────────────────────────────────────────┤
│ LAYER 4: VERIFICATION ── SUPPORTED / UNVERIFIED / CONTRADICTED state    │
├─────────────────────────────────────────────────────────────────────────┤
│ LAYER 3: EVIDENCE     ── Exact source files, line numbers, and revisions│
├─────────────────────────────────────────────────────────────────────────┤
│ LAYER 2: CLAIMS       ── Inferred call graphs & interface relationships │
├─────────────────────────────────────────────────────────────────────────┤
│ LAYER 1: OBSERVATIONS ── Raw Go AST tokens & live OpenTelemetry spans   │
└─────────────────────────────────────────────────────────────────────────┘
```

### 1. Observation

"What does the source code contain?" — Extracted from AST. Raw, unverified.

### 2. Claim

"What does this imply about the system?" — Structured relationships between observations.

### 3. Evidence

"Why do we believe this claim?" — Source files, lines, revisions, runtime traces.

### 4. Verification

"What can we confirm?" — SUPPORTED / UNVERIFIED / CONTRADICTED

### 5. Decision

"What did the team agree on?" — Architectural decisions, policies, intent.

### 6. Lineage

"How did we get here?" — History of decisions, proposals, supersessions.

---

## The Epistemic Tri-State Model

**The Garuda Axiom: Absence of evidence is not evidence of absence.**

### Verification States

| State | Meaning |
|-------|---------|
| `SUPPORTED` | Static declarations and live runtime telemetry agree |
| `UNVERIFIED` | The static entity exists in code, but no runtime traffic has been observed executing through it yet (it is not flagged as dead code) |
| `CONTRADICTED` | Runtime telemetry directly violates static contracts or architectural policies (quarantined as ARCH_DRIFT_001) |

```text
                 SOURCE CODE
                      │
                      ▼
               STATIC ANALYSIS
                      │
                      ▼
               SEMANTIC STATE
                      │
          ┌───────────┴───────────┐
          │                       │
          ▼                       ▼
       CLAIMS                  EVIDENCE
          │
          │
          ▼
   RUNTIME OBSERVATIONS
          │
          ▼
     VERIFICATION
          │
     ┌────┼────┐
     ▼    ▼    ▼
 SUPPORTED  UNVERIFIED  CONTRADICTED
```

This distinction matters:

- **UNVERIFIED** does not mean "broken" or "dead"
- **UNVERIFIED** means "insufficient runtime evidence to claim verification"
- **CONTRADICTED** means "observed runtime behavior conflicts with static expectations"

---

## From Source Code to Shared Engineering State

```text
SOURCE
  │
  ▼
Go AST + go/types
  │
  ▼
Canonical Entities
  │
  ▼
Semantic Relationships
  │
  ▼
Evidence + Revision
  │
  ├───────────────┐
  ▼               ▼
STATIC STATE    RUNTIME OBSERVATION
  │               │
  └───────┬───────┘
          ▼
      VERIFICATION
          │
     ┌────┼────┐
     ▼    ▼    ▼
SUPPORTED UNKNOWN CONTRADICTED
          │
          ▼
     SHARED WORKSPACE
          │
     ┌────┼────────┐
     ▼    ▼        ▼
   HUMAN  IDE      AI
```

---

## The Verification Flow

```mermaid
flowchart TD
    SOURCE["SOURCE CODE"] --> STATIC["STATIC ANALYSIS"]
    STATIC --> GRAPH["SEMANTIC GRAPH"]
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

    SUPPORTED --> TRUSTED["TRUSTED"]
    UNVERIFIED --> UNKNOWN["UNKNOWN"]
    CONTRADICTED --> QUARANTINED["QUARANTINED"]
```

### Technical Precision

- **Telemetry admission:** HTTP `202 Accepted` — asynchronous
- **Measured ingestion response:** approximately `1.8 ms p95` under tested workload
- **Verification:** asynchronous (10-second epoch cycles)
- **Dashboard/IDE propagation:** typically `1–2 minutes` for full consensus convergence

These are **measured observations from current validation**, not universal production SLAs.

---

## The IDE Experience

Garuda integrates directly into VS Code and Cursor via the Language Server Protocol and Model Context Protocol.

### Near-Real-Time Contradiction Detection

When runtime telemetry or static changes violate architectural policy, the exact line of code is flagged immediately in the editor and Problems panel:

<p align="center">
  <img src="assets/screenshots/ide-contradictions.png" width="850" alt="Garuda Inline Contradiction Detection">
</p>

### Instant Blast Radius on Symbol Hover

Hovering over any struct, method, or interface reveals recursive upstream callers, downstream dependencies, and cross-repo bridges:

<p align="center">
  <img src="assets/screenshots/ide-graph-view.png" width="850" alt="Garuda Symbol Blast Radius Hover">
</p>

### Hover Context Example

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

### Interactive Topology Visualizer

The background daemon (`garuda dev`) hosts an interactive D3 dark-mode force-directed graph on `http://localhost:8080/graph`.

Want a complete visual tour of all 10 dashboard screens and IDE workflows? See [docs/WALKTHROUGH.md](docs/WALKTHROUGH.md).

---

## The Workspace: One Truth for All

```mermaid
flowchart TD
    TEAM["TEAM OF 20 DEVELOPERS + 20 AI AGENTS"]
    TEAM --> WS["GARUDA WORKSPACE"]

    WS --> CLAUDE["CLAUDE"]
    WS --> GPT["GPT"]
    WS --> GEMINI["GEMINI"]
    WS --> CURSOR["CURSOR"]
    WS --> HUMANS["HUMANS"]

    CLAUDE --> TRUTH["SAME ENTITIES
    SAME RELATIONSHIPS
    SAME EVIDENCE
    SAME CONTRADICTIONS
    SAME DECISIONS
    SAME RUNTIME TRUTH"]

    GPT --> TRUTH
    GEMINI --> TRUTH
    CURSOR --> TRUTH
    HUMANS --> TRUTH
```

- ✅ Humans see the same thing
- ✅ Claude sees the same thing
- ✅ GPT sees the same thing
- ✅ Gemini sees the same thing
- ✅ Cursor sees the same thing

**One workspace. One truth. All agents.**

> *For multi-agent sync to work, all developers and AI agents must point their MCP configuration to the same Garuda workspace and PostgreSQL database.*

---

## Why This Matters for AI Agents

AI agents are powerful precisely because they can reason, explore and generate code.

But repeatedly asking an agent to reconstruct a large repository from raw source is inefficient.

The same repository structure can be rediscovered again and again:

```text
Agent A
   ↓
Reads repository
   ↓
Builds mental model
   ↓
Investigates task

Agent B
   ↓
Reads repository
   ↓
Builds mental model
   ↓
Investigates same system

Agent C
   ↓
Reads repository
   ↓
Builds mental model
   ↓
Investigates same system
```

Garuda changes the flow:

```text
                 GARUDA
                    │
       ┌────────────┼────────────┐
       │            │            │
   Structure     Evidence      State
       │            │            │
       └────────────┼────────────┘
                    │
                    ▼
              Precise context
                    │
       ┌────────────┼────────────┐
       ▼            ▼            ▼
    Claude         GPT         Gemini
```

### The Duplicate Investigation Problem

Consider a payment service failure.

**Developer A** asks Claude to investigate. Claude explores the repository. Discovers the architecture. Identifies the issue. Fixes it. Records the resulting state.

**Developer C** — two days later, receives a related task. Without shared semantic state:

```text
Developer C
    ↓
GPT
    ↓
Read repository
    ↓
Rediscover architecture
    ↓
Rediscover dependency chain
    ↓
Rediscover previous reasoning
    ↓
Possibly propose an already rejected approach
```

With a shared Garuda workspace:

```text
Developer C
    ↓
GPT
    ↓
Garuda MCP
    ↓
Existing semantic state
    ↓
Evidence
    ↓
Previous decisions
    ↓
Current contradictions
    ↓
Focused investigation
```

The goal is **preventing unnecessary rediscovery**.

---

## Decisions Are Part of Software State

Code is not the only thing that changes.

Teams continuously make decisions:

- which architecture to use
- which approach to reject
- why a dependency exists
- why a service boundary was introduced
- which implementation was replaced
- which proposal was discarded
- which policy should apply
- which issue remains unresolved

If those decisions remain only inside a person's memory or an AI conversation, the rest of the team has no reliable way to recover them.

Garuda therefore treats decisions and their lineage as part of the broader semantic workspace.

```text
PROPOSAL
   │
   ▼
DECISION
   │
   ├── Evidence
   ├── Context
   ├── Revision
   └── Author / Actor
         │
         ▼
     SUPERSEDED?
         │
      ┌──┴──┐
      │     │
     NO    YES
            │
            ▼
       New Decision
            │
            ▼
         Lineage
```

This is particularly important when multiple AI agents are operating simultaneously.

An agent should not repeatedly propose an approach that the team has already considered and rejected when that decision is available to the workspace.

---

## MCP and AI Agents

Garuda exposes its semantic and evidence model to AI agents through Model Context Protocol workflows.

```mermaid
flowchart TD
    AGENT["AI AGENT"] --> MCP["MCP"]
    MCP --> GARUDA["GARUDA"]
    GARUDA --> SEMANTICS["SEMANTICS"]
    GARUDA --> EVIDENCE2["EVIDENCE"]
    GARUDA --> STATE["STATE"]
    SEMANTICS --> CONTEXT["GROUNDED CONTEXT"]
    EVIDENCE2 --> CONTEXT
    STATE --> CONTEXT
```

Examples of structured questions an agent can ask:

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

## Automated CI/CD Governance & PR Gating

Garuda acts as an automated architectural gatekeeper in your continuous integration pipeline via `garuda ci` and `garuda judge`.

```bash
# Evaluate proposed PR snapshot against baseline Merkle ledger state
garuda judge baseline.json proposed.json
```

```text
               PULL REQUEST LIFECYCLE WITH GARUDA
┌─────────────────────────────────────────────────────────────┐
│ 1. Developer / AI Agent opens Pull Request                  │
│ 2. GitHub Actions runs `garuda ci`                          │
│ 3. AST Diff computed against active Merkle baseline         │
│ 4. Transitive blast radius mapped across all 14 repos       │
│                                                             │
│ ❌ BLOCKED: Breaking contract in `payment/service.go:42`    │
│    (Violates Policy: POL_004 - Deprecated Receiver)         │
│                                                             │
│ 5. Sticky Markdown audit report posted directly to PR       │
└─────────────────────────────────────────────────────────────┘
```

### CI Capabilities

- **Contract Breakage Isolation:** Blocks pull requests that alter exported interfaces without updating downstream consumers across the workspace
- **Drift Prevention:** Quarantines pull requests that introduce unauthorized network dependencies or unmapped driver invocations
- **Automated Blast-Radius PR Reports:** Posts sticky comments on GitHub PRs detailing upstream and downstream services affected by the change

---

## Business & Engineering Impact

| Engineering Workflow | Traditional Manual Approach | With Garuda Workspace | Measured Advantage |
| :--- | :--- | :--- | :--- |
| **New Engineer Onboarding** | 3–6 weeks reading docs & tracing code | **2–3 days** via interactive topology & blast radius | **~80% faster time-to-first-PR** |
| **Post-Leave Catch-up** | 3–5 days of meetings & PR diff archaeology | **~15 minutes** reviewing `garuda summary` & hubs | **Zero team sync meetings required** |
| **PR Architectural Review** | 45 min manual verification per PR | **0 min** (Automated CI gate blocks drift) | **100% breaking contract isolation** |
| **Incident Blast-Radius Analysis** | 30–60 min running `grep` and asking in Slack | **Instant symbol hover** in VS Code / Cursor | **Sub-second dependency lookup** |
| **AI Agent Context Overhead** | 4,850 tokens (raw file dumps) | **620 tokens** (verified AST subgraphs) | **87.2% token cost reduction** |

*These represent measured outcomes from controlled benchmarks. Actual results may vary.*

---

## Traditional Approach vs. Garuda

| Traditional Approach | Garuda Epistemic Workspace |
|----------------------|---------------------------|
| ❌ Each AI re-discovers context from raw text | ✅ Shared, compiler-verified semantic workspace |
| ❌ Agents hallucinate signatures and receivers | ✅ 100% symbol precision grounded via MCP |
| ❌ No team-wide architectural consistency | ✅ One deterministic source of truth for humans and AI |
| ❌ Wasted tokens on full-repo prompt dumps | ✅ 87.2% context token compression (GAP-20) |
| ❌ Weeks of developer onboarding | ✅ Immediate architectural navigation |
| ❌ Blind to runtime behavior in production | ✅ Native OpenTelemetry trace ingestion and correlation |
| ❌ No auditable decision lineage | ✅ Immutable dual-root Merkle ledger (Block #898+) |
| ❌ Silent architectural drift in production | ✅ Quarantined violations (ARCH_DRIFT_001) |

---

## The Evidence: 14 Repositories

Garuda has been tested against **14 heterogeneous Go repositories** (including Garuda itself) in controlled test environments.

| Metric | Result |
|--------|--------|
| Repositories | 14 |
| Packages | 143 |
| Entities | 3,675 |
| Relationships | 5,679 |
| Cross‑repository bridges | 55 |
| Contradictions injected | 10 |
| Contradictions detected | 10/10 |
| Entity precision | 100% |
| Relationship precision | 99.9% |
| Token reduction (GAP-20 benchmark) | ~87.2% |
| Merkle Ledger Height | Block #898+ |

*Results from controlled test environments. Actual results may vary depending on hardware, AI models, repository size, and infrastructure.*

### GAP-20 Grounding Benchmark (vs. Unassisted LLMs)

| Metric | Naive LLM | Garuda-Grounded | Improvement |
|--------|-----------|-----------------|-------------|
| Symbol Precision | 40.0% | 100.0% | +150% |
| Structural Hallucination Rate | 66.7% | 0.0% | -66.7% |
| Prompt Token Overhead | 4,850 tokens | 620 tokens | 87.2% reduction |
| Upstream Caller Recall | 20.0% | 100.0% | +80% |
| Downstream Dep Recall | 33.0% | 100.0% | +67% |
| Violation Quarantine Rate | 0.0% | 100.0% | +100% |

**Key takeaway:** Garuda-grounded agents achieved 100% symbol precision and 0% structural hallucination in the evaluated benchmark. Structural hallucinations — fabricated functions, receivers, and symbols — were completely eliminated.

For complete methodology, screenshots, and historical runs: **[→ Open the Garuda Evidence Center](EVIDENCE.md)**.

---

## Capability Status

Garuda is evolving quickly. Capabilities are grouped by maturity.

### Stable

Core capabilities that form the current semantic foundation:

- Go semantic analysis
- Compiler‑backed type resolution
- Deterministic entity identity
- Semantic relationships
- Repository and workspace modelling
- Evidence provenance
- Semantic snapshots
- Semantic diff
- Impact analysis
- Impact‑diff
- Lineage
- Cryptographic state verification
- Global workspace search
- Progressive architecture exploration
- CLI workflows
- Dashboard exploration

### Early Testing

Capabilities that are implemented and actively being validated:

- Multi‑repository intelligence
- OpenTelemetry ingestion
- Runtime observations
- Static ↔ runtime correlation
- Contradiction verification
- Runtime evidence views
- MCP integration
- Garuda IDE
- Grounding benchmark harness
- Agent‑oriented workflows
- **CI integration (`garuda ci`, `garuda judge`)**
- **PR governance and policy automation**

These capabilities should be treated as **early testing rather than universal production guarantees**.

### Launching Soon

The next product‑facing layer is focused on making Garuda easier to operate continuously:

- Richer runtime verification
- Deeper MCP workflows
- Expanded evidence exploration
- Broader grounding benchmarks
- Production‑oriented telemetry workflows
- Additional developer workflow integrations

### Longer‑term Direction

Garuda is being developed toward a broader software intelligence substrate capable of connecting source, runtime, evidence, decisions, policies, AI agents, operational state, and governance:

- Company‑scale software graphs
- Multi‑language semantic intelligence
- Richer runtime correlation
- Stronger architectural governance
- AI‑grounded software operations
- Business‑state integrity
- Autonomous verification and remediation

*These are directions, not current production claims.*

---

## Design Principles

1. **Evidence over inference** — Prefer inspectable evidence over opaque conclusions
2. **Unknown stays unknown** — Do not turn missing evidence into certainty
3. **Identity must be deterministic** — The same semantic entity should remain identifiable across analysis runs
4. **State should be auditable** — Important state transitions should leave an inspectable trail
5. **Architecture should be navigable** — A useful graph should help humans move from macro architecture to specific evidence
6. **Runtime and static knowledge are different** — Static structure tells us what the code contains; runtime observations tell us what has been observed executing. They should not be conflated
7. **AI should consume structured truth** — Agents should not have to rediscover the entire software system every time they work on it
8. **Correctness before scale** — Garuda's architecture favours validated semantics and evidence before expanding scope

---

## What Garuda Is Not

Garuda is not trying to be:

- Another text search engine
- Another code linter
- Another generic dependency graph
- Another observability platform
- Another documentation generator
- Another generic vector database
- An LLM that guesses how your repository works

Garuda sits between these layers. It is intended to become the **semantic and evidence substrate connecting them**.

### A Graph Is Only One View of Garuda

A dependency graph answers: **"What connects to what?"**

Garuda is designed to answer:

- What exists?
- What depends on it?
- What evidence supports that relationship?
- What has actually been observed?
- What contradicts the expected architecture?
- What decision led to the current design?
- What changed?
- Who or what superseded the previous decision?
- What should an AI agent know before modifying this component?

---

## Quick Start

```bash
# One-line installation
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | sh

# Initialize Garuda
garuda init

# Start the stack
garuda up

# Analyze a repository
garuda analyze .

# Start unified daemon (API, OTel Collector, Merkle Worker, UI)
garuda dev

# Open the interactive dark-mode topology visualizer
open http://localhost:8080/graph
```

### Configure MCP for Cursor & Claude Desktop

Add to `.cursor/mcp.json` or `claude_desktop_config.json`:

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

For detailed installation and operational workflows: **[→ Read the Garuda Playbook](PLAYBOOK.md)**.

---

## Documentation

| Document | Purpose |
|----------|---------|
| [Playbook](PLAYBOOK.md) | Installation, commands, workflows, telemetry, MCP and IDE usage |
| [Evidence](EVIDENCE.md) | Validation results, benchmarks, screenshots, methodology |
| [Architecture](docs/) | Technical architecture and design documentation |
| [Security](SECURITY.md) | Security model and vulnerability reporting |
| [Contributing](CONTRIBUTING.md) | Contribution and development workflow |
| [Changelog](CHANGELOG.md) | Project history and implementation changes |

---

## Repository Structure

```
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
│   ├── screenshots/
│   │   ├── Dashboard_overview_workspace_active14repo_and_10_contradictions.png
│   │   ├── global_search.png
│   │   ├── repository-topology.png
│   │   ├── top_level_repo_architecture.png
│   │   ├── cross-repo-graph.png
│   │   ├── dashboard-14repo-contradictions.png
│   │   ├── evidence-cryptographic-trust.png
│   │   ├── ide-contradictions.png
│   │   └── ide-graph-view.png
│   │
│   └── ...
│
├── docs/
├── cmd/
├── internal/
├── migrations/
├── garuda-bench/
├── test/
└── scripts/
```

---

## Early Adopters & Design Partners

Garuda is currently onboarding select engineering teams running multi-service Go architectures with active AI coding fleets.

If your team is experiencing context drift or runaway AI token spend across microservices, open an issue or reach out via [GitHub Discussions](https://github.com/myshra777-ai/garuda/discussions) to set up a shared workspace pilot.

---

## Contributing

Garuda is being built in the open. If you are interested in compiler tooling, developer infrastructure, semantic graphs, runtime verification, OpenTelemetry, MCP, AI coding agents, software architecture, provenance systems, cryptographic state, or developer experience, there are many interesting problems to work on.

Start with: **[CONTRIBUTING.md](CONTRIBUTING.md)**

---

## Security

Security issues should be reported according to **[SECURITY.md](SECURITY.md)**. Garuda treats provenance, state integrity, authentication boundaries, and evidence integrity as first‑class engineering concerns.

---

## License

Garuda is released under the Apache License 2.0. See [LICENSE](LICENSE).

---

## The Direction

Software is becoming increasingly dynamic. Repositories multiply. Dependencies cross organizational boundaries. Runtime behaviour diverges from assumptions. AI agents increasingly modify production systems. And every new agent creates another potential source of duplicated context and reasoning.

Garuda is being built around a different model:

```mermaid
flowchart TD
    SW["SOFTWARE"] --> CODE["CODE"]
    SW --> RUNTIME2["RUNTIME"]
    SW --> INTENT["INTENT"]
    CODE --> GRAPH3["GARUDA GRAPH"]
    RUNTIME2 --> GRAPH3
    INTENT --> GRAPH3
    GRAPH3 --> EVIDENCE3["EVIDENCE"]
    EVIDENCE3 --> VERIFY3["VERIFICATION"]
    VERIFY3 --> HUMAN["HUMAN"]
    VERIFY3 --> AI["AI"]
    VERIFY3 --> SYSTEMS["SYSTEMS"]
```

### Garuda as Engineering Memory

Every engineering organization accumulates knowledge that exists temporarily inside:

- human memory
- Slack conversations
- pull requests
- local AI conversations
- code reviews
- incident discussions
- architecture meetings
- discarded approaches
- previous implementations

Much of it disappears when:

- a conversation ends
- an engineer changes teams
- an engineer goes on leave
- an AI agent starts a new session
- documentation becomes stale

Garuda is designed to preserve the **machine-relevant portion of that engineering state** as structured, inspectable data.

```text
                    ENGINEERING MEMORY

     CODE ────────┐
     RUNTIME ─────┤
     DECISIONS ───┤
     EVIDENCE ────┤
     PRs ─────────┤
     CHANGES ─────┤
     POLICIES ────┤
     OBSERVATIONS ┘
             │
             ▼
         GARUDA
             │
      ┌──────┼──────┐
      ▼      ▼      ▼
    HUMAN   IDE     AI
```

The long‑term goal is not simply to make software easier to search. It is to make software **easier to understand, verify, change, and reason about without repeatedly reconstructing the same reality from scratch.**

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

---

## Disclaimer

This README describes Garuda based on the current implementation and documented validation runs.

**Technical Precision:**

- **Multi-Agent Sync:** Dev A (Claude) and Dev B (Cursor) share the same ground truth only when both point their MCP configuration to the same Garuda workspace and PostgreSQL database. Isolated local databases are not automatically synced.

- **Near Real‑Time vs. Instantaneous:** Span ingestion is immediate (~1.8ms HTTP 202). Full ledger re‑verification and dashboard propagation happen on 10‑second epoch cycles, with full consensus convergence typically completing within 1–2 minutes. We describe this as **near real‑time**, not instantaneous.

- **Language Scope:** Garuda's semantic engine is currently built for Go codebases. AST extraction, `go/types` type checking, and symbol resolution are Go‑specific. Multi‑language support (TypeScript, Rust, Python, Java) is on the roadmap and not currently available in v0.2.0.

Performance figures, precision measurements, token reductions and workflow results are derived from specific test environments and controlled workloads. Actual results may vary depending on repository structure, hardware, infrastructure, telemetry coverage, workload, network conditions, AI model and agent behavior.

Forward-looking capabilities represent development direction and are not guarantees of future functionality.

Garuda is designed to reduce hallucinations, duplicated context reconstruction, architectural drift and engineering friction. It does not claim to eliminate these problems entirely. The GAP-20 benchmark demonstrates **structural hallucination elimination** — no fabricated functions, receivers, or symbols were observed in the evaluated benchmark tasks. This does not claim elimination of all forms of AI hallucination.

Cryptographic mechanisms provide tamper-evident state and verification but do not eliminate the need for proper security controls, credentials, database protection and key management.

EU AI Act and other regulatory references describe technical capabilities that may support governance and compliance workflows. Whether a particular deployment satisfies applicable legal obligations depends on the deployment, risk classification, organizational processes and applicable law. Garuda is not legal advice or a legal certification.

For reproducible validation, run the included benchmark and review the evidence artifacts in [`EVIDENCE.md`](EVIDENCE.md).

---

<p align="center">
  <sub>
    <strong>Version:</strong> v0.2.0 ·
    <strong>License:</strong> Apache 2.0
  </sub>
</p>
```
