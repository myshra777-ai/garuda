#!/usr/bin/env python3
"""
Zero-Dependency Ollama Benchmark Harness.
Talks directly to the Ollama REST API using Python stdlib.
"""

import argparse
import json
import os
import subprocess
import sys
import time
import re
import urllib.request
from typing import Any, Dict, List, Set, Tuple

class GarudaMCPClient:
    def __init__(self, binary_path: str, repo_dir: str):
        env = dict(os.environ)
        env["GARUDA_MCP_QUIET"] = "1"
        self.proc = subprocess.Popen([binary_path], stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL, cwd=repo_dir, env=env)
        self.msg_id = 0
        self.call("initialize", {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "ollama-bench", "version": "1.0"}})

    def call(self, method: str, params: Dict[str, Any]) -> Dict[str, Any]:
        self.msg_id += 1
        req = {"jsonrpc": "2.0", "id": self.msg_id, "method": method, "params": params}
        self.proc.stdin.write((json.dumps(req) + "\n").encode("utf-8"))
        self.proc.stdin.flush()
        raw = self.proc.stdout.readline()
        return json.loads(raw.decode("utf-8")).get("result", {}) if raw else {}

    def call_tool(self, name: str, arguments: Dict[str, Any]) -> Dict[str, Any]:
        result = self.call("tools/call", {"name": name, "arguments": arguments})
        content = result.get("content", [])
        if content and "text" in content[0]:
            try: return json.loads(content[0]["text"])
            except json.JSONDecodeError: return {"raw_text": content[0]["text"]}
        return result

    def close(self):
        self.proc.terminate()

def extract_json_array(text: str) -> List[str]:
    try:
        match = re.search(r'\[.*?\]', text, re.DOTALL)
        if match:
            parsed = json.loads(match.group(0))
            if isinstance(parsed, list): return [str(x).strip() for x in parsed]
    except Exception: pass
    return []

def evaluate_ground_truth(mcp: GarudaMCPClient, task: Dict[str, Any]) -> Set[str]:
    res = mcp.call_tool(task["ground_truth_tool"], task["ground_truth_args"])
    target_key = task.get("target_key")
    match_field = task.get("match_field")
    gt_set = set()
    if target_key and target_key in res:
        items = res[target_key]
        if isinstance(items, list):
            for it in items:
                gt_set.add(str(it[match_field]).strip() if isinstance(it, dict) and match_field in it else str(it).strip())
        else: gt_set.add(str(items).strip())
    return gt_set

def run_ollama(model: str, prompt: str, system_prompt: str) -> Dict[str, Any]:
    """Talks directly to local Ollama REST API."""
    t0 = time.perf_counter()
    data = {
        "model": model,
        "messages": [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": prompt}
        ],
        "stream": False
    }
    req = urllib.request.Request(
        "http://localhost:11434/api/chat",
        data=json.dumps(data).encode("utf-8"),
        headers={"Content-Type": "application/json"}
    )
    
    try:
        with urllib.request.urlopen(req) as resp:
            result = json.loads(resp.read().decode("utf-8"))
            duration = time.perf_counter() - t0
            # Ollama provides exact token counts in the response
            prompt_tokens = result.get("prompt_eval_count", 0)
            eval_tokens = result.get("eval_count", 0)
            return {
                "text": result.get("message", {}).get("content", ""),
                "total_tokens": prompt_tokens + eval_tokens,
                "latency_sec": round(duration, 3)
            }
    except urllib.error.URLError as e:
        print(f"Error connecting to Ollama: {e}. Is 'ollama serve' running?", file=sys.stderr)
        sys.exit(1)

def compute_metrics(predicted_text: str, ground_truth: Set[str]) -> Tuple[float, float, float, float]:
    if not ground_truth: return 1.0, 1.0, 1.0, 0.0
    predicted_set = set(extract_json_array(predicted_text))
    tp = len(predicted_set.intersection(ground_truth))
    fp = len(predicted_set - ground_truth)
    fn = len(ground_truth - predicted_set)
    precision = tp / (tp + fp) if (tp + fp) > 0 else 0.0
    recall = tp / (tp + fn) if (tp + fn) > 0 else 0.0
    f1 = (2 * precision * recall) / (precision + recall) if (precision + recall) > 0 else 0.0
    hallucination_rate = fp / len(predicted_set) if len(predicted_set) > 0 else 0.0
    return round(precision, 3), round(recall, 3), round(f1, 3), round(hallucination_rate, 3)

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo", default=".")
    parser.add_argument("--tasks", default="eval_tasks.json")
    parser.add_argument("--mcp-bin", default="bin/garuda-mcp")
    parser.add_argument("--model", default="llama3.1")
    parser.add_argument("--output", default="ollama_benchmark_results.json")
    args = parser.parse_args()

    mcp = GarudaMCPClient(args.mcp_bin, args.repo)

    with open(args.tasks, "r") as f: tasks = json.load(f)

    print(f"\n=== GARUDA BENCHMARK (OLLAMA {args.model}) ===\n")
    summary = []

    for task in tasks:
        print(f"[*] {task['id']}: {task['name']}")
        gt_set = evaluate_ground_truth(mcp, task)

        ctrl = run_ollama(args.model, task["prompt"], "Respond ONLY with a valid JSON array of strings. No markdown.")
        c_f1, c_hall = compute_metrics(ctrl["text"], gt_set)[2], compute_metrics(ctrl["text"], gt_set)[3]

        gt_payload = mcp.call_tool(task["ground_truth_tool"], task["ground_truth_args"])
        sys_prompt = f"Use this verified data:\n{json.dumps(gt_payload)}\nRespond ONLY with a valid JSON array of strings. No markdown."
        treat = run_ollama(args.model, task["prompt"], sys_prompt)
        t_f1, t_hall = compute_metrics(treat["text"], gt_set)[2], compute_metrics(treat["text"], gt_set)[3]

        savings = round((1.0 - (treat["total_tokens"] / ctrl["total_tokens"])) * 100.0, 1) if ctrl["total_tokens"] > 0 else 0.0

        summary.append({"task_id": task["id"], "control": {"tokens": ctrl["total_tokens"], "f1": c_f1, "hallucination": c_hall}, "treatment": {"tokens": treat["total_tokens"], "f1": t_f1, "hallucination": t_hall}, "savings_pct": savings})
        print(f"    Control -> Tokens: {ctrl['total_tokens']:4d} | F1: {c_f1:.2f} | Hallucination: {c_hall*100:4.1f}%")
        print(f"    Garuda  -> Tokens: {treat['total_tokens']:4d} | F1: {t_f1:.2f} | Hallucination: {t_hall*100:4.1f}% | Savings: {savings}%")

    mcp.close()
    with open(args.output, "w") as f: json.dump(summary, f, indent=2)
    print(f"\n[OK] Results written to {args.output}")

if __name__ == "__main__": main()
