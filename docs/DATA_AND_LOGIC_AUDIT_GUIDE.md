# Garuda Data & Business Logic Audits Guide (`v1.9.9`)

**Resolving Cross-Team Metric Discrepancies, Active Architectural Contradictions, and Undocumented Formula Drift with 0.0–1.0 Team Weights—Without Fintech Regulatory Overhead**

---

## 1. Executive Overview: Why Teams Need Data & Logic Audits

In modern software engineering, **core business logic is split across multiple teams and codebases**. Operations, Marketing, Product Growth, and Management each write code (or direct AI agents) to calculate critical business metrics—such as User Retention, Loss Ratios, Unit Economics, Conversion Funnels, and Settlement Totals.

Historically, when two teams compute contradictory numbers for the same conceptual metric:
1. **Tests Pass**: Each team's unit tests pass because their local logic runs without crashing.
2. **Linters Pass**: The syntax and types are completely valid.
3. **The System Drifts**: Nobody notices that the marketing formula was copied into the core pricing router until an outage, revenue leak, or audit failure occurs.

### Why We Call It "Data & Logic Audits" (Not Financial Audits)

> **Important Positioning & Regulatory Protection**:
> Calling business metric validation "Financial Logic" or "Financial Auditing" causes external stakeholders and government regulators to categorize the product under **Fintech / Financial Services** (subject to strict statutory auditor certification, financial advice liability, and banking compliance approvals like RBI, SEBI, SEC, or FINRA).
>
> **Garuda is an Engineering Architecture & Logic Governance Platform**. It operates at the AST code-graph and documentation level to verify that software adheres to human-approved business contracts and architectural boundaries. Renaming to **Data and Logic Audits** accurately describes our technical domain and eliminates unnecessary regulatory friction.

---

## 2. Understanding Active Contradictions

Gargi and engineering testers frequently ask:
> *"What exactly is an active contradiction? What happens when there ARE active contradictions in a repository?"*

In Garuda, an **Active Contradiction** occurs when the codebase's real AST relationships violate an explicit, approved architectural constraint or decision. There are three distinct categories:

### Category 1: Documentation vs. Code Contradiction (ADR Drift)
- **The Rule**: In `docs/adr/0004-no-direct-db-in-handlers.md`, the architecture team states:
  `- OrderHandler MUST NOT call DirectDB.RawQuery directly`
- **The Code**: An AI coding agent writes `order/handler.go` and calls `h.db.RawQuery(...)` directly to save time.
- **The Detection**: Garuda parses the AST graph, detects the forbidden call edge `OrderHandler -> DirectDB.RawQuery`, and marks the claim status as `CONTRADICTED` with:
  `ARCHITECTURE CONTRADICTION: Forbidden call relationship found in AST`.
- **The Consequence**: `garuda check` exits with a non-zero exit code (`ADR Contradictions: 1`), and the CI gate immediately **BLOCKS** the pull request.

### Category 2: Decision vs. Decision Contradiction (Pre-Flight Shield Quarantine)
- **The Existing State**: Decision #101 approved: *"Use PostgreSQL for session store"* in scope `auth/session`.
- **The Proposed State**: Decision #204 proposed: *"Use Redis for session store"* in scope `auth/session` without explicitly deprecating or superseding Decision #101.
- **The Detection**: Garuda's Contradiction Engine detects conflicting technology tokens in the same domain/system scope.
- **The Consequence**: Decision #204 is placed in **Merkle Quarantine** (`status: QUARANTINED`) and cannot be executed by AI agents.

### Category 3: Cross-Team Formula Discrepancy (Unit Scaling & Metric Variance)
- **The Problem**: Ops calculates `Retention = active_30d / registered`, while Marketing calculates `Retention = active_30d / (registered - churned_trials)`.
- **The Detection**: Garuda discovers both formulas in the AST graph, flags the calculation variance, and assigns 0.0–1.0 weights based on the active team context.
- **The Consequence**: Garuda does **NOT** let AI guess which formula to use. It displays the blast radius, severity tier, and requires a human engineer to approve the canonical version.

---

## 3. Human-in-the-Loop 0.0–1.0 Formula Weighting & Blast Radius

When competing formulas exist for the same concept, Garuda assigns normalized weights (0.0 to 1.0) and computes the **Blast Radius**:

| Team Role | Default Weight | Severity Tier | Downstream Inbound Callers | Human Impact & Policy Action |
| :--- | :--- | :--- | :--- | :--- |
| **Ops Team** | `0.70` | `LOW` | `4 callers` | **Least Architectural Harm**: Isolated to internal monitoring dashboards. Safe to standardize. |
| **Marketing Team** | `0.10` | `MEDIUM` | `19 callers` | **Non-Blocking Trade-off**: Connected to acquisition loaders. Requires human trade-off review. |
| **Management / Core** | `0.10` | `HIGH_SEVERE` | `142 callers` | **Severe Blast Radius**: Connected to Core Homepage & Search Results Redirect Router. **FROZEN — DO NOT TOUCH!** |

### The Golden Rule: AI Informs, Humans Decide
Garuda **never** lets an LLM or autonomous agent silently decide which formula is correct. It provides:
1. Exact file paths and line numbers.
2. Inbound caller counts (fan-in).
3. The severity level (LOW, MEDIUM, HIGH_SEVERE).
4. Recommended least-harm action.
A human engineer or architect reviews the diff and approves the canonical rulebook with 1 click.

---

## 4. CLI & MCP Commands

### Inspecting Competing Formulas
```bash
./bin/garuda audit logic "Retention Rate" --team ops
```

### Auditing Cross-Team Metric Scaling Variances
```bash
./bin/garuda audit data-logic --metric "Active User Retention" --team-a-val 48.0 --team-b-val 4.8
```

### Ingesting ADR Rules and Verifying Against Code AST
```bash
./bin/garuda docs ingest test/fixtures/active_contradictions/docs/adr/0004-no-direct-db-in-handlers.md
./bin/garuda docs verify
```

### Live Demo Harness
```bash
bash scripts/demo_active_contradiction.sh
```
