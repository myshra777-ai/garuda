# Garuda Walkthrough (v1.5.0)

**What each interface looks like, what each view shows, and how to launch it in one command.**

This walkthrough covers the six core surfaces of **Garuda v1.5**:
1. **1-Command Smart Setup & Active Talon Wiring (`./setup.sh` / `garuda init`)**
2. **Garuda Talon Active Agent Execution Harness (`garuda talon`)**
3. **Diff-Scoped Unified Change-Control CLI (`garuda check --diff`)**
4. **6-Surface Software Intelligence Instrument (`/dashboard`)**
5. **23 Active MCP Tools + Prompts & Resources (`bin/garuda-mcp`)**
6. **VS Code / Cursor Extension (`vscode-extension/`)**

---

## 1. 1-Command Smart Setup (`./setup.sh` & `garuda init`)

Garuda v1.5 eliminates manual environment variable exports, manual SQL migration loops, manual JWT key generation, and passive MCP configuration. Everything is discovered dynamically, persisted to `.env`, and wired into your coding agents automatically:

```bash
./setup.sh
```

Or initialize any specific repository directly:

```bash
./bin/garuda init /path/to/your/repo --workspace my-system
```

### What happens automatically
1. **Dynamic Identity & Workspace Discovery**: Reads your developer email and name from `git config user.email` / `git config user.name` (or `$GARUDA_EMAIL`) and derives the active workspace from `$GARUDA_WORKSPACE` or the target repository name.
2. **PostgreSQL & Idempotent Migrations**: Starts or creates the `garuda-postgres` Docker container if PostgreSQL is not running, and runs `store.Migrate` idempotently across all SQL migrations (`001` through `084`).
3. **Persistent `.env` & Ed25519 JWT Keys**: Generates and writes `DATABASE_URL`, `CONTROL_DATABASE_URL`, `GARUDA_WORKSPACE`, `PORT`, `JWT_PRIVATE_KEY_HEX`, and `JWT_PUBLIC_KEY_HEX` into `.env`, which `garuda`, `garuda-api`, and `garuda-mcp` load automatically on every startup.
4. **Garuda Talon Active Agent Harness (`garuda talon init`)**: Compiles live workspace ADRs (`document_claims`), declarative YAML policies (`policies/*.yaml`), and top high-blast-radius hubs into:
   - `.garuda/TALON_CONTRACT.md` (compiled Markdown contract)
   - `.cursor/rules/garuda-talon.mdc` (`alwaysApply: true` Cursor agent rule)
   - `.cursor/mcp.json` (`GARUDA_TALON_ENFORCE=true` active MCP enforcement)
   - `CLAUDE.md` and `.claude/settings.json` (`PostToolUse` hook running `garuda talon guard --hook` after every `Edit|Write|MultiEdit`)
   - `AGENTS.md` and `.windsurfrules` (Codex, Aider, Windsurf contract blocks)
   - `.git/hooks/pre-commit` and `.git/hooks/commit-msg` (diff-scoped policy gate + RFC 6962 Merkle commit trailers)

---

## 2. Garuda Talon Active Agent Execution Harness (`garuda talon`)

A passive MCP server answers when the agent asks. **Garuda Talon adds configured pre-edit and post-edit integration points**: it can compile the workspace contract, provide targeted context, check changed files and return repair instructions. The exact interception boundary depends on the installed client hooks, command wrapper and CI configuration; Talon is not a universal agent sandbox.

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

## 3. Diff-Scoped Unified Change-Control CLI (`garuda check`)

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

## 4. 6-Surface Software Intelligence Instrument (`/dashboard`)

Garuda v1.5 ships a redesigned Software Intelligence Instrument at `http://localhost:8080/dashboard` inspired by Linear, Vercel, Sentry, and Neo4j Bloom.

**Launch it:**

```bash
./bin/garuda dev
```

Visit `http://localhost:8080/login` (or `http://localhost:8080/dashboard`) and press `⌘K` / `Ctrl+K` at any time to open the global command palette.

### The 6 Engineering Surfaces
1. **Workspace Briefing (`/dashboard#briefing`)**: Clean status strip and health KPIs (Indexed Entities, Policy Posture, Active Contradictions, Merkle Ledger Head) paired with an actionable Needs-Attention queue and recent governance verdicts.
2. **Architecture Map (`/dashboard#map`)**: **Neighborhood-first 1-hop / 2-hop topology explorer** (`[Selected Caller] ──> [Focal Symbol] ──> [Direct Dependency / Policy Badge]`) with interactive hop-depth (`1 hop`, `2 hops`) and edge-kind filters (`Calls`, `Imports`, `Implements`, `Claims`) plus a synchronized 4-tab Inspector Drawer (`Source`, `Relationships`, `Policy Results`, `Evidence`).
3. **Change Review Workbench (`/dashboard#workbench`)**: 3-pane pre-merge simulator—select a target symbol and policy bundle on the left, inspect impacted callers and policy rules in the center, and generate a live `ALLOW` / `WARN` / `REVIEW` / `BLOCK` governance gate receipt on the right.
4. **Policies (`/dashboard#policies`)**: Interactive catalog of all 9 deterministic SQL/CTE policy predicates alongside a live YAML Policy Sandbox to test rules before committing them to CI.
5. **Contradictions & Evidence (`/dashboard#contradictions`)**: Sentry-style forensic queue for `CONTRADICTED` and `UNVERIFIED` claims, comparing static AST location against live OpenTelemetry runtime trace spans (`trace_id`, `span_id`, service pair) plus an interactive OTel Span Simulator.
6. **Governance & Audit Ledger (`/dashboard#ledger`)**: Append-only decision table with a bottom **Decision Record Inspector** (`What it did`, `What it looked at`, `What rule applied`, `Epistemic state`, `RFC 6962 Merkle Inclusion Proof`) and 1-click cryptographic verification.

---

## 5. 23 Active MCP Tools + Prompts & Resources (`bin/garuda-mcp`)

Garuda speaks JSON-RPC 2.0 over stdio (`protocolVersion: "2025-06-18"`). With `GARUDA_TALON_ENFORCE=true` (the default unless explicitly disabled), MCP responses can include the active workspace contract and its ADR/policy guardrails. This response injection helps the agent see the rules, but is not by itself an unbypassable enforcement boundary.

### MCP Prompts & Resources
- **Prompt**: `garuda.architectural_contract` (returns compiled Markdown contract for the workspace)
- **Resources**: `garuda://workspace/contract` (`text/markdown`) and `garuda://workspace/adr-guardrails` (`application/json`)

### The 23 MCP Tools
- **Active Talon Enforcement (2)**: `garuda.talon_pre_edit`, `garuda.talon_verify_edit`
- **Orientation & Discovery (5)**: `garuda.briefing`, `garuda.entities`, `garuda.find_entity`, `garuda.inspect`, `garuda.query`
- **Topology & Blast Radius (6)**: `garuda.neighbors`, `garuda.blast_radius`, `garuda.implementers`, `garuda.subclasses`, `garuda.get_impact`, `garuda.get_lineage`
- **Governance & Cryptographic Verification (5)**: `garuda.policy.list`, `garuda.policy.evaluate`, `garuda.verify_policy_evaluation`, `garuda.governance.status`, `garuda.propose_decision`
- **Epistemic Truth & Drift (3)**: `garuda.detect_contradictions`, `garuda.check_drift`, `garuda.query_claims`
- **Multi-Agent Continuity (2)**: `garuda.handoff`, `garuda.resume`

### Verify MCP & Talon Compliance Dynamically
Run `scripts/mcp_verify.py` at any time:

```bash
python3 scripts/mcp_verify.py
```

Expected output:
```text
═══ 30/30 checks passed ═══
```

---

## 6. The VS Code / Cursor Extension (`vscode-extension/`)

Build and compile the extension from `vscode-extension/`:

```bash
cd vscode-extension
npm install
npm run compile
```

- **Inline Contradiction Diagnostics**: Highlights lines where runtime traces or policy evaluations flag `ARCH_DRIFT_001`.
- **Blast-Radius Hover**: Hover over any symbol to view its resolved upstream callers and downstream blast radius via `garuda.blast_radius`.
- **Merkle Ledger Status Bar**: Displays current Merkle block height and active contradiction count via `garuda.governance.status`.

---

## Where to Go Next
- **[README.md](../README.md)** — High-level overview and 5-second workflow.
- **[TALON.md](TALON.md)** — Focused guide to the active agent harness and hook boundaries.
- **[PLAYBOOK.md](../PLAYBOOK.md)** — Full operational runbook, Garuda Talon setup, and troubleshooting.
- **[docs/CAPABILITIES.md](CAPABILITIES.md)** — 6-plane capability and verification matrix.
- **[docs/API.md](API.md)** — Complete CLI, Talon, Policy YAML, MCP, and HTTP API reference.
- **[docs/SPECS.md](SPECS.md)** — Formal system specifications and boundaries.
