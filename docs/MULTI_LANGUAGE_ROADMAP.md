# Garuda Multi-Language Roadmap

**Status:** Active roadmap
**Last updated:** 2026-09-12
**Current coverage:** Go (Stable, type-resolved), Python (Beta, import-resolved), TypeScript (Beta, import-resolved)

## The Core Problem

Every code intelligence product can parse a language's syntax. Parsing is a solved problem — tree-sitter grammars exist for 100+ languages. What separates a code graph from a **verifiable** code graph is resolution: for every edge, can we point to the exact declaration the edge refers to, and can we prove it?

Garuda's differentiation is not breadth. It is verification. A repository with 10,000 HEURISTIC edges is not worth more than a repository with 1,000 RESOLVED edges. It is worth less, because the noise hides the signal.

## The Verification Tiers

Every language Garuda supports progresses through five tiers of resolution. Each tier is a distinct capability. A language is only "stable" when it reaches Tier 4 on real repositories.

| Tier | Name | Method | Proof |
|---|---|---|---|
| 1 | Structural | AST parse only | Source line exists |
| 2 | Import | Import graph resolved | Target file exists |
| 3 | Scope | Local symbol resolution | Target declaration exists in same file/class |
| 4 | Type | Compiler or LSP resolution | Target declaration exists, matches call signature |
| 5 | Cross-repo | Workspace-wide type resolution | Target declaration is in another repo, resolved through build system |

**Current state:**

| Language | Tier | Method constant | Notes |
|---|---|---|---|
| Go | 5 | `GO_TYPES` / `IMPORT_RESOLUTION` / `AST_EXACT` | `go/types` provides tiers 4–5 natively |
| Python | 2 | `PYTHON_IMPORT` for imports, `HEURISTIC` for calls | Tier 3+ requires jedi/pyright |
| TypeScript | 2 | `TS_IMPORT` for imports, `HEURISTIC` for calls | Tier 3+ requires tsserver |
| All others | — | — | Not yet implemented |

## The Language Ladder

Ten languages cover roughly 95% of enterprise software. Priority is by (a) enterprise usage, (b) quality of existing tooling, (c) coverage gap in adjacent products.

### Tier A — Must have (target: within 6 months)

| # | Language | Parser | Resolver | Target tier | Effort |
|---|---|---|---|---|---|
| 1 | **Go** | `go/parser` | `go/types` | 5 | ✅ Done |
| 2 | **Python** | Custom AST scanner | `jedi` or `pyright` | 4 | 2 weeks |
| 3 | **TypeScript** | `tree-sitter-typescript` | `tsserver` or `ts-morph` | 4 | 2 weeks |
| 4 | **Java** | `JavaParser` | `Eclipse JDT` LSP or `javac` API | 4 | 3 weeks |
| 5 | **C#** | `tree-sitter-c-sharp` | `Roslyn` (via LSP) | 4 | 3 weeks |

### Tier B — Should have (target: within 12 months)

| # | Language | Parser | Resolver | Target tier | Effort |
|---|---|---|---|---|---|
| 6 | **Rust** | `tree-sitter-rust` | `rust-analyzer` (LSP) | 4 | 3 weeks |
| 7 | **C / C++** | `tree-sitter-c` / `tree-sitter-cpp` | `clangd` (LSP) | 3 | 4 weeks |
| 8 | **Kotlin** | `tree-sitter-kotlin` | Kotlin compiler analysis API | 4 | 3 weeks |
| 9 | **Ruby** | `tree-sitter-ruby` | `solargraph` (LSP) or Ripper | 3 | 3 weeks |
| 10 | **PHP** | `tree-sitter-php` | `php-parser` + `Psalm` | 3 | 3 weeks |

### Tier C — Nice to have (target: on demand)

| # | Language | Effort | When |
|---|---|---|---|
| 11 | Swift | 4 weeks | When demanded by customer |
| 12 | Scala | 4 weeks | When demanded |
| 13 | Elixir | 3 weeks | When demanded |
| 14 | Dart | 3 weeks | When demanded |

## The Universal Semantics Contract

Every language analyzer must produce the same `RepositorySnapshot` shape:

```go
type RepositorySnapshot struct {
    Entities      []Entity
    Relationships []Relationship
    Stats         SnapshotStats
}

Language-specific logic lives entirely inside the analyzer. The semantic core (identity, claims, evidence, verification, Merkle) is language-agnostic.

Rules every analyzer must satisfy:

Every entity carries a Language field. No null, no empty. The dashboard groups by this.

Every relationship carries a ResolutionMethod. HEURISTIC is honest; GO_TYPES is Go-only. A Python edge must never claim a Go resolution method.

Every relationship carries Evidence with a real file and line. No exceptions.

Entity kinds follow a shared vocabulary. KindClass is class in every language that has classes. KindFunction is function in every language that has functions. Language-specific kinds (e.g., Rust's trait) map to the nearest shared kind and preserve the native name in an attribute.

Detectors must be conservative. requirements.txt alone does not make a repo Python. package.json alone does not make it TypeScript. Require primary evidence (pyproject.toml for Python, tsconfig.json for TS, go.mod for Go).

The Multi-Language Pipeline
text
                        ┌─────────────────┐
                        │  Workspace root  │
                        └────────┬────────┘
                                 │
                    ┌────────────▼────────────┐
                    │  Language detection     │
                    │  Go > TS > Python > ... │
                    └────────────┬────────────┘
                                 │
              ┌──────────────────┼──────────────────┐
              ▼                  ▼                  ▼
        ┌──────────┐      ┌──────────┐      ┌──────────┐
        │ Go lang  │      │ Python   │      │ TS lang  │
        │ analyzer │      │ analyzer │      │ analyzer │
        └────┬─────┘      └────┬─────┘      └────┬─────┘
             │                 │                 │
             └────────────┬────┴─────────────────┘
                          ▼
                 ┌────────────────┐
                 │ Repository     │
                 │ Snapshot       │
                 └────────┬───────┘
                          │
             ┌────────────┼────────────┐
             ▼            ▼            ▼
        ┌────────┐  ┌──────────┐  ┌──────────┐
        │ Identity│  │ Claims  │  │ Evidence │
        │         │  │ Storage │  │ Storage  │
        └────────┘  └──────────┘  └──────────┘
Path B — The Real Resolver Initiative
Path A (import resolution) is committed. It takes Python and TypeScript to Tier 2. The next milestone is Tier 4 for both.

Python resolver — 2 weeks
Tooling: jedi (pure Python, embedded as subprocess).

Why jedi:

Pure Python, no compiled extensions, easy to install and embed

Handles incomplete code gracefully — returns partial results rather than failing

Significantly faster than pyright on large repos

Well-tested, used by IPython, VS Code Python extension, and many IDEs

Architecture:

text
Go analyzer ─── JSON-RPC over stdio ────┐
                                        ▼
                                  ┌──────────┐
                                  │  jedi    │
                                  │ helper   │
                                  └──────────┘
Per file:

Analyzer emits all identifier positions of interest

Helper resolves each position to (file, line, symbol)

Analyzer maps back to AST nodes and produces edges

Expected precision: 85–95% on clean Python codebases. Degrades on dynamically-typed-heavy code.

TypeScript resolver — 2 weeks
Tooling: tsserver via subprocess, or ts-morph as a Node helper.

Why tsserver:

Same engine VS Code uses

Handles TS project references, paths, and incremental updates

Returns fully-resolved symbol identities

Architecture: Same RPC shape as Python. Go spawns Node, sends file/position pairs, receives resolutions.

Expected precision: 90–98% on typed TypeScript. Lower on JS-heavy codebases.

Java resolver — 3 weeks
Tooling: Eclipse JDT Language Server (eclipse.jdt.ls).

Why JDT:

Mature, supports full Java type resolution

Speaks LSP, which is a well-defined protocol

Handles Maven and Gradle projects

Architecture: LSP client in Go. Send textDocument/definition for every identifier. Batch.

Expected precision: 95%+ on JVM codebases with resolvable dependencies.

C# resolver — 3 weeks
Tooling: Roslyn via OmniSharp or the official Microsoft LSP server.

Architecture: LSP client. Same shape as Java.

Expected precision: 95%+.

Benchmark Requirements Per Tier
Each language must pass a per-language benchmark before being declared "stable."

Tier	Benchmark	Passing threshold
1	Parse 1000 files, no crashes, count entities	100% of files parsed
2	Resolve every import in a fixture repo	≥95% import resolution rate
3	Resolve 100 labeled method calls in fixture repo	≥90% precision, ≥85% recall
4	Resolve 500 labeled calls against real repo	≥95% precision, ≥90% recall
5	Cross-repo call resolution across 3 repos	≥90% precision, ≥85% recall
No language claim is published without a passing benchmark.

The Honest Publishing Rule
When you publish multi-language support, publish with the honest breakdown:

Garuda analyzes Go at Tier 5 (compiler-resolved, 99%+ precision).
Python and TypeScript are at Tier 2 (import-resolved, calls marked heuristic).
Java and C# at Tier 4 are on the roadmap. Additional languages will
follow the same pattern: parse, resolve imports, resolve types. No
language claim will be published until it passes its tier benchmark.

This is what an engineer reading your README wants to see. It is also what an investor's technical advisor will check first.

What Changes At Each Tier
Tier	Dashboard shows	Policy engine can	Compliance evidence
1	Entity counts, no relationships	Run structural predicates	None
2	Import graph	Run import-based policies	Import-based
3	Local reference graph	Run scope-based policies	Symbol-level
4	Full call graph	Run architecture policies	Call-level
5	Cross-repo call graph	Run cross-service policies	Full service graph
Current: Go is at Tier 5. Python and TypeScript at Tier 2 after Path A.

Timeline
Quarter	Milestone
Q4 2026	Path A commit (import resolution) for Python + TypeScript
Q4 2026	Python Tier 4 (jedi integration)
Q4 2026	TypeScript Tier 4 (tsserver integration)
Q1 2027	Java Tier 4 (JDT integration)
Q1 2027	C# Tier 4 (Roslyn integration)
Q1 2027	Java + C# benchmarks published
Q2 2027	Rust Tier 4 (rust-analyzer)
Q2 2027	C/C++ Tier 3 (clangd)
Q2 2027	Kotlin, Ruby, PHP on demand
Appendix — Language-Specific Notes
Go
Fully resolved via the standard library's go/types. No external tooling. This is why Go was first.

Python
Import resolution works today. Type resolution requires jedi or pyright. Prefer jedi for speed, pyright for accuracy on typed codebases. Can support both with a selector.

TypeScript / JavaScript
tree-sitter parses both. tsserver resolves both, with lower accuracy on untyped JS. Consider a separate "JavaScript (untyped)" status tier.

Java / Kotlin / Scala
Share the JVM toolchain. JDT handles Java directly. Kotlin requires the Kotlin compiler's analysis API. Scala requires scalameta. Consider these a package.

C / C++
clangd handles both. Build-system awareness (CMake, Bazel) is required to resolve includes and macros correctly. This is the highest-effort language family.

Rust
rust-analyzer is mature. Handles Cargo projects natively. Effort is mostly in the LSP client, not the resolution.

C#
Roslyn is the reference implementation of the C# compiler. Full resolution via LSP. Requires .NET runtime.

Governance
Changes to this roadmap require:

A benchmark result supporting any "Tier N" claim.

A note in the CHANGELOG when a language moves up a tier.

An explicit retraction if a language's benchmark regresses.

The roadmap is the source of truth for what languages Garuda actually supports at what level of verification. Marketing copy derives from the roadmap; the roadmap never derives from marketing copy.

text

---

## Current Measured Coverage

| Language | Total claims | Import-resolved | Heuristic | Resolved % |
|---|---|---|---|---|
| Go | ~3,115 | 3,115 | 0 | 100% |
| Python | 803 | 131 | 672 | 16.3% |
| TypeScript | ~2,036 | TBD after fix | TBD | TBD |

Path A (import resolution only) delivers 15–20% resolution on real
repositories. The remaining 80% requires type resolution — Path B
(jedi / tsserver integration). This matches the roadmap's earlier
prediction that Tier 2 is a lower bound, not the destination.

