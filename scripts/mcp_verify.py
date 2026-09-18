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
        results.append(expect(len(tools) >= 21, f"tools/list >= 21 tools (got {len(tools)})"))

        names = {t["name"] for t in tools}
        results.append(expect("garuda.briefing" in names, "tool garuda.briefing present"))
        results.append(expect(
            "garuda.verify_policy_evaluation" in names,
            "tool garuda.verify_policy_evaluation present"))
        results.append(expect(
            "garuda.blast_radius" in names,
            "tool garuda.blast_radius present"))
        results.append(expect("garuda.handoff" in names, "tool garuda.handoff present"))
        results.append(expect("garuda.resume" in names, "tool garuda.resume present"))
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

        # garuda.blast_radius — SQLModel has 171 inbound edges (proven
        # above). Blast radius must find at least one affected entity
        # and every impacted entity must carry a severity.
        send(proc, {
            "jsonrpc": "2.0",
            "id": 8,
            "method": "tools/call",
            "params": {
                "name": "garuda.blast_radius",
                "arguments": {"workspace": WORKSPACE, "entity_id": entity_id, "depth": 2},
            },
        })
        br = read_response(proc)
        results.append(expect(br.get("id") == 8, "blast_radius id == 8"))
        brp = json.loads(br["result"]["content"][0]["text"])
        results.append(expect(brp.get("target_found") is True, "blast_radius target_found == true"))
        results.append(expect(
            brp.get("total_affected", 0) >= 1,
            "blast_radius found >= 1 affected entity"))
        results.append(expect(
            all("severity" in e for e in brp.get("impacted", [])),
            "every impacted entity carries a severity"))

        # garuda.handoff — random UUIDs. The store's lockAgent will not
        # find the source agent, and ExecuteHandoffTransaction returns
        # an error. The MCP handler translates that to
        # {status: "failed", reason: ...}, not a transport error.
        send(proc, {
            "jsonrpc": "2.0",
            "id": 9,
            "method": "tools/call",
            "params": {
                "name": "garuda.handoff",
                "arguments": {
                    "task_id": "00000000-0000-0000-0000-000000000001",
                    "source_agent_id": "00000000-0000-0000-0000-000000000002",
                    "target_agent_id": "00000000-0000-0000-0000-000000000003",
                },
            },
        })
        ho = read_response(proc)
        results.append(expect(ho.get("id") == 9, "handoff id == 9"))
        hop = json.loads(ho["result"]["content"][0]["text"])
        results.append(expect(hop.get("status") == "failed",
                              "handoff returns status=failed for unknown task/agent"))
        results.append(expect("reason" in hop, "handoff includes a reason"))

        # garuda.resume — random checkpoint UUID. The store cannot find
        # the checkpoint (or the agent). The MCP handler translates to
        # {status: "not_found", reason: ...}.
        send(proc, {
            "jsonrpc": "2.0",
            "id": 10,
            "method": "tools/call",
            "params": {
                "name": "garuda.resume",
                "arguments": {
                    "agent_id": "00000000-0000-0000-0000-000000000004",
                    "checkpoint_id": "00000000-0000-0000-0000-000000000005",
                },
            },
        })
        rs = read_response(proc)
        results.append(expect(rs.get("id") == 10, "resume id == 10"))
        rsp = json.loads(rs["result"]["content"][0]["text"])
        results.append(expect(rsp.get("status") == "not_found",
                              "resume returns status=not_found for unknown checkpoint"))
        results.append(expect("reason" in rsp, "resume includes a reason"))

        # Negative path: unknown tool name. The whitelist must reject
        # with JSON-RPC error -32601 (Method not found) and must not
        # write a row to mcp_tool_calls. The absence of a row is
        # verified out-of-band with the psql check documented next to
        # this file; this script can only assert the wire response.
        send(proc, {
            "jsonrpc": "2.0", "id": 11, "method": "tools/call",
            "params": {"name": "garuda.nonexistent", "arguments": {}},
        })
        unk = read_response(proc)
        results.append(expect(unk.get("id") == 11, "unknown tool id == 11"))
        results.append(expect(
            unk.get("error", {}).get("code") == -32601,
            "unknown tool returns -32601"))

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