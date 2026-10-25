#!/usr/bin/env python3
"""
Build a deterministic documentation context from the current product,
capability and roadmap metadata plus an optional Garuda analyzer snapshot.

The source capability schema is category-based (docs/source/capabilities.yaml),
not the legacy name->capability map this script originally expected.
"""

import argparse
import json
import os
import subprocess
from datetime import datetime, timezone

import yaml


IMPLEMENTED_STATUSES = {"production", "stable", "implemented"}
KNOWN_STATUSES = {
    "production",
    "stable",
    "implemented",
    "implemented-with-ongoing-validation",
    "beta",
    "experimental",
    "planned",
    "deferred",
}


def load_yaml(path):
    with open(path, "r", encoding="utf-8") as f:
        value = yaml.safe_load(f)
    return value or {}


def get_git_info():
    try:
        commit = subprocess.check_output(
            ["git", "rev-parse", "HEAD"], text=True
        ).strip()
        branch = subprocess.check_output(
            ["git", "branch", "--show-current"], text=True
        ).strip()
        return commit, branch
    except (OSError, subprocess.CalledProcessError):
        return "unknown", "unknown"


def flatten_capabilities(capability_doc):
    """Accept the current category-list schema and the former map schema."""
    rows = []
    categories = capability_doc.get("categories", [])
    if isinstance(categories, list):
        for category in categories:
            if not isinstance(category, dict):
                continue
            category_name = category.get("name", "Uncategorized")
            for cap in category.get("capabilities", []) or []:
                if not isinstance(cap, dict):
                    continue
                rows.append({
                    **cap,
                    "category": category_name,
                    "name": cap.get("name", "Unnamed capability"),
                })
    else:
        # Be tolerant of older metadata during a migration window.
        for name, cap in (capability_doc.get("capabilities", {}) or {}).items():
            if isinstance(cap, dict):
                rows.append({**cap, "name": cap.get("name", name)})
    return rows


def build_context(snapshot_path, product_path, capabilities_path, roadmap_path):
    snapshot = {}
    if os.path.exists(snapshot_path):
        with open(snapshot_path, "r", encoding="utf-8") as f:
            snapshot = json.load(f)

    product_doc = load_yaml(product_path)
    product = product_doc.get("product", product_doc)
    capability_doc = load_yaml(capabilities_path)
    roadmap_doc = load_yaml(roadmap_path)
    capability_rows = flatten_capabilities(capability_doc)

    commit, branch = get_git_info()
    commands = sorted({
        cap["command"]
        for cap in capability_rows
        if isinstance(cap.get("command"), str) and cap["command"].strip()
    })

    grouped = {status: [] for status in KNOWN_STATUSES}
    for cap in capability_rows:
        status = cap.get("status", "unknown")
        label = cap["name"]
        if status in grouped:
            grouped[status].append(label)
        else:
            grouped.setdefault("unclassified", []).append(label)

    # Keep the context keys consumed by the current doc-generation scripts,
    # while deriving the values from the actual category-list source metadata.
    stable = sorted(
        grouped["production"] + grouped["stable"] + grouped["implemented"]
    )
    beta = sorted(grouped["beta"])
    experimental = sorted(grouped["experimental"])
    planned = sorted(grouped["planned"])
    deferred = sorted(grouped["deferred"])

    roadmap = roadmap_doc.get("roadmap", {})
    current_phase_id = roadmap.get("current_phase", "unknown")
    phases = roadmap.get("phases", {}) or {}
    current_phase_info = phases.get(current_phase_id, {}) or {}
    phase_keys = list(phases.keys())
    next_phase_id = None
    if current_phase_id in phase_keys:
        idx = phase_keys.index(current_phase_id)
        if idx + 1 < len(phase_keys):
            next_phase_id = phase_keys[idx + 1]

    stats = snapshot.get("stats", {}) or {}
    entity_count = (
        stats.get("entities")
        if stats.get("entities") is not None
        else stats.get("structs", 0)
        + stats.get("interfaces", 0)
        + stats.get("functions", 0)
    )

    def has_capability(*terms):
        terms = tuple(t.lower() for t in terms)
        return any(
            all(term in f"{cap.get('name', '')} {cap.get('category', '')}".lower()
                for term in terms)
            for cap in capability_rows
        )

    repository = product.get("repository", {}) or {}
    repo_url = repository.get("github", "https://github.com/myshra777-ai/garuda")
    language_scope = repository.get(
        "current_language_scope", ["Go", "Python", "TypeScript"]
    )

    return {
        "schema_version": "1.1",
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "source": {
            "repository": repo_url,
            "commit": commit,
            "branch": branch,
            "product_version": product.get("version", "unknown"),
            "docs_contract_version": product.get("docs_contract_version", "unknown"),
        },
        "product": {
            "name": product.get("display_name", product.get("name", "Garuda")),
            "tagline": product.get("tagline", ""),
            "category": product.get("category", {}).get("primary", "")
                if isinstance(product.get("category", {}), dict)
                else product.get("category", ""),
            "description": product.get("short_description", ""),
            "long_description": product.get("long_description", ""),
            "primary_language": repository.get("primary_language", "Go"),
            "language_scope": language_scope,
            "positioning": product.get("positioning", {}),
            "thesis": product.get("thesis", ""),
            "immediate_mission": product.get("immediate_mission", ""),
            "audiences": product.get("audiences", {}),
            "differentiation": product.get("positioning", {}).get("differentiation", []),
        },
        "project_state": {
            "current_phase": {
                "id": current_phase_id,
                "name": current_phase_info.get("name", ""),
                "status": current_phase_info.get("status", ""),
            },
            "repository": {
                "files": stats.get("files"),
                "packages": stats.get("packages"),
                "entities": entity_count,
                "functions": stats.get("functions"),
                "types": stats.get("structs"),
                "interfaces": stats.get("interfaces"),
                # None means this snapshot did not measure this metric.
                "apis": stats.get("apis"),
            },
        },
        "capabilities": {
            "stable": stable,
            "beta": beta,
            "experimental": experimental,
            "planned": planned,
            "deferred": deferred,
            "implemented_with_ongoing_validation": sorted(
                grouped["implemented-with-ongoing-validation"]
            ),
            "unclassified": sorted(grouped.get("unclassified", [])),
            "all": capability_rows,
        },
        "cli": {"commands": commands},
        "architecture": {
            "semantic_model": ["entities", "typed relationships", "evidence", "lineage"],
            "epistemic_classes": ["SUPPORTED", "UNVERIFIED", "CONTRADICTED"],
        },
        "semantic_features": {
            "immutable_ledger": has_capability("Merkle"),
            "merkle_verification": has_capability("Merkle"),
            "semantic_graph": has_capability("Go") or has_capability("graph"),
            "claims": has_capability("claim") or has_capability("documentation"),
            "evidence": has_capability("runtime") or has_capability("evidence"),
            "lineage": has_capability("lineage"),
            "cross_repo": has_capability("cross-repository") or has_capability("cross-repo"),
            "runtime_reconciliation": has_capability("runtime") or has_capability("telemetry"),
            "policy_governance": has_capability("policy"),
            "talon_harness": has_capability("Talon"),
            "owner_control_plane": has_capability("Control Plane"),
        },
        "benchmark": {
            "available": False,
            "metrics": {},
            "source": "EVIDENCE.md; results must be scoped to each named test/corpus",
        },
        "documentation": {
            "sections": {},
            "warnings": [],
            "unsupported_claims": [],
        },
        "roadmap": {
            "current_phase": current_phase_id,
            "next_phase": next_phase_id,
        },
    }


def main():
    parser = argparse.ArgumentParser(description="Build documentation context")
    parser.add_argument("--snapshot", required=True, help="Path to garuda-snapshot.json")
    parser.add_argument("--product", required=True, help="Path to product.yaml")
    parser.add_argument("--capabilities", required=True, help="Path to capabilities.yaml")
    parser.add_argument("--roadmap", required=True, help="Path to roadmap.yaml")
    parser.add_argument("--output", required=True, help="Output JSON file")
    args = parser.parse_args()

    context = build_context(
        args.snapshot, args.product, args.capabilities, args.roadmap
    )
    with open(args.output, "w", encoding="utf-8") as f:
        json.dump(context, f, indent=2, ensure_ascii=False)
        f.write("\n")


if __name__ == "__main__":
    main()
