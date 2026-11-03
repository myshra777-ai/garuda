#!/usr/bin/env python3
# Copyright 2026 Rohit Mishra
# SPDX-License-Identifier: Apache-2.0
"""
Release-consistency checker and synchronizer.

`capabilities.json` is the single source of truth for the release version and
the headline counts. This script makes sure everything else agrees with it.

    python3 scripts/sync_docs_from_manifest.py --check   # CI: report drift, exit 1 if any
    python3 scripts/sync_docs_from_manifest.py           # fix mechanical drift in place
    python3 scripts/sync_docs_from_manifest.py --set v1.9.11 --date 2026-12-01
                                                          # start a new release: rewrite
                                                          # capabilities.json, then sync

What it keeps in sync (see VERSION_TARGETS):
  * version.txt, internal/version/manifest.go, cmd/garuda/root.go
  * the plugin / MCP-registry manifests, Dockerfiles, packaging, OpenAPI info block
  * the MCP tool table in docs/API.md (between the MCP-TOOLS markers)

What it only checks (counts that must match the code):
  * number of MCP tools in cmd/garuda-mcp/main.go, read-only/mutating split
  * number of SQL migrations in migrations/
  * number of skills in skills/, README mentions the tool count
  * release-blocker contracts (Dockerfile builds all 4 binaries, action.yml fails
    closed, build matrix writes to dist/releases)
  * living docs do not mention a different "vX.Y.Z" than the current release

It never rewrites README.md or any prose; those are written by hand.
"""

import argparse
import glob
import json
import os
import re
import subprocess
import sys

SEMVER = r"[0-9]+\.[0-9]+\.[0-9]+"

# (path, [regex with exactly ONE capture group holding the bare X.Y.Z], minimum matches)
VERSION_TARGETS = [
    ("internal/version/manifest.go", [rf'Version\s*=\s*"v({SEMVER})"', rf'SemVer\s*=\s*"({SEMVER})"'], 2),
    ("cmd/garuda/root.go", [rf'var version = "v({SEMVER})"'], 1),
    (".claude-plugin/plugin.json", [rf'"version":\s*"({SEMVER})"'], 1),
    (".claude-plugin/marketplace.json", [rf'"version":\s*"({SEMVER})"'], 1),
    (".codex-plugin/plugin.json", [rf'"version":\s*"({SEMVER})"'], 1),
    (".cursor-plugin/plugin.json", [rf'"version":\s*"({SEMVER})"'], 1),
    ("server.json", [rf'"version":\s*"({SEMVER})"', rf'garuda-mcp:({SEMVER})"'], 2),
    ("skills/architecture-review/SKILL.md", [rf'version:\s*"({SEMVER})"'], 1),
    ("skills/change-impact/SKILL.md", [rf'version:\s*"({SEMVER})"'], 1),
    ("skills/evidence-verification/SKILL.md", [rf'version:\s*"({SEMVER})"'], 1),
    ("skills/post-edit-verification/SKILL.md", [rf'version:\s*"({SEMVER})"'], 1),
    ("skills/pre-edit-governance/SKILL.md", [rf'version:\s*"({SEMVER})"'], 1),
    ("skills/violation-repair/SKILL.md", [rf'version:\s*"({SEMVER})"'], 1),
    ("Dockerfile", [rf"ARG VERSION=v({SEMVER})"], 1),
    ("Dockerfile.mcp", [rf"ARG VERSION=v({SEMVER})"], 1),
    ("action.yml", [rf"Data-Logic Gate \(\\`v({SEMVER})\\`\)"], 1),
    ("internal/api/openapi.yaml", [rf'(?:version|example): "?v?({SEMVER})"?'], 3),
    ("internal/api/system_handlers.go", [rf'"version":\s+"v({SEMVER})"'], 1),
    ("internal/telemetry/collector.go", [rf'Version:\s+"({SEMVER})"'], 1),
    ("internal/telemetry/telemetry.go", [rf'GarudaVersion:\s+"({SEMVER})"'], 1),
    ("deploy/k8s/garuda-stack.yaml", [rf"myshra777-ai/garuda:v({SEMVER})"], 1),
    ("packaging/npm/package.json", [rf'"version":\s*"({SEMVER})"'], 1),
    ("packaging/homebrew/garuda.rb", [rf'version "({SEMVER})"', rf"download/v({SEMVER})/", rf"garuda-v({SEMVER})-", rf'assert_match "({SEMVER})"'], 4),
    ("packaging/aur/PKGBUILD", [rf"pkgver=({SEMVER})"], 1),
    ("packaging/rpm/garuda.spec", [rf"Version:\s+({SEMVER})", rf"Garuda v({SEMVER}) "], 2),
    ("packaging/choco/garuda.nuspec", [rf"<version>({SEMVER})</version>", rf"Garuda v({SEMVER}) ", rf"garuda-v({SEMVER})-"], 3),
    ("docs/GITHUB_MARKETPLACE_STRATEGY.md", [rf"myshra777-ai/garuda@v({SEMVER})", rf"garuda:({SEMVER})", rf"garuda-v({SEMVER})-"], 3),
    ("packaging/winget/garuda.yaml", [rf"PackageVersion:\s+({SEMVER})", rf"garuda-v({SEMVER})-", rf"download/v({SEMVER})/"], 3),
]

# Documents that describe the CURRENT product. They must not name another release.
# (CHANGELOG.md, RELEASE_NOTES.md and the history sections of other docs may.)
LIVING_DOCS = [
    "README.md", "INSTALL.md", "QUICKSTART.md", "PLAYBOOK.md", "SECURITY.md",
    "CONTRIBUTING.md", "docs/API.md", "docs/WALKTHROUGH.md", "docs/POSITIONING.md",
]

# (path, regex with one capture group holding an integer, manifest key path, label)
COUNT_TARGETS = [
    ("internal/version/manifest.go", r"MCPToolCount\s*=\s*([0-9]+)", ("counts", "mcp_tools"), "MCPToolCount"),
    ("internal/version/manifest.go", r"MCPMutatingToolCount\s*=\s*([0-9]+)", ("counts", "mcp_mutating_tools"), "MCPMutatingToolCount"),
    ("internal/version/manifest.go", r"MCPReadOnlyToolCount\s*=\s*([0-9]+)", ("counts", "mcp_readonly_tools"), "MCPReadOnlyToolCount"),
    ("internal/version/manifest.go", r"DashboardSurfaceCount\s*=\s*([0-9]+)", ("counts", "dashboard_surfaces"), "DashboardSurfaceCount"),
    ("internal/version/manifest.go", r"MigrationCount\s*=\s*([0-9]+)", ("counts", "sql_migrations"), "MigrationCount"),
]

MUTATING_TOOLS = {
    "garuda.propose_decision", "garuda.decide_logic_variant", "garuda.reconcile_data_audit",
    "garuda.reconcile_financial_audit", "garuda.approve_base_rulebook", "garuda.synthesize_rulebook",
    "garuda.docs_ingest", "garuda.create_task", "garuda.handoff", "garuda.resume",
    "garuda.reset_audit_state", "garuda.create_workspace", "garuda.talon_verify_edit",
}

TOOLS_BEGIN = "<!-- MCP-TOOLS:BEGIN (generated by scripts/sync_docs_from_manifest.py; do not edit) -->"
TOOLS_END = "<!-- MCP-TOOLS:END -->"


def read(root, rel):
    with open(os.path.join(root, rel), encoding="utf-8") as fh:
        return fh.read()


def write(root, rel, text):
    with open(os.path.join(root, rel), "w", encoding="utf-8", newline="\n") as fh:
        fh.write(text)


def extract_mcp_tools(root):
    main_go = read(root, "cmd/garuda-mcp/main.go")
    return re.findall(r'"name":\s*"(garuda\.[a-z0-9_.]+)",\s*\n\s*"description":\s*"([^"]+)"', main_go)


def count_migrations(root):
    return len(glob.glob(os.path.join(root, "migrations", "*.sql")))


def count_skills(root):
    return len(glob.glob(os.path.join(root, "skills", "*", "SKILL.md")))


def sub_groups(text, patterns, new_value):
    """Replace capture group 1 of every match of every pattern. Returns (text, [old values])."""
    olds = []
    for pat in patterns:
        def repl(m, pat=pat):
            olds.append(m.group(1))
            return m.group(0)[: m.start(1) - m.start(0)] + new_value + m.group(0)[m.end(1) - m.start(0):]
        text = re.sub(pat, repl, text)
    return text, olds


def render_tool_table(counts, tools):
    rows = []
    for idx, (name, desc) in enumerate(tools, 1):
        mode = "Mutating (`GARUDA_MCP_READ_ONLY` gated)" if name in MUTATING_TOOLS else "Read-only"
        rows.append(f"| {idx:02d} | `{name}` | {mode} | {desc} |")
    return (
        f"{TOOLS_BEGIN}\n"
        f"## The {counts['mcp_tools']} MCP tools\n\n"
        f"{counts['mcp_readonly_tools']} are read-only and {counts['mcp_mutating_tools']} can change state "
        "(disabled when `GARUDA_MCP_READ_ONLY=true`).\n\n"
        "| # | Tool | Mode | What it does |\n"
        "| :--- | :--- | :--- | :--- |\n"
        + "\n".join(rows)
        + f"\n{TOOLS_END}\n"
    )


def set_new_release(root, tag, date):
    if not re.fullmatch(r"v" + SEMVER, tag):
        raise SystemExit(f"--set expects a tag like v1.9.11, got {tag!r}")
    path = os.path.join(root, "capabilities.json")
    with open(path, encoding="utf-8") as fh:
        text = fh.read()
    text = re.sub(r'("version":\s*)"v[0-9.]+"', rf'\1"{tag}"', text, count=1)
    text = re.sub(r'("semver":\s*)"[0-9.]+"', rf'\1"{tag[1:]}"', text, count=1)
    if date:
        text = re.sub(r'("release_date":\s*)"[0-9-]+"', rf'\1"{date}"', text, count=1)
    with open(path, "w", encoding="utf-8", newline="\n") as fh:
        fh.write(text)


def run(root, check_only):
    problems = []
    fixed = []

    manifest_path = os.path.join(root, "capabilities.json")
    if not os.path.exists(manifest_path):
        print("capabilities.json is missing (it is the single source of truth).")
        print("If it exists locally but not in git, a global '*.json' rule in .gitignore is hiding it.")
        return 1
    manifest = json.loads(read(root, "capabilities.json"))
    ver, semver, counts = manifest["version"], manifest["semver"], manifest["counts"]
    if ver != "v" + semver:
        problems.append(f"capabilities.json: version {ver!r} and semver {semver!r} disagree")

    # --- counts that must match the code ---------------------------------------------------
    tools = extract_mcp_tools(root)
    if len(tools) != counts["mcp_tools"]:
        problems.append(f"capabilities.json counts.mcp_tools={counts['mcp_tools']} but cmd/garuda-mcp/main.go registers {len(tools)}")
    if counts["mcp_readonly_tools"] + counts["mcp_mutating_tools"] != counts["mcp_tools"]:
        problems.append("capabilities.json: mcp_readonly_tools + mcp_mutating_tools != mcp_tools")
    mutating_found = sum(1 for n, _ in tools if n in MUTATING_TOOLS)
    if mutating_found != counts["mcp_mutating_tools"]:
        problems.append(f"{mutating_found} registered tools are in MUTATING_TOOLS but capabilities.json says {counts['mcp_mutating_tools']}")
    migrations = count_migrations(root)
    if migrations != counts["sql_migrations"]:
        problems.append(f"capabilities.json counts.sql_migrations={counts['sql_migrations']} but migrations/ has {migrations} .sql files")
    skills = count_skills(root)
    if "skills" in counts and skills != counts["skills"]:
        problems.append(f"capabilities.json counts.skills={counts['skills']} but skills/ has {skills}")

    for rel, pat, keypath, label in COUNT_TARGETS:
        text = read(root, rel)
        m = re.search(pat, text)
        want = manifest[keypath[0]][keypath[1]]
        if not m:
            problems.append(f"{rel}: could not find {label}")
        elif int(m.group(1)) != want:
            if check_only:
                problems.append(f"{rel}: {label}={m.group(1)} but capabilities.json says {want}")
            else:
                text = text[: m.start(1)] + str(want) + text[m.end(1):]
                write(root, rel, text)
                fixed.append(f"{rel}: {label} {m.group(1)} -> {want}")

    # --- release date in Go constants ------------------------------------------------------
    rel = "internal/version/manifest.go"
    text = read(root, rel)
    m = re.search(r'ReleaseDate\s*=\s*"([0-9-]+)"', text)
    if m and m.group(1) != manifest["release_date"]:
        if check_only:
            problems.append(f"{rel}: ReleaseDate={m.group(1)} but capabilities.json says {manifest['release_date']}")
        else:
            write(root, rel, text[: m.start(1)] + manifest["release_date"] + text[m.end(1):])
            fixed.append(f"{rel}: ReleaseDate -> {manifest['release_date']}")

    # --- version.txt -----------------------------------------------------------------------
    vtxt = os.path.join(root, "version.txt")
    current = open(vtxt, encoding="utf-8").read().strip() if os.path.exists(vtxt) else None
    if current != ver:
        if check_only:
            problems.append(f"version.txt is {current!r}, expected {ver!r}")
        else:
            write(root, "version.txt", ver + "\n")
            fixed.append(f"version.txt -> {ver}")

    # --- version strings in manifests, Dockerfiles, packaging, Go -------------------------
    for rel, patterns, minimum in VERSION_TARGETS:
        if not os.path.exists(os.path.join(root, rel)):
            problems.append(f"{rel}: file is missing")
            continue
        text = read(root, rel)
        new_text, olds = sub_groups(text, patterns, semver)
        if len(olds) < minimum:
            problems.append(f"{rel}: expected at least {minimum} version strings, found {len(olds)} (pattern drift?)")
            continue
        stale = sorted({o for o in olds if o != semver})
        if stale:
            if check_only:
                problems.append(f"{rel}: has version {', '.join(stale)}; expected {semver}")
            else:
                write(root, rel, new_text)
                fixed.append(f"{rel}: {', '.join(stale)} -> {semver}")

    # --- MCP tool table in docs/API.md -----------------------------------------------------
    rel = "docs/API.md"
    api = read(root, rel)
    table = render_tool_table(counts, tools)
    if TOOLS_BEGIN in api and TOOLS_END in api:
        head, rest = api.split(TOOLS_BEGIN, 1)
        _, tail = rest.split(TOOLS_END, 1)
        wanted = head + table.rstrip("\n") + tail
    else:
        wanted = api.rstrip("\n") + "\n\n" + table
    if wanted != api:
        if check_only:
            problems.append("docs/API.md: the MCP tool table is out of date (run scripts/sync_docs_from_manifest.py)")
        else:
            write(root, rel, wanted)
            fixed.append("docs/API.md: MCP tool table regenerated")

    # --- README facts ------------------------------------------------------------------------
    readme = read(root, "README.md")
    if f"{counts['mcp_tools']} tools" not in readme:
        problems.append(f"README.md does not mention '{counts['mcp_tools']} tools'")
    if f"{skills} skills" not in readme:
        problems.append(f"README.md does not mention '{skills} skills'")

    # --- living docs must not name another release -----------------------------------------
    for rel in LIVING_DOCS:
        if not os.path.exists(os.path.join(root, rel)):
            continue
        others = sorted({v for v in re.findall(r"\bv1\.[0-9]+\.[0-9]+\b", read(root, rel)) if v != ver})
        if others:
            problems.append(f"{rel}: mentions {', '.join(others)} but the current release is {ver}")

    # --- stale headline counts anywhere in the tree -----------------------------------------
    # Tool/migration counts from older releases keep leaking into packaging text, dashboard
    # copy and docs. History files may legitimately quote them.
    stale_rx = re.compile(
        r"\b(?:16|23|35)[- ]tools?\b|\b(?:16|23|35) (?:active |deterministic |registered )?MCP tools\b"
        r"|\b(?:16|23|35)-tool\b|\b(?:93|95)[ -](?:idempotent |SQL |PostgreSQL )*migrations?\b|\bsixteen tools\b",
        re.IGNORECASE)
    skip_prefixes = ("docs/history/", "migrations/", "scripts/sync_docs_from_manifest.py")
    skip_files = {"CHANGELOG.md", "RELEASE_NOTES.md", "BENCHMARKS.md"}
    text_ext = (".md", ".go", ".js", ".html", ".yaml", ".yml", ".json", ".rb", ".nuspec", ".spec", ".py", ".sh", ".ps1")
    tracked = subprocess.run(["git", "-C", root, "ls-files"], capture_output=True, text=True).stdout.splitlines()
    for rel in tracked:
        if rel in skip_files or rel.startswith(skip_prefixes) or not (rel.endswith(text_ext) or rel.endswith("PKGBUILD")):
            continue
        if rel == "docs/INTERNAL_MASTER_ARCHITECTURE_REPORT.md":
            continue  # contains a dated release-history table
        try:
            body = read(root, rel)
        except (OSError, UnicodeDecodeError):
            continue
        for n, line in enumerate(body.splitlines(), 1):
            if stale_rx.search(line):
                problems.append(f"{rel}:{n}: stale tool/migration count ({stale_rx.search(line).group(0)!r}); current: "
                                f"{counts['mcp_tools']} tools, {migrations} migrations")

    # --- release-blocker contracts (unchanged from earlier releases) -----------------------
    dockerfile = read(root, "Dockerfile")
    if f"ARG VERSION={ver}" not in dockerfile or "garuda-mcp" not in dockerfile or "garuda-talon" not in dockerfile:
        problems.append("Dockerfile must default to the canonical version and build all 4 binaries")
    matrix_sh = read(root, "scripts/build_release_matrix.sh")
    if 'OUT_DIR="${1:-dist/releases}"' not in matrix_sh:
        problems.append('scripts/build_release_matrix.sh must write to OUT_DIR="${1:-dist/releases}"')
    action_yml = read(root, "action.yml")
    if "GITHUB_ACTION_PATH" not in action_yml or "--diff" not in action_yml or '|| echo "ALLOW"' in action_yml:
        problems.append("action.yml failed the fail-closed / GITHUB_ACTION_PATH audit")
    pos_md = read(root, "docs/POSITIONING.md")
    if "sixteen tools" in pos_md.lower():
        problems.append("docs/POSITIONING.md still claims sixteen tools")

    # --- report ------------------------------------------------------------------------------
    for line in fixed:
        print("fixed:  " + line)
    if problems:
        for line in problems:
            print("DRIFT:  " + line)
        print(f"\n{len(problems)} problem(s). " + ("Run without --check to fix the mechanical ones." if check_only else "Fix the remaining ones by hand."))
        return 1
    print(f"OK  {ver} | {len(tools)} MCP tools ({counts['mcp_readonly_tools']} read-only, {counts['mcp_mutating_tools']} mutating) | "
          f"{migrations} migrations | {skills} skills | {counts['dashboard_surfaces']} dashboard surfaces")
    return 0


def main():
    parser = argparse.ArgumentParser(description="Check or sync release facts from capabilities.json")
    parser.add_argument("--repo", default=".", help="repository root (default: .)")
    parser.add_argument("--check", action="store_true", help="only report drift, change nothing")
    parser.add_argument("--set", metavar="vX.Y.Z", help="start a new release: set this version in capabilities.json, then sync")
    parser.add_argument("--date", metavar="YYYY-MM-DD", help="release date to record with --set")
    args = parser.parse_args()
    root = os.path.abspath(args.repo)
    if args.set:
        if args.check:
            parser.error("--set and --check cannot be combined")
        set_new_release(root, args.set, args.date)
    sys.exit(run(root, args.check))


if __name__ == "__main__":
    main()
