## [Unreleased]

### Added
- **Canonical Entity Engine & Benchmarking**: Integrated canonical entity engine, V5 truth benchmark suite, and full Apache 2.0 license compliance (`cc9d745`, `fd6e91b`).
- **Graph & Visualization Features**: Added progressive disclosure explorer, company graph, interactive mind map, and impact analysis features (`9551b20`, `0054002`, `4b40135`).
- **CLI Commands**: Introduced CLI commands for interactive graph, policy, topology, and handoff management (`56c7f41`).
- **Documentation & Matrix Generation**: Added a capabilities matrix generator, linked capabilities matrix in the README, and explicitly declared `google-genai` requirement (`fd6e91b`, `0e7b87d`).

### Changed
- **Dependency Management**: Transitioned from vendored dependencies to standard Go module cache (`b506889`, `1b8aa4b`).
- **Documentation**: Revised README for improved structure and clarity (`7435ac8`, `ce46ecb`).

### Fixed
- **Database Migrations**: 
  - Ensured line column exists prior to backfill in migration 035 (`26d6d02`).
  - Wrapped store migration files in atomic transactions without semicolon splitting (`6221f55`).
  - Added `IF NOT EXISTS` guards to `idx_api_contracts_tenant` and related indexes in migration 033 (`18abfbe`).
  - Verified `repositories` and `entities` tables exist prior to cross-repository migrations (`ae6156c`).
- **Analyzer & Telemetry**: Resolved extractor gaps in fingerprint, call, and import extractions, and fixed a Server-Sent Events (SSE) data race (`11dd33d`).
- **CI & Build Tooling**: 
  - Tracked `scripts/requirements.txt` and updated `.gitignore` whitelisting (`f3eb3e8`).
  - Removed `-mod=vendor` flag from CI workflows (`04b3e83`).

### Removed
- Removed the `vendor/` directory in favor of standard Go module caching (`b506889`).

### Semantic Changes
- **Canonical Entity Engine**: Standardized entity processing via a unified canonical engine (`cc9d745`).
- **Semantic Graph Expansion**: Enhanced semantic graph domain model with impact analysis capabilities, progressive disclosure, and topology management (`0054002`, `9551b20`, `56c7f41`).
- **Extraction Accuracy**: Fixed AST/code extraction gaps across fingerprinting, call graphs, and import graphs (`11dd33d`).