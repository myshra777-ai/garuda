## [Unreleased]

### Added
- **Truth Gate**: Completed Stage T1 semantic truth gate v0.1 with P0-09 and P0-10 rules.
- **Truth Gate**: Added pre-ledger validation, canonical identity enforcement, and deterministic snapshot hashing.
- **Types & Manifests**: Added analysis manifest schema, scripts, and build scaffolding.
- **Testing**: Added P0-08 evidence integrity and source-hash tamper detection test suites.

### Changed
- **Licensing**: Added SPDX Apache 2.0 license headers to all Go source files.
- **Documentation**: Overhauled README with AST benchmark suite details, truth fixtures, CLI reference, and auto-generated capabilities matrix and API specs.

### Fixed
- **Types**: Stabilized `internal/types` baseline and isolated semantic AST contracts.

### Semantic Changes
- Enforced canonical identity and deterministic snapshot hashing prior to ledger ingestion.
- Implemented Stage T1 semantic truth gate mechanisms (P0-08 evidence integrity, P0-09, P0-10) for source tamper detection and assertion validation.
Here is the CHANGELOG for Garuda based on the commits since release `v0.1.0`:

## [Unreleased]

### Added
- **Truth Gate Stage T1**: Completed Stage T1 semantic truth gate v0.1 supporting P0-09 and P0-10 rules (`4afce0b`).
- **Canonical Identity & Snapshot Hashing**: Enforced canonical identity resolution, deterministic snapshot hashing, and pre-ledger validation in the truth gate (`4120623`).
- **Analysis Manifest Schema**: Added analysis manifest schema, scripts, and build scaffolding (`307f64a`).
- **Integrity Testing**: Added P0-08 evidence integrity tests and source-hash tamper detection test suite (`ebcc60c`).

### Changed
- **Documentation**: Overhauled README with AST benchmark suite documentation, truth fixtures, CLI reference, auto-generated capabilities matrix, and API specs (`81068cf`, `72c4204`, `e3423dd`, `4375b9f`).
- **Licensing**: Added SPDX Apache 2.0 license headers across all Go source files (`409351c`).

### Fixed
- **Types Baseline**: Stabilized `internal/types` baseline and isolated semantic AST contracts (`ea609b6`).

### Semantic Changes
- Enforced canonical identity and pre-ledger validation routines within the Stage T1 semantic truth gate.
- Introduced deterministic snapshot hashing and source-hash tamper detection for evidence integrity verification.
- Isolated semantic AST contracts from internal baseline types to prevent model drift.
