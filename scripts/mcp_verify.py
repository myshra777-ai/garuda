#!/usr/bin/env python3
"""
mcp_verify.py — a minimal second implementation of an MCP client.

Purpose: validate that bin/garuda-mcp is spec-compliant, not merely
Cursor-compatible. Cursor exposed one real bug (responses to
notifications) that a second implementation would have caught sooner.
This script is that second implementation.

It is deliberately small and dependency-free: stdlib only. It speaks
the same wire protocol every MCP client speaks.

Usage:
  cd ~/garuda
  python3 scripts/mcp_verify.py
"""

import json
import os
import subprocess
import sys

BINARY = os.path.join(os.path.dirname(__file__), "..", "bin", "garuda-mcp")
WORKSPACE = os.environ.get("WORKSPACE", "go-validation-10")


def send(proc, obj):
    proc.stdin.write((json.dumps(obj) + "\n").encode())
    proc.stdin.flush()


def read_response(proc):
    line = proc.stdout.readline()
    if not line:
        raise RuntimeError("server closed stdout unexpectedly")
    return json.loads(line)


def expect(cond, msg):
    if not cond:
        print(f"  FAIL: {msg}", file=sys.stderr)
        return False
    print(f"  ok:   {msg}")
    return True


def main():
    env = dict(os.environ)
    env.setdefault(
        "DATABASE_URL",
        "postgres://test:test@localhost:5433/garuda_test?sslmode=disable",
    )
    env["GARUDA_MCP_QUIET"] = "1"

    proc = subprocess.Popen(
        [BINARY],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.DEVNULL,
        env=env,
    )

    results = []

    try:
        send(proc, {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize",
            "params": {
                "protocolVersion": "2025-06-18",
                "capabilities": {},
                "clientInfo": {"name": "python-verify", "version": "0.1"},
            },
        })
        init = read_response(proc)
        results.append(expect(init.get("id") == 1, "initialize id == 1"))
        results.append(expect(
            init.get("result", {}).get("protocolVersion") == "2025-06-18",
            "protocolVersion == 2025-06-18",
        ))

        # notifications/initialized MUST NOT receive a response. If the
        # server responds, the next read picks up the notification
        # response instead of the tools/list response.
        send(proc, {"jsonrpc": "2.0", "method": "notifications/initialized"})

        send(proc, {"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}})
        tl = read_response(proc)
        results.append(expect(tl.get("id") == 2, "tools/list id == 2 (no notification response)"))
        tools = tl.get("result", {}).get("tools", [])
        results.append(expect(len(tools) >= 18, f"tools/list >= 18 tools (got {len(tools)})"))

        names = {t["name"] for t in tools}
        results.append(expect("garuda.briefing" in names, "tool garuda.briefing present"))
        results.append(expect(
            "garuda.verify_policy_evaluation" in names,
            "tool garuda.verify_policy_evaluation present"))
        for expected in (
            "garuda.entities", "garuda.neighbors", "garuda.find_entity",
            "garuda.subclasses", "garuda.implementers",
        ):
            results.append(expect(expected in names, f"tool {expected} present"))

        send(proc, {
            "jsonrpc": "2.0", "id": 3, "method": "tools/call",
            "params": {
                "name": "garuda.find_entity",
                "arguments": {"workspace": WORKSPACE, "name_pattern": "^SQLModel$", "limit": 1},
            },
        })
        fe = read_response(proc)
        results.append(expect(fe.get("id") == 3, "find_entity id == 3"))
        payload = json.loads(fe["result"]["content"][0]["text"])
        entity_id = payload["results"][0]["id"] if payload["results"] else None
        results.append(expect(bool(entity_id), "find_entity resolved SQLModel"))
        results.append(expect(
            payload["results"][0]["inbound_edges"] == 171,
            "SQLModel has 171 inbound edges",
        ))

        send(proc, {
            "jsonrpc": "2.0", "id": 4, "method": "tools/call",
            "params": {
                "name": "garuda.subclasses",
                "arguments": {"workspace": WORKSPACE, "entity_id": entity_id, "limit": 3},
            },
        })
        sc = read_response(proc)
        scp = json.loads(sc["result"]["content"][0]["text"])
        results.append(expect(scp["count"] >= 1, "subclasses returned >= 1"))
        results.append(expect(
            all(r["claim_type"] == "INHERITS" for r in scp["results"]),
            "all subclass claim_type == INHERITS",
        ))

        send(proc, {
            "jsonrpc": "2.0", "id": 5, "method": "tools/call",
            "params": {
                "name": "garuda.neighbors",
                "arguments": {"workspace": WORKSPACE, "entity_id": entity_id, "limit": 5},
            },
        })
        nb = read_response(proc)
        nbp = json.loads(nb["result"]["content"][0]["text"])
        results.append(expect(nbp["count"] >= 1, "neighbors returned >= 1"))
        results.append(expect(
            all("direction" in r for r in nbp["results"]),
            "every neighbor has a direction field",
        ))

        send(proc, {
            "jsonrpc": "2.0", "id": 6, "method": "tools/call",
            "params": {
                "name": "garuda.implementers",
                "arguments": {"workspace": WORKSPACE, "name": "CmdTyper", "limit": 3},
            },
        })
        im = read_response(proc)
        imp = json.loads(im["result"]["content"][0]["text"])
        results.append(expect(imp["count"] >= 1, "implementers returned >= 1"))

        # garuda.verify_policy_evaluation — a random UUID is not a valid
        # evaluation for this tenant. The tool returns {valid: false,
        # reason: ...}, not a transport error.
        #
        # Request id 7, not 6: id 6 is already the implementers call
        # above. Reusing it would make the response indistinguishable
        # from a duplicate and mask a transport-level bug.
        send(proc, {
            "jsonrpc": "2.0",
            "id": 7,
            "method": "tools/call",
            "params": {
                "name": "garuda.verify_policy_evaluation",
                "arguments": {"evaluation_id": "00000000-0000-0000-0000-000000000000"},
            },
        })
        vpe = read_response(proc)
        results.append(expect(vpe.get("id") == 7, "verify_policy_evaluation id == 7"))
        vpep = json.loads(vpe["result"]["content"][0]["text"])
        results.append(expect(
            vpep.get("valid") is False,
            "verify_policy_evaluation returns valid=false for unknown evaluation"))
        results.append(expect(
            "reason" in vpep,
            "verify_policy_evaluation includes a reason"))

        proc.stdin.close()
        proc.wait(timeout=3)
        results.append(expect(proc.returncode == 0, "server exited cleanly"))

    finally:
        if proc.poll() is None:
            proc.kill()

    passed = sum(results)
    total = len(results)
    print()
    print(f"═══ {passed}/{total} checks passed ═══")
    sys.exit(0 if passed == total else 1)


if __name__ == "__main__":
    main()