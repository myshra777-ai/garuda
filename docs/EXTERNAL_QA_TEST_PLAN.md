# Garuda External QA & GTM Readiness Test Plan

This plan is for an external tester evaluating Garuda before customer outreach. The goal is to discover reproducible bugs and product gaps, not to produce a positive review. Record clean results as carefully as failures.

## 0. Ground rules

- Record the Garuda version and full commit SHA. Do not test an old installation and attribute its behavior to current `main`.
- Use synthetic repositories, fake credentials, and non-production data. Do not put customer source code, credentials, tokens, private keys, or a complete `.env` file in a bug report.
- Do not assume that a successful graph extraction proves a relationship is complete. Garuda distinguishes supported evidence from unverified and contradicted evidence.
- Do not expect a self-repair prompt to fix every violation automatically. Record whether the gate correctly blocks, explains the rule, and offers useful repair guidance.
- Keep tests repeatable: one clean baseline, one intentionally breaking edit, one run with the edit staged, and one run after reverting it.
- Use a separate workspace for each independent project set. Name it clearly, for example `qa-payments` or `qa-typescript`.

## 1. Recommended codebase sizes

Start small enough that every expected relationship can be manually checked. Do not begin with a huge production repository; it makes incorrect output harder to diagnose.

| Stage | Size | Purpose |
|---|---|---|
| Baseline repo A | 20–60 source files; roughly 3,000–10,000 lines | Validate correct/incorrect edges, ADR rules, symbol lookup, policy verdicts, and diff scope. |
| Baseline repo B | 15–50 source files; roughly 1,000–8,000 lines | Validate another repository in the same workspace and cross-repository references where the language/build setup actually declares them. |
| Optional language fixtures | 10–30 files per fixture; roughly 500–3,000 lines | Isolate Python and TypeScript cases that are not covered by the two baseline repositories. |
| Medium-size pass | 10,000–30,000 lines per repository | Check practical usability, indexing stability, navigation, and incremental re-analysis. |
| Stretch/performance pass | 50,000–200,000 lines, only after the earlier stages pass | Record elapsed time, peak memory if available, database growth, and whether the result is usable. This is exploratory; do not treat it as a promised support limit or SLA. |

### Repository A: `orders-api` (Go)

Build a small, realistic service with these layers:

- `cmd/api`: wiring and HTTP startup.
- `internal/handler`: HTTP/request handlers.
- `internal/service`: business logic.
- `internal/repository`: persistence interface/implementation.
- `internal/domain`: shared types and business invariants.

**Intended allowed path:** handler → service → repository → database implementation.

**Intended forbidden path:** handler → database implementation directly.

Include a normal clean implementation plus a separate branch/commit with one deliberate violation (for example, a handler importing `database/sql` and executing a query directly). Keep the violation easy to identify in a short diff.

Suggested source files: `handler.go`, `order_service.go`, `order_repository.go`, `sql_order_repository.go`, and tests covering the normal flow. Do not make tests depend on a real payment provider or other third party.

### Repository B: `payments-contracts` or `orders-sdk` (Go module)

Use this as a second Git repository and workspace member. It should import a small, clearly identifiable type/interface from repo A using a supported local module/workspace arrangement. Keep the dependency explicit in `go.mod`/workspace configuration; document any `replace` directive used for local testing.

Use it to test cross-repository workspace resolution, shared-type impact, symbol ownership, and what happens when a public interface/type changes.

If cross-repository links are not resolved, record the exact modules, import paths, Go workspace configuration, and command used. Do not count a missing edge as a product bug until the workspace setup and supported resolution semantics have been confirmed.

### Language fixtures: Python and TypeScript

To test all three documented languages, create small isolated fixtures in their own directories or repositories; do not put multiple language markers into one root and assume Garuda will merge all analyzers automatically.

- **Python fixture:** modules/classes with direct imports, aliases, methods, and an instance call such as `self.repository.save(...)`.
- **TypeScript fixture:** interfaces/classes, named and namespace import aliases, and an instance call such as `new OrderService().create(...)`.

Keep a known list of 10–20 expected entities and relationships in each fixture. Dynamic reflection, plugin loading, dependency injection, and string-based dispatch are stretch cases: record them as uncertainty/coverage questions, not automatically as bugs.

## 2. Required architecture documents and policy inputs

Each baseline repository should contain:

1. A short `README.md` explaining purpose, modules, and how to run tests.
2. `docs/architecture.md` with the intended layer flow and the key invariants.
3. At least two ADRs in `docs/adr/`:
   - `0001-layering.md`: handlers MUST call the service layer and MUST NOT call repositories/database implementations directly.
   - `0002-shared-contracts.md`: public/shared interfaces have an identified owner and changes to them require checking callers.
4. A `policies/` directory with a small policy bundle. Start from the current checked-in examples instead of inventing YAML syntax:
   - [payment-no-direct-db.yaml](../policies/payment-no-direct-db.yaml)
   - [architecture-layer-guardrails.yaml](../policies/architecture-layer-guardrails.yaml)
   - [require-runtime-evidence.yaml](../policies/require-runtime-evidence.yaml) (optional; use only for the configured runtime-evidence scenarios).

Adapt names and package patterns to the sample modules. Keep the first policy bundle small enough that each rule's expected outcome is clear. Record the exact policy file and rule name for every expected block.

Use normative, explicit wording. Example ADR structure:

```markdown
# ADR-0001: Enforce handler/service boundaries

## Status
Accepted

## Context
HTTP handlers should not own persistence behavior.

## Decision
- Handlers MUST call the appropriate service for business operations.
- Handlers MUST NOT call database implementations directly.
- Services MAY depend on repository interfaces.
- Database implementations MUST remain behind repository interfaces.

## Consequences
A new direct handler-to-database dependency must be reviewed and blocked by the configured architecture policy.
```

This text is a fixture for testing ingestion. It is not a promise that every natural-language statement will be extracted; the tester must verify which claims Garuda actually ingests and evaluates.

## 3. Test sequence and acceptance criteria

Run each stage in order. Record command output, pass/fail, elapsed time, and any manual setup.

### A. Installation, upgrades, and recovery — P0

- Fresh Windows PowerShell 5.1 installation with no Go/PostgreSQL preinstalled.
- Fresh Windows PowerShell 7 installation if available.
- First install when port 8080 is occupied; expected behavior is an actionable conflict message, not killing an unrelated process silently.
- First install when port 5433 is occupied by a non-PostgreSQL listener; installer must not complete successfully based only on a TCP connection; it must fail clearly when the Garuda CLI cannot authenticate to a PostgreSQL database.
- PostgreSQL not yet ready; verify the installer waits for database readiness, or fails with a focused error and the PostgreSQL log location.
- Network interruption/download failure during portable toolchain download; verify retry and cleanup behavior.
- Re-run setup after successful installation; verify idempotence and no duplicate accounts/workspaces/migrations.
- Upgrade an installation with a clean Git checkout.
- Attempt upgrade with a modified tracked file or unpublished local commit; verify the installer refuses to discard work.
- Test a path containing spaces, e.g. `C:\Users\QA Tester\garuda`.
- If possible, repeat install on macOS/Linux to compare the setup contract with Windows.

**Pass:** commands return truthful exit codes; output stays readable; a failed step does not print an unconditional “100% configured” success; logs identify the failing subsystem; no local source changes are destroyed silently.

### B. Indexing and language analysis — P0

- Analyze each baseline repository from a clean checkout and save the result.
- Repeat analysis without changes. Compare entity and relationship counts and stable identifiers/fingerprints where the documented contract says they should be stable.
- Check 10–20 manually selected symbols and relationships against source.
- Include import aliases, receiver/instance calls, interfaces/implementations, multiple modules, build tags/platform conditionals in Go, and a deleted or renamed source file.
- Run analysis from a path with spaces and a path containing non-ASCII characters.
- Run with an unsupported or malformed project marker; the message must explain what is unsupported rather than silently reporting an empty success.

**Pass:** analysis either returns valid, inspectable output or fails with a clear error. Record both false edges and expected edges that are missing, with file/line evidence.

### C. Docs-to-code and policies — P0

- Ingest the ADR directory and inspect each ingested claim, source path, and line.
- Test one clean rule and one deliberate forbidden edge.
- Change the offending file, then evaluate the diff.
- Add a second unrelated clean change; verify it is not blocked by a rule that does not apply.
- Introduce one pre-existing violation before establishing a baseline, then add one new violation. Check that baseline and net-new findings are distinguished as documented.
- Test missing docs, malformed Markdown, malformed YAML, an unsupported predicate, an empty policy directory, and a policy pattern that matches nothing.
- Test a policy contradiction and an unverified claim separately.

**Pass:** the verdict cites the relevant policy/ADR and source location when available; malformed/unsupported policy input is not silently treated as a clean pass; UNKNOWN/UNVERIFIED evidence is not promoted to certainty without evidence.

### D. Diff gate and CI behavior — P0

- Clean diff: expected ALLOW (or the documented non-blocking outcome).
- Diff with one forbidden new dependency: expected BLOCK.
- Existing baseline issue without new violation: should not be misrepresented as a newly introduced issue.
- Stage changes and run the staged mode.
- Run a commit/diff range and a deletion/rename case.
- Run with `--fail-on-block`; verify a blocking verdict exits with code 2 and a clean case returns the documented success code.
- In a disposable CI workflow, confirm the shell/runner propagates the CLI exit code. Do not change production branch protection for this test.

**Pass:** the exact tested revision is checked, and the result/exit code matches the findings. A CLI block alone does not prove that a remote repository prevents merges; branch protection is a separate configuration.

### E. Garuda Talon agent workflow — P0

- Inspect generated `.garuda/TALON_CONTRACT.md` and the applicable agent configuration files.
- Request a pre-edit briefing for a high-impact symbol; confirm that relevant symbols, blast radius, ADRs and active policies are shown.
- Introduce a forbidden edit and run post-edit verification.
- Confirm a BLOCK result includes the violated rule and repair-oriented guidance.
- Run a clean edit and verify it is not falsely blocked by an unrelated rule.
- Test a malformed policy or missing ADR and inspect the failure mode.
- Test each client that is actually available to the tester; label other integrations “not tested” instead of assuming all clients behave identically.
- If using the bounded supervisor/self-repair flow, record the number of attempts and the final state. Do not require an automatic repair to succeed as the sole acceptance criterion.

**Pass:** the configured hook actually runs, surfaces the result, and returns the documented status. A prompt or editor hook should not be described as an unbypassable security boundary.

### F. MCP, dashboard, impact, and workspace boundaries — P1

- Run `scripts/mcp_verify.py` and save the full summary.
- Check tool discovery, argument validation, an unknown entity, a known entity, neighbors/callers, and blast radius.
- Test handoff/resume with a minimal context and then a larger payload; verify checkpoint continuity and collision behavior.
- Inspect dashboard overview, architecture map, change review, policy surfaces, contradictions/evidence, and ledger views.
- Use symbol/package/repository search with an exact query, partial query, no results, special characters, and large result counts.
- Create two workspaces with distinguishable test repositories; verify reads and writes stay in the intended workspace/tenant.
- Test an unavailable database, expired/invalid session, malformed request, and an absent workspace.
- Test repeated init, repeated save, duplicate requests/idempotency, and concurrency where safe.

**Pass:** errors are explicit, workspace selection is respected, and persisted state is consistent with CLI output. Screenshots are useful for layout defects, but include the underlying request/command and response as evidence.

### G. Evidence ledger and runtime observations — P1

- Save a governance decision/evaluation and verify its record/inclusion proof using the documented verifier.
- Repeat a save to inspect idempotency and revision behavior.
- Verify that editing a persisted record or supplied proof causes verification failure in the controlled test.
- Confirm preview-only evaluations are not described as ledger-sealed records.
- If telemetry is configured, send synthetic spans for one declared relationship and one undeclared/contradictory relationship. Verify that runtime states match the documented semantics.
- Confirm missing telemetry remains UNVERIFIED rather than being reported as proof that a relationship never occurs.

**Pass:** record-integrity results are not presented as proof that semantic analysis is complete or that every runtime call was observed.

### H. Usability and GTM readiness — P1

- Ask a developer who has not seen Garuda to install it using only the published instructions.
- Ask them to explain, in their own words, the difference between the architecture map, policy result, Talon block, and ledger evidence.
- Time onboarding steps and record where they hesitate; do not turn a single observation into a general performance claim.
- Ask whether they would keep the gate enabled in a real repository, what false positives they would tolerate, who owns architecture rules, and what must be true before they would pay.
- Distinguish “interested in testing,” “design partner,” “active user,” and “paying customer” in all reporting.

## 4. Bug severity

- **P0 — Stop-ship:** installation fails on a supported clean environment, corrupts/removes user data, reports success after a critical failure, silently passes a configured blocking violation, or crosses workspace/tenant boundaries.
- **P1 — High:** a core workflow fails without a reliable workaround, an important relationship/rule is systematically missed in a supported fixture, or a configured hook/CI gate does not enforce its documented behavior.
- **P2 — Medium:** incorrect or confusing output with a workaround, unstable UI behavior, misleading messages, or incomplete diagnostics.
- **P3 — Low:** polish, non-blocking layout issues, minor documentation typos.

## 5. Bug report template

Copy this for every issue:

```text
Title:
Severity (P0/P1/P2/P3):
Garuda version:
Git commit SHA:
OS + version:
PowerShell/shell version:
Fresh install or upgrade:
Install path (redact username if desired):
Workspace name (synthetic):
Repo language and approximate LOC:
Feature/command under test:
Exact steps to reproduce:
Expected result:
Actual result:
Frequency (1/1, 3/3, intermittent):
Relevant stdout/stderr (redact secrets):
PostgreSQL log excerpt (if install/database issue):
API log excerpt (if API/dashboard issue):
Screenshot/video (optional):
Minimal files or synthetic repo commit that reproduces it:
Workaround, if any:
```

**Never attach `.env`, JWT private keys, API tokens, database passwords, browser cookies, or real customer code.** For the reported Windows database failure, collect the last 100 lines from `%USERPROFILE%\.garuda\data\postgres.log`, plus the setup output and `garuda doctor` output after redacting personal paths and emails.

## 6. Final exit criteria before cold outreach

Do not treat all tests as equal. Before outreach, require:

- P0 and P1 issues fixed or explicitly accepted with a documented workaround.
- A clean install and clean upgrade smoke test on the supported platforms you intend to claim.
- At least one known-clean and one known-blocking diff per supported language fixture.
- CI non-zero behavior verified in a disposable repository.
- A report of false positives, false negatives, unsupported constructs, and tester-not-tested areas.
- A clear onboarding guide and a short demo matching the actual tested behavior.

The deliverable from this test phase is a reproducible bug backlog and a truthful capability/coverage report—not a promise that all possible bugs have been eliminated.
