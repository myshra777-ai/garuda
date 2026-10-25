# Quick start

Go from source code to an architecture-aware agent check. Garuda uses PostgreSQL for persistent workspace state, and Garuda Talon connects that state to supported AI-agent and Git workflows.

## 1. Get the repository

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda
```

The repository includes a convenience setup script. **Review it before running**: it may start Docker services and stop conflicting local processes or containers.

```bash
less setup.sh
./setup.sh
```

The initialization path builds the binaries, initializes a workspace, configures MCP, and attempts to install the active Talon contract and supported hooks. Confirm the terminal output and inspect the generated files in your repository before relying on them.

## 2. Index the repository you want Garuda to check

```bash
./bin/garuda analyze /path/to/repository --workspace my-app --save
```

Use the same workspace name in your later commands. Analysis scope and resolution depend on the language and supported constructs. See [multi-language coverage](MULTI_LANGUAGE_ROADMAP.md).

## 3. Give the agent context before editing

```bash
./bin/garuda talon brief PaymentService \
  --workspace my-app --policies ./policies
```

The briefing can return matching entities, graph-visible blast radius, relevant ADR guardrails, active YAML policies, and a mandatory post-edit verification protocol.

To install or refresh the supported Talon integration files for a repository:

```bash
./bin/garuda talon init /path/to/repository \
  --workspace my-app --policies ./policies
```

See the [Talon guide](TALON.md) for supported Cursor, Claude Code, Codex, Windsurf and Git integration paths and their limits.

## 4. Check an edit

Check staged edits:

```bash
./bin/garuda talon guard /path/to/repository \
  --workspace my-app --staged --fail-on-block
```

Or check a Git revision/diff:

```bash
./bin/garuda talon guard /path/to/repository \
  --workspace my-app --diff HEAD --fail-on-block
```

Talon can incrementally re-index dirty source files, evaluate the changed scope against the active policy bundle and document claims, distinguish net-new violations from baseline findings, and return self-repair instructions for a blocking result. Review the output and confirm that the configured hooks and failure path behave as intended.

## 5. Run the change-control gate in CI

```bash
./bin/garuda check \
  --project my-app \
  --diff HEAD \
  --policy ./policies \
  --save \
  --fail-on-block \
  --json
```

With `--fail-on-block`, a `BLOCK` outcome exits with code `2`. A CI provider can use that exit status to fail the job; you must still configure the job as a required status check to prevent a protected branch from merging. The workspace/database must contain the intended repository state, and the check must be tested against the exact candidate revision before use as a release gate.

## 6. Verify the reference MCP/Talon surface

With the database and workspace configured:

```bash
WORKSPACE=my-app python3 scripts/mcp_verify.py
```

The current evidence record reports 30/30 reference verifier checks. This verifies the covered protocol, tool-discovery, and integration checks; it is not a proof that all semantic outputs are correct.

## Next steps

- [Garuda Talon guide](TALON.md)
- [Capabilities and limitations](CAPABILITIES.md)
- [CLI, MCP and HTTP reference](API.md)
- [System specifications](SPECS.md)
- [End-to-end walkthrough](WALKTHROUGH.md)
- [Current evidence and verification](../EVIDENCE.md)
- [Operational playbook](../PLAYBOOK.md)
- [Security policy](../SECURITY.md)
