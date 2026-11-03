# @myshra777-ai/garuda

Change control for AI-written code. This package downloads the verified Garuda release binary for your system on first run, then runs it.

```bash
npx @myshra777-ai/garuda demo        # see what Garuda catches, no database needed
npx @myshra777-ai/garuda --help
npx -y -p @myshra777-ai/garuda garuda-mcp   # the MCP server, for Claude Code, Cursor, Codex
```

- Needs Node 18 or newer and the `tar` command (included with Windows 10+, macOS and Linux).
- The first run downloads one archive from GitHub Releases and checks its SHA-256 against `SHA256SUMS.txt`. It refuses to run if the check fails.
- Most commands that store decisions need PostgreSQL. To set everything up, including a private database, follow [INSTALL.md](https://github.com/myshra777-ai/garuda/blob/main/INSTALL.md).

Source and docs: <https://github.com/myshra777-ai/garuda>. Apache 2.0.
