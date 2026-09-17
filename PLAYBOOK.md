# Garuda Playbook

**Installation · Setup · Everyday Workflows · Agent Integration · Reference**

This playbook is a practical guide to running Garuda against a real codebase. It assumes you have read the README and want the exact commands, expected outputs, and operating workflows for the current release.

> **Release contract:** Commands and behaviors documented here describe the current release. If an observed behavior differs, file an issue with the command, environment, release commit, and output.

---

## Table of Contents

1. [What Garuda Is For](#what-garuda-is-for)
2. [Architecture at a Glance](#architecture-at-a-glance)
3. [Before You Start](#before-you-start)
4. [Installation](#installation)
5. [Your First Workspace](#your-first-workspace)
6. [Analyzing Code](#analyzing-code)
7. [Connecting an AI Agent](#connecting-an-ai-agent)
8. [Working with Policies](#working-with-policies)
9. [Ingesting Documentation](#ingesting-documentation)
10. [Verifying What You Built](#verifying-what-you-built)
11. [MCP Server Reference](#mcp-server-reference)
12. [Running in CI](#running-in-ci)
13. [Troubleshooting](#troubleshooting)
14. [Command Reference](#command-reference)

---

## What Garuda Is For

Garuda builds a verified semantic model of software and makes that model available to developers and AI agents. The model is stored in PostgreSQL and is derived from compiler-backed type information where supported, rather than from text-pattern matching.

Garuda is primarily used to:

- Give AI agents grounded answers about what actually exists in a codebase.
- Evaluate proposed or existing code against organizational policies.
- Connect normative documentation to implementation evidence.
- Record governance decisions in a cryptographically verifiable Merkle ledger.
- Analyze dependency neighborhoods, callers, implementers, and change impact.
- Coordinate multi-agent work through transactional handoffs and checkpoints.

### Operating model

```text
Repositories ──analyze──▶ Semantic Graph ──query──▶ Developers / AI Agents
       │                         │
       └── docs ingest ──▶ Claims ──verify──▶ Drift and contradictions
                                  │
Policies ──evaluate──▶ Decisions ──anchor──▶ Merkle Ledger ──verify──▶ Evidence
```

Everything is local by default. No data leaves the machine unless you configure an external service or integration.

---

## Architecture at a Glance

| Layer | Responsibility | Primary interface |
|---|---|---|
| Repository analysis | Extract entities, relationships, and evidence | `garuda analyze` |
| PostgreSQL model | Persist workspace, repository, graph, claim, and governance data | `DATABASE_URL` |
| Documentation claims | Extract normative requirements from supported documents | `garuda docs ingest` |
| Verification | Correlate claims with code and runtime evidence | `garuda docs verify` |
| Policy engine | Evaluate organizational rules and produce decisions | `garuda policy evaluate` |
| Trust layer | Anchor persisted decisions in a Merkle ledger | `garuda status`, `garuda policy verify` |
| AI integration | Expose graph and governance tools over MCP | `garuda-mcp` |
| Agent coordination | Create and consume transactional handoffs | `garuda.handoff`, `garuda.resume` |

---

## Before You Start

You need:

| Requirement | Minimum or supported option |
|---|---|
| PostgreSQL | Version 14 or later |
| Go | Version 1.26 or later |
| Codebase | A Go module for compiler-backed analysis; Python and TypeScript are also supported |
| Shell | Bash, Zsh, or an equivalent shell |
| Editor | Any text editor |

You do not need:

- A cloud account.
- An API key.
- A license.

### Analyzer support

| Language | Detection | Analysis mode |
|---|---|---|
| Go | `go.mod` | Compiler-backed and validated |
| Python | `pyproject.toml` or `setup.py` | Structural |
| TypeScript | `tsconfig.json` or `package.json` | Structural; requires CGO and a C compiler |

---

## Installation

### Build from source

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda

go build -o bin/garuda ./cmd/garuda
go build -o bin/garuda-mcp ./cmd/garuda-mcp
```

For TypeScript analysis, enable CGO and ensure a C compiler is available:

```bash
export CGO_ENABLED=1
```

Rebuild `bin/garuda-mcp` whenever you pull changes to `cmd/garuda-mcp/` or `internal/policy/`. MCP clients spawn the binary independently and will not detect a stale build until they are restarted.

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

Persist it in `~/.bashrc`, `~/.zshrc`, or your shell's equivalent if appropriate.

### Apply migrations

```bash
for f in migrations/[0-9]*.sql; do
  psql "$DATABASE_URL" -f "$f"
done
```

Migrations are idempotent. Running them again is safe. When upgrading from an older release, run every migration in order so that the migration files can apply the required upgrade logic.

### Verify the installation

```bash
./bin/garuda --version
./bin/garuda status
```

`garuda status` reports database reachability, Merkle block height, root hash, and verification state distribution.

---

## Your First Workspace

A workspace is a logical group of repositories. Use one workspace per product, domain, or team.

### Create a workspace

```bash
export GARUDA_WORKSPACE=my-first-workspace
./bin/garuda workspace create my-first-workspace
```

Set `GARUDA_WORKSPACE` explicitly. Without it, many commands fall back to the most recently updated workspace, which can produce confusing results.

### Confirm the workspace

```bash
./bin/garuda workspace list
```

The output includes the workspace name and UUID. Other tools, dashboards, and CI runners use this identity to resolve the workspace.

---

## Analyzing Code

### Analyze one repository

```bash
./bin/garuda analyze /path/to/repo --save
```

`--save` persists entities, relationships, and evidence. Without it, Garuda analyzes the repository without making the result visible to other clients.

Language detection is automatic:

| Detected file | Language selected |
|---|---|
| `go.mod` | Go |
| `pyproject.toml` or `setup.py` | Python |
| `tsconfig.json` or `package.json` | TypeScript |

### Analyze multiple repositories

Run the command once for each repository in the same workspace:

```bash
./bin/garuda analyze ~/code/service-a --save
./bin/garuda analyze ~/code/service-b --save
./bin/garuda analyze ~/code/shared-lib --save
```

Cross-repository edges are computed during subsequent analyses when an import path resolves to an entity in another repository.

### Analyzer output

| Output | Description |
|---|---|
| Entities | Structs, interfaces, functions, methods, fields, and packages |
| Relationships | Calls, imports, implements, embeds, and references |
| Evidence | Source file and line supporting each relationship |
| Relationship types | Distinguishes calls from imports, embeds, references, and other edges |

### Check persisted results

```bash
./bin/garuda summary
```

The summary reports repository, entity, and relationship counts. Counts should be non-zero for a successfully analyzed codebase and should increase when additional repositories are saved.

---

## Connecting an AI Agent

Garuda exposes its semantic graph and governance functions through an MCP server using line-delimited JSON-RPC 2.0 over standard input and output.

### 1. Check the MCP binary

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./bin/garuda-mcp
```

The response should contain `protocolVersion: "2025-06-18"` and a `serverInfo` block. A `DATABASE_URL` error means the variable is not available in the current shell.

### 2. Configure Cursor

Add the following to Cursor's MCP configuration. Replace the binary path with an absolute path:

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

Cursor does not expand `~` and does not rely on your interactive shell's `PATH`.

### 3. Configure Claude Desktop

Use the same JSON structure in the platform-specific configuration file:

| Platform | Configuration path |
|---|---|
| macOS | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Linux | `~/.config/Claude/claude_desktop_config.json` |
| Windows | `%APPDATA%\\Claude\\claude_desktop_config.json` |

Restart Claude Desktop after editing the file.

### 4. Verify the connection

```bash
WORKSPACE=my-first-workspace python3 scripts/mcp_verify.py
```

Expected result:

```text
═══ 36/36 checks passed ═══
```

The most common causes of failure are an incorrect binary path, a stale `DATABASE_URL`, a workspace mismatch, or a stale MCP binary.

### 5. Supported clients

| Client | Configuration method |
|---|---|
| Codex CLI | `codex mcp add garuda --env ... -- /path/to/garuda-mcp` |
| Claude Desktop | Platform-specific JSON configuration |
| Cursor | MCP settings panel or `.cursor/mcp.json` |
| Other MCP clients | Must support line-delimited JSON-RPC 2.0 over stdio |

### 6. Use the agent integration

Example prompt:

```text
What calls HandleCharge in the payments service?
```

A connected agent should use `garuda.neighbors` or `garuda.find_entity` and return specific entities, packages, and relationships grounded in the semantic graph.

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
| `claim_exists` | Tests whether matching relationships exist in the workspace |
| `contradiction_exists` | Tests whether a `CONTRADICTED` verification exists |
| `verification_missing` | Tests whether an entity lacks runtime verification |
| `language_matches` | Tests whether the scope contains entities of a given language |

### Decision levels

| Decision | Meaning |
|---|---|
| `ALLOW` | The change or state is permitted |
| `WARN` | The condition is reported but does not necessarily block |
| `REVIEW` | Human or governance review is required |
| `BLOCK` | The condition must prevent the operation or merge |

### Validate, evaluate, and anchor

Validate syntax without evaluating against the database:

```bash
./bin/garuda policy validate ./policies
```

Preview policy decisions without persistence:

```bash
./bin/garuda policy evaluate ./policies
```

Persist the evaluation and anchor it in the Merkle ledger:

```bash
./bin/garuda policy evaluate ./policies --save
```

Preview mode leaves no trace. `--save` creates an auditable evaluation at a specific point in time. The CLI refuses `--reconcile` without `--save` because reconciliation retires policy rows that are no longer represented on disk.

### Inspect and verify an evaluation

```bash
./bin/garuda policy show <evaluation-id>
./bin/garuda policy verify <evaluation-id>
```

`policy show` displays the fired policy, matching entities and claims, matched predicates, and Merkle block height. `policy verify` re-derives the inclusion proof against the current epoch root and reports any failure.

---

## Ingesting Documentation

Garuda connects normative documentation to implementation evidence by extracting claims from Markdown, ADR, and plain-text files.

### Supported claim format

A line is eligible for extraction only when both conditions are true:

1. It contains a modal term such as `MUST`, `SHOULD`, `MUST NOT`, or `SHALL`.
2. It appears under a heading suggesting normative content, such as `Decision`, `Requirement`, `Constraint`, `Policy`, or `Specification`.

Prose, tables, and code blocks are intentionally excluded to reduce false positives. See `docs/DOCUMENT_FORMATS.md` for the complete format contract.

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

Synchronize all configured sources:

```bash
./bin/garuda docs sync
```

### Verify claims against code

```bash
./bin/garuda docs verify
```

| Verification result | Meaning |
|---|---|
| `SUPPORTED` | The claim matches a semantic entity or relationship |
| `UNVERIFIED` | No matching implementation evidence was found |
| `CONTRADICTED` | Runtime behavior conflicts with the claim |

Currently, only `CALLS` predicates are verified against the graph. Other predicates remain `UNVERIFIED` by design.

---

## Verifying What You Built

### Workspace and ledger status

```bash
./bin/garuda status
```

Reports the Merkle block height, root hash, verification distribution, and daemon status.

### Inspect an entity

```bash
./bin/garuda inspect <entity-name>
```

Displays the entity's package, kind, incoming relationships, outgoing relationships, and matching documentation claims.

### Analyze change impact

```bash
./bin/garuda impact <entity-name>
```

Reports callers transitively and provides the full graph-visible blast radius. Results reflect the edges present in the semantic graph; sparse leaf call edges are a property of the analyzed code, not an assumption that every function has many callers.

### Compare snapshots

```bash
garuda analyze . --save -o before.json
garuda analyze . --save -o after.json
./bin/garuda diff before.json after.json
```

The semantic diff reports added, removed, or changed entities and relationships. Both snapshots must be produced by `garuda analyze`.

---

## MCP Server Reference

The MCP server exposes twenty-one tools over line-delimited JSON-RPC 2.0 over stdio.

### Tool catalog

| Group | Tool | Description | Access |
|---|---|---|---|
| Session / overview | `garuda.briefing` | Reports workspace state, trust anchor, scale, hubs, policies, contradictions, and changes since the previous briefing | Read-only |
| Code graph | `garuda.entities` | Lists semantic entities, filterable by package and kind | Read-only |
| Code graph | `garuda.find_entity` | Finds entities by name pattern, kind, package, or file path | Read-only |
| Code graph | `garuda.inspect` | Inspects one entity and its incoming and outgoing relationships | Read-only |
| Code graph | `garuda.neighbors` | Lists all entities connected to a subject in one hop, in both directions | Read-only |
| Code graph | `garuda.subclasses` | Lists entities that inherit from or embed the subject | Read-only |
| Code graph | `garuda.implementers` | Lists entities that implement the subject interface | Read-only |
| Code graph | `garuda.blast_radius` | Computes the transitive impact of changing a subject | Read-only |
| Code graph | `garuda.query_claims` | Queries documentation claims matching a subject | Read-only |
| Code graph | `garuda.query` | Queries the knowledge graph using natural language | Read-only |
| Governance | `garuda.policy.list` | Lists policies registered for the tenant | Read-only |
| Governance | `garuda.policy.evaluate` | Performs a non-persisting policy evaluation | Read-only |
| Governance | `garuda.verify_policy_evaluation` | Verifies a Merkle inclusion proof for a persisted evaluation | Read-only |
| Governance | `garuda.governance.status` | Aggregates active policies, documentation health, and contradiction count | Read-only |
| Governance | `garuda.check_drift` | Produces a documentation-to-code drift report | Read-only |
| Governance | `garuda.get_lineage` | Returns the full lineage of a decision | Read-only |
| Governance | `garuda.get_impact` | Reports what breaks if a decision changes | Read-only |
| Governance | `garuda.detect_contradictions` | Lists unresolved contradictions in the tenant graph | Read-only |
| Governance | `garuda.propose_decision` | Creates a budget-checked decision draft | Mutating |
| Multi-agent coordination | `garuda.handoff` | Atomically hands off work between agents and creates a checkpoint | Mutating |
| Multi-agent coordination | `garuda.resume` | Restores an active checkpoint and consumes it transactionally | Mutating |

### Mutation semantics

`garuda.propose_decision`, `garuda.handoff`, and `garuda.resume` are the mutating tools.

- `garuda.handoff` runs in a Serializable transaction, creates a checkpoint, records the handoff, and transitions both agent statuses.
- `garuda.resume` consumes a checkpoint in a Serializable transaction and marks it restored.
- A second resume of the same checkpoint returns `{ "status": "not_found" }` because double consumption is rejected transactionally.

### MCP compliance verification

```bash
WORKSPACE=my-first-workspace python3 scripts/mcp_verify.py
```

The reference client runs 36 checks, including negative-path assertions for every tool. Unknown IDs must return structured failures rather than transport errors. The verification suite also checks that notifications do not receive forbidden responses.

---

## Running in CI

### Block policy violations

```bash
./bin/garuda policy evaluate ./policies --fail-on BLOCK
```

`--fail-on` accepts `BLOCK`, `REVIEW`, `WARN`, or `ALLOW`. The command exits non-zero when an evaluation reaches the selected level or higher.

### Annotate pull requests with impact

```bash
./bin/garuda impact <changed-symbol> --format github
```

The GitHub format emits annotations for graph-visible callers of the changed symbol.

### Compare semantic snapshots

```bash
git stash
./bin/garuda analyze . --save -o /tmp/before.json
git stash pop
./bin/garuda analyze . --save -o /tmp/after.json
./bin/garuda diff /tmp/before.json /tmp/after.json
```

The resulting structured JSON can be used to comment on a pull request, fail a build, or publish a CI artifact.

### Recommended CI sequence

```text
1. Start PostgreSQL.
2. Apply migrations.
3. Select the target workspace.
4. Analyze the repository with --save.
5. Validate policies.
6. Evaluate policies with --fail-on BLOCK.
7. Verify documentation claims when documentation is part of the change.
8. Publish semantic diff and impact artifacts.
```

---

## Troubleshooting

| Symptom | Likely cause | Resolution |
|---|---|---|
| `garuda dev` reports `bind: address already in use` | Another process is listening on port 8080 | Stop the process or run `./bin/garuda dev --port 8081` |
| `garuda analyze` finds zero entities | Not inside a Go module, build errors, or no exported/referenced symbols | Check for `go.mod`, run `go build ./...`, and inspect package visibility |
| MCP connects but tools return empty results | Workspace or database mismatch | Compare `GARUDA_WORKSPACE` and `DATABASE_URL` with the values used during analysis |
| MCP reports a missing tool or unexpected response | Stale `bin/garuda-mcp` | Rebuild the binary and restart the MCP client |
| `policy verify` reports a proof failure | Row tampering, reset root hash, or verification-format mismatch | Check database history, workspace lifecycle, and `verification_version` |
| Dashboard is empty | Analysis was not saved or targeted another workspace | Run `./bin/garuda summary` and confirm non-zero counts |

### Inspect workspace resolution

```bash
psql "$DATABASE_URL" -c "SELECT id, name FROM workspaces;"
```

Compare the returned workspace with the configured `GARUDA_WORKSPACE`.

### Proof failure details

A policy proof failure can indicate:

- A database row was edited after the proof was written.
- The tenant's root hash was reset, commonly after dropping and recreating a workspace.
- The proof format changed between the writing and reading versions.

Check `verification_version` on the evaluation row. The current format is version 1. If the cause remains unclear, file an issue with the evaluation ID and block height.

---

## Command Reference

### Analysis

| Command | Purpose |
|---|---|
| `garuda analyze <path>` | Analyze a repository |
| `garuda diff <before> <after>` | Compare semantic snapshots |
| `garuda inspect <entity>` | Inspect one entity and relationships |
| `garuda entities` | List entities in the workspace |
| `garuda impact <entity>` | Compute transitive blast radius |
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
| `garuda status` | Inspect ledger and daemon status |
| `garuda verify` | Verify ledger integrity |
| `garuda bench` | Run the GAP-20 grounding benchmark |
| `garuda ci` | Run CI mode with baseline comparison |
| `garuda judge <baseline> <proposed>` | Produce a governance judgment between snapshots |

### Help and support

```bash
garuda <command> --help
```

For issue reports, use GitHub Issues. For design discussions, use GitHub Discussions. Architecture decisions are documented in `docs/adr/`, and evidence and benchmarks are documented in `EVIDENCE.md`.