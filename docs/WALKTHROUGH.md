# Garuda Walkthrough (v1.2.0)

**What each interface looks like, what each view shows, and how to launch it in one command.**

This walkthrough covers the five surfaces you will use most often: the **1-Command Smart Setup (`./setup.sh` / `garuda init`)**, the **Unified Change-Control CLI (`garuda check`)**, the **Web Dashboard & Interactive Graph**, the **AI Agent MCP Integration**, and the **VS Code Extension**.

---

## 1. 1-Command Smart Setup (`./setup.sh` & `garuda init`)

Garuda v1.2 eliminates manual environment variable exports, manual SQL migration loops, and manual JWT key copy-pasting. Everything is discovered dynamically and persisted to `.env`:

```bash
./setup.sh
```

Or initialize any specific repository directly:

```bash
./bin/garuda init /path/to/your/repo
```

### What happens automatically
1. **Dynamic Identity & Workspace Discovery**: Reads your developer email and name from `git config user.email` / `git config user.name` (or `$GARUDA_EMAIL`) and derives the active workspace from `$GARUDA_WORKSPACE` or the target repository name.
2. **PostgreSQL & Idempotent Migrations**: Starts or creates the `garuda-postgres` Docker container if PostgreSQL is not running, and runs `store.Migrate` idempotently across all SQL migrations.
3. **Persistent `.env` & Ed25519 JWT Keys**: Generates and writes `DATABASE_URL`, `CONTROL_DATABASE_URL`, `GARUDA_WORKSPACE`, `PORT`, `JWT_PRIVATE_KEY_HEX`, and `JWT_PUBLIC_KEY_HEX` into `.env`, which `garuda`, `garuda-api`, and `garuda-mcp` load automatically on every startup.
4. **IDE & Agent MCP Configuration**: Writes `.cursor/mcp.json` and updates Claude Desktop (`~/.config/Claude/claude_desktop_config.json`) with your resolved binary path and workspace.

---

## 2. Unified Change-Control CLI (`garuda check`)

Before committing or merging an AI-generated change, run `garuda check` to evaluate all five Garuda planes (Software State + Blast Radius + 3-State Evidence + Policy Evaluation + RFC 6962 Merkle Seal) in a single command:

```bash
./bin/garuda check --project go-validation-10 --target SQLModel --policy ./policies --save --fail-on-block
```

```text
╔═══════════════════════════════════════════════════════════════════╗
║               GARUDA CHANGE-CONTROL DECISION                      ║
╚═══════════════════════════════════════════════════════════════════╝
  Change:      SQLModel (workspace: go-validation-10)
  Affected:    171 entities (CRITICAL=12, HIGH=48)
  Evidence:    SUPPORTED=142  |  UNVERIFIED=5  |  CONTRADICTED=0
  Policies:    3 evaluated in 38ms
  Merkle Seal: RFC 6962 block #7
───
  GOVERNANCE OUTCOME: WARN
```

Add `--json` when running inside CI/CD pipelines to emit a machine-readable JSON decision payload.

---

## 3. The Web Dashboard (`/dashboard`)

The web dashboard displays everything about a workspace in one view: repositories, entities, relationships, policies, documentation health, runtime drift, and the Merkle ledger.

**Open it:**

```bash
./bin/garuda dev
```

Then visit `http://localhost:8080/login` (sign in with the owner email and password provisioned by `./setup.sh` or `garuda init`) and open `http://localhost:8080/dashboard`.

### What you see
- **Sidebar & Workspace Picker**: Switch between any workspace you own or belong to (`workspace_members`), including workspaces created from the CLI via `garuda analyze --save`.
- **Overview KPIs**: Repositories, Packages, Entities, Typed Relationships, Active Policies, and Drift Prevented.
- **Policy Enforcement Panel**: Counts of `BLOCK`, `REVIEW`, `WARN`, and `ALLOW` evaluations, plus the ten most recent evaluations with their RFC 6962 Merkle block heights and clickable evidence drawers.
- **Trust Strip (3-State Epistemic Model)**: Displays `SUPPORTED`, `UNVERIFIED`, and `CONTRADICTED` claim counts. When runtime verification has not yet been ingested for an entity, it remains honestly marked `UNVERIFIED`.
- **Architectural Hubs**: The highest-impact entities in the workspace ranked by incoming caller/dependent count.
- **Needs Attention (Quarantined Contradictions)**: Lists any runtime observation that contradicted the static architecture graph (including undeclared runtime calls between services).

---

## 4. The Interactive Force Graph (`/graph`)

Visit `http://localhost:8080/graph` while `garuda dev` or `garuda-api` is running.

- **Nodes**: Repositories, packages, or entities scaled by inbound centrality so core architectural hubs stand out visually.
- **Edges**: Solid lines represent static AST relationships (`CALLS`, `IMPORTS`, `IMPLEMENTS`, `INHERITS`, `EMBEDS`); red dashed lines highlight `CONTRADICTED` runtime drift edges.
- **Navigation**: Single-click a node to inspect its source file, line numbers, and callers; double-click to drill down from repository to package to entity level.

---

## 5. AI Agent MCP Integration (`bin/garuda-mcp`)

Garuda speaks JSON-RPC 2.0 over stdio (`protocolVersion: "2025-06-18"`). Any MCP client (Cursor, Claude Desktop, Codex CLI, or VS Code) can query the workspace directly.

### The 21+ MCP Tools
- **Session Context**: `garuda.briefing`
- **Graph & Architecture Exploration**: `garuda.find_entity`, `garuda.entities`, `garuda.neighbors`, `garuda.subclasses`, `garuda.implementers`, `garuda.blast_radius`, `garuda.inspect`, `garuda.query_claims`, `garuda.query`
- **Governance & Cryptographic Verification**: `garuda.policy.list`, `garuda.policy.evaluate`, `garuda.verify_policy_evaluation`, `garuda.governance.status`, `garuda.check_drift`, `garuda.get_lineage`, `garuda.get_impact`, `garuda.detect_contradictions`, `garuda.propose_decision`
- **Multi-Agent Coordination**: `garuda.handoff`, `garuda.resume`

### Verify MCP Compliance Dynamically
`scripts/mcp_verify.py` automatically loads `.env` and dynamically discovers the top hub and interface entities in your active workspace (no hardcoded repository names or edge counts):

```bash
python3 scripts/mcp_verify.py
```

Expected output:
```text
═══ 23/23 checks passed ═══
```

---

## 6. The VS Code / Cursor Extension (`vscode-extension/`)

Build and install the extension from `vscode-extension/`:

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
- **[README.md](../README.md)** — High-level overview and 1-command quickstart.
- **[PLAYBOOK.md](../PLAYBOOK.md)** — Full operational runbook and troubleshooting.
- **[docs/CAPABILITIES.md](CAPABILITIES.md)** — 5-plane capability and verification matrix.
- **[docs/API.md](API.md)** — Complete CLI, Policy YAML, MCP, and HTTP API reference.
- **[docs/SPECS.md](SPECS.md)** — Formal system specifications and boundaries.
