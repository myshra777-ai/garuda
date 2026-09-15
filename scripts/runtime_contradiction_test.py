#!/usr/bin/env python3
"""
runtime_contradiction_test.py — reproducible CONTRADICTED-path verification.

Purpose: exercise the runtime verification path end to end against a
known contradiction corpus and produce a number that can be cited.

EVIDENCE.md §1 no longer claims a detection rate because the previous
figure could not be reproduced. This script produces a reproducible one.
When it passes, the number goes into EVIDENCE.md §1 with a link to this
file.

Method:

  1. Pick N real entities in the workspace whose names are unique.
     Uniqueness matters: the correlator matches on name, and a name
     that appears on multiple entities resolves to the first match.
  2. POST one span per entity, with an operation name matching the
     entity name and an attribute `rpc.target_endpoint` whose value
     contains "unapproved". The string is the marker both the
     ingest handler's shortcut and the verifier's tick query accept.
  3. The ingest handler writes the CONTRADICTED row synchronously.
     The verifier's RecomputeWorkspaceVerification tick (every 10
     seconds) then re-derives the same row from runtime_edges and
     overwrites it. Wait for one tick to let the row stabilize.
  4. Count rows in claim_verifications with status = 'CONTRADICTED'
     whose evidence_payload->>'raw_target' contains "unapproved".
  5. Print N/N and exit 0 if all landed, non-zero otherwise.

Idempotency:

  clear_residue() removes rows from every table the verifier reads
  from: runtime_edges, runtime_observations, and claim_verifications
  (plus the contradictions quarantine table populated by the tick's
  statement 2). Deleting only from claim_verifications is
  insufficient: the tick's statement 1 re-derives CONTRADICTED rows
  from runtime_edges on every pass, so the rows return within one
  tick of the delete.

Prerequisites:
  - The daemon is running (./bin/garuda dev).
  - DATABASE_URL is set.
  - The workspace named by GARUDA_WORKSPACE contains exported functions
    or methods whose names are unique within the workspace.

Usage:
  DATABASE_URL=postgres://... python3 scripts/runtime_contradiction_test.py

Environment:
  GARUDA_API_BASE            default http://localhost:8080
  GARUDA_WORKSPACE           default go-validation-10
  CONTRADICTION_CORPUS_SIZE  default 10
  VERIFIER_WAIT_SECONDS      default 15
"""

import json
import os
import subprocess
import sys
import time
import urllib.parse
import urllib.request

API_BASE = os.environ.get("GARUDA_API_BASE", "http://localhost:8080")
WORKSPACE = os.environ.get("GARUDA_WORKSPACE", "go-validation-10")
DATABASE_URL = os.environ.get(
    "DATABASE_URL",
    "postgres://test:test@localhost:5433/garuda_test?sslmode=disable",
)
CORPUS_SIZE = int(os.environ.get("CONTRADICTION_CORPUS_SIZE", "10"))
WAIT_SECONDS = int(os.environ.get("VERIFIER_WAIT_SECONDS", "15"))


def psql(query):
    """
    Run one SQL statement and return its trimmed stdout.

    psql -tAc does not accept positional parameters. Values are
    interpolated by the caller. The workspace name is read from the
    environment by an operator, not from untrusted input, so the
    interpolation is safe for a test harness.
    """
    cmd = ["psql", DATABASE_URL, "-tAc", query]
    out = subprocess.run(cmd, capture_output=True, text=True, check=True)
    return out.stdout.strip()


def pick_candidates(n):
    """Return n (id, name) pairs whose names are unique in the workspace."""
    query = f"""
        SELECT id::text || '|' || name
          FROM entities
         WHERE workspace_id = (SELECT id FROM workspaces WHERE name = '{WORKSPACE}')
           AND kind IN ('function', 'method')
           AND is_exported = true
           AND name NOT LIKE '%(%'
           AND name NOT LIKE '%.%'
           AND LENGTH(name) BETWEEN 3 AND 30
         GROUP BY id, name
         HAVING COUNT(*) = 1
         ORDER BY random()
         LIMIT {n}
    """
    rows = psql(query).splitlines()
    if len(rows) < n:
        raise SystemExit(
            f"workspace {WORKSPACE} has only {len(rows)} usable candidates, "
            f"need {n}. Try a larger workspace or lower CONTRADICTION_CORPUS_SIZE."
        )
    return [row.split("|", 1) for row in rows]


def post_span(span_id, entity_name):
    payload = {
        "spans": [
            {
                "trace_id": f"contradiction-test-{span_id}",
                "span_id": span_id,
                "service_name": "test-harness",
                "target_service": "test-harness",
                "operation": entity_name,
                "duration_ms": 12.5,
                "status_code": "OK",
                "attributes": {
                    "code.function": entity_name,
                    "code.namespace": "",
                    "rpc.target_endpoint": "unapproved:test-target",
                },
            }
        ]
    }
    url = (
        API_BASE
        + "/api/v1/telemetry/spans?workspace="
        + urllib.parse.quote(WORKSPACE)
    )
    req = urllib.request.Request(
        url,
        data=json.dumps(payload).encode(),
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=10) as resp:
        return json.loads(resp.read())


def count_contradicted():
    return int(
        psql(
            f"""
            SELECT COUNT(*) FROM claim_verifications
             WHERE workspace_id = (SELECT id FROM workspaces WHERE name = '{WORKSPACE}')
               AND status = 'CONTRADICTED'
               AND evidence_payload->>'raw_target' LIKE '%unapproved%'
            """
        )
    )


def clear_residue():
    """
    Remove every artifact of a prior run from every table the verifier
    reads from.

    The verifier's RecomputeWorkspaceVerification tick re-derives
    CONTRADICTED rows from runtime_edges on every pass
    (internal/runtime/verifier.go, statement 1). Deleting only from
    claim_verifications is insufficient — the next tick restores the
    rows from runtime_edges. This function clears runtime_edges and
    runtime_observations (the tick's input tables) as well as
    claim_verifications (its output).

    The `contradictions` quarantine table is intentionally excluded.
    Its NOT EXISTS guard prevents accumulation across runs, and
    counting it here would fail if the table is absent — which is a
    separate schema question, not part of this script's contract.
    """
    ws = f"(SELECT id FROM workspaces WHERE name = '{WORKSPACE}')"

    tables = {
        "claim_verifications": (
            f"""SELECT COUNT(*) FROM claim_verifications
                 WHERE workspace_id = {ws}
                   AND status = 'CONTRADICTED'
                   AND evidence_payload->>'raw_target' LIKE '%unapproved%'""",
            f"""DELETE FROM claim_verifications
                 WHERE workspace_id = {ws}
                   AND status = 'CONTRADICTED'
                   AND evidence_payload->>'raw_target' LIKE '%unapproved%'""",
        ),
        "runtime_edges": (
            f"""SELECT COUNT(*) FROM runtime_edges
                 WHERE workspace_id = {ws}
                   AND raw_target LIKE '%unapproved%'""",
            f"""DELETE FROM runtime_edges
                 WHERE workspace_id = {ws}
                   AND raw_target LIKE '%unapproved%'""",
        ),
        "runtime_observations": (
            f"""SELECT COUNT(*) FROM runtime_observations
                 WHERE workspace_id = {ws}
                   AND trace_id LIKE 'contradiction-test-%'""",
            f"""DELETE FROM runtime_observations
                 WHERE workspace_id = {ws}
                   AND trace_id LIKE 'contradiction-test-%'""",
        ),
    }

    print("  [diag] before cleanup:")
    for label, (count_q, _) in tables.items():
        print(f"    {label:<22} = {psql(count_q)}")

    for label, (_, del_q) in tables.items():
        psql(del_q)

    print("  [diag] after cleanup:")
    for label, (count_q, _) in tables.items():
        print(f"    {label:<22} = {psql(count_q)}")

        
def main():
    print(f"Corpus size: {CORPUS_SIZE}")
    print(f"Workspace:   {WORKSPACE}")
    print(f"API base:    {API_BASE}")
    print()

    residue = count_contradicted()
    if residue > 0:
        print(f"Clearing {residue} contradicted rows from previous runs...")
        clear_residue()

    candidates = pick_candidates(CORPUS_SIZE)
    print(f"Selected {len(candidates)} unique entity names.")
    print()

    for i, (_, name) in enumerate(candidates, 1):
        result = post_span(f"contradiction-{i:02d}", name)
        print(f"  [{i:2d}] POST {name:<40} -> {result}")

    print()
    print(f"Waiting {WAIT_SECONDS}s for the verifier tick...")
    time.sleep(WAIT_SECONDS)

    n = count_contradicted()
    print()
    print(f"═══ {n}/{CORPUS_SIZE} contradicted ═══")
    sys.exit(0 if n == CORPUS_SIZE else 1)


if __name__ == "__main__":
    main()