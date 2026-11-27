# Garuda v1.9.12

**Release date:** 2026-11-28 · **License:** Apache-2.0 · [Full changelog](CHANGELOG.md) · [Install guide](INSTALL.md)

v1.9.12 makes `garuda check` say how much of a change it actually analyzed and which code the change can affect, and fixes bugs found while adding unit tests to packages that had none.

## Highlights

- **Evidence coverage.** In diff mode, `garuda check --json` and `garuda talon guard --json` include `evidence_coverage`: changed files analyzed out of the total, the files that were not analyzed and why, and a `COMPLETE` / `PARTIAL` / `INCONCLUSIVE` verdict. An `ALLOW` on a partly analyzed change is no longer indistinguishable from a fully checked one. The terminal output shows a `Coverage:` line.
- **Impacted entities.** `garuda check --json` lists what a change can affect, most severe first, with repository-relative file paths (capped at 200), so tools can tell which teams own the affected code.
- **Migrations embedded in the binary.** Installs that keep the SQL files elsewhere (for example Homebrew) no longer start with an empty schema.
- **Deterministic impact receipts.** The same change now always produces the same `evidence_root`.
- **Document ingestion fixed.** `.txt` and `.rst` are read as the CLI promised, DOCX files are parsed properly, and PDFs are skipped with a clear message instead of being read as raw bytes.
- **Fewer false matches.** ADR rules match modal words only as whole words, and runtime correlation no longer claims full confidence without a namespace.
- **`check --diff` affected count.** A diff that touches no indexed symbols reports 0 affected entities instead of the whole workspace.

## Downloads

Attached to this release (`<tag>` is `v1.9.12`):

| Platform | File |
| :--- | :--- |
| Windows x64 / ARM64 | `garuda-<tag>-windows-amd64.zip`, `garuda-<tag>-windows-arm64.zip` |
| macOS Apple Silicon / Intel | `garuda-<tag>-darwin-arm64.tar.gz`, `garuda-<tag>-darwin-amd64.tar.gz` |
| Linux x64 / ARM64 (and ChromeOS Linux) | `garuda-<tag>-linux-amd64.tar.gz`, `garuda-<tag>-linux-arm64.tar.gz`, `garuda_1.9.12_amd64.deb`, `garuda_1.9.12_arm64.deb` |
| Android (Termux, ARM64) | `garuda-<tag>-android-arm64-termux.tar.gz` |
| FreeBSD x64 | `garuda-<tag>-freebsd-amd64.tar.gz` |
| Source plus installers | `garuda-<tag>-universal-source.tar.gz`, `install.sh`, `install.ps1`, `setup.sh`, `setup.ps1` |
| Checksums | `SHA256SUMS.txt` |
| Package-manager manifests with checksums filled in | `packaging/` (Homebrew, AUR, winget, RPM, Chocolatey; templates, not yet published to those repositories) |

Containers: `ghcr.io/myshra777-ai/garuda:v1.9.12` (API and dashboard) and `ghcr.io/myshra777-ai/garuda-mcp:1.9.12` (MCP server over stdio).

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
- **Benchmarks and MCP checks** in [EVIDENCE.md](EVIDENCE.md) were recorded on an earlier release line; the analyzers did not change in this release.
- **New in this release:** unit tests cover the impact engine, graph export, runtime correlation, document ingestion, the ADR parser, evidence coverage and the impacted-entities list. Contract tests pin the keys of `check --json` and `talon guard --json`. The database-backed parts of the runtime verifier, knowledge evaluator and graph service are covered by the existing integration tests in CI, not by new ones.
- Package-manager manifests (Homebrew, winget, Chocolatey, AUR, RPM) are templates and have not been submitted to those repositories.

## Known limitations

- The MCP server needs a PostgreSQL database. Directory users must run the installer first; a database-free demo mode is not part of this release.
- The Python, TypeScript, and Java (Beta) analyzers resolve cross-file imports, barrel re-exports, `tsconfig.json` path aliases, constructor fields, inheritance (`super()`), and parameter/return type annotations at `0.95` confidence (`AST_EXACT`), but do not run external language servers on unannotated dynamic duck-typing (see [SECURITY.md](SECURITY.md) and [docs/CAPABILITIES.md](docs/CAPABILITIES.md)).
- The Merkle ledger is anchored locally; there is no external timestamp authority.
