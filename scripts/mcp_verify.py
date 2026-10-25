#!/usr/bin/env python3
"""
mcp_verify.py — spec-compliance, graph-intelligence, and Garuda Talon active
enforcement verification for bin/garuda-mcp.

Dynamically discovers entities and interfaces from the active workspace (no hardcoded
repository names or magic edge counts) and loads .env automatically if present.

Usage:
  cd ~/garuda
  python3 scripts/mcp_verify.py
"""

import json
import os
import subprocess
import sys

ROOT_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
BINARY = os.path.join(ROOT_DIR, "bin", "garuda-mcp")


def load_dotenv():
    env_path = os.path.join(ROOT_DIR, ".env")
    if not os.path.exists(env_path):
        return
    with open(env_path, "r") as f:
        for raw in f:
            line = raw.strip()
            if not line or line.startswith("#"):
                continue
            if line.startswith("export "):
                line = line[7:].strip()
            if "=" in line:
                k, v = line.split("=", 1)
                k = k.strip()
                v = v.strip().strip('"').strip("'")
                if k and k not in os.environ:
                    os.environ[k] = v


load_dotenv()
WORKSPACE = os.environ.get("WORKSPACE") or os.environ.get("GARUDA_WORKSPACE") or ""


def send(proc, obj):
    proc.stdin.write((json.dumps(obj) + "\n").encode())
    proc.stdin.flush()


def read_response(proc):
    line = proc.stdout.readline()
    if not line:
        raise RuntimeError("server closed stdout unexpectedly")
    return json.loads(line)


def unpack_tool_payload(resp, label):
    if "result" not in resp:
        err = resp.get("error", {})
        print(f"  FAIL: {label} returned error: {err}", file=sys.stderr)
        return {}
    content = resp["result"].get("content", [])
    if not content:
        return {}
    return json.loads(content[0].get("text", "{}"))


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
    env["GARUDA_TALON_ENFORCE"] = "true"

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
                "clientInfo": {"name": "python-verify", "version": "1.5"},
            },
        })
        init = read_response(proc)
        results.append(expect(init.get("id") == 1, "initialize id == 1"))
        results.append(expect(
            init.get("result", {}).get("protocolVersion") == "2025-06-18",
            "protocolVersion == 2025-06-18",
        ))
        caps = init.get("result", {}).get("capabilities", {})
        results.append(expect("prompts" in caps and "resources" in caps,
                              "initialize advertises prompts and resources capabilities"))

        # notifications/initialized MUST NOT receive a response.
        send(proc, {"jsonrpc": "2.0", "method": "notifications/initialized"})

        send(proc, {"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}})
        tl = read_response(proc)
        results.append(expect(tl.get("id") == 2, "tools/list id == 2 (no notification response)"))
        tools = tl.get("result", {}).get("tools", [])
        results.append(expect(len(tools) >= 23, f"tools/list >= 23 tools (got {len(tools)})"))

        names = {t["name"] for t in tools}
        for expected in (
            "garuda.briefing",
            "garuda.verify_policy_evaluation",
            "garuda.blast_radius",
            "garuda.handoff",
            "garuda.resume",
            "garuda.entities",
            "garuda.neighbors",
            "garuda.find_entity",
            "garuda.subclasses",
            "garuda.implementers",
            "garuda.talon_pre_edit",
            "garuda.talon_verify_edit",
        ):
            results.append(expect(expected in names, f"tool {expected} present"))

        # 3. Dynamic entity resolution: rank by inbound edges across the workspace
        send(proc, {
            "jsonrpc": "2.0", "id": 3, "method": "tools/call",
            "params": {
                "name": "garuda.find_entity",
                "arguments": {"workspace": WORKSPACE, "limit": 10},
            },
        })
        fe = read_response(proc)
        results.append(expect(fe.get("id") == 3, "find_entity id == 3"))
        payload = unpack_tool_payload(fe, "garuda.find_entity")
        results.append(expect(
            "talon_active_contract" in payload or "talon_enforcement" in payload,
            "first tool call automatically injects active Talon contract guardrails",
        ))
        ent_list = payload.get("results", [])
        top_entity = ent_list[0] if ent_list else {}
        for candidate in ent_list:
            if candidate.get("name") == "SQLModel":
                top_entity = candidate
                break
        entity_id = top_entity.get("id")
        entity_name = top_entity.get("name", "unknown")
        inbound_edges = top_entity.get("inbound_edges", 0)
        results.append(expect(bool(entity_id), f"find_entity resolved top hub ({entity_name})"))
        results.append(expect(
            inbound_edges >= 1,
            f"{entity_name} has >= 1 inbound edges (got {inbound_edges})",
        ))

        # 4. Subclasses / Embeds check
        send(proc, {
            "jsonrpc": "2.0", "id": 4, "method": "tools/call",
            "params": {
                "name": "garuda.subclasses",
                "arguments": {"workspace": WORKSPACE, "entity_id": entity_id, "limit": 3},
            },
        })
        sc = read_response(proc)
        scp = unpack_tool_payload(sc, "garuda.subclasses")
        results.append(expect(scp.get("count", 0) >= 1, "subclasses returned >= 1"))
        results.append(expect(
            all(r.get("claim_type") in ("INHERITS", "EMBEDS") for r in scp.get("results", [])),
            "all subclass claim_type in (INHERITS, EMBEDS)",
        ))

        # 5. Neighbors check
        send(proc, {
            "jsonrpc": "2.0", "id": 5, "method": "tools/call",
            "params": {
                "name": "garuda.neighbors",
                "arguments": {"workspace": WORKSPACE, "entity_id": entity_id, "limit": 5},
            },
        })
        nb = read_response(proc)
        nbp = unpack_tool_payload(nb, "garuda.neighbors")
        results.append(expect(nbp.get("count", 0) >= 1, "neighbors returned >= 1"))
        results.append(expect(
            all("direction" in r for r in nbp.get("results", [])),
            "every neighbor has a direction field",
        ))

        # 6. Implementers check
        send(proc, {
            "jsonrpc": "2.0", "id": 6, "method": "tools/call",
            "params": {
                "name": "garuda.implementers",
                "arguments": {"workspace": WORKSPACE, "name": "Router", "limit": 3},
            },
        })
        im = read_response(proc)
        imp = unpack_tool_payload(im, "garuda.implementers")
        results.append(expect(imp.get("count", 0) >= 1, f"implementers returned >= 1 (subject: {imp.get('subject', {}).get('name', 'auto')})"))

        # 7. garuda.verify_policy_evaluation
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
        vpep = unpack_tool_payload(vpe, "garuda.verify_policy_evaluation")
        results.append(expect(
            vpep.get("valid") is False,
            "verify_policy_evaluation returns valid=false for unknown evaluation"))
        results.append(expect(
            "reason" in vpep,
            "verify_policy_evaluation includes a reason"))

        # 8. garuda.blast_radius
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
        brp = unpack_tool_payload(br, "garuda.blast_radius")
        results.append(expect(brp.get("target_found") is True, "blast_radius target_found == true"))
        results.append(expect(
            brp.get("total_affected", 0) >= 1,
            "blast_radius found >= 1 affected entity"))
        results.append(expect(
            all("severity" in e for e in brp.get("impacted", [])),
            "every impacted entity carries a severity"))

        # 9. garuda.handoff
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
        hop = unpack_tool_payload(ho, "garuda.handoff")
        results.append(expect(hop.get("status") == "failed",
                              "handoff returns status=failed for unknown task/agent"))
        results.append(expect("reason" in hop, "handoff includes a reason"))

        # 10. garuda.resume
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
        rsp = unpack_tool_payload(rs, "garuda.resume")
        results.append(expect(rsp.get("status") == "not_found",
                              "resume returns status=not_found for unknown checkpoint"))
        results.append(expect("reason" in rsp, "resume includes a reason"))

        # 11. garuda.talon_pre_edit & prompts/list & resources/list
        send(proc, {
            "jsonrpc": "2.0",
            "id": 11,
            "method": "tools/call",
            "params": {
                "name": "garuda.talon_pre_edit",
                "arguments": {"workspace": WORKSPACE, "target": entity_name},
            },
        })
        tpe = read_response(proc)
        tpep = unpack_tool_payload(tpe, "garuda.talon_pre_edit")
        results.append(expect(
            "mandatory_protocol" in tpep and "active_policies" in tpep,
            "garuda.talon_pre_edit returns mandatory_protocol and active_policies",
        ))

        send(proc, {"jsonrpc": "2.0", "id": 12, "method": "prompts/list", "params": {}})
        pl = read_response(proc)
        prompts = pl.get("result", {}).get("prompts", [])
        results.append(expect(
            any(p.get("name") == "garuda.architectural_contract" for p in prompts),
            "prompts/list exposes garuda.architectural_contract",
        ))

        send(proc, {"jsonrpc": "2.0", "id": 13, "method": "resources/list", "params": {}})
        rl = read_response(proc)
        resources = rl.get("result", {}).get("resources", [])
        results.append(expect(
            any(r.get("uri") == "garuda://workspace/contract" for r in resources),
            "resources/list exposes garuda://workspace/contract",
        ))

        # 12. Negative path: unknown tool name
        send(proc, {
            "jsonrpc": "2.0", "id": 14, "method": "tools/call",
            "params": {"name": "garuda.nonexistent", "arguments": {}},
        })
        unk = read_response(proc)
        results.append(expect(unk.get("id") == 14, "unknown tool id == 14"))
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