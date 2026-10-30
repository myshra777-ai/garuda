<p align="center">
  <img src="assets/garuda_logo.png" width="140" alt="Garuda Logo">
</p>

<h1 align="center">Garuda</h1>

<p align="center">
  <strong>AI writes code at 100 mph. Garuda makes sure your architecture, business formulas, and financial numbers survive the ride.</strong><br>
  <em>The local-first change-control gate, active AI agent harness (<code>Garuda Talon</code>), human-in-the-loop formula &amp; financial audit engine, and 10-surface software intelligence instrument.</em>
</p>

<p align="center">
  <a href="https://github.com/myshra777-ai/garuda/actions/workflows/ci.yml"><img src="https://github.com/myshra777-ai/garuda/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/myshra777-ai/garuda/actions/workflows/garuda-ci.yml"><img src="https://github.com/myshra777-ai/garuda/actions/workflows/garuda-ci.yml/badge.svg" alt="Garuda CI & Governance"></a>
  <a href="https://github.com/myshra777-ai/garuda/actions/workflows/garuda-gate.yml"><img src="https://github.com/myshra777-ai/garuda/actions/workflows/garuda-gate.yml/badge.svg" alt="Garuda Integrity Gate"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/version-v1.9.6-2563eb.svg" alt="Version v1.9.6"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache--2.0-blue.svg" alt="Apache 2.0 License"></a>
  <a href="INSTALL.md"><img src="https://img.shields.io/badge/install-1_Command_%7C_Zero_Dependencies-10b981.svg" alt="1-Command Universal Install"></a>
  <a href="EVIDENCE.md"><img src="https://img.shields.io/badge/verified-24%2F24_Benchmarks_%7C_35_MCP_Tools_%7C_10%2F10_Runtime-f59e0b.svg" alt="Verification Evidence"></a>
</p>

<p align="center">
  <a href="#watch-the-20-second-cinematic-intro">20s Cinematic Demo</a> ·
  <a href="#the-story-why-we-built-garuda">Why Garuda Exists</a> ·
  <a href="#new-in-v191-human-in-the-loop-weights-criticality-tiers--financial-audits">Human-in-the-loop audits</a> ·
  <a href="#a-visual-tour-of-garuda">Visual Product Tour</a> ·
  <a href="#install-in-one-command-zero-dependencies--foolproof">1-Command Install</a> ·
  <a href="#when-to-use-what-cli-command--mcp-tool">CLI &amp; MCP Quick Guide</a> ·
  <a href="#documentation-suite-by-role">Documentation by Role</a>
</p>

---

## Dashboard: a calmer command surface

The current dashboard groups its ten working surfaces into **Explore**, **Enforce**, and **Verify**. It uses a graphite-first theme, restrained mint signal color, keyboard navigation, progressive disclosure, responsive navigation, and reduced-motion support. The main overview, search, docs, Talon, session, decision, and review flows use live workspace results by default; fallback sample data in those flows is opt-in via `?demo=1`. The Data & Logic Audit API also includes built-in illustrative formula/financial scenarios alongside live workspace-enriched criticality data. Treat those baseline scenarios as examples, not automatically detected findings against the selected repository.

See the [Dashboard Walkthrough](docs/DASHBOARD_WALKTHROUGH.md) for every surface, panel, control, and recommended workflow. The dashboard currently ingests ADRs, specifications, invariants, and policy documents. Repository registration and synchronization are available through the CLI (`garuda repo add` and `garuda workspace sync`); direct repository upload/import is not yet a dashboard capability.

## Watch the 20-Second Cinematic Intro

Experience what Garuda is and how its **Software Intelligence Instrument** turns invisible code structure, architecture rules, business formulas, and AI agent edits into a living control room:

<div align="center">
  <video src="assets/garuda_dashboard_demo_video.mp4" poster="assets/Screenshot%20From%202026-10-25%2002-30-49.png" controls autoplay loop muted playsinline width="100%">
    <source src="assets/garuda_dashboard_demo_video.mp4" type="video/mp4">
    <source src="https://raw.githubusercontent.com/myshra777-ai/garuda/main/assets/garuda_dashboard_demo_video.mp4" type="video/mp4">
    <a href="assets/garuda_dashboard_demo_video.mp4">
      <img src="assets/Screenshot%20From%202026-10-25%2002-30-49.png" width="100%" alt="Watch the 20-Second Garuda Introduction Video">
    </a>
  </video>
  <p><em>🎬 <strong>20-Second Product Introduction:</strong> If your browser blocks inline video playback on GitHub, <a href="assets/garuda_dashboard_demo_video.mp4"><strong>click here to play <code>assets/garuda_dashboard_demo_video.mp4</code></strong></a>.</em></p>
</div>

---

## The Story: Why We Built Garuda

Today, coding agents like **Cursor**, **Claude Code**, **Codex**, and **Windsurf** can generate a thousand lines of working code in thirty seconds. Unit tests pass. Linter checks turn green. The pull request looks clean.

**Then, six months later—or during a statutory financial audit—your team hits a wall.**

Why? Because AI agents optimize for *making the local prompt work right now*, not for protecting your architecture or business logic:

- **Silent Layer Bypasses:** An agent makes an HTTP handler talk directly to `database/sql` or `DirectBankAPI`, bypassing your idempotency store (`ADR-0001`).
- **Undocumented Competing Business Formulas:** Ops, Marketing, and Management each code their own undocumented formula for **Loss Ratio**. An AI agent overwrites the core pricing gate with a marketing campaign formula, taking down search redirection across 142 callers.
- **Hidden 10x Financial Discrepancies (48 Crore vs. 4.8 Crore):** Claims Ops reports **₹48.00 Crore** in settled claims (`SUM(settled_claims_inr) / 1e7`), while Finance's executive export reports **₹4.80 Crore** because a rollup query divides Lakhs by `1000.0` instead of `100.0`. Nobody notices until an external financial audit.
- **No Written Rules in Unorganized Repos:** Most growing companies don't have time to write formal ADRs from scratch, so their architecture rules live only in senior engineers' heads.

**Linters check syntax. Unit tests check small functions. Neither guards your architecture or your financial lineage.**

We built **Garuda** to be the deterministic referee for the AI coding era—combining compiler-grade AST analysis, active AI agent guardrails (`Garuda Talon`), human-in-the-loop formula weighting (`0.0–1.0`), continuous financial data-lineage audits, and RFC 6962 cryptographic Merkle proofs.

---

## New in `v1.9.1`: Human-in-the-Loop Weights, Criticality Tiers & Financial Audits

1. **Human-in-the-Loop Multi-Team Formula Disambiguation (`0.0–1.0` Weights)**:
   - When multiple teams implement different formulas for the same concept (e.g., `Loss Ratio` across **Ops**, **Marketing**, and **Management**), **Garuda never lets AI guess**.
   - Calculates dynamic **`0.00–1.00` team weights** (`0.70` active team vs `0.10`/`0.10`), ranks variants by **Least Architectural Harm**, and presents plain-English blast radius impacts so the human developer makes the final call (`garuda audit logic` / `garuda.disambiguate_logic`).
2. **3-Tier Surface Criticality Classification (Amazon Search vs. Wishlist Model)**:
   - **`HIGH_SEVERE` (Core Surfaces — Freeze / Do Not Touch)**: Homepage, Search Bar, Search Results Redirect (142 callers), and Claims Settlement Ledger. Frozen against automated AI edits.
   - **`MEDIUM` (Non-Blocking Features — Human Trade-Off Review)**: Wishlist Loader and Promo Badges (19 callers). Secondary features where trade-offs can be weighed.
   - **`LOW` (Minor UI & Formatting — Allow with Info)**: Search Box visual alignment and badge helpers (4 callers).
3. **Continuous Financial & Business Data Discrepancy Audit (48 Crore vs. 4.8 Crore)**:
   - Traces AST and SQL unit-scaling lineage across team pipelines to quarantine **10x Lakhs-to-Crore divisor bugs** (`48.00 Cr` vs `4.80 Cr`) and omitted reinsurance deductions *before* external audits occur (`garuda audit financial` / `garuda.audit_financial_discrepancy`). See **[`docs/FINANCE_AND_AUDIT_GUIDE.md`](docs/FINANCE_AND_AUDIT_GUIDE.md)**.
4. **Auto-Synthesized Base Rulebook Studio (Grunt Work Done by Garuda, Approved by You)**:
   - Automatically reverse-engineers unorganized repositories into `.garuda/DRAFT_BASE_RULEBOOK.md`. Stakeholders review, edit inline in `/dashboard`, and click **Approve & Seal Base Rulebook** (`garuda rulebook generate` / `garuda rulebook approve`).
5. **Extended 35-Tool Production MCP Server & Bulletproof All-OS Installer**:
   - Expanded `bin/garuda-mcp` to **35 deterministic tools** (`43/43` verification checks passing in `scripts/mcp_verify.py`) and hardened `setup.sh` / `setup.ps1` with `pg_isready` WAL polling and automatic role/database self-healing.

---

## A Visual Tour of Garuda (`9-Surface Dashboard PWA`)

Garuda ships with a local, zero-cloud **9-Surface Software Intelligence Instrument** (`/dashboard`) that runs 100% offline in your browser or as an installable desktop/mobile PWA:

### 1. Executive Workspace Briefing (`01 — Overview`)
Immediate system command center: indexed topology, 3-state epistemic claims (`SUPPORTED` / `UNVERIFIED` / `CONTRADICTED`), active Talon policy gate, RFC 6962 Merkle block height, and direct jump buttons to **Docs & Base Rulebook**, **Criticality & Risk Hub**, and **Financial & Formula Audit (48Cr vs 4.8Cr)**.

<p align="center">
  <img src="assets/Screenshot%20From%202026-10-25%2002-30-49.png" width="100%" alt="Garuda Workspace Briefing & System Health">
</p>

### 2. Explore Index & Instant Symbol Intelligence (`02 — Explore Index`)
Surfaces your most critical exported structs, interfaces, functions, and packages ranked by caller fan-in, with instant `⌘K` / `Ctrl+K` search and 1-click neighborhood preview.

<p align="center">
  <img src="assets/Screenshot%20From%202026-10-25%2002-31-11.png" width="100%" alt="Garuda Explore Index & High-Impact Symbol Directory">
</p>

### 3. Neighborhood-First Architecture Map (`03 — Architecture Map`)
Renders a clean **1-hop or 2-hop neighborhood map** around any symbol or package (`[Callers] -> [Selected Symbol] -> [Dependencies & Policy Gate]`) with a synchronized side Inspector.

<p align="center">
  <img src="assets/Screenshot%20From%202026-10-25%2002-31-24.png" width="100%" alt="Garuda Neighborhood Architecture Map & Inspector Drawer">
</p>

### 4. Docs, Specs & Auto-Synthesized Base Rulebook Studio (`04 — Docs & Base Rulebook`)
Connects **Tier 1 Documents** → **Tier 2 Extracted Rules (`MUST` / `MUST_NOT`)** → **Tier 3 Active Guardrails** → **Tier 4 Code Symbols**, and includes the **Auto-Synthesized Base Rulebook Studio** for 1-click human stakeholder approval.

<p align="center">
  <img src="assets/Screenshot%20From%202026-10-25%2002-31-40.png" width="100%" alt="Garuda Docs, Specs & Guardrails 4-Tier Lineage Graph">
</p>

### 5. Garuda Talon, Surface Criticality Tiers & Risk Hub (`05 — Talon & Risk Hub`)
Displays all 6 active IDE/Git hooks, the **Deterministic `BuildContext` Provenance Receipt**, the **4-Target Cross-Platform Matrix (`linux`, `darwin`, `windows`, `cgo`)**, **Surface Criticality Tiers (`HIGH_SEVERE` / `MEDIUM` / `LOW`)**, and the **Risk Hub Leaderboard**.

<p align="center">
  <img src="assets/Screenshot%20From%202026-10-25%2002-31-56.png" width="100%" alt="Garuda Talon Active Agent Harness & Architectural Risk Hub">
</p>

### 6. Human-in-the-Loop Change Review Workbench (`06 — Change Review`)
A 3-pane pre-merge workbench inspecting **Change Context**, **Blast-Radius Impact Map**, and **3-Plane Epistemic, Policy & Merkle Evidence**.

<p align="center">
  <img src="assets/Screenshot%20From%202026-10-25%2002-32-11.png" width="100%" alt="Garuda Change Review Workbench & Policy Sandbox">
</p>

### 7. Evidence Ledger, Agents & Financial/Logic Audit (`07 — Evidence`, `08 — Agents`, `09 — Financial & Logic Audit`)
- **07 — Evidence & Ledger**: Append-only RFC 6962 SHA-256 Merkle audit log with 1-click cryptographic inclusion proof verification.
- **08 — Agents & Sessions**: Live MCP client session timeline and multi-agent handoff tracking.
- **09 — Financial & Logic Audit**: Side-by-side `0.0–1.0` multi-team formula disambiguation (`Ops` vs `Marketing` vs `Management`) and continuous financial discrepancy detector (`48.00 Cr` vs `4.80 Cr`).

<p align="center">
  <img src="assets/Screenshot%20From%202026-10-25%2002-32-25.png" width="100%" alt="Garuda Contradictions Forensic Queue & Cryptographic Audit Ledger">
</p>

---

## Active IDE Integration (Cursor, VS Code, Claude Code, Windsurf)

| Real-Time Knowledge Graph in IDE | Live Contradiction & Rule Violation Alerts |
| :---: | :---: |
| <img src="assets/screenshots/ide-graph-view.png" width="100%" alt="Garuda IDE Graph View"> | <img src="assets/screenshots/ide-contradictions.png" width="100%" alt="Garuda IDE Contradictions Alert"> |

---

## Install in One Command (Zero Dependencies & Foolproof)

You do **not** need Go, PostgreSQL, Docker, GCC, Homebrew, or Xcode pre-installed, and you do **not** need Administrator or `sudo` privileges. Garuda `v1.9.1` provisions pre-built Portable PostgreSQL `16.4` (`127.0.0.1:5433` with `pg_isready` WAL readiness polling and automatic `pg_hba.conf` trust healing) and Portable Go `1.26.0` in `~/.garuda`:

### macOS (Apple Silicon & Intel), Linux, Chromebook & Android (Termux)

```bash
# Option A: 1-Command Remote Install & Launch
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash

# Option B: Inside an existing cloned repository
git clone --depth 1 https://github.com/myshra777-ai/garuda.git && cd garuda && ./setup.sh
```

### Windows 10 / 11 (PowerShell 5.1 & PowerShell 7 — Pure ASCII, No Admin Required)

```powershell
# Option A: 1-Command Remote Install & Launch
Set-ExecutionPolicy Bypass -Scope Process -Force; irm https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.ps1 | iex

# Option B: Inside an existing cloned repository
cd $HOME\garuda
powershell -NoProfile -ExecutionPolicy Bypass -File .\setup.ps1
```

### Instant Self-Diagnosis & Auto-Repair (`garuda doctor`)

```bash
./bin/garuda doctor --fix
```

---

## When to Use What CLI Command & MCP Tool

For the complete decision matrix, see **[`docs/CLI_DECISION_GUIDE.md`](docs/CLI_DECISION_GUIDE.md)**. Here are the most common workflows at a glance:

| Workflow | CLI Command (`./bin/garuda ...`) | MCP Tool (`bin/garuda-mcp`) |
| :--- | :--- | :--- |
| **Pre-Edit Blast Radius & ADR Briefing** | `./bin/garuda talon brief <Symbol>` | `garuda.talon_pre_edit` |
| **Post-Edit Diff Gate & Self-Repair Prompt** | `./bin/garuda talon guard --diff HEAD --fail-on-block` | `garuda.talon_verify_edit` |
| **Cross-OS Build Matrix (`linux`/`darwin`/`windows`/`cgo`)** | `./bin/garuda talon guard --matrix` | `garuda.build_matrix_check` |
| **Compare Competing Team Formulas (`0.0–1.0` Weights)** | `./bin/garuda audit logic "Loss Ratio" --team ops` | `garuda.disambiguate_logic` / `garuda.decide_logic_variant` |
| **Audit Cross-Team Financial Numbers (`48 Cr` vs `4.8 Cr`)** | `./bin/garuda audit financial` | `garuda.audit_financial_discrepancy` / `garuda.reconcile_financial_audit` |
| **Auto-Generate & Approve Base Rulebook** | `./bin/garuda rulebook generate && ./bin/garuda rulebook approve` | `garuda.synthesize_rulebook` / `garuda.approve_base_rulebook` |
| **Inspect 3-Tier Surface Criticality** | `./bin/garuda audit criticality` | `garuda.criticality_tiers` |
| **Ingest ADRs & View 4-Tier Docs-to-Code Graph** | `./bin/garuda docs ingest ./docs/adr && ./bin/garuda docs graph` | `garuda.docs_ingest` / `garuda.docs_graph` |

---

## Documentation Suite by Role

| Document | Target Audience | What's Covered |
| :--- | :--- | :--- |
| **[`docs/PERSONA_ONBOARDING_GUIDE.md`](docs/PERSONA_ONBOARDING_GUIDE.md)** | **All Users (Start Here!)** | Tailored 5-minute onboarding paths for Engineers, Architects, Finance/Audit Leads, Founders/PMs, DevOps, AI Agents, and Investors. |
| **[`docs/FINANCE_AND_AUDIT_GUIDE.md`](docs/FINANCE_AND_AUDIT_GUIDE.md)** | **Finance, CFOs, Actuaries & Auditors** | Catching 48 Cr vs 4.8 Cr unit-scaling discrepancies pre-audit, `0.0–1.0` formula weighting, and RFC 6962 statutory audit proofs. |
| **[`docs/CLI_DECISION_GUIDE.md`](docs/CLI_DECISION_GUIDE.md)** | **Developers, Architects & AI Agents** | Complete "When-to-Use-What" decision matrix across all 22 CLI commands and all **35 MCP tools**. |
| **[`INSTALL.md`](INSTALL.md)** | **Everyone** | Foolproof 1-command zero-dependency setup across Windows PS 5.1/7, macOS (Intel & Apple Silicon), Linux, Chromebook, and Android + 20-scenario troubleshooting table. |
| **[`PLAYBOOK.md`](PLAYBOOK.md)** | **Engineering & DevOps Teams** | End-to-end operational runbook covering all 9 dashboard surfaces, Talon harness, YAML policies, OTel runtime verification, and CI gates. |
| **[`docs/WALKTHROUGH.md`](docs/WALKTHROUGH.md)** | **Evaluators & Product Teams** | Visual screen-by-screen tour of the 9-surface `/dashboard` PWA, Base Rulebook Studio, Financial Audit Workbench, and IDE extension. |
| **[`docs/CAPABILITIES.md`](docs/CAPABILITIES.md)** | **Staff Architects & Technical DD** | 8-plane production capability matrix with verification tiers and enforced invariants. |
| **[`docs/API.md`](docs/API.md)** | **Platform & Tooling Engineers** | Complete CLI flag reference, 35 MCP tool schemas, 9 YAML policy predicates, and HTTP REST endpoints. |
| **[`docs/SPECS.md`](docs/SPECS.md)** | **Systems & Security Engineers** | Formal v1.9.1 system specification, PostgreSQL schemas (`001`–`093`), `BuildContext` hashing, and RFC 6962 Merkle internals. |
| **[`EVIDENCE.md`](EVIDENCE.md)** | **Investors & Technical Reviewers** | Reproducible verification results (`24/24` compiler benchmarks, `43/43` MCP checks across 35 tools, `10/10` runtime contradiction quarantine). |
| **[`CHANGELOG.md`](CHANGELOG.md)** | **Everyone** | Complete release history from `v1.1.0` through **`v1.9.1`**. |

---

<p align="center">
  <strong>Let agents write fast. Let Garuda keep your architecture, formulas, and financial numbers whole.</strong><br>
  Released under the <a href="LICENSE">Apache 2.0 License</a>.
</p>
