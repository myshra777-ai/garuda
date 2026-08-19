# CHANGELOG

## [Unreleased]

### Added
- **Canonical Semantic & Entity Engine**: Introduced core canonical semantic engine and entity engine supporting claims, evidence, and lineage.
- **Graph Visualization & Exploration**: Added progressive disclosure graph explorer, company graph, and interactive mind map rendering.
- **CLI Tooling**: Added new CLI commands for interactive graph exploration, policy governance, topology inspection, and handoff workflows.
- **Impact Analysis & Intelligence**: Integrated impact analysis capabilities and Company Brain MVP functionality.
- **Truth Benchmark Suite**: Implemented V5 truth benchmark suite alongside full Apache 2.0 license compliance updates.
- **Automated Docgen**: Introduced automated capability matrix and API specification generators integrated into CI/documentation workflows.

### Changed
- **MVP Refactor**: Refactored core engine to standardize canonical semantic models and impact analysis integrations.
- **Documentation Workflows**: Standardized documentation generation scripts to utilize `gemini-3.6-flash`.
- **Documentation Updates**: Updated project README and added capabilities matrix linking for improved project structure and clarity.

### Fixed
- **Database & Migration Stability**:
  - Executed database migrations inside atomic transactions to eliminate issues caused by manual semicolon splitting.
  - Hardened database migrations `040-041`.
  - Added existence check for `line` column prior to running backfill logic in migration `035`.
  - Added `IF NOT EXISTS` clauses for `idx_api_contracts_tenant` and related indexes in migration `033`.
  - Ensured `repositories` and `entities` tables exist before executing cross-repo migrations.
- **Analyzer & Telemetry**:
  - Resolved gaps in fingerprint, function call, and import extraction logic.
  - Fixed Server-Sent Events (SSE) data race in telemetry handling.
- **CI & Build Pipeline**:
  - Removed `-mod=vendor` flag requirement from CI workflow steps.
  - Tracked and explicitly whitelisted `scripts/requirements.txt` in `.gitignore`.
  - Added explicit dependency for `google-genai` to `requirements.txt`.

### Removed
- Removed project `vendor/` directory in favor of Go module cache management.

### Semantic Changes
- **Canonical Entity Engine**: Transitioned semantic models toward canonical entity schema, stabilizing claims, evidence, and cross-repository lineage tracking.