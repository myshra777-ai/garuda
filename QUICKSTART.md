# 🦅 Garuda Quickstart: Your First 3 Rules in 10 Minutes

Garuda governs autonomous coding agents (Claude Code, Cursor, OpenAI Codex, Windsurf) and CI/CD pipelines using **compiler-grade AST call graphs**, an **Auto-Synthesized Base Rulebook**, **weighted multi-team formula disambiguation (`0.0–1.0`)**, and an **RFC 6962 SHA-256 Merkle ledger**.

---

## ⚡ Try It in 60 Seconds (Zero Setup Demo)

Want to see what Garuda is built to catch, on a 200-line sample fintech microservice, before configuring anything? The demo needs no database.

```bash
./bin/garuda demo          # Windows: .\bin\garuda.exe demo
```

This spins up `acme-pay` (4 Go files, ~200 lines) with 3 intentional violations:
1. **`[CRITICAL]` Direct `database/sql` bypass in Payment Handler** (`internal/payment/checkout.go:5`)
2. **`[CRITICAL]` `float64` currency rounding + 10x Crore/Lakh unit scale drift (`48 Cr` vs `4.8 Cr`)** (`internal/payment/settlement.go:7`)
3. **`[HIGH]` Upward layer dependency (`internal/store` → `internal/payment`, created by `garuda demo`)** (`internal/store/ledger_repo.go:8`)

The demo writes this sample, runs the real Garuda Go analyzer on it (the entity and relationship counts and the timing are measured, and analyzing Go code needs a Go toolchain), and lists the three planted problems with the repair instructions an agent would receive. It does not evaluate policies or seal a Merkle receipt: those need a database. Finish [INSTALL.md](INSTALL.md), then use the steps below to see real checks.

---

## 🚀 Your First 3 Rules in 10 Minutes (The Happy Path)

### Step 1: Initialize With a Starter Template (or the 5-Question Wizard) — *2 mins*

Pick one of Garuda's **10 pre-built starter policy templates** (`hexagonal`, `layered`, `microservices`, `no-circular-imports`, `financial`, `zero-trust-auth`, `ai-agent-safety`, `pii-gdpr-hipaa`, `idempotent-mutations`, or `all`):

```bash
# Option A: One-command init with a starter template
./bin/garuda init . --template financial

# Option B: Interactive 5-question CLI wizard (auto-detects your stack & critical paths)
./bin/garuda init . --wizard
```

What just happened:
- Provisioned local/Docker PostgreSQL, ran migrations, and indexed your repository's AST & call graph.
- Installed starter YAML policies into `./policies/`.
- Configured MCP & active pre-edit/post-edit hooks for Claude Code, Cursor, Codex, and Git.
- Reverse-engineered `.garuda/DRAFT_BASE_RULEBOOK.md` directly from your codebase.

### Step 2: Review Your Auto-Synthesized Base Rulebook in the Dashboard — *3 mins*

Start the Garuda Gateway (or run `./setup.sh` which starts it automatically) and open **`http://localhost:8080/dashboard`**:

1. Navigate to **Surface `04 — Docs & Rules Graph`** → click tab **`02 · Base Rulebook Studio`**.
2. Inspect the reverse-engineered layer boundaries, **Surface Criticality Tiers (`SEVERE / HIGH / MEDIUM / LOW`)**, and **Multi-Team Formula Weights (`0.0–1.0`)**.
3. Edit the Markdown directly in the studio if needed.

### Step 3: Approve & Seal Your Base Rulebook — *2 mins*

Click **✓ Approve & Seal Rulebook** in the Dashboard, or run from your terminal:

```bash
./bin/garuda rulebook generate --approve --approver "Staff Architect"
```

This writes `.garuda/APPROVED_BASE_RULEBOOK.md` and anchors its SHA-256 digest into the RFC 6962 Merkle ledger.

### Step 4: Run Your First Diff-Scoped Governance Check — *3 mins*

Evaluate your working tree or PR diff against your active policies and approved rulebook:

```bash
./bin/garuda check . \
  --diff HEAD \
  --policy policies \
  --pr-comment garuda-pr-report.md \
  --sarif garuda.sarif
```

To send instant violation alerts to Slack or Microsoft Teams:

```bash
./bin/garuda check . --diff HEAD --notify-slack "https://hooks.slack.com/services/T00/B00/X00"
```

---

## 📦 Pre-Built Starter Policy Templates (`policies/templates/`)

List or install any template at any time with `./bin/garuda templates list` or `./bin/garuda templates install <name>`:

| Template Key | File | What It Enforces |
|---|---|---|
| `hexagonal` | `policies/templates/hexagonal.yaml` | Domain/Core packages must never import infrastructure adapters (`database/sql`, `http`, `redis`, `grpc`). |
| `layered` | `policies/templates/layered.yaml` | Persistence (`store`/`repo`) must never import service or transport (`api`/`handler`) layers. |
| `microservices` | `policies/templates/microservices.yaml` | Services must never call or import another microservice's internal persistence store directly. |
| `no-circular-imports` | `policies/templates/no-circular-imports.yaml` | Blocks bidirectional package dependency cycles across internal modules. |
| `financial-no-direct-sql` | `policies/templates/financial-no-direct-sql.yaml` | Payment, refund, and billing code must route through an audited Repository interface, never `database/sql`. |
| `financial-decimal-precision` | `policies/templates/financial-decimal-precision.yaml` | Monetary, fee, tax, and loss-ratio calculations must use fixed-point Decimal or `int64` minor units, never `float64`. |
| `zero-trust-auth` | `policies/templates/zero-trust-auth.yaml` | Exported API handlers must enforce `tenant_id` and `workspace_id` scope before store queries. |
| `ai-agent-safety` | `policies/templates/ai-agent-safety.yaml` | Freezes `SEVERE` criticality Merkle & auth signing code against autonomous AI edits without human co-sign. |
| `pii-gdpr-hipaa` | `policies/templates/pii-gdpr-hipaa.yaml` | Blocks customer PII/PHI structs from flowing into unredacted stdout/file loggers. |
| `idempotent-mutations` | `policies/templates/idempotent-mutations.yaml` | External payment/transfer mutations must verify an idempotency key via `IdempotencyStore`. |

---

## 🔄 Import Existing Rules (ArchUnit, Dependency-Cruiser, Import Linter)

Already have architecture tests in Java, TypeScript, or Python? Translate them directly into Garuda YAML without re-authoring:

```bash
# Java ArchUnit -> policies/imported-archunit.yaml
./bin/garuda import --from archunit --input src/test/java/ArchitectureTest.java

# JS/TS Dependency-Cruiser -> policies/imported-dependency-cruiser.yaml
./bin/garuda import --from dependency-cruiser --input .dependency-cruiser.json

# Python Import Linter -> policies/imported-importlinter.yaml
./bin/garuda import --from importlinter --input .importlinter
```

---

## 📜 Versioned Rulebooks in Git & CI/CD Workflow

Rulebooks are first-class policy artifacts that live in Git alongside your code:

1. **Commit `.garuda/APPROVED_BASE_RULEBOOK.md` and `policies/*.yaml` to Git.**
2. **PR Diff Review**: Any change to `.garuda/APPROVED_BASE_RULEBOOK.md` or `policies/*.yaml` is diffable in GitHub PRs and requires human architect co-sign (`garuda rulebook generate --approve`).
3. **GitHub Actions Gate**: Copy `.github/workflows/garuda-gate.yml` into your repository. On every PR, it runs `garuda check` against the PR diff, posts a Markdown table on the PR, uploads `garuda.sarif` to GitHub Code Scanning, and fails the workflow with exit code `2` on `BLOCK` violations.

---

## 🏢 Multi-Repo & Monorepo Support

For monorepos with 10+ microservices (`services/payments`, `services/orders`, `packages/core`), run:

```bash
./bin/garuda monorepo init .
```

Garuda discovers every service root (`go.mod`, `package.json`, `pyproject.toml`, `Cargo.toml`, `pom.xml`) and writes `.garuda/monorepo.yaml`, composing root baseline policies (`policies/*.yaml`) with service-local policies (`services/*/policies/*.yaml`) while enforcing cross-service database isolation.
