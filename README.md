<p align="center">
  <img src="assets/garuda-logo.png" width="130" alt="Garuda Logo">
</p>

<h1 align="center">Garuda — Architectural &amp; Data-Logic Governance for AI-Generated Software</h1>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache_2.0-blue.svg" alt="License: Apache 2.0"></a>
  <a href="mcp-registry.json"><img src="https://img.shields.io/badge/MCP_Server-garuda--mcp_Ready-10b981.svg" alt="MCP Server Ready"></a>
  <a href=".codex-plugin/plugin.json"><img src="https://img.shields.io/badge/OpenAI_Codex-Plugin_Ready-4f46e5.svg" alt="Codex Plugin Ready"></a>
  <a href="action.yml"><img src="https://img.shields.io/badge/GitHub_Action-SARIF_%26_PR_Gate-0f172a.svg" alt="GitHub Action Ready"></a>
  <a href="EVIDENCE.md"><img src="https://img.shields.io/badge/Verified-63%2F63_MCP_%7C_10%2F10_Isolation-f59e0b.svg" alt="Verified"></a>
</p>

<p align="center">
  <strong>Run <code>garuda init</code> and get a draft architectural &amp; data-logic rulebook reverse-engineered from your codebase in seconds.</strong>
</p>

Garuda is the deterministic **Change-Control & Epistemic Governance Layer** for autonomous coding agents (**Claude Code**, **Cursor**, **OpenAI Codex**, **Windsurf**) and regulated engineering teams. Instead of relying on probabilistic LLM prompts, Garuda builds a compiler-grade AST call graph, synthesizes an editable **Base Rulebook**, resolves competing cross-team business formulas using **`0.0–1.0` team weights**, catches unit-scale data discrepancies (**Data & Logic Audit Ledger**, e.g. `48.00 Crore` vs `4.80 Crore`), and seals every governance verdict in an **RFC 6962 SHA-256 Merkle tree**.

> **AI can write the code. Who governs the change?**
> An AI agent modifies a payment service. Garuda discovers that the change bypasses the service layer, violates an architectural decision (`ADR-0001`), and introduces a 10x unit-scaling precision risk. **Garuda Talon** evaluates the diff, identifies the violations, outputs a deterministic self-repair prompt, and blocks the broken change from passing the gate—while recording a cryptographic RFC 6962 Merkle proof of the decision.
> **Understand the system. Govern the change. Preserve the proof.**

---

## 🚀 Start Here: Productive in 10 Minutes (`The Aha! Moment`)

- 📖 **[First 3 Rules in 10 Minutes (`QUICKSTART.md`)](QUICKSTART.md)** — The complete step-by-step happy path (`garuda init` → Auto-Rulebook Studio → `garuda rulebook approve` → `garuda check`).
- ⚡ **60-Second Live Demo (Zero Config)**:
  ```bash
  ./bin/garuda demo
  ```
  Spins up a 200-line sample fintech microservice (`acme-pay`) with **3 intentional violations** (*direct `database/sql` in payment handler*, *`float64` currency math + `10x` Crore/Lakh unit drift*, and *upward `store → payment` layer bypass*), runs the full pipeline, and outputs the deterministic agent self-repair plan.

---

## ✨ Killer Workflow: Auto-Synthesized Base Rulebook Studio

Most codebases don't have up-to-date architecture rules written down. Garuda does the grunt work for you:

```bash
# 1. Interactive 5-question setup wizard (or pass --template financial-no-direct-sql / hexagonal / layered)
./bin/garuda init . --wizard

# 2. Reverse-engineer your repo's implicit layer boundaries, criticality tiers & formula weights
./bin/garuda rulebook generate

# 3. Review & edit in the Web Dashboard (Surface 04 — Base Rulebook Studio) or seal via CLI:
./bin/garuda rulebook generate --approve --approver "Staff Architect"

# 4. Enforce on every git diff, PR, and AI agent edit
./bin/garuda check . --diff HEAD --policy policies --fail-on-block
```

---

## 🛡️ 10 Pre-Built Starter Policy Templates ([`policies/templates/`](policies/templates/))

Run `./bin/garuda init --template <name>` or `./bin/garuda templates install <name>` to activate working guardrails in seconds:

| Template | Command | Enforced Invariant |
| :--- | :--- | :--- |
| **Hexagonal / Ports & Adapters** | `garuda init --template hexagonal` | Domain/core packages never import SQL, HTTP, Redis, or gRPC adapters ([`hexagonal.yaml`](policies/templates/hexagonal.yaml)). |
| **Layered Architecture** | `garuda init --template layered` | Persistence (`store`/`repo`) never imports upper `service` or `api` layers ([`layered.yaml`](policies/templates/layered.yaml)). |
| **Microservices DB Isolation** | `garuda init --template microservices` | Services never access another microservice's internal store package ([`microservices.yaml`](policies/templates/microservices.yaml)). |
| **Zero Circular Imports** | `garuda init --template no-circular-imports` | Blocks cyclic package dependencies across internal modules ([`no-circular-imports.yaml`](policies/templates/no-circular-imports.yaml)). |
| **Data/Logic: No Direct SQL** | `garuda init --template financial-no-direct-sql` | Payment/refund handlers must route through `PaymentRepository`, never `database/sql` ([`financial-no-direct-sql.yaml`](policies/templates/financial-no-direct-sql.yaml)). |
| **Data/Logic: Decimal Precision** | `garuda init --template financial-decimal-precision` | Monetary, fee, tax, and loss-ratio math must use `Decimal` / `int64` paise, never `float64` ([`financial-decimal-precision.yaml`](policies/templates/financial-decimal-precision.yaml)). |
| **Zero-Trust Tenant Scope** | `garuda init --template zero-trust-auth` | Exported API handlers must verify `tenant_id` & `workspace_id` scope ([`zero-trust-auth.yaml`](policies/templates/zero-trust-auth.yaml)). |
| **AI Agent Safety Gate** | `garuda init --template ai-agent-safety` | Freezes `HIGH_SEVERE` criticality Merkle & auth signing code against unapproved agent edits ([`ai-agent-safety.yaml`](policies/templates/ai-agent-safety.yaml)). |
| **PII / GDPR / HIPAA Guard** | `garuda init --template pii-gdpr-hipaa` | Customer PII/PHI structs never flow into unredacted raw log sinks ([`pii-gdpr-hipaa.yaml`](policies/templates/pii-gdpr-hipaa.yaml)). |
| **Idempotent Mutations** | `garuda init --template idempotent-mutations` | External charge/payout mutations must verify an idempotency key ([`idempotent-mutations.yaml`](policies/templates/idempotent-mutations.yaml)). |

Already using **ArchUnit** (Java), **Dependency-Cruiser** (JS/TS), or **Import Linter** (Python)? Migrate in one command:

```bash
./bin/garuda import --from archunit --input src/test/java/ArchitectureTest.java
./bin/garuda import --from dependency-cruiser --input .dependency-cruiser.json
./bin/garuda import --from importlinter --input .importlinter
```

---

## ⚙️ Drop-In GitHub Action, SARIF & Slack/Teams Alerts

Copy **[`.github/workflows/garuda-gate.yml`](.github/workflows/garuda-gate.yml)** (or use [`action.yml`](action.yml)) into any repository to:

- Run `garuda check` against every PR diff (`--diff origin/$BASE_REF...HEAD`).
- Post a rich Markdown governance & blast-radius table directly on the PR (`--pr-comment`).
- Upload inline code annotations to the **GitHub Security** tab (`--sarif garuda.sarif`).
- Send real-time violation alerts to **Slack** (`--notify-slack`) or **Microsoft Teams** (`--notify-teams`).
- Fail the workflow (`exit 2`) if `CRITICAL` (`BLOCK`) violations are introduced.

---

## 🔌 MCP Server, Agent Skills & Ecosystem Plugins

Garuda ships a hardened MCP server (`garuda-mcp`), **6 specialist agent skills** ([`skills/`](skills/)), and native plugin manifests ready for local use and registry publication:

- **Claude Code & Claude Desktop**: Auto-configured by `garuda init` (`.claude/settings.json` + [`CLAUDE.md`](CLAUDE.md) + [`.claude-plugin/plugin.json`](.claude-plugin/plugin.json) + [`.claude-plugin/marketplace.json`](.claude-plugin/marketplace.json)).
- **Cursor IDE**: Auto-configured by `garuda init` (`.cursor/rules/garuda-talon.mdc` + `.cursor/mcp.json` + [`.cursor-plugin/plugin.json`](.cursor-plugin/plugin.json)).
- **OpenAI Codex Plugin**: Native manifest at **[`.codex-plugin/plugin.json`](.codex-plugin/plugin.json)** and [`AGENTS.md`](AGENTS.md).
- **6 Bundled Agent Skills ([`skills/`](skills/))**:
  - [`garuda-pre-edit-governance`](skills/pre-edit-governance/SKILL.md) · [`garuda-post-edit-verification`](skills/post-edit-verification/SKILL.md) · [`garuda-violation-repair`](skills/violation-repair/SKILL.md) · [`garuda-architecture-review`](skills/architecture-review/SKILL.md) · [`garuda-change-impact`](skills/change-impact/SKILL.md) · [`garuda-evidence-verification`](skills/evidence-verification/SKILL.md)
- **MCP Registries**: Ready-to-publish manifests for **Official MCP Registry** (**[`mcp-registry.json`](mcp-registry.json)**), **Smithery** (**[`smithery.yaml`](smithery.yaml)**), and **Glama** (**[`glama.json`](glama.json)**).

---

## 💎 Open-Source Core vs. Garuda Enterprise & Cloud

Garuda is built so individual developers and teams get an immediate, zero-friction **Aha! moment** locally for free—while scaling seamlessly into regulated multi-team organizations:

| Capability | Garuda Open-Source (Free Local & CI) | Garuda Team & Enterprise (Paid Tier) |
| :--- | :--- | :--- |
| **60-Second Demo, Auto-Rulebook Studio & 10 Policy Templates** | ✅ Included (`garuda demo`, `garuda rulebook generate`) | ✅ Included + Org-Wide Policy Inheritance |
| **Local AST Graph, Talon Pre-Commit Gate & MCP Server** | ✅ Included (Go, Python, TypeScript + `garuda-mcp`) | ✅ Included + Hosted Remote MCP Gateway |
| **10-Surface Progressive-Disclosure Local Dashboard (`/dashboard`)** | ✅ Included (Light & Graphite Dark Mode PWA) | ✅ Included + Multi-Tenant RBAC & Invite Codes (`GRD-XXXXXX`) |
| **Cross-Team Formula Disambiguation (`0.0–1.0` Weights) & Data-Logic Audit** | ✅ Local Repo & CLI (`garuda audit logic` / `data-logic`) | 🔒 **Cross-Org Pipeline Lineage, Live Warehouse Reconciliation & Slack/PagerDuty Escalation** |
| **Cryptographic RFC 6962 Merkle Ledger & Compliance Exports** | ✅ Local Workspace Merkle Proofs | 🔒 **Tamper-Evident Multi-Region Audit Vault, SOC2 / ISO-27001 / RBI Signed Bundles & SSO/SAML** |

---

## 📚 Architecture Decision Records (ADRs), Playbooks & Deep Dives

- **Quickstart & Monorepo Guide**: **[`QUICKSTART.md`](QUICKSTART.md)**
- **Persona Onboarding Guide**: **[`docs/PERSONA_ONBOARDING_GUIDE.md`](docs/PERSONA_ONBOARDING_GUIDE.md)**
- **Data & Logic Audit Guide (`0.0–1.0` Weights & Unit Drift)**: **[`docs/DATA_AND_LOGIC_AUDIT_GUIDE.md`](docs/DATA_AND_LOGIC_AUDIT_GUIDE.md)**
- **Technical Deep-Dive Blog Post**: **[`docs/BLOG_AST_MERKLE_AGENT_GOVERNANCE.md`](docs/BLOG_AST_MERKLE_AGENT_GOVERNANCE.md)**
- **Architectural Decision Records ([`docs/adr/`](docs/adr/))**:
  - **[`ADR-0001: Payment Idempotency & Hexagonal Store Isolation`](docs/adr/0001-payment-idempotency.md)**
  - **[`ADR-0002: RFC 6962 Canonical SHA-256 Merkle Provenance Layer`](docs/adr/0002-rfc6962-merkle-ledger.md)**
  - **[`ADR-0003: Three-State Epistemic Honesty (SUPPORTED, UNVERIFIED, CONTRADICTED)`](docs/adr/0003-epistemic-honesty.md)**
- **Operational Playbooks ([`docs/playbooks/`](docs/playbooks/README.md))**: **[`docs/playbooks/README.md`](docs/playbooks/README.md)**
- **Database Schema & 98 Migrations ([`migrations/`](migrations/))**: **[`migrations/`](migrations/)** (including [`095_data_logic_audit_ledger_and_hygiene.sql`](migrations/095_data_logic_audit_ledger_and_hygiene.sql), [`096_policy_evaluations_decision_check.sql`](migrations/096_policy_evaluations_decision_check.sql), [`097_multi_tenant_rbac_isolation.sql`](migrations/097_multi_tenant_rbac_isolation.sql), and [`098_strict_workspace_and_tenant_isolation.sql`](migrations/098_strict_workspace_and_tenant_isolation.sql))
- **10/10 Multi-Tenant Isolation & Evidence Regression Suite**: **[`internal/api/tenant_isolation_test.go`](internal/api/tenant_isolation_test.go)**
- **API & MCP Reference**: **[`docs/API.md`](docs/API.md)** · **[`EVIDENCE.md`](EVIDENCE.md)** · **[`INSTALL.md`](INSTALL.md)**

---

<p align="center">
  <strong>Understand the system. Govern the change. Preserve the proof.</strong><br>
  Released under the <a href="LICENSE">Apache 2.0 License</a>.
</p>
