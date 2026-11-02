#!/usr/bin/env bash
# Removes the "// Law Enforcement. I am bound by the ACGM ..." line (and the bare "//"
# spacer above it) from the license header of tracked source files. Keeps Copyright + SPDX.
# usage: scripts/strip-law-header.sh   (then review: git diff --stat)
set -euo pipefail
python3 -I - <<'PY'
import re, subprocess
files = subprocess.check_output(["git", "ls-files", "*.go", "*.py", "*.sh"], text=True).split("\n")
pat = re.compile(r"^(?:(?://|#)[ ]*\n)?(?://|#) Law Enforcement\. I am bound by the ACGM[^\n]*\n", re.M)
n = 0
for f in filter(None, files):
    try:
        s = open(f, encoding="utf-8").read()
    except Exception:
        continue
    t = pat.sub("", s, count=1)
    if t != s:
        open(f, "w", encoding="utf-8").write(t)
        n += 1
print(f"stripped header line from {n} files")
PY
