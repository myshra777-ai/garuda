#!/usr/bin/env python3
"""
runtime_contradiction_test.py — reproducible CONTRADICTED-path verification.

Purpose: exercise the runtime verification path end to end against a
known contradiction corpus and produce a number that can be cited.
"""

import json
import os
import subprocess
import sys
import time
import urllib.error
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
    cmd = ["psql", DATABASE_URL, "-tAc", query]
    out = subprocess.run(cmd, capture_output=True, text=True, check=True)
    return out.stdout.strip()


def obtain_auth_headers():
    """
    Obtain an authenticated Ed25519 Bearer JWT or session cookie for /api/v1 requests.
    """
    token = os.environ.get("GARUDA_TOKEN", "").strip()
    if token:
        return {"Content-Type": "application/json", "Authorization": f"Bearer {token}"}

    email = os.environ.get("GARUDA_EMAIL", "admin@garuda.local")
    password = os.environ.get("GARUDA_PASSWORD", "garuda2026")

    # Attempt token login first
    for endpoint in ("/api/v1/auth/login", "/api/v1/auth/signup"):
        try:
            body = json.dumps({"email": email, "password": password}).encode()
            req = urllib.request.Request(
                API_BASE + endpoint,
                data=body,
                headers={"Content-Type": "application/json"},
                method="POST",
            )
            with urllib.request.urlopen(req, timeout=5) as resp:
                data = json.loads(resp.read())
                if "access_token" in data:
                    return {
                        "Content-Type": "application/json",
                        "Authorization": f"Bearer {data['access_token']}",
                    }
        except Exception:
            pass

    # Fallback: form signup/login to obtain garuda_session cookie
    form_data = urllib.parse.urlencode({"email": email, "password": password}).encode()
    for endpoint in ("/signup", "/login"):
        try:
            req = urllib.request.Request(
                API_BASE + endpoint,
                data=form_data,
                headers={"Content-Type": "application/x-www-form-urlencoded"},
                method="POST",
            )
            with urllib.request.urlopen(req, timeout=5) as resp:
                cookie = resp.headers.get("Set-Cookie", "")
                if "garuda_session=" in cookie:
                    sess_kv = cookie.split(";", 1)[0]
                    return {"Content-Type": "application/json", "Cookie": sess_kv}
        except urllib.error.HTTPError as e:
            cookie = e.headers.get("Set-Cookie", "") if e.headers else ""
            if "garuda_session=" in cookie:
                sess_kv = cookie.split(";", 1)[0]
                return {"Content-Type": "application/json", "Cookie": sess_kv}
        except Exception:
            pass

    return {"Content-Type": "application/json"}


def pick_candidates(n):
    """Return n (id, name) pairs whose names are unique in the workspace."""
    query = f"""
        SELECT (array_agg(id::text))[1] || '|' || name
          FROM entities
         WHERE workspace_id = (SELECT id FROM workspaces WHERE name = '{WORKSPACE}')
           AND kind IN ('function', 'method')
           AND is_exported = true
           AND name NOT LIKE '%(%'
           AND name NOT LIKE '%.%'
           AND LENGTH(name) BETWEEN 3 AND 30
         GROUP BY name
         HAVING COUNT(*) = 1
         ORDER BY random()
         LIMIT {n}
    """
    rows = [r for r in psql(query).splitlines() if "|" in r]
    if len(rows) < n:
        raise SystemExit(f"Need {n} unique exported functions in workspace {WORKSPACE}, found {len(rows)}")
    return [row.split("|", 1) for row in rows]


def post_span(span_id, entity_name, headers):
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
        headers=headers,
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

    headers = obtain_auth_headers()

    residue = count_contradicted()
    if residue > 0:
        print(f"Clearing {residue} contradicted rows from previous runs...")
        clear_residue()

    candidates = pick_candidates(CORPUS_SIZE)
    print(f"Selected {len(candidates)} unique entity names.")
    print()

    for i, (_, name) in enumerate(candidates, 1):
        result = post_span(f"contradiction-{i:02d}", name, headers)
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