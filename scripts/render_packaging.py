#!/usr/bin/env python3
# Copyright 2026 Rohit Mishra
# SPDX-License-Identifier: Apache-2.0
"""
Fill the download checksums into the package-manager manifests in packaging/.

    python3 scripts/render_packaging.py --sums dist/releases/SHA256SUMS.txt --out dist/releases/packaging

The manifests in packaging/ carry no checksums because the archives do not exist until a
release is built. This reads the release's SHA256SUMS.txt and writes ready-to-submit copies:

  homebrew/garuda.rb      sha256 "<hash>" after every url line
  aur/PKGBUILD            sha256sums_x86_64 / sha256sums_aarch64
  winget/garuda.yaml      InstallerSha256 after every InstallerUrl
  rpm/garuda.spec         copied unchanged (rpm takes no checksum)
  choco/garuda.nuspec     copied unchanged (packs local files)

Exits 1 if a checksum that a manifest needs is missing from SHA256SUMS.txt.
"""

import argparse
import os
import re
import shutil
import sys


def load_sums(path):
    sums = {}
    with open(path, encoding="utf-8") as fh:
        for line in fh:
            m = re.match(r"^([0-9a-fA-F]{64})\s+\*?(\S+)\s*$", line)
            if m:
                sums[os.path.basename(m.group(2))] = m.group(1).lower()
    return sums


def need(sums, name, missing):
    if name not in sums:
        missing.append(name)
        return "0" * 64
    return sums[name]


def render_homebrew(text, sums, missing):
    out = []
    for line in text.splitlines():
        out.append(line)
        m = re.match(r'^(\s*)url "[^"]*/([^/"]+)"\s*$', line)
        if m:
            out.append(f'{m.group(1)}sha256 "{need(sums, m.group(2), missing)}"')
    # drop any pre-existing duplicate sha256 line that directly follows another sha256 line
    cleaned = []
    for line in out:
        if cleaned and line.strip().startswith("sha256 ") and cleaned[-1].strip().startswith("sha256 "):
            cleaned[-1] = line
            continue
        cleaned.append(line)
    return "\n".join(cleaned) + "\n"


def render_pkgbuild(text, sums, version, missing):
    text = re.sub(r"sha256sums_x86_64=\([^)]*\)",
                  f"sha256sums_x86_64=('{need(sums, f'garuda-v{version}-linux-amd64.tar.gz', missing)}')", text)
    text = re.sub(r"sha256sums_aarch64=\([^)]*\)",
                  f"sha256sums_aarch64=('{need(sums, f'garuda-v{version}-linux-arm64.tar.gz', missing)}')", text)
    return text


def render_winget(text, sums, missing):
    out = []
    for line in text.splitlines():
        if line.strip().startswith("InstallerSha256:"):
            continue
        out.append(line)
        m = re.match(r"^(\s*)InstallerUrl:\s*\S+/([^/\s]+)\s*$", line)
        if m:
            out.append(f"{m.group(1)}InstallerSha256: {need(sums, m.group(2), missing).upper()}")
    return "\n".join(out) + "\n"


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--sums", required=True, help="path to the release's SHA256SUMS.txt")
    ap.add_argument("--out", required=True, help="directory to write the rendered manifests to")
    ap.add_argument("--root", default=os.path.join(os.path.dirname(os.path.abspath(__file__)), ".."))
    args = ap.parse_args()

    root = os.path.abspath(args.root)
    with open(os.path.join(root, "version.txt"), encoding="utf-8") as fh:
        version = fh.read().strip().lstrip("v")
    sums = load_sums(args.sums)
    missing = []

    jobs = [
        ("homebrew/garuda.rb", lambda t: render_homebrew(t, sums, missing)),
        ("aur/PKGBUILD", lambda t: render_pkgbuild(t, sums, version, missing)),
        ("winget/garuda.yaml", lambda t: render_winget(t, sums, missing)),
        ("rpm/garuda.spec", lambda t: t),
        ("choco/garuda.nuspec", lambda t: t),
    ]
    for rel, fn in jobs:
        with open(os.path.join(root, "packaging", rel), encoding="utf-8") as fh:
            rendered = fn(fh.read())
        dest = os.path.join(args.out, rel)
        os.makedirs(os.path.dirname(dest), exist_ok=True)
        with open(dest, "w", encoding="utf-8", newline="\n") as fh:
            fh.write(rendered)
        print(f"wrote {os.path.relpath(dest)}")

    if missing:
        print("\nERROR: SHA256SUMS.txt has no entry for: " + ", ".join(sorted(set(missing))), file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
