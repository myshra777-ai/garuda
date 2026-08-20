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