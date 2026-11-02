# Garuda System Walkthrough

**A complete visual and technical tour of Garuda's 10-surface Software Intelligence Instrument (`/dashboard` PWA), the `Garuda Talon` active AI execution harness, deterministic `BuildContext` provenance receipts, cross-platform `--matrix` gates, 54 active MCP tools, and IDE integrations.**

---

## 1. Setup

Install and first-run instructions for Windows, macOS, Linux, ChromeOS and Android (Termux) are in [INSTALL.md](../INSTALL.md). In short:

```bash
# macOS, Linux, ChromeOS (Linux), Termux
curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash
```

```powershell
# Windows 10 / 11
irm https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.ps1 | iex
```

The installer provisions PostgreSQL 16 and Go in user space under `~/.garuda` when they are missing, builds the four binaries, applies the migrations, writes `.env`, indexes the repository, wires the Talon hooks and starts the API on `127.0.0.1:8080`.

---

## 2. Visual Tour of the 10-Surface Software Intelligence Instrument (`/dashboard` PWA)

Open `http://localhost:8080/dashboard` (or install it as a standalone desktop/Chromebook/Android Progressive Web App via `/manifest.json` and `/sw.js`). Every chart and graph uses Garuda's built-in **Zero-CDN Offline SVG Layout Engine** so it works 100% air-gapped.

The **00 — About Garuda** page inside the dashboard explains every surface with expandable cards:

<p align="center">
  <img src="../assets/screenshots/topology-overview.png" width="100%" alt="About Garuda page: orient, search and map topology (surfaces 01-03)">
</p>

### Surface 01 — Executive Workspace Briefing (`/dashboard#briefing`)
Provides an immediate system health strip (Indexed Entities, Verified Edges, Active Policy Posture, Contradictions Count, and RFC 6962 Merkle Head), a prioritized **Needs Attention** queue, and recent governance verdicts.


### Surface 02 — Explore Index & High-Impact Symbol Directory (`/dashboard#explore`)
Automatically loads the workspace's top exported symbols ranked by inbound caller fan-in and blast radius on initial boot. Engineers can filter by language (`Go`, `Python`, `TypeScript`), entity kind (`struct`, `interface`, `func`, `method`, `class`, `package`), or press `⌘K` / `Ctrl+K` from anywhere to jump straight to a symbol.


### Surface 03 — Neighborhood-First Architecture Map (`/dashboard#map`)
Replaces unreadable 10,000-node hairballs with a focused **1-hop or 2-hop ego-network explorer** (`[Upstream Callers] ──> [Focal Symbol] ──> [Downstream Dependencies & Policy Badges]`). Clicking any node opens the synchronized **4-Tab Inspector Drawer** (`Source`, `Relationships`, `Policy Results`, `Evidence`), and toggling **`✦ Docs & Rules Graph`** overlays normative ADR/spec links directly onto the map.


### Surface 04 — Docs, Specs & Guardrails Lineage Graph (`/dashboard#docs-graph`)
This surface visualizes how human-written engineering standards govern live code across a **4-tier DAG**:
- **Tier 1 — Documents**: Markdown ADRs (`docs/adr/*.md`), Design Specs (`docs/specs/*.md`), System Invariants (`docs/invariants.md`), YAML Policy Bundles (`policies/*.yaml`), and `.garudarules`.
- **Tier 2 — Extracted Normative Rules**: Machine-extracted `MUST`, `MUST_NOT`, `BLOCK`, `REVIEW`, `DENY`, and `INVARIANT` claims.
- **Tier 3 — Active Policy Guardrails**: Executable SQL/CTE predicates enforcing those rules.
- **Tier 4 — Code Symbols & Packages**: The exact Go, Python, and TypeScript symbols bound by each rule, complete with a **Document-to-Code Component Matrix** and an in-browser **Ingest Document / ADR Modal** (`POST /api/v1/dashboard/docs-ingest`).


<p align="center">
  <img src="../assets/screenshots/docs-to-guardrails.png" width="100%" alt="About Garuda page: 4-tier docs-to-code lineage and the Talon pre-edit briefing (surfaces 04-06)">
</p>

### Surface 05 — Garuda Talon & Architectural Risk Hub (`/dashboard#talon-hub`)
The operational nerve center for AI agent governance and build-context determinism:
- **Deterministic `BuildContext` Provenance Receipt & CI-vs-Editor Drift Banner**: Displays the active `receipt_hash`, `graph_fingerprint`, `GOOS/GOARCH`, `CGO_ENABLED`, `-tags`, `go.mod` hash, dirty worktree state, drift status (`CLEAN` vs `DRIFT DETECTED`), and a 1-click **Copy Reproduce Cmd** button.
- **Cross-Platform Build Matrix Gate (`--matrix`)**: Shows real-time verification across `linux/amd64 (cgo=0)`, `linux/amd64 (cgo=1, tags=cgo)`, `darwin/arm64 (cgo=1)`, and `windows/amd64 (cgo=0)`.
- **Live Talon Active Guard Terminal**: Run diff-scoped `talon guard` checks or simulate an architectural violation to inspect the exact **Self-Repair Prompt** fed back to Cursor/Claude Code.
- **Active IDE, MCP & Git Hook Status Panel**: Tracks all 6 enforcement hooks with 1-click installation/sync.
- **Architectural Risk Hub Leaderboard**: Scores symbols from `0–100` (`CRITICAL`, `HIGH`, `MEDIUM`, `LOW`) by caller fan-in, transitive blast radius, and ADR sensitivity, with 1-click **Pre-Edit Briefing** drawers.


### Surface 06 — Change Review Workbench & Interactive Policy Sandbox (`/dashboard#workbench` & `/dashboard#policies`)
A 3-pane pre-merge simulator where engineers select a target symbol and policy bundle on the left, inspect impacted upstream callers and applicable predicates in the center, and generate a live `ALLOW` / `WARN` / `REVIEW` / `BLOCK` receipt on the right—paired with a live YAML Policy Sandbox for testing new rules across all 9 graph predicates.


### Surface 07 & 08 — Forensic Contradictions Queue & RFC 6962 Cryptographic Audit Ledger (`/dashboard#contradictions` & `/dashboard#ledger`)
- **Contradictions & Evidence (`#contradictions`)**: Compares static AST source locations against live OpenTelemetry runtime trace spans (`trace_id`, `span_id`, caller/callee service pair) and includes an interactive **OTel Runtime Span Simulator** to test shadow-call quarantine.
- **Governance & Audit Ledger (`#ledger`)**: Displays the append-only RFC 6962 SHA-256 Merkle ledger table with a bottom **Decision Record Inspector** (`What it did`, `What it looked at`, `BuildContext Receipt`, `Epistemic state`, `Merkle Inclusion Proof`) and 1-click cryptographic verification.


---

## 3. Deterministic `BuildContext` Receipts & Cross-Platform Matrix CLI

One of the hardest problems in static analysis and architectural governance is **environment drift**: a developer runs a check on `darwin/arm64` with `CGO_ENABLED=1`, while CI runs on `linux/amd64` with `CGO_ENABLED=0` or custom `-tags`.

Every `garuda analyze`, `garuda check`, and `garuda talon guard` execution computes and seals a deterministic **`BuildContext` receipt**:

```bash
# Re-index workspace with a sealed BuildContext receipt and Merkle block
./bin/garuda analyze . --workspace default --save

# Run the active Talon guard across your current git diff AND verify all 4 OS/CGO target profiles
./bin/garuda talon guard --workspace default --matrix --fail-on-block
```

Example terminal output:
```text
🔒 Garuda Talon Active Guard
   Workspace:         default
   Build Receipt:     bc_9f82d4a1e7c03b12 (graph: gf_4e19a82c01b7d3ef)
   Target Env:        linux/amd64 | CGO_ENABLED=0 | tags=[] | go1.26
   Environment Drift: CLEAN (matches indexed workspace graph receipt)
   Cross-OS Matrix:   [PASS] linux/amd64 (cgo=0) | [PASS] linux/amd64 (cgo=1) | [PASS] darwin/arm64 (cgo=1) | [PASS] windows/amd64 (cgo=0)
   Reproduce Command: CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ./bin/garuda talon guard --workspace default --diff HEAD --fail-on-block
   Governance Verdict: ALLOW ✓ (0 net-new violations; baseline isolated)
```

---

## 4. VS Code & Cursor IDE Extension (`vscode-extension/`)

Alongside the 54-tool MCP server and `.cursor/rules/garuda-talon.mdc` harness, Garuda ships a VS Code / Cursor extension (`vscode-extension/`) that shows inline neighborhood graphs and live contradiction diagnostics next to your code. Its TypeScript source is in `vscode-extension/src/`.


### Surface 09 — Data & Logic Audit Workbench (`/dashboard?tab=audit`)

<p align="center">
  <img src="../assets/screenshots/data-logic-audit.png" width="100%" alt="About Garuda page: data and logic audit, 48 Crore vs 4.8 Crore discrepancy example (surface 09)">
</p>

Surface `09 — Data & Logic Audit` brings together:
1. **Multi-Team Formula Disambiguation & 0.0–1.0 Weight Allocation**: Switch between **Ops Team (0.70 Weight)**, **Marketing Team (0.70 Weight)**, and **Management Team (0.70 Weight)** to compare competing `Loss Ratio` formulas, inspect caller counts and surface criticality tiers (`LOW`, `MEDIUM`, `HIGH_SEVERE`), and record human decisions (`✓ Take Human Call: Approve` or `🔒 Keep Frozen`).
2. **Continuous Financial & Business Data Discrepancy Detector (48 Crore vs. 4.8 Crore)**: Displays side-by-side team figures (`48.00 Crore INR` vs `4.80 Crore INR` — `10.00x Mismatch`), the AST & SQL data-lineage root cause trace, and 1-click **`✓ Approve Canonical Crore Fix`**.
