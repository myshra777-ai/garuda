# Garuda Universal Installation & Troubleshooting Guide (`v1.8.1`)

**Zero pre-installed dependencies required.** Whether you are a Product Manager, Founder, Designer, Architect, or Software Engineer on **Windows 10/11 (PowerShell)**, **macOS (Apple Silicon / Intel)**, **Linux**, **Chromebook (ChromiumOS)**, or **Android**, Garuda automatically manages and provisions Go, PostgreSQL, schema migrations, Active MCP, and the Garuda Talon Harness in **1 or 2 commands**—without requiring GCC/C compilers, Xcode Command Line Tools, Docker Desktop, or Administrator privileges.

## 1. One-Command Setup by Operating System

### Windows 10 / 11 (PowerShell — Zero Dependencies Required)

Open **Windows PowerShell** (you do **not** need Administrator privileges, Go, PostgreSQL, GCC/MinGW, or Docker installed) and run:

```powershell
# Option A: 1-Command Remote Install & Launch (even if Git is not installed!)
Set-ExecutionPolicy Bypass -Scope Process -Force; irm https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.ps1 | iex
```

Or if you have already cloned or downloaded the repository:

```powershell
# Option B: 2-Command Local Setup inside the repository
cd $HOME\garuda
powershell -ExecutionPolicy Bypass -File .\setup.ps1
```

**What `setup.ps1` does automatically for you:**
1. **Zero-Admin Portable PostgreSQL (`:5433`)**: Uses Docker if running, or automatically downloads Portable PostgreSQL 16 into `$HOME\.garuda\runtime\postgres`, initializes a local database cluster in `$HOME\.garuda\data\pg`, and starts it on port `5433`.
2. **Zero-Admin Portable Go & Pure-Go Build (`CGO_ENABLED=0`)**: Uses system Go if present, or downloads Portable Go 1.24 into `$HOME\.garuda\toolchain\go` and compiles `garuda.exe`, `garuda-api.exe`, `garuda-mcp.exe`, and `garuda-talon.exe` in 100% Pure-Go mode (no GCC/MinGW C compiler needed).
3. **1-Click Onboarding**: Runs migrations, creates your owner account, wires **Cursor** (`.cursor\mcp.json`) and **Claude Desktop** (`%APPDATA%\Claude\claude_desktop_config.json`), activates **Garuda Talon**, starts the gateway, and opens `http://localhost:8080/login` in your browser.

---

### macOS (Apple Silicon M1/M2/M3/M4 & Intel — Zero Dependencies Required)

Open **Terminal** (you do **not** need Xcode Command Line Tools, Go, Docker, or PostgreSQL pre-installed) and run:

```bash
# Option A: 1-Command Remote Install & Launch
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash
```

Or inside a cloned repository:

```bash
# Option B: 2-Command Local Setup
git clone https://github.com/myshra777-ai/garuda.git && cd garuda
./setup.sh
```

**What `setup.sh` does automatically on macOS:**
1. **No Xcode CLI / GCC Required**: Automatically detects if `clang`/`gcc` is installed; if not, compiles in 100% Pure-Go mode (`CGO_ENABLED=0`) so macOS never blocks you with an Xcode Command Line Tools download popup.
2. **Portable Go Toolchain**: Downloads official `darwin-arm64` or `darwin-amd64` Go into `~/.garuda/toolchain/go` if `go` is not installed (no `sudo` required).
3. **Automatic PostgreSQL**: Uses Docker if running, or initializes a local user-space PostgreSQL cluster in `~/.garuda/data/pg` on port `5433`.
4. **Auto-Wires Claude Desktop & Cursor**: Writes `~/Library/Application Support/Claude/claude_desktop_config.json` and `.cursor/mcp.json`.

---

### Linux (Ubuntu, Debian, Fedora, Arch, RHEL, WSL2)

```bash
# Option A: 1-Command Remote Install & Launch
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash

# Option B: Inside cloned repository
git clone https://github.com/myshra777-ai/garuda.git && cd garuda && ./setup.sh
```

---

### Chromebook / ChromiumOS (Crostini Linux + 1-Click Chromium App)

1. Open the ChromeOS **Linux Terminal** (Settings -> Advanced -> Developers -> Linux development environment) and run:
   ```bash
   curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash
   ```
2. Open **Google Chrome / Chromium** to `http://localhost:8080/login` and sign in.
3. Click the **"Install Garuda"** icon in the right side of the Chromium address bar to install Garuda as a standalone ChromeOS shelf application.

---

### Android (Termux Local Runtime + Android Chrome PWA)

- **Option 1 — Install the Garuda Mobile App (PWA) on Android Chrome**:
  Open your team or laptop Garuda URL in **Android Chrome**, tap the browser menu (`⋮`), and tap **"Add to Home screen" / "Install app"**. The 6-surface Software Intelligence Instrument renders in a responsive standalone mobile layout.
- **Option 2 — Run the Full Garuda CLI, MCP & Talon Engine Locally on Android (Termux)**:
  Open **Termux** on Android and run:
  ```bash
  pkg update -y && pkg install -y curl git golang postgresql
  curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash
  ```

---

## 2. Instant Self-Diagnosis & Auto-Repair (`garuda doctor`)

At any time, you can verify your installation and automatically repair missing `.env` keys, stopped PostgreSQL instances, missing binaries, or unlinked IDE hooks with a single command:

```bash
# macOS / Linux / Chromebook / Android
./bin/garuda doctor --fix

# Windows PowerShell
.\bin\garuda.exe doctor --fix
```

---

## 3. Troubleshooting & Installation Error Reference

| # | Symptom or Error Message | Why It Happens | 1-Command Fix |
|---|---|---|---|
| **1** | **Windows PowerShell**: `setup.ps1 cannot be loaded because running scripts is disabled on this system (PSSecurityException)` | Windows PowerShell blocks unsigned local scripts by default in `Restricted` mode. | Run with process-scoped bypass:<br>`powershell -ExecutionPolicy Bypass -File .\setup.ps1` |
| **2** | **`go: command not found`** or **`'go' is not recognized as an internal or external command`** | Go is not installed in your system `PATH` when running manual `go build` commands. | Run `./setup.sh` (or `.\setup.ps1` on Windows), which automatically installs Portable Go into `~/.garuda/toolchain/go`, or add it to your shell:<br>`export PATH="$HOME/.garuda/toolchain/go/bin:$PATH"` |
| **3** | **`cgo: C compiler "gcc" not found`** or **`xcrun: error: invalid active developer path`** | You ran `go build` manually on a machine without GCC, MinGW-w64, or macOS Xcode CLI Tools. | Compile in Pure-Go mode (`CGO_ENABLED=0`):<br>**Bash**: `CGO_ENABLED=0 ./setup.sh`<br>**PowerShell**: `$env:CGO_ENABLED="0"; .\setup.ps1` |
| **4** | **`could not connect to PostgreSQL (connection refused on 127.0.0.1:5433)`** | The local PostgreSQL cluster or Docker container stopped after a system reboot. | Run automatic recovery:<br>**Bash**: `./bin/garuda doctor --fix` (or `./setup.sh`)<br>**PowerShell**: `.\bin\garuda.exe doctor --fix` (or `.\setup.ps1`) |
| **5** | **`bind: address already in use (:8080)`** | Another process or previous `garuda-api` instance is already listening on port `8080`. | Re-run `./setup.sh` / `.\setup.ps1` (which automatically frees port `8080`), or run on a custom port:<br>`GARUDA_PORT=8090 ./setup.sh` |
| **6** | **Windows SmartScreen / Defender popup**: `Windows protected your PC` when running `garuda.exe` | Locally compiled or freshly downloaded `.exe` binaries do not yet have a commercial EV code-signing certificate. | Click **More info** -> **Run anyway**, or unblock via PowerShell:<br>`Get-ChildItem .\bin\*.exe \| Unblock-File` |
| **7** | **Cursor or Claude Desktop does not show `garuda` MCP tools** | IDE was already open before `garuda init` wrote `.cursor/mcp.json` or `claude_desktop_config.json`. | Run `./bin/garuda talon init .`, then press `Cmd+Shift+P` / `Ctrl+Shift+P` -> **Developer: Reload Window** in Cursor (or restart Claude Desktop). |
| **8** | **Git commit blocked by `GARUDA TALON ACTIVE GUARD`** | Your staged changes introduced a net-new policy violation or ADR contradiction, or your workspace index was out of sync. | Inspect the exact violation and repair instructions:<br>`./bin/garuda talon guard --staged`<br>To re-sync the workspace baseline first:<br>`./bin/garuda analyze . --workspace default --save` |
| **9** | **`git commit -m "..."` says `Changes not staged for commit`** | Modified files (such as `README.md`) were edited after `git add`, or `git commit` was run without `git add -A`. | Stage all changes before committing:<br>`git add -A && git commit -m "your message" && git push origin main` |
| **10** | **`git push origin main` fails with `! [rejected] main -> main (non-fast-forward)`** | Remote `origin/main` has commits that have not yet been pulled into your local branch. | Rebase your local commit onto `origin/main` and push:<br>`git pull --rebase origin main && git push origin main` |
| **11** | **Database migration error (`dirty database version`)** | A previous migration was interrupted mid-execution. | Reset and re-apply idempotent schema migrations:<br>`./bin/garuda init --skip-scan` |
| **12** | **Chromebook Crostini**: Browser cannot reach `localhost:8080` | ChromeOS Linux container port forwarding was paused. | Open ChromeOS Settings -> Developers -> Linux -> **Port forwarding** and enable TCP port `8080`, or visit `http://penguin.linux.test:8080/login`. |
| **13** | **Android Termux**: `Permission denied` when accessing `/sdcard` repos | Termux has not yet been granted shared storage access. | Run `termux-setup-storage` inside Termux and keep Git repositories inside `$HOME/` for native POSIX permissions. |
