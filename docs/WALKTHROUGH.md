# Garuda System Walkthrough (`v1.9.1` Universal)

**A complete visual and technical tour of Garuda's 9-surface Software Intelligence Instrument (`/dashboard` PWA), the `Garuda Talon` active AI execution harness, deterministic `BuildContext` provenance receipts, cross-platform `--matrix` gates, 35 active MCP tools, and IDE integrations.**

---

## 0. Watch the 20-Second Cinematic Product Introduction

Before diving into the technical surfaces, watch the 20-second introduction video exploring Garuda's live control plane (`../assets/garuda_dashboard_demo_video.mp4`):

&lt;div align=&quot;center&quot;&gt;
  &lt;video src=&quot;../assets/garuda_dashboard_demo_video.mp4&quot; poster=&quot;../assets/Screenshot%20From%202026-10-25%2002-30-49.png&quot; controls autoplay loop muted playsinline width=&quot;100%&quot;&gt;
    &lt;source src=&quot;../assets/garuda_dashboard_demo_video.mp4&quot; type=&quot;video/mp4&quot;&gt;
    &lt;a href=&quot;../assets/garuda_dashboard_demo_video.mp4&quot;&gt;
      &lt;img src=&quot;../assets/Screenshot%20From%202026-10-25%2002-30-49.png&quot; width=&quot;100%&quot; alt=&quot;Watch the 20-Second Garuda Introduction Video&quot;&gt;
    &lt;/a&gt;
  &lt;/video&gt;
  &lt;p&gt;&lt;em&gt;🎬 &lt;a href=&quot;../assets/garuda_dashboard_demo_video.mp4&quot;&gt;&lt;strong&gt;Click here to play &lt;code&gt;assets/garuda_dashboard_demo_video.mp4&lt;/code&gt;&lt;/strong&gt;&lt;/a&gt;&lt;/em&gt;&lt;/p&gt;
&lt;/div&gt;

---

## 1. One-Command Zero-Dependency Universal Setup (`install.sh` / `install.ps1`)

Garuda `v1.9.1` eliminates manual toolchain downloads, manual SQL migration loops, manual JWT key generation, and passive MCP configuration. Whether you are on Windows 10/11, macOS (Apple Silicon or Intel), Linux, Chromebook (Crostini), or Android (Termux), Garuda provisions its own portable runtime in user-space (`~/.garuda`) and boots in one command:

### macOS, Linux, Chromebook (Crostini) & Android (Termux)
```bash
# Remote 1-command install & launch
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash

# Or inside a local repository
./setup.sh
```

### Windows 10 / 11 (PowerShell)
```powershell
# Remote 1-command install & launch
Set-ExecutionPolicy Bypass -Scope Process -Force; irm https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.ps1 | iex

# Or inside a local repository
powershell -ExecutionPolicy Bypass -File .\setup.ps1
```

### What Happens Automatically Under the Hood (`v1.9.1`)
1. **Fast Shallow Clone & Zero-Admin Portable PostgreSQL 16 (`127.0.0.1:5433`)**: `install.sh` and `install.ps1` clone with `--depth 1` (or unpack release archives via native `curl`/`tar`). If PostgreSQL is not running, Garuda downloads pre-built Portable PostgreSQL `16.4` (`~18 MB` on macOS/Linux, `~55 MB` on Windows) into `~/.garuda/runtime/postgres`, initializes a locale-safe cluster (`LC_ALL=C --no-locale`) in `~/.garuda/data/pg` with isolated sockets in `~/.garuda/data/run`, and binds to explicit IPv4 `127.0.0.1:5433`.
2. **Single-Download Portable Go `1.26.0` & 1-Pass Pure-Go Build (`CGO_ENABLED=0`)**: Downloads Portable Go `1.26.0` (matching `go.mod` so Go never downloads a second toolchain) into `~/.garuda/toolchain/go` and compiles all four binaries (`garuda`, `garuda-api`, `garuda-mcp`, `garuda-talon`) in one pass with commit build-stamp caching (`bin/.build_stamp`).
3. **Dynamic Identity, Schema Hardening (`057_schema_hardening.sql`) & Persistent `.env`**: Auto-discovers your email from `git config user.email`, runs all idempotent SQL migrations, and persists `DATABASE_URL` (normalized to `127.0.0.1:5433`), `GARUDA_WORKSPACE`, `GARUDA_TALON_ENFORCE=true`, `PORT=8080`, and Ed25519 JWT keys into `.env`.
4. **Garuda Talon Active Agent Harness (`garuda talon init`)**: Compiles live workspace ADRs (`document_claims`), declarative YAML policies (`policies/*.yaml`), and top high-blast-radius hubs into `.garuda/TALON_CONTRACT.md`, `.cursor/rules/garuda-talon.mdc`, `.cursor/mcp.json`, `CLAUDE.md`, `.claude/settings.json` (`PostToolUse` hook), `AGENTS.md`, `.windsurfrules`, and `.git/hooks/pre-commit`, then verifies all 43/43 dynamic MCP checks (`scripts/mcp_verify.py`).

---

## 2. Visual Tour of the 9-Surface Software Intelligence Instrument (`/dashboard` PWA)

Open `http://localhost:8080/dashboard` (or install it as a standalone desktop/Chromebook/Android Progressive Web App via `/manifest.json` and `/sw.js`). Every chart and graph uses Garuda's built-in **Zero-CDN Offline SVG Layout Engine** so it works 100% air-gapped.

### Surface 01 — Executive Workspace Briefing (`/dashboard#briefing`)
Provides an immediate system health strip (Indexed Entities, Verified Edges, Active Policy Posture, Contradictions Count, and RFC 6962 Merkle Head), a prioritized **Needs Attention** queue, and recent governance verdicts.

&lt;p align=&quot;center&quot;&gt;
  &lt;img src=&quot;../assets/Screenshot%20From%202026-10-25%2002-30-49.png&quot; width=&quot;100%&quot; alt=&quot;Surface 01: Executive Workspace Briefing&quot;&gt;
&lt;/p&gt;

### Surface 02 — Explore Index & High-Impact Symbol Directory (`/dashboard#explore`)
Automatically loads the workspace's top exported symbols ranked by inbound caller fan-in and blast radius on initial boot. Engineers can filter by language (`Go`, `Python`, `TypeScript`), entity kind (`struct`, `interface`, `func`, `method`, `class`, `package`), or press `⌘K` / `Ctrl+K` from anywhere to jump straight to a symbol.

&lt;p align=&quot;center&quot;&gt;
  &lt;img src=&quot;../assets/Screenshot%20From%202026-10-25%2002-31-11.png&quot; width=&quot;100%&quot; alt=&quot;Surface 02: Explore Index&quot;&gt;
&lt;/p&gt;

### Surface 03 — Neighborhood-First Architecture Map (`/dashboard#map`)
Replaces unreadable 10,000-node hairballs with a focused **1-hop or 2-hop ego-network explorer** (`[Upstream Callers] ──> [Focal Symbol] ──> [Downstream Dependencies & Policy Badges]`). Clicking any node opens the synchronized **4-Tab Inspector Drawer** (`Source`, `Relationships`, `Policy Results`, `Evidence`), and toggling **`✦ Docs & Rules Graph`** overlays normative ADR/spec links directly onto the map.

&lt;p align=&quot;center&quot;&gt;
  &lt;img src=&quot;../assets/Screenshot%20From%202026-10-25%2002-31-24.png&quot; width=&quot;100%&quot; alt=&quot;Surface 03: Neighborhood Architecture Map&quot;&gt;
&lt;/p&gt;

### Surface 04 — Docs, Specs & Guardrails Lineage Graph (`/dashboard#docs-graph`)
Introduced in `v1.7` and refined in `v1.8.1`, this surface visualizes how human-written engineering standards govern live code across a **4-tier DAG**:
- **Tier 1 — Documents**: Markdown ADRs (`docs/adr/*.md`), Design Specs (`docs/specs/*.md`), System Invariants (`docs/invariants.md`), YAML Policy Bundles (`policies/*.yaml`), and `.garudarules`.
- **Tier 2 — Extracted Normative Rules**: Machine-extracted `MUST`, `MUST_NOT`, `BLOCK`, `REVIEW`, `DENY`, and `INVARIANT` claims.
- **Tier 3 — Active Policy Guardrails**: Executable SQL/CTE predicates enforcing those rules.
- **Tier 4 — Code Symbols & Packages**: The exact Go, Python, and TypeScript symbols bound by each rule, complete with a **Document-to-Code Component Matrix** and an in-browser **Ingest Document / ADR Modal** (`POST /api/v1/dashboard/docs-ingest`).

&lt;p align=&quot;center&quot;&gt;
  &lt;img src=&quot;../assets/Screenshot%20From%202026-10-25%2002-31-40.png&quot; width=&quot;100%&quot; alt=&quot;Surface 04: Docs, Specs &amp; Guardrails Lineage Graph&quot;&gt;
&lt;/p&gt;

### Surface 05 — Garuda Talon & Architectural Risk Hub (`/dashboard#talon-hub`)
The operational nerve center for AI agent governance and build-context determinism (`v1.9.1`):
- **Deterministic `BuildContext` Provenance Receipt & CI-vs-Editor Drift Banner**: Displays the active `receipt_hash`, `graph_fingerprint`, `GOOS/GOARCH`, `CGO_ENABLED`, `-tags`, `go.mod` hash, dirty worktree state, drift status (`CLEAN` vs `DRIFT DETECTED`), and a 1-click **Copy Reproduce Cmd** button.
- **Cross-Platform Build Matrix Gate (`--matrix`)**: Shows real-time verification across `linux/amd64 (cgo=0)`, `linux/amd64 (cgo=1, tags=cgo)`, `darwin/arm64 (cgo=1)`, and `windows/amd64 (cgo=0)`.
- **Live Talon Active Guard Terminal**: Run diff-scoped `talon guard` checks or simulate an architectural violation to inspect the exact **Self-Repair Prompt** fed back to Cursor/Claude Code.
- **Active IDE, MCP & Git Hook Status Panel**: Tracks all 6 enforcement hooks with 1-click installation/sync.
- **Architectural Risk Hub Leaderboard**: Scores symbols from `0–100` (`CRITICAL`, `HIGH`, `MEDIUM`, `LOW`) by caller fan-in, transitive blast radius, and ADR sensitivity, with 1-click **Pre-Edit Briefing** drawers.

&lt;p align=&quot;center&quot;&gt;
  &lt;img src=&quot;../assets/Screenshot%20From%202026-10-25%2002-31-56.png&quot; width=&quot;100%&quot; alt=&quot;Surface 05: Garuda Talon &amp; Architectural Risk Hub&quot;&gt;
&lt;/p&gt;

### Surface 06 — Change Review Workbench & Interactive Policy Sandbox (`/dashboard#workbench` & `/dashboard#policies`)
A 3-pane pre-merge simulator where engineers select a target symbol and policy bundle on the left, inspect impacted upstream callers and applicable predicates in the center, and generate a live `ALLOW` / `WARN` / `REVIEW` / `BLOCK` receipt on the right—paired with a live YAML Policy Sandbox for testing new rules across all 9 graph predicates.

&lt;p align=&quot;center&quot;&gt;
  &lt;img src=&quot;../assets/Screenshot%20From%202026-10-25%2002-32-11.png&quot; width=&quot;100%&quot; alt=&quot;Surface 06: Change Review Workbench &amp; Policies&quot;&gt;
&lt;/p&gt;

### Surface 07 & 08 — Forensic Contradictions Queue & RFC 6962 Cryptographic Audit Ledger (`/dashboard#contradictions` & `/dashboard#ledger`)
- **Contradictions & Evidence (`#contradictions`)**: Compares static AST source locations against live OpenTelemetry runtime trace spans (`trace_id`, `span_id`, caller/callee service pair) and includes an interactive **OTel Runtime Span Simulator** to test shadow-call quarantine.
- **Governance & Audit Ledger (`#ledger`)**: Displays the append-only RFC 6962 SHA-256 Merkle ledger table with a bottom **Decision Record Inspector** (`What it did`, `What it looked at`, `BuildContext Receipt`, `Epistemic state`, `Merkle Inclusion Proof`) and 1-click cryptographic verification.

&lt;p align=&quot;center&quot;&gt;
  &lt;img src=&quot;../assets/Screenshot%20From%202026-10-25%2002-32-25.png&quot; width=&quot;100%&quot; alt=&quot;Surface 07 &amp; 08: Contradictions Queue &amp; Audit Ledger&quot;&gt;
&lt;/p&gt;

---

## 3. Deterministic `BuildContext` Receipts & Cross-Platform Matrix CLI (`v1.9.1`)

One of the hardest problems in static analysis and architectural governance is **environment drift**: a developer runs a check on `darwin/arm64` with `CGO_ENABLED=1`, while CI runs on `linux/amd64` with `CGO_ENABLED=0` or custom `-tags`.

In `v1.9.1`, every `garuda analyze`, `garuda check`, and `garuda talon guard` execution computes and seals a deterministic **`BuildContext` receipt**:

```bash
# Re-index workspace with a sealed v1.9.1 BuildContext receipt and Merkle block
./bin/garuda analyze . --workspace default --save

# Run the active Talon guard across your current git diff AND verify all 4 OS/CGO target profiles
./bin/garuda talon guard --workspace default --matrix --fail-on-block
```

Example terminal output:
```text
🔒 Garuda Talon Active Guard (v1.9.1)
   Workspace:         default
   Build Receipt:     bc_9f82d4a1e7c03b12 (graph: gf_4e19a82c01b7d3ef)
   Target Env:        linux/amd64 | CGO_ENABLED=0 | tags=[] | go1.24
   Environment Drift: CLEAN (matches indexed workspace graph receipt)
   Cross-OS Matrix:   [PASS] linux/amd64 (cgo=0) | [PASS] linux/amd64 (cgo=1) | [PASS] darwin/arm64 (cgo=1) | [PASS] windows/amd64 (cgo=0)
   Reproduce Command: CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ./bin/garuda talon guard --workspace default --diff HEAD --fail-on-block
   Governance Verdict: ALLOW ✓ (0 net-new violations; baseline isolated)
```

---

## 4. VS Code & Cursor IDE Extension (`vscode-extension/`)

Alongside the 23-tool MCP server and `.cursor/rules/garuda-talon.mdc` harness, Garuda ships a native VS Code / Cursor extension providing inline neighborhood graph inspection and live contradiction diagnostics right next to your code editor:

| IDE Graph View (`../assets/screenshots/ide-graph-view.png`) | IDE Contradictions Panel (`../assets/screenshots/ide-contradictions.png`) |
| :---: | :---: |
| &lt;img src=&quot;../assets/screenshots/ide-graph-view.png&quot; width=&quot;100%&quot; alt=&quot;IDE Graph View&quot;&gt; | &lt;img src=&quot;../assets/screenshots/ide-contradictions.png&quot; width=&quot;100%&quot; alt=&quot;IDE Contradictions Panel&quot;&gt; |


### Surface 09 — Financial & Logic Audit Workbench (`/dashboard?tab=audit`)
Added in `v1.9.0`–`v1.9.1`, Surface `09 — Financial & Logic Audit` brings together:
1. **Multi-Team Formula Disambiguation & 0.0–1.0 Weight Allocation**: Switch between **Ops Team (0.70 Weight)**, **Marketing Team (0.70 Weight)**, and **Management Team (0.70 Weight)** to compare competing `Loss Ratio` formulas, inspect caller counts and surface criticality tiers (`LOW`, `MEDIUM`, `HIGH_SEVERE`), and record human decisions (`✓ Take Human Call: Approve` or `🔒 Keep Frozen`).
2. **Continuous Financial & Business Data Discrepancy Detector (48 Crore vs. 4.8 Crore)**: Displays side-by-side team figures (`48.00 Crore INR` vs `4.80 Crore INR` — `10.00x Mismatch`), the AST & SQL data-lineage root cause trace, and 1-click **`✓ Approve Canonical Crore Fix`**.
