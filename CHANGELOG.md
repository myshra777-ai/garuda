## [Unreleased]

### Added
- **Canonical Entity Engine & V5 Truth Benchmark**: Added V5 truth benchmark suite, canonical entity engine, and complete Apache 2.0 compliance tooling (`cc9d745`).
- **Semantic Graph & Impact Analysis**: Implemented semantic graph capabilities, impact analysis, and Company Brain MVP integration (`0054002`, `4b40135`).
- **Graph Explorers & Visualizations**: Added progressive disclosure explorer, company graph visualizer, and interactive mind map (`9551b20`).
- **CLI Commands**: Introduced interactive commands for graph, policy, topology, and handoff (`56c7f41`).
- **Documentation Utilities**: Added a capabilities matrix generator and updated license reference links (`fd6e91b`).

### Changed
- Standardized documentation generation scripts on `gemini-3.6-flash` (`99087b7`).
- Revised and restructured the README for improved clarity (`7435ac8`, `ce46ecb`).

### Fixed
- **Database Migrations & Store**:
  - Hardened migrations 040 and 041 (`99087b7`).
  - Ensured line column exists before backfill in migration 035 (`26d6d02`).
  - Executed migration files in atomic transactions without relying on semicolon splitting (`6221f55`).
  - Added `IF NOT EXISTS` guards to `idx_api_contracts_tenant` and related indexes in migration 033 (`18abfbe`).
  - Verified `repositories` and `entities` tables exist prior to executing cross-repository migrations (`ae6156c`).
- **Analyzer & Telemetry**: Resolved extractor gaps for fingerprints, calls, and imports, and fixed a Server-Sent Events (SSE) data race (`11dd33d`).
- **CI & Dependencies**:
  - Removed `-mod=vendor` flag from CI workflow (`04b3e83`).
  - Explicitly added `google-genai` to `requirements.txt` and updated `.gitignore` rules for script tracking (`f3eb3e8`, `0e7b87d`).
  - Resolved missing `gopkg.in/yaml.v3` dependency (`b6f32a7`, `a2aba4f`).

### Removed
- Removed the `vendor/` directory in favor of standard Go module caching (`b506889`).

### Semantic Changes
- **Canonical Entity Model**: Introduced canonical entity engine powering repository entities, claims, evidence, and lineage tracking.
- **Graph & Impact Analysis**: Added semantic graph capabilities providing impact analysis and cross-entity relationships across Go packages.