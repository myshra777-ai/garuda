# Garuda Hygiene

Garuda Hygiene is a read-only, report-only analysis surface for advisory static-code findings derived from the scoped semantic workspace graph.

The canonical command is:

```bash
garuda hygiene [path]
```

The historical command remains available as a deprecated compatibility alias:

```bash
garuda ponytail [path]
```

`garuda ponytail` and `garuda hygiene` use the same analyzer and produce the same report. New scripts and documentation should use `garuda hygiene`.

## What Hygiene reports

Hygiene currently reports three advisory categories.

| Category | Meaning | Does not prove |
| :--- | :--- | :--- |
| Static unreferenced candidate | A non-package, non-file entity has no incoming relationship in the indexed workspace graph. | Dead code, unreachable code, or safe removal. |
| Duplicate symbol-name candidate | A non-empty symbol name appears in more than one package. | Duplicated implementation, a conflict, or a refactoring opportunity. |
| Standard-library alternative suggestion | A symbol name contains `Contains` or `Sort`, matching a naming heuristic. | That the implementation is replaceable by `slices.Contains` or `slices.Sort`. |

The word `candidate` is intentional. Hygiene makes the evidence boundary visible; it does not convert missing graph evidence into a definitive claim.

## Commands

Human-readable output:

```bash
GARUDA_WORKSPACE=go-validation-10 \
GARUDA_TENANT_ID=00000000-0000-0000-0000-000000000001 \
./bin/garuda hygiene .
```

JSON output written to a file:

```bash
GARUDA_WORKSPACE=go-validation-10 \
GARUDA_TENANT_ID=00000000-0000-0000-0000-000000000001 \
./bin/garuda hygiene . --json -o /tmp/garuda-hygiene.json
```

The command accepts at most one positional path. The current report reads the already indexed semantic graph for the resolved workspace; it does not analyze the path or refresh the graph.

## Workspace scope

Hygiene resolves the workspace using the active tenant and workspace configuration before reading entities and relationships. The entity and graph reads are scoped by both tenant ID and workspace ID.

The relevant environment variables are:

```bash
export GARUDA_TENANT_ID=00000000-0000-0000-0000-000000000001
export GARUDA_WORKSPACE=go-validation-10
```

Environment variables select the requested context. Tenant and workspace isolation are enforced by the Go store queries; an environment variable alone is not an authorization boundary.

## JSON contract

The current top-level JSON contract is:

```json
{
  "dead_code": [],
  "duplications": [],
  "stdlib_alternatives": [],
  "summary": {
    "dead_code": 0,
    "duplications": 0,
    "stdlib_alts": 0
  },
  "total_entities": 0,
  "total_relationships": 0
}
```

The legacy top-level field names are preserved for compatibility:

- `dead_code` contains static unreferenced candidates.
- `duplications` contains aggregated duplicate symbol-name candidates.
- `stdlib_alternatives` contains standard-library alternative suggestions.
- `summary` contains the corresponding counts.
- `total_entities` and `total_relationships` describe the scoped graph input.

The internal analyzer uses more precise category names and returns typed findings. The CLI projects those findings into the legacy report fields so existing scripts continue to work.

## Finding semantics

### Static unreferenced candidates

An entity is included when it has zero incoming relationships in the indexed graph and is not a package or file entity.

The result means:

```text
No incoming graph reference was found in the indexed workspace graph.
```

It does not mean:

```text
This entity is dead and may be deleted safely.
```

Entry points, reflection, generated registration, external consumers, build tags, dynamic loading, incomplete indexing, and unsupported language behavior can all make a zero-incoming result insufficient for removal.

### Duplicate symbol-name candidates

Hygiene groups entities by non-empty symbol name and reports one finding when that name occurs in more than one package. Package names are sorted deterministically and repeated entities in the same package do not create additional package entries.

The finding is a naming signal, not an implementation-equivalence test. Two functions with the same name may be unrelated and entirely correct.

### Standard-library alternative suggestions

Hygiene currently applies a naming heuristic:

- Names containing `Contains` suggest `slices.Contains`.
- Names containing `Sort` suggest `slices.Sort`.

This does not inspect function bodies, generic constraints, ordering semantics, comparator behavior, or API compatibility. Review the implementation before changing code.

## Determinism and malformed input

The pure analyzer does not mutate its input slices or external state. Findings are sorted deterministically by category, name, package, file, line range, entity ID, and message.

The CLI compatibility adapter converts the existing untyped graph edge maps into typed relationships. Rows with missing, non-string, or empty `from`, `to`, or `type` values are skipped rather than converted into fabricated relationships.

The current adapter does not persist a warning count for skipped rows. This is a known limitation of the existing untyped `GetGraphData` contract.

## Read-only boundary

The current Hygiene report performs scoped reads and pure in-memory analysis. It does not:

- Modify source files.
- Write graph entities or relationships.
- Persist Hygiene findings.
- Create governance decisions.
- Anchor a Merkle record.
- Apply code fixes.
- Suppress findings.
- Merge changes.

## Reference verification

The current reference workspace verification was run with:

```bash
GARUDA_WORKSPACE=go-validation-10 \
GARUDA_TENANT_ID=00000000-0000-0000-0000-000000000001 \
./bin/garuda hygiene . --json -o /tmp/garuda-hygiene.json
```

The measured report contained:

```text
Entities: 22905
Relationships: 40956
Static unreferenced candidates: 11919
Duplicate symbol-name candidates: 1480
Standard-library alternatives: 59
```

These are counts from one indexed workspace snapshot. They are not quality scores, defect counts, or proof that every finding is actionable.

The compatibility alias was verified with:

```bash
GARUDA_WORKSPACE=go-validation-10 \
GARUDA_TENANT_ID=00000000-0000-0000-0000-000000000001 \
./bin/garuda ponytail . --json -o /tmp/garuda-ponytail.json

diff -u \
  <(jq -S . /tmp/garuda-hygiene.json) \
  <(jq -S . /tmp/garuda-ponytail.json)
```

The sorted JSON reports were identical.

## Current limitations

- Hygiene reads the existing indexed graph; it does not refresh analysis.
- The current graph-edge store contract is untyped at the CLI boundary.
- Static unreferenced analysis does not account for all dynamic or external references.
- Duplicate detection is based on symbol names and package identity, not body similarity.
- Standard-library suggestions are naming heuristics.
- Findings are not persisted, suppressed, baselined, or exposed through MCP.
- There is no VS Code diagnostic lifecycle.
- There is no autonomous remediation or merge behavior.

These limitations are intentional for the report-only Hygiene arc.

## Future work

Future work may add a read-only MCP surface, analysis freshness metadata, persisted observations, suppression and baseline semantics, and IDE presentation. Each requires a separate contract and evidence review; none should be inferred from the current CLI.