# Garuda System Walkthrough (v1.6 Universal)

**What each interface looks like, how it helps you, and how to launch it in one command across Windows, macOS, Linux, Chromebook, and Android.**

This walkthrough covers all seven core surfaces of **Garuda v1.6**:
1. **1-Command Zero-Dependency Universal Setup (`install.ps1` / `setup.ps1` & `install.sh` / `setup.sh`)**
2. **Self-Diagnosis & Auto-Repair CLI (`garuda doctor --fix`)**
3. **Garuda Talon Active Agent Execution Harness (`garuda talon`)**
4. **Diff-Scoped Unified Change-Control CLI (`garuda check --diff`)**
5. **6-Surface Software Intelligence Instrument & PWA (`/dashboard`)**
6. **23 Active MCP Tools + Prompts & Resources (`bin/garuda-mcp`)**
7. **VS Code / Cursor Extension (`vscode-extension/`)**

---

## 1. 1-Command Zero-Dependency Universal Setup

Garuda v1.6 eliminates manual toolchain downloads, manual SQL migration loops, manual JWT key generation, and passive MCP configuration. Whether you are on a fresh Windows laptop, a MacBook without Xcode CLI tools, a Linux workstation, a Chromebook, or Android Termux, Garuda provisions its own portable runtime in user-space (`~/.garuda`) and boots in one command:

**Windows 10 / 11 (PowerShell):**
```powershell
# Remote 1-command install & launch
Set-ExecutionPolicy Bypass -Scope Process -Force; irm https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.ps1 | iex

# Or inside a local repository
powershell -ExecutionPolicy Bypass -File .\setup.ps1
```

**macOS, Linux, Chromebook (Crostini) & Android (Termux):**
```bash
# Remote 1-command install & launch
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash

# Or inside a local repository
./setup.sh
```

### What happens automatically
1. **Zero-Admin Portable Toolchain & Database**: If Go or PostgreSQL are not installed on your device, Garuda downloads Portable Go and Portable PostgreSQL into `~/.garuda/`, initializes a local database cluster on port `5433`, and builds all four binaries (`garuda`, `garuda-api`, `garuda-mcp`, `garuda-talon`) using `CGO_ENABLED=0` Pure-Go mode when no C compiler (`gcc`/`clang`) is present.
2. **Dynamic Identity & Persistent `.env`**: Reads your identity from `git config` (or defaults), runs all SQL migrations idempotently, and persists `DATABASE_URL`, `GARUDA_WORKSPACE`, `GARUDA_TALON_ENFORCE=true`, `PORT=8080`, and Ed25519 JWT keys into `.env`.
3. **Garuda Talon Active Agent Harness (`garuda talon init`)**: Compiles live workspace ADRs (`document_claims`), declarative YAML policies (`policies/*.yaml`), and top high-blast-radius hubs into `.garuda/TALON_CONTRACT.md`, `.cursor/rules/garuda-talon.mdc`, `.cursor/mcp.json`, `CLAUDE.md`, `.claude/settings.json` (`PostToolUse` hook), `AGENTS.md`, `.windsurfrules`, and `.git/hooks/pre-commit`.

---

## 2. Self-Diagnosis & Auto-Repair (`garuda doctor`)

If a background PostgreSQL database stops after a reboot, a binary is missing, or an IDE hook needs re-linking, run:

```bash
./bin/garuda doctor --fix
```

```text
╔═══════════════════════════════════════════════════════════════════╗
║               🦅 GARUDA SYSTEM DOCTOR & DIAGNOSTICS               ║
╚═══════════════════════════════════════════════════════════════════╝
  Platform:  linux/amd64  |  Workspace: default

  ✓ [PASS] Platform & Runtime Environment
      OS=linux ARCH=amd64
  ✓ [PASS] Environment File (.env)
      Loaded .env (GARUDA_WORKSPACE=default, GARUDA_TALON_ENFORCE=true)
  ✓ [PASS] PostgreSQL Database & Migrations
      PostgreSQL reachable and all schema migrations verified
  ✓ [PASS] Garuda Binaries (CLI, API, MCP, Talon)
      All 4 binaries present in /home/user/garuda/bin
  ✓ [PASS] Garuda Talon Active Harness & MCP Hooks
      .garuda/TALON_CONTRACT.md, .cursor/mcp.json, .claude/settings.json, CLAUDE.md, and AGENTS.md active
  ✓ [PASS] Garuda API & Software Intelligence Dashboard
      Online at http://localhost:8080/dashboard (PWA enabled for Chromium/Android/Desktop)

🎉 All core systems healthy! Open http://localhost:8080/dashboard in your browser.
```

---

## 3. Garuda Talon Active Agent Execution Harness (`garuda talon`)

Standard MCP servers are passive—if an AI agent forgets to query a tool, it writes code blind. **Garuda Talon** enforces architectural boundaries before, during, and after every agent edit.

### Pre-Edit Briefing (`garuda talon brief`)
Before modifying a high-impact struct, interface, or package, inspect its exact blast radius, applicable ADR `MUST`/`MUST_NOT` rules, and active YAML policies:

```bash
./bin/garuda talon brief SQLModel --workspace go-validation-10
```

```text
🦅 GARUDA TALON PRE-EDIT BRIEFING (workspace: go-validation-10, target: SQLModel)
  • Transitive Blast Radius: 171 entities (CRITICAL=12, HIGH=48)
  • Symbol: sqlmodel.main.SQLModel (class, 67 inbound edges) at sqlmodel/main.py
  • ADR Rule [MUST_NOT]: CheckoutHandler MUST_NOT CALLS OrderRepository (ADR-0002)
  • Policy [BLOCK] Block Direct Handler-to-DB Layer Violations (arch-001): HTTP handlers must not call DB repositories directly
```

### Active Diff Guard (`garuda talon guard`)
Incrementally re-indexes dirty working-tree files in milliseconds, maps git diff hunks (`--staged`, `--diff <rev>`, or `--files`) to modified AST entities, evaluates net-new policy and ADR violations, and outputs a deterministic **Self-Repair Prompt** on `BLOCK`:

```bash
./bin/garuda talon guard --workspace go-validation-10 --fail-on-block
```

```text
╔═══════════════════════════════════════════════════════════════════╗
║                  GARUDA TALON ACTIVE GUARD                        ║
╚═══════════════════════════════════════════════════════════════════╝
  Workspace:   go-validation-10
  Diff Scope:  2 file(s), 3 touched symbol(s)
  Blast Radius:18 affected entities (CRITICAL=1, HIGH=4)
  Violations:  Net-New=0  |  Baseline=2  |  ADR Contradictions=0
  Merkle Seal: RFC 6962 block #14
  OUTCOME:     ALLOW
```

### Autonomous Supervisor Loop (`garuda talon run`)
Wrap any CLI coding agent (`claude`, `aider`, `codex`) in a closed self-repair loop that automatically feeds policy/ADR violation fixes back to the agent (up to `--max-retries 3`) until `ALLOW` is achieved:

```bash
./bin/garuda talon run --max-retries 3 -- claude "Refactor CheckoutHandler to use OrderService"
```

---

## 4. Diff-Scoped Unified Change-Control CLI (`garuda check`)

Run `garuda check` against a specific symbol (`--target`), a git revision range (`--diff HEAD`), staged index changes (`--staged`), or explicit files (`--files`). By default, diff mode isolates **net-new violations** introduced by your change from pre-existing baseline technical debt:

```bash
# Evaluate only the symbols modified in your current working-tree diff
./bin/garuda check --diff HEAD --project go-validation-10 --policy ./policies --save --fail-on-block

# Or evaluate a specific symbol across all 5 planes
./bin/garuda check --project go-validation-10 --target SQLModel --policy ./policies --save --fail-on-block
```

```text
╔═══════════════════════════════════════════════════════════════════╗
║               GARUDA CHANGE-CONTROL DECISION                      ║
╚═══════════════════════════════════════════════════════════════════╝
  Change:      SQLModel (workspace: go-validation-10)
  Affected:    171 entities (CRITICAL=12, HIGH=48)
  Evidence:    SUPPORTED=142  |  UNVERIFIED=5  |  CONTRADICTED=0
  Policies:    3 evaluated (Net-New Violations=0, Baseline=1)
  Merkle Seal: RFC 6962 block #15
───
  GOVERNANCE OUTCOME: WARN
```

Add `--json` inside CI/CD pipelines to emit a machine-readable JSON decision payload including `touched_files`, `touched_symbols`, `net_new_violations`, `baseline_violations`, `self_repair_instructions`, and `merkle_block_height`.

---

## 5. 6-Surface Software Intelligence Instrument & PWA (`/dashboard`)

Open `http://localhost:8080/dashboard` in any browser (or click **Install Garuda** in Chrome/Edge/Chromebook/Android to run it as a standalone Progressive Web App):

1. **Workspace Briefing (`/dashboard#briefing`)** — Clean status strip (`Workspace`, `Indexed Entities`, `Policy Posture`, `Active Contradictions`, `Merkle Ledger Head`) with an actionable **Needs Attention** queue and recent governance verdicts.
2. **Architecture Map (`/dashboard#map`)** — **Neighborhood-first 1-hop / 2-hop topology explorer** (`[Selected Caller] ──> [Focal Symbol] ──> [Direct Dependency / Policy Badge]`) with hop-depth (`1 hop`, `2 hops`) and edge-kind filters (`Calls`, `Imports`, `Implements`, `Claims`) plus a synchronized 4-tab Inspector Drawer (`Source`, `Relationships`, `Policy Results`, `Evidence`).
3. **Change Review Workbench (`/dashboard#workbench`)** — 3-pane pre-merge simulator: choose a target symbol and policy bundle on the left, inspect impacted callers and policy rules in the center, and generate a live `ALLOW` / `WARN` / `REVIEW` / `BLOCK` governance gate receipt on the right.
4. **Policies (`/dashboard#policies`)** — Interactive catalog of all 9 deterministic SQL/CTE policy predicates (`layer_violation`, `transitive_claim_exists`, `cycle_exists`, `blast_radius_exceeds`, `contradiction_exists`, `verification_missing`, `claim_exists`, `entity_exists`, `language_matches`) plus a live YAML Policy Sandbox.
5. **Contradictions & Evidence (`/dashboard#contradictions`)** — Sentry-style forensic queue for `CONTRADICTED` and `UNVERIFIED` claims, comparing static AST location against live OpenTelemetry runtime trace spans (`trace_id`, `span_id`, service pair) with a 1-click OTel Span Simulator.
6. **Governance & Audit Ledger (`/dashboard#ledger`)** — Append-only decision table with a bottom **Decision Record Inspector** (`What it did`, `What it looked at`, `What rule applied`, `Epistemic state`, `RFC 6962 Merkle Inclusion Proof`) and 1-click cryptographic verification.

Press `⌘K` / `Ctrl+K` from any surface to open the global command palette.

---

## 6. 23 Active MCP Tools + Prompts & Resources (`bin/garuda-mcp`)

Connected AI coding agents automatically receive the workspace's active architectural contract on their first tool call (`GARUDA_TALON_ENFORCE=true`) and can invoke 23 tools across 6 categories:
- **Active Talon Enforcement**: `garuda.talon_pre_edit`, `garuda.talon_verify_edit`
- **Orientation & Discovery**: `garuda.briefing`, `garuda.entities`, `garuda.find_entity`, `garuda.inspect`, `garuda.query`
- **Topology & Blast Radius**: `garuda.neighbors`, `garuda.blast_radius`, `garuda.implementers`, `garuda.subclasses`, `garuda.get_impact`, `garuda.get_lineage`
- **Governance & Policy Gates**: `garuda.policy.list`, `garuda.policy.evaluate`, `garuda.verify_policy_evaluation`, `garuda.governance.status`, `garuda.propose_decision`
- **Epistemic Truth & Drift**: `garuda.detect_contradictions`, `garuda.check_drift`, `garuda.query_claims`
- **Multi-Agent Continuity**: `garuda.handoff`, `garuda.resume`

---

## 7. VS Code / Cursor Extension (`vscode-extension/`)

Install the extension from `vscode-extension/` to view the **Merkle Decision Ledger**, **Active Contradictions**, **Inline Hover Blast Radius**, and **Interactive Graph Panel** directly inside VS Code or Cursor.
