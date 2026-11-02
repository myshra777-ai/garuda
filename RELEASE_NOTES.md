# Garuda v1.9.10

**Release date:** 2026-11-03 · **License:** Apache-2.0 · [Full changelog](CHANGELOG.md) · [Install guide](INSTALL.md)

v1.9.10 is a reliability and release-readiness update. It rewrites the installers, corrects the documentation against the code, hardens the shipped container and Kubernetes files, and puts the plugin and MCP-registry manifests into the repository so Garuda can be listed in the marketplaces.

## Highlights

- **Installers rewritten and hardened** (`install.sh`, `setup.sh`, `install.ps1`, `setup.ps1`). No default password or token is ever used; a random owner password is generated (or take yours from `GARUDA_PASSWORD`) and a random `GARUDA_CONTROL_TOKEN` goes into `.env`. Safe updates: a folder that is not a Garuda checkout is never touched, and local edits or commits are never discarded. Better messages for missing tools, busy ports, root users, a missing Visual C++ runtime on Windows and failed downloads. The unsafe duplicate `scripts/install.sh` is gone.
- **Documentation corrected against the code.** One accurate [INSTALL.md](INSTALL.md) for Windows, macOS, Linux, ChromeOS and Android (Termux); wrong tool and migration counts fixed everywhere (54 tools, 91 migrations); broken links and screenshots fixed; duplicate documents removed; `SECURITY.md` merged from two conflicting policies into one; example credentials removed. CI now fails when docs, counts or links drift.
- **Marketplace-ready repository.** The plugin manifests for Claude Code, Codex and Cursor, `.mcp.json`, `server.json` (MCP Registry), `glama.json` and `smithery.yaml` are tracked. `.gitignore` no longer has a global `*.json` rule that kept them off GitHub. A CI check keeps it that way.
- **One release pipeline.** Pushing a version tag verifies the repository, builds archives for Windows, macOS, Linux, Android (Termux) and FreeBSD, writes `SHA256SUMS.txt`, attests build provenance, and creates a **draft** GitHub release. Container images go to GHCR; publishing the release publishes `server.json` to the MCP Registry.
- **Safer deployment files.** `docker-compose.yml` no longer exposes PostgreSQL with a known password on all interfaces or ships a guessable secret; the Kubernetes manifest points at an image that is actually published; the API image builds with the Go version `go.mod` requires.

## Downloads

Attached to this release (`<tag>` is `v1.9.10`):

| Platform | File |
| :--- | :--- |
| Windows x64 / ARM64 | `garuda-<tag>-windows-amd64.zip`, `garuda-<tag>-windows-arm64.zip` |
| macOS Apple Silicon / Intel | `garuda-<tag>-darwin-arm64.tar.gz`, `garuda-<tag>-darwin-amd64.tar.gz` |
| Linux x64 / ARM64 (and ChromeOS Linux) | `garuda-<tag>-linux-amd64.tar.gz`, `garuda-<tag>-linux-arm64.tar.gz`, `garuda_1.9.10_amd64.deb`, `garuda_1.9.10_arm64.deb` |
| Android (Termux, ARM64) | `garuda-<tag>-android-arm64-termux.tar.gz` |
| FreeBSD x64 | `garuda-<tag>-freebsd-amd64.tar.gz` |
| Source plus installers | `garuda-<tag>-universal-source.tar.gz`, `install.sh`, `install.ps1`, `setup.sh`, `setup.ps1` |
| Checksums | `SHA256SUMS.txt` |
| Package-manager manifests with checksums filled in | `packaging/` (Homebrew, AUR, winget, RPM, Chocolatey; templates, not yet published to those repositories) |

Containers: `ghcr.io/myshra777-ai/garuda:v1.9.10` (API and dashboard) and `ghcr.io/myshra777-ai/garuda-mcp:1.9.10` (MCP server over stdio).

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
