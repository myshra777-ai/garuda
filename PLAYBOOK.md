# Garuda Operational Playbook (v1.8.1 Universal)

**Zero-Dependency Setup · Active Agent Harness (`Garuda Talon`) · Diff-Scoped Change Control · 6-Surface Dashboard PWA · Reference**

This playbook is the complete operational guide for installing, configuring, and running Garuda v1.6 against real-world Go, Python, and TypeScript codebases across **Windows 10/11 (PowerShell)**, **macOS**, **Linux**, **Chromebook (ChromiumOS)**, and **Android (Termux & Chrome PWA)**.

---

## Table of Contents

1. [What Garuda Is & Why It Exists](#1-what-garuda-is--why-it-exists)
2. [How Garuda Helps Every Persona](#2-how-garuda-helps-every-persona)
3. [1-Command Universal Setup (Zero Dependencies)](#3-1-command-universal-setup-zero-dependencies)
4. [Instant Self-Diagnostics & Repair (`garuda doctor`)](#4-instant-self-diagnostics--repair-garuda-doctor)
5. [Garuda Talon: Active Agent Execution Harness](#5-garuda-talon-active-agent-execution-harness)
6. [Diff-Scoped Change-Control Gate (`garuda check`)](#6-diff-scoped-change-control-gate-garuda-check)
7. [Analyzing Go, Python & TypeScript Repositories](#7-analyzing-go-python--typescript-repositories)
8. [Authoring & Evaluating Declarative YAML Policies](#8-authoring--evaluating-declarative-yaml-policies)
9. [Ingesting & Verifying Markdown ADRs (`garuda docs`)](#9-ingesting--verifying-markdown-adrs-garuda-docs)
10. [Runtime OpenTelemetry Verification & Contradiction Quarantine](#10-runtime-opentelemetry-verification--contradiction-quarantine)
11. [6-Surface Software Intelligence Instrument (`/dashboard` PWA)](#11-6-surface-software-intelligence-instrument-dashboard-pwa)
12. [23 Active MCP Tools, Prompts & Resources](#12-23-active-mcp-tools-prompts--resources)
13. [Multi-Agent Checkpoints, Handoff & Resume](#13-multi-agent-checkpoints-handoff--resume)
14. [Enforcing Gates in CI/CD & Pre-Commit](#14-enforcing-gates-in-cicd--pre-commit)
15. [Troubleshooting & Common Error Reference](#15-troubleshooting--common-error-reference)

---

## 1. What Garuda Is & Why It Exists

AI coding agents (**Cursor**, **Claude Code**, **Codex**, **Devin**, **Windsurf**) write code faster than human reviewers can audit. A pull request can pass unit tests and linters while quietly breaking your architecture:
- Introducing a forbidden `Handler -> Database` direct call or import.
- Creating a circular package dependency (`A -> B -> C -> A`).
- Modifying a core shared struct or class with 170+ downstream callers without checking blast radius.
- Violating documented Architecture Decision Records (ADRs) or making undeclared runtime service calls observed in OpenTelemetry spans.

**Garuda v1.6 is the deterministic change-control gate, active AI execution harness (`Garuda Talon`), and software intelligence instrument for AI-written code.**

### The 5-Plane Architecture

| Plane | Responsibility | Primary CLI / API Surface |
|---|---|---|
| **Plane 1: Compiler & AST Graph Engine** | Extracts deterministic symbols (`UUIDv5`) and confidence-weighted edges (`CALLS`, `IMPORTS`, `IMPLEMENTS`, `INHERITS`, `EMBEDS`, `REFERENCES`) across Go (`go/types` + `go.work` at `1.0`), Python (`0.95`), and TypeScript (`0.95`, supporting both Tree-sitter CGO and Pure-Go `!cgo`). | `garuda analyze <path> [--incremental] [--save]` |
| **Plane 2: 3-State Epistemic & Runtime Verifier** | Classifies every claim as `SUPPORTED`, `UNVERIFIED`, or `CONTRADICTED` by correlating static AST edges with ingested Markdown ADRs and live OpenTelemetry trace spans (`caller_symbol` + `target_endpoint`). | `garuda docs ingest`, `garuda docs verify`, `POST /api/v1/telemetry/spans` |
| **Plane 3: Fail-Closed Graph Policy Engine** | Evaluates 9 deterministic SQL/CTE policy predicates (`layer_violation`, `transitive_claim_exists`, `cycle_exists`, `blast_radius_exceeds`, `claim_exists`, `entity_exists`, `contradiction_exists`, `verification_missing`, `language_matches`) with Git-diff scoping (`net_new_violations` vs. `baseline_violations`). | `garuda policy evaluate`, `garuda check --diff` |
| **Plane 4: Unified Change-Control & RFC 6962 Merkle Ledger** | Computes a deterministic `ALLOW`, `WARN`, `REVIEW`, or `BLOCK` verdict and anchors persisted evaluations in an append-only SHA-256 RFC 6962 Merkle tree with verifiable inclusion proofs. | `garuda check`, `GET /api/v1/evidence/verify/{id}` |
| **Plane 5: Garuda Talon Active Harness & 6-Surface Instrument** | Forces AI coding agents to respect ADR and policy contracts before and after every edit via Active MCP (`GARUDA_TALON_ENFORCE=true`), IDE hooks, Git hooks, supervisor loops, and the `/dashboard` PWA. | `garuda talon`, `bin/garuda-mcp`, `http://localhost:8080/dashboard` |

---

## 2. How Garuda Helps Every Persona

- **Product Managers & Non-Technical Founders:** Run one command in Windows PowerShell (`install.ps1`) or macOS Terminal (`install.sh`) with **zero pre-installed dependencies**. Open `http://localhost:8080/dashboard` (or install it as a desktop/mobile PWA) to inspect workspace health, high-risk coupling hubs, whether AI-generated features obey documented architecture rules, and an auditable decision log.
- **Software Engineers:** Run `garuda talon brief <Symbol>` before refactoring to see every caller and active rule, and run `garuda check --diff HEAD` to check only your uncommitted changes without getting blocked by pre-existing baseline debt.
- **Staff / Principal Architects:** Commit declarative rules in `policies/*.yaml` and Markdown ADRs in `docs/adr/*.md`. Garuda compiles them into `.garuda/TALON_CONTRACT.md` and enforces them automatically across every developer's editor, pre-commit hook, and CI pipeline.
- **AI Coding Agents (Cursor, Claude Code, Codex, Windsurf):** Receive active architectural guardrails on the very first MCP call (`GARUDA_TALON_ENFORCE=true`), query 23 deterministic MCP tools, and receive structured **Self-Repair Prompts** whenever an edit triggers a `BLOCK`.
- **Security & Governance Leads:** Verify cryptographic RFC 6962 Merkle inclusion proofs for every saved architecture decision and inspect quarantined runtime OpenTelemetry contradictions.

---

## 3. 1-Command Universal Setup (Zero Dependencies)

You do **not** need Go, PostgreSQL, GCC/Xcode/MinGW, or Docker pre-installed. Garuda provisions a portable user-space toolchain (`~/.garuda/toolchain/go`) and database (`~/.garuda/data/pg` on port `5433`) automatically:

### Windows 10 / 11 (PowerShell)

```powershell
# Option A: 1-Command Remote Install & Launch
Set-ExecutionPolicy Bypass -Scope Process -Force; irm https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.ps1 | iex

# Option B: From inside a cloned/extracted repository
cd $HOME\garuda
powershell -ExecutionPolicy Bypass -File .\setup.ps1
```

### macOS, Linux, Chromebook (Crostini) & Android (Termux)

```bash
# Option A: 1-Command Remote Install & Launch
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash

# Option B: From inside a cloned/extracted repository
git clone https://github.com/myshra777-ai/garuda.git && cd garuda && ./setup.sh
```

### Direct CLI Repository Initialization (`garuda init`)

Once compiled, you can initialize any repository on your machine in one command:

```bash
./bin/garuda init /path/to/repo --workspace my-system --email myshra777@gmail.com --password garuda2026
```

All Garuda binaries (`garuda`, `garuda-api`, `garuda-mcp`, `garuda-talon`) automatically load `.env` from your working directory—you never need to manually export `DATABASE_URL`, `GARUDA_WORKSPACE`, `PORT`, or `JWT_PRIVATE_KEY_HEX` across terminal tabs.

---

## 4. Instant Self-Diagnostics & Repair (`garuda doctor`)

Run `garuda doctor` at any time to verify your platform environment, `.env` configuration, PostgreSQL connectivity and schema migrations, compiled binaries, Garuda Talon active hooks, and `/dashboard` gateway health. Pass `--fix` to automatically repair any missing or stopped components:

```bash
# macOS / Linux / Chromebook / Android
./bin/garuda doctor --fix

# Windows PowerShell
.\bin\garuda.exe doctor --fix

# Emit machine-readable JSON diagnostics
./bin/garuda doctor --json
```

---

## 5. Garuda Talon: Active Agent Execution Harness

Standard MCP servers are passive—if an agent forgets to call a tool, it writes code blind. **Garuda Talon** makes architectural enforcement active across every stage of agent execution:

### 5.1 Compile Contract & Install Active Hooks (`garuda talon init`)

```bash
./bin/garuda talon init . --workspace my-system --policies ./policies
```

This single command compiles live workspace ADRs (`document_claims`), YAML policies (`policies/*.yaml`), and top high-blast-radius hubs into:
1. `.garuda/TALON_CONTRACT.md` — Human- and machine-readable architectural contract.
2. `.cursor/rules/garuda-talon.mdc` — Cursor rule with `alwaysApply: true` requiring pre-edit `garuda.talon_pre_edit` and post-edit `garuda.talon_verify_edit`.
3. `.cursor/mcp.json` — Registers `garuda-mcp` with `GARUDA_TALON_ENFORCE=true`.
4. `CLAUDE.md` & `.claude/settings.json` — Registers a Claude Code `PostToolUse` hook on `Edit|Write|MultiEdit` running `garuda talon guard --hook`.
5. `AGENTS.md` & `.windsurfrules` — Idempotent `<!-- GARUDA-TALON:START -->` contract blocks for Codex, Aider, and Windsurf.
6. `.git/hooks/pre-commit` & `.git/hooks/commit-msg` — Blocks commits with net-new policy or ADR violations and stamps `Garuda-Governance-Outcome` and `Garuda-Merkle-Block` trailers into verified commits.

### 5.2 Pre-Edit Briefing (`garuda talon brief`)

Before modifying a symbol, inspect its blast radius, callers, ADR rules, and active policies:

```bash
./bin/garuda talon brief CheckoutHandler --workspace my-system
```

### 5.3 Post-Edit Active Diff Guard (`garuda talon guard`)

Incrementally re-indexes dirty working-tree files in milliseconds, maps git diff hunks (`--staged`, `--diff <rev>`, or `--files`) to modified AST symbols, evaluates net-new policy and ADR violations, and outputs a deterministic **Self-Repair Prompt** on `BLOCK`:

```bash
# Check uncommitted working-tree changes
./bin/garuda talon guard --workspace my-system --fail-on-block

# Check staged git index changes (used by .git/hooks/pre-commit)
./bin/garuda talon guard --staged --workspace my-system --fail-on-block
```

### 5.4 Autonomous Supervisor Loop (`garuda talon run`)

Wrap any CLI coding agent (`claude`, `aider`, `codex`) in a closed self-repair loop that automatically feeds policy/ADR violation fixes back to the agent (up to `--max-retries 3`) until `ALLOW` is achieved:

```bash
./bin/garuda talon run --max-retries 3 -- claude "Refactor CheckoutHandler to use OrderService"
```

---

## 6. Diff-Scoped Change-Control Gate (`garuda check`)

`garuda check` runs all 5 planes across either a specific symbol (`--target`) or a Git diff (`--diff`, `--staged`, `--files`), isolating `net_new_violations` touching your diff from pre-existing `baseline_violations` and sealing an RFC 6962 Merkle receipt:

```bash
# 1. Evaluate only the symbols modified in your current working-tree diff against HEAD
./bin/garuda check --diff HEAD --project my-system --policy ./policies --save --fail-on-block

# 2. Evaluate a PR branch range in CI/CD and emit structured JSON
./bin/garuda check --diff origin/main...HEAD --project my-system --policy ./policies --save --fail-on-block --json

# 3. Evaluate a specific symbol across all 5 planes
./bin/garuda check --project my-system --target SQLModel --policy ./policies --save --fail-on-block
```

**Exit Code Contract:**
- Exit `0`: Governance outcome is `ALLOW`, `WARN`, or `REVIEW` (or `BLOCK` when `--fail-on-block` is not set).
- Exit `2`: Governance outcome is `BLOCK` and `--fail-on-block` is set.
- Exit `1`: Execution, configuration, or fail-closed policy parse error.

---

## 7. Analyzing Go, Python & TypeScript Repositories

```bash
# Full repository scan with RFC 6962 Merkle baseline revision
./bin/garuda analyze /path/to/repo --workspace my-system --save

# Sub-second incremental sync of specific modified files
./bin/garuda analyze . --workspace my-system --incremental --files internal/api/handler.go --save
```

- **Go (`1.0` compiler confidence):** Uses `go/types`, `go/ssa`, `go.work`, and `vendor/` to resolve structs, interfaces, methods, pointer/value receivers, `IMPLEMENTS`, `EMBEDS`, `IMPORTS`, and intra/cross-module `CALLS`.
- **Python (`0.95` structural + alias/instance confidence):** Resolves modules, classes, functions, `import X as Y`, `from X import Y as Z`, `self.attr = Class()`, and cross-file instance method calls.
- **TypeScript (`0.95` Tree-sitter + Pure-Go `!cgo` confidence):** Resolves interfaces, classes, functions, named/namespace import aliases (`import { X as Y }`, `import * as NS`), `new Class()` instance receivers, and cross-file method calls—working seamlessly both with CGO Tree-sitter and in Pure-Go (`CGO_ENABLED=0`) builds.

---

## 8. Authoring & Evaluating Declarative YAML Policies

Store policy files in `./policies/*.yaml`. Garuda supports **9 deterministic graph-native SQL/CTE predicates**:

| Predicate | Purpose | Key Parameters |
|---|---|---|
| `layer_violation` | Block direct `CALLS` or `IMPORTS` across architectural layers | `from_layer`, `to_layer` (regex) |
| `transitive_claim_exists` | Block multi-hop backdoor dependency chains (`A -> B -> C`) | `from_name_pattern`, `to_name_pattern`, `claim_type`, `max_depth` |
| `cycle_exists` | Detect circular package dependencies (`A -> B -> C -> A`) | `max_depth` (2..6) |
| `blast_radius_exceeds` | Escalate changes to high-fan-in hubs | `max_callers` |
| `claim_exists` | Match specific static edges (`CALLS`, `IMPORTS`, `IMPLEMENTS`) | `claim_type`, `from_name_pattern`, `to_name_pattern`, `min_confidence` |
| `entity_exists` | Match symbols by kind, naming pattern, and visibility | `kind`, `name_pattern`, `exported` |
| `contradiction_exists` | Block workspaces with unresolved runtime or ADR contradictions | `severity` |
| `verification_missing` | Warn when exported symbols lack runtime `SUPPORTED` verification | `kind` |
| `language_matches` | Scope policies to `go`, `python`, or `typescript` | `language` |

Evaluate policies directly from the CLI:

```bash
./bin/garuda policy evaluate --policy ./policies --workspace my-system --save
```

---

## 9. Ingesting & Verifying Markdown ADRs (`garuda docs`)

Garuda extracts normative `MUST`, `REQUIRED`, `SHALL`, `MUST NOT`, `SHALL NOT`, and `FORBIDDEN` architectural claims from Markdown ADRs and verifies them against the semantic code graph:

```bash
# 1. Ingest Markdown ADRs from a directory or file
./bin/garuda docs ingest ./docs/adr --workspace my-system

# 2. Verify extracted ADR claims against static & runtime graph edges
./bin/garuda docs verify --workspace my-system

# 3. View verification summary (SUPPORTED / UNVERIFIED / CONTRADICTED)
./bin/garuda docs status --workspace my-system
```

---

## 10. Runtime OpenTelemetry Verification & Contradiction Quarantine

Ingest runtime spans via `POST /api/v1/telemetry/spans` (or run the verification script `python3 scripts/runtime_contradiction_test.py`):

- **Edge-Level `SUPPORTED` Promotion:** A static edge (`Caller -> Callee`) is promoted from `UNVERIFIED` to `SUPPORTED` only when runtime spans corroborate **both** ends of the relationship.
- **Shadow Call `CONTRADICTED` Quarantine:** If runtime spans observe a service or symbol calling an unapproved target that does not exist in the static graph (or violates an ADR), Garuda automatically creates a `CONTRADICTED` quarantine record and surfaces it in `/dashboard#contradictions` and `garuda.detect_contradictions`.

---

## 11. 6-Surface Software Intelligence Instrument (`/dashboard` PWA)

Open `http://localhost:8080/dashboard` (or click **"Install Garuda"** in Chrome/Edge/Chromebook/Android to install it as a standalone Progressive Web App):

1. **Workspace Briefing (`#briefing`)** — Executive health strip, Needs-Attention queue, and recent governance verdicts.
2. **Architecture Map (`#map`)** — 1-hop / 2-hop neighborhood topology explorer (`[Caller] -> [Focal Symbol] -> [Dependency]`) with edge-kind filters and a 4-tab Inspector Drawer (`Source`, `Relationships`, `Policy Results`, `Evidence`).
3. **Change Review Workbench (`#workbench`)** — Interactive 3-pane pre-merge simulator (`Target & Policy Selector` -> `Blast Radius & Rule Breakdown` -> `Live Gate Receipt`).
4. **Policies (`#policies`)** — Catalog of all 9 predicates plus a live YAML Policy Sandbox.
5. **Contradictions & Evidence (`#contradictions`)** — Forensic queue comparing static AST source lines against live OpenTelemetry runtime trace spans, plus an interactive OTel Span Simulator.
6. **Governance & Audit Ledger (`#ledger`)** — Append-only decision table with a bottom Decision Record Inspector and 1-click RFC 6962 Merkle Inclusion Proof verification.

Press `⌘K` (macOS) or `Ctrl+K` (Windows/Linux/Chromebook) from any surface to search symbols, jump between views, or trigger a live change review.

---

## 12. 23 Active MCP Tools, Prompts & Resources

`bin/garuda-mcp` implements JSON-RPC 2.0 over stdio and exposes:
- **MCP Prompts (`prompts/list`, `prompts/get`):** `garuda.architectural_contract`
- **MCP Resources (`resources/list`, `resources/read`):** `garuda://workspace/contract`, `garuda://workspace/adr-guardrails`
- **23 MCP Tools (`tools/list`, `tools/call`):**
  - *Active Talon Enforcement:* `garuda.talon_pre_edit`, `garuda.talon_verify_edit`
  - *Orientation & Discovery:* `garuda.briefing`, `garuda.entities`, `garuda.find_entity`, `garuda.inspect`, `garuda.query`
  - *Topology & Impact:* `garuda.neighbors`, `garuda.blast_radius`, `garuda.implementers`, `garuda.subclasses`, `garuda.get_impact`, `garuda.get_lineage`
  - *Policy & Governance:* `garuda.policy.list`, `garuda.policy.evaluate`, `garuda.verify_policy_evaluation`, `garuda.governance.status`, `garuda.propose_decision`
  - *Epistemic Evidence & Drift:* `garuda.detect_contradictions`, `garuda.check_drift`, `garuda.query_claims`
  - *Multi-Agent Continuity:* `garuda.handoff`, `garuda.resume`

Verify all 30 MCP & Talon protocol checks at any time with:

```bash
python3 scripts/mcp_verify.py
```

---

## 13. Multi-Agent Checkpoints, Handoff & Resume

When multiple coding agents collaborate across sessions (`planner-agent` -> `implementer-agent` -> `reviewer-agent`), Garuda preserves task lineage and state transactionally:
- `garuda.handoff` / `POST /api/v1/agents/handoff`: Creates a serialized checkpoint and transfers task ownership.
- `garuda.resume` / `POST /api/v1/agents/resume`: Atomically consumes the pending handoff (`UPDATE ... RETURNING` with `409 Conflict` replay protection) so two agents can never resume the same handoff twice.

---

## 14. Enforcing Gates in CI/CD & Pre-Commit

Add Garuda to GitHub Actions (`.github/workflows/garuda-gate.yml`) or GitLab CI:

```yaml
- name: Run Garuda Talon Diff-Scoped Architecture Gate
  run: |
    ./bin/garuda talon guard --diff origin/main...HEAD --workspace default --policies ./policies --fail-on-block
```

---

## 15. Troubleshooting & Common Error Reference

For platform-specific installation errors (Windows PowerShell ExecutionPolicy, missing GCC/Xcode, stopped PostgreSQL, port conflicts, Chromebook port forwarding, Android Termux storage, or `git push origin main` non-fast-forward rejections), run `./bin/garuda doctor --fix` or consult the complete 13-item table in **[`INSTALL.md`](INSTALL.md)**.

---

## 16. v1.8.1 Operational Workflows: Docs-to-Code Lineage Graph, Talon Risk Hub & Deterministic BuildContext Matrix

### 16.1 Exploring & Ingesting into the 4-Tier Docs, Specs & Guardrails Graph
Every Markdown ADR (`docs/adr/*.md`), design spec (`docs/specs/*.md`), system invariant (`docs/invariants.md`), and YAML policy (`policies/*.yaml`) is automatically parsed into a 4-tier lineage graph (`Document -> Normative Rule -> Active Guardrail -> Code Symbol`):

```bash
# Ingest Markdown ADRs and design specs from the CLI
./bin/garuda docs ingest ./docs/adr --workspace default

# Inspect the 4-tier Docs-to-Code lineage graph as JSON or in the dashboard
curl -s http://localhost:8080/api/v1/dashboard/docs-graph?workspace=default | jq .summary
```
Or open **`http://localhost:8080/dashboard#docs-graph`** (`04 — Docs & Rules Graph`) to visually trace any ADR rule directly to the Go, Python, or TypeScript symbols it governs, or click **`+ Ingest ADR / Spec`** to author and bind a new architectural rule live in your browser.

### 16.2 Verifying Deterministic `BuildContext` Receipts & Catching CI-vs-Editor Drift
When a check passes in a developer's local IDE (`darwin/arm64`, `CGO_ENABLED=1`) but fails or behaves differently in Linux CI (`linux/amd64`, `CGO_ENABLED=0`), Garuda's `BuildContext` provenance receipt pinpoints the exact environmental difference and gives you a 1-line reproduction command:

```bash
# Run diff-scoped Talon guard with cross-platform matrix verification
./bin/garuda talon guard --workspace default --diff HEAD --matrix --fail-on-block

# Run unified check across all 4 OS/CGO matrix profiles and emit JSON receipt
./bin/garuda check --diff HEAD --project default --policy ./policies --matrix --save --json
```

If environment drift is detected between the indexed workspace graph and the active check runner (`TARGET_PLATFORM_MISMATCH`, `CGO_MISMATCH`, `BUILD_TAGS_MISMATCH`, `COMMIT_DRIFT`, or `WORKTREE_DIRTY_DRIFT`), Garuda surfaces the exact drift code and outputs the deterministic `reproduce_command` (`CGO_ENABLED=... GOOS=... GOARCH=... ./bin/garuda talon guard ...`) so any engineer or CI runner can reproduce the exact graph and verdict.
