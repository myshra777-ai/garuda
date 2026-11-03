# Installing Garuda

Garuda runs on **Windows 10/11, macOS, Linux, ChromeOS (Linux container) and Android (Termux)**. One command installs everything into your home folder. No `sudo` for the default path, and no administrator rights, except that on a Windows PC without the Microsoft Visual C++ runtime, Windows asks once for permission to install it.

- [Quick start](#quick-start)
- [Install with an AI assistant](docs/INSTALL_WITH_AI.md) (ChatGPT, Claude, Codex, Cursor)
- [What you need first](#what-you-need-first)
- [What the installer does](#what-the-installer-does)
- [Per-OS notes](#per-os-notes) (Windows, macOS, Linux, ChromeOS, Android)
- [Installing the dashboard as an app (PWA)](#installing-the-dashboard-as-an-app-pwa)
- [Settings you can change](#settings-you-can-change)
- [Other ways to install](#other-ways-to-install) (Docker, release archives, package managers)
- [Check that it works](#check-that-it-works)
- [Connect an AI agent](#connect-an-ai-agent)
- [Update, stop, uninstall](#update-stop-uninstall)
- [Troubleshooting](#troubleshooting)
- [What has been tested](#what-has-been-tested)

## Quick start

**macOS, Linux, ChromeOS (Linux), Android (Termux):**

```bash
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash
```

**Windows 10 / 11 (PowerShell 5.1 or 7):**

```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force; irm https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.ps1 | iex
```

`Set-ExecutionPolicy ... -Scope Process` only affects that one PowerShell window.

The installer clones Garuda into `~/garuda` (`%USERPROFILE%\garuda` on Windows), builds it, starts it and prints the dashboard address and your login. To read the script before running it, download `install.sh` / `install.ps1` first, open it, then run it.

Already cloned the repository? Run the setup script from inside it:

```bash
./setup.sh                                                   # macOS, Linux, ChromeOS, Termux
```

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\setup.ps1   # Windows
```

You can pass extra repositories to index: `./setup.sh ~/code/service-a ~/code/service-b`.

**Just want to try it?** `npx garuda-guard demo` needs only Node 18 or newer and no database.
**Want an AI assistant to do it?** Use the [copy-and-paste prompt](docs/INSTALL_WITH_AI.md).

## What you need first

The installer downloads Go and PostgreSQL itself, so you do **not** need those, a C compiler, Homebrew or Docker. It does need a few basic tools and an internet connection.

| OS | Needed before you start | Install it with |
| :--- | :--- | :--- |
| **Windows 10 (1803+) / 11** | PowerShell 5.1 or 7 (already included). `git` is optional; without it the installer downloads a zip. | nothing to do |
| **macOS** (Intel or Apple Silicon) | `git`, `curl`, `tar`, `unzip` (all ship with the Xcode command line tools) | `xcode-select --install` |
| **Debian / Ubuntu / ChromeOS Linux** | `git`, `curl`, `xz-utils`, `unzip` | `sudo apt-get install -y git curl xz-utils unzip` |
| **Fedora / RHEL** | `git`, `curl`, `xz`, `unzip` | `sudo dnf install -y git curl xz unzip` |
| **Arch** | `git`, `curl`, `xz`, `unzip` | `sudo pacman -S git curl xz unzip` |
| **Android (Termux)** | `git`, `curl` (PostgreSQL and Go are installed with `pkg`) | `pkg install -y git curl` |

Also needed:

- A 64-bit CPU (x86-64 or ARM64).
- About 1 GB of free disk space (Go toolchain, PostgreSQL, build cache).
- A regular user account. **PostgreSQL refuses to run as root**, so on Linux do not run the installer with `sudo`. In a root-only container, create a normal user first.
- Python 3 is optional; it is only used for the MCP self-check at the end of setup.

## What the installer does

1. **Gets the code.** `install.sh` / `install.ps1` clone the repository (shallow). An existing install is updated only if it is a clean Garuda checkout; the installer never deletes a folder it does not recognize and never discards local edits or local commits.
2. **PostgreSQL 16.** Uses `DATABASE_URL` if you set it. Otherwise uses a PostgreSQL already listening on `127.0.0.1:5433`, otherwise downloads a portable PostgreSQL 16 into `~/.garuda/runtime/postgres` and starts a private cluster in `~/.garuda/data/pg` on `127.0.0.1:5433` (database `garuda_test`, role `test`, local connections only). Fallbacks: a `garuda-postgres` Docker container if Docker is running (Windows prefers Docker when it is already running), `pkg install postgresql` on Termux, or your system package manager as a last resort.
3. **Go.** Uses the Go on your `PATH` if it is 1.21 or newer (newer Go versions fetch the exact toolchain named in `go.mod` on their own). Otherwise downloads a matching portable Go into `~/.garuda/toolchain/go`.
4. **Build.** Compiles `garuda`, `garuda-api`, `garuda-mcp` and `garuda-talon` into `./bin` with `CGO_ENABLED=0`, so no C compiler is needed. It also copies them into `~/.local/bin` and `~/go/bin` if those folders already exist.
5. **Initialize.** `garuda init` applies the SQL migrations, writes `.env` (database URL, Ed25519 JWT keys, a random `GARUDA_CONTROL_TOKEN`), creates the owner account and workspace, and indexes the repository.
6. **Talon.** `garuda talon init` writes the agent contract and hooks (`.garuda/TALON_CONTRACT.md`, `CLAUDE.md`, `AGENTS.md`, `.cursor/`, `.windsurfrules`, and a git `pre-commit` hook).
7. **Start and check.** Starts `garuda-api` on `127.0.0.1:8080`, runs `garuda doctor`, and (if Python 3 is present) `scripts/mcp_verify.py`.

Files outside the repository folder all live under `~/.garuda`:

| Path | Contents |
| :--- | :--- |
| `~/.garuda/runtime/postgres` | portable PostgreSQL binaries |
| `~/.garuda/data/pg` | the PostgreSQL data directory |
| `~/.garuda/data/run` | PostgreSQL unix socket, API pid file |
| `~/.garuda/toolchain/go` | portable Go (only if none was found) |
| `~/.garuda/owner-login` (`owner-login.txt` on Windows) | the generated owner password, if you did not set one |

**Login.** The installer never uses a default password. Set `GARUDA_PASSWORD` (8+ characters) before running to choose your own; otherwise a random one is generated, printed once and saved to the file above. The dashboard is at `http://localhost:8080/login`.

**Network exposure.** The API listens on `127.0.0.1` only. Nothing is reachable from other machines unless you set `GARUDA_BIND` yourself.

## Per-OS notes

### Windows 10 / 11

- Works in Windows PowerShell 5.1 and PowerShell 7. The scripts are plain ASCII on purpose, so 5.1 does not mangle them.
- PostgreSQL comes from the EnterpriseDB portable zip (about 320 MB, downloaded once). It needs the **Microsoft Visual C++ 2015-2022 Redistributable (x64)**. `setup.ps1` checks for it and installs it for you (Windows asks for permission; click Yes). If you decline, or the automatic install fails, install it yourself from <https://aka.ms/vs/17/release/vc_redist.x64.exe>. If `initdb` fails with exit code `-1073741515` or `0xC0000135`, install it from <https://aka.ms/vs/17/release/vc_redist.x64.exe> and run the installer again.
- If Docker Desktop is already running, the installer uses a `garuda-postgres` container instead of the portable PostgreSQL.
- Windows Defender or SmartScreen may scan the downloads; the first run can take a few minutes.
- Windows on ARM: the installer builds native ARM64 Garuda binaries, but the PostgreSQL zip is x64 and runs under Windows' x64 emulation.
- Git Bash / MSYS2: `install.sh` detects it and hands over to `setup.ps1`.
- If port 8080 is busy: `.\setup.ps1 -Port 8081`.

### macOS

- Apple Silicon and Intel are both supported. PostgreSQL binaries from a Homebrew install (`postgresql@15` to `@17`) are reused to create the private cluster; otherwise the portable PostgreSQL 16 is downloaded, so there is no long Homebrew compile.
- Everything is bound to `127.0.0.1`. The setup script uses IPv4 on purpose, because `localhost` can resolve to IPv6 `::1` where nothing is listening.
- If macOS asks whether to allow incoming connections for `postgres`/`garuda-api`, you can deny it; both only listen on loopback.
- Gatekeeper does not block the binaries the installer builds locally. Binaries downloaded from the Releases page are **not code-signed or notarized**; clear the quarantine flag with `xattr -d com.apple.quarantine <file>` if macOS refuses to open them.

### Linux

- Debian/Ubuntu, Fedora/RHEL and Arch are covered by the prerequisites table above. WSL2 works like Linux.
- Do not run as root; PostgreSQL will refuse to start.
- If you get `initdb: invalid locale settings`, the installer already forces `LC_ALL=C` for the cluster. If you still see it, check that `xz`/`xz-utils` is installed.

### ChromeOS (Chromebook)

ChromeOS runs Garuda in its **Linux development environment** (Crostini, Debian). Turn it on in *Settings → Advanced → Developers → Linux development environment*, open the Terminal app, then follow the Debian/Ubuntu steps:

```bash
sudo apt-get update && sudo apt-get install -y git curl xz-utils unzip
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash
```

Open `http://localhost:8080/login` in Chrome on the Chromebook; ChromeOS forwards `localhost` from the Linux container. ChromeOS devices with ARM CPUs are supported (ARM64 builds).

### Android (Termux)

Install Termux from **F-Droid** (the Play Store build is outdated and no longer updated), then:

```bash
pkg update && pkg upgrade -y
pkg install -y git curl
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash
```

- On Termux the installer uses `pkg install postgresql golang` rather than downloading portable binaries.
- Open `http://localhost:8080/login` in Chrome on the same phone. Keep Termux running (use *Acquire wakelock* in the notification) or Android may stop it in the background.
- Building takes a few minutes on a phone. If the build fails, run `pkg upgrade -y golang` and re-run the installer.
- Android 12+ can kill background processes (the "phantom process killer"). If PostgreSQL stops on its own, see the Termux wiki for disabling it.

## Installing the dashboard as an app (PWA)

The dashboard is a Progressive Web App (`/manifest.json`, `/sw.js`). Open `http://localhost:8080/login` in Chrome, Edge, Brave or Chromium and choose *Install Garuda* from the address-bar icon or the browser menu. On Android Chrome use *Add to Home screen* / *Install app*.

Notes:

- `localhost` counts as a secure context, so installation works over plain `http://localhost`. If you expose Garuda on another hostname you need HTTPS for the install prompt.
- The installed app is only a window onto your local server. The Garuda server (`garuda-api`) and PostgreSQL must be running.
- Safari on macOS and iOS can add the page to the Dock/Home Screen, but does not offer the Chromium install prompt.

## Settings you can change

Set these in your shell **before** running the installer or `setup.sh`.

| Variable | Default | Purpose |
| :--- | :--- | :--- |
| `GARUDA_PASSWORD` | random, saved to `~/.garuda/owner-login` | Owner password (8+ characters) |
| `GARUDA_EMAIL` | `git config user.email` | Owner email |
| `GARUDA_WORKSPACE` | folder name | Workspace name |
| `GARUDA_PORT` (`-Port` on Windows) | `8080` | Dashboard port |
| `DATABASE_URL` | local cluster on `127.0.0.1:5433` | Use your own PostgreSQL instead (13 or newer, with the `pgcrypto` extension available) |
| `GARUDA_INSTALL_DIR` | `~/garuda` | Where `install.sh` / `install.ps1` clone the repository |
| `GARUDA_BIND` | loopback only | Address `garuda-api` listens on. Use `0.0.0.0:8080` only if you add TLS and access control |
| `GARUDA_MCP_READ_ONLY` | unset | `true` blocks every mutating MCP tool |
| `GARUDA_MCP_APPROVER` | unset | Identity recorded for human approvals made over MCP |

`.env` in the repository folder is written by `garuda init` and read by every Garuda binary. The setup scripts read it without executing it, and anything you export in your shell wins over it.

## Other ways to install

### Docker

The published image runs the API only; you still provide PostgreSQL. For a local trial with both:

```bash
POSTGRES_PASSWORD='choose-one' docker compose up -d --build
# dashboard: http://localhost:8080  (both ports are bound to 127.0.0.1)
```

For production-style setups see `deploy/compose/docker-compose.prod.yml` and `deploy/k8s/garuda-stack.yaml`. Images are published to `ghcr.io/myshra777-ai/garuda` (API) and `ghcr.io/myshra777-ai/garuda-mcp` (MCP server, stdio) with a tag per release.

### Prebuilt release archives

Each [GitHub release](https://github.com/myshra777-ai/garuda/releases) attaches static binaries (no C library needed) plus `SHA256SUMS.txt`. File names follow `garuda-<tag>-<os>-<arch>` where `<tag>` is the release tag, for example `garuda-vX.Y.Z-linux-amd64.tar.gz`:

| Platform | Archive |
| :--- | :--- |
| Windows x64 / ARM64 | `garuda-<tag>-windows-amd64.zip`, `...-windows-arm64.zip` |
| macOS Apple Silicon / Intel | `garuda-<tag>-darwin-arm64.tar.gz`, `...-darwin-amd64.tar.gz` |
| Linux x64 / ARM64 (also ChromeOS Linux) | `garuda-<tag>-linux-amd64.tar.gz`, `...-linux-arm64.tar.gz`, plus `.deb` packages |
| Android Termux (ARM64) | `garuda-<tag>-android-arm64-termux.tar.gz` (a static Linux ARM64 build) |
| FreeBSD x64 | `garuda-<tag>-freebsd-amd64.tar.gz` |

Verify before running:

```bash
sha256sum -c --ignore-missing SHA256SUMS.txt      # Linux
shasum -a 256 -c --ignore-missing SHA256SUMS.txt  # macOS
```

```powershell
(Get-FileHash .\garuda-<tag>-windows-amd64.zip -Algorithm SHA256).Hash   # compare with SHA256SUMS.txt
```

The binaries still need a PostgreSQL database (`DATABASE_URL`). The simplest way to get one is the setup script in the source archive.

### Package managers

The files in `packaging/` (Homebrew, winget, Chocolatey, AUR, RPM) are **templates that are not yet published** to those package repositories, so `brew install garuda` or `winget install garuda` will not work yet. `scripts/render_packaging.py` fills in the download checksums from a release's `SHA256SUMS.txt` so the manifests can be submitted.

## Check that it works

```bash
./bin/garuda doctor                      # platform, .env, database, migrations, binaries, hooks, API
./bin/garuda doctor --fix                # repair what it can
python3 scripts/mcp_verify.py            # MCP protocol self-check against your workspace
```

On Windows use `.\bin\garuda.exe doctor`. Then open `http://localhost:8080/login` and sign in with the owner email and password.

## Connect an AI agent

`garuda talon init` already wrote the config for Cursor, Claude Code, Codex and Windsurf in the repository you indexed. To register the MCP server by hand, point your agent at `garuda-mcp` over stdio ([example config](.mcp.json)):

```json
{
  "mcpServers": {
    "garuda": {
      "command": "garuda-mcp",
      "env": {
        "DATABASE_URL": "postgres://test:test@127.0.0.1:5433/garuda_test?sslmode=disable",
        "GARUDA_WORKSPACE": "default"
      }
    }
  }
}
```

Use the path `./bin/garuda-mcp` (or an absolute path) if the binary is not on your `PATH`, and use your own `DATABASE_URL` if you set one. Claude Code and Codex users can also install the plugin: see the README for `/plugin marketplace add myshra777-ai/garuda`.

## Update, stop, uninstall

**Update:** run the installer again. It first checks that your checkout has no tracked local changes and no local commits, then updates to `origin/main` and rebuilds. If either check fails it stops with a message and changes nothing.

**Stop:**

```bash
./bin/garuda status                              # what is running
kill "$(cat ~/.garuda/data/garuda-api.pid)"      # stop the API (Linux/macOS/Termux)
~/.garuda/runtime/postgres/bin/pg_ctl -D ~/.garuda/data/pg stop   # stop the private PostgreSQL
```

On Windows, stop `garuda-api.exe` and `postgres.exe` from Task Manager, or run `docker stop garuda-postgres` if you used Docker.

**Uninstall:** stop the processes above, then delete the repository folder (`~/garuda`) and `~/.garuda`. If you set up Docker, also `docker rm -f garuda-postgres` and `docker volume rm garuda-pgdata`. Remove copies of the binaries in `~/.local/bin` and `~/go/bin` if you want a clean slate. `garuda talon init` added hook files to the repositories you indexed (`.garuda/`, `CLAUDE.md`, `AGENTS.md`, `.cursor/`, `.windsurfrules`, `.git/hooks/pre-commit`); delete those there if you no longer want them.

## Troubleshooting

| Symptom | Cause and fix |
| :--- | :--- |
| `PostgreSQL cannot be started by the root user` | Run the script as a normal user (not with `sudo`). |
| `'<tool>' is required but was not found` | Install the tool named in the message (see [What you need first](#what-you-need-first)) and re-run. |
| `connection refused` to `[::1]:5433` | `localhost` resolved to IPv6. Use `127.0.0.1` in `DATABASE_URL`; `garuda doctor --fix` rewrites it. |
| `port 8080 is already in use` | Another program owns the port. Choose another: `GARUDA_PORT=8081 ./setup.sh` (`-Port 8081` on Windows). |
| `initdb: invalid locale settings` | Make sure `xz`/`xz-utils` is installed; the script forces `LC_ALL=C` for the cluster. |
| `PostgreSQL did not start` | Read `~/.garuda/data/postgres.log`. A stale `postmaster.pid` from a crash is cleaned up automatically on the next run. |
| Windows: `blocked by your organization's Device Guard policy` / exit code `4551` | Windows security (Smart App Control, Device Guard or a work or school policy) refuses programs that are not code-signed, and Garuda's are not signed yet. Use WSL2 (`wsl --install -d Ubuntu`, then the Linux steps), or on a personal PC turn Smart App Control off (Windows cannot turn it back on without reinstalling), or ask your IT team to allow the `bin` folder. |
| Windows: `initdb` exit code `-1073741515` / `0xC0000135` | Install the VC++ 2015-2022 Redistributable (x64), see above. |
| Windows: `running scripts is disabled on this system` | Use the one-liner with `Set-ExecutionPolicy Bypass -Scope Process -Force`, or `powershell -ExecutionPolicy Bypass -File .\setup.ps1`. |
| Windows: `Unexpected token` / `Missing closing '}'` in `setup.ps1` | Your copy was saved with a different encoding. Re-download it or `git checkout setup.ps1`; the file must stay plain ASCII. |
| `Tracked local changes found ... no files were changed` | The installer refuses to overwrite your edits. Commit or stash them, then re-run. |
| `Install target ... is not a Garuda installation` | `GARUDA_INSTALL_DIR` points at a folder that has other files in it. Pick an empty folder; the installer will not delete anything. |
| `Could not download Go` / PostgreSQL download fails | Network, proxy or firewall. Install Go (<https://go.dev/dl/>) and/or point `DATABASE_URL` at an existing PostgreSQL, then re-run. |
| Termux: build fails | `pkg upgrade -y golang` and run the installer again. |
| Login says the password is wrong | Use the password in `~/.garuda/owner-login` (`%USERPROFILE%\.garuda\owner-login.txt`), or the value you set in `GARUDA_PASSWORD` **the first time** you ran setup. `garuda init` never overwrites an existing user's password. |
| `scripts/mcp_verify.py` fails on one check | Make sure `./bin/garuda-mcp` is built and the workspace has been indexed (`./bin/garuda analyze . --workspace <name> --save`), then re-run it and open an issue with the output. |

Still stuck? Open an issue at <https://github.com/myshra777-ai/garuda/issues> with the output of `./bin/garuda doctor --json` (look it over for secrets such as database URLs before you post it).

## What has been tested

Being upfront about what is and is not covered by automated or manual runs:

| Area | Status |
| :--- | :--- |
| Linux (x86-64) installer end to end with a user-space PostgreSQL, reruns, update guards, port/pid handling | Exercised manually, including with stubbed binaries for the Go build and API |
| `setup.ps1` / `install.ps1` | Parsed with real PowerShell 7 and checked for ASCII-only content; contract strings enforced by the `windows-installer-smoke` workflow. The first manual run on one Windows 11 PC found and fixed four problems (missing Visual C++ runtime, PostgreSQL start hang, API health probe, programs blocked by Smart App Control). **A clean re-run has not been recorded yet.** |
| macOS (Intel and Apple Silicon) | Logic reviewed against Homebrew and macOS conventions; **not yet run end to end** |
| ChromeOS (Linux/Crostini) | Same code path as Debian; **not yet run on a Chromebook** |
| Android (Termux) | Uses `pkg` packages; **not yet run on a device** |
| Download sources (go.dev, Maven Central, EnterpriseDB) | URLs follow the vendors' published layouts; availability depends on your network |

The `installer-e2e` workflow (manual, in GitHub Actions) runs the full installer on Ubuntu, macOS and Windows runners and uploads the logs. Run it from the Actions tab to produce a current result for your commit.
