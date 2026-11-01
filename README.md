# 🦅 Garuda — Architectural & Financial Governance for AI-Generated Software

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-emerald.svg)](LICENSE)
[![MCP Tools: 24 Deterministic Tools](https://img.shields.io/badge/MCP_Server-24_Tools_Verified-indigo.svg)](mcp-registry.json)
[![Codex Plugin Ready](https://img.shields.io/badge/OpenAI_Codex-Plugin_Ready-blue.svg)](.codex-plugin/plugin.json)

> **Run `garuda init` and get a draft architectural & financial rulebook reverse-engineered from your codebase in seconds.**

Garuda is the deterministic **Change-Control & Epistemic Governance Layer** for autonomous coding agents (**Claude Code, Cursor, OpenAI Codex, Windsurf**) and regulated engineering teams. Instead of relying on probabilistic LLM prompts, Garuda builds a **compiler-grade AST call graph**, synthesizes an editable **Base Rulebook**, resolves competing cross-team business formulas using **`0.0–1.0` team weights**, catches unit-scale data discrepancies (**Data & Logic Audit Ledger**), and seals every governance verdict in an **RFC 6962 SHA-256 Merkle tree**.

---

## 🚀 Start Here: Productive in 10 Minutes

- 📖 **[First 3 Rules in 10 Minutes (`QUICKSTART.md`)](QUICKSTART.md)** — The complete step-by-step happy path (`garuda init` → Auto-Rulebook Studio → `garuda rulebook approve` → `garuda check`).
- ⚡ **60-Second Live Demo (Zero Config)**:
  ```bash
  ./bin/garuda demo
  ```
  Spins up a 200-line sample fintech microservice (`acme-pay`) with **3 intentional violations** (direct `database/sql` in payment handler, `float64` currency math + 10x Crore/Lakh unit drift, and upward `store → payment` layer bypass), runs the full pipeline, and outputs the deterministic agent self-repair plan.

---

## ✨ Killer Workflow: Auto-Synthesized Base Rulebook Studio

Most codebases don't have up-to-date architecture rules written down. Garuda does the grunt work for you:

```bash
# 1. Interactive 5-question setup wizard (or pass --template financial / hexagonal / layered)
./bin/garuda init . --wizard

# 2. Reverse-engineer your repo's implicit layer boundaries, criticality tiers & formula weights
./bin/garuda rulebook generate

# 3. Review & edit in the Web Dashboard (Surface 04 — Base Rulebook Studio) or seal via CLI:
./bin/garuda rulebook generate --approve --approver "Staff Architect"

# 4. Enforce on every git diff, PR, and AI agent edit
./bin/garuda check . --diff HEAD --policy policies --fail-on-block
```

---

## 🛡️ 10 Pre-Built Starter Policy Templates (`policies/templates/`)

Run `./bin/garuda init --template <name>` or `./bin/garuda templates install <name>` to activate working guardrails in seconds:

| Template | Command | Enforced Invariant |
|---|---|---|
| **Hexagonal / Ports & Adapters** | `garuda init --template hexagonal` | Domain/core packages never import SQL, HTTP, Redis, or gRPC adapters. |
| **Layered Architecture** | `garuda init --template layered` | Persistence (`store`/`repo`) never imports upper service or API layers. |
| **Microservices DB Isolation** | `garuda init --template microservices` | Services never access another microservice's internal store package. |
| **Zero Circular Imports** | `garuda init --template no-circular-imports` | Blocks cyclic package dependencies across internal modules. |
| **Financial: No Direct SQL** | `garuda init --template financial-no-direct-sql` | Payment/refund handlers must route through `PaymentRepository`, never `database/sql`. |
| **Financial: Decimal Precision** | `garuda init --template financial-decimal-precision` | Monetary, fee, tax, and loss-ratio math must use Decimal / `int64` paise, never `float64`. |
| **Zero-Trust Tenant Scope** | `garuda init --template zero-trust-auth` | Exported API handlers must verify `tenant_id` & `workspace_id` scope. |
| **AI Agent Safety Gate** | `garuda init --template ai-agent-safety` | Freezes `SEVERE` criticality Merkle & auth signing code against unapproved agent edits. |
| **PII / GDPR / HIPAA Guard** | `garuda init --template pii-gdpr-hipaa` | Customer PII/PHI structs never flow into unredacted raw log sinks. |
| **Idempotent Mutations** | `garuda init --template idempotent-mutations` | External charge/payout mutations must verify an idempotency key. |

Already using **ArchUnit** (Java), **Dependency-Cruiser** (JS/TS), or **Import Linter** (Python)? Migrate in one command:
```bash
./bin/garuda import --from archunit --input src/test/java/ArchitectureTest.java
./bin/garuda import --from dependency-cruiser --input .dependency-cruiser.json
./bin/garuda import --from importlinter --input .importlinter
```

---

## ⚙️ Drop-In GitHub Action, SARIF & Slack/Teams Alerts

Copy [`.github/workflows/garuda-gate.yml`](.github/workflows/garuda-gate.yml) into any repository to:
1. Run `garuda check` against every PR diff (`--diff origin/$BASE_REF...HEAD`).
2. Post a rich Markdown governance & blast-radius table directly on the PR (`--pr-comment`).
3. Upload inline code annotations to the GitHub Security tab (`--sarif garuda.sarif`).
4. Send real-time violation alerts to Slack (`--notify-slack`) or Microsoft Teams (`--notify-teams`).
5. Fail the workflow (`exit 2`) if `CRITICAL` (`BLOCK`) violations are introduced.

---

## 🔌 MCP Server & AI Coding Agent Integrations

Garuda ships a hardened 24-tool MCP server (`garuda-mcp`) and native plugin manifests:
- **Claude Code & Claude Desktop**: Auto-configured by `garuda init` (`.claude/settings.json` + `CLAUDE.md`).
- **Cursor IDE**: Auto-configured by `garuda init` (`.cursor/rules/garuda-talon.mdc` + `.cursor/mcp.json`).
- **OpenAI Codex Plugin**: Native manifest at [`.codex-plugin/plugin.json`](.codex-plugin/plugin.json).
- **MCP Registries**: Ready-to-publish manifests for [Official MCP Registry (`mcp-registry.json`)](mcp-registry.json), [Smithery (`smithery.yaml`)](smithery.yaml), and [Glama (`glama.json`)](glama.json).

---

## 📚 Architecture Decision Records (ADRs), Playbooks & Deep Dives

- **Quickstart & Monorepo Guide**: [`QUICKSTART.md`](QUICKSTART.md)
- **Technical Deep-Dive Blog Post**: [`docs/BLOG_AST_MERKLE_AGENT_GOVERNANCE.md`](docs/BLOG_AST_MERKLE_AGENT_GOVERNANCE.md)
- **Architectural Decision Records (`docs/adr/`)**:
  - [`ADR-0001`: Payment Idempotency & Hexagonal Store Isolation](docs/adr/0001-payment-idempotency.md)
  - [`ADR-0002`: RFC 6962 Canonical SHA-256 Merkle Provenance Layer](docs/adr/0002-rfc6962-merkle-ledger.md)
  - [`ADR-0003`: Three-State Epistemic Honesty (`SUPPORTED`, `UNVERIFIED`, `CONTRADICTED`)](docs/adr/0003-epistemic-honesty.md)
- **Operational Playbooks (`docs/playbooks/`)**: [`docs/playbooks/`](docs/playbooks/)
- **Database Schema & 95 Migrations (`migrations/`)**: [`migrations/`](migrations/) (including `095_data_logic_audit_ledger_and_hygiene.sql`)
