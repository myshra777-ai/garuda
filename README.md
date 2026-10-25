<p align="center">
  <img src="assets/garuda-logo.png" width="104" alt="Garuda Logo">
</p>

<h1 align="center">Garuda</h1>

<p align="center">
  <strong>Let agents write. Garuda guards the architecture.</strong><br>
  <em>Garuda watches your architecture even when you aren't watching the diff — it maps, verifies, and gates every change before it merges.</em>
</p>

<p align="center">
  <a href="https://github.com/myshra777-ai/garuda/actions/workflows/ci.yml"><img src="https://github.com/myshra777-ai/garuda/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/myshra777-ai/garuda/actions/workflows/garuda-ci.yml"><img src="https://github.com/myshra777-ai/garuda/actions/workflows/garuda-ci.yml/badge.svg" alt="Garuda CI & Governance"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache--2.0-blue.svg" alt="Apache 2.0 License"></a>
  <a href="PLAYBOOK.md"><img src="https://img.shields.io/badge/docs-Operational_Playbook-0f172a.svg" alt="Operational Playbook"></a>
  <a href="EVIDENCE.md"><img src="https://img.shields.io/badge/evidence-38%2F38_MCP_%7C_10%2F10_Runtime-16a34a.svg" alt="Verification Evidence"></a>
</p>

## Why Garuda? (The Gap in Today's AI Tooling)

AI coding agents (Cursor, Claude Code, Devin, Codex) are incredible at writing code, but they are architecturally blind. A pull request can pass every unit test and linter check while quietly introducing a forbidden `Handler -> Database` dependency, creating a circular package dependency, breaking a shared struct with 170+ downstream callers, or drifting away from documented ADRs and live production behavior.

Here is what happens without Garuda:

| What You Use Today | What It Does | Where It Fails You |
| :--- | :--- | :--- |
| **Vector RAG / Embeddings**<br>(Cursor codebase index) | Guesses related files via semantic similarity. | Hallucinates dependencies, misses exact compiler types, has zero notion of rules. |
| **Linters & Unit Tests**<br>(golangci-lint, pytest) | Verifies syntax and isolated functional logic. | Lets forbidden imports slip by (e.g., Handler calling DB), blind to system-wide drift. |
| **Knowledge Graph Tools** | Shows you a static map of code entities. | Only tells you *what is*. Cannot gate CI, cannot catch runtime boundary bypasses. |
| **✨ GARUDA** | Deterministic AST mapping + Live OpenTelemetry traces + Fail-closed policy gates. | Gates bad code before merge. Proves compliance with cryptographic receipts. |

## What Garuda Can Do for You

- **Know Your Real Blast Radius Before Touching Code:** Ask Garuda about any struct, function, or service. It traces exact upstream callers, interface implementers, and downstream impacts in milliseconds—not by guessing with an LLM, but through compiler-verified ASTs.
- **Stop Architectural Decay in Pull Requests:** Write policies in simple YAML (e.g., "API handlers must never talk directly to SQL" or "Packages must have zero circular imports"). Garuda fails your CI build with exit code 2 if an agent or engineer violates them.
- **Catch "Shadow" Runtime Calls (Static vs. Production):** Static code might look clean, but reflection and dynamic dispatch bypass boundaries in production. Garuda ingests OpenTelemetry spans and flags undeclared runtime calls as `CONTRADICTED` in real time.
- **Give Agents Grounded Truth via 21 MCP Tools:** Stop agents from burning context tokens grepping through thousands of files. Garuda feeds Cursor and Claude exact symbol IDs, neighbors, and policy constraints directly through MCP.
- **Cryptographic Proof of Every Decision:** Every policy check and governance approval is anchored to an append-only RFC 6962 Merkle tree. You get non-repudiable proof of architectural compliance for security teams and compliance audits.

## The 5-Second Workflow

Boot the entire local stack, index your repository, and wire up your coding agents in one command:

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda
./setup.sh
```

Or run the unified change-control gate against any symbol or pull request target:

```bash
./bin/garuda check \
  --project my-system \
  --target PaymentService \
  --policy ./policies \
  --save \
  --fail-on-block \
  --json
```

**Here is what happens immediately:**
1. **Compiler & AST Extraction (`100%` Local, Zero LLM Guessing)** — Garuda parses your workspace using `go/types` + `go.work` + `vendor/` for Go, Python ASTs, and Tree-sitter TypeScript ASTs. No source code leaves your machine, and there is no embedding or vector-store hallucination.
2. **Transitive Blast Radius & Topology Mapping** — Computes confidence-weighted downstream callers, interface implementers, structural hubs, and package cycles in milliseconds.
3. **3-State Epistemic Truth (`SUPPORTED` / `UNVERIFIED` / `CONTRADICTED`)** — Correlates static code edges with ingested Markdown/ADR requirements and live OpenTelemetry runtime spans. Shadow runtime calls that bypass static boundaries are automatically quarantined.
4. **Fail-Closed Policy Evaluation (`ALLOW` / `WARN` / `REVIEW` / `BLOCK`)** — Runs 9 deterministic SQL/CTE policy predicates. With `--fail-on-block`, a `BLOCK` verdict exits with code `2` to fail your CI pipeline.
5. **RFC 6962 Merkle Audit Receipt** — When persisted with `--save`, the governance decision and its exact input hashes are anchored in an append-only SHA-256 Merkle tree so any auditor or CI job can cryptographically verify inclusion and tamper-evidence.

## Setup & Deployment Guide

Garuda adapts to whether you are an individual developer hacking locally or an enterprise team protecting a multi-repo monorepo.

### Path A: For Individual Developers (Local, Zero-Config)
Get up and running locally in under 60 seconds with Docker and your choice of IDE (Cursor, Claude Desktop, or VS Code):

```bash
# 1. Clone the repository
git clone https://github.com/myshra777-ai/garuda.git
cd garuda

# 2. Run the automated setup script
# (Starts Postgres, runs migrations, compiles binaries, & sets up MCP)
./setup.sh

# 3. Analyze your own project
./bin/garuda analyze /path/to/your/project --save --workspace my-app

# 4. Launch the local dashboard
./bin/garuda dev
```
Your Cursor and Claude Desktop environments will automatically have `garuda-mcp` registered. You can immediately ask your agent: *"What calls PaymentHandler and what policy rules apply?"*

### Path B: For Engineering Teams & Organizations
Standardize architecture policies across your entire engineering team and enforce gates on every pull request.

**1. Centralized Infrastructure (Postgres & Garuda Gateway)**
Deploy Garuda centrally so your team and CI runners share the same semantic model and Merkle ledger:

```bash
docker run -d \
  --name garuda-api \
  -p 8080:8080 \
  -e DATABASE_URL="postgres://user:pass@shared-db:5432/garuda?sslmode=require" \
  -e JWT_PRIVATE_KEY_HEX="<your-central-key>" \
  garuda/garuda-api:v1.0.0
```

**2. Team-Wide Repository Policies**
Check your architecture guardrails directly into Git under `policies/` so they are version-controlled alongside your code:

```yaml
id: no-handler-db-bypass
scope:
  package: "internal/handlers"
when:
  - type: layer_violation
    params:
      from_layer: "internal/handlers"
      to_layer: "internal/db"
then:
  decision: BLOCK
  reason: "Handlers must call services, not raw database packages."
```

**3. Enforce the Gate in GitHub Actions**
Add this step to your PR check workflow (e.g., `.github/workflows/architecture-gate.yml`) to automatically block PRs that violate policy:

```yaml
name: Garuda Architecture Gate
on: [pull_request]

jobs:
  verify-architecture:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install Garuda
        run: |
          curl -sSL https://get.garuda-engine.dev | bash
      - name: Run Change-Control Gate
        env:
          DATABASE_URL: ${{ secrets.GARUDA_DATABASE_URL }}
        run: |
          garuda check \
            --project ${{ github.event.repository.name }} \
            --policy ./policies \
            --fail-on-block
```

## Software Intelligence Instrument (`/dashboard`)

Your command deck when developing, reviewing changes, or debugging architectural drift. When you run `garuda dev` and navigate to `http://localhost:8080/dashboard`, you don’t get an overwhelming, unreadable graph of 10,000 nodes. Instead, hit `⌘K` / `Ctrl+K` to access six focused workspaces built for how you actually work:

**1. Workspace Briefing (`/dashboard#briefing`)**
- **What you see:** A real-time health readout of your codebase—indexed symbols, active policies, Merkle ledger height, and a prioritized "Needs Attention" queue.
- **Why it matters:** Instantly know if an agent introduced an architectural contradiction, broken interface, or unverified documentation claim before you start your day.

**2. Architecture Map (`/dashboard#map`)**
- **What you see:** A neighborhood-first visual explorer. Click any focal symbol to see immediate 1-hop and 2-hop dependencies: `[Upstream Callers] ──> [Focal Symbol] ──> [Dependencies / Policies]`.
- **Why it matters:** Filter cleanly by Calls, Imports, Implements, or Claims. Open the slide-out Inspector Drawer to see raw source lines, active relationships, and evidence backing every edge.

**3. Change Review Workbench (`/dashboard#workbench`)**
- **What you see:** A 3-pane pre-merge flight simulator. Pick your target symbol and policy bundle on the left, see impacted downstream callers in the center, and watch Garuda evaluate the gate on the right.
- **Why it matters:** Preview whether a proposed PR will trigger an `ALLOW`, `WARN`, `REVIEW`, or `BLOCK` verdict before pushing to GitHub.

**4. Interactive Policy Sandbox (`/dashboard#policies`)**
- **What you see:** An interactive catalog of all 9 deterministic SQL/CTE predicates (`layer_violation`, `cycle_exists`, `blast_radius_exceeds`, etc.) alongside a live YAML editor.
- **Why it matters:** Test new architectural guardrails in a safe sandbox against your live graph before committing them to CI.

**5. Contradictions & Runtime Evidence (`/dashboard#contradictions`)**
- **What you see:** A dedicated forensic queue highlighting where reality disagrees with your code or documentation.
- **Why it matters:** It compares static AST definitions with real OpenTelemetry trace spans (`trace_id`, `span_id`). If production traffic takes a route your architecture forbids, it gets quarantined here with full stack traces.

**6. Governance & Audit Ledger (`/dashboard#ledger`)**
- **What you see:** An immutable log of every governance decision, policy check, and Merkle root transition.
- **Why it matters:** Click any entry to inspect what the engine checked, which rule applied, its epistemic state (`SUPPORTED`, `UNVERIFIED`, `CONTRADICTED`), and verify its RFC 6962 inclusion proof with one click.

## 21 MCP Tools for Coding Agents

`./setup.sh` automatically registers `garuda-mcp` with **Cursor** and **Claude Desktop** (and prints ready-to-paste config for **Codex CLI**). Connected agents query the shared workspace model directly via 21 deterministic tools:

| Category | MCP Tools | Purpose |
|---|---|---|
| **Orientation & Discovery** | `garuda.briefing`, `garuda.entities`, `garuda.find_entity`, `garuda.inspect` | Orient the agent to workspace health, resolve exact symbol IDs, and inspect source signatures before editing. |
| **Topology & Blast Radius** | `garuda.neighbors`, `garuda.blast_radius`, `garuda.get_impact`, `garuda.implementers`, `garuda.subclasses`, `garuda.query_claims` | Trace upstream callers, downstream dependencies, interface implementations, and transitive change impact. |
| **Governance & Policy Gates** | `garuda.verify_policy_evaluation`, `garuda.policy.evaluate`, `garuda.governance.status`, `garuda.policy.list` | Pre-check proposed refactors against active YAML policies and verify RFC 6962 Merkle receipts. |
| **Epistemic Truth & Runtime** | `garuda.detect_contradictions`, `garuda.check_drift`, `garuda.get_lineage`, `garuda.propose_decision` | Query static-vs-runtime contradictions, ADR drift, and decision history. |
| **Multi-Agent Continuity** | `garuda.handoff`, `garuda.resume` | Persist and resume structured multi-agent task state, pinned symbols, and architectural constraints across sessions. |

## Evidence, Not False Certainty

Garuda keeps the boundaries of its evidence explicit so teams never mistake a clean check for magical omniscience:

- **`SUPPORTED`** — Available static and/or runtime evidence corroborates the claim within the verified scope.
- **`UNVERIFIED`** — Static structure exists, but runtime or secondary corroborating evidence has not yet been observed. It does *not* mean the relationship is broken.
- **`CONTRADICTED`** — Available evidence actively conflicts with a claim (for example, an ADR states `ServiceA` never calls `StoreB`, or an OpenTelemetry span observes an undeclared runtime call).

**What the RFC 6962 Merkle Ledger Proves (and What It Doesn't):**
The Merkle ledger cryptographically proves the **inclusion, ordering, and tamper-evidence of persisted governance records** (`--save`). It does **not** prove that an unindexed dynamic call cannot exist, that a human-written YAML policy is wise, or that the software is bug-free. Preview evaluations run without `--save` are ephemeral and are not anchored to the ledger.

## Supported Languages & Verification Benchmarks

| Language | Analysis Engine | Confidence & Scope | Status |
|---|---|---|---|
| **Go** | `go/types`, `go/ssa`, `go.work`, `vendor/` | Exact compiler-grade type, interface, and cross-module call resolution (`1.0` static confidence). | **Production** |
| **Python** | Structural AST Analyzer | Module, class, function, import, and call-graph extraction. | **Beta** |
| **TypeScript** | Tree-sitter AST Analyzer | Interface, class, function, import, and structural relationship extraction. | **Beta** |

- **Reproducible Verification Suite:** `make benchmark-strict` executes 24 ground-truth compiler and cross-module benchmarks (`TestDoubleRunDeterminism` across all 24 cases), `38/38` reference MCP verifier checks (`scripts/mcp_verify.py`), and `10/10` edge-level runtime contradiction detections (`scripts/runtime_contradiction_test.py`). See [EVIDENCE.md](EVIDENCE.md) for full methodology and scope boundaries.

## Documentation

- **[Operational Playbook](PLAYBOOK.md)** — 1-command onboarding, MCP client wiring (Cursor, Claude Desktop, Codex), policy authoring, runtime OpenTelemetry ingestion, and troubleshooting.
- **[System Walkthrough](docs/WALKTHROUGH.md)** — End-to-end CLI, v1.4 Dashboard, MCP, and VS Code extension workflows.
- **[Capabilities & Verification Tiers](docs/CAPABILITIES.md)** — Detailed breakdown of production vs. beta capabilities and verification guarantees.
- **[CLI, Policy, MCP & HTTP API Reference](docs/API.md)** — Complete reference for all CLI flags, 9 policy predicates, 21 MCP tools, and REST endpoints.
- **[System Specifications](docs/SPECS.md)** — Architecture internals, PostgreSQL schema, Ed25519 JWT auth, and Merkle tree specifications.
- **[Verification Evidence](EVIDENCE.md)** — Reproducible benchmark output, test coverage, and explicit security/epistemic boundaries.
- **[Changelog](CHANGELOG.md)** · **[Codebase Hygiene](docs/hygiene.md)** · **[License (Apache 2.0)](LICENSE)**