#!/usr/bin/env python3
# Copyright 2026 Rohit Mishra
# SPDX-License-Identifier: Apache-2.0
"""
Check that everything a plugin marketplace or MCP directory needs is present, tracked by git,
not hidden by .gitignore, parseable, and consistent with the release version.

    python3 scripts/check_marketplace_manifests.py

The files here once vanished from GitHub because a global `*.json` rule in .gitignore hid them.
This check makes that kind of mistake fail CI.
"""

import json
import os
import re
import subprocess
import sys

ROOT = subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True, check=True).stdout.strip()
os.chdir(ROOT)

REQUIRED = [
    ".claude-plugin/plugin.json",
    ".claude-plugin/marketplace.json",
    ".codex-plugin/plugin.json",
    ".agents/plugins/marketplace.json",
    ".cursor-plugin/plugin.json",
    ".mcp.json",
    "server.json",
    "glama.json",
    "smithery.yaml",
    "action.yml",
    "Dockerfile.mcp",
    "assets/garuda-logo.png",
]

errors = []


def err(msg):
    errors.append(msg)


def sh(*args):
    return subprocess.run(args, capture_output=True, text=True)


def load(path):
    try:
        with open(path, encoding="utf-8") as fh:
            return json.load(fh)
    except (OSError, ValueError) as exc:
        err(f"{path}: cannot parse JSON ({exc})")
        return {}


version = open("version.txt", encoding="utf-8").read().strip().lstrip("v")

# 1. present, tracked, not ignored
for path in REQUIRED:
    if not os.path.exists(path):
        err(f"{path}: missing")
        continue
    if sh("git", "ls-files", "--error-unmatch", path).returncode != 0:
        err(f"{path}: exists but is not tracked by git (git add it)")
    if sh("git", "check-ignore", "--no-index", "-q", path).returncode == 0:
        r = sh("git", "check-ignore", "--no-index", "-v", path).stdout.strip()
        err(f"{path}: matched by an ignore rule -> {r}")

# 2. no global rule that would hide manifests again
gi = open(".gitignore", encoding="utf-8").read().splitlines()
for line in gi:
    if re.fullmatch(r"\s*(\*\.json|\*\.html|\*\.txt|\*\.yml|\*\.yaml|\*\.md)\s*", line):
        err(f".gitignore: global rule {line.strip()!r} hides files from GitHub; list specific paths instead")

# 3. versions agree
for path in (".claude-plugin/plugin.json", ".codex-plugin/plugin.json", ".cursor-plugin/plugin.json", "server.json"):
    if os.path.exists(path):
        v = load(path).get("version")
        if v != version:
            err(f"{path}: version {v!r} != version.txt {version!r}")
mk = load(".claude-plugin/marketplace.json")
for plugin in mk.get("plugins", []):
    if plugin.get("version") != version:
        err(f".claude-plugin/marketplace.json: plugin {plugin.get('name')} version {plugin.get('version')!r} != {version!r}")
    src = plugin.get("source")
    if isinstance(src, str) and not os.path.exists(os.path.join(src)):
        err(f".claude-plugin/marketplace.json: plugin source {src!r} does not exist")

# 4. MCP registry entry
srv = load("server.json")
name = srv.get("name", "")
if not name.startswith("io.github."):
    err(f"server.json: name {name!r} must start with io.github.<owner>/")
pk = (srv.get("packages") or [{}])[0]
if not str(pk.get("identifier", "")).endswith(":" + version):
    err(f"server.json: image identifier {pk.get('identifier')!r} must end with :{version}")
label = re.search(r'LABEL\s+io\.modelcontextprotocol\.server\.name="([^"]+)"', open("Dockerfile.mcp", encoding="utf-8").read())
if not label or label.group(1) != name:
    err(f"Dockerfile.mcp: LABEL io.modelcontextprotocol.server.name must equal server.json name {name!r}")

# 5. Codex plugin pointers and skills
codex = load(".codex-plugin/plugin.json")
for key in ("skills", "mcpServers"):
    target = codex.get(key)
    if target and not os.path.exists(target):
        err(f".codex-plugin/plugin.json: {key} -> {target!r} does not exist")
logo = (codex.get("interface") or {}).get("logo")
if logo and not os.path.exists(logo):
    err(f".codex-plugin/plugin.json: logo {logo!r} does not exist")

skills_dir = "skills"
count = 0
for entry in sorted(os.listdir(skills_dir)):
    md = os.path.join(skills_dir, entry, "SKILL.md")
    if not os.path.isfile(md):
        continue
    count += 1
    text = open(md, encoding="utf-8").read()
    m = re.match(r"^---\n(.*?)\n---\n", text, re.S)
    front = m.group(1) if m else ""
    if not re.search(r"^name:\s*\S", front, re.M) or not re.search(r"^description:\s*\S", front, re.M):
        err(f"{md}: needs YAML frontmatter with name and description")
    elif not re.search(rf"^name:\s*{re.escape(entry)}\s*$", front, re.M):
        err(f"{md}: frontmatter name must equal the folder name {entry!r}")
    else:
        dm = re.search(r'^description:\s*"?(.*?)"?\s*$', front, re.M)
        words = len(dm.group(1).split()) if dm else 0
        if not 60 <= words <= 120:
            err(f"{md}: description must be 60-120 words so agents can decide when to use the skill (found {words})")
    if sh("git", "ls-files", "--error-unmatch", md).returncode != 0:
        err(f"{md}: not tracked by git")
if count == 0:
    err("skills/: no SKILL.md files found")

# 6. .mcp.json
mcp = load(".mcp.json")
if "garuda" not in (mcp.get("mcpServers") or {}):
    err(".mcp.json: must define mcpServers.garuda")

if errors:
    for e in errors:
        print("ERROR  " + e)
    print(f"\n{len(errors)} problem(s).")
    sys.exit(1)
print(f"OK  {len(REQUIRED)} marketplace/registry files tracked, visible and consistent at v{version}; {count} skills valid.")
