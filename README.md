<p align="center">
  <img src="assets/garuda-logo.png" width="88" alt="Garuda">
</p>

<h1 align="center">Garuda</h1>

<p align="center">
  <strong>Architecture-aware change control for AI-written code.</strong><br>
  Help coding agents work within the constraints your team intends to keep.
</p>

<p align="center">
  <a href="https://github.com/myshra777-ai/garuda/actions/workflows/ci.yml"><img src="https://github.com/myshra777-ai/garuda/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache--2.0-blue.svg" alt="Apache 2.0 License"></a>
</p>

AI coding agents can produce code that compiles and passes tests while still breaking an architectural rule: a handler reaches directly into a database layer, a dependency cycle appears, or a change to a shared symbol affects far more callers than expected.

**Garuda builds a semantic model of a codebase and checks changes against explicit architecture policies and documented constraints.** It brings those checks into a developer's CLI, AI-agent workflow and CI process, with evidence available to inspect and governance decisions that can be recorded and cryptographically verified.

## See the workflow

1. **Map the code.** Index supported Go, Python and TypeScript structures and relationships.
2. **Understand the change.** Use a Git diff to identify touched symbols and estimate their dependency impact.
3. **Check the rules.** Evaluate supported YAML policies and ingested architecture claims against the indexed model.
4. **Review the result.** Inspect `ALLOW`, `WARN`, `REVIEW` or `BLOCK` outcomes, including the distinction between new violations and existing baseline issues where supported.
5. **Keep evidence.** Persist selected governance decisions and verify their Merkle inclusion proofs.

Garuda is designed to make a defined class of architectural problems easier to detect. It does not prove that every behavior in a program has been discovered or that a passing check means a change is bug-free.

## Try Garuda

The repository includes a setup script for a local stack. Review the script before running it, especially on a machine with existing Docker containers or services.

```bash
git clone https://github.com/myshra777-ai/garuda.git
cd garuda
less setup.sh
./setup.sh
```

After setup, initialize or index the repository you want to analyze using the CLI commands supported by your checkout. For example:

```bash
./bin/garuda analyze /path/to/repository --save
./bin/garuda check --diff HEAD --project my-system --policy ./policies --save --fail-on-block
```

The exact flags and required workspace/database configuration can vary by command. Use the [CLI and API reference](docs/API.md) and [quick start](docs/QUICKSTART.md) for setup and command details.

## Garuda Talon

**Talon connects Garuda's checks to AI-assisted development.** Its implementation includes workspace contract generation, pre-edit context, post-edit diff checks, agent-facing MCP tools and a supervisor loop that can feed repair instructions back to a CLI coding agent.

Talon can help guide or gate changes at configured integration points. The degree of enforcement depends on the installed hook, client configuration, command, policy scope and failure behavior. An agent instruction or MCP tool alone is not a security boundary, so teams should validate their own workflow before relying on it as a mandatory gate.

## What is included

- **Semantic analysis:** Go analysis uses compiler/type information for supported resolution paths; Python and TypeScript analysis use AST-based techniques with language-specific coverage limits.
- **Impact analysis:** Explore relationships and estimate the downstream or upstream impact of a changed symbol.
- **Architecture policies:** Evaluate supported declarative rules against the indexed model.
- **Documentation claims:** Ingest supported Markdown/ADR content and compare extracted architectural claims with modeled code relationships.
- **Runtime evidence:** Ingest supported OpenTelemetry span data and correlate observations with static entities. This is evidence from observed traces, not complete runtime discovery.
- **Agent integrations:** Expose tools and context through MCP, with Talon hooks and generated configuration for supported workflows.
- **Governance records:** Persist selected decisions and verify their cryptographic integrity using the Merkle implementation.
- **Web dashboard:** Inspect workspace status, architecture neighborhoods, policy outcomes, contradictions and recorded decisions.

See the [capability matrix](docs/CAPABILITIES.md) for implementation scope, the [walkthrough](docs/WALKTHROUGH.md) for examples, and [EVIDENCE.md](EVIDENCE.md) for reported validation results and their limits.

## Evidence, not false certainty

Garuda uses three evidence states:

- `SUPPORTED` — available evidence supports a claim within the checked scope.
- `UNVERIFIED` — evidence is insufficient; this does not mean the claim is false.
- `CONTRADICTED` — available evidence conflicts with a claim.

A Merkle inclusion proof can show that a persisted record is consistent with a supplied root and proof path. It does not establish that the original analysis was complete, that a policy was well designed, or that the software is correct.

Benchmark results describe the specific tests and corpora used. They should not be read as universal accuracy guarantees.

## Documentation

| Start here | What you'll find |
|---|---|
| [Quick start](docs/QUICKSTART.md) | Local setup, first analysis and first change check |
| [Capabilities](docs/CAPABILITIES.md) | Implemented surfaces and known limits |
| [CLI, MCP and HTTP API](docs/API.md) | Commands, configuration, tools and API reference |
| [Specifications](docs/SPECS.md) | System architecture, data model, interfaces and boundaries |
| [Walkthrough](docs/WALKTHROUGH.md) | End-to-end examples and dashboard tour |
| [Evidence and verification](EVIDENCE.md) | Benchmarks, validation corpora and methodology notes |
| [Operational playbook](PLAYBOOK.md) | Running and operating Garuda |
| [Positioning](docs/POSITIONING.md) | Internal product language and claim discipline |
| [Security policy](SECURITY.md) | Responsible vulnerability reporting |

## Current boundaries

- Results are limited by the repositories, files, language constructs and relationships Garuda successfully indexes.
- Static analysis cannot discover every dynamic behavior, reflection path, generated-code edge or runtime configuration.
- Runtime correlation only reflects telemetry actually ingested and successfully mapped.
- Policy outcomes depend on the configured rules and the evidence available to the evaluator.
- Agent hooks and generated instructions have client-specific enforcement limits; validate bypass and failure behavior before treating them as mandatory controls.
- Garuda is not a coding model, a replacement for tests or code review, or a compliance certification.

## Contributing

Contributions and reproducible bug reports are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md). Please keep documentation claims aligned with tests and measured evidence.

---

Licensed under [Apache 2.0](LICENSE).
