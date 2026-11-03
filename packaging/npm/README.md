# garuda-guard

Guardrails for AI coding agents. Garuda gives Claude Code, Cursor and Codex the context before they edit, checks the change against your architecture rules, blocks what breaks them, and keeps a tamper-evident record. This package downloads the verified Garuda release binary for your system on first run, then runs it.

```bash
npx garuda-guard demo                 # see what Garuda catches, no database needed
npx garuda-guard --help
npx -y -p garuda-guard garuda-mcp     # the MCP server, for Claude Code, Cursor, Codex
```

- Needs Node 18 or newer and the `tar` command (included with Windows 10+, macOS and Linux).
- The first run downloads one archive from GitHub Releases and checks its SHA-256 against `SHA256SUMS.txt`. It refuses to run if the check fails.
- Most commands that store decisions need PostgreSQL. To set everything up, including a private database, follow [INSTALL.md](https://github.com/myshra777-ai/garuda/blob/main/INSTALL.md).
- On Windows PCs with Smart App Control or a work policy, Windows can refuse the unsigned program. See the troubleshooting table in INSTALL.md.

Source and docs: <https://github.com/myshra777-ai/garuda>. Apache 2.0.
