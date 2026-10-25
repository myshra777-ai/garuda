# Garuda Talon

**Talon brings architecture checks into the coding-agent workflow.** It connects Garuda's indexed software model and active policies to supported pre-edit, post-edit, Git-hook, MCP and CLI-supervisor integration points.

A passive MCP server answers when an agent asks. Talon adds a workflow around that context: brief the agent before a change, check the change after editing, return repair-oriented feedback when blocked, and optionally supervise a CLI agent for a bounded number of attempts.

Talon is not a universal sandbox and does not guarantee that every agent action can be intercepted. Use the CLI/CI gate as an explicit enforcement boundary and test the behavior of every configured hook.

## The lifecycle

1. **Compile the contract.** Talon combines ingested workspace document claims, YAML policies and selected high-impact symbols into a Markdown contract.
2. **Ground the agent.** A pre-edit briefing provides matching symbols, graph-visible impact, relevant ADR guardrails and active policies.
3. **Check the edit.** The guard can sync dirty source files, map the Git diff to touched entities, evaluate diff-scoped policies and ADR claims, and report net-new findings separately from baseline findings.
4. **Repair or review.** A blocking result includes instructions intended to help the agent fix the change. The supervisor can pass those instructions back to a CLI coding agent for a bounded number of attempts.
5. **Keep a record.** Guard results can be persisted and sealed in the Merkle ledger. The ledger proves integrity/inclusion of recorded information—not that the analysis was complete.

## Setup

Run initialization after Garuda has a database and a workspace. Review generated changes before using them in a development repository.

```bash
./bin/garuda talon init . \
  --workspace my-app \
  --policies ./policies
```

The installer can generate or update integration files for supported clients and Git hooks, including:

- `.garuda/TALON_CONTRACT.md` — compiled workspace contract.
- `.cursor/rules/garuda-talon.mdc` — Cursor rule.
- `.cursor/mcp.json` — MCP configuration with Talon enforcement enabled.
- `CLAUDE.md` and `.claude/settings.json` — Claude Code instructions and configured edit hook.
- `AGENTS.md` and `.windsurfrules` — agent-tooling contract sections.
- `.git/hooks/pre-commit` and `commit-msg` — local check and governance-trailer hooks when a Git hooks directory is available.

The exact files created depend on the repository and client environment. Inspect them after installation, especially if they overlap with existing rules or hooks. Re-run the installer after reviewing your active policy/ADR contract when you intentionally want to refresh the generated content.

## Commands

### 1. Print the current contract

```bash
./bin/garuda talon contract . \
  --workspace my-app \
  --policies ./policies
```

Use this to inspect what the agent-facing contract actually contains. It is based on the workspace claims and policy directory available to Talon; it cannot infer undocumented rules that are not present in those inputs.

### 2. Get pre-edit context

```bash
./bin/garuda talon brief PaymentService \
  --workspace my-app \
  --policies ./policies
```

The briefing can return matched symbols, their graph-visible blast radius, applicable ADR guardrails, active policy rules, and the protocol for checking the change after editing.

The same pre-edit operation is exposed to connected agents as `garuda.talon_pre_edit`.

### 3. Guard a change

Check staged edits:

```bash
./bin/garuda talon guard . \
  --workspace my-app \
  --staged \
  --fail-on-block
```

Check a revision or diff:

```bash
./bin/garuda talon guard . \
  --workspace my-app \
  --diff HEAD \
  --fail-on-block
```

Check a supplied file list:

```bash
./bin/garuda talon guard . \
  --workspace my-app \
  --files internal/handlers/payment.go,internal/service/payment.go \
  --fail-on-block
```

The guard reports the touched files/symbols, impact estimates, policy/ADR findings, net-new and baseline counts, outcome and any Merkle block height. A `BLOCK` outcome exits with code `2` when `--fail-on-block` is used. Hook mode also exits `2` on a blocking verdict and writes the repair prompt to stderr.

The default command enables incremental indexing, but the exact behavior depends on repository state and the selected mode. Inspect the guard output and confirm the indexed workspace corresponds to the checkout being checked. Generated files, unsupported syntax and unresolved symbols may not map cleanly to graph entities.

The agent-facing post-edit operation is `garuda.talon_verify_edit`.

### 4. Run a CLI agent under the supervisor

```bash
./bin/garuda talon run \
  --workspace my-app \
  --policies ./policies \
  --max-retries 3 \
  -- claude "Refactor PaymentService to use PaymentRepository"
```

Talon runs the given CLI agent command, checks the resulting working-tree diff, and can feed a structured self-repair prompt back to the agent. The retry count is bounded by `--max-retries`; it does not guarantee that a blocked change will be repaired or reach `ALLOW`. Review the final result and any file changes before committing.

## MCP integration

The MCP server exposes 23 tools, including:

- `garuda.talon_pre_edit` — targeted pre-edit context and guardrails.
- `garuda.talon_verify_edit` — post-edit sync/check and repair feedback.
- `garuda.architectural_contract` — a prompt that returns the compiled contract.
- `garuda://workspace/contract` — contract resource.
- `garuda://workspace/adr-guardrails` — structured ADR guardrail resource.

With `GARUDA_TALON_ENFORCE` enabled (default unless set to `false` or `0`), the server injects active contract information into tool responses. This improves the chance that the agent sees the rules, but response injection is not itself a hard security boundary: a client may ignore instructions, a tool may not be called, or a client configuration may be changed.

Verify the protocol and integration surface with:

```bash
WORKSPACE=my-app python3 scripts/mcp_verify.py
```

The current evidence record reports 30/30 reference checks. That number covers the verifier's specified protocol and integration checks; it is not a semantic accuracy score for all tools.

## CI is the merge-control boundary

Talon hooks help catch problems near the edit, but a team should still configure a CI job that runs against the exact proposed revision. For example:

```bash
./bin/garuda check \
  --project my-app \
  --diff HEAD \
  --policy ./policies \
  --save \
  --fail-on-block \
  --json
```

When a `BLOCK` verdict causes the job to fail, configure that job as a required status check in the repository's branch-protection/ruleset settings. Garuda does not configure branch protection for you.

Before relying on this path, validate that:
- The workspace is indexed from the candidate checkout or is incrementally synchronized to it.
- The policy directory is the one intended for the repository.
- Malformed policies, database errors, unavailable services and missing evidence behave as expected.
- The CI process propagates Garuda's exit code and the check is required for merging.
- A test violation fails the job, while a clean test case passes for the intended reasons.

## Operational boundaries

- Contracts only reflect information available from ingested document claims, active YAML policies and the indexed workspace.
- A zero violation count does not prove the graph is complete.
- Runtime checks only see telemetry that has been captured, ingested and correlated.
- A `SUPPORTED` / `UNVERIFIED` / `CONTRADICTED` label describes evidence state within the checked scope, not a universal truth certificate.
- Hooks may be bypassed or misconfigured. Use CI branch protection for the merge boundary.
- The supervisor is a bounded wrapper around a CLI command, not an isolated sandbox, independent reasoning model, or automatic approval system.
- Merkle proof verification can establish consistency/inclusion of recorded data against a supplied root and proof; it cannot prove semantic completeness or correctness.

## Related documentation

- [Quick start](QUICKSTART.md)
- [Capabilities](CAPABILITIES.md)
- [CLI, MCP and HTTP API](API.md)
- [End-to-end walkthrough](WALKTHROUGH.md)
- [Operational playbook](../PLAYBOOK.md)
- [Evidence and verification](../EVIDENCE.md)
