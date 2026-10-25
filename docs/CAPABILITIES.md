# 🧠 Garuda Capabilities & Verification Matrix (v1.6.0 Universal)

> Grounded in the Garuda v1.6.0 multi-language compiler/AST engine (`go/types`, Python AST, and dual Tree-sitter / Pure-Go `CGO_ENABLED=0` TypeScript AST), Garuda Talon active agent execution harness, diff-scoped graph policy evaluator, 3-state runtime verifier, RFC 6962 Merkle ledger, and zero-dependency cross-OS installer matrix.

This is a current capability guide—not a certification of production readiness. Feature behavior depends on the current binary, database migrations, workspace state, supported syntax, configured rules and integration path.

## At a glance

| Capability | Status | Verification Tier | Supported Semantics | Enforced Invariant |
| :--- | :--- | :--- | :--- | :--- |
| **Go Struct, Interface & Field Extraction** | 🟢 **Production (GA)** | `100% Benchmark (24/24 Strict Suite)` | Struct tags (`json`, `db`, `validate`), line spans, pointer/slice flags, method sets | `ACGM Resolution Invariant` |
| **Go Method & Receiver Disambiguation** | 🟢 **Production (GA)** | `100% Benchmark (002-method-identity)` | Pointer and value receivers (`ReceiverType`), exported/unexported methods | `Canonical UUIDv5` |
| **Go Cross-Package, Intra-Package & `go.work` `CALLS`** | 🟢 **Production (GA)** | `go/types` + `go.work` + `vendor/` (`1.0` confidence) | Direct intra-package calls, cross-package calls, and multi-module `go.work` workspace calls | `AST_EXACT Provenance` |
| **Go Interface Satisfaction (`IMPLEMENTS`)** | 🟢 **Production (GA)** | `100% Benchmark (003, 009, 010, 024)` | Type-checked `types.Implements` + cross-module `go.work` structural satisfaction at `1.0` confidence | `Full Recall Gate` |
| **Go Struct & Interface Embedding (`EMBEDS`)** | 🟢 **Production (GA)** | `100% Benchmark (006-embedding)` | Anonymous struct fields and embedded interfaces | `Composition Invariant` |
| **Sub-Second Incremental Working-Tree Sync** | 🟢 **Production (GA)** | `garuda analyze --incremental --files` | Shared in-memory package cache + targeted file re-indexing for dirty working-tree edits | `Deterministic Graph Parity` |
| **Python AST, Alias & Instance Call Graph** | 🟢 **Production (GA)** | `Multi-Repo Corpus + Alias/Receiver Tests` | Modules, classes, methods, functions, `import X as Y`, `from X import Y as Z`, and `self.attr.method()` / instance calls (`0.95` confidence) | `Deterministic SHA-1 UUID` |
| **TypeScript Tree-Sitter & Pure-Go (`!cgo`) AST Graph** | 🟢 **Production (GA)** | `Tree-sitter (cgo) + Pure-Go Lexical/Scope (!cgo)` | Interfaces, classes, functions, named/namespace import aliases (`import { X as Y }`, `import * as NS`), `new Class()` instance calls (`0.95` confidence, zero C compiler needed) | `Deterministic SHA-1 UUID` |
| **Cross-Repository Dependency Resolution** | 🟢 **Production (GA)** | `Postgres Workspace Symbol Cache` | Multi-repository module path and symbol bridge mapping within a workspace | `Workspace Isolation` |

## Garuda's differentiator: connect the loop

### 1. Before the edit
Talon can compile active ADR claims, YAML policies and high-impact symbols into a workspace contract. A pre-edit briefing gives the agent or developer relevant rules and graph-visible impact before a change is made.

### 2. After the edit
Talon can re-index dirty files, target the Git diff, check net-new violations, verify relevant ADR claims, and return a structured verdict plus repair-oriented instructions. Its supervisor command can rerun a CLI coding agent with that feedback for a configured number of attempts.

### 3. Before the merge
`garuda check` or a policy evaluation can be wired into CI. With `--fail-on-block`, a `BLOCK` verdict exits with code `2`; the CI provider can fail the job and a required branch-protection check can stop the merge. Garuda does not configure repository branch protection on its own.

### 4. After the decision
When a decision is persisted, the Merkle implementation can supply an inclusion proof. That receipt lets reviewers verify the integrity and inclusion of recorded decision data; it does not make the original analysis complete by itself.

## Evidence terminology

| Term | Meaning |
|---|---|
| `SUPPORTED` | Available evidence supports a claim within the checked scope |
| `UNVERIFIED` | Evidence is insufficient; not proof of a defect |
| `CONTRADICTED` | Available evidence conflicts with a claim |
| `ALLOW` | The configured evaluator found no blocking condition in the evaluated scope |
| `WARN` / `REVIEW` | A non-blocking warning or review outcome |
| `BLOCK` | An active policy/evidence condition produced a blocking outcome |

A passing result applies to the scope actually evaluated. Confirm index freshness, touched-symbol mapping, policy files, telemetry availability and integration failure behavior before using it as a mandatory merge gate.

## Current validation record

The current [evidence report](../EVIDENCE.md) records:

- 24/24 strict compiler/AST benchmark cases passed.
- 30/30 reference MCP/Talon verifier checks passed.
- 10/10 injected contradictions detected in the controlled runtime test.
- A 14-repository Go-only validation corpus and a separate 9-repository mixed-language workspace.

| Capability | Status | Verification Tier | Supported Semantics | Enforced Invariant |
| :--- | :--- | :--- | :--- | :--- |
| **Active Contract Compilation (`garuda talon init`)** | 🟢 **Production (GA)** | `internal/talon/talon.go` | Synthesizes live ADRs (`document_claims`), YAML policies, and top hubs into `.garuda/TALON_CONTRACT.md`, `.cursor/rules/garuda-talon.mdc`, `CLAUDE.md`, `AGENTS.md`, and `.windsurfrules` | `Always-On Agent Context` |
| **Active MCP Session Enforcement (`GARUDA_TALON_ENFORCE=true`)** | 🟢 **Production (GA)** | `30/30 checks in scripts/mcp_verify.py` | 23 MCP tools + `prompts/list` (`garuda.architectural_contract`) + `resources/list` (`garuda://workspace/contract`) + automatic first-call guardrail injection | `Fail-Closed Agent Orientation` |
| **Pre-Edit & Post-Edit Guard (`garuda.talon_pre_edit` / `verify_edit`)** | 🟢 **Production (GA)** | `internal/talon` + `cmd/garuda-mcp` | Pre-flight blast radius & ADR check before edits; incremental AST sync + diff-scoped policy check + deterministic Self-Repair Prompt after edits | `Zero Blind Agent Edits` |
| **Claude Code & Git Hook Enforcement** | 🟢 **Production (GA)** | `.claude/settings.json` + `.git/hooks/` | `PostToolUse` hook (`garuda talon guard --hook` exiting `2` on `BLOCK`) and Git `pre-commit` / `commit-msg` Merkle provenance trailers | `Pre-Commit & Edit-Time Gate` |
| **Autonomous Supervisor Loop (`garuda talon run`)** | 🟢 **Production (GA)** | `talon.RunHarnessLoop` | Wraps any CLI coding agent, intercepts dirty edits, evaluates `talon guard`, and loops the agent with self-repair prompts until `ALLOW` | `Automated Self-Repair Convergence` |
| **6-Surface Software Intelligence Instrument (`/dashboard` PWA)** | 🟢 **Production (GA)** | `cmd/garuda-api` & `garuda dev` | `Briefing`, `Architecture Map` (1-hop/2-hop), `Change Review Workbench`, `Policies`, `Contradictions & Evidence`, `Governance & Audit Ledger` + `⌘K` palette + `/manifest.json` & `/sw.js` PWA | `Ed25519 JWT & Session Isolation` |

---

## 6. Plane 06 — Universal Zero-Dependency Runtime & Cross-OS Diagnostics

| Capability | Status | Verification Tier | Supported Semantics | Enforced Invariant |
| :--- | :--- | :--- | :--- | :--- |
| **Windows PowerShell Zero-Dependency Installer (`install.ps1` / `setup.ps1`)** | 🟢 **Production (GA)** | `Windows 10/11 PowerShell` | Auto-provisions Portable PostgreSQL 16 (`$HOME\.garuda\runtime\postgres`) and Portable Go (`$HOME\.garuda\toolchain\go`) without Admin or GCC | `1-Command Windows Onboarding` |
| **macOS / Linux / Chromebook / Android Installer (`install.sh` / `setup.sh`)** | 🟢 **Production (GA)** | `POSIX / Darwin / Crostini / Termux` | Auto-provisions Portable Go, user-space PostgreSQL (`~/.garuda/data/pg`), and compiles in Pure-Go (`CGO_ENABLED=0`) when `gcc`/`clang` is absent | `1-Command Universal Onboarding` |
| **Self-Diagnosis & Auto-Repair CLI (`garuda doctor`)** | 🟢 **Production (GA)** | `cmd/garuda/doctor_cmd.go` (`--fix`, `--json`) | Verifies and repairs `.env`, PostgreSQL & migrations, all 4 binaries, Talon hooks, and `/dashboard` gateway health | `Self-Healing Local Runtime` |
