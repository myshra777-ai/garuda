#!/usr/bin/env python3
"""
garuda_bench.py — Public, Reproducible Benchmark Harness
Compares raw LLM performance (Control) vs. LLM + Garuda MCP Server (Treatment).
Measures: Hallucination Rate, Exact Token Usage, Wall-Clock Latency, F1-Accuracy.
"""

import argparse
import json
import os
import subprocess
import sys
import time
from typing import Any, Dict, List, Set, Tuple

try:
    import anthropic
except ImportError:
    print("Error: anthropic package required. Run: pip install anthropic", file=sys.stderr)
    sys.exit(1)


class GarudaMCPClient:
    """Spawns bin/garuda-mcp as a subprocess and speaks JSON-RPC over stdio."""

    def __init__(self, binary_path: str, repo_dir: str):
        env = dict(os.environ)
        env["GARUDA_MCP_QUIET"] = "1"
        self.proc = subprocess.Popen(
            [binary_path],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            cwd=repo_dir,
            env=env,
        )
        self.msg_id = 0
        self._init_session()

    def _init_session(self):
        self.call("initialize", {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {"name": "garuda-bench-harness", "version": "1.0"}
        })

    def call(self, method: str, params: Dict[str, Any]) -> Dict[str, Any]:
        self.msg_id += 1
        req = {
            "jsonrpc": "2.0",
            "id": self.msg_id,
            "method": method,
            "params": params,
        }
        payload = json.dumps(req) + "\n"
        self.proc.stdin.write(payload.encode("utf-8"))
        self.proc.stdin.flush()
        raw = self.proc.stdout.readline()
        if not raw:
            return {}
        res = json.loads(raw.decode("utf-8"))
        return res.get("result", {})

    def call_tool(self, name: str, arguments: Dict[str, Any]) -> Dict[str, Any]:
        result = self.call("tools/call", {"name": name, "arguments": arguments})
        content = result.get("content", [])
        if content and "text" in content[0]:
            try:
                return json.loads(content[0]["text"])
            except json.JSONDecodeError:
                return {"raw_text": content[0]["text"]}
        return result

    def close(self):
        try:
            self.proc.terminate()
            self.proc.wait(timeout=2)
        except Exception:
            self.proc.kill()


def evaluate_ground_truth(mcp: GarudaMCPClient, task: Dict[str, Any]) -> Set[str]:
    """Retrieves deterministic ground-truth items from Garuda's compiler engine."""
    tool = task["ground_truth_tool"]
    args = task["ground_truth_args"]
    res = mcp.call_tool(tool, args)

    target_key = task.get("target_key")
    match_field = task.get("match_field")

    gt_set = set()
    if target_key and target_key in res:
        items = res[target_key]
        if isinstance(items, list):
            for it in items:
                if isinstance(it, dict) and match_field in it:
                    gt_set.add(str(it[match_field]).strip())
                else:
                    gt_set.add(str(it).strip())
        else:
            gt_set.add(str(items).strip())
    return gt_set


def run_control_arm(client: anthropic.Anthropic, model: str, prompt: str) -> Dict[str, Any]:
    """Arm A: Raw Claude relying on internalized pre-training and guessing."""
    t0 = time.perf_counter()
    resp = client.messages.create(
        model=model,
        max_tokens=4096,
        system="You are a senior software architect. Analyze the requested codebase. Respond with exact lists in valid JSON format.",
        messages=[{"role": "user", "content": prompt}]
    )
    duration = time.perf_counter() - t0
    return {
        "text": resp.content[0].text,
        "input_tokens": resp.usage.input_tokens,
        "output_tokens": resp.usage.output_tokens,
        "total_tokens": resp.usage.input_tokens + resp.usage.output_tokens,
        "latency_sec": round(duration, 3)
    }


def run_treatment_arm(client: anthropic.Anthropic, model: str, prompt: str, tool_context: Dict[str, Any]) -> Dict[str, Any]:
    """Arm B: Claude augmented with Garuda's verified graph context."""
    t0 = time.perf_counter()
    ctx_str = json.dumps(tool_context, indent=2)
    resp = client.messages.create(
        model=model,
        max_tokens=4096,
        system=f"You are a verified architectural evaluator. Use the following ground-truth Garuda graph data to answer:\n{ctx_str}",
        messages=[{"role": "user", "content": prompt}]
    )
    duration = time.perf_counter() - t0
    return {
        "text": resp.content[0].text,
        "input_tokens": resp.usage.input_tokens,
        "output_tokens": resp.usage.output_tokens,
        "total_tokens": resp.usage.input_tokens + resp.usage.output_tokens,
        "latency_sec": round(duration, 3)
    }


def compute_metrics(predicted_text: str, ground_truth: Set[str]) -> Tuple[float, float, float, float]:
    """
    Computes Precision, Recall, F1, and Hallucination Rate.
    Hallucination = generated entity names that do NOT exist in the compiler ground-truth.
    """
    if not ground_truth:
        return 1.0, 1.0, 1.0, 0.0

    # Extract words/tokens that look like symbols
    import re
    tokens = set(re.findall(r'[A-Za-z_][A-Za-z0-9_.]+', predicted_text))

    # Match tokens against ground truth
    true_positives = len(tokens.intersection(ground_truth))
    false_positives = len(tokens - ground_truth)
    false_negatives = len(ground_truth - tokens)

    precision = true_positives / (true_positives + false_positives) if (true_positives + false_positives) > 0 else 0.0
    recall = true_positives / (true_positives + false_negatives) if (true_positives + false_negatives) > 0 else 0.0
    f1 = (2 * precision * recall) / (precision + recall) if (precision + recall) > 0 else 0.0
    hallucination_rate = false_positives / len(tokens) if len(tokens) > 0 else 0.0

    return round(precision, 3), round(recall, 3), round(f1, 3), round(hallucination_rate, 3)


def main():
    parser = argparse.ArgumentParser(description="Garuda Rigorous Evaluation Harness")
    parser.add_argument("--repo", default=".", help="Path to target repository")
    parser.add_argument("--tasks", default="eval_tasks.json", help="Path to evaluation tasks JSON")
    parser.add_argument("--mcp-bin", default="bin/garuda-mcp", help="Path to garuda-mcp executable")
    parser.add_argument("--model", default="claude-3-5-sonnet-20241022", help="Anthropic model name")
    parser.add_argument("--output", default="benchmark_results.json", help="Output results path")
    args = parser.parse_args()

    api_key = os.environ.get("ANTHROPIC_API_KEY")
    if not api_key:
        print("Error: ANTHROPIC_API_KEY environment variable is required.", file=sys.stderr)
        sys.exit(1)

    client = anthropic.Anthropic(api_key=api_key)
    mcp = GarudaMCPClient(args.mcp_bin, args.repo)

    with open(args.tasks, "r") as f:
        tasks = json.load(f)

    print(f"\n==================================================================")
    print(f"GARUDA SCIENTIFIC EVALUATION HARNESS: {len(tasks)} RIGOROUS TASKS")
    print(f"Target Repo: {os.path.abspath(args.repo)}")
    print(f"Model:       {args.model}")
    print(f"==================================================================\n")

    summary = []

    for task in tasks:
        print(f"[*] Running {task['id']}: {task['name']}...")
        gt_set = evaluate_ground_truth(mcp, task)

        # Control Arm (Without Garuda)
        ctrl = run_control_arm(client, args.model, task["prompt"])
        c_p, c_r, c_f1, c_hall = compute_metrics(ctrl["text"], gt_set)

        # Treatment Arm (With Garuda)
        gt_payload = mcp.call_tool(task["ground_truth_tool"], task["ground_truth_args"])
        treat = run_treatment_arm(client, args.model, task["prompt"], gt_payload)
        t_p, t_r, t_f1, t_hall = compute_metrics(treat["text"], gt_set)

        token_savings_pct = round((1.0 - (treat["total_tokens"] / ctrl["total_tokens"])) * 100.0, 1) if ctrl["total_tokens"] > 0 else 0.0

        item = {
            "task_id": task["id"],
            "name": task["name"],
            "ground_truth_count": len(gt_set),
            "control": {
                "tokens": ctrl["total_tokens"],
                "latency_sec": ctrl["latency_sec"],
                "precision": c_p,
                "recall": c_r,
                "f1": c_f1,
                "hallucination_rate": c_hall,
            },
            "treatment": {
                "tokens": treat["total_tokens"],
                "latency_sec": treat["latency_sec"],
                "precision": t_p,
                "recall": t_r,
                "f1": t_f1,
                "hallucination_rate": t_hall,
            },
            "token_savings_pct": token_savings_pct
        }
        summary.append(item)
        print(f"    Control  -> Tokens: {ctrl['total_tokens']:5d} | F1: {c_f1:.2f} | Hallucination: {c_hall*100:4.1f}%")
        print(f"    Garuda   -> Tokens: {treat['total_tokens']:5d} | F1: {t_f1:.2f} | Hallucination: {t_hall*100:4.1f}% | Savings: {token_savings_pct}%")

    mcp.close()

    with open(args.output, "w") as f:
        json.dump(summary, f, indent=2)

    print(f"\n[OK] Results written to {args.output}")


if __name__ == "__main__":
    main()