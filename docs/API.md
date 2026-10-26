# Garuda CLI, Talon Harness, Policy, MCP & HTTP API Reference (v1.6.0 Universal)

This reference documents the complete command-line interface (`bin/garuda` and `bin/garuda-talon`), the **Garuda Talon** active agent harness, declarative policy specification (`policies/*.yaml`), Model Context Protocol surface (`bin/garuda-mcp` with 23 active tools), and HTTP REST / PWA gateway (`bin/garuda-api`).

---

## 1. Environment & Zero-Config `.env` Resolution

All Garuda binaries (`garuda`, `garuda-talon`, `garuda-api`, `garuda-mcp`) automatically load `.env` from the current working directory or repository root on startup via `internal/config/env.go`.

| Variable | Default (Dev) | Purpose |
| :--- | :--- | :--- |
| `DATABASE_URL` | `postgres://test:test@localhost:5433/garuda_test?sslmode=disable` | Primary PostgreSQL connection string |
| `GARUDA_WORKSPACE` | Most recently updated workspace | Default workspace name for CLI, Talon, and MCP commands |
| `GARUDA_TALON_ENFORCE` | `true` | Enables active MCP contract injection on the agent's first tool call |
| `GARUDA_POLICY_DIR` | `./policies` | Default directory for declarative YAML architecture policies |
| `PORT` | `8080` | HTTP port for `garuda-api` and `garuda dev` |
| `GARUDA_API_URL` | `http://localhost:${PORT}` | Base URL used by `garuda` CLI when calling the HTTP gateway |
| `JWT_PRIVATE_KEY_HEX` | Auto-generated & persisted to `.env` | Ed25519 64-byte private key (hex) for JWT signing |
| `JWT_PUBLIC_KEY_HEX` | Auto-generated & persisted to `.env` | Ed25519 32-byte public key (hex) for JWT verification |
| `CONTROL_DATABASE_URL` | Defaults to `DATABASE_URL` when token set | Read-only connection pool for the `/admin` Control Plane |
| `GARUDA_CONTROL_TOKEN` | `devtoken-changeme-please` | Operator bearer token for `/admin` Control Plane routes |

---

## 2. CLI Commands (`bin/garuda` & `bin/garuda-talon`)

### `garuda init [target_repo_path]`
Initializes Garuda in a single command: verifies or starts PostgreSQL (local user-space cluster or Docker container), runs all database migrations idempotently, persists `.env` and Ed25519 JWT keys, provisions the default workspace and owner account, indexes the target repository, and runs `garuda talon init` to wire Cursor, Claude Code, Codex, Windsurf, and Git hooks across Windows, macOS, Linux, Chromebook, and Android.

```bash
./bin/garuda init . --workspace go-validation-10 --email myshra777@gmail.com --password garuda2026
```

**Flags:**
- `--workspace string` — Workspace name to initialize (defaults to `GARUDA_WORKSPACE` or target folder name).
- `--email string` — Owner email to provision (default `myshra777@gmail.com`).
- `--password string` — Owner password to provision (default `garuda2026`).
- `--target string` — Repository path to index (default `.`).
- `--skip-scan` — Skip initial repository AST scan during setup.

---

### `garuda doctor [--fix] [--json]`
Diagnoses and auto-repairs your Garuda installation across Windows PowerShell, macOS, Linux, Chromebook (Crostini), and Android (Termux). Verifies:
1. Platform & Runtime Environment (`GOOS/GOARCH`, Crostini, Termux)
2. `.env` Configuration
3. PostgreSQL Reachability & Idempotent Schema Migrations
4. Core Binaries (`garuda`, `garuda-api`, `garuda-mcp`, `garuda-talon`)
5. Garuda Talon Active Contract & IDE/Git Hooks (`.garuda/TALON_CONTRACT.md`, `.cursor/mcp.json`, `.claude/settings.json`, `CLAUDE.md`, `AGENTS.md`)
6. API Gateway & `/dashboard` PWA Live Reachability

```bash
# Diagnose and automatically repair any missing or stopped components
./bin/garuda doctor --fix

# Emit machine-readable JSON diagnostic report
./bin/garuda doctor --json
```

---

### `garuda talon <init|contract|brief|guard|run>` (or `bin/garuda-talon`)
Manages the **Garuda Talon** Active Agent Execution Harness.

```bash
# 1. Compile .garuda/TALON_CONTRACT.md and install Cursor, Claude Code, Codex, Windsurf & Git hooks
./bin/garuda talon init . --workspace go-validation-10 --policies ./policies

# 2. Print the active workspace ADR + Policy contract (Markdown or --json)
./bin/garuda talon contract . --workspace go-validation-10

# 3. Emit deterministic pre-flight context, blast radius, and ADR rules before editing a symbol
./bin/garuda talon brief SQLModel --workspace go-validation-10

# 4. Run incremental AST sync + diff-scoped policy & ADR check + Self-Repair Prompt output
./bin/garuda talon guard . --workspace go-validation-10 --staged --fail-on-block

# 5. Supervisor loop: wrap a CLI coding agent and auto-loop self-repair until ALLOW
./bin/garuda talon run --max-retries 3 -- claude "Refactor CheckoutHandler"
```

**Flags (`garuda talon guard`):**
- `--workspace`, `-w string` — Workspace name.
- `--policies string` — Directory containing policy YAML files (default `./policies`).
- `--staged` — Check staged git index changes (`git diff --cached`).
- `--diff string` — Git revision or range to check (e.g. `HEAD`, `origin/main...HEAD`).
- `--files string` — Comma-separated list of modified files.
- `--hook` — Run in IDE / Claude Code `PostToolUse` hook mode (exits `2` with Self-Repair Prompt on `BLOCK`).
- `--fail-on-block` — Exit with code `2` when outcome is `BLOCK`.
- `--auto-index` — Incrementally re-index dirty files before evaluating (default `true`).
- `--json` — Emit structured JSON `GuardVerdict` payload.

---

### `garuda check [dir]`
Executes the unified 5-plane Change-Control workflow across either a target symbol (`--target`) or a Git diff (`--diff`, `--staged`, `--files`), isolating `net_new_violations` from `baseline_violations` and sealing an RFC 6962 Merkle receipt.

```bash
# Diff-scoped check (isolates net-new violations touching modified symbols)
./bin/garuda check --diff HEAD --project go-validation-10 --policy ./policies --save --fail-on-block

# Staged pre-commit check
./bin/garuda check --staged --project go-validation-10 --policy ./policies --save --fail-on-block

# Symbol-targeted check
./bin/garuda check --project go-validation-10 --target SQLModel --policy ./policies --save --fail-on-block
```

**Flags:**
- `--workspace`, `-w`, `--project string` — Workspace name.
- `--symbol`, `--target string` — Target symbol or service name being modified.
- `--diff string` — Git revision or range to map to affected AST symbols (e.g. `HEAD`, `origin/main...HEAD`).
- `--staged` — Evaluate only symbols touched in the staged git index.
- `--files string` — Comma-separated list of modified files to evaluate.
- `--net-new-only` — Partition and gate only on net-new violations touching the diff (default `true`).
- `--auto-index` — Incrementally re-index dirty working-tree files before evaluating (default `true`).
- `--policies`, `--policy string` — Directory containing policy YAML files (default `./policies`).
- `--save`, `-s` — Persist evaluations and seal RFC 6962 Merkle receipt (default `true`).
- `--fail-on-block` — Exit with code `2` when governance outcome is `BLOCK`.
- `--json` — Emit structured JSON `ChangeControlDecision` payload.

---

### `garuda analyze <path>`
Analyzes a Go (`go/types`, `go.work`, `vendor/`), Python (AST + cross-file alias/instance resolution), or TypeScript (Tree-sitter CGO or Pure-Go `!cgo` + cross-file alias/instance resolution) repository, extracts entities and typed relationships (`CALLS`, `IMPORTS`, `IMPLEMENTS`, `INHERITS`, `EMBEDS`, `REFERENCES`), and persists the semantic graph and Merkle decision revision to PostgreSQL.

```bash
# Full repository analysis
./bin/garuda analyze ~/garuda-validation/go/chi --workspace go-validation-10 --save

# Sub-second incremental re-index of dirty working-tree files
./bin/garuda analyze . --workspace go-validation-10 --incremental --files internal/api/handler.go --save
```

**Flags:**
- `--save`, `-s` — Persist extracted entities, claims, and Merkle revision to PostgreSQL.
- `--workspace`, `-w string` — Target workspace name.
- `--incremental` — Reuse in-memory package cache and re-index only modified packages/files.
- `--files string` — Comma-separated list of dirty files for incremental sync.
- `--output`, `-o string` — Write JSON snapshot to file.
- `--json` — Print JSON snapshot to stdout.
- `--diff <before.json> <after.json>` — Compute structural schema and breaking-change diff between two snapshots.
- `--fail-on-breaking` — Exit non-zero if breaking changes are detected during `--diff`.

---

### `garuda policy <subcommand>`
Deterministic policy enforcement lifecycle commands (fail-closed on malformed YAML or unknown predicates):

```bash
./bin/garuda policy validate ./policies
./bin/garuda policy evaluate ./policies --workspace go-validation-10 --save --fail-on-block
./bin/garuda policy list
./bin/garuda policy show <evaluation-id>
./bin/garuda policy verify <evaluation-id>
```

---

### `garuda docs <ingest|verify|status>`
Extracts normative architectural claims (`MUST` / `MUST_NOT`) from Markdown ADRs and specifications and verifies them against the static codebase graph.

```bash
./bin/garuda docs ingest docs/adr/0001-payment-idempotency.md
./bin/garuda docs verify
./bin/garuda docs status
```

---

### `garuda summary`, `garuda topology`, `garuda impact`, & `garuda hygiene`
- `./bin/garuda summary` — Executive architectural summary of repositories, entities, relationships, top hubs, and Merkle ledger height.
- `./bin/garuda topology` — Package-level layering, fan-in/fan-out, and circular dependency report.
- `./bin/garuda impact --workspace <ws-uuid> --target <entity-uuid> --depth 3` — Confidence-weighted transitive blast radius.
- `./bin/garuda hygiene [path]` (alias `garuda ponytail`) — Advisory static report for unreferenced symbols and duplicate candidates.

---

### `garuda dev` / `garuda up` / `garuda status`
- `./bin/garuda dev` — Starts the unified Garuda daemon (HTTP API server + 6-surface `/dashboard` Software Intelligence Instrument + 10-second background runtime verification & Merkle epoch worker).
- `./bin/garuda up` — Starts PostgreSQL container, runs migrations, and launches the gateway stack.
- `./bin/garuda status` — Verifies gateway health (`/health`).

---

## 3. Declarative Policy Predicates (`policies/*.yaml`)

Every policy file specifies `when` predicates evaluated against the workspace graph in PostgreSQL. If any `.yaml`/`.yml` file in the policy directory is malformed or references an unknown predicate, Garuda **fails closed** immediately.

| Predicate | Parameters | Description |
| :--- | :--- | :--- |
| `layer_violation` | `from_layer` (regex), `to_layer` (regex), `limit` | Fires when any entity in `from_layer` directly `CALLS` or `IMPORTS` an entity in `to_layer`. |
| `transitive_claim_exists` | `claim_type`, `from_name_pattern`, `to_name_pattern`, `max_depth` (1–10), `limit` | Recursive CTE traversal that fires when a multi-hop chain connects `from_name_pattern` to `to_name_pattern`. |
| `cycle_exists` | `max_depth` (2–8), `limit` | Recursive CTE that detects circular package dependencies (`PkgA -> PkgB -> PkgA`). |
| `blast_radius_exceeds` | `max_callers` (int), `name_pattern` (regex), `limit` | Fires when an entity has more than `max_callers` distinct inbound dependents. |
| `claim_exists` | `claim_type`, `from_name_pattern`, `to_name_pattern`, `min_confidence` | Fires when a direct relationship edge matches the filters and `scope.package`/`scope.pattern`. |
| `entity_exists` | `kind`, `name_pattern`, `exported` | Fires when an entity matches the filter and `scope.package`/`scope.pattern`. |
| `contradiction_exists` | `severity`, `unresolved_only` | Fires when unresolved runtime or decision contradictions exist. |
| `verification_missing` | `kind`, `exported`, `limit` | Fires when matching entities have no `SUPPORTED` runtime verification row. |
| `language_matches` | `language` (`go`, `python`, `typescript`) | Scopes evaluation to repositories/entities of the target language. |

---

## 4. Active MCP Surface (`bin/garuda-mcp`)

`bin/garuda-mcp` speaks JSON-RPC 2.0 over stdio (`protocolVersion: "2025-06-18"`), advertises `tools`, `prompts`, and `resources` capabilities, and exposes **23 deterministic tools**:

| Category | Tool Name | Whitelisted Arguments | Purpose |
| :--- | :--- | :--- | :--- |
| **Active Talon Harness** | `garuda.talon_pre_edit` | `workspace`, `target`, `tenant_id` | Mandatory pre-edit briefing returning blast radius, callers, ADR rules, and active policies. |
| **Active Talon Harness** | `garuda.talon_verify_edit` | `workspace`, `repo_dir`, `staged`, `diff`, `files`, `tenant_id` | Mandatory post-edit incremental AST sync, diff-scoped policy/ADR evaluation, Merkle seal, and Self-Repair Prompt generator. |
| **Orientation & Discovery** | `garuda.briefing` | `workspace`, `tenant_id` | Workspace overview, entity/claim counts, and governance posture. |
| **Orientation & Discovery** | `garuda.entities` | `workspace`, `package`, `kind`, `limit` | List indexed entities filtered by package or kind. |
| **Orientation & Discovery** | `garuda.find_entity` | `workspace`, `name_pattern`, `kind`, `package`, `limit` | Resolve symbols ranked by inbound edge centrality. |
| **Orientation & Discovery** | `garuda.inspect` | `workspace`, `entity_id` | Deep inspection of an entity's metadata, file span, and claims. |
| **Orientation & Discovery** | `garuda.query` | *(none)* | Execute read-only structural workspace overview query. |
| **Topology & Blast Radius** | `garuda.neighbors` | `workspace`, `entity_id`, `limit` | Retrieve 1-hop inbound and outbound relationships with direction. |
| **Topology & Blast Radius** | `garuda.blast_radius` | `workspace`, `entity_id`, `depth`, `min_confidence` | Compute confidence-weighted transitive impact and severity breakdown. |
| **Topology & Blast Radius** | `garuda.implementers` | `workspace`, `entity_id`, `name`, `limit` | Find all structs/classes implementing an interface (`IMPLEMENTS`). |
| **Topology & Blast Radius** | `garuda.subclasses` | `workspace`, `entity_id`, `name`, `language`, `limit` | Find all types inheriting or embedding a base type (`INHERITS`, `EMBEDS`). |
| **Topology & Blast Radius** | `garuda.get_impact` | `decision_id`, `tenant_id` | Inspect downstream impact associated with a governance decision. |
| **Topology & Blast Radius** | `garuda.get_lineage` | `decision_id`, `tenant_id` | Traverse ancestor and descendant decision lineage. |
| **Governance & Policy** | `garuda.policy.list` | `tenant_id` | List active declarative architecture policies. |
| **Governance & Policy** | `garuda.policy.evaluate` | `workspace` | Evaluate active policies against the workspace graph. |
| **Governance & Policy** | `garuda.verify_policy_evaluation` | `evaluation_id`, `tenant_id` | Cryptographically verify an evaluation's RFC 6962 Merkle inclusion proof. |
| **Governance & Policy** | `garuda.governance.status` | `workspace` | Return current Merkle block height, root hash, and contradiction counts. |
| **Governance & Policy** | `garuda.propose_decision` | `scope_domain`, `scope_system`, `tenant_id` | Propose and record a scoped architectural decision. |
| **Epistemic Truth & Drift** | `garuda.detect_contradictions` | `tenant_id` | List quarantined runtime and ADR contradictions (`CONTRADICTED`). |
| **Epistemic Truth & Drift** | `garuda.check_drift` | `workspace` | Check static-vs-runtime and documentation-to-code drift status. |
| **Epistemic Truth & Drift** | `garuda.query_claims` | `workspace`, `subject` | Query extracted documentation/ADR claims (`document_claims`) and static edges. |
| **Multi-Agent Continuity** | `garuda.handoff` | `task_id`, `source_agent_id`, `target_agent_id` | Atomically transfer task ownership and create a resumable checkpoint. |
| **Multi-Agent Continuity** | `garuda.resume` | `agent_id`, `checkpoint_id` | Consume a handoff checkpoint and restore agent context. |

Verify MCP and Talon compliance at any time by running:
```bash
python3 scripts/mcp_verify.py
```
Expected result: `═══ 30/30 checks passed ═══`.

---

## 5. Key HTTP Gateway & PWA Endpoints (`bin/garuda-api`)

All `/api/v1/*` routes enforce cryptographic Ed25519 JWT validation (`AuthMiddleware`) or authenticated session cookies (`garuda_session`).

| Method & Path | Auth | Description |
| :--- | :--- | :--- |
| `GET /health`, `GET /healthz`, `GET /readyz` | Public | Gateway liveness and PostgreSQL readiness checks |
| `GET /manifest.json`, `GET /sw.js`, `GET /icon.svg` | Public | Progressive Web App (PWA) manifest, service worker, and vector icon for Chromium/Android/Desktop installation |
| `GET /signup`, `POST /signup` | Public | Browser & CLI/API account + tenant + workspace provisioning |
| `GET /login`, `POST /login` | Public | Session authentication (`garuda_session` cookie) |
| `GET /dashboard` | Session | 6-surface Software Intelligence Instrument (`Briefing`, `Map`, `Workbench`, `Policies`, `Contradictions`, `Ledger`) |
| `GET /api/v1/dashboard/instrument` | Session / Bearer | Complete JSON payload powering the 6-surface dashboard instrument |
| `GET /api/v1/dashboard/neighborhood` | Session / Bearer | 1-hop / 2-hop focal symbol neighborhood graph & 4-tab Inspector Drawer payload |
| `POST /api/v1/dashboard/workbench/simulate` | Session / Bearer | Live Change Review Workbench pre-merge policy & blast-radius simulator |
| `GET /graph`, `GET /api/v1/graph` | Session / Bearer | Interactive D3 force graph & JSON graph payload |
| `POST /api/v1/telemetry/spans` | Bearer / Session | OpenTelemetry runtime span ingestion & synchronous edge-level contradiction detection |
| `GET /api/v1/evidence/verify/{id}` | Bearer / Session | RFC 6962 cryptographic Merkle inclusion proof verification for a decision |
| `GET /api/v1/merkle/state` | Bearer / Session | Current tenant Merkle root hash and epoch block height |
| `GET /admin` | Control Token | Multi-tenant operator Control Plane dashboard |
