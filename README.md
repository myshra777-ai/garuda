<p align="center">
  <img src="assets/garuda_logo.png" width="140" alt="Garuda Logo">
</p>

<h1 align="center">Garuda</h1>

<p align="center">
  <strong>AI writes code at 100 mph. Garuda makes sure your architecture, business formulas, and cross-team numbers survive the ride.</strong><br>
  <em>The local-first change-control gate, active AI agent harness (<code>Garuda Talon</code>), human-in-the-loop Data &amp; Logic Audit engine, and 10-surface progressive-disclosure Software Intelligence Dashboard.</em>
</p>

<p align="center">
  <a href="https://github.com/myshra777-ai/garuda/actions/workflows/ci.yml"><img src="https://github.com/myshra777-ai/garuda/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/myshra777-ai/garuda/actions/workflows/garuda-ci.yml"><img src="https://github.com/myshra777-ai/garuda/actions/workflows/garuda-ci.yml/badge.svg" alt="Garuda CI & Governance"></a>
  <a href="https://github.com/myshra777-ai/garuda/actions/workflows/garuda-gate.yml"><img src="https://github.com/myshra777-ai/garuda/actions/workflows/garuda-gate.yml/badge.svg" alt="Garuda Integrity Gate"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/version-v1.9.9-10b981.svg" alt="Version v1.9.9"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache--2.0-blue.svg" alt="Apache 2.0 License"></a>
  <a href="INSTALL.md"><img src="https://img.shields.io/badge/install-1_Command_%7C_Zero_Dependencies-10b981.svg" alt="1-Command Universal Install"></a>
  <a href="EVIDENCE.md"><img src="https://img.shields.io/badge/verified-24%2F24_Benchmarks_%7C_63%2F63_MCP_Checks_(54_Tools)-f59e0b.svg" alt="Verification Evidence"></a>
</p>

<p align="center">
  <a href="#what-is-garuda-in-plain-english">What Is Garuda?</a> ·
  <a href="#watch-the-20-second-cinematic-intro">20s Video Demo</a> ·
  <a href="#install-in-one-command-zero-dependencies">1-Command Install</a> ·
  <a href="#whats-new-in-v199">What's New in v1.9.9</a> ·
  <a href="#tour-the-futuristic-10-surface-dashboard">Visual Dashboard Tour</a> ·
  <a href="#everyday-cli--mcp-workflows">Everyday CLI &amp; MCP Guide</a> ·
  <a href="#documentation-suite-by-role">Full Documentation Suite</a>
</p>

---

## What Is Garuda? (In Plain English)

The current dashboard groups its ten working surfaces into **Explore**, **Enforce**, and **Verify**. It uses a graphite-first theme, restrained mint signal color, keyboard navigation, progressive disclosure, responsive navigation, and reduced-motion support. The Talon hub, pre-edit drawer, change-review checks and build matrix use returned workspace/API data; missing or failed checks are shown as unavailable rather than turned into an ALLOW verdict. The Data & Logic Audit API still includes illustrative formula/financial scenarios alongside workspace-enriched criticality context; treat those rows as examples until linked to imported source evidence. Risk-hub rankings are degree-based heuristics, not policy decisions.
Today, AI coding tools like **Cursor**, **Claude Code**, **Codex**, and **Windsurf** can write a thousand lines of code in thirty seconds. The code compiles, unit tests pass, and the pull request looks green.

**But how do you know that fast AI-generated code didn't quietly break something important?**

In real teams, four silent failures happen every week when humans and AI agents ship fast:

1. **Silent Architecture Shortcuts**: An AI agent makes an API handler talk straight to the database or an external payment API to "make the prompt work," bypassing your security, retry, or store layers.
2. **Competing Business Formulas Across Teams**: Your **Ops**, **Marketing**, and **Management** teams each write their own formula for the same metric (like *Loss Ratio*, *Retention*, or *Net Margin*). Without a referee, an AI agent copies a marketing campaign formula into your core pricing router—breaking 142 downstream callers.
3. **10x Cross-Team Number Mismatches (48 Crore vs. 4.8 Crore)**: One team's code divides raw Rupees by `10,000,000` and reports **48.00 Crore**, while another team's rollup query divides Lakhs by `1,000` instead of `100` and reports **4.80 Crore**. Nobody catches the 10x error until an executive review or audit.
4. **Undocumented "Tribal Knowledge"**: Most growing codebases don't have up-to-date Architecture Decision Records (ADRs). Rules live only in senior engineers' heads, so new developers and AI agents don't know which files are safe to edit and which ones are mission-critical.

**Garuda is your local-first architectural safety net and truth engine.**

It sits quietly on your machine (and in CI), reads your real code (Go, Python, and TypeScript), connects it to your architecture rules and business formulas, guides your AI coding agents before and after every edit, and gives you a calm, futuristic **10-Surface Web Dashboard** where humans make the final call.

---

## Watch the 20-Second Cinematic Intro

See how Garuda turns invisible code dependencies, architecture rules, business formulas, and AI agent edits into a calm, interactive control room:

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

## Install in One Command (Zero Dependencies)

You do **not** need Go, PostgreSQL, Docker, GCC, Homebrew, or Xcode pre-installed, and you do **not** need `sudo` or Administrator rights. Garuda automatically sets up a portable local database (`127.0.0.1:5433`), compiles all 4 binaries (`garuda`, `garuda-api`, `garuda-mcp`, `garuda-talon`), wires your AI coding tools, and opens the dashboard:

### Linux, macOS (Apple Silicon & Intel), Chromebook & Android (Termux)

```bash
# Option A: 1-Command Remote Install & Launch
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash

# Option B: Inside an existing cloned repository
git clone --depth 1 https://github.com/myshra777-ai/garuda.git && cd garuda && ./setup.sh
```

### Windows 10 / 11 (PowerShell 5.1 & PowerShell 7 — No Admin Required)

```powershell
# Option A: 1-Command Remote Install & Launch
Set-ExecutionPolicy Bypass -Scope Process -Force; irm https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.ps1 | iex

# Option B: Inside an existing cloned repository
cd $HOME\garuda
powershell -NoProfile -ExecutionPolicy Bypass -File .\setup.ps1
```

Once setup finishes, open **`http://localhost:8080/login`** in your browser. Any time your computer restarts or you want to check system health, run:

```bash
./bin/garuda doctor --fix
```

---

## What's New in `v1.9.9`

### 1. Futuristic Progressive-Disclosure Dashboard (`10 Surfaces`, Light & Graphite Dark Mode)
- **Easy on the Eyes, Zero Clutter**: Inspired by Linear, Vercel, Stripe, and Apple Human Interface Guidelines, every dashboard surface shows **at most 4 primary cards on first view**. Secondary tables, drift breakdowns, and raw telemetry sit one click away inside smooth expandable `<details>` drawers and sub-tabs.
- **Smooth `cubic-bezier(0.16, 1, 0.3, 1)` Motion**: GPU-accelerated staggered entrance transitions (`surfaceIn`), subtle card elevation on hover, a 3-step tabular-numeral type hierarchy, and a 1-click toggle between **Warm Alabaster Light Mode** and **Obsidian Graphite Dark Mode** (`☀ / ☾`).

### 2. Human-in-the-Loop Data & Logic Audit Ledger (`09 — Data & Logic Audit`)
- **0.0–1.0 Multi-Team Formula Weighting**: When **Ops**, **Marketing**, and **Management** code different formulas for the same concept (such as *Loss Ratio* or *Retention*), Garuda never lets AI silently pick or overwrite one. It computes **`0.00–1.00` team weights** (`0.70` active team vs `0.10`/`0.10`), ranks each option by **Least Architectural Harm**, and requires human co-signing (`garuda audit logic`).
- **Cross-Team Data & Unit-Scale Discrepancy Detector**: Traces AST and SQL data lineage across teams to catch **10x unit-scaling bugs** (such as **48.00 Crore vs. 4.80 Crore** from dividing Lakhs by `1000` instead of `100`) before numbers ever reach an external review (`garuda audit data-logic`, backed by `data_logic_audit_ledger` and `data_audit_reconciliations`).

### 3. 3-Tier Surface Criticality Classification (`HIGH_SEVERE` · `MEDIUM` · `LOW`)
- **`HIGH_SEVERE` (Core Surfaces — Frozen)**: Homepage, Search Redirect Router (142 callers), and Core Settlement Ledgers are locked against automated AI edits (`HUMAN_FREEZE_DO_NOT_TOUCH`).
- **`MEDIUM` (Non-Blocking Features — Human Trade-Off)**: Secondary features like Wishlist Loaders (19 callers) allow informed human trade-offs.
- **`LOW` (UI & Formatting — Safe)**: Isolated UI box alignment and formatting helpers (4–6 callers) proceed cleanly.

### 4. Auto-Synthesized Base Rulebook Studio (`04 — Docs & Rules Graph`)
- Don't have time to write ADRs from scratch? Garuda scans your repository, discovers implicit layer boundaries, criticality tiers, and formula weights, and drafts `.garuda/DRAFT_BASE_RULEBOOK.md`. Review and edit it right in the browser, then click **Approve & Seal Rulebook** (`garuda rulebook generate` / `garuda rulebook approve`).

### 5. Hardened 54-Tool MCP Server & Deterministic Dirty-Worktree Receipts (`63/63` Checks)
- **54 Deterministic MCP Tools**: Includes `garuda.preflight_bundle` (1-call pre-edit context), `garuda.verify_receipt` (RFC 6962 cryptographic verification), `garuda.diff_snapshots` (revision-to-revision graph diff), `garuda.audit_export` (CSV/JSON/Markdown compliance export), and `garuda.create_workspace` (isolated sandbox workspaces).
- **Exact Dirty-Worktree & Cross-OS Reproducibility (`--worktree-dirty`, `--dirty-tree-hash`, `--matrix`)**: Every `garuda check` and `garuda talon guard` verdict binds `GOOS/GOARCH`, `CGO_ENABLED`, build tags, git commit SHA, and dirty working-tree hash into a single copy-pasteable reproduce command.

---

## Tour the Futuristic 10-Surface Dashboard (`/dashboard` PWA)

Garuda's web interface runs 100% locally at `http://localhost:8080/dashboard` (and installs as a standalone desktop or mobile app):

### `01` — Overview: Calm Command Center
Shows **1 Hero System Integrity Card** (`ALLOW` / `BLOCK`) + **3 clickable core metric cards** (*Indexed Topology*, *Epistemic & Policy Gate*, *RFC 6962 Provenance Ledger*), with progressive-disclosure drawers for *Needs Attention & Documentation Drift* and *Architectural Hubs & Indexed Repositories*.

<p align="center">
  <img src="assets/Screenshot%20From%202026-10-25%2002-30-49.png" width="100%" alt="Garuda Workspace Briefing & System Health">
</p>

### `02` — Explore Index: Instant Symbol Search (`⌘K`)
Search packages, functions, methods, structs, and interfaces with instant signature preview and 1-click neighborhood actions.

<p align="center">
  <img src="assets/Screenshot%20From%202026-10-25%2002-31-11.png" width="100%" alt="Garuda Explore Index & High-Impact Symbol Directory">
</p>

### `03` — Architecture Map: Full-Viewport Neighborhood Graph
Explores 1-hop and 2-hop symbol neighborhoods (`Callers → Symbol → Dependencies`) on an interactive D3 canvas with collapsible depth/edge filters and a synchronized entity inspector.

<p align="center">
  <img src="assets/Screenshot%20From%202026-10-25%2002-31-24.png" width="100%" alt="Garuda Neighborhood Architecture Map & Inspector Drawer">
</p>

### `04` — Docs & Rules Graph: 4-Tab Progressive Studio
Switches cleanly between **01 · Interactive 4-Tier Graph** (`Document → Rule → Policy → Code`), **02 · Base Rulebook Studio**, **03 · Ingested Documents**, and **04 · Extracted Rules Matrix**.

<p align="center">
  <img src="assets/Screenshot%20From%202026-10-25%2002-31-40.png" width="100%" alt="Garuda Docs, Specs & Guardrails 4-Tier Lineage Graph">
</p>

### `05` — Talon & Risk Hub: Active AI Agent Guardrails
Organizes AI agent governance into 3 focused tabs: **01 · Live Enforcement Gate & Hooks** (6 IDE/Git hooks + self-repair simulator), **02 · Risk Hub Leaderboard & Surface Criticality**, and **03 · Deterministic Build Matrix & Receipt**.

<p align="center">
  <img src="assets/Screenshot%20From%202026-10-25%2002-31-56.png" width="100%" alt="Garuda Talon Active Agent Harness & Architectural Risk Hub">
</p>

### `06` — Change Review: 3-Pane Pre-Merge Workbench
Inspect any symbol or git diff across three balanced cards: **01 · Change Context**, **02 · Blast-Radius Impact Map**, and **03 · 3-Plane Verification & Decision Record**.

<p align="center">
  <img src="assets/Screenshot%20From%202026-10-25%2002-32-11.png" width="100%" alt="Garuda Change Review Workbench & Policy Sandbox">
</p>

### `07` — Evidence & Ledger, `08` — Agents, `09` — Data & Logic Audit, `§` — Decision Lineage
- **`07 — Evidence & Ledger`**: Consolidated RFC 6962 Merkle root header + append-only audit table with an expandable Decision Record & Raw Telemetry drawer.
- **`08 — Agents & Sessions`**: Live MCP agent sessions and multi-agent task handoff timeline.
- **`09 — Data & Logic Audit`**: Tabbed **Pillar 1 · Multi-Team Formula Weights (`0.0–1.0`)** and **Pillar 2 · Cross-Team Data & Logic Discrepancies (`48 Cr vs 4.8 Cr`)**.
- **`§ — Decision Lineage`**: MADR-compatible Architectural Decision Records and supersession chains.

<p align="center">
  <img src="assets/Screenshot%20From%202026-10-25%2002-32-25.png" width="100%" alt="Garuda Contradictions Forensic Queue & Cryptographic Audit Ledger">
</p>

---

## Everyday CLI & MCP Workflows

For the full decision matrix across all CLI commands and **54 MCP tools**, see **[`docs/CLI_DECISION_GUIDE.md`](docs/CLI_DECISION_GUIDE.md)**:

| What You Want to Do | CLI Command (`garuda ...`) | MCP Tool (`garuda-mcp`) |
| :--- | :--- | :--- |
| **1-Call Pre-Edit Context & Blast Radius** | `garuda talon brief <Symbol>` | `garuda.preflight_bundle` / `garuda.talon_pre_edit` |
| **Post-Edit Diff Check & Self-Repair Prompt** | `garuda talon guard --diff HEAD --fail-on-block` | `garuda.talon_verify_edit` |
| **Reproduce an Exact Build & Dirty-Tree Verdict** | `garuda check --target linux/amd64 --cgo=false --commit <sha> --worktree-dirty --dirty-tree-hash <hash> --workspace <ws>` | `garuda.verify_receipt` |
| **Verify Cross-OS Build Matrix (`linux`/`darwin`/`windows`/`cgo`)** | `garuda check --matrix --workspace default` | `garuda.build_matrix_check` |
| **Compare Competing Team Formulas (`0.0–1.0` Weights)** | `garuda audit logic "Loss Ratio" --team ops` | `garuda.disambiguate_logic` / `garuda.decide_logic_variant` |
| **Audit Cross-Team Data & Unit-Scaling (`48 Cr` vs `4.8 Cr`)** | `garuda audit data-logic --metric "Claims Paid" --team-a-cr 48.0 --team-b-cr 4.8` | `garuda.audit_data_discrepancy` / `garuda.reconcile_data_audit` |
| **Inspect 3-Tier Surface Criticality** | `garuda audit criticality` | `garuda.criticality_tiers` |
| **Auto-Generate & Seal Base Rulebook** | `garuda rulebook generate && garuda rulebook approve --by "Lead Architect"` | `garuda.synthesize_rulebook` / `garuda.approve_base_rulebook` |
| **Export Signed Compliance Audit Trail** | `garuda policy list` | `garuda.audit_export` (`csv` / `json` / `md`) |

---

## Documentation Suite by Role

| Document | Who It's For | What You'll Learn |
| :--- | :--- | :--- |
| **[`docs/PERSONA_ONBOARDING_GUIDE.md`](docs/PERSONA_ONBOARDING_GUIDE.md)** | **All Users (Start Here!)** | 5-minute tailored onboarding paths for Software Engineers, Staff Architects, Data/Audit Leads, Founders/PMs, DevOps, and AI Agents. |
| **[`INSTALL.md`](INSTALL.md)** | **Everyone** | 1-command zero-dependency setup across Linux, macOS, Windows PowerShell, Chromebook, and Android + 20-scenario troubleshooting table. |
| **[`docs/WALKTHROUGH.md`](docs/WALKTHROUGH.md)** | **Product & Engineering Teams** | Screen-by-screen tour of the redesigned 10-surface progressive-disclosure `/dashboard` PWA, Talon harness, and IDE integrations. |
| **[`docs/DATA_AND_LOGIC_AUDIT_GUIDE.md`](docs/DATA_AND_LOGIC_AUDIT_GUIDE.md)** | **Engineering, Data & Audit Leads** | How `0.0–1.0` formula weighting, active contradiction quarantine, and the `data_logic_audit_ledger` work without FinOps regulatory overhead. |
| **[`docs/FINANCE_AND_AUDIT_GUIDE.md`](docs/FINANCE_AND_AUDIT_GUIDE.md)** | **RevOps, Actuaries & Controllers** | Deep dive into the 48 Crore vs. 4.8 Crore unit-scaling case study, surface criticality freezes, and RFC 6962 audit proofs. |
| **[`docs/CLI_DECISION_GUIDE.md`](docs/CLI_DECISION_GUIDE.md)** | **Developers & AI Agent Builders** | Complete "When-to-Use-What" decision guide mapping every CLI command to all **54 MCP tools**. |
| **[`PLAYBOOK.md`](PLAYBOOK.md)** | **DevOps & Platform Engineers** | End-to-end operational runbook covering CI gates, `--worktree-dirty` receipts, OTel runtime telemetry, and multi-agent handoffs. |
| **[`docs/API.md`](docs/API.md)** | **Tooling & Integration Engineers** | Complete CLI flag reference, 54 MCP tool schemas, `GARUDA_MCP_READ_ONLY` / `GARUDA_MCP_APPROVER` controls, 9 YAML predicates, and HTTP REST API. |
| **[`docs/CAPABILITIES.md`](docs/CAPABILITIES.md)** | **Staff Architects & Technical Evaluators** | 8-plane capability and verification matrix (`24/24` compiler benchmarks, `63/63` MCP checks, `18/18` security unit tests). |
| **[`docs/SPECS.md`](docs/SPECS.md)** | **Systems & Security Engineers** | Formal `v1.9.9` architecture specification, PostgreSQL migrations (`001`–`095`), `BuildContext` hashing, and RFC 6962 Merkle proofs. |
| **[`EVIDENCE.md`](EVIDENCE.md)** | **Technical Due Diligence** | Reproducible verification suite outputs and benchmark evidence. |
| **[`CHANGELOG.md`](CHANGELOG.md)** | **Everyone** | Full release history through **`v1.9.9`**. |

---

<p align="center">
  <strong>Let AI agents write fast. Let Garuda keep your architecture, formulas, and data whole.</strong><br>
  Released under the <a href="LICENSE">Apache 2.0 License</a>.
</p>
