#!/usr/bin/env python3
# Copyright 2026 Rohit Mishra
# SPDX-License-Identifier: Apache-2.0

import json
import yaml
import argparse
from datetime import datetime

STATUS_BADGES = {
    "production": "🟢 **Production (GA)**",
    "beta": "🟡 **Beta (AST Verified)**",
    "experimental": "🟣 **Experimental (Heuristic)**",
    "planned": "⚪ **Planned (Roadmap)**"
}

def generate_markdown(caps_data, snapshot_data):
    md = []
    md.append("# 🧠 Garuda Capabilities & AST Verification Matrix\n")
    md.append(f"> Auto-generated on `{datetime.utcnow().strftime('%Y-%m-%d %H:%M:%S UTC')}`. Grounded in AST snapshot and benchmark gates.\n")
    
    if snapshot_data and "stats" in snapshot_data:
        stats = snapshot_data["stats"]
        md.append("## Snapshot Extraction Metrics\n")
        md.append("| Metric | Count |")
        md.append("| :--- | :--- |")
        md.append(f"| **Parsed Files** | `{stats.get('files', 0)}` |")
        md.append(f"| **Packages** | `{stats.get('packages', 0)}` |")
        md.append(f"| **Discovered Structs** | `{stats.get('structs', 0)}` |")
        md.append(f"| **Discovered Interfaces** | `{stats.get('interfaces', 0)}` |")
        md.append(f"| **Functions & Methods** | `{stats.get('functions', 0)}` |")
        md.append(f"| **Total Struct Fields** | `{stats.get('total_fields', 0)}` |\n")

    md.append("## Feature Verification & Status\n")
    for category in caps_data.get("categories", []):
        md.append(f"### {category['name']}\n")
        if category.get("description"):
            md.append(f"*{category['description']}*\n")
        md.append("| Capability | Status | Verification Tier | Supported Semantics | Invariant |")
        md.append("| :--- | :--- | :--- | :--- | :--- |")
        
        for cap in category.get("capabilities", []):
            badge = STATUS_BADGES.get(cap.get("status", "planned"), cap.get("status", ""))
            md.append(
                f"| **{cap['name']}** | {badge} | `{cap.get('verification_tier', 'N/A')}` | {cap.get('supported_semantics', '')} | `{cap.get('invariant', '')}` |"
            )
        md.append("")

    return "\n".join(md)

def main():
    parser = argparse.ArgumentParser(description="Generate CAPABILITIES.md")
    parser.add_argument("--capabilities", required=True, help="Path to capabilities.yaml")
    parser.add_argument("--snapshot", required=False, help="Path to garuda-snapshot.json")
    parser.add_argument("--output", required=True, help="Path to output markdown file")
    args = parser.parse_args()

    with open(args.capabilities, "r") as f:
        caps_data = yaml.safe_load(f)

    snapshot_data = None
    if args.snapshot:
        try:
            with open(args.snapshot, "r") as f:
                snapshot_data = json.load(f)
        except Exception:
            snapshot_data = None

    content = generate_markdown(caps_data, snapshot_data)
    with open(args.output, "w") as f:
        f.write(content)

    print(f"✅ Generated {args.output} successfully.")

if __name__ == "__main__":
    main()
