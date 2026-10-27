# Garuda Universal Installation & Troubleshooting Guide (`v1.8.4`)

**Zero pre-installed dependencies required.** Whether you are a Product Manager, Founder, Designer, Staff Architect, or Software Engineer on **Windows 10/11 (PowerShell 5.1 or PowerShell 7)**, **macOS (Apple Silicon M1/M2/M3/M4 & Intel `x86_64`)**, **Linux**, **Chromebook (ChromiumOS Crostini)**, or **Android (Termux & Chrome PWA)**, Garuda automatically provisions Portable Go `1.26.0`, Portable PostgreSQL `16.4`, idempotent schema migrations (`057_schema_hardening.sql`), Active MCP (`23` tools), and the **Garuda Talon** Active Agent Harness in **1 command**—without requiring GCC/C compilers, Xcode Command Line Tools, MinGW, Docker Desktop, or Administrator/`sudo` privileges.

---

## 1. One-Command Setup by Operating System

### Windows 10 / 11 (PowerShell 5.1 & PowerShell 7 — Zero Dependencies Required)

Open **Windows PowerShell** (you do **not** need Administrator privileges, Git, Go, PostgreSQL, GCC/MinGW, or Docker installed) and run:

```powershell
# Option A: 1-Command Remote Install & Launch (works even if Git is not installed!)
Set-ExecutionPolicy Bypass -Scope Process -Force; irm https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.ps1 | iex
```

Or if you have already cloned or downloaded the repository:

```powershell
# Option B: Local Setup inside an existing repository checkout
cd $HOME\garuda
git pull --rebase origin main
powershell -NoProfile -ExecutionPolicy Bypass -File .\setup.ps1
```

> **Using Git Bash / MSYS2 / Cygwin on Windows?** You can also run `./setup.sh` directly inside Git Bash—it automatically detects `msys*`/`mingw*`/`cygwin*` and delegates to `powershell.exe -NoProfile -ExecutionPolicy Bypass -File ./setup.ps1`.

**What `install.ps1` and `setup.ps1` (`v1.8.4`) do automatically on Windows:**
1. **7-Bit Pure-ASCII PowerShell 5.1 Compatibility**: Both `install.ps1` and `setup.ps1` are strictly 7-bit ASCII and suppress PowerShell byte-stream progress bars (`$ProgressPreference = "SilentlyContinue"`), preventing Windows PowerShell 5.1 (`CP1252` default encoding) from ever misinterpreting UTF-8 bytes as smart-quote string delimiters.
2. **Fast Shallow Clone (`--depth 1`) & Native `curl.exe` / `tar.exe` Acceleration**: Uses shallow Git clones (fetching only the latest commit instead of 180+ MB of pack history) and native Windows `System32\curl.exe` + `System32\tar.exe` for 10x faster archive downloads and extraction.
3. **Zero-Admin Portable PostgreSQL 16 (`127.0.0.1:5433`)**: Uses Docker if running, or automatically downloads Portable PostgreSQL 16.4 (`~55 MB`, one-time download) into `$HOME\.garuda\runtime\postgres`, initializes a locale-safe UTF-8 database cluster (`LC_ALL=C --no-locale`) in `$HOME\.garuda\data\pg`, and starts it on explicit IPv4 `127.0.0.1:5433`.
4. **Single-Download Portable Go `1.26.0` & 1-Pass Pure-Go Build (`CGO_ENABLED=0`)**: Downloads Portable Go `1.26.0` (matching `go.mod` so Go never downloads a second toolchain) into `$HOME\.garuda\toolchain\go` and compiles `garuda.exe`, `garuda-api.exe`, `garuda-mcp.exe`, and `garuda-talon.exe` in a single `go build -o bin\ ...` pass without needing GCC or MinGW.
5. **1-Click Onboarding & Active Harness**: Runs idempotent migrations (`057_schema_hardening.sql`), auto-discovers your email from `git config user.email`, wires **Cursor** (`.cursor\mcp.json`) and **Claude Desktop** (`%APPDATA%\Claude\claude_desktop_config.json`), activates **Garuda Talon**, starts the gateway, runs `garuda.exe doctor`, and opens `http://localhost:8080/login`.

---

### macOS (Apple Silicon M1/M2/M3/M4 & Intel `x86_64` — Zero Dependencies & Fast Setup)

Open **Terminal** (you do **not** need Xcode Command Line Tools, Homebrew, Go, Docker, or PostgreSQL pre-installed) and run:

```bash
# Option A: 1-Command Remote Install & Launch
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash
```

Or inside an existing cloned repository:

```bash
# Option B: Local Setup inside the repository
cd ~/garuda
git pull --rebase origin main
./setup.sh
```

**What `install.sh` and `setup.sh` (`v1.8.4`) do automatically on macOS:**
1. **3-Second Pre-Built Portable PostgreSQL 16 (`~18 MB`, Zero Homebrew Source Builds)**: If PostgreSQL is not running on `127.0.0.1:5433` and Docker is not active, `setup.sh` downloads pre-built Portable PostgreSQL `16.4.0` binaries (`darwin-arm64v8` or `darwin-amd64`, `~18 MB`) directly into `~/.garuda/runtime/postgres/bin` in seconds—completely avoiding 30–45 minute `brew install postgresql@16` C/C++ source compilations on Intel Macs.
2. **IPv4 `127.0.0.1:5433` Normalization & Dedicated Unix Socket Directory (`~/.garuda/data/run`)**: macOS resolves `localhost` to IPv6 `[::1]` first. Garuda normalizes all `DATABASE_URL` strings from `@localhost:` to `@127.0.0.1:`, binds PostgreSQL explicitly to `listen_addresses = '127.0.0.1'`, and places Unix sockets inside `~/.garuda/data/run` (`chmod 700`) with automatic self-healing re-initialization if a prior cluster directory was interrupted.
3. **Single-Download Portable Go `1.26.0` (Matching `go.mod`)**: Reads the exact Go version from `go.mod` (`1.26.0`) and downloads `go1.26.0.darwin-arm64.tar.gz` or `go1.26.0.darwin-amd64.tar.gz` into `~/.garuda/toolchain/go`, preventing Go from downloading a second toolchain during compilation.
4. **1-Pass Pure-Go Build (`CGO_ENABLED=0`) with Commit Build-Stamp Caching**: Compiles all 4 binaries in a single `go build -o bin/ ...` invocation (and skips rebuilding when `bin/.build_stamp` matches `HEAD`), never triggering the macOS Xcode Command Line Tools popup.
5. **Auto-Wires Claude Desktop, Cursor & Garuda Talon**: Writes `~/Library/Application Support/Claude/claude_desktop_config.json`, `.cursor/mcp.json`, `.garuda/TALON_CONTRACT.md`, and `.git/hooks/pre-commit`, then verifies all **25/25** dynamic MCP & Talon checks (`scripts/mcp_verify.py`).

---

### Linux (Ubuntu, Debian, Fedora, Arch, RHEL, WSL2)

```bash
# Option A: 1-Command Remote Install & Launch (Zero sudo required!)
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash

# Option B: Inside a cloned repository
git clone --depth 1 https://github.com/myshra777-ai/garuda.git && cd garuda && ./setup.sh
```

On Linux, `setup.sh` automatically provisions pre-built Portable PostgreSQL 16 (`linux-amd64` or `linux-arm64v8`) into `~/.garuda/runtime/postgres` and Portable Go `1.26.0` into `~/.garuda/toolchain/go` in user-space—no `sudo apt-get` or root privileges required.

---

### Chromebook / ChromiumOS (Crostini Linux + 1-Click Chromium App)

1. Open the ChromeOS **Linux Terminal** (`Settings -> Advanced -> Developers -> Linux development environment`) and run:
   ```bash
   curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash
   ```
2. Open **Google Chrome / Chromium** to `http://localhost:8080/login` and sign in with the credentials printed in your terminal.
3. Click the **"Install Garuda"** icon in the right side of the Chromium address bar to install Garuda as a standalone ChromeOS shelf application.

---

### Android (Termux Local Runtime + Android Chrome PWA)

- **Option 1 — Install the Garuda Mobile App (PWA) on Android Chrome**:
  Open your team or laptop Garuda URL in **Android Chrome**, tap the browser menu (`⋮`), and tap **"Add to Home screen" / "Install app"**. The 8-surface Software Intelligence Instrument renders in a responsive standalone mobile layout.
- **Option 2 — Run the Full Garuda CLI, MCP & Talon Engine Locally on Android (Termux)**:
  Open **Termux** on Android and run:
  ```bash
  pkg update -y && pkg install -y curl git golang postgresql
  curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash
  ```

---

## 2. Instant Self-Diagnosis & Auto-Repair (`garuda doctor`)

At any time, you can verify your installation and automatically repair missing `.env` keys, stopped PostgreSQL clusters (in `~/.garuda/data/pg` or Docker), pending database migrations, missing binaries, or unlinked IDE hooks with a single command:

```bash
# macOS / Linux / Chromebook / Android — Diagnose & Auto-Repair
./bin/garuda doctor --fix

# Windows PowerShell — Diagnose & Auto-Repair
.\bin\garuda.exe doctor --fix

# Emit machine-readable JSON diagnostic report
./bin/garuda doctor --json
```

**What `garuda doctor [--fix]` checks and repairs:**
1. **Platform & Runtime Environment**: Verifies `GOOS/GOARCH`, Crostini, or Android Termux environment.
2. **Environment File (`.env`)**: Ensures `DATABASE_URL` (`127.0.0.1:5433`), `GARUDA_WORKSPACE`, `GARUDA_TALON_ENFORCE=true`, and `PORT=8080` are persisted.
3. **PostgreSQL Database & Migrations**: Automatically restarts the local user-space PostgreSQL cluster in `~/.garuda/data/pg` (or Docker `garuda-postgres`) if port `5433` stopped after a reboot, and applies all migrations through `057_schema_hardening.sql`.
4. **Garuda Binaries (`CLI`, `API`, `MCP`, `Talon`)**: Verifies all 4 binaries exist in `bin/` and compiles any missing binary using system or Portable Go (`~/.garuda/toolchain/go`).
5. **Garuda Talon Active Harness & MCP Hooks**: Verifies `.garuda/TALON_CONTRACT.md`, `.cursor/mcp.json`, `.claude/settings.json`, `CLAUDE.md`, and `AGENTS.md`.
6. **Garuda API & Dashboard Gateway**: Verifies live HTTP reachability at `http://localhost:8080/health`.

---

## 3. Complete Troubleshooting & Installation Error Reference

| # | Symptom or Error Message | Why It Happens | 1-Command Fix |
|---|---|---|---|
| **1** | **Windows PowerShell 5.1**: `Unexpected token ... in expression or statement` or `Missing closing '}' in statement block` in `setup.ps1` | Older versions of `setup.ps1` contained multi-byte UTF-8 box-drawing/emoji bytes (`0x93`/`0x94`) that Windows PowerShell 5.1 (`CP1252` default encoding without BOM) parsed as smart-quote string delimiters. | Pull or re-run the `v1.8.4` pure 7-bit ASCII installer:<br>`Set-ExecutionPolicy Bypass -Scope Process -Force; irm https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.ps1 \| iex` |
| **2** | **Windows PowerShell**: `setup.ps1 cannot be loaded because running scripts is disabled on this system (PSSecurityException)` | Windows PowerShell blocks unsigned local `.ps1` scripts by default in `Restricted` mode. | Run with process-scoped execution policy bypass:<br>`powershell -NoProfile -ExecutionPolicy Bypass -File .\setup.ps1` |
| **3** | **Windows PowerShell**: `NativeCommandError` when `git clone`, `initdb`, or `docker` prints progress messages | PowerShell treats native command `stderr` progress output as a terminating exception when `$ErrorActionPreference = "Stop"`. | Upgraded in `v1.8.4` via `Invoke-QuietCmd` / `Invoke-LoggedCmd` (`cmd.exe /c`). Update with:<br>`git pull --rebase origin main; powershell -NoProfile -ExecutionPolicy Bypass -File .\setup.ps1` |
| **4** | **macOS / Linux**: `database migration failed: failed to connect to user=test database=garuda_test: [::1]:5433 (localhost): connection refused` | macOS resolves `localhost` to IPv6 `[::1]` first, or the user-space PostgreSQL cluster in `~/.garuda/data/pg` is not running. | Run `./setup.sh` or `./bin/garuda doctor --fix`. In `v1.8.4`, `DATABASE_URL` automatically normalizes `@localhost:` to IPv4 `@127.0.0.1:5433` and auto-starts `~/.garuda/data/pg`. |
| **5** | **macOS Setup Takes Too Long**: `brew install postgresql@16` compiling from source (30+ mins on Intel Mac) or Go downloading two toolchains (`go1.24` then `go1.26`) | Homebrew compiles C/C++ dependencies from source on older/Intel macOS releases, and Go downloads `go1.26.0` if bootstrapped with `go1.24.0`. | Fixed in `v1.8.4`: `setup.sh` downloads pre-built Portable PostgreSQL 16 (`~18 MB` in ~3s), downloads Portable Go `1.26.0` matching `go.mod`, and uses `git clone --depth 1`. Re-run `./setup.sh`. |
| **6** | **macOS / Linux PostgreSQL**: `initdb: invalid locale settings` or Unix socket permission failure in `/tmp` | Non-standard shell `LC_ALL`/`LANG` locale variables or restricted `/tmp` permissions block `initdb`/`pg_ctl`. | `setup.sh` (`v1.8.4`) runs `LC_ALL=C LANG=C initdb --no-locale -E UTF8` and isolates sockets in `~/.garuda/data/run` (`chmod 700`). Run `./setup.sh` to auto-heal. |
| **7** | **`go: command not found`** or **`'go' is not recognized as an internal or external command`** | Go is not installed in your global system `PATH` when running manual `go build` commands. | Run `./setup.sh` (or `.\setup.ps1`), or add Portable Go to your current shell:<br>**Bash/Zsh**: `export PATH="$HOME/.garuda/toolchain/go/bin:$PATH"`<br>**PowerShell**: `$env:PATH = "$HOME\.garuda\toolchain\go\bin;$env:PATH"` |
| **8** | **`cgo: C compiler "gcc" not found`** or **`xcrun: error: invalid active developer path`** | You ran `go build` manually on a machine without GCC, MinGW-w64, or macOS Xcode Command Line Tools. | Compile in Pure-Go mode (`CGO_ENABLED=0` is default in `setup.sh` & `setup.ps1`):<br>**Bash**: `CGO_ENABLED=0 go build -o bin/ ./cmd/...`<br>**PowerShell**: `$env:CGO_ENABLED="0"; go build -o bin\ .\cmd\...` |
| **9** | **`bind: address already in use (:8080)`** | Another process or previous `garuda-api` instance is already listening on port `8080`. | Re-run `./setup.sh` / `.\setup.ps1` (which automatically frees port `8080`), or choose a custom port:<br>`GARUDA_PORT=8090 ./setup.sh` |
| **10** | **Windows SmartScreen / Defender popup**: `Windows protected your PC` when running `garuda.exe` | Locally compiled `.exe` binaries do not carry an external commercial EV code-signing certificate. | Click **More info** -> **Run anyway**, or unblock in PowerShell:<br>`Get-ChildItem .\bin\*.exe \| Unblock-File` |
| **11** | **Cursor or Claude Desktop does not show `garuda` MCP tools** | The IDE was already open before `garuda init` wrote `.cursor/mcp.json` or `claude_desktop_config.json`. | Run `./bin/garuda talon init .`, then press `Cmd+Shift+P` / `Ctrl+Shift+P` -> **Developer: Reload Window** in Cursor (or restart Claude Desktop). |
| **12** | **`python3 scripts/mcp_verify.py` fails on `garuda.implementers` or `garuda.subclasses`** | Earlier verifier versions queried hardcoded sample names (`CmdTyper`) absent when indexing the Garuda repo itself. | Fixed in `v1.8.4` (`cmd/garuda-mcp/tools_graph.go` smart `resolveTopClaimTarget` fallback + dynamic `scripts/mcp_verify.py`). Pull latest and run `python3 scripts/mcp_verify.py` (`25/25` passed). |
| **13** | **`⚠️ Environment drift detected (N difference(s) between receipt and active environment)`** | Your active terminal `GOOS/GOARCH`, `CGO_ENABLED`, or `HEAD` commit differs from the last indexed workspace graph receipt. | Run the exact `Reproduce:` command printed in the banner, or re-sync the workspace baseline for your current machine:<br>`./bin/garuda analyze . --workspace default --save` |
| **14** | **Git commit blocked by `GARUDA TALON ACTIVE GUARD`** | Your staged changes introduced a net-new policy violation or ADR contradiction. | Inspect the exact violation and copy-ready repair prompt:<br>`./bin/garuda talon guard --staged`<br>To refresh baseline first: `./bin/garuda analyze . --workspace default --save` |
| **15** | **`bash: install_garuda_v1.8.x.sh: No such file or directory`** | The downloaded updater script is in `~/Downloads` rather than `~/garuda`. | Run it from `~/Downloads` inside your repo:<br>`cd ~/garuda && bash ~/Downloads/<script_name>.sh` |
| **16** | **`git commit -m "..."` says `Changes not staged for commit` (e.g. `modified: go.mod / go.sum`)** | Files were modified after `git add`, or `go build` touched `go.mod`/`go.sum`. | Stage all intended changes before committing:<br>`git add -A && git commit -m "your message"` |
| **17** | **`git push origin main` fails with `! [rejected] main -> main (non-fast-forward)`** | Remote `origin/main` has commits that have not yet been pulled into your local branch. | Stash any dirty `go.mod`/`go.sum`, rebase onto `origin/main`, and push:<br>`git checkout -- go.mod go.sum 2>/dev/null; git pull --rebase origin main && git push origin main` |
| **18** | **Database migration error (`dirty database version` or missing `canonical_name` / `relationships` column)** | A previous migration was interrupted or an older schema was initialized before `057_schema_hardening.sql`. | Run automatic schema repair:<br>`./bin/garuda doctor --fix` (or `./bin/garuda init --skip-scan`) |
| **19** | **Chromebook Crostini**: Browser cannot reach `localhost:8080` | ChromeOS Linux container port forwarding was paused after sleep. | Open ChromeOS `Settings -> Developers -> Linux -> Port forwarding` and enable TCP port `8080`, or open `http://penguin.linux.test:8080/login`. |
| **20** | **Android Termux**: `Permission denied` when accessing `/sdcard` repositories | Termux has not yet been granted shared storage permissions, or `/sdcard` mounts `noexec`. | Run `termux-setup-storage` inside Termux and keep Git repositories inside `$HOME/` (`~/garuda`) for native POSIX permissions. |
