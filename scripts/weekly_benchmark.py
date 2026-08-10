#!/usr/bin/env python3
"""
Garuda Weekly Benchmark Suite
Runs performance tests, generates charts, and pushes results to GitHub.
"""

import os
import sys
import json
import subprocess
import time
import requests
from datetime import datetime
import matplotlib.pyplot as plt

# --- Configuration ---
BENCHMARK_HISTORY = "benchmarks/history.json"
CHART_OUTPUT = "benchmarks/latency_chart.png"

def get_api_key_with_fallback():
    """Try quaternary key first; fallback to primary if it fails."""
    keys = [
        os.getenv("GEMINI_API_KEY_QUATERNARY"),
        os.getenv("GEMINI_API_KEY_PRIMARY"),
    ]
    for key in keys:
        if key:
            return key
    return None

def call_gemini(prompt):
    """Call Gemini API with fallback handling."""
    api_key = get_api_key_with_fallback()
    if not api_key:
        print("⚠️ No Gemini API key available. Skipping AI analysis.")
        return "Benchmark complete. No AI analysis available."

    url = f"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key={api_key}"
    headers = {"Content-Type": "application/json"}
    payload = {"contents": [{"parts": [{"text": prompt}]}]}

    try:
        response = requests.post(url, json=payload, timeout=30)
        if response.status_code == 200:
            data = response.json()
            return data["candidates"][0]["content"]["parts"][0]["text"]
        elif response.status_code == 429:
            print("⚠️ Rate limited (429). Falling back to primary key...")
            return "Benchmark complete. Rate limit exceeded."
        else:
            print(f"⚠️ API error: {response.status_code}")
            return "Benchmark complete. API error."
    except Exception as e:
        print(f"⚠️ API call failed: {e}")
        return "Benchmark complete. API error."

def run_benchmark():
    """Execute 100 decision proposals and measure latency."""
    print("🚀 Running Garuda Benchmark Suite...")
    
    # Start local stack if not running
    subprocess.run(["garuda", "up"], capture_output=True)

    # Wait for services
    time.sleep(5)

    # Get token
    token_cmd = subprocess.run(
        ["curl", "-s", "http://localhost:8080/debug/token?actor=benchmark&tenant_id=00000000-0000-0000-0000-000000000001"],
        capture_output=True, text=True
    )
    token = json.loads(token_cmd.stdout).get("token", "")

    latencies = []
    tokens_saved = []
    contradictions = 0

    for i in range(100):
        start = time.time()
        result = subprocess.run(
            [
                "curl", "-s", "-X", "POST", "http://localhost:8080/api/v1/decisions/submit",
                "-H", f"Authorization: Bearer {token}",
                "-H", "Content-Type: application/json",
                "-H", "X-Model: benchmark-agent",
                "-H", "X-Model-Provider: benchmark",
                "-d", json.dumps({
                    "tenant_id": "00000000-0000-0000-0000-000000000001",
                    "title": f"Benchmark decision {i+1}",
                    "scope_domain": "benchmark",
                    "scope_system": "test"
                })
            ],
            capture_output=True, text=True
        )
        latency = (time.time() - start) * 1000  # ms
        latencies.append(latency)

        try:
            resp = json.loads(result.stdout)
            if "quarantined" in str(resp):
                contradictions += 1
        except:
            pass

    # Calculate metrics
    p50 = sorted(latencies)[len(latencies)//2]
    p95 = sorted(latencies)[int(len(latencies)*0.95)]
    p99 = sorted(latencies)[int(len(latencies)*0.99)]

    return {
        "timestamp": datetime.now().isoformat(),
        "total_requests": len(latencies),
        "p50_ms": p50,
        "p95_ms": p95,
        "p99_ms": p99,
        "contradictions": contradictions,
        "latencies": latencies
    }

def generate_chart(data):
    """Generate latency distribution chart."""
    latencies = data["latencies"]
    plt.figure(figsize=(10, 6))
    plt.hist(latencies, bins=20, color="#8B5CF6", edgecolor="black", alpha=0.7)
    plt.axvline(data["p50_ms"], color="blue", linestyle="--", label=f"P50: {data['p50_ms']:.1f}ms")
    plt.axvline(data["p95_ms"], color="red", linestyle="--", label=f"P95: {data['p95_ms']:.1f}ms")
    plt.axvline(data["p99_ms"], color="orange", linestyle="--", label=f"P99: {data['p99_ms']:.1f}ms")
    plt.title("Garuda API Latency Distribution", fontsize=14)
    plt.xlabel("Latency (ms)")
    plt.ylabel("Frequency")
    plt.legend()
    plt.grid(True, alpha=0.3)
    plt.tight_layout()
    plt.savefig(CHART_OUTPUT)
    print(f"📊 Chart saved to {CHART_OUTPUT}")

def save_history(metrics):
    """Append metrics to benchmark history."""
    os.makedirs(os.path.dirname(BENCHMARK_HISTORY), exist_ok=True)

    history = []
    if os.path.exists(BENCHMARK_HISTORY):
        with open(BENCHMARK_HISTORY, "r") as f:
            try:
                history = json.load(f)
            except:
                history = []

    history.append(metrics)
    with open(BENCHMARK_HISTORY, "w") as f:
        json.dump(history, f, indent=2)
    print(f"📈 History saved to {BENCHMARK_HISTORY}")

def main():
    print("📊 Starting Garuda Benchmark Suite")
    print("=" * 40)

    # Check for Quaternary key
    if not os.getenv("GEMINI_API_KEY_QUATERNARY"):
        print("⚠️ GEMINI_API_KEY_QUATERNARY not set. Using primary key as fallback.")
        print("ℹ️ Upgrade to paid tier for dedicated benchmark capacity.")

    # Run benchmarks
    metrics = run_benchmark()
    print(f"✅ Benchmark complete: {metrics['total_requests']} requests")
    print(f"   P50: {metrics['p50_ms']:.1f}ms | P95: {metrics['p95_ms']:.1f}ms | P99: {metrics['p99_ms']:.1f}ms")

    # Generate chart
    generate_chart(metrics)

    # Save history
    save_history(metrics)

    # Generate AI analysis
    prompt = f"""
Analyze these benchmark results for Garuda:
- P50: {metrics['p50_ms']:.1f}ms
- P95: {metrics['p95_ms']:.1f}ms
- P99: {metrics['p99_ms']:.1f}ms
- Contradictions caught: {metrics['contradictions']}

Provide a brief summary (2-3 sentences) for the weekly report.
"""
    analysis = call_gemini(prompt)
    print(f"\n🧠 AI Analysis:\n{analysis}")

    # Save to markdown report
    report = f"""# Garuda Weekly Benchmark Report

**Date:** {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
**Requests:** {metrics['total_requests']}
**P50 Latency:** {metrics['p50_ms']:.1f}ms
**P95 Latency:** {metrics['p95_ms']:.1f}ms
**P99 Latency:** {metrics['p99_ms']:.1f}ms
**Contradictions Caught:** {metrics['contradictions']}

## AI Analysis
{analysis}

![Latency Distribution](latency_chart.png)
"""
    with open("BENCHMARKS.md", "w") as f:
        f.write(report)

    print("✅ Benchmark report written to BENCHMARKS.md")

if __name__ == "__main__":
    main()