# Garuda CLI, Talon Harness, Policy, MCP & HTTP API Reference (v1.5.0)

This reference documents the complete command-line interface (`bin/garuda` and `bin/garuda-talon`), the **Garuda Talon** active agent harness, declarative policy specification (`policies/*.yaml`), Model Context Protocol surface (`bin/garuda-mcp` with 23 active tools), and HTTP REST gateway (`bin/garuda-api`).

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
Initializes Garuda in a single command: verifies or starts the `garuda-postgres` Docker container, runs all database migrations idempotently, persists `.env` and Ed25519 JWT keys, provisions the default workspace and owner account, indexes the target repository, and runs `garuda talon init` to wire Cursor, Claude Code, Codex, Windsurf, and Git hooks.

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
Analyzes a Go (`go/types`, `go.work`, `vendor/`), Python (AST + cross-file alias/instance resolution), or TypeScript (Tree-sitter + cross-file alias/instance resolution) repository, extracts entities and typed relationships (`CALLS`, `IMPORTS`, `IMPLEMENTS`, `INHERITS`, `EMBEDS`, `REFERENCES`), and persists the semantic graph and Merkle decision revision to PostgreSQL.

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

### Additional CLI commands

These utility commands are implemented alongside the main analysis, policy and Talon workflow.

| Command | Purpose | Important behavior |
|---|---|---|
| `garuda ci` | Analyze the current repository and compare it with a saved semantic baseline | Use `--baseline <file>`; `--block-on-break` exits non-zero when the assessment is breaking |
| `garuda diff <before.json> <after.json>` | Compare two serialized analysis snapshots | `--fail-on-breaking` exits non-zero when breaking changes are detected |
| `garuda self-describe [path]` | Emit a structured JSON product description or Markdown skeleton from source analysis, CLI structure and optional workspace evidence | Use `--markdown`, `--workspace`, and `--output`; product metadata is read from `docs/source/product.yaml` when available |
| `garuda plan` | Request a structured plan from the configured Garuda API truth graph | Requires the API service and accepts `--scope-domain`, `--scope-system`, `--status`, `--at`, and `--format markdown|json` |
| `garuda hygiene [path]` | Produce advisory unreferenced-symbol, duplicate-name and standard-library-alternative candidates from the indexed graph | Does not prove code is dead, does not refresh analysis, and does not persist findings; `garuda ponytail` is a deprecated alias |

Examples:

```bash
# Compare an analysis snapshot with a stored baseline in CI.
./bin/garuda ci --baseline garuda-baseline.json --block-on-break

# Compare two saved analyzer outputs.
./bin/garuda diff before.json after.json --fail-on-breaking

# Emit product description metadata as Markdown.
./bin/garuda self-describe . --markdown --output PRODUCT.md

# Request a scoped plan from the running API.
./bin/garuda plan --scope-domain infrastructure --format markdown
```

---

### `garuda dev` / `garuda up` / `garuda status`
- `./bin/garuda dev` — Starts the unified Garuda daemon (HTTP API server + six core `/dashboard` engineering surfaces plus ADR Decision Lineage + 10-second background runtime verification & Merkle epoch worker).
- `./bin/garuda up` — Starts PostgreSQL container, runs migrations, and launches the gateway stack.
- `./bin/garuda status` — Verifies gateway health (`/health`).

---

## 3. Declarative Policy Predicates (`policies/*.yaml`)

Every policy file specifies `when` predicates evaluated against the workspace graph in PostgreSQL. The policy engine rejects malformed YAML and unknown predicate types rather than treating them as valid rules. Confirm that the CLI/CI caller propagates these errors as a failed check; the exact failure behavior still depends on the command path.

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

## 5. Current HTTP routes (`bin/garuda-api`)

This list is cross-checked against `Server.RegisterRoutes` in `internal/api/handler.go`. The live router is the authority for route availability; a URL present only in a generated or historical OpenAPI file is not necessarily registered.

### Public and browser-session routes

| Method & path | Access | Purpose |
|---|---|---|
| `GET /health`, `GET /healthz` | Public | Liveness |
| `GET /readyz`, `GET /api/v1/system/health` | Public | Readiness; checks the configured database connection when available |
| `GET /favicon.ico` | Public | Empty favicon response |
| `GET /login`, `POST /login` | Public entry; POST validates credentials | Browser sign-in |
| `GET /signup`, `POST /signup` | Public entry; POST provisions the account/tenant/workspace path | Browser sign-up |
| `POST /logout` | Browser session | End the tenant browser session |
| `GET /dashboard` | Tenant session | Workspace Software Intelligence Instrument |
| `GET /api/v1/dashboard/stats` | Tenant session | Workspace dashboard stats |
| `GET /api/v1/dashboard/search` | Tenant session | Workspace search |
| `GET /api/v1/dashboard/sessions` | Tenant session | Session view data |
| `GET /api/v1/dashboard/decisions`, `GET /api/v1/dashboard/decisions/{id}` | Tenant session | Decision list and detail |
| `GET /api/v1/dashboard/policies`, `GET /api/v1/dashboard/policies/verify` | Tenant session | Policy view and verification |
| `GET /api/v1/graph` | Tenant session | Interactive graph data |
| `GET /api/v1/events` | Tenant session | Live events/SSE path |
| `GET /api/v1/workspaces` | Tenant session | Workspace list |
| `GET /system/discover` | Tenant session | System endpoint discovery |
| `GET /admin` | Tenant session plus `GARUDA_ADMIN_EMAILS` allowlist | Legacy deployment-admin dashboard; returns 404 when the allowlist disables it |

### Authenticated API routes

The `/api/v1` API subrouter accepts the configured Ed25519 JWT/token path and the authenticated tenant session cookie, subject to each handler's own access checks.

| Method & path | Purpose |
|---|---|
| `POST /api/v1/decisions/propose` | Propose a scoped governance decision |
| `GET /api/v1/evidence/verify/{id}` | Verify the recorded evidence/Merkle proof for a decision |
| `GET /api/v1/agent/warmup` | Agent context/bootstrap path |
| `POST /api/v1/agents/checkpoint` | Persist an agent checkpoint |
| `POST /api/v1/agents/handoff` | Handoff task ownership |
| `POST /api/v1/agents/resume` | Resume from a checkpoint |
| `GET /api/v1/tasks/{task_id}/lineage` | Read decision/task lineage |
| `POST /api/v1/router/evaluate` | Evaluate the configured routing request |
| `GET /api/v1/telemetry/stream` | Event stream |
| `POST /api/v1/telemetry/spans`, `POST /api/v1/runtime/spans` | Ingest OpenTelemetry runtime spans and correlate observations |
| `GET /api/v1/runtime/coverage` | Read runtime evidence coverage |
| `GET /api/v1/merkle/state` | Read the current Merkle state |

### Owner Control Plane

The operator Control Plane is **separate from** the tenant dashboard and is not under `/api/v1`.

| Method & path | Access / purpose |
|---|---|
| `GET /_/control/login`, `POST /_/control/login` | Token login page and submission |
| `POST /_/control/logout` | Clear the owner Control Plane session |
| `GET /_/control?tab=business` | Owner aggregate business/workspace metrics |
| `GET /_/control?tab=tenants` | Tenant summary and derived health |
| `GET /_/control?tab=operations` | MCP activity, stalled sessions, errors and process-local metrics |
| `GET /_/control?tab=internal` | Internal status view |

The owner Control Plane is served only when `GARUDA_CONTROL_TOKEN` and `CONTROL_DATABASE_URL` are configured and a separate read-only Control Plane pool is available. Unauthenticated/disabled control routes return 404 by design. The older `/admin` route is separate and uses the tenant session plus `GARUDA_ADMIN_EMAILS` allowlist.

### Development-only route

| Method & path | Access / purpose |
|---|---|
| `GET /debug/token` | Debug token helper; deliberately returns 404 when `GARUDA_ENV=production` |

### API contract boundary

The route inventory above describes the current Gorilla router. It does not imply that every endpoint has a stable, fully typed request/response schema. The dashboard workbench and policy UI should not be described as performing a live server-side simulation unless a route and implementation for that operation are present in the release being documented. The top-level [`openapi.yaml`](../openapi.yaml) lists registered routes; `internal/api/openapi.yaml` remains a separate generated-client schema and is not the authoritative inventory of live routes.
