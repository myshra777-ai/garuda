# Why Linting Isn't Enough for AI Coding Agents: Deterministic AST + RFC 6962 Merkle Governance in Regulated Software

*Ready-to-publish technical deep-dive for LinkedIn, Dev.to, Hashnode, and Hacker News.*

---

Autonomous coding agents (Claude Code, Cursor, Codex, Windsurf) can write 1,000 lines of syntactically valid Go, Python, or TypeScript in under a minute. Every linter passes. Every unit test in the touched package passes.

And yet, in regulated domains—fintech, insurance, banking, healthcare—three classes of silent architectural failures slip through daily:

1. **Boundary Bypass**: An agent imports `database/sql` directly inside a payment handler instead of routing through your audited, idempotency-enforced `PaymentRepository`.
2. **Silent Formula Overwrites & Unit-Scale Drift**: The Ops team defines `Loss Ratio` or `Claims Paid` using a `1e6` (10 Lakh) divisor while Finance uses `1e7` (1 Crore), or an agent introduces `float64` math into currency settlement—creating a **10x discrepancy (`48 Crore` vs `4.8 Crore`)** that stays hidden until an external audit.
3. **High-Blast-Radius Edits Without Caller Context**: An agent modifies a shared signature with 142 transitive callers across 6 services without knowing the surface is classified as `SEVERE` criticality.

Here is how we built **Garuda** to solve this deterministically—and how you can run it on a 200-line sample codebase in 60 seconds.

---

## 1. The 5-Plane Architecture: AST + Auto-Rulebook + Weighted Formulas + Merkle Provenance

Unlike LLM-based code reviewers that guess whether a diff looks safe, Garuda operates on five deterministic planes:

1. **Compiler-Grade AST & Call Graph (`internal/analyzer`)**: Extracts every symbol, package boundary, `IMPORTS` edge, and `CALLS` edge, bound to a deterministic `BuildContext` (`GOOS/GOARCH`, `CGO_ENABLED`, build tags, and exact Git tree SHA).
2. **Auto-Synthesized Base Rulebook (`garuda rulebook generate`)**: Most teams don't have time to hand-author 50 YAML rules from scratch. Garuda reverse-engineers your codebase's implicit layer boundaries, criticality tiers (`SEVERE`, `HIGH`, `MEDIUM`, `LOW`), and team formulas into `.garuda/DRAFT_BASE_RULEBOOK.md` for human review and 1-click approval.
3. **Human-in-the-Loop Weighted Logic Disambiguation (`0.0–1.0`)**: When competing formulas exist across teams (e.g., Ops vs Marketing vs Management), Garuda computes `0.0–1.0` role weights and blast-radius harm scores—refusing to let an AI agent silently overwrite a formula without an attested human CLI co-sign (`GARUDA_MCP_APPROVER`).
4. **Data & Logic Audit Ledger (`garuda audit data-logic`)**: Traces AST and SQL constant divisors across pipelines to catch unit-scale mismatches (like `1e6` vs `1e7`) before numbers ever reach an executive dashboard.
5. **Dual-Root RFC 6962 SHA-256 Merkle Ledger (`internal/store/merkle_v1.go`)**: Every policy evaluation, rulebook approval, and agent verification is cryptographically anchored with an inclusion proof.

---

## 2. See It Live in 60 Seconds: `garuda demo`

We shipped a built-in `garuda demo` command that generates a 200-line fintech microservice (`acme-pay`) with 3 intentional violations and runs the entire pipeline:

```bash
$ garuda demo
╔══════════════════════════════════════════════════════════════════════════════════╗
║     🦅 GARUDA LIVE PIPELINE DEMO — 200-LINE FINTECH MICROSERVICE (acme-pay)      ║
╚══════════════════════════════════════════════════════════════════════════════════╝
  [1/5] Sample Codebase Generated: /tmp/garuda-demo-acme-pay (156 lines across 4 Go files)
  [2/5] AST & Call-Graph Extraction:
        ✓ Found N entities and M relationships in X ms   (real numbers; they depend on your machine)
  [3/5] Auto-Synthesized Base Rulebook:
        ✓ Reverse-engineered 4 layer & financial invariants into .garuda/DRAFT_BASE_RULEBOOK.md
  [4/5] Evaluating Declarative Policies & Data/Logic Audit Ledger:
        ✗ [BLOCK · CRITICAL] tpl-financial-no-direct-sql
          Location: internal/payment/checkout.go:5 (CheckoutHandler.ProcessInstantCheckout)
          Reason:   Payment code imports 'database/sql' directly instead of domain.PaymentRepository.
        ✗ [BLOCK · CRITICAL] tpl-financial-decimal-precision + DATA-LOGIC-AUDIT
          Location: internal/payment/settlement.go:7 (CalculateSettlementFeeFloat64 / ComputeOpsClaimsPaidCrore)
          Reason:   Uses float64 for currency math AND divides by 1e6 instead of 1e7 (48 Cr vs 4.8 Cr 10x drift).
        ✗ [BLOCK · HIGH]     tpl-layered-no-upward-deps
          Location: internal/store/ledger_repo.go:8 (PostgresLedgerRepo.RecordChargeTx)
          Reason:   Persistence layer 'internal/store' imports upper service layer 'internal/payment'.
```

Instead of just failing the build, Garuda Talon emits deterministic **Self-Repair Instructions** directly back to Claude Code, Cursor, or Codex via MCP (`garuda.talon_verify_edit`) so the agent fixes its own architectural violation before the developer even reviews the PR.

---

## 3. Get Your First 3 Rules Running in 10 Minutes

1. **Initialize with a pre-built template** (or run the 5-question interactive wizard):
   ```bash
   garuda init . --template financial
   # or: garuda init . --wizard
   ```
2. **Import existing rules** from Java ArchUnit, JS/TS Dependency-Cruiser, or Python Import Linter:
   ```bash
   garuda import --from archunit
   ```
3. **Approve your reverse-engineered Base Rulebook & gate your CI**:
   ```bash
   garuda rulebook generate --approve
   garuda check . --diff HEAD --pr-comment garuda-pr.md --sarif garuda.sarif --fail-on-block
   ```

Check out the repository and the 10-minute quickstart at **[github.com/myshra777-ai/garuda](https://github.com/myshra777-ai/garuda)**. We're actively partnering with engineering teams in fintech, insurance, and regulated SaaS—if you're governing AI coding agents in production, we'd love your feedback!
