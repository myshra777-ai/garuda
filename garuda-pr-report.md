## 🦅 Garuda Architectural & Financial Governance Gate

| Metric | Value |
|---|---|
| **Governance Outcome** | ✅ **ALLOW** |
| **Workspace / Target** | `go-validation-10` (`diff:diff (2 symbols: persistEvaluation, Migrate)`) |
| **Blast Radius** | `20` affected entities (`CRITICAL=0`, `HIGH=5`) |
| **Policies Evaluated** | `16` policies in `6968ms` (`Net-New=0`, `Baseline=36`) |
| **Merkle Provenance** | RFC 6962 Block `#537` |

### 🔧 Deterministic Self-Repair Instructions

- Policy archunit-imported-01 (ArchUnit Imported: domain must not depend on infrastructure): INCONCLUSIVE: policy scope matched 0 entities in workspace (vacuous ALLOW prevented for policy archunit-imported-01)
- Policy importlinter-contract-01 (Import-Linter Imported: app.domain -> app.adapters forbidden): INCONCLUSIVE: policy scope matched 0 entities in workspace (vacuous ALLOW prevented for policy importlinter-contract-01)
- Policy tpl-hexagonal-ports-adapters (Hexagonal Architecture: Domain core must not import infrastructure adapters): INCONCLUSIVE: policy scope matched 0 entities in workspace (vacuous ALLOW prevented for policy tpl-hexagonal-ports-adapters)
- Policy depcruise-no-ui-in-services-01 (Dependency-Cruiser Imported: no-ui-in-services): INCONCLUSIVE: policy scope matched 0 entities in workspace (vacuous ALLOW prevented for policy depcruise-no-ui-in-services-01)

**Reproduce Locally:**
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 garuda check --target linux/amd64 --cgo=true --commit 68fffe7f7760 --worktree-dirty --dirty-tree-hash d8a3173825b9 --workspace go-validation-10
```
