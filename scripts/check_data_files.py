#!/usr/bin/env python3
# Copyright 2026 Rohit Mishra
# SPDX-License-Identifier: Apache-2.0
"""Fail if a tracked .json, .yml or .yaml file does not parse (needs PyYAML).

    python3 scripts/check_data_files.py

Catches things like an unquoted "default: 7 days" inside a YAML description, which broke
the embedded OpenAPI document that the API serves to Swagger UI.
"""

import json
import subprocess
import sys

try:
    import yaml
except ImportError:
    sys.exit("PyYAML is required: pip install pyyaml")

root = subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True, check=True).stdout.strip()
files = subprocess.run(["git", "-C", root, "ls-files"], capture_output=True, text=True, check=True).stdout.splitlines()
bad = 0
checked = 0
for rel in files:
    path = f"{root}/{rel}"
    try:
        if rel.endswith(".json"):
            with open(path, encoding="utf-8") as fh:
                json.load(fh)
        elif rel.endswith((".yml", ".yaml")):
            with open(path, encoding="utf-8") as fh:
                list(yaml.safe_load_all(fh))
        else:
            continue
        checked += 1
    except FileNotFoundError:
        continue
    except Exception as exc:  # noqa: BLE001
        bad += 1
        print(f"ERROR  {rel}: {str(exc).splitlines()[0]}")
if bad:
    print(f"\n{bad} file(s) do not parse.")
    sys.exit(1)
print(f"OK  {checked} JSON/YAML files parse.")
