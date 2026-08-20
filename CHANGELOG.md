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