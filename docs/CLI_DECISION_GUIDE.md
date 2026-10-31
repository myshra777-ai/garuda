# Garuda CLI & MCP Decision Guide: How & When to Use What Command (`v1.9.9`)

**A Complete Decision Matrix, Workflow Playbook, and CLI-to-MCP Mapping for Developers, Architects, Finance Auditors, CI/CD Pipelines, and AI Coding Agents**

---

## 1. Quick-Decision Matrix: "I Want To... Which Command & MCP Tool Do I Use?"

Use this table whenever you need to pick the exact right Garuda command or MCP tool for your task:

| Your Goal / Scenario | When to Use It | Exact CLI Command (`./bin/garuda ...`) | Equivalent MCP Tool for LLMs (`bin/garuda-mcp`) | Mutates State? |
| :--- | :--- | :--- | :--- | :--- |
| **1. First-Time Setup or Onboarding a New Repo** | Once per machine or when initializing a new repository. Starts Postgres, runs 93 migrations, seeds user, scans AST, and wires all 6 IDE/Git hooks. | `./setup.sh` *(macOS/Linux)*<br>`.\setup.ps1` *(Windows)*<br>or `./bin/garuda init .` | *(Automated by `init` / `talon init`)* | Yes (DB + `.env` + Hooks) |
| **2. Diagnose or Fix Any Setup / DB / Hook Issue** | Anytime Postgres stops after a reboot, `.env` is missing, or MCP hooks need re-syncing. | `./bin/garuda doctor --fix` | `garuda.doctor_diagnose` | Read-only (or repairs with `--fix`) |
| **3. Check Blast Radius & Rules BEFORE Editing Code** | **Right before** you or an AI agent touch a struct, function, or file. Shows callers, criticality tier, and ADR/policy rules. | `./bin/garuda talon brief <SymbolOrFile>` | `garuda.talon_pre_edit` | No (Read-only) |
| **4. Verify Your Uncommitted Changes AFTER Editing** | **Immediately after** editing code (or in pre-commit). Runs sub-second incremental AST sync, checks only your diff (`net_new_violations`), and emits a Self-Repair Prompt if `BLOCK`. | `./bin/garuda talon guard --diff HEAD --fail-on-block` | `garuda.talon_verify_edit` | Yes (Incremental AST + Merkle Seal) |
| **5. Run a Full 5-Plane Change-Control Gate (PR / CI)** | In pull requests, pre-merge reviews, or CI/CD pipelines. Evaluates AST blast radius, 3-state epistemics, YAML policies, and seals an RFC 6962 Merkle receipt. | `./bin/garuda check --diff origin/main...HEAD --save --fail-on-block` | `garuda.talon_verify_edit` / `garuda.policy.evaluate` | Yes (with `--save`) |
| **6. Test Code Across Linux, macOS, Windows & CGO** | When your code uses `//go:build` tags, OS suffixes (`_linux.go`, `_windows.go`), or `cgo` and you want to prevent "works on Mac, fails in Linux CI" drift. | `./bin/garuda talon guard --matrix`<br>or `./bin/garuda check --matrix` | `garuda.build_matrix_check` | No (or seals with `check --save`) |
| **7. Full or Incremental Repository AST Scan** | After pulling major upstream changes, switching branches, or onboarding a Go, Python, or TypeScript repo. | `./bin/garuda analyze . --workspace default --save` | *(Triggered automatically by `garuda.talon_verify_edit`)* | Yes (Updates AST Graph + Merkle Revision) |
| **8. Compare Competing Team Formulas (0.0–1.0 Weights)** | When multiple teams (Ops, Marketing, Management) have different undocumented formulas for the same metric (e.g., `Loss Ratio`) and a human must choose safely. | `./bin/garuda audit logic "Loss Ratio" --team ops` | `garuda.disambiguate_logic`<br>`garuda.decide_logic_variant` | Read-only (or saves with `--decide`) |
| **9. Audit Cross-Team Financial Numbers (48 Cr vs 4.8 Cr)** | Before board meetings, quarterly closes, or statutory audits to catch 10x unit-scaling bugs (Lakhs vs Crore) across team pipelines. | `./bin/garuda audit data-logic --metric "Claims Paid" --team-a-cr 48.0 --team-b-cr 4.8` | `garuda.audit_financial_discrepancy`<br>`garuda.reconcile_financial_audit` | Yes (Logs & reconciles audit record) |
| **10. Inspect 3-Tier Surface Criticality (`HIGH_SEVERE` / `MEDIUM` / `LOW`)** | To see which surfaces are frozen (`Homepage`, `Search Redirect`, `Claims Ledger`) vs open to trade-offs (`Wishlist`) vs low-impact (`UI Box`). | `./bin/garuda audit criticality` | `garuda.criticality_tiers` | Read-only (or registers custom tier) |
| **11. Auto-Generate & Approve a Base Rulebook** | When a repository has **no formal documentation** and you want Garuda to synthesize a structured rulebook for 1-click human approval. | `./bin/garuda rulebook generate`<br>`./bin/garuda rulebook approve --by "Lead Architect"` | `garuda.synthesize_rulebook`<br>`garuda.approve_base_rulebook` | Yes (Writes `.garuda/*_RULEBOOK.md` + DB) |
| **12. Ingest Markdown ADRs & View 4-Tier Docs-to-Code Graph** | When adding or updating `docs/adr/*.md` or specs and verifying which AST symbols implement or violate them. | `./bin/garuda docs ingest ./docs/adr`<br>`./bin/garuda docs verify`<br>`./bin/garuda docs graph` | `garuda.docs_ingest`<br>`garuda.docs_graph`<br>`garuda.check_drift` | `ingest`/`verify` mutate DB; `graph` is read-only |
| **13. Author, Validate & Evaluate YAML Policies** | When writing or testing declarative architectural rules (`layer_violation`, `cycle_exists`, `blast_radius_exceeds`) in `./policies/*.yaml`. | `./bin/garuda policy validate ./policies`<br>`./bin/garuda policy evaluate ./policies --save` | `garuda.policy.list`<br>`garuda.policy.evaluate`<br>`garuda.verify_policy_evaluation` | Preview by default; `--save` anchors to Merkle |
| **14. Check Static Code Hygiene (Unreferenced / Duplicates)** | During refactoring sprints to find advisory unreferenced symbols, duplicate symbol names across packages, and stdlib alternatives. | `./bin/garuda hygiene .` | `garuda.hygiene_report` | No (Pure in-memory advisory report) |
| **15. Verify Cryptographic RFC 6962 Merkle Ledger Integrity** | During security reviews or compliance audits to prove zero tampering across all historical decisions and policy evaluations. | `./bin/garuda verify`<br>`./bin/garuda policy verify <eval-id>` | `garuda.verify_policy_evaluation`<br>`garuda.governance.status` | No (Cryptographic verification) |

---

## 2. Deep Comparison: When to Use Similar Commands

### 2.1 `garuda talon guard` vs. `garuda check` vs. `garuda policy evaluate`

Many users ask how these three verification commands differ. Here is the exact distinction:

| Dimension | `garuda talon guard` | `garuda check` | `garuda policy evaluate` |
| :--- | :--- | :--- | :--- |
| **Primary Audience** | AI Coding Agents (`Cursor`, `Claude Code`), Git `pre-commit` hooks, and rapid developer inner loops. | Pull Request gates, Change Review Workbench, and symbol-targeted impact reviews. | Architects authoring and testing `./policies/*.yaml` bundles. |
| **Incremental Dirty-File AST Sync** | **Yes (Default `--auto-index=true`)**: Automatically re-indexes uncommitted working-tree files in milliseconds before checking. | **Yes (Default `--auto-index=true`)**: Syncs dirty files when `--diff` or `--staged` is used. | **No**: Evaluates against the already-indexed PostgreSQL workspace graph. |
| **Checks ADR `document_claims`?** | **Yes**: Evaluates both YAML policies AND ingested ADR `MUST`/`MUST_NOT` claims. | **Yes**: Combines blast radius, 3-state epistemics, YAML policies, and Merkle sealing. | **YAML Policies Only**: Evaluates the 9 declarative YAML predicates. |
| **Emits Copy-Ready Agent Self-Repair Prompt?** | **Yes**: Outputs a structured, file-and-line-specific repair prompt for LLMs on `BLOCK`. | Outputs a human & JSON 5-plane Change-Control Decision Record. | Outputs per-policy `ALLOW`/`WARN`/`REVIEW`/`BLOCK` table. |

### 2.2 `garuda docs ingest` vs. `garuda rulebook generate`

- **Use `garuda docs ingest ./docs/adr` when you ALREADY have written documents** (Markdown ADRs, design specs, `.txt`, `.docx`, `.pdf`) and want Garuda to extract `MUST` / `MUST_NOT` rules and link them to code symbols.
- **Use `garuda rulebook generate` when your repo is UNORGANIZED or undocumented** and you want Garuda to automatically reverse-engineer your codebase's layer boundaries, surface criticality tiers (`HIGH_SEVERE` / `MEDIUM` / `LOW`), competing team formula weights (`0.0–1.0`), and financial invariants into `.garuda/DRAFT_BASE_RULEBOOK.md` so a human can review and approve it in one click (`garuda rulebook approve`).

---

## 3. Complete CLI Command & Flag Reference (`./bin/garuda`)

### 3.1 Setup, Diagnostics & Daemon Lifecycle
```bash
# 1. Initialize database, migrations, .env, owner account, AST index & Talon hooks
./bin/garuda init [repo_path] --workspace <name> --email <email> --password <pass> [--skip-scan]

# 2. Diagnose and auto-repair Postgres, .env, binaries, migrations & IDE hooks
./bin/garuda doctor [--fix] [--json]

# 3. Start the unified HTTP API Gateway + 10-Surface Web Dashboard + background verifier
./bin/garuda dev

# 4. Check live gateway health
./bin/garuda status
```

### 3.2 Codebase Analysis & Topology Inspection
```bash
# 1. Full multi-language AST analysis (Go 1.0, Python 0.95, TypeScript 0.95)
./bin/garuda analyze <path> --workspace <name> --save [--tags cgo,enterprise] [--target linux/amd64] [--cgo=false]

# 2. Sub-second incremental working-tree sync of modified files
./bin/garuda analyze . --workspace <name> --incremental --files internal/api/handler.go --save

# 3. Compare two JSON snapshots for breaking signature/contract changes
./bin/garuda analyze --diff before.json after.json [--fail-on-breaking]

# 4. Executive workspace summary, package topology, symbol inspection & blast radius
./bin/garuda summary [--workspace <name>]
./bin/garuda inspect <pkg.Symbol> [--workspace <name>]
./bin/garuda justify <entity-uuid-or-pkg.Symbol>
./bin/garuda impact --workspace <ws-uuid> --target <entity-uuid> --depth 3
./bin/garuda hygiene . [--json] [-o hygiene.json]
```

### 3.3 Garuda Talon Active Agent Harness (`garuda talon`)
```bash
# 1. Compile .garuda/TALON_CONTRACT.md and install all 6 IDE, MCP & Git hooks
./bin/garuda talon init . --workspace <name> --policies ./policies [--json]

# 2. Print the compiled workspace architectural contract
./bin/garuda talon contract . --workspace <name> [--json]

# 3. Pre-edit briefing: blast radius, callers, criticality tier & governing rules
./bin/garuda talon brief <SymbolOrFile> --workspace <name> [--json]

# 4. Post-edit active gate: diff-scoped check + BuildContext receipt + 4-OS matrix + Self-Repair Prompt
./bin/garuda talon guard . --workspace <name> [--staged | --diff HEAD | --files a.go,b.go] [--matrix] [--fail-on-block] [--hook] [--json]

# 5. Autonomous supervisor loop: run an AI coding agent and auto-loop self-repair until ALLOW
./bin/garuda talon run --max-retries 3 -- <agent-command> [args...]
```

### 3.4 Unified Change-Control & Policy Engine (`garuda check` & `garuda policy`)
```bash
# 1. Unified 5-plane Change-Control check (diff-scoped or symbol-targeted)
./bin/garuda check [dir] --project <ws> [--diff HEAD | --staged | --target <Symbol>] --policy ./policies [--matrix] --save --fail-on-block [--json]

# 2. Validate, evaluate, list, show, and cryptographically verify YAML policies
./bin/garuda policy validate ./policies
./bin/garuda policy evaluate ./policies --workspace <ws> --save [--reconcile] [--fail-on-block] [--matrix] [--json]
./bin/garuda policy list
./bin/garuda policy show <evaluation-uuid>
./bin/garuda policy verify <evaluation-uuid>
```

### 3.5 Human-in-the-Loop Logic Audit, Criticality Tiers, Financial Audit & Base Rulebook (`v1.9.9`)
```bash
# 1. Multi-team formula disambiguation (0.0-1.0 weights + Least Architectural Harm)
./bin/garuda audit logic [concept] --team <ops|marketing|management> [--decide <variant_id> --action APPROVED_BY_HUMAN] [--json]

# 2. Cross-team financial discrepancy audit (e.g., 48 Crore vs 4.8 Crore Claims Paid)
./bin/garuda audit data-logic --metric "Total Claims Paid" --team-a-cr 48.0 --team-b-cr 4.8 [--reconcile <audit_id>] [--json]

# 3. 3-Tier Surface Criticality classification report
./bin/garuda audit criticality [--tier HIGH_SEVERE|MEDIUM|LOW] [--json]

# 4. Auto-synthesize and approve the workspace Base Rulebook
./bin/garuda rulebook generate [--output .garuda/DRAFT_BASE_RULEBOOK.md]
./bin/garuda rulebook approve --by "Primary Stakeholder / Staff Architect"
```

### 3.6 Documentation Lineage & Cryptographic Ledger Verification
```bash
# 1. Ingest ADRs/specs, verify against AST, and print 4-tier Docs-to-Code Lineage Graph
./bin/garuda docs ingest ./docs/adr
./bin/garuda docs verify
./bin/garuda docs status
./bin/garuda docs graph

# 2. Verify RFC 6962 Merkle revision chains and explain historical decisions
./bin/garuda verify
./bin/garuda explain <decision-uuid>
```

---

## 4. Complete 54-Tool MCP Surface for LLMs (`bin/garuda-mcp`)

Every tool below is exposed over stdio JSON-RPC 2.0 (`protocolVersion: "2025-06-18"`), protected by `toolArgWhitelist`, and verified by `python3 scripts/mcp_verify.py` (`63/63` checks passing):

| Category | MCP Tool Name | When an LLM / Coding Agent Should Call It |
| :--- | :--- | :--- |
| **1. Active Talon Harness** | `garuda.talon_pre_edit` | **Mandatory before any edit**: Retrieves target blast radius, callers, ADR rules, and active policies. |
| **1. Active Talon Harness** | `garuda.talon_verify_edit` | **Mandatory after any edit**: Incrementally re-indexes dirty files, runs diff-scoped gate, seals Merkle block, and returns Self-Repair Prompt if `BLOCK`. |
| **2. Logic, Finance & Rulebook (`v1.9.9`)** | `garuda.disambiguate_logic` | When encountering multiple formulas for the same business concept; returns `0.0–1.0` team weights and blast-radius severity for human decision. |
| **2. Logic, Finance & Rulebook (`v1.9.9`)** | `garuda.decide_logic_variant` | After the human user chooses a formula variant (`APPROVE_VARIANT` or `FREEZE_SEVERE`), persists the decision in PostgreSQL. |
| **2. Logic, Finance & Rulebook (`v1.9.9`)** | `garuda.audit_financial_discrepancy` | When auditing financial/business metrics across teams (e.g., 48 Cr vs 4.8 Cr Claims Paid) to trace unit-scaling bugs in AST/SQL lineage. |
| **2. Logic, Finance & Rulebook (`v1.9.9`)** | `garuda.reconcile_financial_audit` | Records human auditor sign-off on a flagged financial discrepancy. |
| **2. Logic, Finance & Rulebook (`v1.9.9`)** | `garuda.synthesize_rulebook` | Synthesizes a structured draft Base Rulebook (`.garuda/DRAFT_BASE_RULEBOOK.md`) from an unorganized repository. |
| **2. Logic, Finance & Rulebook (`v1.9.9`)** | `garuda.approve_base_rulebook` | Locks and seals the human-reviewed Base Rulebook into `.garuda/APPROVED_BASE_RULEBOOK.md` with an RFC 6962 SHA-256 hash. |
| **2. Logic, Finance & Rulebook (`v1.9.9`)** | `garuda.criticality_tiers` | Queries or registers `HIGH_SEVERE` (frozen), `MEDIUM` (trade-off), and `LOW` (UI box) architectural surfaces. |
| **3. Docs Lineage & Build Matrix (`v1.9.9`)** | `garuda.docs_graph` | Retrieves the 4-tier Document → Normative Rule → Policy Guardrail → AST Symbol lineage matrix. |
| **3. Docs Lineage & Build Matrix (`v1.9.9`)** | `garuda.docs_ingest` | Ingests Markdown ADRs or specs from `docs/adr` and correlates them against workspace AST entities. |
| **3. Docs Lineage & Build Matrix (`v1.9.9`)** | `garuda.build_matrix_check` | Evaluates the 4-target cross-OS/CGO build matrix and returns the deterministic `BuildContext` receipt and 1-line reproduce command. |
| **3. Docs Lineage & Build Matrix (`v1.9.9`)** | `garuda.doctor_diagnose` | Runs the System Doctor check so the agent can verify DB, Talon contract, and gateway health. |
| **3. Docs Lineage & Build Matrix (`v1.9.9`)** | `garuda.hygiene_report` | Runs the static code hygiene analyzer to report unreferenced symbol candidates, duplicate names, and stdlib alternatives. |
| **4. Orientation & Discovery** | `garuda.briefing` | Session-start workspace orientation: scale, top hubs, active policies, and changes since last watermark. |
| **4. Orientation & Discovery** | `garuda.find_entity` | Searches entities by regex, kind, package, or file path, ranked by inbound caller fan-in. |
| **4. Orientation & Discovery** | `garuda.entities` | Lists entities in a workspace filtered by package or kind. |
| **4. Orientation & Discovery** | `garuda.inspect` | Inspects a single entity's full signature, file/line span, and incoming/outgoing claims. |
| **4. Orientation & Discovery** | `garuda.query` | Queries canonical architectural decisions by natural-language keyword. |
| **5. Graph Topology & Impact** | `garuda.neighbors` | Lists all 1-hop incoming and outgoing connected entities with edge direction and confidence. |
| **5. Graph Topology & Impact** | `garuda.blast_radius` | Computes transitive upstream impact (`CRITICAL`, `HIGH`, `MEDIUM`, `LOW`) up to depth `N`. |
| **5. Graph Topology & Impact** | `garuda.implementers` | Finds all concrete structs/classes implementing an interface (`IMPLEMENTS`). |
| **5. Graph Topology & Impact** | `garuda.subclasses` | Finds all types inheriting from or embedding a base type (`INHERITS`, `EMBEDS`). |
| **5. Graph Topology & Impact** | `garuda.get_impact` | Finds downstream decision-lineage impact if an architectural decision changes. |
| **5. Graph Topology & Impact** | `garuda.get_lineage` | Traverses parent and child decision supersession chains. |
| **6. Policy, Governance & Proofs** | `garuda.policy.list` | Lists registered declarative policies for the tenant. |
| **6. Policy, Governance & Proofs** | `garuda.policy.evaluate` | Dry-runs active YAML policies against the workspace graph. |
| **6. Policy, Governance & Proofs** | `garuda.verify_policy_evaluation` | Cryptographically verifies the RFC 6962 Merkle inclusion proof for a saved policy evaluation. |
| **6. Policy, Governance & Proofs** | `garuda.governance.status` | Returns current Merkle block height, root hash, active policies, and contradiction counts. |
| **6. Policy, Governance & Proofs** | `garuda.propose_decision` | Proposes a new scoped architectural decision draft. |
| **7. Epistemic Evidence & Drift** | `garuda.detect_contradictions` | Lists quarantined runtime OpenTelemetry and ADR contradictions (`CONTRADICTED`). |
| **7. Epistemic Evidence & Drift** | `garuda.check_drift` | Reports bidirectional Documentation-to-Code and Code-to-Documentation drift. |
| **7. Epistemic Evidence & Drift** | `garuda.query_claims` | Queries extracted normative ADR claims (`document_claims`) by subject or object. |
| **8. Multi-Agent Continuity** | `garuda.handoff` | Atomically transfers task ownership between agents and saves a Serializable state checkpoint. |
| **8. Multi-Agent Continuity** | `garuda.resume` | Consumes a handoff checkpoint (with replay protection) to resume multi-agent work. |


---

## 5. `v1.9.9` Additions: Dirty-Worktree Provenance Flags, Security Controls & Extended MCP Tools

### 5.1 Dirty-Worktree Provenance Flags on `garuda check` & `garuda talon guard`
When a working tree has uncommitted edits, `BuildContext.ReproduceCommand` emits `--worktree-dirty` and `--dirty-tree-hash <12-char-sha256>` so CI or another developer can verify the exact dirty-tree state:

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 garuda check \
  --target linux/amd64 \
  --cgo=false \
  --commit ef749a162e01 \
  --worktree-dirty \
  --dirty-tree-hash 417801f9a311 \
  --workspace default
```

- `--worktree-dirty` (`bool`) — Assert that the working tree is expected to be dirty when verifying the `BuildContext` receipt.
- `--dirty-tree-hash string` — Assert the exact 12-character SHA-256 digest of modified/untracked source files (`BuildContext.DirtyFilesHash`).

### 5.2 MCP Security & Attestation Environment Variables (`v1.9.9`)
- `GARUDA_MCP_READ_ONLY=true` — Blocks all 13 mutating MCP tools (`propose_decision`, `decide_logic_variant`, `reconcile_data_audit`, `reconcile_financial_audit`, `approve_base_rulebook`, `synthesize_rulebook`, `docs_ingest`, `create_task`, `handoff`, `resume`, `reset_audit_state`, `create_workspace`, and `talon_verify_edit` when `save_to_ledger=true`), returning a structured `{"status": "rejected", "read_only": true}` response.
- `GARUDA_MCP_APPROVER=<email-or-handle>` — Required for an MCP call to record `APPROVED_BY_HUMAN` or `RECONCILED_BY_HUMAN`. When unset, `decide_logic_variant`, `reconcile_data_audit`, and `approve_base_rulebook` record `PROPOSED_BY_AGENT_AWAITING_HUMAN_CLI_COSIGN` (`human_cli_verified: false`) until co-signed via the CLI or Web Dashboard.

### 5.3 New `v1.9.9` MCP Tools (Completing the 54-Tool Surface)
| Tool Name | Whitelisted Arguments | Purpose |
| :--- | :--- | :--- |
| `garuda.preflight_bundle` | `workspace`, `target`, `symbol`, `file`, `tenant_id` | Single-call pre-edit bundle combining `talon_pre_edit`, `blast_radius`, `neighbors`, `criticality_tiers`, `query_claims`, and `snapshot_receipt`. |
| `garuda.verify_receipt` | `workspace`, `receipt_hash`, `graph_fingerprint`, `commit_sha`, `tenant_id` | Cryptographically verifies a `snapshot_receipt` or `BuildContext` hash against the live PostgreSQL workspace and RFC 6962 Merkle root. |
| `garuda.diff_snapshots` | `workspace`, `from_revision`, `to_revision`, `tenant_id` | Diffs two workspace snapshots or git revisions, returning added/removed files, entities, claims, and policy posture delta. |
| `garuda.audit_export` | `workspace`, `format` (`csv`/`json`/`md`), `limit`, `tenant_id` | Exports the tamper-evident governance, policy, logic-variant, and data-audit ledger in CSV, JSON, or Markdown format. |
| `garuda.create_workspace` | `workspace`, `description`, `tenant_id` | Provisions an isolated workspace (returning an isolated empty `snapshot_receipt` until repositories are analyzed into it). |
| `garuda.audit_data_discrepancy` | `workspace`, `metric_name`, `team_a_name`, `team_a_crore`, `team_b_name`, `team_b_crore`, `tenant_id` | FinOps-safe alias for cross-team metric & unit-scale discrepancy auditing (`data_logic_audit_ledger`). |
| `garuda.reconcile_data_audit` | `workspace`, `audit_id`, `approved_by`, `resolution_note`, `tenant_id` | Reconciles a flagged data/logic discrepancy (requiring `GARUDA_MCP_APPROVER` for human sign-off). |
| `garuda.reset_audit_state` | `workspace`, `confirm`, `tenant_id` | Resets demo/sandbox logic-variant and data-audit overrides back to clean baseline state. |
