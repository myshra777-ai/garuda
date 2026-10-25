# Multi-language coverage and roadmap

**Current reference:** Garuda v1.5.0, 2026-10-25.  
**Purpose:** Describe what the current analyzers resolve, what remains structural or heuristic, and what evidence is needed before stronger language claims are made.

Garuda should not treat “the parser supports a language” as equivalent to “every call in that language is resolved.” The useful unit is the relationship type, resolution method and test evidence for the construct being analyzed.

## Current implementation

| Language | Current approach | Supported coverage | Boundary |
|---|---|---|---|
| **Go** | Go AST plus compiler/type information, workspace/module loading and `go.work` support | Supported declarations, calls, receivers, interfaces, embedding and tested cross-module relationships | Resolution applies to packages and build configurations the toolchain can load. Dynamic behavior and unsupported/generated configurations can be absent. |
| **Python** | Structural AST analysis with selected cross-file import-alias and instance-call patterns | Modules, classes, functions, imports, selected aliases and supported instance/method calls | Not a full type checker or runtime resolver. Dynamic dispatch, reflection, monkey-patching and unsupported syntax may remain unresolved. |
| **TypeScript** | Tree-sitter parsing with selected cross-file alias, namespace-import, instance and method-call patterns | Supported classes, interfaces, imports, declarations and selected call relationships | Not a full TypeScript compiler or language-service resolution pass. Project-reference and dynamic behavior may not be represented. |
| Other languages | No current general analyzer contract | Not currently supported by the product contract | A roadmap entry is not an implemented analyzer. |

The current implementation and limits are summarized in [CAPABILITIES.md](CAPABILITIES.md) and the canonical [capability metadata](source/capabilities.yaml). The latest measured runs belong in [EVIDENCE.md](../EVIDENCE.md).

## Confidence is not the same as measured precision

Some supported Python and TypeScript cross-file alias/instance paths assign a static confidence value of `0.95`. That value is an implementation confidence marker; it must not be described as an experimentally measured 95% precision rate unless an independently labeled ground-truth corpus demonstrates that rate.

Similarly, compiler/type-aware Go resolution gives stronger evidence for the supported constructs, but passing the strict benchmark suite does not prove every possible relationship in every Go repository is present.

## Resolution tiers

Use these tiers to describe a particular relationship class—not as a blanket grade for an entire language.

| Tier | Meaning | Evidence expected |
|---|---|---|
| 1 — Structural | Syntax parsed and entities extracted | Parser coverage, stable source locations and deterministic extraction |
| 2 — Import | Import/module target resolved | Labeled import fixtures and unresolved-import reporting |
| 3 — Local/supported call pattern | Selected identifiers, aliases or instance calls map to declarations under documented rules | Labeled fixture cases, with heuristic paths distinguished |
| 4 — Compiler/type-aware | Compiler or language-service resolution identifies the declaration for supported constructs | Ground-truth precision/recall on representative repositories |
| 5 — Workspace/cross-repository | Resolution crosses configured module/repository boundaries using the language's build model | Multi-module tests and labeled cross-repository edges |

Go has compiler/type-aware paths and cross-module cases represented in the strict suite. Python and TypeScript currently provide selected structural and cross-file resolution, not universal Tier 4 resolution. Do not infer that all their edges have equal confidence.

## Current validation evidence

The current repository evidence reports:

- `make benchmark-strict`: 24/24 strict compiler/AST benchmark cases passed.
- A Go-only validation corpus covering 14 repositories, with counts documented separately in [EVIDENCE.md](../EVIDENCE.md).
- A separate blended workspace of 9 repositories across Go, Python and TypeScript, with 14,333 entities and 40,956 relationships recorded for that snapshot.
- Cross-file alias/instance tests for supported Python and TypeScript patterns.
- `scripts/mcp_verify.py`: 30/30 reference MCP/Talon verifier checks.
- `scripts/runtime_contradiction_test.py`: 10/10 controlled injected contradictions detected.

These results do not establish real-repository precision/recall for all languages or complete runtime discovery. The mixed-workspace counts and Go-only corpus are different datasets.

## What would justify stronger language claims?

For every new resolver or relationship class:

1. Define the source syntax and resolution contract.
2. Record known-unresolved cases and whether edges are compiler-resolved, structurally resolved or heuristic.
3. Build labeled fixtures and representative repository cases.
4. Report precision, recall, unresolved rate and false-positive examples where measurable.
5. Test deterministic output across repeated runs.
6. Publish the command, commit, environment, corpus and timestamp with the result.
7. Downgrade the claim if a benchmark regresses; do not allow marketing wording to outrun evidence.

## Roadmap

The next multi-language work is to expand labeled ground-truth evaluation and improve resolution quality for supported Python and TypeScript patterns. Additional language analyzers should be prioritized only when there is a concrete user need and a credible verification plan. Java, C#, Rust, C/C++, Kotlin, Ruby, PHP and other languages are not presented here as current analyzer capabilities.

## Historical results

Earlier milestones remain in [EVIDENCE_MULTI_LANG.md](../EVIDENCE_MULTI_LANG.md). That file records the September 2026 state of the analyzer and should be read as a historical snapshot, not as the current Python/TypeScript capability definition.
