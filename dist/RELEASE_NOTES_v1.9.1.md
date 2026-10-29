# Garuda `v1.9.1` — Official Release Notes & Multi-OS Package Downloads

**Release Date:** October 29, 2026  
**Release Tag:** `v1.9.1` (`Apache-2.0`)  
**Verification Status:** `24/24` Strict Compiler Benchmarks · `43/43` MCP Verification Checks (35 Tools) · `10/10` Runtime OTel Contradiction Quarantine · `CGO_ENABLED=0` Zero-Dependency All-OS Matrix

---

## 1. Executive Summary: What's New in Garuda `v1.9.1`

Garuda `v1.9.1` is our largest production milestone to date, evolving Garuda from an architectural change-control gate into a **complete Software & Financial Intelligence Instrument** with **35 deterministic MCP tools**, **Human-in-the-Loop multi-team formula disambiguation (`0.0–1.0` weights)**, **3-Tier Surface Criticality classification**, **continuous cross-team financial discrepancy auditing (48 Crore vs. 4.8 Crore)**, an **Auto-Synthesized Base Rulebook Studio**, and **pre-packaged standalone downloads for Windows, macOS, Linux, Chromebook, Android, and FreeBSD**.

### Five Headline Upgrades in `v1.9.1`

1. **Human-in-the-Loop Multi-Team Formula Disambiguation (`0.0–1.0` Weights)**:
   - When multiple teams implement undocumented competing formulas for the same business metric (e.g., `Loss Ratio` across **Ops**, **Marketing**, and **Management**), **Garuda never lets an AI agent silently guess or overwrite a formula**.
   - Dynamically computes **`0.00–1.00` team relevance weights** (`0.70` active team vs. `0.10`/`0.10`), ranks each formula by **Least Architectural Harm**, and presents plain-English blast radius impacts so a human stakeholder makes the final call (`garuda audit logic` / `garuda.disambiguate_logic`).
2. **3-Tier Surface Criticality Classification (Amazon Search vs. Wishlist Model)**:
   - **`HIGH_SEVERE` (Core Surfaces — Freeze / Do Not Touch)**: Homepage, Search Bar, Search Results Redirect Router (142 callers), and Claims Settlement Ledger. Frozen against automated AI modifications.
   - **`MEDIUM` (Non-Blocking Features — Human Trade-Off Review)**: Wishlist Loader and Promotional Badges (19 callers). Secondary features where trade-offs can be weighed.
   - **`LOW` (Minor UI & Formatting — Allow with Info)**: Search Box UI alignment and badge helpers (4 callers).
3. **Continuous Financial & Business Data Discrepancy Audit (48 Crore vs. 4.8 Crore)**:
   - Traces AST and SQL unit-scaling lineage across team pipelines to quarantine **10x Lakhs-to-Crore divisor bugs** (`48.00 Crore INR` in Claims Ops vs. `4.80 Crore INR` in Finance's executive export) and omitted reinsurance cessions *before* statutory financial audits occur (`garuda audit financial` / `garuda.audit_financial_discrepancy`).
4. **Auto-Synthesized Base Rulebook Studio**:
   - Reverse-engineers unorganized repositories into `.garuda/DRAFT_BASE_RULEBOOK.md`. Engineers, architects, and finance leads can review, edit inline in `/dashboard` (`04 — Docs & Base Rulebook`), and lock the baseline in one click (`garuda rulebook generate` / `garuda rulebook approve`).
5. **Extended 35-Tool Production MCP Server & Multi-OS Standalone Packages**:
   - Expanded `bin/garuda-mcp` from 26 to **35 zero-stub tools** (`43/43` checks passing in `scripts/mcp_verify.py`), upgraded the offline web console to **9 interactive surfaces**, and added pre-built standalone archives + package-manager manifests for **Windows (`x64` & `ARM64`), macOS (`Apple Silicon` & `Intel`), Linux (`x86_64` & `ARM64`, `.tar.gz`, `.deb`, `.rpm`, `AUR`), Chromebook (`Crostini` + `PWA`), Android (`Termux` + `PWA`), and FreeBSD**.

---

## 2. Direct Downloads & Packages by Operating System (`v1.9.1`)

Every release on the [**GitHub Releases Page (`releases/tag/v1.9.1`)**](https://github.com/myshra777-ai/garuda/releases/tag/v1.9.1) includes 100% standalone, zero-CGO (`CGO_ENABLED=0`) bundles containing all four binaries (`garuda`, `garuda-api`, `garuda-mcp`, `garuda-talon`), SQL migrations (`001`–`093`), starter YAML policies, and the 1-command setup scripts:

| Operating System & Architecture | Standalone Release Archive | Package Manager / Native Installer |
| :--- | :--- | :--- |
| **Windows 10 / 11 (64-bit x64)** | [`garuda-v1.9.1-windows-amd64.zip`](https://github.com/myshra777-ai/garuda/releases/download/v1.9.1/garuda-v1.9.1-windows-amd64.zip) | **PowerShell**: `irm https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.ps1 \| iex`<br>**Winget / Scoop / Choco**: `packaging/winget`, `packaging/scoop`, `packaging/choco` |
| **Windows 11 ARM64 (Snapdragon)** | [`garuda-v1.9.1-windows-arm64.zip`](https://github.com/myshra777-ai/garuda/releases/download/v1.9.1/garuda-v1.9.1-windows-arm64.zip) | **PowerShell**: `irm https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.ps1 \| iex` |
| **macOS Apple Silicon (M1/M2/M3/M4)** | [`garuda-v1.9.1-darwin-arm64.tar.gz`](https://github.com/myshra777-ai/garuda/releases/download/v1.9.1/garuda-v1.9.1-darwin-arm64.tar.gz) | **Homebrew**: `brew install --formula packaging/homebrew/garuda.rb`<br>**1-Command**: `curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh \| bash` |
| **macOS Intel (`x86_64`)** | [`garuda-v1.9.1-darwin-amd64.tar.gz`](https://github.com/myshra777-ai/garuda/releases/download/v1.9.1/garuda-v1.9.1-darwin-amd64.tar.gz) | **Homebrew**: `brew install --formula packaging/homebrew/garuda.rb`<br>**1-Command**: `curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh \| bash` |
| **Linux x86_64 (Ubuntu, Debian, Fedora, Arch, WSL2)** | [`garuda-v1.9.1-linux-amd64.tar.gz`](https://github.com/myshra777-ai/garuda/releases/download/v1.9.1/garuda-v1.9.1-linux-amd64.tar.gz) | **Debian/Ubuntu `.deb`**: [`garuda_1.9.1_amd64.deb`](https://github.com/myshra777-ai/garuda/releases/download/v1.9.1/garuda_1.9.1_amd64.deb) (`sudo dpkg -i garuda_1.9.1_amd64.deb`)<br>**RPM / Arch AUR**: `packaging/rpm/garuda.spec` · `packaging/aur/PKGBUILD` |
| **Linux ARM64 & Chromebook (Crostini)** | [`garuda-v1.9.1-linux-arm64.tar.gz`](https://github.com/myshra777-ai/garuda/releases/download/v1.9.1/garuda-v1.9.1-linux-arm64.tar.gz) | **Debian/Crostini `.deb`**: [`garuda_1.9.1_arm64.deb`](https://github.com/myshra777-ai/garuda/releases/download/v1.9.1/garuda_1.9.1_arm64.deb) + Chrome PWA (`/manifest.json`) |
| **Android (Termux ARM64 + Chrome PWA)** | [`garuda-v1.9.1-android-arm64-termux.tar.gz`](https://github.com/myshra777-ai/garuda/releases/download/v1.9.1/garuda-v1.9.1-android-arm64-termux.tar.gz) | **Termux**: `curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh \| bash` |
| **FreeBSD (`x86_64`)** | [`garuda-v1.9.1-freebsd-amd64.tar.gz`](https://github.com/myshra777-ai/garuda/releases/download/v1.9.1/garuda-v1.9.1-freebsd-amd64.tar.gz) | Extract `tar -xzf garuda-v1.9.1-freebsd-amd64.tar.gz` and run `./bin/garuda` |
| **Docker / OCI Container & K8s** | `ghcr.io/myshra777-ai/garuda-api:v1.9.1` | `docker compose -f deploy/compose/docker-compose.prod.yml up -d`<br>`kubectl apply -f deploy/k8s/garuda-stack.yaml` |
| **Checksums & Provenance** | [`SHA256SUMS.txt`](https://github.com/myshra777-ai/garuda/releases/download/v1.9.1/SHA256SUMS.txt) | Verify with `sha256sum -c SHA256SUMS.txt` or `shasum -a 256 -c SHA256SUMS.txt` |

---

## 3. Quick-Install Commands Across All Platforms

### Option A: 1-Command Zero-Dependency Universal Install (Recommended)
Automatically provisions Portable PostgreSQL `16.4` (`127.0.0.1:5433` with `pg_isready` WAL readiness polling), Portable Go `1.26.0`, database migrations, and all 6 IDE/Git hooks without needing `sudo` or Administrator rights:

```bash
# macOS (Apple Silicon & Intel), Linux, Chromebook & Android (Termux)
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash
```

```powershell
# Windows 10 / 11 (PowerShell 5.1 & PowerShell 7 — Pure ASCII, No Admin Required)
Set-ExecutionPolicy Bypass -Scope Process -Force; irm https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.ps1 | iex
```

### Option B: Download & Run Pre-Compiled Release Bundle Directly
If you prefer to download a pre-compiled binary archive from GitHub Releases without compiling from source:

```bash
# Example: macOS Apple Silicon (darwin-arm64) or Linux (linux-amd64)
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"; [[ "$ARCH" == "x86_64" ]] && ARCH="amd64"; [[ "$ARCH" == "aarch64" ]] && ARCH="arm64"
curl -fsSLO "https://github.com/myshra777-ai/garuda/releases/download/v1.9.1/garuda-v1.9.1-${OS}-${ARCH}.tar.gz"
tar -xzf "garuda-v1.9.1-${OS}-${ARCH}.tar.gz"
cd "garuda-v1.9.1-${OS}-${ARCH}" && ./setup.sh
```

```bash
# Example: Ubuntu / Debian / Chromebook Crostini (.deb package)
curl -fsSLO https://github.com/myshra777-ai/garuda/releases/download/v1.9.1/garuda_1.9.1_amd64.deb
sudo dpkg -i garuda_1.9.1_amd64.deb
garuda doctor --fix
```

---

## 4. Complete Release Evolution (`v1.1.0` → `v1.9.1`)

| Release | Date | Key Milestone & Capabilities |
| :--- | :--- | :--- |
| **`v1.9.1`** | **2026-10-29** | **35-Tool MCP Server (`43/43` checks)**, zero-stub CLI (`garuda mcp install`, `audit`, `rulebook`), Finance & Audit Guide, CLI Decision Guide, Persona Onboarding Guide, and **Multi-OS Release & Package Matrix** (`Windows`, `macOS`, `Linux`, `.deb`, `Homebrew`, `Winget`, `Scoop`, `Chocolatey`, `AUR`, `RPM`, `FreeBSD`, `Android`). |
| **`v1.9.0`** | **2026-10-28** | **Human-in-the-Loop Formula Disambiguation (`0.0–1.0` weights)**, **3-Tier Surface Criticality (`HIGH_SEVERE` / `MEDIUM` / `LOW`)**, **Continuous Financial Discrepancy Audit (`48 Cr` vs `4.8 Cr`)**, **Auto-Synthesized Base Rulebook Studio**, 9-Surface `/dashboard`, and `pg_isready` WAL Postgres handshake healing. |
| **`v1.8.4`** | **2026-10-27** | 3-second pre-built Portable PostgreSQL `16.4` on macOS/Linux (`~18 MB`, eliminating Homebrew source compiles), explicit IPv4 `127.0.0.1:5433` normalization, single-download Portable Go `1.26.0`, and `057_schema_hardening.sql`. |
| **`v1.8.2`–`v1.8.3`** | **2026-10-27** | Pure 7-bit ASCII Windows PowerShell 5.1 bootstrapper (`install.ps1` / `setup.ps1`), `cmd.exe /c` stderr-safe native wrappers, `System32\curl.exe` + `tar.exe` fast path, and self-healing `garuda doctor --fix`. |
| **`v1.8.0`–`v1.8.1`** | **2026-10-26** | **Deterministic `BuildContext` Provenance Receipts**, `go/build.Context` conditional compilation & build-tag filtering (`TestTwoBuildsExposeDifferentEdges`), CI-vs-Editor drift detector, and 4-target cross-OS `--matrix` gate. |
| **`v1.7.0`** | **2026-10-26** | **4-Tier Docs, Specs & Guardrails Lineage Graph**, **Garuda Talon & Risk Hub** leaderboard (`0–100` risk score), and zero-CDN offline SVG force/DAG layout engine. |
| **`v1.6.0`** | **2026-10-25** | Universal zero-dependency installers across Windows, macOS, Linux, Chromebook, and Android; Pure-Go (`!cgo`) TypeScript AST parser; `garuda doctor`; and installable PWA (`/manifest.json`, `/sw.js`). |
| **`v1.5.0`** | **2026-10-25** | **Garuda Talon Active Agent Execution Harness** (`.garuda/TALON_CONTRACT.md`, 6 IDE/MCP/Git hooks, `talon brief/guard/run`), Git-diff AST hunk targeting, net-new vs. baseline violation partitioning, and Python/TS alias/instance call resolution (`0.95`). |
| **`v1.3.0`–`v1.4.0`** | **2026-10-24** | Cryptographic Ed25519 JWT enforcement in `AuthMiddleware`, edge-level `SUPPORTED` runtime promotion, fail-closed YAML policy parser, and 6-surface Software Intelligence Instrument. |
| **`v1.1.0`–`v1.2.0`** | **2026-09-20 – 10-24** | Initial public release of `garuda`, `garuda-api`, `garuda-mcp`, Dual-Root RFC 6962 Merkle Ledger, `garuda check`, and recursive SQL CTE policy predicates (`layer_violation`, `transitive_claim_exists`, `cycle_exists`, `blast_radius_exceeds`). |

---

## 5. Verification & Cryptographic Integrity Summary (`v1.9.1`)

```bash
# 1. Verify all 24 strict compiler & multi-module ground-truth benchmarks (100% precision/recall, 0.0% FP)
make benchmark-strict

# 2. Verify all 35 MCP tools, argument whitelists, and active Talon contract injection (43/43 checks)
python3 scripts/mcp_verify.py

# 3. Verify 10/10 controlled OpenTelemetry runtime contradiction quarantine
python3 scripts/runtime_contradiction_test.py

# 4. Run System Doctor & Cross-Platform Build Matrix Gate
./bin/garuda doctor --fix
./bin/garuda talon guard --matrix --fail-on-block
```
