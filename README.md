<p align="center">
  <img src="assets/Garuda%20Final%20logo.png" width="150" alt="Garuda Logo">
</p>

<h1 align="center">Garuda (<code>v1.9.9</code>)</h1>

<p align="center">
  <strong>AI writes code at 100 mph. Garuda makes sure your architecture, business formulas, and cross-team numbers survive the ride.</strong><br>
  <em>The local-first change-control gate, active AI agent harness (<code>Garuda Talon</code>), human-in-the-loop Data &amp; Logic Audit engine, <strong>54-tool MCP server</strong>, and <strong>10-surface progressive-disclosure Software Intelligence Dashboard</strong>.</em>
</p>

<p align="center">
  <a href="https://github.com/myshra777-ai/garuda/actions/workflows/ci.yml"><img src="https://github.com/myshra777-ai/garuda/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/myshra777-ai/garuda/actions/workflows/garuda-ci.yml"><img src="https://github.com/myshra777-ai/garuda/actions/workflows/garuda-ci.yml/badge.svg" alt="Garuda CI & Governance"></a>
  <a href="https://github.com/myshra777-ai/garuda/actions/workflows/release-packages.yml"><img src="https://img.shields.io/badge/GHCR_&_Marketplace-Ready-10b981.svg" alt="GitHub Packages & Marketplace"></a>
  <a href="capabilities.json"><img src="https://img.shields.io/badge/version-v1.9.9-10b981.svg" alt="Version v1.9.9"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache--2.0-blue.svg" alt="Apache 2.0 License"></a>
  <a href="EVIDENCE.md"><img src="https://img.shields.io/badge/verified-24%2F24_Fixtures_%7C_63%2F63_MCP_Checks_(54_Tools)-f59e0b.svg" alt="Verification Evidence"></a>
</p>

<p align="center">
  <a href="#what-is-garuda-in-plain-english">What Is Garuda?</a> ·
  <a href="#watch-the-20-second-cinematic-intro">20s Video Demo</a> ·
  <a href="#install-in-one-command-zero-dependencies">1-Command Install</a> ·
  <a href="#whats-new-in-v199">What's New in v1.9.9</a> ·
  <a href="#tour-the-futuristic-10-surface-dashboard-dashboard-pwa">Visual Dashboard Tour</a> ·
  <a href="#everyday-cli--mcp-workflows">Everyday CLI &amp; MCP Guide</a> ·
  <a href="#honest-capabilities-scope--current-boundaries">Capabilities &amp; Honest Limits</a> ·
  <a href="#documentation-suite-by-role">Full Documentation Suite</a>
</p>

---

## What Is Garuda? (In Plain English)

Today, AI coding tools like **Cursor**, **Claude Code**, **Codex**, and **Windsurf** can write a thousand lines of code in thirty seconds. The code compiles, unit tests pass, and the pull request looks green.

**But how do you know that fast AI-generated code didn't quietly break something important?**

In real engineering teams, four silent failures happen every week when humans and AI agents ship fast:

1. **Silent Architecture Shortcuts**: An AI agent makes an API handler talk straight to `database/sql` or an external payment API to "make the prompt work," bypassing your security, retry, idempotency, or store layers.
2. **Competing Business Formulas Across Teams**: Your **Ops**, **Marketing**, and **Management** teams each write their own formula for the same metric (like *Loss Ratio*, *Retention*, or *Net Margin*). Without a referee, an AI agent copies a marketing campaign formula into your core pricing router—breaking 142 downstream callers.
3. **10x Cross-Team Number Mismatches (`48.00 Crore` vs. `4.80 Crore`)**: One team's code divides raw Rupees by `10,000,000` and reports **`48.00 Crore`**, while another team's rollup query divides Lakhs by `1,000` instead of `100` and reports **`4.80 Crore`**. Nobody catches the 10x unit-scaling bug until an executive review or external audit.
4. **Undocumented "Tribal Knowledge"**: Most growing codebases don't have up-to-date Architecture Decision Records (ADRs). Rules live only in senior engineers' heads, so new developers and AI agents don't know which files are safe to edit (`LOW`) and which ones are mission-critical (`HIGH_SEVERE`).

**Garuda (`v1.9.9`) is your local-first architectural safety net and truth engine.**
It sits quietly on your machine (and in CI), reads your real code (**Go**, **Python**, and **TypeScript**), connects it to your architecture rules and business formulas, guides your AI coding agents before and after every edit via **Garuda Talon** and **54 deterministic MCP tools**, and gives you a calm, futuristic **10-Surface Web Dashboard** where humans make the final call.

<p align="center">
  <img src="assets/garuda%20core%20about%20page.png" width="100%" alt="Garuda Core Overview — Epistemic Software Intelligence & Active Governance">
</p>

---

## Watch the 20-Second Cinematic Intro

See how Garuda turns invisible code dependencies, architecture rules, business formulas, and AI agent edits into a calm, interactive control room:

<p align="center">
  <a href="assets/Garuda%20dashboard%20demo.mp4" title="Click to play the 20-second Garuda Dashboard Demo Video">
    <img src="assets/how%20to%20get%20started%20in%205%20minutes.png" width="92%" alt="Watch the 20-Second Garuda Dashboard Demo Video">
  </a>
</p>

<p align="center">
  🎬 <strong><a href="assets/Garuda%20dashboard%20demo.mp4">▶ Click here to play / download <code>assets/Garuda dashboard demo.mp4</code></a></strong><br>
  <sub><em>Note: GitHub's Markdown security filter strips inline HTML <code>&lt;video&gt;</code> players for repository paths; clicking the preview or link above streams <code>Garuda dashboard demo.mp4</code> directly.</em></sub>
</p>

---

## Install in One Command (Zero Dependencies)

You do **not** need Go, PostgreSQL, Docker, GCC, Homebrew, or Xcode pre-installed, and you do **not** need `sudo` or Administrator rights. Garuda automatically sets up a portable local database (`127.0.0.1:5433`), compiles all **4 zero-CGO binaries** (`garuda`, `garuda-api`, `garuda-mcp`, `garuda-talon`), wires your AI coding tools, and opens the dashboard:

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

Once setup finishes, open **`http://localhost:8080/login`** in your browser. Any time your computer restarts or you want to verify system health, run:

```bash
./bin/garuda doctor --fix
```

<p align="center">
  <img src="assets/Create%20new%20workspace.png" width="100%" alt="Create a New Garuda Workspace & Index Repositories in Minutes">
</p>

---

## What's New in `v1.9.9`

### 1. Futuristic Progressive-Disclosure Dashboard (`10` Surfaces, Light & Graphite Dark Mode)
- **Easy on the Eyes, Zero Clutter**: Inspired by Linear, Vercel, Stripe, and Apple Human Interface Guidelines, every dashboard surface shows **at most 4 primary cards on first view**. Secondary tables, drift breakdowns, and raw telemetry sit one click away inside smooth expandable `<details>` drawers and sub-tabs.
- **Smooth `cubic-bezier(0.16, 1, 0.3, 1)` Motion**: GPU-accelerated staggered entrance transitions (`surfaceIn`), subtle card elevation on hover, a 3-step tabular-numeral type hierarchy, and a 1-click toggle between **Warm Alabaster Light Mode** and **Obsidian Graphite Dark Mode** (`☀ / ☾`).

### 2. Human-in-the-Loop Data & Logic Audit Ledger (`09 — Data & Logic Audit`)
- **`0.0–1.0` Multi-Team Formula Weighting**: When **Ops**, **Marketing**, and **Management** code different formulas for the same concept (such as *Loss Ratio* or *Retention*), Garuda never lets AI silently pick or overwrite one. It computes `0.00–1.00` team weights (`0.70` active team vs. `0.10`/`0.10`), ranks each option by *Least Architectural Harm*, and requires human co-signing (`garuda audit logic`).
- **Cross-Team Data & Unit-Scale Discrepancy Detector**: Traces AST and SQL data lineage across teams to catch 10x unit-scaling bugs (such as `48.00 Crore` vs. `4.80 Crore` from dividing Lakhs by `1000` instead of `100`) before numbers ever reach an external review (`garuda audit data-logic`, backed by `data_logic_audit_ledger` and `data_audit_reconciliations`).

### 3. 3-Tier Surface Criticality Classification (`HIGH_SEVERE` · `MEDIUM` · `LOW`)
- **`HIGH_SEVERE` (Core Surfaces — Frozen)**: Homepage, Search Redirect Router (`142` callers), and Core Settlement Ledgers are locked against automated AI edits (`HUMAN_FREEZE_DO_NOT_TOUCH`).
- **`MEDIUM` (Non-Blocking Features — Human Trade-Off)**: Secondary features like Wishlist Loaders (`19` callers) allow informed human trade-offs.
- **`LOW` (UI & Formatting — Safe)**: Isolated UI box alignment and formatting helpers (`4–6` callers) proceed cleanly.

### 4. Auto-Synthesized Base Rulebook Studio (`04 — Docs & Rules Graph`)
- Don't have time to write ADRs from scratch? Garuda scans your repository, discovers implicit layer boundaries, criticality tiers, and formula weights, and drafts `.garuda/DRAFT_BASE_RULEBOOK.md`. Review and edit it right in the browser, then click **Approve & Seal Rulebook** (`garuda rulebook generate` / `garuda rulebook approve`).

### 5. Hardened 54-Tool MCP Server & Deterministic Dirty-Worktree Receipts (`63/63` Checks)
- **54 Deterministic MCP Tools**: Includes `garuda.preflight_bundle` (1-call pre-edit context), `garuda.verify_receipt` (RFC 6962 cryptographic verification), `garuda.diff_snapshots` (revision-to-revision graph diff), `garuda.audit_export` (CSV/JSON/Markdown compliance export), and `garuda.create_workspace` (isolated sandbox workspaces).
- **Exact Dirty-Worktree & Cross-OS Reproducibility (`--worktree-dirty`, `--dirty-tree-hash`, `--matrix`)**: Every `garuda check` and `garuda talon guard` verdict binds `GOOS/GOARCH`, `CGO_ENABLED`, build tags, git commit SHA, and dirty working-tree hash into a single copy-pasteable reproduce command.

---

## Tour the Futuristic 10-Surface Dashboard (`/dashboard` PWA)

Garuda's web interface runs 100% locally at `http://localhost:8080/dashboard` (and installs as a standalone desktop or mobile PWA):

### `01` — Overview & 5-Minute Onboarding: Calm Command Center
Shows **1 Hero System Integrity Card** (`ALLOW` / `BLOCK`) + **3 clickable core metric cards** (*Indexed Topology*, *Epistemic & Policy Gate*, *RFC 6962 Provenance Ledger*), with progressive-disclosure drawers for *Needs Attention & Documentation Drift* and *Architectural Hubs & Indexed Repositories*.

<p align="center">
  <img src="assets/how%20to%20get%20started%20in%205%20minutes.png" width="100%" alt="01 — Overview & How to Get Started in 5 Minutes">
</p>

### `02` & `03` — Explore Index (`⌘K`) & Architecture Map Without the Hairball
Search packages, functions, methods, structs, and interfaces with instant signature preview, and explore 1-hop and 2-hop symbol neighborhoods (`Callers → Symbol → Dependencies`) on a clean interactive canvas without the unreadable node hairball.

<p align="center">
  <img src="assets/see%20the%20entire%20system%20without%20hairball.png" width="100%" alt="03 — See the Entire System Without the Hairball">
</p>

### `04` & `05` — Docs & Rules Graph + Talon Risk Hub: Decode Signal Across Your Architecture
Switches cleanly between the **4-Tier Document-to-Code Lineage Graph** (`Document → Rule → Policy → Code`), the **Auto-Synthesized Base Rulebook Studio**, and **Garuda Talon's** active IDE/Git enforcement gates, risk hub leaderboard, and cross-OS build matrix.

<p align="center">
  <img src="assets/decode%20signal%20across%20architecture.png" width="100%" alt="04 & 05 — Decode Signal Across Architecture, Docs & Talon Guardrails">
</p>

### `06`, `07` & `08` — Change Review Workbench, Cryptographic Evidence Ledger & Agent Sessions
Inspect any symbol or git diff across three balanced cards (**01 · Change Context**, **02 · Blast-Radius Impact Map**, **03 · 3-Plane Verification & Decision Record**), verify RFC 6962 Merkle roots, and monitor active MCP agent sessions and task handoffs.

<p align="center">
  <img src="assets/Screenshot%20From%202026-11-01%2022-31-46.png" width="100%" alt="06 & 07 — Change Review Workbench & Cryptographic Evidence Ledger">
</p>

### `09` & `§` — Human-in-the-Loop Data & Logic Audit Ledger & ADR Decision Lineage
Resolves competing cross-team formulas using **`0.0–1.0` team weights** (*Pillar 1*) and catches **10x unit-scaling discrepancies** like `48.00 Crore` vs. `4.80 Crore` (*Pillar 2*) before external reviews, backed by MADR decision lineage.

<p align="center">
  <img src="assets/cross%20team%20audit%20and%20analysis.png" width="100%" alt="09 — Cross-Team Data & Logic Audit and Formula Weight Analysis">
</p>

---

## Everyday CLI & MCP Workflows (`54` Registered Tools)

Every action in the dashboard is also available from a clean, scannable CLI and across all **54 MCP tools** in Cursor, Claude Code, Codex, and Windsurf (see [`docs/CLI_DECISION_GUIDE.md`](docs/CLI_DECISION_GUIDE.md)):

<p align="center">
  <img src="assets/scannable%20command%20line.png" width="100%" alt="Scannable Garuda CLI & MCP Workflows">
</p>

| What You Want to Do | CLI Command (`garuda ...`) | MCP Tool (`garuda-mcp`) |
| :--- | :--- | :--- |
| **1-Call Pre-Edit Context & Blast Radius** | `garuda talon brief <Symbol>` | `garuda.preflight_bundle` / `garuda.talon_pre_edit` |
| **Post-Edit Diff Check & Self-Repair Prompt** | `garuda talon guard --diff HEAD --fail-on-block` | `garuda.talon_verify_edit` |
| **Reproduce an Exact Build & Dirty-Tree Verdict** | `garuda check --target linux/amd64 --cgo=false --commit <sha> --worktree-dirty --dirty-tree-hash <hash> --workspace <ws>` | `garuda.verify_receipt` / `garuda.snapshot` |
| **Verify Cross-OS Build Matrix (`4` Profiles)** | `garuda check --matrix --workspace default` | `garuda.build_matrix_check` |
| **Compare Competing Team Formulas (`0.0–1.0` Weights)** | `garuda audit logic "Loss Ratio" --team ops` | `garuda.disambiguate_logic` / `garuda.decide_logic_variant` |
| **Audit Cross-Team Data & Unit-Scaling (`48 Cr` vs `4.8 Cr`)** | `garuda audit data-logic --metric "Claims Paid" --team-a-cr 48.0 --team-b-cr 4.8` | `garuda.audit_data_discrepancy` / `garuda.reconcile_data_audit` |
| **Inspect 3-Tier Surface Criticality** | `garuda audit criticality` | `garuda.criticality_tiers` |
| **Auto-Generate & Seal Base Rulebook** | `garuda rulebook generate && garuda rulebook approve --by "Lead Architect"` | `garuda.synthesize_rulebook` / `garuda.approve_base_rulebook` |
| **Export Signed Compliance Audit Trail** | `garuda policy list` | `garuda.audit_export` (`csv` / `json` / `md`) |

---

## Canonical Capability & Verification Manifest (`P3`)

All counts, versions, and tool definitions across Garuda's binaries and documentation are synchronized from **[`capabilities.json`](capabilities.json)** and **[`internal/version/manifest.go`](internal/version/manifest.go)** (verified in CI via `python3 scripts/sync_docs_from_manifest.py --check`):

| Canonical Dimension | Count / Value | Source of Truth |
| :--- | :--- | :--- |
| **Canonical Release Version** | **`v1.9.9`** (`2026-11-01`) | `capabilities.json` · `internal/version/manifest.go` · `version.txt` |
| **Core Zero-CGO Binaries** | **`4`** (`garuda`, `garuda-mcp`, `garuda-api`, `garuda-talon`) | `cmd/*` (`CGO_ENABLED=0`) |
| **Registered MCP Tools (`bin/garuda-mcp`)** | **`54`** (`41` read-only · `13` mutating) | `cmd/garuda-mcp/main.go` · `tools_whitelist.go` |
| **MCP Positive & Negative Behavioral Tests** | **`54/54` tools** (`108` positive + negative subtests) | `cmd/garuda-mcp/behavioral_tools_test.go` |
| **Live MCP Protocol Verification Suite** | **`63/63` checks** | `scripts/mcp_verify.py` |
| **Dashboard Working Surfaces (`/dashboard`)** | **`10` surfaces** (`≤ 4` primary cards on first view) | `internal/api/templates/dashboard.html` |
| **Dashboard HTTP API Endpoints** | **`19` endpoints** | `internal/api/static/dashboard.js` |
| **Idempotent PostgreSQL Migrations** | **`95`** (`001`–`095_data_logic_audit_ledger_and_hygiene.sql`) | `migrations/` |
| **Graph-Native YAML Policy Predicates** | **`9` predicates** | `internal/policy/` |

### Strict Separation of Evaluation Tiers (Fixtures vs. Simulations vs. Live AST/LLM)

| Evaluation Tier | Scope & Activation Flag | Default in Production Workspaces | What It Measures & Verifies |
| :--- | :--- | :--- | :--- |
| **Tier 1 — Compiler Fixture Results** (`24/24`) | `make benchmark-strict` or `include_fixtures=true` | **EXCLUDED BY DEFAULT** (`ExcludeFixturesSQL` filters out `test/fixtures/`, `testdata/`, `garuda-bench/`, `conftest`) | Static `go/types`, `go.work`, Python AST, and TypeScript AST extraction precision/recall (`100%`) and checkout-path-invariant cache fingerprints (`TestDoubleRunDeterminism`). |
| **Tier 2 — Synthetic Grounding & Contradiction Simulations** (`10/10`) | `scripts/runtime_contradiction_test.py`, `--demo-fixtures`, `simulate_violation`, `simulate_drift` | **OPT-IN ONLY** (Empty workspaces return `status: "INSUFFICIENT_DATA"` when `include_fixtures=false`) | Quarantine of injected unapproved OTel runtime spans (`10/10`), Talon self-repair prompt synthesis under simulated `BLOCK`/drift, and reference case studies (`Loss Ratio` `0.0–1.0` weights & `48.00 Cr vs 4.80 Cr`). |
| **Tier 3 — Live AST, RFC 6962 Ledger & Live LLM/MCP Sessions** (`63/63` & `54/54`) | Default live mode (`scope: "production_only"`, `GARUDA_TALON_ENFORCE=true`) | **ACTIVE BY DEFAULT** | Live repository symbol & blast-radius resolution, cryptographic tenant/workspace isolation, `GARUDA_MCP_READ_ONLY=true` mutation blocking, `GARUDA_MCP_APPROVER` human co-sign verification, and SHA-256 `receipt_hash` validation. |

---

## Honest Capabilities, Scope & Current Boundaries

Garuda is built on epistemic honesty (`SUPPORTED`, `UNVERIFIED`, `CONTRADICTED`). Here is an upfront summary of what Garuda does deterministically today and where its current boundaries lie:

| Area | What Works Today (`v1.9.9`) | Current Boundary / Honest Limitation |
| :--- | :--- | :--- |
| **Supported Languages** | **Go** (`go/types` + `go.work`), **Python** (AST), and **TypeScript / JavaScript** (AST) with `24/24` strict benchmark fixtures at 100% precision/recall. | Rust, Java, C#, and C++ are not yet parsed by the zero-CGO static analyzer; polyglot repos containing those languages index Go/Python/TS files while skipping unsupported extensions. |
| **Static AST vs. Runtime Telemetry** | Static call/import graphs are extracted immediately on `garuda analyze`. Dynamic reflection, interface dispatch, or runtime-only HTTP hops are marked `UNVERIFIED` until OpenTelemetry spans are ingested via `/api/v1/telemetry/ingest`. | Garuda never guesses or halluccinates runtime edges from static code alone—unobserved runtime paths remain explicitly `UNVERIFIED` until runtime telemetry is provided. |
| **Human-in-the-Loop Formula & Data Audit** | Detects competing formula variants (`0.0–1.0` weights) and unit-scale mismatches (`48.00 Cr vs 4.80 Cr`). Unauthenticated MCP agents can only record `PROPOSED_BY_AGENT_AWAITING_HUMAN_CLI_COSIGN`. | Final approval (`APPROVED_BY_HUMAN` / `RECONCILED_BY_HUMAN`) intentionally requires human CLI co-signing or an explicit `GARUDA_MCP_APPROVER` identity—AI agents cannot self-approve business formulas. |
| **Empty Workspaces vs. Demo Case Studies** | In production (`include_fixtures=false`), an unindexed or empty workspace returns `status: "INSUFFICIENT_DATA"` rather than fake demo rows. | To explore the pre-loaded *Loss Ratio* and *48 Cr vs 4.8 Cr* interactive case studies on a fresh workspace without matching domain code, pass `--demo-fixtures` (CLI) or `include_fixtures=true` (MCP/UI). |

---

## GitHub Packages & GitHub Marketplace Integration

```yaml
# Use Garuda Talon directly from GitHub Marketplace in any workflow:
- name: Garuda Talon Architecture & Data-Logic Gate
  uses: myshra777-ai/garuda@v1.9.9
  with:
    workspace: default
    policy-dir: policies
    fail-on-block: "true"
    matrix: "true"
```

```bash
# Pull the multi-arch OCI container from GitHub Packages (GHCR):
docker pull ghcr.io/myshra777-ai/garuda:v1.9.9
```

---

## Documentation Suite by Role

| Document | Who It's For | What You'll Learn |
| :--- | :--- | :--- |
| **[`docs/PERSONA_ONBOARDING_GUIDE.md`](docs/PERSONA_ONBOARDING_GUIDE.md)** | **All Users (Start Here!)** | 5-minute tailored onboarding paths for Software Engineers, Staff Architects, Data/Audit Leads, Founders/PMs, DevOps, and AI Agents. |
| **[`INSTALL.md`](INSTALL.md)** | **Everyone** | 1-command zero-dependency setup across Linux, macOS, Windows PowerShell, Chromebook, and Android + 20-scenario troubleshooting table. |
| **[`docs/WALKTHROUGH.md`](docs/WALKTHROUGH.md)** | **Product & Engineering Teams** | Screen-by-screen tour of the redesigned 10-surface progressive-disclosure `/dashboard` PWA, Talon harness, and IDE integrations. |
| **[`docs/DATA_AND_LOGIC_AUDIT_GUIDE.md`](docs/DATA_AND_LOGIC_AUDIT_GUIDE.md)** | **Engineering, Data & Audit Leads** | How `0.0–1.0` formula weighting, active contradiction quarantine, and the `data_logic_audit_ledger` work without FinOps regulatory overhead. |
| **[`docs/FINANCE_AND_AUDIT_GUIDE.md`](docs/FINANCE_AND_AUDIT_GUIDE.md)** | **RevOps, Actuaries & Controllers** | Deep dive into the `48 Crore vs. 4.8 Crore` unit-scaling case study, surface criticality freezes, and RFC 6962 audit proofs. |
| **[`docs/CLI_DECISION_GUIDE.md`](docs/CLI_DECISION_GUIDE.md)** | **Developers & AI Agent Builders** | Complete "When-to-Use-What" decision guide mapping every CLI command to all 54 MCP tools. |
| **[`PLAYBOOK.md`](PLAYBOOK.md)** | **DevOps & Platform Engineers** | End-to-end operational runbook covering CI gates, `--worktree-dirty` receipts, OTel runtime telemetry, and multi-agent handoffs. |
| **[`docs/API.md`](docs/API.md)** | **Tooling & Integration Engineers** | Complete CLI flag reference, 54 MCP tool schemas, `GARUDA_MCP_READ_ONLY` / `GARUDA_MCP_APPROVER` controls, 9 YAML predicates, and HTTP REST API. |
| **[`docs/CAPABILITIES.md`](docs/CAPABILITIES.md)** | **Staff Architects & Evaluators** | 8-plane capability and verification matrix (`24/24` compiler benchmarks, `63/63` MCP checks, `54/54` behavioral tool tests). |
| **[`docs/SPECS.md`](docs/SPECS.md)** | **Systems & Security Engineers** | Formal `v1.9.9` architecture specification, PostgreSQL migrations (`001`–`095`), `BuildContext` hashing, and RFC 6962 Merkle proofs. |
| **[`EVIDENCE.md`](EVIDENCE.md)** | **Technical Due Diligence** | Reproducible verification suite outputs and 3-tier benchmark evidence. |
| **[`CHANGELOG.md`](CHANGELOG.md)** | **Everyone** | Full release history through `v1.9.9`. |

---

<p align="center">
  <strong>Let AI agents write fast. Let Garuda keep your architecture, formulas, and data whole.</strong><br>
  Released under the <a href="LICENSE">Apache 2.0 License</a>.
</p>
