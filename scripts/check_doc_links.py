#!/usr/bin/env python3
# Copyright 2026 Rohit Mishra
# SPDX-License-Identifier: Apache-2.0
"""
Fail if any relative link or image in the Markdown docs points at a file that does not exist.

    python3 scripts/check_doc_links.py            # check every tracked *.md
    python3 scripts/check_doc_links.py README.md docs/INSTALL.md

Absolute http(s) links, mailto: links and in-page #anchors are not checked.
Links inside fenced code blocks and `inline code` are ignored.
"""

import os
import re
import subprocess
import sys
import urllib.parse

INLINE_LINK = re.compile(r"!?\[[^\]]*\]\(\s*<?([^)\s>]+)>?(?:\s+\"[^\"]*\")?\s*\)")
HTML_ATTR = re.compile(r"""(?:href|src)\s*=\s*["']([^"']+)["']""", re.IGNORECASE)
REF_DEF = re.compile(r"^\s*\[[^\]]+\]:\s*<?(\S+?)>?(?:\s+[\"'(].*)?$")
INLINE_CODE = re.compile(r"`[^`\n]*`")
SKIP_PREFIXES = ("http://", "https://", "mailto:", "tel:", "data:", "#", "javascript:")


def tracked_markdown(root):
    out = subprocess.run(["git", "-C", root, "ls-files", "*.md"], capture_output=True, text=True, check=True).stdout
    return [p for p in out.splitlines() if p and os.path.exists(os.path.join(root, p))]


def links_in(path):
    in_fence = False
    with open(path, encoding="utf-8", errors="replace") as fh:
        for lineno, line in enumerate(fh, 1):
            if line.lstrip().startswith(("```", "~~~")):
                in_fence = not in_fence
                continue
            if in_fence:
                continue
            scrubbed = INLINE_CODE.sub("", line)
            for rx in (INLINE_LINK, HTML_ATTR):
                for m in rx.finditer(scrubbed):
                    yield lineno, m.group(1)
            m = REF_DEF.match(scrubbed)
            if m:
                yield lineno, m.group(1)


def main():
    root = subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True, check=True).stdout.strip()
    files = sys.argv[1:] or tracked_markdown(root)
    broken = []
    checked = 0
    for rel in files:
        full = rel if os.path.isabs(rel) else os.path.join(root, rel)
        base = os.path.dirname(full)
        for lineno, target in links_in(full):
            if target.startswith(SKIP_PREFIXES) or "{" in target or "<" in target:
                continue
            path_part = urllib.parse.unquote(target.split("#", 1)[0].split("?", 1)[0])
            if not path_part:
                continue
            resolved = os.path.join(root, path_part.lstrip("/")) if path_part.startswith("/") else os.path.normpath(os.path.join(base, path_part))
            checked += 1
            if not os.path.exists(resolved):
                broken.append(f"{os.path.relpath(full, root)}:{lineno}: broken link -> {target}")
    for line in broken:
        print(line)
    if broken:
        print(f"\n{len(broken)} broken link(s) in {len(files)} file(s) ({checked} relative links checked).")
        return 1
    print(f"OK  {checked} relative links in {len(files)} Markdown files all resolve.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
