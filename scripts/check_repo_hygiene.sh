#!/usr/bin/env bash
# Fails if build artefacts or oversized files are tracked by git.
# usage: scripts/check_repo_hygiene.sh [max_kb=2048]
set -euo pipefail
MAX_KB="${1:-2048}"; bad=0
while IFS= read -r f; do
  [[ -f "$f" ]] || continue
  case "$f" in dist/*|bin/*|*.test|c.out|coverage.out) echo "tracked build artefact: $f"; bad=1; continue;; esac
  case "$f" in assets/*) continue;; esac
  kb=$(( ($(stat -c %s "$f") + 1023) / 1024 ))
  if (( kb > MAX_KB )); then echo "tracked file larger than ${MAX_KB}KB: $f (${kb}KB)"; bad=1; fi
done < <(git ls-files)
exit $bad
