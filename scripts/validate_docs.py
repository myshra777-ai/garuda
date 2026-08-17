#!/usr/bin/env python3
"""
Validate generated documentation against the context and YAML sources.
"""

import argparse
import json
import os
import sys

def load_context(path):
    with open(path, 'r') as f:
        return json.load(f)

def validate_readme(readme_path, context):
    errors = []
    with open(readme_path, 'r') as f:
        content = f.read()

    # Check that capabilities in README match context
    # This is a basic check; we could parse Markdown more thoroughly.
    stable = context.get('capabilities', {}).get('stable', [])
    planned = context.get('capabilities', {}).get('planned', [])
    # For each stable capability, it should be mentioned in README? Not necessarily.
    # But we can check that no planned capability is described as available.
    for cap in planned:
        # Look for the capability name or its command
        command = cap
        # In capabilities.yaml, we have a command field; we might need to map.
        # For simplicity, we just check if the capability name appears in the README with "planned" or "future".
        # This is a placeholder.
        pass

    # Check that the README contains the correct tagline
    tagline = context.get('product', {}).get('tagline', '')
    if tagline and tagline not in content:
        errors.append("README does not contain the product tagline from product.yaml")

    return errors

def validate_changelog(changelog_path):
    # Ensure the file exists and is non-empty
    if not os.path.exists(changelog_path) or os.path.getsize(changelog_path) == 0:
        return ["CHANGELOG.md is empty or missing"]
    return []

def validate_api(api_path):
    if not os.path.exists(api_path) or os.path.getsize(api_path) == 0:
        return ["API.md is empty or missing"]
    return []

def main():
    parser = argparse.ArgumentParser(description="Validate generated documentation")
    parser.add_argument("--context", required=True, help="Path to docs-context.json")
    parser.add_argument("--readme", default="README.md", help="Path to README.md")
    parser.add_argument("--changelog", default="CHANGELOG.md", help="Path to CHANGELOG.md")
    parser.add_argument("--api", default="docs/API.md", help="Path to API.md")
    args = parser.parse_args()

    context = load_context(args.context)
    all_errors = []
    all_errors.extend(validate_readme(args.readme, context))
    all_errors.extend(validate_changelog(args.changelog))
    all_errors.extend(validate_api(args.api))

    if all_errors:
        print("❌ Documentation validation failed:")
        for e in all_errors:
            print(f"  - {e}")
        sys.exit(1)
    else:
        print("✅ Documentation validation passed.")
        sys.exit(0)

if __name__ == "__main__":
    main()