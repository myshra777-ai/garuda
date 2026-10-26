#!/usr/bin/env python3
"""
Deterministically build the documentation context from:
- garuda-snapshot.json (from 'garuda analyze --output')
- source YAML files (product, capabilities, roadmap)
- git metadata (commit, branch)
"""

import argparse
import json
import os
import subprocess
import sys
from datetime import datetime, timezone

import yaml

def load_yaml(path):
    with open(path, 'r') as f:
        return yaml.safe_load(f)

def get_git_info():
    try:
        commit = subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip()
        branch = subprocess.check_output(['git', 'branch', '--show-current'], text=True).strip()
        return commit, branch
    except Exception:
        return "unknown", "unknown"

def build_context(snapshot_path, product_path, capabilities_path, roadmap_path):
    snapshot = {}
    if os.path.exists(snapshot_path):
        with open(snapshot_path, 'r') as f:
            snapshot = json.load(f)

    product = load_yaml(product_path)
    capabilities = load_yaml(capabilities_path)
    roadmap = load_yaml(roadmap_path)

    commit, branch = get_git_info()

    # Extract CLI commands from snapshot if present
    cli_commands = []
    # If you have a separate source for CLI commands, you can supply it.
    # For now, we'll parse from capabilities that have 'command' field.
    for cap in capabilities.get('capabilities', {}).values():
        if 'command' in cap:
            cli_commands.append(cap['command'])

    # Classify capabilities by status
    stable = [k for k, v in capabilities.get('capabilities', {}).items() if v.get('status') == 'stable']
    beta = [k for k, v in capabilities.get('capabilities', {}).items() if v.get('status') == 'beta']
    experimental = [k for k, v in capabilities.get('capabilities', {}).items() if v.get('status') == 'experimental']
    planned = [k for k, v in capabilities.get('capabilities', {}).items() if v.get('status') == 'planned']
    deferred = [k for k, v in capabilities.get('capabilities', {}).items() if v.get('status') == 'deferred']

    # Determine current phase from roadmap
    current_phase_id = roadmap.get('roadmap', {}).get('current_phase', 'P2')
    phases = roadmap.get('roadmap', {}).get('phases', {})
    current_phase_info = phases.get(current_phase_id, {})
    next_phase_id = None
    phase_keys = list(phases.keys())
    if current_phase_id in phase_keys:
        idx = phase_keys.index(current_phase_id)
        if idx + 1 < len(phase_keys):
            next_phase_id = phase_keys[idx+1]

    # Aggregate semantic features from snapshot (very basic)
    entities = snapshot.get('entities', [])
    claims = snapshot.get('relationships', [])
    stats = snapshot.get('stats', {})

    context = {
        "schema_version": "1.0",
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "source": {
            "repository": product.get('product', {}).get('repository', {}).get('github', 'unknown'),
            "commit": commit,
            "branch": branch,
            "analyzer_version": "0.1.0",  # could derive from snapshot if present
            "docs_contract_version": "1.0"
        },
        "product": {
            "name": product.get('product', {}).get('name', 'Garuda'),
            "tagline": product.get('product', {}).get('tagline', ''),
            "category": product.get('product', {}).get('category', {}).get('primary', ''),
            "primary_language": product.get('product', {}).get('repository', {}).get('primary_language', 'Go'),
            "language_scope": product.get('product', {}).get('repository', {}).get('current_language_scope', ['Go']),
            "positioning": product.get('product', {}).get('positioning', {}),
            "thesis": product.get('product', {}).get('thesis', ''),
            "immediate_mission": product.get('product', {}).get('immediate_mission', '')
        },
        "project_state": {
            "current_phase": {
                "id": current_phase_id,
                "name": current_phase_info.get('name', ''),
                "status": current_phase_info.get('status', '')
            },
            "repository": {
                "files": stats.get('files', 0),
                "packages": stats.get('packages', 0),
                "entities": stats.get('structs', 0) + stats.get('interfaces', 0) + stats.get('functions', 0),
                "functions": stats.get('functions', 0),
                "types": stats.get('structs', 0),
                "interfaces": stats.get('interfaces', 0),
                "apis": 0   # placeholder
            }
        },
        "capabilities": {
            "stable": stable,
            "beta": beta,
            "experimental": experimental,
            "planned": planned,
            "deferred": deferred
        },
        "cli": {
            "commands": cli_commands
        },
        "architecture": {
            "semantic_model": [
                "entities",
                "claims",
                "evidence",
                "lineage"
            ],
            "epistemic_classes": [
                "OBSERVED_CODE",
                "OBSERVED_RUNTIME",
                "OBSERVED_CONFIG",
                "OBSERVED_DOC",
                "INFERRED",
                "CONFLICTED",
                "DECISION",
                "POLICY",
                "VERIFIED"
            ]
        },
        "semantic_features": {
            "immutable_ledger": "stable" in stable or "semantic_graph" in stable,
            "merkle_verification": "merkle_verification" in stable,
            "semantic_graph": "semantic_graph" in stable,
            "claims": True,   # hardcoded
            "evidence": True,
            "lineage": True,
            "cross_repo": "cross_repo" in planned,
            "runtime_reconciliation": "runtime_reconciliation" in planned,
            "policy_governance": "policy_governance" in planned
        },
        "benchmark": {
            "available": False,
            "metrics": {}
        },
        "documentation": {
            "sections": {},
            "warnings": [],
            "unsupported_claims": []
        },
        "roadmap": {
            "current_phase": current_phase_id,
            "next_phase": next_phase_id
        }
    }

    return context

def main():
    parser = argparse.ArgumentParser(description="Build documentation context")
    parser.add_argument("--snapshot", required=True, help="Path to garuda-snapshot.json")
    parser.add_argument("--product", required=True, help="Path to product.yaml")
    parser.add_argument("--capabilities", required=True, help="Path to capabilities.yaml")
    parser.add_argument("--roadmap", required=True, help="Path to roadmap.yaml")
    parser.add_argument("--output", required=True, help="Output JSON file")
    args = parser.parse_args()

    context = build_context(args.snapshot, args.product, args.capabilities, args.roadmap)
    with open(args.output, 'w') as f:
        json.dump(context, f, indent=2)

if __name__ == "__main__":
    main()