<p align="center">
  <img src="assets/garuda-logo.png" width="96" alt="Garuda Logo">
</p>

<h1 align="center">Garuda</h1>

<p align="center">
  <strong>Architecture guardrails and change control for AI-written code.</strong><br>
  Give coding agents an accurate map of your codebase before they write, and block changes that break your architecture before they merge. Same input, same verdict, every time.
</p>

<p align="center">
  <a href="https://github.com/myshra777-ai/garuda/releases"><img src="https://img.shields.io/badge/release-v1.2.0-blue.svg?style=flat-square" alt="Release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache_2.0-green.svg?style=flat-square" alt="License"></a>
  <a href="PLAYBOOK.md"><img src="https://img.shields.io/badge/Playbook-Operational%20Guide-orange.svg?style=flat-square" alt="Playbook"></a>
  <a href="docs/CAPABILITIES.md"><img src="https://img.shields.io/badge/Capabilities-Verification%20Matrix-teal.svg?style=flat-square" alt="Capabilities"></a>
  <a href="EVIDENCE.md"><img src="https://img.shields.io/badge/Benchmarks-View%20Evidence-purple.svg?style=flat-square" alt="Benchmarks"></a>
</p>

---

## The Problem

AI coding agents (Cursor, Claude Code, Codex, Devin) write code faster than humans can review it. Each pull request can look clean on its own while quietly damaging your system architecture:

- An HTTP handler starts calling or importing the database layer directly (`Handler -> Helper -> Database`).
- Three packages end up depending on each other in a cycle (`A -> B -> C -> A`).
- A core shared struct with 170+ downstream callers is modified without architectural review.
- Production OpenTelemetry traces show runtime calls to unapproved services that do not exist in the static codebase.

Unit tests pass. The diff looks small. A busy reviewer approves it. Six months later, nobody can change anything without breaking something else.

**Garuda catches this deterministically — while the agent writes, in your editor, and in CI.**

---

## How It Works

Garuda connects **software state + blast radius + runtime evidence + policy + cryptographic history** using the same rules everywhere:

| Where | What Happens |
| :--- | :--- |
| **1. While the agent writes (MCP)** | Your coding agent queries `garuda-mcp` (21+ tools) for exact callers, implementers, subclasses, and blast radius instead of guessing imports. |
| **2. In your editor (VS Code)** | The Garuda VS Code extension flags architectural drift, runtime contradictions, and blast radius before you commit. |
| **3. In CI / Pre-Merge (`garuda check`)** | `garuda check` evaluates declarative YAML policies against the semantic graph, outputs a Change-Control Decision (`ALLOW`, `WARN`, `REVIEW`, `BLOCK`), and seals an RFC 6962 Merkle receipt. |

---

## 1-Command Quickstart

Garuda includes a zero-config setup script that starts PostgreSQL (via Docker if needed), builds all binaries, persists `.env` and Ed25519 JWT keys automatically, runs idempotent migrations, provisions your local owner account, configures Cursor/Claude MCP, and verifies the stack:

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda
./setup.sh
```

Or initialize any repository directly with the CLI:

```bash
./bin/garuda init /path/to/your/repo --workspace my-system
```

---

## See It in Action

### 1. Run a Unified Change-Control Check
```bash
./bin/garuda check --project go-validation-10 --target SQLModel --policy ./policies --save --fail-on-block
```

```text
╔═══════════════════════════════════════════════════════════════════╗
║               GARUDA CHANGE-CONTROL DECISION                      ║
╚═══════════════════════════════════════════════════════════════════╝
  Change:      SQLModel (workspace: go-validation-10)
  Affected:    171 entities (CRITICAL=12, HIGH=48)
  Evidence:    SUPPORTED=142  |  UNVERIFIED=5  |  CONTRADICTED=0
  Policies:    3 evaluated in 38ms
  Merkle Seal: RFC 6962 block #7
───
  GOVERNANCE OUTCOME: WARN
```

### 2. Declarative Architecture Policy (`policies/architecture-layer-guardrails.yaml`)
Garuda supports 9 deterministic SQL/CTE policy predicates (`layer_violation`, `transitive_claim_exists`, `cycle_exists`, `blast_radius_exceeds`, `claim_exists`, `entity_exists`, `contradiction_exists`, `verification_missing`, `language_matches`):

```yaml
id: "arch-no-handler-to-db"
version: "1.0"
title: "Prevent HTTP handlers from directly or transitively calling database drivers"
priority: 100
authority: "architecture-board"
language: "any"
scope:
  domain: "workspace"
  system: "core"
when:
  - type: "layer_violation"
    params:
      from_layer: "(handler|controller|api)"
      to_layer: "(database|sql|pgx|gorm)"
      limit: 10
then:
  decision: "BLOCK"
  reason: "Presentation/API handlers must route through domain/store interfaces rather than coupling to database packages."
```

---

## Language Support & Honest Boundaries

| Language | Analysis Engine | Extracted Semantics | Status |
| :--- | :--- | :--- | :--- |
| **Go** | Compiler-backed (`go/ast` + `go/types`) | Packages, structs, interfaces, functions, methods, intra/cross-package `CALLS`, `IMPORTS`, `IMPLEMENTS`, `EMBEDS`, `REFERENCES` | 🟢 **Production (Tier 5)** |
| **Python** | Structural AST | Modules, classes, functions, methods, `IMPORTS`, `INHERITS`, intra/cross-module `CALLS` (`self.method()` & functions) | 🟢 **Beta (Tier 2+)** |
| **TypeScript** | Tree-sitter AST | Modules, classes, interfaces, functions, methods, `IMPORTS`, `INHERITS`, `IMPLEMENTS`, `CALLS` | 🟢 **Beta (Tier 2+)** |

### What Garuda Is Not
- **Not an LLM code reviewer:** Garuda does not use probabilistic prompts to guess if code "looks good." Every verdict comes from deterministic graph queries and SQL predicates.
- **Not a replacement for unit tests or SAST:** Garuda enforces architectural boundaries, blast-radius thresholds, and runtime-to-static alignment. Pair it with your test suite and security scanners.
- **Honest epistemic states:** `UNVERIFIED` means no runtime trace or supporting evidence has been observed yet — it is never silently upgraded to `SUPPORTED`.

---

## Documentation

- 📘 **[Operational Playbook](PLAYBOOK.md)** — Complete setup, multi-repo workflows, policy authoring, runtime OTel correlation, and troubleshooting.
- 📗 **[Capabilities & Verification Matrix](docs/CAPABILITIES.md)** — Detailed feature matrix across all 5 architectural planes.
- 📙 **[CLI, Policy & MCP API Reference](docs/API.md)** — Every CLI command, flag, policy predicate, MCP tool, and REST endpoint.
- 📓 **[Changelog](CHANGELOG.md)** · **[Verification Evidence](EVIDENCE.md)** · **[System Specs](docs/SPECS.md)**
