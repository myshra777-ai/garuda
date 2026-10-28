# Garuda Role-by-Role Persona Onboarding & Readability Guide (`v1.9.1`)

**Start Here Based on Your Role: Software Engineer · Staff Architect · Finance/Audit Controller · Founder/PM · DevOps/SRE · AI Coding Agent · Investor/Technical Evaluator**

Garuda `v1.9.1` serves both technical and non-technical stakeholders across the entire software and business lifecycle. Use this guide to jump straight to the workflows, dashboard surfaces, and commands designed for your exact role.

---

## Quick Role Selector

| Your Role / Persona | Your Primary Goal with Garuda | Best Starting Surface | 5-Minute Quickstart Section |
| :--- | :--- | :--- | :--- |
| **1. Software Engineer / Full-Stack Developer** | Know what breaks before editing code; verify only your own `git diff` without getting blocked by old legacy debt. | `06 — Change Review` + `garuda talon brief/guard` | [Section 1](#1-software-engineer--full-stack-developer) |
| **2. Staff / Principal Architect** | Turn undocumented repo architecture into an approved Base Rulebook; enforce layer boundaries and ADRs across all IDEs & CI. | `04 — Docs & Base Rulebook` + `03 — Architecture Map` | [Section 2](#2-staff--principal-architect) |
| **3. Finance Controller, CFO, Actuary & Statutory Auditor** | Catch cross-team financial number drift (e.g., 48 Cr vs 4.8 Cr Claims Paid) pre-audit; govern competing business formulas (`0.0–1.0` weights). | `09 — Financial & Logic Audit` + `07 — Evidence & Ledger` | [Section 3](#3-finance-controller-cfo-actuary--statutory-auditor) |
| **4. Non-Technical Founder, Product Manager & Engineering Lead** | Install in 1 command with zero setup friction; view a visual health scorecard of architecture, blast-radius tiers, and AI agent activity. | `01 — Overview` + `05 — Talon, Criticality & Risk Hub` | [Section 4](#4-non-technical-founder-product-manager--engineering-lead) |
| **5. DevOps, Platform Engineer & SRE** | Prevent "works on Mac, fails in Linux CI" build-tag drift; correlate live OpenTelemetry runtime spans against static AST claims. | `05 — Talon & Risk Hub` (`--matrix`) + `POST /api/v1/telemetry/spans` | [Section 5](#5-devops-platform-engineer--sre) |
| **6. AI Coding Agent (`Cursor`, `Claude Code`, `Codex`, `Windsurf`)** | Receive active architectural contracts automatically; query 35 MCP tools; self-repair violations in <20 ms before committing. | `bin/garuda-mcp` (35 Tools) + `.garuda/TALON_CONTRACT.md` | [Section 6](#6-ai-coding-agents-cursor-claude-code-codex-windsurf) |
| **7. Investor & Technical Due-Diligence Evaluator** | Verify reproducible benchmarks (`24/24` compiler suite, `43/43` MCP checks, `10/10` runtime quarantine, `18 ms` gate latency) in minutes. | `EVIDENCE.md` + `make benchmark-strict` + `scripts/mcp_verify.py` | [Section 7](#7-investor--technical-due-diligence-evaluator) |

---

## 1. Software Engineer / Full-Stack Developer

### Why You'll Love Garuda
- **Never get blamed for legacy technical debt**: `garuda talon guard` and `garuda check --diff HEAD` filter findings to **only the symbols touched in your diff** (`net_new_violations`).
- **No more blind refactors**: See every upstream caller and whether a symbol belongs to a `HIGH_SEVERE` (frozen), `MEDIUM` (trade-off), or `LOW` (UI box) surface before you type a single line.

### Your 3 Daily Commands
```bash
# 1. Before editing a struct, function, or file: see callers, risk score & ADR rules
./bin/garuda talon brief HandleRefund

# 2. After editing code: verify 0 net-new policy or ADR violations in your working tree
./bin/garuda talon guard --diff HEAD --fail-on-block

# 3. If working on a multi-team business formula (e.g., Loss Ratio): check team weights (0.0-1.0)
./bin/garuda audit logic "Loss Ratio" --team ops
```
- **Next Read**: [`docs/CLI_DECISION_GUIDE.md`](CLI_DECISION_GUIDE.md)

---

## 2. Staff / Principal Architect

### Why You'll Love Garuda
- **Eliminates documentation grunt work**: If your repository has no formal ADRs yet, Garuda's **Auto-Synthesized Base Rulebook Studio** extracts implicit layer boundaries, fan-in criticality tiers, and multi-team formula weights into `.garuda/DRAFT_BASE_RULEBOOK.md` so you can review, edit, and approve it in one click.
- **4-Tier Docs-to-Code Lineage**: Every Markdown ADR (`docs/adr/*.md`) and YAML policy (`policies/*.yaml`) is visually connected to the exact AST symbols it governs (`Document -> Extracted Rule -> Policy Guardrail -> Code Symbol`).

### Your 4 Core Workflows
```bash
# 1. Auto-synthesize a structured Base Rulebook from an unorganized repository
./bin/garuda rulebook generate
./bin/garuda rulebook approve --by "Principal Architect"

# 2. Ingest existing Markdown ADRs and view the 4-tier Docs-to-Code Lineage Matrix
./bin/garuda docs ingest ./docs/adr
./bin/garuda docs graph

# 3. Validate and evaluate declarative YAML policies (layer_violation, cycle_exists, etc.)
./bin/garuda policy validate ./policies
./bin/garuda policy evaluate ./policies --save --fail-on-block

# 4. Inspect 3-Tier Surface Criticality classifications
./bin/garuda audit criticality
```
- **Next Read**: [`PLAYBOOK.md`](../PLAYBOOK.md) and [`docs/SPECS.md`](SPECS.md)

---

## 3. Finance Controller, CFO, Actuary & Statutory Auditor

### Why You'll Love Garuda
- **Catches 10x number discrepancies before external audits**: Traces AST and SQL aggregation formulas across teams to flag unit-scaling mismatches—such as Claims Ops reporting **48.00 Crore INR** (`SUM(settled_claims_inr) / 1e7`) while Finance's executive export reports **4.80 Crore INR** (`SUM(settled_claims_lakhs) / 1000.0` instead of `/ 100.0`).
- **Prevents unauthorized AI formula changes**: Assigns `0.0–1.0` team weights to competing formulas and locks statutory/core surfaces as `HIGH_SEVERE`.
- **Tamper-evident RFC 6962 Merkle proofs**: Every approved formula, rulebook, and reconciliation is sealed with a SHA-256 Merkle inclusion proof.

### Your Dashboard & CLI Workflow
1. Open **`http://localhost:8080/dashboard`** and click **`09 — Financial & Logic Audit`**.
2. Inspect the **48.00 Crore vs. 4.80 Crore** side-by-side lineage card and click **`✓ Approve Canonical Crore Fix`**.
3. Or run from the terminal:
   ```bash
   ./bin/garuda audit financial --metric "Total Claims Paid" --team-a-cr 48.0 --team-b-cr 4.8
   ./bin/garuda verify
   ```
- **Next Read**: **[`docs/FINANCE_AND_AUDIT_GUIDE.md`](FINANCE_AND_AUDIT_GUIDE.md)**

---

## 4. Non-Technical Founder, Product Manager & Engineering Lead

### Why You'll Love Garuda
- **True 1-Command Zero-Dependency Install**: Works out of the box on macOS, Windows PowerShell, Linux, Chromebook, and Android without needing Docker, Go, Postgres, or Xcode pre-installed.
- **Plain-English Business Impact (Amazon Search vs. Wishlist Model)**: Translates technical call graphs into clear business criticality tiers:
  - **`HIGH_SEVERE` (Core Homepage, Search Redirect, Settlement Ledger — 142 callers)**: Frozen against automated AI breakage so customers are never blocked from searching or paying.
  - **`MEDIUM` (Wishlist Loader, Promo Badges — 19 callers)**: Non-blocking features where product/engineering trade-offs can be weighed.
  - **`LOW` (UI Box Alignment — 4 callers)**: Safe for fast visual iteration.

### Your 1-Command Launch
```bash
# macOS / Linux / Chromebook
./setup.sh

# Windows PowerShell
powershell -NoProfile -ExecutionPolicy Bypass -File .\setup.ps1
```
Then open **`http://localhost:8080/login`** (or install as a desktop/mobile Chrome PWA).
- **Next Read**: [`INSTALL.md`](../INSTALL.md) and [`docs/WALKTHROUGH.md`](WALKTHROUGH.md)

---

## 5. DevOps, Platform Engineer & SRE

### Why You'll Love Garuda
- **Deterministic `BuildContext` Receipts & Cross-OS Matrix (`--matrix`)**: Binds `GOOS`, `GOARCH`, `CGO_ENABLED`, `-tags`, and `go.mod` / `tsconfig.json` hashes into every check, and tests `linux/amd64 (cgo=0)`, `linux/amd64 (cgo=1)`, `darwin/arm64 (cgo=1)`, and `windows/amd64 (cgo=0)` in one pass.
- **Live OpenTelemetry Runtime Contradiction Quarantine**: Correlates production OTel spans (`POST /api/v1/telemetry/spans`) against static AST edges to quarantine shadow runtime calls (`CONTRADICTED`).

### Your CI/CD & SRE Commands
```bash
# 1. Run cross-platform build matrix gate in CI/CD
./bin/garuda talon guard --diff origin/main...HEAD --matrix --fail-on-block

# 2. Self-heal local or CI runtime environment
./bin/garuda doctor --fix
```
- **Next Read**: [`PLAYBOOK.md`](../PLAYBOOK.md) §10 & §14

---

## 6. AI Coding Agents (`Cursor`, `Claude Code`, `Codex`, `Windsurf`)

### How Agents Interact with Garuda (`bin/garuda-mcp`)
1. **First-Call Auto-Orientation**: On the very first MCP tool call in any session (`GARUDA_TALON_ENFORCE=true`), `garuda-mcp` automatically injects the workspace's active ADR `MUST`/`MUST_NOT` rules and YAML policy guardrails into the JSON response.
2. **Pre-Edit Briefing**: Call `garuda.talon_pre_edit` before modifying any file or symbol.
3. **Human-in-the-Loop Formula Disambiguation**: Call `garuda.disambiguate_logic` when touching business metrics; present the `0.0–1.0` weights and severity tiers to the human user and record their choice via `garuda.decide_logic_variant`.
4. **Post-Edit Gate & Self-Repair**: Call `garuda.talon_verify_edit` immediately after editing. If `outcome == "BLOCK"`, apply the `self_repair_instructions` and re-verify until `ALLOW`.
- **Next Read**: [`docs/API.md`](API.md) §4

---

## 7. Investor & Technical Due-Diligence Evaluator

### Reproducible Proof Points in 3 Commands
```bash
# 1. Zero-dependency 1-command provisioning, migration, AST scan, and 43/43 MCP verification
./setup.sh

# 2. Run the 24/24 ground-truth compiler & determinism benchmark suite (100% precision/recall, 0.0% FP)
make benchmark-strict

# 3. Verify all 35 MCP tools, Talon active contract injection, and v1.9.1 financial/logic audit tools
python3 scripts/mcp_verify.py
```
- **Next Read**: [`EVIDENCE.md`](../EVIDENCE.md) and [`docs/CAPABILITIES.md`](CAPABILITIES.md)
