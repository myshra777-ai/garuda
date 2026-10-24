# Garuda CLI, Policy, MCP & HTTP API Reference (v1.2.0)

This reference documents the complete command-line interface (`bin/garuda`), declarative policy specification (`policies/*.yaml`), Model Context Protocol surface (`bin/garuda-mcp`), and HTTP REST gateway (`bin/garuda-api`).

---

## 1. Environment & Zero-Config `.env` Resolution

All Garuda binaries (`garuda`, `garuda-api`, `garuda-mcp`) automatically load `.env` from the current working directory or repository root on startup via `internal/config/env.go`.

| Variable | Default (Dev) | Purpose |
| :--- | :--- | :--- |
| `DATABASE_URL` | `postgres://test:test@localhost:5433/garuda_test?sslmode=disable` | Primary PostgreSQL connection string |
| `GARUDA_WORKSPACE` | Most recently updated workspace | Default workspace name for CLI and MCP commands |
| `PORT` | `8080` | HTTP port for `garuda-api` and `garuda dev` |
| `GARUDA_API_URL` | `http://localhost:${PORT}` | Base URL used by `garuda` CLI when calling the HTTP gateway |
| `JWT_PRIVATE_KEY_HEX` | Auto-generated & persisted to `.env` | Ed25519 64-byte private key (hex) for JWT signing |
| `JWT_PUBLIC_KEY_HEX` | Auto-generated & persisted to `.env` | Ed25519 32-byte public key (hex) for JWT verification |
| `CONTROL_DATABASE_URL` | Defaults to `DATABASE_URL` when token set | Read-only connection pool for the `/admin` Control Plane |
| `GARUDA_CONTROL_TOKEN` | `devtoken-changeme-please` | Operator bearer token for `/admin` Control Plane routes |

---

## 2. CLI Commands (`bin/garuda`)

### `garuda init [target_repo_path]`
Initializes Garuda in a single command: verifies or starts the `garuda-postgres` Docker container, runs all database migrations idempotently, persists `.env` and Ed25519 JWT keys, provisions the default workspace and owner account, configures Cursor (`.cursor/mcp.json`) and Claude Desktop MCP, and indexes the target repository.

```bash
./bin/garuda init . --workspace go-validation-10 --email myshra777@gmail.com --password garuda2026
```

**Flags:**
- `--workspace string` — Workspace name to initialize (defaults to `GARUDA_WORKSPACE` or `go-validation-10`).
- `--email string` — Owner email to provision (default `myshra777@gmail.com`).
- `--password string` — Owner password to provision (default `garuda2026`).
- `--target string` — Repository path to index (default `.`).
- `--skip-scan` — Skip initial repository AST scan during setup.

---

### `garuda check [dir]`
Executes the unified 5-plane Change-Control workflow (Blast Radius + 3-State Evidence + Policy Evaluation + Governance Verdict + RFC 6962 Merkle Seal) and prints the Change-Control Decision Card.

```bash
./bin/garuda check --project go-validation-10 --target SQLModel --policy ./policies --save --fail-on-block
```

**Flags:**
- `--workspace`, `-w`, `--project string` — Workspace name.
- `--symbol`, `--target string` — Target symbol or service name being modified.
- `--policies`, `--policy string` — Directory containing policy YAML files (default `./policies`).
- `--save`, `-s` — Persist evaluations and seal RFC 6962 Merkle receipt (default `true`).
- `--fail-on-block` — Exit with code `2` when governance outcome is `BLOCK`.
- `--json` — Emit structured JSON decision payload.

---

### `garuda analyze <path>`
Analyzes a Go, Python, or TypeScript repository, extracts entities and typed relationships (`CALLS`, `IMPORTS`, `IMPLEMENTS`, `INHERITS`, `EMBEDS`, `REFERENCES`), and optionally persists the semantic graph and Merkle decision revision to PostgreSQL.

```bash
./bin/garuda analyze ~/garuda-validation/go/chi --workspace go-validation-10 --save
./bin/garuda analyze ~/garuda-validation/py/sqlmodel --workspace go-validation-10 --save
```

**Flags:**
- `--save`, `-s` — Persist extracted entities, claims, and Merkle revision to PostgreSQL.
- `--workspace`, `-w string` — Target workspace name.
- `--output`, `-o string` — Write JSON snapshot to file.
- `--json` — Print JSON snapshot to stdout.
- `--diff <before.json> <after.json>` — Compute structural schema and breaking-change diff between two snapshots.
- `--fail-on-breaking` — Exit non-zero if breaking changes are detected during `--diff`.

---

### `garuda policy <subcommand>`
Deterministic policy enforcement lifecycle commands:

```bash
./bin/garuda policy validate ./policies
./bin/garuda policy evaluate ./policies --workspace go-validation-10 --save --fail-on-block
./bin/garuda policy list
./bin/garuda policy show <evaluation-id>
./bin/garuda policy verify <evaluation-id>
```

---

### `garuda summary`
Prints an executive architectural summary of the active workspace, including total repositories, AST entities, typed relationships, top architectural hubs ranked by incoming callers, and current Merkle ledger block status.

```bash
./bin/garuda summary
```

---

### `garuda impact`
Computes the confidence-weighted transitive blast radius (`CRITICAL`, `HIGH`, `MEDIUM`, `LOW`) for a target entity UUID across the workspace graph.

```bash
./bin/garuda impact --workspace <workspace-uuid> --target <entity-uuid> --depth 3 --min-confidence 0.50 --json
```

---

### `garuda docs <ingest|verify|status>`
Extracts normative architectural claims from Markdown ADRs and specifications and verifies them against the static codebase graph.

```bash
./bin/garuda docs ingest docs/adr/0001-payment-idempotency.md
./bin/garuda docs verify
./bin/garuda docs status
```

---

### `garuda dev` / `garuda up` / `garuda status`
- `./bin/garuda dev` — Starts the unified Garuda daemon (HTTP API server + D3 graph visualizer + 10-second background runtime verification & Merkle epoch worker).
- `./bin/garuda up` — Starts PostgreSQL container, runs migrations, and launches the gateway stack.
- `./bin/garuda status` — Verifies gateway health (`/health`).

---

## 3. Declarative Policy Predicates (`policies/*.yaml`)

Every policy file specifies `when` predicates evaluated against the workspace graph in PostgreSQL:

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

## 4. MCP Tools Surface (`bin/garuda-mcp`)

`bin/garuda-mcp` speaks JSON-RPC 2.0 over stdio (`protocolVersion: "2025-06-18"`) and exposes 21+ tools:

- **Graph & Architecture Discovery**: `garuda.find_entity`, `garuda.entities`, `garuda.neighbors`, `garuda.subclasses`, `garuda.implementers`, `garuda.blast_radius`, `garuda.briefing`, `garuda.summary`, `garuda.inspect`
- **Governance & Cryptographic Verification**: `garuda.verify_policy_evaluation`, `garuda.detect_contradictions`, `garuda.governance.status`, `garuda.decisions.active`
- **Multi-Agent Orchestration**: `garuda.handoff`, `garuda.resume`, `garuda.checkpoint`

Verify MCP compliance at any time by running:
```bash
python3 scripts/mcp_verify.py
```

---

## 5. Key HTTP Gateway Endpoints (`bin/garuda-api`)

| Method & Path | Auth | Description |
| :--- | :--- | :--- |
| `GET /health` | Public | Gateway & database liveness check |
| `GET /signup`, `POST /signup` | Public | Browser & CLI/API account + tenant + workspace provisioning (supports form-urlencoded and JSON) |
| `GET /login`, `POST /login` | Public | Session authentication (`garuda_session` cookie) |
| `GET /dashboard` | Session | Multi-workspace engineering intelligence & governance console |
| `GET /graph`, `GET /api/v1/graph` | Session / Bearer | Interactive D3 force graph & JSON graph payload |
| `POST /api/v1/telemetry/spans` | Header / Bearer | OpenTelemetry runtime span ingestion & synchronous contradiction detection |
| `GET /api/v1/evidence/verify/{id}` | Bearer | RFC 6962 cryptographic Merkle inclusion proof verification for a decision |
| `GET /api/v1/merkle/state` | Bearer | Current tenant Merkle root hash and epoch block height |
| `GET /admin` | Control Token | Multi-tenant operator Control Plane dashboard |