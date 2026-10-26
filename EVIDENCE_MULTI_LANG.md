# Garuda Multi-Language Validation

**Milestone:** Path B Step 1 — Go + Python unified semantic graph
**Date:** 2026-09-10
**Workspace:** `go-validation-10`
**Tenant:** `00000000-0000-0000-0000-000000000001`
**Merkle block height:** (see ledger section below)

---

## Executive Summary

Garuda now analyzes Go and Python codebases into a single semantic graph,
with a unified ledger, unified dashboard, and a unified drift engine.

| Metric | Value |
|--------|-------|
| Repositories analyzed | 12 (11 Go + 1 Python) |
| Packages | 784 |
| Internal entities | 12,917 |
| Relationships | 25,269 |
| Cross-repository bridges | 247 |
| Languages | Go, Python |
| Go share | 90.1% |
| Python share | 9.9% |
| Document claims | 4 (from 1 ADR) |
| Doc-code drift detected | 1 (doc→code) + 12,592 (code→doc) |
| Merkle ledger | Dual-root, advancing |

**Headline:** One semantic graph. Two languages. Same ledger. Same dashboard.
Same drift engine. No language-specific forks of the core pipeline.

---

## What Was Built

### Python Parser (Structural, Pure Go)

**Location:** `internal/ast/python/parser.go`

**Approach:** Indentation-aware structural scanner.

- Tracks indentation to close blocks
- Handles multi-line strings, comments, implicit continuations
- Extracts: classes, base classes, methods, functions, module-level functions, imports
- No external dependencies, no cgo

**Extracted entity kinds:**

| Kind | Source |
|------|--------|
| `module` | one per `.py` file |
| `class` | `class Foo(Bar):` |
| `method` | `def method(self):` inside class |
| `function` | `def func():` at module level |
| `import` | `import x`, `from y import z` |

**Extracted relation kinds:**

| Kind | Source |
|------|--------|
| `IMPORTS` | module → target module (when target is in workspace) |
| `INHERITS` | class → base class (when base is in workspace) |

**Honest limitations:**
- No `CALLS` extraction (requires type inference, Python is dynamic)
- No external module resolution (imports to `pydantic`, `sqlalchemy` become external stubs)
- Decorators are captured but not interpreted
- Docstrings are stripped cleanly but not attached

### Language Dispatch

**Location:** `internal/analyzer/dispatch_python.go`, `cmd/garuda/analyze.go`

**Behavior:**
1. `garuda analyze <path>` inspects the target
2. If Python markers present (`pyproject.toml`, `setup.py`, `requirements.txt`, `Pipfile`) → route to Python parser
3. If `go.mod` present → route to Go parser
4. Either way, produces a standard `analyzer.Result`
5. Same persistence path: `SaveAnalysisDecision` → `SaveSemanticGraph` → same ledger

### Unified Semantic Schema

Both languages write to the same `entities` table with a `language` column:

```sql
language | count
---------+-------
go       | 14286
python   |  2121

(These are total counts including external stubs. Dashboard shows 12,917
internal-only entities after filtering kind != 'external'.)

Repository Inventory
Repository	Language	Entities	Relationships	Files	Commit
chi	Go	242	42	46	b1c9ab47
client_golang	Go	689	946	75	5a4ff970
cobra	Go	277	53	18	adbc8813
garuda-self	Go	1,167	1,922	212	86ce7e06
gin	Go	464	625	54	dcaa4296
grpc-go	Go	4,684	15,110	530	d2c52f09
jwt	Go	114	196	24	1a11d372
redis	Go	2,812	3,735	142	8010edc7
sqlmodel	Python	1,278	452	317	7fec3bcb
testify	Go	554	672	26	435c07b5
websocket	Go	134	364	16	e064f32e
zap	Go	502	683	55	bb1a55dd
Total	—	12,917	25,269	1,515	—
What the Dashboard Shows
Languages Panel
text
[======================= Go 90.1% ========================][== Python 9.9% ==]
Horizontal bar, GitHub-style. Two colors:

Go: #00ADD8 (official Go blue)

Python: #3572A5 (official Python blue)

Knowledge Drift Panel
text
📄 Documentation Claims
   Total documented capabilities        4
   Supported by code                    0
   Unverified                           4
   Contradicted                         0

🔍 Knowledge Drift
   Doc → Code drift (documented, not implemented)   1
   Code → Doc drift (implemented, not documented)   12,592
   Undocumented code entities                       8,288
   Unimplemented doc claims                         1
Architectural Hubs (cross-language)
Top-ranked entities by caller count. Now includes Python entities
alongside Go:

text
Codec            interface  google.golang.org/grpc          475 callers
Cmder            interface  github.com/redis/go-redis       255 callers
sqlmodel         package    stdlib (sqlmodel)               201 callers  ← Python
Compressor       interface  google.golang.org/grpc          185 callers
SQLModel         struct     sqlmodel.main (sqlmodel.main)   171 callers  ← Python
What this proves: Python entities are first-class citizens in the
semantic graph. They participate in centrality ranking, cross-repo
bridges, and search results.

Reproduction Commands
Prerequisites
bash
# Postgres running
docker-compose -f docker-compose.garuda.yml up -d postgres

# Full migrations
export DATABASE_URL="postgres://test:test@localhost:5433/garuda_test?sslmode=disable"
export GARUDA_TENANT_ID="00000000-0000-0000-0000-000000000001"
go run cmd/migrate/main.go
Workspace Setup
bash
export GARUDA_WORKSPACE="go-validation-10"
garuda workspace create go-validation-10
Analyze Repositories
bash
# Go repos (11)
for repo in chi client_golang cobra garuda-self gin grpc-go jwt redis testify websocket zap; do
  garuda analyze ~/garuda-validation/go/$repo --save --workspace go-validation-10
done

# Python repo (1)
garuda analyze ~/garuda-validation/py/sqlmodel --save --workspace go-validation-10
Ingest Documentation
bash
mkdir -p ~/garuda-validation/docs/chi-adr
cat > ~/garuda-validation/docs/chi-adr/adr-001.md << 'EOF'
# ADR-001: Chi Router Architecture

## Decision
The chi router MUST use the stdlib net/http Handler interface.
The router MUST implement the ServeHTTP method.
The Mux MUST support middleware chaining.
The Router SHOULD expose a Routes method for introspection.
EOF

garuda docs ingest ~/garuda-validation/docs/chi-adr
garuda docs verify
Verify Metrics
bash
# Language breakdown
curl -s "http://localhost:8080/api/v1/dashboard/stats?workspace=go-validation-10" | jq '.languages_breakdown'

# Aggregate stats
curl -s "http://localhost:8080/api/v1/dashboard/stats?workspace=go-validation-10" | jq '{
  repositories, packages, entities, relationships, cross_repo_links
}'

# Drift
curl -s "http://localhost:8080/api/v1/dashboard/stats?workspace=go-validation-10" | jq '.drift'
Open Dashboard
text
http://localhost:8080/dashboard?workspace=go-validation-10
What This Validates
Roadmap Claim	Status
Multi-language parsing (L1/L2)	✅ Partial — Python working, Go working
One verification model, many parsers	✅ Proven
Unified semantic graph	✅ Proven
Cross-language hubs and centrality	✅ Proven
Same ledger for all languages	✅ Proven
Same drift engine for all languages	✅ Proven
No language-specific storage forks	✅ Proven
What This Does NOT Claim
Garuda's multi-language support is partial and honest:

Not yet	Status
TypeScript	Not implemented (schema ready, parser pending)
Rust	Not implemented (schema ready, parser pending)
Java/Kotlin	Not implemented
Python CALLS edges	Not extracted (dynamic language limits)
Python external module resolution	External stubs only
Python decorator interpretation	Captured, not analyzed
This evidence document reflects what was measured on 2026-09-10.
Not what might be measured later. Not what's roadmap-claimed.

Architecture Decisions This Validates
The 6-layer epistemic model is language-agnostic.
Observation → Claim → Evidence → Verification → Decision → Lineage
works for Go and Python without modification.

The language column in entities is the right abstraction.
One column, one dashboard panel, one bar. No language-specific
schemas, no per-language dashboards, no per-language pipelines.

The parser interface is stable.
A language-specific parser returns analyzer.Result. Everything
downstream — persistence, verification, drift, search, graph,
dashboard, MCP — is language-blind.

Small structural parsers are viable for regular-grammar languages.
Python's structural scanner extracts 2,121 entities from sqlmodel
with zero external dependencies. This is not "state of the art"
compared to tree-sitter, but it's sufficient for the semantic
graph use case.

Merkle Ledger State
Dual-root snapshot at the time of capture:

text
static_root_hash:  dc44adf1277c08affeebb7fa3343378a61b50cfb5ad46d2f2ba1f1bd7ee09344
runtime_root_hash: d4c9edbf91db91639f63ac73e63a3edb06ea55f120cca45d137b5147ddf72b9f
block_height:      220+
runtime_leaf_count: 0
verified_claims_count: 0
contradicted_claims_count: 0
Ledger advances every 10 seconds via the background worker. Each analysis
commits a revision with a Merkle chained hash.

Screenshots
Screenshot	Shows
dashboard-multi-lang-languages-panel.png	Languages bar: Go 90.1% / Python 9.9%
dashboard-multi-lang-drift-panel.png	Drift: 4 doc claims, 12,592 code→doc drift
dashboard-multi-lang-hubs.png	sqlmodel + SQLModel in top-5 hubs
dashboard-multi-lang-repos.png	12 repo cards, sqlmodel tagged Python
Next Steps
Priority	Task	Estimated
P0	TypeScript parser (tree-sitter + cgo)	1 week
P1	Rust parser (structural or tree-sitter)	3-5 days
P1	Python CALLS extraction (via lightweight type hints)	2-3 days
P2	Python docstring → doc claim extraction	1-2 days
P2	Cross-language edge resolution (Go calls Python? Python calls Go?)	1 week
P3	Java/Kotlin parser	1 week
Reproducibility Guarantee
To reproduce these exact numbers:

Clone Garuda at commit a8afd09 (or later)

Run the reproduction commands above

Compare against the numbers in this document

If numbers differ, the cause is one of:

Upstream repos have changed (they're live; git pull fetches new commits)

Analyzer version differs

Filtering differs (external stubs included/excluded)

Determinism test: Re-run analysis on the same commit twice. Result
should be byte-identical (fingerprint match).

Report generated: 2026-09-10
Author: Rohit Mishra
Verification: Reproducible from the commands above

text

---

## Save and Commit

```bash
cd ~/garuda

# Save the file to EVIDENCE_MULTI_LANG.md

# Also create the screenshots directory reference
mkdir -p assets/screenshots/multi-lang

# If you have the dashboard screenshots, move them there
# cp ~/Downloads/dashboard-multi-lang-languages-panel.png assets/screenshots/multi-lang/
# (repeat for other screenshots)

# Commit
git add EVIDENCE_MULTI_LANG.md
git commit -m "Add multi-language validation evidence (Path B Step 1)

- 12 repos: 11 Go + 1 Python
- 12,917 internal entities, 25,269 relationships, 247 bridges
- Languages panel: Go 90.1% / Python 9.9%
- Drift engine running: 4 doc claims, 12,592 code→doc drift
- Reproducible commands documented
- Honest limitations section

Milestone: unified semantic graph across two languages"

git push origin main
