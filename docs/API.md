# Garuda CLI, Talon Harness, Policy, MCP & HTTP API Reference

This reference documents the complete command-line interface (`bin/garuda` and `bin/garuda-talon`), the **Garuda Talon** active agent harness, declarative policy specification (`policies/*.yaml`), Model Context Protocol surface (`bin/garuda-mcp` with 54 active tools), and HTTP REST / PWA gateway (`bin/garuda-api`).

---

## 1. Environment & Zero-Config `.env` Resolution

All Garuda binaries (`garuda`, `garuda-talon`, `garuda-api`, `garuda-mcp`) automatically load `.env` from the current working directory or repository root on startup via `internal/config/env.go`.

| Variable | Default (Dev) | Purpose |
| :--- | :--- | :--- |
| `DATABASE_URL` | `postgres://test:test@127.0.0.1:5433/garuda_test?sslmode=disable` | Primary PostgreSQL connection string (`config.ResolveDBURL()` automatically normalizes `@localhost:` to IPv4 `@127.0.0.1:`) |
| `GARUDA_WORKSPACE` | Most recently updated workspace | Default workspace name for CLI, Talon, and MCP commands |
| `GARUDA_TALON_ENFORCE` | `true` | Enables active MCP contract injection on the agent's first tool call |
| `GARUDA_POLICY_DIR` | `./policies` | Default directory for declarative YAML architecture policies |
| `PORT` | `8080` | HTTP port for `garuda-api` and `garuda dev` |
| `GARUDA_API_URL` | `http://localhost:${PORT}` | Base URL used by `garuda` CLI when calling the HTTP gateway |
| `JWT_PRIVATE_KEY_HEX` | Auto-generated & persisted to `.env` | Ed25519 64-byte private key (hex) for JWT signing |
| `JWT_PUBLIC_KEY_HEX` | Auto-generated & persisted to `.env` | Ed25519 32-byte public key (hex) for JWT verification |
| `CONTROL_DATABASE_URL` | Defaults to `DATABASE_URL` when token set | Read-only connection pool for the `/admin` Control Plane |
| `GARUDA_CONTROL_TOKEN` | Random value written to `.env` by `garuda init` / `setup.sh` / `setup.ps1` | Operator bearer token for `/admin` Control Plane routes. There is no built-in default: when it is unset the Control Plane stays disabled. |

---

## 2. CLI Commands (`bin/garuda` & `bin/garuda-talon`)

### `garuda init [target_repo_path]`
Initializes Garuda in a single command: verifies or starts PostgreSQL (local user-space cluster or Docker container), runs all database migrations idempotently, persists `.env` and Ed25519 JWT keys, provisions the default workspace and owner account, indexes the target repository, and runs `garuda talon init` to wire Cursor, Claude Code, Codex, Windsurf, and Git hooks across Windows, macOS, Linux, Chromebook, and Android.

```bash
export GARUDA_PASSWORD='choose-a-long-passphrase'   # at least 8 characters
./bin/garuda init . --workspace my-system --email you@example.com
```

**Flags:**
- `--workspace string` — Workspace name to initialize (defaults to `GARUDA_WORKSPACE` or target folder name).
- `--email string` — Owner email to provision (auto-discovered from `GARUDA_EMAIL`, `git config user.email`, or OS username if omitted).
- `--password string` — Owner password to provision (defaults to `GARUDA_PASSWORD`; if unset, a random password is generated and **not** shown, so set one to be able to log in).
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
- `./bin/garuda summary [--workspace <name>] [--json]` — Executive architectural summary of repositories, entities, relationships, top hubs, and Merkle ledger height.
- `./bin/garuda topology` — Package-level layering, fan-in/fan-out, and circular dependency report.
- `./bin/garuda impact --workspace <ws-uuid> --target <entity-uuid> --depth 3` — Confidence-weighted transitive blast radius.
- `./bin/garuda hygiene [path]` (alias `garuda ponytail`) — Advisory static report for unreferenced symbols and duplicate candidates.

---

### `garuda dev` / `garuda up` / `garuda status`
- `./bin/garuda dev` — Starts the unified Garuda daemon (HTTP API server + 10-surface `/dashboard` Software Intelligence Instrument + 10-second background runtime verification & Merkle epoch worker).
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
| `language_matches` | `language` (`go`, `python`, `typescript`, `java`) | Scopes evaluation to repositories/entities of the target language. |

---

## 4. Active MCP Surface (`bin/garuda-mcp`)

`bin/garuda-mcp` speaks JSON-RPC 2.0 over stdio (`protocolVersion: "2025-06-18"`), advertises `tools`, `prompts`, and `resources` capabilities, and exposes **54 deterministic tools**:

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
Expected result: `═══ 25/25 checks passed ═══` (self-hosted dynamic verifier) or `═══ 63/63 checks passed ═══` (multi-repo benchmark suite).

---

## 5. Key HTTP Gateway & PWA Endpoints (`bin/garuda-api`)

All `/api/v1/*` routes enforce cryptographic Ed25519 JWT validation (`AuthMiddleware`) or authenticated session cookies (`garuda_session`).

| Method & Path | Auth | Description |
| :--- | :--- | :--- |
| `GET /health`, `GET /healthz`, `GET /readyz` | Public | Gateway liveness and PostgreSQL readiness checks |
| `GET /manifest.json`, `GET /sw.js`, `GET /icon.svg` | Public | Progressive Web App (PWA) manifest, service worker, and vector icon for Chromium/Android/Desktop installation |
| `GET /signup`, `POST /signup` | Public | Browser & CLI/API account + tenant + workspace provisioning |
| `GET /login`, `POST /login` | Public | Session authentication (`garuda_session` cookie) |
| `GET /dashboard` | Session | 10-surface Software Intelligence Instrument (`Briefing`, `Map`, `Workbench`, `Policies`, `Contradictions`, `Ledger`) |
| `GET /api/v1/dashboard/instrument` | Session / Bearer | Complete JSON payload powering the 10-surface dashboard instrument |
| `GET /api/v1/dashboard/neighborhood` | Session / Bearer | 1-hop / 2-hop focal symbol neighborhood graph & 4-tab Inspector Drawer payload |
| `POST /api/v1/dashboard/workbench/simulate` | Session / Bearer | Live Change Review Workbench pre-merge policy & blast-radius simulator |
| `GET /graph`, `GET /api/v1/graph` | Session / Bearer | Interactive D3 force graph & JSON graph payload |
| `POST /api/v1/telemetry/spans` | Bearer / Session | OpenTelemetry runtime span ingestion & synchronous edge-level contradiction detection |
| `GET /api/v1/evidence/verify/{id}` | Bearer / Session | RFC 6962 cryptographic Merkle inclusion proof verification for a decision |
| `GET /api/v1/merkle/state` | Bearer / Session | Current tenant Merkle root hash and epoch block height |
| `GET /admin` | Control Token | Multi-tenant operator Control Plane dashboard |


---

## 6. Build Matrix Flags, Dashboard Lineage Endpoints & BuildContext Schema

### 6.1 Cross-Platform Build Matrix & Build-Tag Flags (`garuda analyze`, `garuda check`, `garuda talon guard`)
- `--tags string` — Comma-separated Go build tags (e.g., `--tags cgo,enterprise`) applied via `go/build.Context` during conditional AST file matching.
- `--target-os string` — Override target `GOOS` (`linux`, `darwin`, `windows`) for conditional compilation analysis.
- `--target-arch string` — Override target `GOARCH` (`amd64`, `arm64`) for conditional compilation analysis.
- `--matrix` — Execute the 4-profile cross-platform build matrix (`linux/amd64 cgo=0`, `linux/amd64 cgo=1 tags=cgo`, `darwin/arm64 cgo=1`, `windows/amd64 cgo=0`) in a single pass and report per-profile active/skipped files, graph fingerprints, and policy verdicts.

### 6.2 Docs-to-Code Lineage & Talon Risk Hub HTTP Endpoints (`bin/garuda-api`)

| Method & Path | Description |
| :--- | :--- |
| `GET /api/v1/dashboard/docs-graph?workspace=<ws>` | Returns the 4-tier Docs, Specs & Guardrails lineage graph (`documents`, `rules`, `guardrails`, `code_nodes`, `edges`, `component_matrix`, and `summary` counts). |
| `POST /api/v1/dashboard/docs-ingest` | Ingests a new Markdown ADR/spec payload (`title`, `path`, `doc_type`, `content`) or triggers a repository directory scan, extracts normative `MUST`/`MUST_NOT` rules, and links them to matching workspace code symbols. |
| `GET /api/v1/dashboard/talon-hub?workspace=<ws>` | Returns the active `Garuda Talon` contract, 6 IDE/Git hook statuses, top high-blast-radius `risk_hubs` (0–100 risk score), the active `build_context` receipt, `environment_drift` report, and 4-target `build_matrix` results. |
| `POST /api/v1/dashboard/talon-check` | Executes a live diff-scoped `talon guard` check (`mode: "guard"`), simulates a layer violation with self-repair output (`mode: "simulate"`), or re-syncs all 6 IDE/Git hooks (`mode: "init"`). |

### 6.3 Deterministic `BuildContext` Receipt JSON Schema (`internal/analyzer/build_context.go`)

```json
{
  "receipt_hash": "bc_9f82d4a1e7c03b12",
  "graph_fingerprint": "gf_4e19a82c01b7d3ef",
  "captured_at": "2026-10-27T07:15:00Z",
  "commit_sha": "df3f897248a7a0",
  "branch": "main",
  "worktree_dirty": false,
  "dirty_files_count": 0,
  "go_context": {
    "target_os": "linux",
    "target_arch": "amd64",
    "cgo_enabled": false,
    "build_tags": [],
    "compiler_version": "go1.24.0",
    "go_mod_hash": "sha256:8a4f...",
    "active_files_count": 142,
    "skipped_by_build_tags_count": 4
  },
  "reproduce_command": "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ./bin/garuda talon guard --workspace default --diff HEAD --fail-on-block"
}
```


---

## 7. Audit & Rulebook CLI Commands, MCP Tool Arguments & HTTP Endpoints

### 7.1 `garuda audit` & `garuda rulebook` CLI Commands (`cmd/garuda/audit_cmd.go`)
- `garuda audit logic [concept] --team <ops|marketing|management> [--decide <variant_id> --action APPROVED_BY_HUMAN] [--json]`
- `garuda audit data-logic --metric <name> --team-a-cr <float> --team-b-cr <float> [--reconcile <audit_id>] [--json]`
- `garuda audit criticality [--tier HIGH_SEVERE|MEDIUM|LOW] [--json]`
- `garuda rulebook generate [--output .garuda/DRAFT_BASE_RULEBOOK.md]`
- `garuda rulebook approve [--by "Stakeholder Name"]`

### 7.2 Audit, Rulebook & Docs MCP Tools (arguments)

| Tool Name | Whitelisted Arguments | Purpose |
| :--- | :--- | :--- |
| `garuda.disambiguate_logic` | `workspace`, `team`, `concept`, `tenant_id` | Returns competing team formulas with `0.0–1.0` weights, criticality tiers, and blast radius. |
| `garuda.decide_logic_variant` | `workspace`, `team`, `variant_id`, `decision_action`, `custom_weight`, `tenant_id` | Persists human stakeholder formula approval or freeze in PostgreSQL. |
| `garuda.audit_financial_discrepancy` | `workspace`, `metric`, `team_a_name`, `team_a_crore`, `team_b_name`, `team_b_crore`, `tenant_id` | Audits cross-team Crore/INR discrepancies and traces AST/SQL unit-scaling root causes. |
| `garuda.reconcile_financial_audit` | `workspace`, `audit_id`, `metric`, `approved_by`, `tenant_id` | Marks a flagged financial discrepancy as reconciled by a human auditor. |
| `garuda.synthesize_rulebook` | `workspace`, `tenant_id` | Generates `.garuda/DRAFT_BASE_RULEBOOK.md` and persists draft to PostgreSQL. |
| `garuda.approve_base_rulebook` | `workspace`, `approved_by`, `tenant_id` | Seals `.garuda/APPROVED_BASE_RULEBOOK.md` with an RFC 6962 SHA-256 digest. |
| `garuda.criticality_tiers` | `workspace`, `tier_filter`, `surface_name`, `package_pattern`, `criticality_tier`, `tenant_id` | Lists or registers `HIGH_SEVERE`, `MEDIUM`, and `LOW` architectural surfaces. |
| `garuda.docs_graph` | `workspace`, `doc_type`, `tenant_id` | Returns the 4-tier Document-to-Code & Guardrails Lineage Matrix. |
| `garuda.docs_ingest` | `workspace`, `path`, `title`, `doc_type`, `tenant_id` | Ingests Markdown ADRs/specs into `document_claims` and correlates with AST entities. |
| `garuda.doctor_diagnose` | `workspace`, `fix`, `tenant_id` | Executes the System Doctor check over MCP. |
| `garuda.build_matrix_check` | `workspace`, `repo_dir`, `diff`, `tenant_id` | Evaluates the 4-target Cross-Platform Build Matrix and returns the `BuildContext` receipt. |
| `garuda.hygiene_report` | `workspace`, `limit`, `tenant_id` | Runs the static code hygiene analyzer over MCP. |

### 7.3 Audit & Rulebook HTTP Endpoints (`internal/api/dashboard_audit_rulebook.go`)
- `GET  /api/v1/dashboard/logic-weights?workspace=<ws>&team=<ops|marketing|management>`
- `POST /api/v1/dashboard/logic-weights/decide`
- `GET  /api/v1/dashboard/criticality-tiers?workspace=<ws>`
- `POST /api/v1/dashboard/criticality-tiers`
- `GET  /api/v1/dashboard/financial-audit?workspace=<ws>`
- `POST /api/v1/dashboard/financial-audit`
- `GET  /api/v1/dashboard/rulebook-draft?workspace=<ws>`
- `POST /api/v1/dashboard/rulebook-approve?workspace=<ws>`


---

## 8. Dirty-Worktree Provenance Flags, MCP Security Controls & Additional MCP Tools

### 8.1 Dirty-Worktree Provenance Flags on `garuda check` & `garuda talon guard`
When a working tree has uncommitted edits, `BuildContext.ReproduceCommand` emits `--worktree-dirty` and `--dirty-tree-hash <12-char-sha256>` so CI or another developer can verify the exact dirty-tree state:

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 garuda check \
  --target linux/amd64 \
  --cgo=false \
  --commit ef749a162e01 \
  --worktree-dirty \
  --dirty-tree-hash 417801f9a311 \
  --workspace default
```

- `--worktree-dirty` (`bool`) — Assert that the working tree is expected to be dirty when verifying the `BuildContext` receipt.
- `--dirty-tree-hash string` — Assert the exact 12-character SHA-256 digest of modified/untracked source files (`BuildContext.DirtyFilesHash`).

### 8.2 MCP Security & Attestation Environment Variables
- `GARUDA_MCP_READ_ONLY=true` — Blocks all 13 mutating MCP tools (`propose_decision`, `decide_logic_variant`, `reconcile_data_audit`, `reconcile_financial_audit`, `approve_base_rulebook`, `synthesize_rulebook`, `docs_ingest`, `create_task`, `handoff`, `resume`, `reset_audit_state`, `create_workspace`, and `talon_verify_edit` when `save_to_ledger=true`), returning a structured `{"status": "rejected", "read_only": true}` response.
- `GARUDA_MCP_APPROVER=<email-or-handle>` — Required for an MCP call to record `APPROVED_BY_HUMAN` or `RECONCILED_BY_HUMAN`. When unset, `decide_logic_variant`, `reconcile_data_audit`, and `approve_base_rulebook` record `PROPOSED_BY_AGENT_AWAITING_HUMAN_CLI_COSIGN` (`human_cli_verified: false`) until co-signed via the CLI or Web Dashboard.

### 8.3 Additional MCP Tools (arguments)
| Tool Name | Whitelisted Arguments | Purpose |
| :--- | :--- | :--- |
| `garuda.preflight_bundle` | `workspace`, `target`, `symbol`, `file`, `tenant_id` | Single-call pre-edit bundle combining `talon_pre_edit`, `blast_radius`, `neighbors`, `criticality_tiers`, `query_claims`, and `snapshot_receipt`. |
| `garuda.verify_receipt` | `workspace`, `receipt_hash`, `graph_fingerprint`, `commit_sha`, `tenant_id` | Cryptographically verifies a `snapshot_receipt` or `BuildContext` hash against the live PostgreSQL workspace and RFC 6962 Merkle root. |
| `garuda.diff_snapshots` | `workspace`, `from_revision`, `to_revision`, `tenant_id` | Diffs two workspace snapshots or git revisions, returning added/removed files, entities, claims, and policy posture delta. |
| `garuda.audit_export` | `workspace`, `format` (`csv`/`json`/`md`), `limit`, `tenant_id` | Exports the tamper-evident governance, policy, logic-variant, and data-audit ledger in CSV, JSON, or Markdown format. |
| `garuda.create_workspace` | `workspace`, `description`, `tenant_id` | Provisions an isolated workspace (returning an isolated empty `snapshot_receipt` until repositories are analyzed into it). |
| `garuda.audit_data_discrepancy` | `workspace`, `metric_name`, `team_a_name`, `team_a_crore`, `team_b_name`, `team_b_crore`, `tenant_id` | FinOps-safe alias for cross-team metric & unit-scale discrepancy auditing (`data_logic_audit_ledger`). |
| `garuda.reconcile_data_audit` | `workspace`, `audit_id`, `approved_by`, `resolution_note`, `tenant_id` | Reconciles a flagged data/logic discrepancy (requiring `GARUDA_MCP_APPROVER` for human sign-off). |
| `garuda.reset_audit_state` | `workspace`, `confirm`, `tenant_id` | Resets demo/sandbox logic-variant and data-audit overrides back to clean baseline state. |

---

<!-- MCP-TOOLS:BEGIN (generated by scripts/sync_docs_from_manifest.py; do not edit) -->
## The 54 MCP tools

41 are read-only and 13 can change state (disabled when `GARUDA_MCP_READ_ONLY=true`).

| # | Tool | Mode | What it does |
| :--- | :--- | :--- | :--- |
| 01 | `garuda.briefing` | Read-only | Session-start briefing: workspace state, policy verdict, epistemic counts, trust anchor, scale, top hubs, active policies, open contradictions, and what changed since this agent's last briefing. Read-only. Call this first. |
| 02 | `garuda.query` | Read-only | Query the Garuda knowledge graph (entities, document claims, and decisions) with natural language keywords. Rejects raw SQL statements. |
| 03 | `garuda.get_lineage` | Read-only | Get the full governance and AST lineage of a decision UUID, actor name (e.g. 'cli-operator'), or symbol name. |
| 04 | `garuda.detect_contradictions` | Read-only | Detect unresolved contradictions across decisions and ADR document claims in the tenant knowledge graph. |
| 05 | `garuda.get_impact` | Read-only | Find what breaks if a decision UUID or code symbol (e.g. 'AppendLeafAndSealTx') is changed. Computes transitive blast radius and child decision impact. |
| 06 | `garuda.blast_radius` | Read-only | Compute the blast radius of a symbol change: which entities consume the target entity, directly or transitively, up to a depth. Read-only. Returns a critical-first list and per-severity counts. |
| 07 | `garuda.propose_decision` | Mutating (`GARUDA_MCP_READ_ONLY` gated) | Propose a new architectural decision (creates a draft). |
| 08 | `garuda.policy.list` | Read-only | List policies registered for a tenant. Read-only. |
| 09 | `garuda.governance.status` | Read-only | Aggregate governance state: top-level decision, active policies, documentation health, blocked_by list, and unified entity/claim counts. Read-only. |
| 10 | `garuda.check_drift` | Read-only | Bidirectional documentation-to-code (doc_to_code) and code-to-documentation (code_to_doc) drift report. Read-only. |
| 11 | `garuda.query_claims` | Read-only | Query document claims matching a subject and/or epistemic status (SUPPORTED, UNVERIFIED, CONTRADICTED). Read-only. |
| 12 | `garuda.policy.evaluate` | Read-only | Dry-run the policy engine against a workspace and return deterministic policy decisions. Does NOT persist or Merkle-anchor. Use 'garuda policy evaluate' from the CLI to anchor a real governance decision. |
| 13 | `garuda.verify_policy_evaluation` | Read-only | Verify the Merkle inclusion proof for a persisted policy evaluation. Returns valid=true if the proof recomputes to the anchored root. Read-only; no side effects. |
| 14 | `garuda.verify_evidence` | Read-only | Verify the RFC 6962 Merkle inclusion proof or SHA-256 evidence digest for a policy evaluation, evidence block, or decision UUID. Read-only; no side effects. |
| 15 | `garuda.entities` | Read-only | List semantic entities in a workspace. Filterable by package and kind. Read-only. |
| 16 | `garuda.inspect` | Read-only | Inspect one entity with its signature and incoming and outgoing relationships. Read-only. |
| 17 | `garuda.subclasses` | Read-only | List entities that inherit from or embed the subject. Uses INHERITS and EMBEDS claims. Read-only. |
| 18 | `garuda.implementers` | Read-only | List entities that implement the subject interface. Uses IMPLEMENTS claims. Read-only. |
| 19 | `garuda.neighbors` | Read-only | List every entity connected to the subject in one hop, both incoming and outgoing directions. External stubs excluded by default. Read-only. |
| 20 | `garuda.create_task` | Mutating (`GARUDA_MCP_READ_ONLY` gated) | Create a new task and provision source/target agent UUIDs in the workspace so garuda.handoff and garuda.resume can operate on persisted UUIDs. |
| 21 | `garuda.handoff` | Mutating (`GARUDA_MCP_READ_ONLY` gated) | Initiate an atomic task handoff between two agents. Creates a checkpoint of the source agent's state, transfers task ownership, and transitions both agents' statuses. Call garuda.create_task first to provision UUIDs. |
| 22 | `garuda.resume` | Mutating (`GARUDA_MCP_READ_ONLY` gated) | Restore the state from an active checkpoint created by garuda.handoff, and mark the checkpoint as consumed. A second resume of the same checkpoint returns {status: 'not_found'}. |
| 23 | `garuda.find_entity` | Read-only | Find entities by exact name, regex name_pattern, kind, package, or file path. Returns status='not_found' with empty results when a filter matches 0 entities. Read-only. |
| 24 | `garuda.talon_pre_edit` | Read-only | MANDATORY pre-edit check: call before modifying any file or symbol. Returns transitive blast radius, callers, applicable ADR MUST/MUST_NOT claims, and active policy guardrails. |
| 25 | `garuda.talon_verify_edit` | Mutating (`GARUDA_MCP_READ_ONLY` gated) | MANDATORY post-edit verification gate: call immediately after modifying code. Runs in dry-run mode (side_effects=[]) by default; set save_to_ledger=true to append a Merkle ledger block (side_effects=['merkle_ledger_append']). |
| 26 | `garuda.disambiguate_logic` | Read-only | Compare competing undocumented formulas across teams with 0.0-1.0 team weights summing to 1.00, surface criticality tiers (HIGH_SEVERE, MEDIUM, LOW), and explicit mode ('live_ast' vs 'fixture'). |
| 27 | `garuda.audit_data_discrepancy` | Read-only | Audit cross-team business logic and metric number discrepancies in the Data & Logic Audit Ledger by tracing AST and SQL lineage before production deployment. |
| 28 | `garuda.reconcile_data_audit` | Mutating (`GARUDA_MCP_READ_ONLY` gated) | Record a reconciliation proposal for a flagged cross-team data discrepancy in the Data & Logic Audit Ledger. Requires GARUDA_MCP_APPROVER or CLI co-signing for human approval. |
| 29 | `garuda.audit_financial_discrepancy` | Read-only | Legacy alias for garuda.audit_data_discrepancy: audit cross-team business metric number discrepancies in the Data & Logic Audit Ledger. |
| 30 | `garuda.synthesize_rulebook` | Mutating (`GARUDA_MCP_READ_ONLY` gated) | Synthesize unorganized repository architecture, surface criticality tiers, and multi-team formula weights into a structured draft Base Rulebook for human stakeholder review and approval. |
| 31 | `garuda.decide_logic_variant` | Mutating (`GARUDA_MCP_READ_ONLY` gated) | Propose or record a decision on a competing multi-team formula variant in PostgreSQL. Requires GARUDA_MCP_APPROVER or CLI co-signing for APPROVED_BY_HUMAN. |
| 32 | `garuda.reconcile_financial_audit` | Mutating (`GARUDA_MCP_READ_ONLY` gated) | Legacy alias for garuda.reconcile_data_audit: record reconciliation for a flagged cross-team metric discrepancy in the Data & Logic Audit Ledger. |
| 33 | `garuda.approve_base_rulebook` | Mutating (`GARUDA_MCP_READ_ONLY` gated) | Propose or approve the synthesized Base Rulebook with an RFC 6962 SHA-256 digest. Requires GARUDA_MCP_APPROVER or CLI co-signing for human approval. |
| 34 | `garuda.criticality_tiers` | Read-only | List or register 3-Tier Surface Criticality classifications (HIGH_SEVERE freeze, MEDIUM trade-off, LOW UI box alignment). |
| 35 | `garuda.docs_graph` | Read-only | Retrieve the 4-Tier Document-to-Code & Guardrails Lineage Matrix (Documents -> Normative Rules -> Active Policies -> AST Symbols). |
| 36 | `garuda.docs_ingest` | Mutating (`GARUDA_MCP_READ_ONLY` gated) | Ingest Markdown ADRs or design specifications into document_claims and correlate them against workspace AST symbols. |
| 37 | `garuda.doctor_diagnose` | Read-only | Run the Garuda System Doctor diagnostic suite across platform, PostgreSQL, Talon contract, and API gateway health. |
| 38 | `garuda.build_matrix_check` | Read-only | Verify the codebase across the 4-target Cross-Platform Build Matrix (linux/amd64, darwin/arm64, windows/amd64, cgo/purego) and return the deterministic BuildContext receipt. |
| 39 | `garuda.hygiene_report` | Read-only | Run the static code hygiene analyzer on the workspace graph to report advisory unreferenced candidates, duplicate symbol names, and stdlib alternatives with structured file:line locations. |
| 40 | `garuda.reset_audit_state` | Mutating (`GARUDA_MCP_READ_ONLY` gated) | Reset smoke-test or agent-proposed audit state in the workspace back to clean baseline. |
| 41 | `garuda.create_workspace` | Mutating (`GARUDA_MCP_READ_ONLY` gated) | Provision an isolated sandbox workspace (e.g. 'audit-sandbox') so mutating governance or audit checks can run safely without affecting the default workspace. |
| 42 | `garuda.verify_receipt` | Read-only | Recompute receipt_hash, graph_fingerprint, and policy_set_version from the current AST and git commit to cryptographically verify a pre-edit or guard receipt. |
| 43 | `garuda.diff_snapshots` | Read-only | Compare the indexed graph snapshot receipt against current HEAD and list active vs superseded policies, commit sync state, and drift deltas. |
| 44 | `garuda.diff_impact` | Read-only | Compute diff-level blast radius in one call: returns per-symbol impact with file, line_start, and source snippets for every symbol touched in the git diff. |
| 45 | `garuda.symbol_test_coverage` | Read-only | Show which unit tests (*_test.go) and runtime traces exercise a symbol or package, explaining exported handler verification status. |
| 46 | `garuda.snapshot` | Read-only | Return the current workspace snapshot ID, commit SHA, dirty-tree status, policy-set version, and freshness verdict (fresh, stale, dirty) in one lightweight call. |
| 47 | `garuda.explain_claim` | Read-only | Return the compiler AST or policy derivation path and source snippet for a claim_id or symbol name. |
| 48 | `garuda.snippet` | Read-only | Return hash-verified source lines around a symbol or file_path:line_start against the active snapshot. |
| 49 | `garuda.summary` | Read-only | Return a compact token-saving workspace summary (counts, active policies, and snapshot ID). |
| 50 | `garuda.owners` | Read-only | Return the declared owner and team for an entity from the ownership mapping, or UNASSIGNED if undeclared. |
| 51 | `garuda.verify_claim` | Read-only | Verify a structured edge (from_symbol, predicate, to_symbol) or prose claim against the indexed graph and return SUPPORTED, CONTRADICTED, or NOT_IN_GRAPH with a cryptographic receipt. |
| 52 | `garuda.test_coverage` | Read-only | Alias for garuda.symbol_test_coverage: inspect which unit tests and runtime traces exercise a symbol. |
| 53 | `garuda.audit_export` | Read-only | Export workspace decisions, policy evaluations, and Merkle block anchors as structured JSON or CSV for compliance filing. |
| 54 | `garuda.preflight_bundle` | Read-only | Single-call token-saving pre-edit bundle combining snapshot freshness, talon_pre_edit briefing, test coverage, and declared ownership for a symbol or file. |
<!-- MCP-TOOLS:END -->
