```markdown
# Garuda Playbook

**Installation · Setup · Everyday workflows · Agent integration · Reference**

This playbook covers the practical steps for running Garuda on a real codebase. It assumes you have read the README and want to know what to actually type.

Everything here is executed against the current release. If a command behaves differently than described, that is a bug — file an issue.

---

## Table of contents

1. [What Garuda is for](#what-garuda-is-for)
2. [Before you start](#before-you-start)
3. [Installation](#installation)
4. [Your first workspace](#your-first-workspace)
5. [Analyzing code](#analyzing-code)
6. [Connecting an AI agent](#connecting-an-ai-agent)
7. [Working with policies](#working-with-policies)
8. [Ingesting documentation](#ingesting-documentation)
9. [Verifying what you've built](#verifying-what-youve-built)
10. [The MCP server in detail](#the-mcp-server-in-detail)
11. [Running in CI](#running-in-ci)
12. [Troubleshooting](#troubleshooting)
13. [Command reference](#command-reference)

---

## What Garuda is for

Garuda builds a verified model of your software and gives it to every developer and every AI agent working on it. The model lives in a PostgreSQL database, is shared across every client, and answers questions about your code from compiler type information rather than from pattern matching.

The two things people use it for, most often:

1. **Telling an AI agent what actually exists in your code.** Instead of the agent opening files and guessing, it asks Garuda and gets a compiler-backed answer.
2. **Checking that a change is allowed.** A policy declares what your organization requires. Garuda evaluates every candidate change against it and returns a decision — allow, warn, review, or block — anchored to a cryptographic ledger.

Everything else is machinery that makes those two things reliable.

---

## Before you start

You need three things:

- **A PostgreSQL database.** Version 14 or later. Local, Docker, or hosted — any Postgres works.
- **A Go module to analyze.** The Go analyzer is compiler-backed and validated. Python and TypeScript analyzers work but are structural, not compiler-backed.
- **A shell, a text editor, and about twenty minutes.**

You do not need:

- A cloud account
- An API key
- A license

Everything is local. The Merkle trust layer runs in-process. No data leaves your machine unless you configure it to.

---

## Installation

### Build from source

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda
go build -o bin/garuda ./cmd/garuda
go build -o bin/garuda-mcp ./cmd/garuda-mcp
```

Requires Go 1.26 or later. The TypeScript analyzer uses tree-sitter and requires CGO — set `CGO_ENABLED=1` and have a C compiler on your `PATH`. If you skip the `garuda-mcp` binary, you can still use the CLI but AI agents will not be able to connect.

### Set up the database

If you do not have Postgres running locally, the quickest path is Docker:

```bash
docker run -d \
  --name garuda-pg \
  -e POSTGRES_USER=garuda \
  -e POSTGRES_PASSWORD=garuda \
  -e POSTGRES_DB=garuda \
  -p 5432:5432 \
  postgres:16
```

Then export the connection string in every shell where you use Garuda:

```bash
export DATABASE_URL="postgres://garuda:garuda@localhost:5432/garuda?sslmode=disable"
```

Add that line to your `~/.bashrc` or `~/.zshrc` so you do not have to type it every time.

### Apply migrations

```bash
for f in migrations/[0-9]*.sql; do
  psql "$DATABASE_URL" -f "$f"
done
```

Migrations are idempotent. Running them twice is safe. If you are upgrading from an older release, run all migrations in order — the migration files themselves contain the upgrade logic.

### Verify the install

```bash
./bin/garuda --version
./bin/garuda status
```

`garuda status` reports whether it can reach the database and what state the ledger is in.

---

## Your first workspace

A **workspace** is a logical group of repositories. Most teams use one workspace per product, domain, or team. You will create one, name it, and point your analysis at it.

### Create a workspace

```bash
export GARUDA_WORKSPACE=my-first-workspace
./bin/garuda workspace create my-first-workspace
```

`GARUDA_WORKSPACE` tells Garuda which workspace every subsequent command should target. If you do not set it, most commands fall back to the most recently updated workspace, which is rarely what you want. Set it.

### Confirm it exists

```bash
./bin/garuda workspace list
```

You should see `my-first-workspace` in the list with a UUID. That UUID is what other tools — dashboards, MCP clients, CI runners — will use to refer to this workspace.

---

## Analyzing code

### Analyze a single repository

```bash
./bin/garuda analyze /path/to/repo --save
```

`--save` is important. Without it, Garuda analyzes the code but does not persist the result. With it, entities, relationships, and evidence are written to the database and become visible to every client.

Language is detected automatically. A Go module is detected by `go.mod`. Python by `pyproject.toml` or `setup.py`. TypeScript by `tsconfig.json` or `package.json`.

### Analyze multiple repositories

Run `garuda analyze --save` once per repository, all in the same workspace:

```bash
./bin/garuda analyze ~/code/service-a --save
./bin/garuda analyze ~/code/service-b --save
./bin/garuda analyze ~/code/shared-lib --save
```

Cross-repository edges are computed automatically during the second and subsequent analyses, whenever an import path in one repository resolves to an entity in another.

### What the analyzer produces

For every file it reads, the analyzer extracts:

- **Entities.** Structs, interfaces, functions, methods, fields, packages.
- **Relationships.** Calls, imports, implements, embeds, references.
- **Evidence.** For every relationship, the source file and line that justifies it.

Every relationship is typed. A call is not the same as an import. An embed is not the same as a reference. This matters when you ask a policy question later — the policy can target a specific relationship type.

### Check what was saved

```bash
./bin/garuda summary
```

The summary reports repository count, entity count, and relationship count for the workspace. On the reference corpus — nine repositories across Go, Python, and TypeScript — the summary reports 9 repositories, 14,333 entities, and 40,956 relationships. Yours will differ; the important thing is that the numbers are non-zero and grow when you add repositories.

---

## Connecting an AI agent

This is what most people install Garuda for. It takes about five minutes.

### Step 1 — confirm the MCP binary works

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./bin/garuda-mcp
```

You should see a single line of JSON with `"protocolVersion":"2025-06-18"` and a `"serverInfo"` block. If you see an error about `DATABASE_URL`, the environment variable is not set in this shell.

### Step 2 — point Cursor at it

Open Cursor's settings, find the MCP configuration, and add:

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

The path must be absolute. Cursor does not expand `~` and does not use your shell's `PATH`.

### Step 3 — point Claude Desktop at it

Claude Desktop uses the same configuration format. The file is at:

- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Linux: `~/.config/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`

Restart Claude Desktop after editing the file.

### Step 4 — verify the connection

```bash
WORKSPACE=my-first-workspace python3 scripts/mcp_verify.py
```

Expected output:

```
═══ 18/18 checks passed ═══
```

If any check fails, the script prints which one. The most common failure is a wrong binary path. The second most common is a stale `DATABASE_URL`.

### Step 5 — use it

Once connected, ask your agent a question about your code:

> *What calls `HandleCharge` in the payments service?*

If Garuda is connected, the agent will call `garuda.neighbors` or `garuda.find_entity` and get a real answer. If it is not connected, the agent will open files and guess. You can tell the difference immediately — the Garuda answer will name specific entities with specific packages, and the guess will be hedged.

---

## Working with policies

Policies are how you tell Garuda what your organization requires. They are YAML files checked into your repository, evaluated against your workspace.

### Write a policy

Create a directory called `policies/` at the root of your repo. Add a file called `payment-no-direct-db.yaml`:

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

The `when` block is a list of predicates. A policy fires when every predicate in the list is true. The `then` block says what to do when it fires.

Five predicates are available:

- `entity_exists` — matching entities exist in the workspace
- `claim_exists` — matching relationships exist in the workspace
- `contradiction_exists` — a `CONTRADICTED` verification exists
- `verification_missing` — an entity has no runtime verification
- `language_matches` — the scope contains entities of a given language

Four outcomes are available: `ALLOW`, `WARN`, `REVIEW`, `BLOCK`.

### Validate without evaluating

```bash
./bin/garuda policy validate ./policies
```

This parses every policy in the directory and reports syntax errors without touching the database. Run it before every commit that changes a policy.

### Evaluate

```bash
./bin/garuda policy evaluate ./policies
```

Garuda evaluates each policy against the current workspace. Every evaluation produces one of the four outcomes and is anchored to the Merkle ledger.

### Show an evaluation

```bash
./bin/garuda policy show <evaluation-id>
```

The output includes the policy that fired, the entities and claims that matched, the matched predicates, and the Merkle block height at which the evaluation was committed.

### Verify an evaluation

```bash
./bin/garuda policy verify <evaluation-id>
```

This re-derives the Merkle inclusion proof from the stored proof and the current epoch root. Anyone with the source of the Merkle package can run this independently — the verification does not require trust in Garuda's database, servers, or keys.

If the proof verifies, the output says so and prints the block height. If it does not, the output says which part of the proof failed. A tampered evaluation will fail to verify.

---

## Ingesting documentation

Garuda can extract claims from your documentation and match them against the code. This is how intent gets connected to implementation.

### What the parser reads

The parser reads Markdown, ADR, and plain-text files. It extracts claims from lines that meet two conditions:

1. The line uses a modal verb — `MUST`, `SHOULD`, `MUST NOT`, `SHALL`.
2. The line lives inside a section whose heading suggests normative content — a heading containing words like "decision", "requirement", "constraint", "policy", "specification".

Lines that do not meet both conditions are ignored. Prose sentences, tables, and code blocks are not parsed.

This is deliberate. A parser that tries to extract every possible requirement from free-form prose will produce false positives, and false positives in a trust system are worse than false negatives. The format contract is documented in full in [`docs/DOCUMENT_FORMATS.md`](../docs/DOCUMENT_FORMATS.md).

### Ingest a directory

```bash
./bin/garuda docs ingest ./docs
```

The command reports how many files it discovered, how many it processed, how many were empty (no matching lines), and how many claims it extracted. The counts add up: `discovered = processed + empty + skipped`.

### Ingest from a workspace config

If you want to ingest documentation from multiple directories, create `.garuda/workspace.yaml`:

```yaml
workspace: my-first-workspace
tenant_id: 00000000-0000-0000-0000-000000000001
documents:
  - path: ./docs
  - path: ../shared-docs
  - path: /absolute/path/to/adr
```

Then:

```bash
./bin/garuda docs sync
```

### Verify claims against code

```bash
./bin/garuda docs verify
```

This correlates every ingested claim against the semantic graph. A claim that matches an entity or relationship in the code becomes `SUPPORTED`. A claim that does not match stays `UNVERIFIED` with the reason. A claim that contradicts runtime behavior becomes `CONTRADICTED`.

Only `CALLS` predicates are verified against the graph. Other predicates remain `UNVERIFIED` by design — the current verification engine only knows how to check call edges.

---

## Verifying what you've built

### Check the ledger

```bash
./bin/garuda status
```

Reports the current Merkle block height, the current root hash, and the verification state distribution across the workspace.

### Inspect a single entity

```bash
./bin/garuda inspect <entity-name>
```

Reports the entity's package, kind, outgoing relationships, and incoming relationships. If the entity was matched by any documentation claims, those are listed too.

### Check the impact of a change

```bash
./bin/garuda impact <entity-name>
```

Reports every caller of the entity, transitively — the full blast radius. On a workspace with tens of thousands of edges, the response comes back in under a second.

One caveat, and it is important: call edges are sparse at the leaf. Many functions have exactly one caller, and 90% have four or fewer. This is a property of the code, not a limitation of the analyzer. What you get from `garuda impact` is the truth about what the graph contains, including the truth that the graph does not contain everything.

### Compare two snapshots

```bash
./bin/garuda diff before.json after.json
```

Semantic diff, not text diff. Reports which entities were added, removed, or changed, and which relationships changed. Requires both snapshots to have been produced by `garuda analyze`.

---

## The MCP server in detail

The MCP server is the primary integration point for AI agents. It exposes sixteen tools over stdio, using line-delimited JSON-RPC 2.0.

### The sixteen tools

| Tool | What it answers |
| :--- | :--- |
| `garuda.entities` | List semantic entities in the workspace, filterable by package and kind |
| `garuda.inspect` | Inspect one entity with its incoming and outgoing relationships |
| `garuda.neighbors` | Every entity connected to the subject in one hop, both directions |
| `garuda.subclasses` | Who inherits from or embeds the subject |
| `garuda.implementers` | Who implements the subject interface |
| `garuda.find_entity` | Find entities by name pattern, kind, package, or file path |
| `garuda.policy.list` | List policies registered for the tenant |
| `garuda.policy.evaluate` | Dry-run the policy engine and return the decisions it would make |
| `garuda.governance.status` | Aggregate governance state: active policies, documentation health, contradiction count |
| `garuda.check_drift` | Documentation-to-code drift report |
| `garuda.query_claims` | Query document claims matching a subject |
| `garuda.query` | Query the Garuda knowledge graph with natural language |
| `garuda.get_lineage` | Get the full lineage of a decision |
| `garuda.get_impact` | Find what breaks if a decision is changed |
| `garuda.detect_contradictions` | Detect unresolved contradictions in the tenant knowledge graph |
| `garuda.propose_decision` | Propose a new decision (creates a draft) |

Read-only tools — everything except `garuda.propose_decision` — do not modify the workspace. `garuda.propose_decision` writes a draft decision, and budget-checks before it does.

### Verifying spec compliance

The repository includes a reference MCP client at [`scripts/mcp_verify.py`](../scripts/mcp_verify.py). It runs 18 checks against any MCP server that speaks line-delimited JSON-RPC 2.0 over stdio. Run it:

```bash
WORKSPACE=my-first-workspace python3 scripts/mcp_verify.py
```

If you are building your own MCP server, run it against yours. It takes about ten seconds and has caught real bugs — including one in Garuda itself, where the server was responding to JSON-RPC notifications that the spec forbids a response to.

---

## Running in CI

Garuda integrates into CI in two modes: as a gate that blocks PRs, and as a reporting step that annotates them.

### Block PRs that violate policy

```bash
./bin/garuda policy evaluate ./policies --fail-on BLOCK
```

`--fail-on` accepts `BLOCK`, `REVIEW`, `WARN`, or `ALLOW`. The command exits non-zero if any evaluation produces a decision at that level or higher. Wire this into your CI pipeline and any change that would violate a policy stops before merge.

### Annotate PRs with impact

```bash
./bin/garuda impact <changed-symbol> --format github
```

The `--format github` flag produces output in the annotation format that GitHub Actions understands. Every caller of the changed symbol becomes a check annotation on the PR.

### Semantic diff between commits

```bash
git stash
./bin/garuda analyze . --save -o /tmp/before.json
git stash pop
./bin/garuda analyze . --save -o /tmp/after.json
./bin/garuda diff /tmp/before.json /tmp/after.json
```

The diff output names every entity and relationship that changed, in structured JSON. What you do with it depends on your pipeline — comment on the PR, fail the build, or write it to an artifact.

---

## Troubleshooting

### `garuda dev` fails with "bind: address already in use"

Something is already listening on port 8080. Either stop it, or run Garuda on a different port:

```bash
./bin/garuda dev --port 8081
```

The port flag is honored by every command that starts a server.

### `garuda analyze` finds zero entities

Three possibilities, in order of likelihood:

1. You are not in a Go module directory. Check for `go.mod` in the current directory or any parent.
2. The module has a build error. Run `go build ./...` and fix any compile errors first — the analyzer uses the same type resolution the compiler does and will fail on the same code.
3. The module has no exported entities. Garuda indexes exported and referenced symbols; a package that contains only unexported helpers will produce few entities.

### The MCP server connects but tools return empty results

The most common cause is a workspace mismatch. Check that:

- `GARUDA_WORKSPACE` in the MCP configuration matches the workspace you actually analyzed.
- The `DATABASE_URL` in the MCP configuration matches the one you used for `garuda analyze`.

Run `psql "$DATABASE_URL" -c "SELECT id, name FROM workspaces;"` to see which workspaces exist, and compare the ID against what the MCP server is resolving.

### `garuda policy verify` reports a proof failure

A proof failure means one of three things:

1. The database row was manually edited after the proof was written. This is the failure case the ledger is designed to detect.
2. The tenant's root hash has been reset — usually by dropping and recreating the workspace.
3. The proof format has changed between the version that wrote it and the version that is reading it. Check `verification_version` on the row; the current format is version 1.

If none of those apply and the proof still fails, file an issue with the evaluation ID and the block height.

### The daemon runs but the dashboard is empty

Check that `garuda analyze --save` actually wrote to the database. `garuda summary` should report non-zero counts. If it reports zero, the analysis either did not run with `--save` or was pointed at a different workspace.

---

## Command reference

Every command Garuda ships. Run `garuda <command> --help` for the full flag list.

### Analysis

| Command | Purpose |
| :--- | :--- |
| `garuda analyze <path>` | Analyze a repository |
| `garuda diff <before> <after>` | Semantic diff between two snapshots |
| `garuda inspect <entity>` | Inspect one entity with its relationships |
| `garuda entities` | List entities in the workspace |
| `garuda impact <entity>` | Blast-radius analysis for a symbol |
| `garuda summary` | Workspace architectural summary |

### Workspaces and repositories

| Command | Purpose |
| :--- | :--- |
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
| :--- | :--- |
| `garuda policy list` | List active policies |
| `garuda policy validate <dir>` | Parse and validate policy YAML |
| `garuda policy evaluate <dir>` | Run policies and anchor decisions |
| `garuda policy show <id>` | Show one evaluation with evidence and Merkle proof |
| `garuda policy verify <id>` | Re-derive the Merkle inclusion proof |

### Documentation

| Command | Purpose |
| :--- | :--- |
| `garuda docs ingest <path>` | Extract claims from documents |
| `garuda docs sync` | Ingest all sources declared in `.garuda/workspace.yaml` |
| `garuda docs status` | Show freshness of configured document sources |
| `garuda docs verify` | Correlate claims against AST entities and detect drift |

### Servers and integrations

| Command | Purpose |
| :--- | :--- |
| `garuda dev` | Start the unified daemon (API, worker, dashboard) |
| `garuda mcp` | Print the path to the MCP server binary |
| `garuda dashboard` | Open the web dashboard |

### Verification

| Command | Purpose |
| :--- | :--- |
| `garuda status` | Inspect Merkle root and daemon status |
| `garuda verify` | Verify ledger integrity |
| `garuda bench` | Run the GAP-20 grounding benchmark |
| `garuda ci` | Run in CI mode with baseline comparison |
| `garuda judge <baseline> <proposed>` | Governance judgement between snapshots |

---

## Getting help

- **Issues and bug reports:** [GitHub Issues](https://github.com/myshra777-ai/garuda/issues)
- **Design discussions:** [GitHub Discussions](https://github.com/myshra777-ai/garuda/discussions)
- **Architecture decisions:** [`docs/adr/`](../docs/adr/)
- **Evidence and benchmarks:** [`EVIDENCE.md`](../EVIDENCE.md)
```
````

---

