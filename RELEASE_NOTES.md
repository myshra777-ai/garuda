# Garuda v1.9.11

**Release date:** 2026-11-03 · **License:** Apache-2.0 · [Full changelog](CHANGELOG.md) · [Install guide](INSTALL.md)

v1.9.11 makes Garuda easier to try and fixes an overstated demo. Run `npx garuda-guard demo` (Node 18 or newer, no database) to see it work.

## Highlights

- **Honest `garuda demo`.** The demo now runs the real analyzer on the sample service and prints its real symbol, edge and timing numbers. The fake Merkle root hash is gone. The three planted problems are still listed from a fixed list, not detected; the demo says so.
- **One-command try-it: `npx garuda-guard demo`.** A small Node wrapper downloads the release archive for your system, checks it against `SHA256SUMS.txt`, caches it and runs it. `garuda-guard` also provides the `garuda` and `garuda-mcp` commands.
- **Windows setup fixes.** `setup.ps1` installs the Visual C++ runtime when it is missing, starts PostgreSQL reliably, checks the API over IPv4, and explains error 4551 when Windows blocks the unsigned binary (options: WSL2, Smart App Control, or asking IT to allow the folder).
- **Install with an AI assistant.** [docs/INSTALL_WITH_AI.md](docs/INSTALL_WITH_AI.md) has one copy-and-paste prompt for ChatGPT, Claude, Codex and Cursor, with a checklist for verifying what the assistant returns.
- **README and business case.** A rewritten README and a new [Why Garuda](docs/WHY_GARUDA.md) page for engineering leaders.

## Downloads

Attached to this release (`<tag>` is `v1.9.11`):

| Platform | File |
| :--- | :--- |
| Windows x64 / ARM64 | `garuda-<tag>-windows-amd64.zip`, `garuda-<tag>-windows-arm64.zip` |
| macOS Apple Silicon / Intel | `garuda-<tag>-darwin-arm64.tar.gz`, `garuda-<tag>-darwin-amd64.tar.gz` |
| Linux x64 / ARM64 (and ChromeOS Linux) | `garuda-<tag>-linux-amd64.tar.gz`, `garuda-<tag>-linux-arm64.tar.gz`, `garuda_1.9.11_amd64.deb`, `garuda_1.9.11_arm64.deb` |
| Android (Termux, ARM64) | `garuda-<tag>-android-arm64-termux.tar.gz` |
| FreeBSD x64 | `garuda-<tag>-freebsd-amd64.tar.gz` |
| Source plus installers | `garuda-<tag>-universal-source.tar.gz`, `install.sh`, `install.ps1`, `setup.sh`, `setup.ps1` |
| Checksums | `SHA256SUMS.txt` |
| Package-manager manifests with checksums filled in | `packaging/` (Homebrew, AUR, winget, RPM, Chocolatey; templates, not yet published to those repositories) |

Containers: `ghcr.io/myshra777-ai/garuda:v1.9.11` (API and dashboard) and `ghcr.io/myshra777-ai/garuda-mcp:1.9.11` (MCP server over stdio).

Binaries are static (`CGO_ENABLED=0`) but **not code-signed or notarized**. Verify them against `SHA256SUMS.txt` (see [INSTALL.md](INSTALL.md#prebuilt-release-archives)). Build provenance is attested with GitHub's artifact attestations: `gh attestation verify <file> --repo myshra777-ai/garuda`.

## Install

```bash
# macOS, Linux, ChromeOS (Linux), Android (Termux)
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash
```

```powershell
# Windows 10 / 11
irm https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.ps1 | iex
```

Claude Code: `/plugin marketplace add myshra777-ai/garuda`. Codex: `codex plugin marketplace add myshra777-ai/garuda`. Both need `garuda-mcp` and a database, so run the installer first.

## Upgrading

- Re-run the installer. It updates a clean checkout and rebuilds; it stops without changing anything if you have local edits or commits.
- Existing `.env` files and user accounts are kept. `garuda init` never overwrites an existing user's password.
- The MCP Registry entry and the Kubernetes manifest now use the image names `ghcr.io/myshra777-ai/garuda-mcp` and `ghcr.io/myshra777-ai/garuda`. If you referenced `ghcr.io/myshra777-ai/garuda-api`, switch to `ghcr.io/myshra777-ai/garuda`.
- If you use `docker-compose.yml`, set `POSTGRES_PASSWORD` (it is now required) and note that ports are bound to `127.0.0.1`.

## What has and has not been verified

- **Verified:** the Linux user-space install path end to end (fresh run, re-run, update guards, port and pid handling); PowerShell scripts parse under PowerShell 7 and stay ASCII; shell scripts pass `bash -n`; documentation links, counts and release facts are checked by `scripts/sync_docs_from_manifest.py --check` and `scripts/check_doc_links.py`; marketplace files are checked by `scripts/check_marketplace_manifests.py`.
- **Not yet recorded:** a full install on Windows, macOS, ChromeOS and Android (Termux). The manual **Installer end-to-end** workflow runs Ubuntu, macOS and Windows runners and produces that evidence.
- **Benchmarks and MCP checks** in [EVIDENCE.md](EVIDENCE.md) were recorded on the previous release line; the analyzers did not change in this release.
- Package-manager manifests (Homebrew, winget, Chocolatey, AUR, RPM) are templates and have not been submitted to those repositories.

## Known limitations

- The MCP server needs a PostgreSQL database. Directory users must run the installer first; a database-free demo mode is not part of this release.
- The Python and TypeScript analyzers are structural, not full type resolution (see [SECURITY.md](SECURITY.md) and [docs/CAPABILITIES.md](docs/CAPABILITIES.md)).
- The Merkle ledger is anchored locally; there is no external timestamp authority.
