# Changelog

All notable changes to Garuda since `v0.1.0` will be documented in this file.

## [Unreleased]

### Added
- **Stage T1 Semantic Truth Gate (v0.1)**: Completed initial Stage T1 truth gate implementation incorporating P0-09 and P0-10 rules (`4afce0b`).
- **Truth Gate Validation**: Added canonical identity enforcement, deterministic snapshot hashing, and pre-ledger validation checks (`4120623`).
- **Analysis Manifest & Build Scaffolding**: Introduced the analysis manifest schema, supporting build scripts, and initial scaffolding (`307f64a`).
- **Integrity Testing**: Added P0-08 evidence integrity tests and source-hash tamper detection test suite (`ebcc60c`).

### Changed
- Updated README and documentation suite to include AST benchmark suite, truth fixtures, CLI reference, automatically generated capabilities matrix, and API specifications (`81068cf`, `72c4204`, `4375b9f`).

### Fixed
- Stabilized `internal/types` baseline and isolated semantic AST contract definitions (`ea609b6`).

### Semantic Changes
- **Pre-Ledger Identity & Hashing**: Enforced canonical identity validation and deterministic snapshot hashing prior to ledger entry (`4120623`).
- **Evidence Integrity Verification**: Established evidence integrity requirements and source-hash tamper detection baseline (`ebcc60c`).