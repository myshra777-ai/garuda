# Quick start

This guide helps you get from checkout to a first Garuda analysis. Garuda's setup path uses local services and may alter Docker/process state, so read the setup script before running it.

## 1. Get the repository

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda
```

Review the setup script and configuration before execution:

```bash
less setup.sh
```

The bundled setup path is intended to bootstrap a local development environment. Do not run it unreviewed on a shared or production host.

## 2. Build or bootstrap

If you want to build the CLI and MCP server without running the setup script:

```bash
go build -o bin/garuda ./cmd/garuda
go build -o bin/garuda-mcp ./cmd/garuda-mcp
```

The full application uses PostgreSQL for persistent state. See [the API reference](API.md) for environment variables, database setup, migrations and the API server. The setup script may perform additional environment-specific initialization.

## 3. Analyze a repository

After configuring the database and workspace required by your checkout, analyze a target repository:

```bash
./bin/garuda analyze /path/to/repository --save
```

Consult `./bin/garuda analyze --help` and [the CLI reference](API.md) for supported flags. Analysis coverage depends on the language, syntax and dependencies that Garuda can resolve.

## 4. Check a change

From an initialized workspace, evaluate a Git diff against a configured policy directory. The exact project/workspace flags depend on the current CLI contract:

```bash
./bin/garuda check --diff HEAD --project my-system --policy ./policies --save --fail-on-block
```

Review the full output, including the evaluated scope, unresolved relationships, baseline findings and policy results. Do not treat a passing result as proof that every behavior was discovered.

## 5. Try Talon

Talon provides pre-edit context and post-edit checks for configured agent workflows:

```bash
./bin/garuda talon --help
./bin/garuda talon brief --help
./bin/garuda talon guard --help
```

Before relying on Talon as a mandatory gate, confirm the correct client hooks are installed and test that blocked edits, analysis failures and unavailable services behave as intended.

## Next steps

- [Capabilities and limitations](CAPABILITIES.md)
- [CLI, MCP and HTTP reference](API.md)
- [System specifications](SPECS.md)
- [Walkthrough](WALKTHROUGH.md)
- [Evidence and verification results](../EVIDENCE.md)
- [Security policy](../SECURITY.md)
