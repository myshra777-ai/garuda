#!/usr/bin/env python3
# Copyright 2026 Rohit Mishra
# SPDX-License-Identifier: Apache-2.0
"""
Canonical Documentation, Assets & Capability Manifest Synchronizer (P3).
Reads capabilities.json + cmd/garuda-mcp/main.go + assets/ and generates/updates:
  - version.txt
  - docs/source/capabilities.yaml
  - README.md
  - docs/API.md & docs/API_REFERENCE.md
  - docs/WALKTHROUGH.md
  - docs/CAPABILITIES.md
  - EVIDENCE.md
  - RELEASE_NOTES.md
  - CHANGELOG.md
"""

import argparse
import glob
import json
import os
import re
import shutil
import sys


def discover_assets(repo_root: str) -> dict:
    assets_dir = os.path.join(repo_root, "assets")
    shots_dir = os.path.join(assets_dir, "screenshots")
    os.makedirs(assets_dir, exist_ok=True)
    os.makedirs(shots_dir, exist_ok=True)

    # 1. Find the newest logo in assets/
    logo_candidates = []
    for ext in ("*.png", "*.svg", "*.webp", "*.jpg"):
        for p in glob.glob(os.path.join(assets_dir, ext)):
            base = os.path.basename(p).lower()
            if "logo" in base or "garuda_minimal" in base or base == "garuda.png":
                logo_candidates.append((os.path.getmtime(p), p))

    primary_logo_rel = "assets/garuda-logo.png"
    if logo_candidates:
        logo_candidates.sort(reverse=True)
        newest_logo = logo_candidates[0][1]
        for alias in ("garuda-logo.png", "Garuda_minimal_logo.png", "garuda_logo.png"):
            dst = os.path.join(assets_dir, alias)
            if os.path.abspath(newest_logo) != os.path.abspath(dst):
                try:
                    shutil.copy2(newest_logo, dst)
                except Exception:
                    pass

    # 2. Collect all screenshot files in assets/ and assets/screenshots/
    all_images = []
    for folder, rel_prefix in [(assets_dir, "assets"), (shots_dir, "assets/screenshots")]:
        if not os.path.isdir(folder):
            continue
        for fname in sorted(os.listdir(folder)):
            low = fname.lower()
            if not low.endswith((".png", ".webp", ".jpg", ".gif", ".svg")):
                continue
            if "logo" in low:
                continue
            # Prefer accurately named screenshots over raw 'Screenshot From ...' timestamps
            if low.startswith("screenshot from") or low.startswith("screenshot_from"):
                continue
            rel_path = f"{rel_prefix}/{fname}"
            all_images.append((fname, rel_path))

    def pick_best(keywords, fallback):
        # First check direct assets/ (newest/custom uploaded), then assets/screenshots/
        for fname, rel_path in all_images:
            low = fname.lower()
            if any(k in low for k in keywords):
                return rel_path
        return fallback

    surface_shots = {
        "overview": pick_best(["overview", "briefing", "command-center", "01"], "assets/screenshots/workspace-overview.png"),
        "search": pick_best(["search", "explore", "02"], "assets/screenshots/global_search.png"),
        "architecture": pick_best(["top_level_repo_architecture", "architecture", "neighborhood", "03"], "assets/screenshots/top_level_repo_architecture.png"),
        "docs_graph": pick_best(["docs-graph", "docs_graph", "rulebook", "repository-topology", "04"], "assets/screenshots/repository-topology.png"),
        "talon": pick_best(["talon", "risk", "clusters", "05"], "assets/screenshots/workspace-clusters-graph.png"),
        "review": pick_best(["review", "workbench", "cross-repo", "06"], "assets/screenshots/cross-repo-graph.png"),
        "trust": pick_best(["trust", "evidence", "ledger", "merkle", "07"], "assets/screenshots/evidence-cryptographic-trust.png"),
        "contradictions": pick_best(["contradiction", "14repo-contradictions", "drift"], "assets/screenshots/dashboard-14repo-contradictions.png"),
        "ast_graph": pick_best(["low-level-ast", "ast-graph", "ast"], "assets/screenshots/low-level-ast-graph.png"),
        "multi_repo": pick_best(["14-repo", "multi-repo", "workspace-graph"], "assets/screenshots/14-repo-workspace-graph.png"),
        "ide_graph": pick_best(["ide-graph"], "assets/screenshots/ide-graph-view.png"),
        "ide_contradictions": pick_best(["ide-contradiction"], "assets/screenshots/ide-contradictions.png"),
    }

    return {
        "logo": primary_logo_rel,
        "surface_shots": surface_shots,
        "all_images": all_images,
    }


def extract_mcp_tools(repo_root: str):
    main_go = open(os.path.join(repo_root, "cmd/garuda-mcp/main.go")).read()
    tools = re.findall(r'"name":\s*"(garuda\.[a-z0-9_.]+)",\s*\n\s*"description":\s*"([^"]+)"', main_go)
    return tools


def generate_readme(manifest: dict, assets_info: dict, tools: list) -> str:
    ver = manifest["version"]
    counts = manifest["counts"]
    shots = assets_info["surface_shots"]
    logo = assets_info["logo"]

    # Build gallery rows for any additional custom screenshots in assets/
    seen_paths = set(shots.values())
    extra_gallery = []
    for fname, rel_path in assets_info["all_images"]:
        if rel_path not in seen_paths:
            seen_paths.add(rel_path)
            label = fname.rsplit(".", 1)[0].replace("-", " ").replace("_", " ").title()
            extra_gallery.append((label, rel_path))

    extra_gallery_md = ""
    if extra_gallery:
        rows = "\n".join(
            f"| **{lbl}** (`{path}`) | <img src=\"{path}\" width=\"420\" alt=\"{lbl}\"> |"
            for lbl, path in extra_gallery
        )
        extra_gallery_md = f"""
### Additional Workspace, AST & Multi-Repository Views (`assets/`)

| View & Asset Path | Preview |
| :--- | :--- |
{rows}
"""

    return f"""<p align="center">
  <img src="{logo}" width="140" alt="Garuda Logo">
</p>

<h1 align="center">Garuda (<code>{ver}</code>)</h1>

<p align="center">
  <strong>AI writes code at 100 mph. Garuda makes sure your architecture, business formulas, and cross-team numbers survive the ride.</strong><br>
  <em>The local-first change-control gate, active AI agent harness (<code>Garuda Talon</code>), human-in-the-loop Data &amp; Logic Audit engine, <strong>{counts['mcp_tools']}-tool MCP server</strong>, and <strong>{counts['dashboard_surfaces']}-surface progressive-disclosure Software Intelligence Dashboard</strong>.</em>
</p>

<p align="center">
  <a href="https://github.com/myshra777-ai/garuda/actions/workflows/ci.yml"><img src="https://github.com/myshra777-ai/garuda/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/myshra777-ai/garuda/actions/workflows/garuda-ci.yml"><img src="https://github.com/myshra777-ai/garuda/actions/workflows/garuda-ci.yml/badge.svg" alt="Garuda CI & Governance"></a>
  <a href="https://github.com/myshra777-ai/garuda/actions/workflows/release-packages.yml"><img src="https://img.shields.io/badge/GHCR_&_Marketplace-Ready-10b981.svg" alt="GitHub Packages & Marketplace"></a>
  <a href="capabilities.json"><img src="https://img.shields.io/badge/version-{ver}-10b981.svg" alt="Version {ver}"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache--2.0-blue.svg" alt="Apache 2.0 License"></a>
  <a href="EVIDENCE.md"><img src="https://img.shields.io/badge/verified-{counts['compiler_fixture_benchmarks']}%2F{counts['compiler_fixture_benchmarks']}_Fixtures_%7C_{counts['mcp_protocol_verification_checks']}%2F{counts['mcp_protocol_verification_checks']}_MCP_Checks_({counts['mcp_tools']}_Tools)-f59e0b.svg" alt="Verification Evidence"></a>
</p>

<p align="center">
  <a href="#what-is-garuda-in-plain-english">What Is Garuda?</a> ·
  <a href="#install-in-one-command-zero-dependencies">1-Command Install</a> ·
  <a href="#canonical-capability--verification-manifest-p3">Canonical Manifest &amp; 3-Tier Evidence</a> ·
  <a href="#tour-the-futuristic-10-surface-dashboard-dashboard-pwa">Visual Dashboard Tour</a> ·
  <a href="#everyday-cli--mcp-workflows">CLI &amp; {counts['mcp_tools']}-Tool MCP Guide</a> ·
  <a href="#github-packages--github-marketplace-integration">GitHub Marketplace &amp; Packages</a> ·
  <a href="#documentation-suite-by-role">Documentation Suite</a>
</p>

---

## What Is Garuda? (In Plain English)

AI coding agents (**Cursor**, **Claude Code**, **Codex**, **Windsurf**) can generate a thousand lines of code in thirty seconds. The code compiles and unit tests pass—yet four critical classes of drift silently slip into production:

1. **Silent Layer & Boundary Shortcuts**: An AI agent makes an API handler query `database/sql` or an external payment client directly, bypassing retry, idempotency, or store layers.
2. **Competing Undocumented Business Formulas Across Teams**: **Ops**, **Marketing**, and **Management** each implement their own formula for the same concept (such as *Loss Ratio* or *Retention*). Without a referee, an agent copies a campaign formula into a core settlement router with 142 callers.
3. **10x Cross-Team Unit-Scaling Mismatches (`48.00 Crore` vs. `4.80 Crore`)**: One team's pipeline divides raw Rupees by `10,000,000` (`48.00 Crore`), while another divides Lakhs by `1,000` instead of `100` (`4.80 Crore`).
4. **Undocumented Tribal Architecture**: Most repositories lack up-to-date ADRs, leaving engineers and AI agents blind to which surfaces are mission-critical (`HIGH_SEVERE`) and which are safe to refactor (`LOW`).

**Garuda (`{ver}`) is your local-first architectural safety net and truth engine.** It indexes Go, Python, and TypeScript ASTs into a single PostgreSQL + RFC 6962 Merkle graph, enforces pre-edit and post-edit guardrails via **Garuda Talon** and **`{counts['mcp_tools']}` deterministic MCP tools**, and provides a calm, progressive-disclosure **`{counts['dashboard_surfaces']}`-Surface Web Dashboard**.

---

## Install in One Command (Zero Dependencies)

You do **not** need Go, PostgreSQL, Docker, GCC, Homebrew, or Xcode pre-installed, and you do **not** need `sudo` or Administrator rights:

```bash
# Linux, macOS (Apple Silicon & Intel), Chromebook (Crostini) & Android (Termux)
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash

# Or inside an existing cloned repository
./setup.sh
```

```powershell
# Windows 10 / 11 (PowerShell 5.1 & PowerShell 7 — No Admin Required)
Set-ExecutionPolicy Bypass -Scope Process -Force; irm https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.ps1 | iex
```

---

## Canonical Capability & Verification Manifest (`P3`)

All counts, versions, and tool definitions across Garuda's binaries and documentation are synchronized from **[`capabilities.json`](capabilities.json)** and **[`internal/version/manifest.go`](internal/version/manifest.go)** (verified in CI via `python3 scripts/sync_docs_from_manifest.py --check`):

| Canonical Dimension | Count / Value | Source of Truth |
| :--- | :--- | :--- |
| **Canonical Release Version** | **`{ver}`** (`{manifest['release_date']}`) | `capabilities.json` · `internal/version/manifest.go` · `version.txt` |
| **Core Zero-CGO Binaries** | **`{counts['binaries']}`** (`garuda`, `garuda-mcp`, `garuda-api`, `garuda-talon`) | `cmd/*` (`CGO_ENABLED=0`) |
| **Registered MCP Tools (`bin/garuda-mcp`)** | **`{counts['mcp_tools']}`** (`{counts['mcp_readonly_tools']}` read-only · `{counts['mcp_mutating_tools']}` mutating) | `cmd/garuda-mcp/main.go` · `tools_whitelist.go` |
| **MCP Positive & Negative Behavioral Tests** | **`{counts['mcp_behavioral_unit_tests']}/{counts['mcp_tools']}` tools** (`108` positive + negative subtests) | `cmd/garuda-mcp/behavioral_tools_test.go` |
| **Live MCP Protocol Verification Suite** | **`{counts['mcp_protocol_verification_checks']}/{counts['mcp_protocol_verification_checks']}` checks** | `scripts/mcp_verify.py` |
| **Dashboard Working Surfaces (`/dashboard`)** | **`{counts['dashboard_surfaces']}` surfaces** (`≤ 4` primary cards on first view) | `internal/api/templates/dashboard.html` |
| **Dashboard HTTP API Endpoints** | **`{counts['dashboard_api_endpoints']}` endpoints** | `internal/api/static/dashboard.js` |
| **Idempotent PostgreSQL Migrations** | **`{counts['sql_migrations']}`** (`001`–`095_data_logic_audit_ledger_and_hygiene.sql`) | `migrations/` |
| **Graph-Native YAML Policy Predicates** | **`{counts['yaml_policy_predicates']}` predicates** | `internal/policy/` |


### The Four-Tier Document-to-Code Governance Graph (`garuda.docs_graph` & Surface `04`)

Garuda connects four layers of an engineering organization into one inspectable document-to-code governance graph:

| Governance Tier | Layer | What It Connects |
| :--- | :--- | :--- |
| **Tier 1 — Organizational Intent** | ADRs, Specs, Invariants & YAML Policies | `docs/adr/*.md`, `docs/specs/*.md`, `policies/*.yaml`, `.garudarules` |
| **Tier 2 — Extracted Normative Rules** | `MUST`, `MUST_NOT`, `BLOCK`, `REVIEW`, `DENY` | Collision-free SHA-256 component nodes extracted from documents |
| **Tier 3 — Executable Governance** | Active YAML Policies & Talon Guardrails | Compiled `.garuda/TALON_CONTRACT.md` & deterministic policy verdicts |
| **Tier 4 — Actual Software (AST)** | Indexed Functions, Methods & Packages | Go (`go/types`), Python AST, and TypeScript AST entities & call edges |

> **Link Provenance Transparency**: Every document-to-code link returned by `garuda.docs_graph` and `/api/v1/dashboard/docs-graph` carries an explicit `match_method` (`"ast_exact"`, `"package_or_keyword_match"`, or `"fallback_unindexed_preview"`) and `graph_source_mode` (`"live_indexed_ast"` vs. `"fallback_preview_unindexed"`), distinguishing compiler-resolved AST links from keyword/package rules and unindexed workspace previews.

### Multi-Agent Coordination & Operational Continuity Layer

Garuda understands the work of agents, not just their code—preserving shared operational context as tasks move across concurrent AI coding agents:

| Capability | Implementation Surface | What It Enables |
| :--- | :--- | :--- |
| **MCP Session Tracking** | `cmd/garuda-mcp/session_tracking.go` | Identifies connected MCP clients/agents, active activity, and session duration. |
| **Tool Execution Telemetry** | `cmd/garuda-mcp/telemetry.go` | Records tool names, execution latency, success/error counts, and whitelist-filtered arguments (`filterArgs`). |
| **Agent-Specific Briefings** | `garuda.briefing` / `garuda.preflight_bundle` | Tracks when an agent was last briefed and reports workspace deltas since that watermark. |
| **Task Checkpoints** | `garuda.create_task` / `checkpoints` | Persists structured execution context and scope metadata for a task UUID. |
| **Serializable Atomic Handoffs** | `garuda.handoff` (`internal/store/handoff_store.go`) | Locks and validates agents and task ownership inside a serializable transaction, saves a checkpoint, transfers ownership, and fails closed on unknown task UUIDs. |
| **Checkpoint Resume** | `garuda.resume` | Restores saved checkpoint state for the receiving agent and marks the checkpoint consumed. |
| **Decision Lineage** | `garuda.propose_decision` & Surface `§` | Makes MADR architectural and formula decisions and their supersession chain inspectable. |

### Strict Separation of Evaluation Tiers (Fixtures vs. Simulations vs. Live AST/LLM)

Garuda strictly separates **compiler benchmark fixtures**, **synthetic fault-injection simulations**, and **live production repository / LLM agent results** so benchmark or demo data never pollutes live workspace queries:

| Evaluation Tier | Scope & Activation Flag | Default in Production Workspaces | What It Measures & Verifies |
| :--- | :--- | :--- | :--- |
| **Tier 1 — Compiler Fixture Results** (`{counts['compiler_fixture_benchmarks']}/{counts['compiler_fixture_benchmarks']}`) | `make benchmark-strict` or `include_fixtures=true` | **EXCLUDED BY DEFAULT** (`ExcludeFixturesSQL` filters out `test/fixtures/`, `testdata/`, `garuda-bench/`, `conftest`) | Static `go/types`, `go.work`, Python AST, and TypeScript AST extraction precision/recall (`100%`) and checkout-path-invariant cache fingerprints (`TestDoubleRunDeterminism`). |
| **Tier 2 — Synthetic Grounding & Contradiction Simulations** (`{counts['synthetic_runtime_contradiction_simulations']}/{counts['synthetic_runtime_contradiction_simulations']}`) | `scripts/runtime_contradiction_test.py`, `--demo-fixtures`, `simulate_violation`, `simulate_drift` | **OPT-IN ONLY** (Empty workspaces return `status: "INSUFFICIENT_DATA"` when `include_fixtures=false`) | Quarantine of injected unapproved OTel runtime spans (`10/10`), Talon self-repair prompt synthesis under simulated `BLOCK`/drift, and reference case studies (`Loss Ratio` `0.0–1.0` weights & `48.00 Cr vs 4.80 Cr`). |
| **Tier 3 — Live AST, RFC 6962 Ledger & Live LLM/MCP Sessions** (`{counts['mcp_protocol_verification_checks']}/{counts['mcp_protocol_verification_checks']}` & `{counts['mcp_behavioral_unit_tests']}/{counts['mcp_tools']}`) | Default live mode (`scope: "production_only"`, `GARUDA_TALON_ENFORCE=true`) | **ACTIVE BY DEFAULT** | Live repository symbol & blast-radius resolution, cryptographic tenant/workspace isolation, `GARUDA_MCP_READ_ONLY=true` mutation blocking, `GARUDA_MCP_APPROVER` human co-sign verification, and SHA-256 `receipt_hash` validation. |

---

## Tour the Futuristic 10-Surface Dashboard (`/dashboard` PWA)

Garuda's web interface runs 100% locally at `http://localhost:8080/dashboard` with progressive disclosure (**at most 4 cards on first view**), smooth `cubic-bezier(0.16, 1, 0.3, 1)` animations, and a 1-click toggle between **Warm Alabaster Light Mode** and **Obsidian Graphite Dark Mode** (`☀ / ☾`):

### `01` — Overview: Calm Command Center (`4` Primary Cards + `2` Expandable Drawers)
Displays **1 Hero System Integrity Card** (`ALLOW` / `BLOCK`) + **3 clickable core metric cards** (*Indexed Topology*, *Epistemic & Policy Gate*, *RFC 6962 Provenance Ledger*), with collapsible `<details>` drawers for *Needs Attention & Documentation Drift* and *Architectural Hubs & Indexed Repositories*.

<p align="center">
  <img src="{shots['overview']}" width="100%" alt="Surface 01 — Garuda Workspace Overview & System Command Center">
</p>

### `02` — Explore Index: Instant Semantic Symbol Search (`⌘K`)
Search repositories, packages, functions, methods, structs, and interfaces with instant signature preview and 1-hop neighborhood actions.

<p align="center">
  <img src="{shots['search']}" width="100%" alt="Surface 02 — Garuda Semantic Codebase Explorer">
</p>

### `03` — Architecture Map: Full-Viewport Neighborhood & AST Graph
Explores 1-hop and 2-hop symbol neighborhoods (`Callers → Symbol → Dependencies`) on an interactive D3 canvas (`74vh+`) with a collapsible depth/edge filter bar and synchronized entity inspector.

<p align="center">
  <img src="{shots['architecture']}" width="100%" alt="Surface 03 — Garuda Architecture Map & Neighborhood Graph">
</p>

### `04` — Docs & Rules Graph: 4-Tab Progressive Lineage & Base Rulebook Studio
Switches cleanly between **01 · Interactive 4-Tier Graph** (`Document → Rule → Policy → Code`), **02 · Base Rulebook Studio** (`.garuda/APPROVED_BASE_RULEBOOK.md`), **03 · Ingested Documents**, and **04 · Extracted Rules Matrix**.

<p align="center">
  <img src="{shots['docs_graph']}" width="100%" alt="Surface 04 — Garuda Docs, Rules & Repository Topology">
</p>

### `05` — Talon & Risk Hub: Active AI Agent Guardrails (`3` Sub-Tabs)
Organizes AI agent governance into **01 · Live Enforcement Gate & Hooks** (6 IDE/Git hooks + self-repair simulator), **02 · Risk Hub Leaderboard & Surface Criticality**, and **03 · Deterministic Build Matrix & Receipt**.

<p align="center">
  <img src="{shots['talon']}" width="100%" alt="Surface 05 — Garuda Talon & Workspace Clusters">
</p>

### `06` — Change Review Workbench (`3`-Pane Pre-Merge Simulator)
Inspect any symbol or git diff across three balanced cards: **01 · Change Context**, **02 · Blast-Radius Impact Map**, and **03 · 3-Plane Verification & Decision Record**.

<p align="center">
  <img src="{shots['review']}" width="100%" alt="Surface 06 — Garuda Cross-Repo & Change Review Impact Graph">
</p>

### `07` — Evidence & Ledger, `08` — Agents, `09` — Data & Logic Audit, `§` — Decision Lineage
- **`07 — Evidence & Ledger`**: Consolidated RFC 6962 Merkle epoch header + append-only audit table with an expandable Decision Record & Raw Telemetry drawer.
- **`08 — Agents & Sessions`**: Active MCP client sessions and multi-agent task handoff timeline.
- **`09 — Data & Logic Audit`**: Tabbed **Pillar 1 · Multi-Team Formula Weights (`0.0–1.0`)** and **Pillar 2 · Cross-Team Data & Logic Discrepancies (`48.00 Cr vs 4.80 Cr`)** backed by `data_logic_audit_ledger`.
- **`§ — Decision Lineage`**: MADR-compatible Architectural Decision Records and supersession chains.

<p align="center">
  <img src="{shots['trust']}" width="100%" alt="Surface 07 — Garuda Evidence & Cryptographic Trust Ledger">
</p>

<p align="center">
  <img src="{shots['contradictions']}" width="100%" alt="Surface 09 — Garuda Multi-Repo Contradictions & Data/Logic Audit">
</p>

### VS Code & Cursor IDE Extension (`vscode-extension/`)

| IDE Graph View (`{shots['ide_graph']}`) | IDE Contradictions Panel (`{shots['ide_contradictions']}`) |
| :---: | :---: |
| <img src="{shots['ide_graph']}" width="100%" alt="IDE Graph View"> | <img src="{shots['ide_contradictions']}" width="100%" alt="IDE Contradictions Panel"> |
{extra_gallery_md}
---

## Everyday CLI & MCP Workflows (`{counts['mcp_tools']}` Registered Tools)

| Workflow | CLI Command (`garuda ...`) | MCP Tool (`garuda-mcp`) |
| :--- | :--- | :--- |
| **1-Call Pre-Edit Context & Blast Radius** | `garuda talon brief <Symbol>` | `garuda.preflight_bundle` / `garuda.talon_pre_edit` |
| **Post-Edit Diff Check & Self-Repair Prompt** | `garuda talon guard --diff HEAD --fail-on-block` | `garuda.talon_verify_edit` |
| **Verify Deterministic Build & Graph Receipt** | `garuda check --workspace default` | `garuda.verify_receipt` / `garuda.snapshot` |
| **Diff Workspace Snapshots & Symbol Impact** | `garuda analyze --diff before.json after.json` | `garuda.diff_snapshots` / `garuda.diff_impact` |
| **Verify Cross-OS Build Matrix (`4` Profiles)** | `garuda check --matrix --workspace default` | `garuda.build_matrix_check` |
| **Compare Competing Team Formulas (`0.0–1.0`)** | `garuda audit logic "Loss Ratio" --team ops` | `garuda.disambiguate_logic` / `garuda.decide_logic_variant` |
| **Audit Cross-Team Data & Unit-Scaling (`48 Cr` vs `4.8 Cr`)** | `garuda audit data-logic --metric "Claims Paid"` | `garuda.audit_data_discrepancy` / `garuda.reconcile_data_audit` |
| **Inspect 3-Tier Surface Criticality** | `garuda audit criticality` | `garuda.criticality_tiers` |
| **Auto-Generate & Seal Base Rulebook** | `garuda rulebook generate && garuda rulebook approve` | `garuda.synthesize_rulebook` / `garuda.approve_base_rulebook` |
| **Anti-Hallucination Claim & Coverage Check** | `garuda docs verify` | `garuda.verify_claim` / `garuda.explain_claim` / `garuda.symbol_test_coverage` |
| **Provision Isolated Sandbox & Export Audit Trail** | `garuda init --workspace audit-sandbox` | `garuda.create_workspace` / `garuda.audit_export` |

---

## Five Core Governance Capabilities & End-to-End Scenario

> **AI can write the code. Who governs the change?**
> An AI agent modifies a payment service. Garuda discovers that the change bypasses the service layer, violates an architectural decision (`ADR-0001`), and introduces a financial unit-scaling precision risk (`48.00 Cr` vs `4.80 Cr`). **Garuda Talon** evaluates the change, identifies the violations, provides an actionable self-repair direction, and prevents the unacceptable change from passing the configured gate. The resulting governance decision is anchored with verifiable RFC 6962 Merkle evidence.
> **Understand the system. Govern the change. Preserve the proof.**

| Pillar | What It Establishes |
| :--- | :--- |
| **1. Software Intelligence** | Understands code structure, dependencies, relationships, and blast-radius impact across Go (`go/types`), Python, and TypeScript. |
| **2. Architecture Intelligence** | Discovers, synthesizes, and maintains architectural rules, auto-generated Base Rulebooks (`.garuda/APPROVED_BASE_RULEBOOK.md`), and ADRs. |
| **3. Business Logic Governance** | Evaluates competing team formulas (`0.0–1.0` weights), data relationships, and unit-scaling discrepancies (`48.00 Cr` vs `4.80 Cr`). |
| **4. Agent Change Control** | Guides agents (`garuda.talon_pre_edit`), evaluates modifications (`garuda.talon_verify_edit`), and enforces fail-closed policies. |
| **5. Verifiable Governance** | Records human-in-the-loop decisions and provides cryptographic RFC 6962 SHA-256 Merkle inclusion proofs. |

---

## 🔌 MCP Server & AI Coding Agent Integrations

Garuda ships a hardened `{counts['mcp_tools']}`-tool MCP server (`garuda-mcp`) and native plugin manifests:

- **Claude Code & Claude Desktop**: Auto-configured by `garuda init` (`.claude/settings.json` + [`CLAUDE.md`](CLAUDE.md) + [`.claude-plugin/plugin.json`](.claude-plugin/plugin.json)).
- **Cursor IDE**: Auto-configured by `garuda init` (`.cursor/rules/garuda-talon.mdc` + `.cursor/mcp.json` + [`.cursor-plugin/plugin.json`](.cursor-plugin/plugin.json)).
- **OpenAI Codex Plugin**: Native manifest at [`.codex-plugin/plugin.json`](.codex-plugin/plugin.json) and [`AGENTS.md`](AGENTS.md).
- **MCP Registries**: Ready-to-publish manifests for Official MCP Registry ([`mcp-registry.json`](mcp-registry.json)), Smithery ([`smithery.yaml`](smithery.yaml)), and Glama ([`glama.json`](glama.json)).

## 📚 Architecture Decision Records (ADRs), Playbooks & Deep Dives

- **Quickstart & Monorepo Guide**: [`QUICKSTART.md`](QUICKSTART.md)
- **Technical Deep-Dive Blog Post**: [`docs/BLOG_AST_MERKLE_AGENT_GOVERNANCE.md`](docs/BLOG_AST_MERKLE_AGENT_GOVERNANCE.md)
- **Architectural Decision Records ([`docs/adr/`](docs/adr/))**:
  - [`ADR-0001: Payment Idempotency & Hexagonal Store Isolation`](docs/adr/0001-payment-idempotency.md)
  - [`ADR-0002: RFC 6962 Canonical SHA-256 Merkle Provenance Layer`](docs/adr/0002-rfc6962-merkle-ledger.md)
  - [`ADR-0003: Three-State Epistemic Honesty (SUPPORTED, UNVERIFIED, CONTRADICTED)`](docs/adr/0003-epistemic-honesty.md)
- **Operational Playbooks ([`docs/playbooks/`](docs/playbooks/README.md))**: [`docs/playbooks/README.md`](docs/playbooks/README.md)
- **Database Schema & Migrations ([`migrations/`](migrations/))**: [`migrations/`](migrations/) (including [`095_data_logic_audit_ledger_and_hygiene.sql`](migrations/095_data_logic_audit_ledger_and_hygiene.sql), [`096_policy_evaluations_decision_check.sql`](migrations/096_policy_evaluations_decision_check.sql), [`097_multi_tenant_rbac_isolation.sql`](migrations/097_multi_tenant_rbac_isolation.sql), and [`098_strict_workspace_and_tenant_isolation.sql`](migrations/098_strict_workspace_and_tenant_isolation.sql))
- **10/10 Multi-Tenant Isolation & Evidence Regression Suite**: [`internal/api/tenant_isolation_test.go`](internal/api/tenant_isolation_test.go) (`go test -v ./internal/api -run "Test0|Test10"`)

---

## GitHub Packages & GitHub Marketplace Integration

Garuda `{ver}` ships with first-class support for **GitHub Packages (`ghcr.io`)** and the **GitHub Marketplace (`action.yml`)**. See **[`docs/GITHUB_MARKETPLACE_STRATEGY.md`](docs/GITHUB_MARKETPLACE_STRATEGY.md)** for the full launch playbook:

```yaml
# Use Garuda Talon directly from GitHub Marketplace in any workflow:
- name: Garuda Talon Architecture & Data-Logic Gate
  uses: myshra777-ai/garuda@{ver}
  with:
    workspace: default
    policy-dir: policies
    fail-on-block: "true"
    matrix: "true"
```

```bash
# Pull the multi-arch OCI container from GitHub Packages (GHCR):
docker pull ghcr.io/myshra777-ai/garuda:{ver}
```

---

## Documentation Suite by Role

| Document | Audience | Description |
| :--- | :--- | :--- |
| **[`capabilities.json`](capabilities.json)** | **All Tools & Evaluators** | Single canonical version & capability manifest (`{ver}`, `{counts['mcp_tools']}` MCP tools, `{counts['dashboard_surfaces']}` surfaces, `{counts['sql_migrations']}` migrations). |
| **[`docs/PERSONA_ONBOARDING_GUIDE.md`](docs/PERSONA_ONBOARDING_GUIDE.md)** | **All Users (Start Here!)** | 5-minute onboarding paths for Engineers, Architects, Data/Audit Leads, PMs, DevOps, and AI Agents. |
| **[`INSTALL.md`](INSTALL.md)** | **Everyone** | 1-command zero-dependency setup across Linux, macOS, Windows PowerShell, Chromebook, and Android. |
| **[`docs/WALKTHROUGH.md`](docs/WALKTHROUGH.md)** | **Product & Engineering Teams** | Visual walkthrough of the redesigned `{counts['dashboard_surfaces']}`-surface progressive-disclosure `/dashboard` PWA. |
| **[`docs/DATA_AND_LOGIC_AUDIT_GUIDE.md`](docs/DATA_AND_LOGIC_AUDIT_GUIDE.md)** | **Engineering & Data Leads** | `0.0–1.0` formula weighting, active contradiction quarantine, and `data_logic_audit_ledger`. |
| **[`docs/API.md`](docs/API.md)** | **Tooling & Integration Engineers** | Complete CLI flag reference, all **`{counts['mcp_tools']}` MCP tool schemas**, `GARUDA_MCP_READ_ONLY` / `GARUDA_MCP_APPROVER` controls, and HTTP REST API. |
| **[`docs/CAPABILITIES.md`](docs/CAPABILITIES.md)** | **Staff Architects & Evaluators** | 8-plane capability matrix + 3-tier separation of compiler fixtures, synthetic simulations, and live AST/LLM results. |
| **[`docs/GITHUB_MARKETPLACE_STRATEGY.md`](docs/GITHUB_MARKETPLACE_STRATEGY.md)** | **Maintainers & DevOps** | Step-by-step launch guide for GitHub Packages (`ghcr.io`) and GitHub Marketplace (`action.yml`). |
| **[`EVIDENCE.md`](EVIDENCE.md)** | **Technical Due Diligence** | Reproducible verification suite outputs (`{counts['compiler_fixture_benchmarks']}/{counts['compiler_fixture_benchmarks']}` fixtures, `{counts['synthetic_runtime_contradiction_simulations']}/{counts['synthetic_runtime_contradiction_simulations']}` OTel simulations, `{counts['mcp_protocol_verification_checks']}/{counts['mcp_protocol_verification_checks']}` MCP checks, `{counts['mcp_behavioral_unit_tests']}/{counts['mcp_tools']}` behavioral tests). |
| **[`RELEASE_NOTES.md`](RELEASE_NOTES.md)** | **Everyone** | Official `{ver}` release notes and multi-OS download matrix. |

---

<p align="center">
  <strong>Let AI agents write fast. Let Garuda keep your architecture, formulas, and data whole.</strong><br>
  Released under the <a href="LICENSE">Apache 2.0 License</a>.
</p>
"""


def sync_all(repo_root: str, check_only: bool = False) -> int:
    manifest_path = os.path.join(repo_root, "capabilities.json")
    if not os.path.exists(manifest_path):
        print(f"❌ Canonical manifest missing at {manifest_path}!")
        print("   Root cause: capabilities.json is ignored by '*.json' in .gitignore or was not committed to git.")
        print("   Fix: Add '!capabilities.json' to .gitignore, run 'git add -f capabilities.json .gitignore', and commit.")
        return 1
    manifest = json.load(open(manifest_path))
    ver = manifest["version"]
    semver = manifest["semver"]
    counts = manifest["counts"]

    tools = extract_mcp_tools(repo_root)
    if len(tools) != counts["mcp_tools"]:
        print(f"❌ Tool count mismatch: capabilities.json has {counts['mcp_tools']}, cmd/garuda-mcp/main.go has {len(tools)}")
        return 1

    assets_info = discover_assets(repo_root)

    if check_only:
        readme = open(os.path.join(repo_root, "README.md")).read()
        if ver not in readme or f"{counts['mcp_tools']}" not in readme:
            print("❌ README.md is out of sync with capabilities.json")
            return 1
        vtxt_path = os.path.join(repo_root, "version.txt")
        if not os.path.exists(vtxt_path) or open(vtxt_path).read().strip() != ver:
            print(f"❌ version.txt missing or does not match {ver}")
            return 1
        dockerfile = open(os.path.join(repo_root, "Dockerfile")).read()
        if f"ARG VERSION={ver}" not in dockerfile or "garuda-mcp" not in dockerfile or "garuda-talon" not in dockerfile:
            print("❌ Dockerfile is out of sync (must default to canonical version and build all 4 binaries)")
            return 1
        matrix_sh = open(os.path.join(repo_root, "scripts/build_release_matrix.sh")).read()
        if ver not in matrix_sh or 'OUT_DIR="${1:-dist/releases}"' not in matrix_sh:
            print("❌ scripts/build_release_matrix.sh is out of sync (must default to canonical version and dist/releases)")
            return 1
        action_yml = open(os.path.join(repo_root, "action.yml")).read()
        if "GITHUB_ACTION_PATH" not in action_yml or "--diff" not in action_yml or '|| echo "ALLOW"' in action_yml:
            print("❌ action.yml failed fail-closed & GITHUB_ACTION_PATH audit")
            return 1
        pos_md = open(os.path.join(repo_root, "docs/POSITIONING.md")).read()
        if ver not in pos_md or "sixteen tools" in pos_md.lower():
            print("❌ docs/POSITIONING.md is out of sync with canonical manifest")
            return 1
        print(f"✅ Canonical manifest & release-blocker check passed: {ver} | {len(tools)} MCP tools | {counts['dashboard_surfaces']} surfaces | {counts['sql_migrations']} migrations | action.yml + Dockerfile + dist/releases verified")
        return 0

    # 1. Write version.txt
    open(os.path.join(repo_root, "version.txt"), "w").write(ver + "\n")

    # 2. Write README.md
    readme_md = generate_readme(manifest, assets_info, tools)
    open(os.path.join(repo_root, "README.md"), "w").write(readme_md)

    # 3. Update docs/WALKTHROUGH.md (fix escaped HTML tags, old Screenshot From links, and 35/23 tool counts)
    wt_path = os.path.join(repo_root, "docs/WALKTHROUGH.md")
    if os.path.exists(wt_path):
        wt = open(wt_path).read()
        wt = wt.replace("&lt;", "<").replace("&gt;", ">").replace("&quot;", '"').replace("&amp;", "&")
        wt = wt.replace("35 active MCP tools", f"{counts['mcp_tools']} active MCP tools")
        wt = wt.replace("23-tool MCP server", f"{counts['mcp_tools']}-tool MCP server")
        shots = assets_info["surface_shots"]
        wt = wt.replace("../assets/Screenshot%20From%202026-10-25%2002-30-49.png", "../" + shots["overview"])
        wt = wt.replace("../assets/Screenshot%20From%202026-10-25%2002-31-11.png", "../" + shots["search"])
        wt = wt.replace("../assets/Screenshot%20From%202026-10-25%2002-31-24.png", "../" + shots["architecture"])
        wt = wt.replace("../assets/Screenshot%20From%202026-10-25%2002-31-40.png", "../" + shots["docs_graph"])
        wt = wt.replace("../assets/Screenshot%20From%202026-10-25%2002-31-56.png", "../" + shots["talon"])
        wt = wt.replace("../assets/Screenshot%20From%202026-10-25%2002-32-11.png", "../" + shots["review"])
        wt = wt.replace("../assets/Screenshot%20From%202026-10-25%2002-32-25.png", "../" + shots["trust"])
        if not os.path.exists(os.path.join(repo_root, "assets/garuda_dashboard_demo_video.mp4")):
            idx0 = wt.find("## 0. Watch the 20-Second")
            idx1 = wt.find("## 1. One-Command")
            if idx0 != -1 and idx1 != -1:
                wt = wt[:idx0] + wt[idx1:]
        open(wt_path, "w").write(wt)

    # 4. Update RELEASE_NOTES.md (eliminate 35/26 tool count and 9 surface drift, add 3-tier separation)
    rn_path = os.path.join(repo_root, "RELEASE_NOTES.md")
    if os.path.exists(rn_path):
        rn = open(rn_path).read()
        rn = rn.replace("35 deterministic MCP tools", f"{counts['mcp_tools']} deterministic MCP tools ({counts['mcp_readonly_tools']} read-only, {counts['mcp_mutating_tools']} mutating)")
        rn = rn.replace("from 26 to **35 zero-stub tools**", f"to **{counts['mcp_tools']} zero-stub tools** (`{counts['mcp_behavioral_unit_tests']}/{counts['mcp_tools']}` positive & negative behavioral tests + `{counts['mcp_protocol_verification_checks']}/{counts['mcp_protocol_verification_checks']}` protocol checks)")
        rn = rn.replace("to **9 interactive surfaces**", f"to **{counts['dashboard_surfaces']} progressive-disclosure surfaces**")
        open(rn_path, "w").write(rn)

    # 5. Update EVIDENCE.md (fix 'All 6 surfaces' -> 'All 10 progressive-disclosure surfaces' and prepend 3-Tier Separation section)
    ev_path = os.path.join(repo_root, "EVIDENCE.md")
    if os.path.exists(ev_path):
        ev = open(ev_path).read()
        ev = ev.replace("All 6 surfaces (`Briefing`, `Architecture Map`, `Change Review Workbench`, `Policies`, `Contradictions & Evidence`, `Governance & Audit Ledger`)", f"All {counts['dashboard_surfaces']} progressive-disclosure surfaces (`Overview`, `Explore Index`, `Architecture Map`, `Docs & Rules Graph`, `Talon & Risk Hub`, `Change Review`, `Evidence & Ledger`, `Agents & Sessions`, `Data & Logic Audit`, `Decision Lineage`)")
        tier_section = f"""## Three-Tier Evaluation Separation (`P3` Canonical Standard)

To prevent benchmark fixtures or synthetic demos from being conflated with live repository measurements, Garuda partitions all verification evidence into three strict tiers:

1. **Tier 1 — Deterministic Compiler Fixture Results (`{counts['compiler_fixture_benchmarks']}/{counts['compiler_fixture_benchmarks']}` Strict Suite)**:
   - **Activation**: `make benchmark-strict` or `include_fixtures=true`.
   - **Default Production Behavior**: **Excluded by default** (`ExcludeFixturesSQL` filters out `test/fixtures/`, `testdata/`, `garuda-bench/`, and `conftest`).
   - **What It Proves**: 100% extraction precision and recall on controlled Go (`go/types`, `go.work`, `vendor/`), Python AST, and TypeScript AST ground-truth fixtures, plus checkout-path-invariant cache fingerprints (`TestDoubleRunDeterminism`).
2. **Tier 2 — Synthetic Grounding & Contradiction Simulations (`{counts['synthetic_runtime_contradiction_simulations']}/{counts['synthetic_runtime_contradiction_simulations']}` Fault Injections)**:
   - **Activation**: `python3 scripts/runtime_contradiction_test.py`, `--demo-fixtures`, `simulate_violation`, `simulate_drift`.
   - **Default Production Behavior**: **Opt-in only**. Calling `garuda.disambiguate_logic` or `garuda.audit_data_discrepancy` on an empty workspace with `include_fixtures=false` returns `status: "INSUFFICIENT_DATA"` with zero synthetic leakage.
   - **What It Proves**: Quarantine of injected unapproved OTel service edges (`10/10`), deterministic self-repair prompt generation under simulated `BLOCK`/drift, and human-in-the-loop disambiguation of competing formulas (`0.0–1.0` weights) and 10x unit-scaling bugs (`48.00 Cr vs 4.80 Cr`).
3. **Tier 3 — Live Repository AST, RFC 6962 Ledger & Live LLM/MCP Agent Sessions (`{counts['mcp_protocol_verification_checks']}/{counts['mcp_protocol_verification_checks']}` Protocol Checks + `{counts['mcp_behavioral_unit_tests']}/{counts['mcp_tools']}` Behavioral Tool Tests)**:
   - **Activation**: Default production mode (`scope: "production_only"`, `GARUDA_TALON_ENFORCE=true`).
   - **What It Proves**: Real-time AST symbol & blast-radius resolution on live repositories, cryptographic tenant/workspace isolation (`TestNewTools_TenantIsolation_ReadOnly_HumanApproval_ReceiptHandling`), fail-closed read-only enforcement (`GARUDA_MCP_READ_ONLY=true`), attested human co-signing (`GARUDA_MCP_APPROVER`), and tamper-evident SHA-256 `receipt_hash` verification.

---

## Evidence at a Glance"""
        if "## Three-Tier Evaluation Separation" not in ev:
            ev = ev.replace("## Evidence at a Glance", tier_section)
        open(ev_path, "w").write(ev)

    # 6. Update docs/API.md and create docs/API_REFERENCE.md with the full 54-tool table generated from cmd/garuda-mcp/main.go
    api_path = os.path.join(repo_root, "docs/API.md")
    if os.path.exists(api_path):
        api_md = open(api_path).read()
        tool_rows = []
        for idx, (tname, tdesc) in enumerate(tools, 1):
            mode = "Mutating (`GARUDA_MCP_READ_ONLY` gated)" if tname in {
                "garuda.propose_decision", "garuda.decide_logic_variant", "garuda.reconcile_data_audit",
                "garuda.reconcile_financial_audit", "garuda.approve_base_rulebook", "garuda.synthesize_rulebook",
                "garuda.docs_ingest", "garuda.create_task", "garuda.handoff", "garuda.resume",
                "garuda.reset_audit_state", "garuda.create_workspace", "garuda.talon_verify_edit",
            } else "Read-Only"
            tool_rows.append(f"| `{idx:02d}` | `{tname}` | {mode} | {tdesc} |")
        manifest_table = (
            f"\n\n---\n\n## Canonical {counts['mcp_tools']}-Tool MCP Registry (Auto-Synced from `capabilities.json` & `cmd/garuda-mcp/main.go`)\n\n"
            "| # | Tool Name | Mode & Safety Gate | Description |\n"
            "| :--- | :--- | :--- | :--- |\n"
            + "\n".join(tool_rows)
            + "\n"
        )
        if "## Canonical 54-Tool MCP Registry" not in api_md:
            api_md = api_md.rstrip() + manifest_table
        open(api_path, "w").write(api_md)
        open(os.path.join(repo_root, "docs/API_REFERENCE.md"), "w").write(api_md)

    # 7. Sync docs/source/capabilities.yaml version and tool counts
    cap_yaml_path = os.path.join(repo_root, "docs/source/capabilities.yaml")
    if os.path.exists(cap_yaml_path):
        cy = open(cap_yaml_path).read()
        cy = re.sub(r'version:\s*"[^"]+"', f'version: "{semver}"', cy)
        cy = re.sub(r'last_updated:\s*"[^"]+"', f'last_updated: "{manifest["release_date"]}"', cy)
        cy = cy.replace("23-Tool MCP Surface", f"{counts['mcp_tools']}-Tool MCP Surface")
        cy = cy.replace("23 MCP tools", f"{counts['mcp_tools']} MCP tools")
        cy = cy.replace("30/30 Checks in scripts/mcp_verify.py", f"{counts['mcp_protocol_verification_checks']}/{counts['mcp_protocol_verification_checks']} Checks in scripts/mcp_verify.py")
        open(cap_yaml_path, "w").write(cy)

    print(f"✅ Synchronized all documentation, assets, and capability counts from capabilities.json ({ver}, {len(tools)} MCP tools).")
    return 0


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Sync docs and assets from capabilities.json")
    parser.add_argument("--repo", default=".", help="Repository root")
    parser.add_argument("--check", action="store_true", help="Check consistency without writing")
    args = parser.parse_args()
    sys.exit(sync_all(os.path.abspath(args.repo), check_only=args.check))
