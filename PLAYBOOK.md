# Garuda Playbook

**Installation · Setup · Everyday Workflows · Agent Integration · Reference**

This playbook is the operational guide for running Garuda against a real codebase. It assumes you have read [`README.md`](README.md) and want exact commands, expected behavior, integration examples, and troubleshooting guidance for the current release.

> **Release contract:** Commands and behaviors documented here describe the current release. If observed behavior differs, record the command, environment, release commit, database configuration, and output before filing an issue.

---

## Table of Contents

1. [What Garuda Is For](#what-garuda-is-for)
2. [Architecture at a Glance](#architecture-at-a-glance)
3. [Before You Start](#before-you-start)
4. [Installation](#installation)
5. [Your First Workspace](#your-first-workspace)
6. [Analyzing Code](#analyzing-code)
7. [Connecting AI Clients](#connecting-ai-clients)
8. [Workspace Console and IDE](#workspace-console-and-ide)
9. [Working with Policies](#working-with-policies)
10. [Ingesting Documentation](#ingesting-documentation)
11. [Runtime Evidence and Contradictions](#runtime-evidence-and-contradictions)
12. [Verifying What You Built](#verifying-what-you-built)
13. [MCP Server Reference](#mcp-server-reference)
14. [Multi-Agent Coordination](#multi-agent-coordination)
15. [Running in CI](#running-in-ci)
16. [Troubleshooting](#troubleshooting)
17. [Command Reference](#command-reference)
18. [Evidence and Specifications](#evidence-and-specifications)

---

## What Garuda Is For

Garuda is a verified state and governance plane for AI-assisted software development. It builds a shared semantic model of code, relationships, documentation claims, runtime observations, policies, evidence, and decisions, then exposes that model through CLI, HTTP, MCP, dashboards, CI, and VS Code.

Garuda is primarily used to:

- Give AI agents grounded answers about what exists and how software is connected.
- Evaluate implementation state against organizational policies.
- Connect normative documentation to code and runtime evidence.
- Surface `SUPPORTED`, `UNVERIFIED`, and `CONTRADICTED` states.
- Record committed governance decisions in a cryptographically verifiable Merkle ledger.
- Analyze callers, implementers, dependencies, neighborhoods, and graph-visible impact.
- Coordinate multi-agent task handoffs through checkpoints and transactional resume operations.
- Display policy violations, contradictions, unverified claims, graph relationships, and evidence in the Workspace Console and IDE integration.

### Operating model

```text
Repositories ──analyze──▶ Semantic Graph ──query──▶ Developers / AI Agents
       │                         │
       └── docs ingest ──▶ Claims ──verify──▶ Drift and contradictions
                                  │
Runtime observations ──correlate─▶ Evidence and verification states
                                  │
Policies ──evaluate──▶ Decisions ──anchor──▶ Merkle Ledger ──verify──▶ Proof
```

Everything is local by default. No data leaves the machine unless you configure an external service, telemetry source, or integration.

Garuda is not an agent framework or coding agent. It is the shared semantic, evidence, and governance layer underneath those tools.

---

## Architecture at a Glance

| Layer | Responsibility | Primary interface |
|---|---|---|
| Repository analysis | Extract entities, relationships, and source evidence | `garuda analyze` |
| PostgreSQL model | Persist tenant, workspace, repository, graph, claims, runtime, and governance state | `DATABASE_URL` |
| Documentation claims | Extract normative requirements from supported documents | `garuda docs ingest` |
| Verification | Correlate claims with code and runtime evidence | `garuda docs verify` |
| Runtime evidence | Ingest observations and identify supported or contradictory behavior | Runtime API and telemetry paths |
| Policy engine | Evaluate rules and produce `ALLOW`, `WARN`, `REVIEW`, or `BLOCK` | `garuda policy evaluate` |
| Trust layer | Anchor committed evaluations and decisions in a Merkle ledger | `garuda status`, `garuda verify`, `garuda policy verify` |
| AI integration | Expose graph and governance tools over MCP | `garuda-mcp` |
| Agent coordination | Transfer tasks and consume checkpoints transactionally | `garuda.handoff`, `garuda.resume` |
| Workspace Console | Present workspace, agent, governance, decision, and graph views | `garuda dev` |
| IDE integration | Present ledger, contradictions, diagnostics, graph, and hover impact | `vscode-extension/` |

---

## Before You Start

### Requirements

| Requirement | Minimum or supported option |
|---|---|
| PostgreSQL | Version 14 or later |
| Go | Version 1.26 or later |
| Codebase | Go for compiler-backed analysis; Python and TypeScript are also supported structurally |
| Shell | Bash, Zsh, or equivalent |
| Node.js/npm | Required to compile the VS Code extension |
| C compiler | Required for TypeScript analyzer support |
| Editor | Any text editor; VS Code for the extension workflow |

You do not need:

- A cloud account for local operation.
- An external API key for the core local workflow.
- A license for the open-source local workflow.

### Analyzer support

| Language | Detection | Analysis mode | Maturity |
|---|---|---|---|
| Go | `go.mod` | Compiler-backed and validated | Production within supported scope |
| Python | `pyproject.toml` or `setup.py` | Structural | Beta |
| TypeScript | `tsconfig.json` or `package.json` | Structural; requires CGO and a C compiler | Beta |
| Rust | — | Planned | Upcoming |
| Java | — | Planned | Upcoming |
| Elixir | — | Planned | Upcoming |

---

## Installation

### Build Garuda from source

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda

go build -o bin/garuda ./cmd/garuda
go build -o bin/garuda-mcp ./cmd/garuda-mcp
```

For TypeScript analysis:

```bash
export CGO_ENABLED=1
```

Ensure a C compiler is available on `PATH`.

Rebuild `bin/garuda-mcp` whenever changes affect `cmd/garuda-mcp/`, MCP tool definitions, or governance code used by the server. MCP clients spawn the binary independently and will not detect a stale build until restarted.

### Start PostgreSQL with Docker

```bash
docker run -d \
  --name garuda-pg \
  -e POSTGRES_USER=garuda \
  -e POSTGRES_PASSWORD=garuda \
  -e POSTGRES_DB=garuda \
  -p 5432:5432 \
  postgres:16
```

Export the connection string in every shell that uses Garuda:

```bash
export DATABASE_URL="postgres://garuda:garuda@localhost:5432/garuda?sslmode=disable"
```

The URL must match the database container you started. If multiple PostgreSQL containers exist, Garuda uses whichever database is specified by the exported `DATABASE_URL`.

### Apply migrations

```bash
for f in migrations/[0-9]*.sql; do
  psql "$DATABASE_URL" -f "$f"
done
```

Migrations are intended to be idempotent. When upgrading, apply them in release order. If the migration directory contains duplicate numeric prefixes, inspect the filenames and verify ordering before applying them automatically.

### Compile the VS Code extension

```bash
cd vscode-extension
npm install
npm run compile
cd ..
```

The extension entry point is `vscode-extension/out/extension.js`.

### Verify the installation

```bash
./bin/garuda --version
./bin/garuda status
```

`garuda status` reports database reachability, Merkle block height, root hash, verification distribution, and daemon state.

---

## Your First Workspace

A workspace is a logical group of repositories and semantic state. Use one workspace per product, domain, or team.

### Create and select a workspace

```bash
export GARUDA_WORKSPACE=my-first-workspace
./bin/garuda workspace create my-first-workspace
```

Set `GARUDA_WORKSPACE` explicitly in every shell and MCP client configuration. Without an explicit workspace, commands may resolve the most recently updated workspace within the applicable scope.

### Confirm the workspace

```bash
./bin/garuda workspace list
```

The output includes workspace names and UUIDs. Dashboards, MCP clients, CI runners, and APIs use this identity to resolve the active workspace.

### Inspect workspace state

```bash
./bin/garuda summary
./bin/garuda status
```

---

## Analyzing Code

### Analyze one repository

```bash
./bin/garuda analyze /path/to/repo --save
```

`--save` persists entities, relationships, and evidence. Without it, Garuda analyzes the repository without making the result available to other clients.

Language detection is automatic:

| Detected file | Language selected |
|---|---|
| `go.mod` | Go |
| `pyproject.toml` or `setup.py` | Python |
| `tsconfig.json` or `package.json` | TypeScript |

### Analyze multiple repositories

Run the command once per repository in the same workspace:

```bash
./bin/garuda analyze ~/code/service-a --save
./bin/garuda analyze ~/code/service-b --save
./bin/garuda analyze ~/code/shared-lib --save
```

Supported cross-repository edges are computed when import paths resolve to entities in other repositories.

### What analysis produces

| Output | Description |
|---|---|
| Entities | Structs, interfaces, functions, methods, fields, packages, and supported external entities |
| Relationships | Calls, imports, implements, embeds, inherits, references, and cross-repository bridges |
| Evidence | Source file, line span, and derivation context supporting a relationship |
| Identity | Stable semantic identity for query and persistence across analyses |
| Maturity | Compiler-backed, structural, or heuristic resolution tier |

### Check persisted results

```bash
./bin/garuda summary
```

The summary reports repository, package, entity, and relationship counts. A successful saved analysis should produce non-zero counts for a non-empty codebase.

### View impact

```bash
./bin/garuda impact <entity-name>
```

The CLI reports graph-visible callers and transitive impact. Results describe the evidence present in the graph; they are not a guarantee that every dynamic runtime dependency has been discovered.

---

## Connecting AI Clients

Garuda exposes its semantic graph, governance functions, evidence queries, and multi-agent coordination through MCP using line-delimited JSON-RPC 2.0 over stdio.

### 1. Check the MCP binary

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./bin/garuda-mcp
```

The response should contain protocol version `2025-06-18` and a `serverInfo` block.

### 2. Configure Cursor

Add Garuda to Cursor's MCP configuration with an absolute binary path:

```json
{
  "mcpServers": {
    "garuda": {
      "command": "/absolute/path/to/bin/garuda-mcp",
      "env": {
        "DATABASE_URL": "postgres://garuda:garuda@localhost:5432/garuda?sslmode=disable",
        "GARUDA_WORKSPACE": "my-first-workspace"
      }
    }
  }
}
```

Cursor does not expand `~` and does not depend on your interactive shell's `PATH`.

### 3. Configure Claude Desktop

Use the same JSON structure in the platform-specific file:

| Platform | Configuration path |
|---|---|
| macOS | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Linux | `~/.config/Claude/claude_desktop_config.json` |
| Windows | `%APPDATA%\\Claude\\claude_desktop_config.json` |

Restart Claude Desktop after editing the configuration.

### 4. Configure Codex CLI

```bash
codex mcp add garuda \
  --env DATABASE_URL="postgres://garuda:garuda@localhost:5432/garuda?sslmode=disable" \
  --env GARUDA_WORKSPACE="my-first-workspace" \
  -- /absolute/path/to/bin/garuda-mcp
```

### 5. Verify the reference client

```bash
WORKSPACE=my-first-workspace python3 scripts/mcp_verify.py
```

Current reference-verifier result:

```text
═══ 38/38 checks passed ═══
```

The verifier exercises tool discovery, positive and negative paths, structured errors, protocol notifications, and clean shutdown behavior.

### 6. Verified client matrix

| Client | What was validated |
|---|---|
| Cursor | MCP integration in the developer IDE |
| Claude Desktop | MCP integration in a desktop client with strict response handling |
| Codex CLI | MCP integration in an independent CLI client |
| Reference verifier | Dependency-free protocol and negative-path contract checks |

### 7. Use the agent integration

Example prompt:

```text
What calls HandleCharge in the payments service, and what policy constraints apply?
```

A connected agent should query Garuda and return specific entities, packages, relationships, claims, and policy context rather than relying only on local file context.

---

## Workspace Console and IDE

### Workspace Console

Start the service:

```bash
./bin/garuda dev
```

Open:

```text
http://localhost:8080
```

The tenant-facing Workspace Console provides:

| Surface | What it shows |
|---|---|
| Workspace | Repository, package, entity, and relationship counts; language composition; architectural hubs; recent evidence; workspace health |
| Agents | MCP surface, active and recent sessions, tool-call activity, and agent coordination state |
| Governance | Policy enforcement, documentation claims, drift, contradictions, verification states, and Merkle trust |
| Decisions | Decision history, evidence, ancestors, descendants, revision chains, and anchor status |
| Graph | Interactive topology, communities, repositories, packages, relationships, and exploration controls |

The Console can display policy violations, contradictions, unverified claims, graph relationships, runtime evidence, governance attention items, and impact-related information within the selected workspace.

### VS Code extension

From the repository root:

```bash
cd vscode-extension
npm install
npm run compile
```

Open the repository in VS Code and use the Garuda Architecture Shield commands and views.

| Extension capability | Description |
|---|---|
| Cryptographic Ledger | View ledger and verification state |
| Quarantined Contradictions | Inspect contradiction diagnostics |
| Problems-panel diagnostics | Surface analyzer-backed violations and runtime contradictions |
| Graph visualizer | Open the architecture graph |
| Workspace reanalysis | Re-index the workspace AST |
| State refresh | Refresh ledger and verification state |
| Go symbol hover | Display blast-radius and dependency information when enabled |

Configure these settings when required:

| Setting | Purpose |
|---|---|
| `garuda.executablePath` | Path to the Garuda CLI binary |
| `garuda.databaseUrl` | PostgreSQL connection string |
| `garuda.daemonUrl` | Unified daemon HTTP API URL |
| `garuda.enableHoverBlastRadius` | Enable blast-radius and dependency information on Go symbol hover |

The extension is tested for editor-integrated diagnostics and exploration. Its update behavior depends on configured daemon/database connectivity and refresh or reanalysis events. Do not interpret “real-time” as a guarantee that every file mutation is continuously analyzed without an analysis event.

---

## Working with Policies

Policies are YAML files checked into the repository and evaluated against the current workspace.

### Example policy

Create `policies/payment-no-direct-db.yaml`:

```yaml
id: payment-no-direct-db
version: v1
title: "Payment code must not import database/sql directly"
priority: 200
language: go
authority: "team-platform@company.com"
scope:
  domain: payments
when:
  - type: claim_exists
    params:
      claim_type: IMPORTS
      from_name_pattern: ".*Payment.*"
      to_name_pattern: ".*sql.*"
then:
  decision: BLOCK
  reason: "Direct database/sql import from payment code violates hexagonal boundary."
```

The `when` block contains predicates. A policy fires when all predicates in the list are true. The `then` block defines the decision and reason.

### Supported predicates

| Predicate | Purpose |
|---|---|
| `entity_exists` | Tests whether matching entities exist in the workspace |
| `claim_exists` | Tests whether matching relationships or claims exist |
| `contradiction_exists` | Tests whether a `CONTRADICTED` verification exists |
| `verification_missing` | Tests whether an entity lacks required runtime verification |
| `language_matches` | Tests whether the scope contains entities of a given language |

### Decision levels

| Decision | Meaning |
|---|---|
| `ALLOW` | No blocking condition was found |
| `WARN` | A non-blocking condition is reported |
| `REVIEW` | Human or specialist review is required |
| `BLOCK` | The policy result should prevent the governed action or merge |

### Validate, preview, reconcile, and anchor

Validate syntax without evaluating against the database:

```bash
./bin/garuda policy validate ./policies
```

Preview policy decisions without persistence:

```bash
./bin/garuda policy evaluate ./policies
```

Persist and anchor the evaluation:

```bash
./bin/garuda policy evaluate ./policies --save
```

Preview mode leaves no trace. `--save` persists the evaluation and anchors it in the Merkle ledger. `--reconcile` retires active policy rows whose YAML file is absent and requires `--save`:

```bash
./bin/garuda policy evaluate ./policies --save --reconcile
```

### Inspect and verify an evaluation

```bash
./bin/garuda policy show <evaluation-id>
./bin/garuda policy verify <evaluation-id>
```

`policy show` displays the policy, evidence, matching entities and claims, predicates, decision, and Merkle block height. `policy verify` re-derives and checks the inclusion proof.

---

## Ingesting Documentation

Garuda connects normative documentation to implementation evidence by extracting claims from Markdown, ADR, and plain-text files.

### Supported claim format

A line is eligible for extraction only when both conditions are true:

1. It contains a modal term such as `MUST`, `SHOULD`, `MUST NOT`, or `SHALL`.
2. It appears under a heading suggesting normative content, such as `Decision`, `Requirement`, `Constraint`, `Policy`, or `Specification`.

Prose, tables, and code blocks are intentionally excluded to reduce false positives. See [`docs/DOCUMENT_FORMATS.md`](docs/DOCUMENT_FORMATS.md) for the complete format contract.

### Ingest a directory

```bash
./bin/garuda docs ingest ./docs
```

The command reports discovered, processed, empty, skipped, and extracted-claim counts.

### Configure multiple sources

Create `.garuda/workspace.yaml`:

```yaml
workspace: my-first-workspace
tenant_id: 00000000-0000-0000-0000-000000000001
documents:
  - path: ./docs
  - path: ../shared-docs
  - path: /absolute/path/to/adr
```

Synchronize configured sources:

```bash
./bin/garuda docs sync
```

### Verify claims against code

```bash
./bin/garuda docs verify
```

| Verification result | Meaning |
|---|---|
| `SUPPORTED` | The claim matches available semantic or runtime evidence |
| `UNVERIFIED` | Sufficient matching evidence was not found |
| `CONTRADICTED` | Available runtime or semantic evidence conflicts with the claim |

An unverified claim is not automatically false. It means the current evidence is insufficient.

---

## Runtime Evidence and Contradictions

Garuda can ingest runtime observations and correlate them with semantic entities.

### Runtime workflow

```text
Telemetry or runtime source
          ↓
Runtime observation ingestion
          ↓
Entity and operation correlation
          ↓
Evidence or contradiction record
          ↓
Policy evaluation
          ↓
ALLOW / WARN / REVIEW / BLOCK
```

### What runtime evidence can support

- Runtime observation records.
- Trace and span association.
- Entity and operation correlation.
- Supported or contradicted verification state.
- `verification_missing` policy evaluation.
- Dashboard attention items.
- IDE contradiction diagnostics where configured.

### Current boundary

The runtime path is implemented and exercised with controlled and synthetic evidence. Larger-scale autonomous correlation, dead-code detection, and unused-import detection are next-arc capabilities.

### Review workflow

When an exported handler lacks runtime evidence, a policy can produce `REVIEW` rather than incorrectly declaring the handler broken. The review record can show the matched entities, predicate, reason, decision, and Merkle block or anchor status.

---

## Verifying What You Built

### Workspace and ledger status

```bash
./bin/garuda status
```

Reports database reachability, Merkle block height, root hash, verification distribution, and daemon status.

### Inspect an entity

```bash
./bin/garuda inspect <entity-name>
```

Displays package, kind, incoming relationships, outgoing relationships, and matching documentation claims.

### Analyze change impact

```bash
./bin/garuda impact <entity-name>
```

Reports graph-visible callers and transitive impact. Sparse leaf call edges are a property of the analyzed code; the command reports what the graph contains.

### Compare snapshots

```bash
./bin/garuda analyze . --save -o /tmp/before.json
./bin/garuda analyze . --save -o /tmp/after.json
./bin/garuda diff /tmp/before.json /tmp/after.json
```

The semantic diff reports added, removed, or changed entities and relationships. Both snapshots must be produced by `garuda analyze`.

### Verify ledger integrity

```bash
./bin/garuda verify
```

Use this to verify the ledger and decision-chain integrity exposed by the CLI.

---

## MCP Server Reference

The MCP server exposes 21 tools over line-delimited JSON-RPC 2.0 over stdio.

### Tool catalog

| Group | Tool | Description | Access |
|---|---|---|---|
| Session / overview | `garuda.briefing` | Workspace state, trust anchor, scale, hubs, policies, contradictions, and recent changes | Read-only |
| Code graph | `garuda.entities` | Lists semantic entities, filterable by package and kind | Read-only |
| Code graph | `garuda.find_entity` | Finds entities by name pattern, kind, package, or file path | Read-only |
| Code graph | `garuda.inspect` | Inspects one entity and incoming/outgoing relationships | Read-only |
| Code graph | `garuda.neighbors` | Lists one-hop inbound and outbound connections | Read-only |
| Code graph | `garuda.subclasses` | Finds inheritance or embedding relationships | Read-only |
| Code graph | `garuda.implementers` | Finds interface implementers | Read-only |
| Code graph | `garuda.blast_radius` | Computes graph-visible impact of changing a symbol | Read-only |
| Code graph | `garuda.query_claims` | Queries documentation claims related to a subject | Read-only |
| Code graph | `garuda.query` | Queries the knowledge graph using natural language | Read-only |
| Governance | `garuda.policy.list` | Lists policies registered for the tenant | Read-only |
| Governance | `garuda.policy.evaluate` | Performs a non-persisting policy evaluation | Read-only |
| Governance | `garuda.verify_policy_evaluation` | Verifies a persisted evaluation proof | Read-only |
| Governance | `garuda.governance.status` | Aggregates active policies, documentation health, and contradiction count | Read-only |
| Governance | `garuda.check_drift` | Produces a documentation-to-code drift report | Read-only |
| Governance | `garuda.get_lineage` | Returns full decision lineage | Read-only |
| Governance | `garuda.get_impact` | Reports impact if a decision changes | Read-only |
| Governance | `garuda.detect_contradictions` | Lists unresolved contradictions | Read-only |
| Governance | `garuda.propose_decision` | Creates a budget-checked decision draft | Mutating |
| Multi-agent coordination | `garuda.handoff` | Atomically transfers work and creates a checkpoint | Mutating |
| Multi-agent coordination | `garuda.resume` | Restores and consumes an active checkpoint | Mutating |

### Mutation semantics

The mutating tools are `garuda.propose_decision`, `garuda.handoff`, and `garuda.resume`.

- `garuda.handoff` runs in a Serializable transaction, creates a checkpoint, records the handoff, and transitions agent/task state.
- `garuda.resume` consumes a checkpoint in a Serializable transaction and marks it restored.
- A second resume of the same checkpoint returns `{ "status": "not_found" }` because double consumption is rejected transactionally.

### Tool behavior clarifications

- `garuda.policy.evaluate` is a preview by default. The CLI uses `--save` for persistence and anchoring.
- `garuda.blast_radius` answers what may be affected if a symbol changes from the semantic graph.
- `garuda.get_impact` answers what may be affected if a decision changes from the lineage DAG.
- Unknown tools, evaluations, entities, tasks, or checkpoints should return structured failures rather than transport-level failures.

### MCP compliance verification

```bash
WORKSPACE=my-first-workspace python3 scripts/mcp_verify.py
```

The current reference verifier reports 38/38 checks. It covers initialization, protocol version, tool discovery, graph queries, policy-proof failure handling, blast-radius output, handoff and resume negative paths, unknown-tool errors, notification behavior, and clean shutdown.

---

## Multi-Agent Coordination

Garuda coordinates agent work at the task-transfer and checkpoint layer. It does not itself provide model inference or general-purpose agent-process orchestration.

### Handoff and resume

A handoff can:

1. Validate source and target agent state.
2. Validate task ownership and scope.
3. Create a checkpoint.
4. Record the handoff.
5. Transfer task ownership.
6. Transition agent state.
7. Commit atomically.

Resume restores an active checkpoint and consumes it transactionally. Duplicate resume attempts are rejected with a structured `not_found` response.

### Sessions and tool calls

The MCP activity surface can record:

- Client name and version.
- Agent identity.
- Session identity and lifecycle.
- Workspace and tenant association.
- Tool name.
- Duration.
- Success or error status.
- Whitelisted argument summaries.

Free-form query and business-context fields should not be recorded in argument summaries.

---

## Running in CI

### Block policy violations

```bash
./bin/garuda policy evaluate ./policies --fail-on BLOCK
```

`--fail-on` accepts `BLOCK`, `REVIEW`, `WARN`, or `ALLOW`. The command exits non-zero when an evaluation reaches the selected threshold or higher.

### Annotate pull requests with impact

```bash
./bin/garuda impact <changed-symbol> --format github
```

The GitHub format emits annotations for graph-visible callers and impact results.

### Compare semantic snapshots

```bash
git stash
./bin/garuda analyze . --save -o /tmp/before.json
git stash pop
./bin/garuda analyze . --save -o /tmp/after.json
./bin/garuda diff /tmp/before.json /tmp/after.json
```

### Recommended CI sequence

```text
1. Start PostgreSQL.
2. Apply migrations.
3. Resolve the tenant and target workspace.
4. Analyze the repository with --save.
5. Validate policies.
6. Evaluate policies with --fail-on BLOCK.
7. Verify documentation or runtime evidence when relevant.
8. Publish semantic diff, impact, and verification artifacts.
```

---

## Troubleshooting

| Symptom | Likely cause | Resolution |
|---|---|---|
| `garuda dev` reports `bind: address already in use` | Another process is listening on port 8080 | Stop it or run `./bin/garuda dev --port 8081` |
| `garuda analyze` finds zero entities | Wrong module location, build errors, or no exported/referenced symbols | Check for `go.mod`, run `go build ./...`, and inspect package visibility |
| MCP connects but tools return empty results | Workspace or database mismatch | Compare `GARUDA_WORKSPACE` and `DATABASE_URL` with the values used during analysis |
| MCP reports a missing tool or unexpected response | Stale MCP binary | Rebuild `bin/garuda-mcp` and restart the MCP client |
| MCP verifier count differs from documentation | Stale evidence or verifier output | Run `scripts/mcp_verify.py`, then synchronize README, PLAYBOOK, EVIDENCE, and SPECS |
| VS Code diagnostics are empty | Extension not compiled, wrong executable/database/daemon settings, or state not refreshed | Run `npm run compile`, verify settings, re-index the workspace, and refresh Garuda state |
| `policy verify` reports a proof failure | Row tampering, reset root hash, or verification-format mismatch | Check database history, workspace lifecycle, and verification version |
| Dashboard is empty | Analysis was not saved or targeted another workspace | Run `garuda summary` and confirm non-zero counts |
| Policy evaluation returns no matches | Wrong workspace, unsupported predicate path, or policy scope mismatch | Confirm workspace, language, predicate, and analyzed entities |

### Inspect workspace resolution

```bash
psql "$DATABASE_URL" -c "SELECT id, tenant_id, name FROM workspaces ORDER BY updated_at DESC;"
```

Compare the returned workspace and tenant with `GARUDA_WORKSPACE` and the configured tenant context.

### Inspect MCP tool count

```bash
WORKSPACE=my-first-workspace python3 scripts/mcp_verify.py
```

The current reference result should be:

```text
═══ 38/38 checks passed ═══
```

### Proof failure details

A policy proof failure can indicate:

- A database row was edited after the proof was written.
- The tenant root hash was reset, commonly after dropping and recreating state.
- The proof format changed between writing and reading versions.

Check the evaluation identifier, block height, and verification version before filing an issue.

---

## Command Reference

Run `garuda <command> --help` for the authoritative flags for the installed binary.

### Analysis

| Command | Purpose |
|---|---|
| `garuda analyze <path>` | Analyze a repository |
| `garuda diff <before> <after>` | Compare semantic snapshots |
| `garuda inspect <entity>` | Inspect an entity and relationships |
| `garuda entities` | List entities in the workspace |
| `garuda impact <entity>` | Compute graph-visible transitive impact |
| `garuda summary` | Show workspace architectural counts |

### Workspaces and repositories

| Command | Purpose |
|---|---|
| `garuda workspace create <name>` | Create a workspace |
| `garuda workspace list` | List workspaces |
| `garuda workspace delete <name>` | Delete a workspace |
| `garuda repo add <url>` | Register a repository |
| `garuda repo list` | List repositories |
| `garuda repo enable <id>` | Enable a disabled repository |
| `garuda repo disable <id>` | Disable a repository |
| `garuda repo remove <id>` | Remove a repository |

### Policies

| Command | Purpose |
|---|---|
| `garuda policy list` | List active policies |
| `garuda policy validate <dir>` | Parse and validate policy YAML |
| `garuda policy evaluate <dir>` | Preview or persist policy evaluations |
| `garuda policy show <id>` | Show evaluation evidence and proof details |
| `garuda policy verify <id>` | Verify a Merkle inclusion proof |

### Documentation

| Command | Purpose |
|---|---|
| `garuda docs ingest <path>` | Extract claims from documents |
| `garuda docs sync` | Ingest configured document sources |
| `garuda docs status` | Show freshness of configured sources |
| `garuda docs verify` | Correlate claims with code and detect drift |

### Servers and integrations

| Command | Purpose |
|---|---|
| `garuda dev` | Start the unified API, worker, and dashboard daemon |
| `garuda mcp` | Print the MCP server binary path |
| `garuda dashboard` | Open the web dashboard |

### Verification and governance

| Command | Purpose |
|---|---|
| `garuda status` | Inspect ledger, verification, and daemon status |
| `garuda verify` | Verify ledger and decision-chain integrity |
| `garuda bench` | Run the grounding benchmark |
| `garuda ci` | Run CI mode with baseline comparison |
| `garuda judge <baseline> <proposed>` | Produce a governance judgment between snapshots |

### VS Code extension

| Command or operation | Purpose |
|---|---|
| `npm install` in `vscode-extension/` | Install extension dependencies |
| `npm run compile` | Compile the extension |
| `Garuda: Refresh Ledger & Verification State` | Refresh ledger and verification state |
| `Garuda: Open Graph Visualizer` | Open graph visualization |
| `Garuda: Re-index Workspace AST` | Re-analyze the workspace |

### MCP tools

The current MCP surface contains 21 tools grouped into session context, graph exploration, governance and decisions, and multi-agent coordination. See the [MCP Server Reference](#mcp-server-reference) for the complete catalog.

---

## Evidence and Specifications

| Document | Purpose |
|---|---|
| [`README.md`](README.md) | Product positioning and overview |
| [`docs/SPECS.md`](docs/SPECS.md) | Detailed product and system specification |
| [`EVIDENCE.md`](EVIDENCE.md) | Validation results and methodology |
| [`docs/CAPABILITIES.md`](docs/CAPABILITIES.md) | Analyzer and product capability maturity matrix |
| [`docs/specs/`](docs/specs/) | Subsystem and operational specifications |
| [`docs/DOCUMENT_FORMATS.md`](docs/DOCUMENT_FORMATS.md) | Supported document formats and claim extraction |
| [`docs/invariants.md`](docs/invariants.md) | Invariant contract for the trust substrate |
| [`docs/adr/`](docs/adr/) | Architecture decision records |
| [`vscode-extension/`](vscode-extension/) | VS Code integration |
| [`scripts/mcp_verify.py`](scripts/mcp_verify.py) | Dependency-free MCP verifier |
| [`openapi.yaml`](openapi.yaml) | HTTP API contract |
| [`SECURITY.md`](SECURITY.md) | Security model and reporting |
| [`CONTRIBUTING.md`](CONTRIBUTING.md) | Contribution guidelines |

For implementation details, consult `docs/SPECS.md`. For current validation numbers, prefer the latest reproducible command output and synchronized evidence documents over historical drafts.

---

## Support and Contribution

For issue reports, use GitHub Issues. For design discussions, use GitHub Discussions. Architecture decisions are documented in `docs/adr/`, and reproducible evidence is documented in `EVIDENCE.md`.

Contributions are welcome across analyzers, semantic resolution, policy evaluation, runtime evidence, MCP, dashboards, VS Code tooling, testing, and documentation.