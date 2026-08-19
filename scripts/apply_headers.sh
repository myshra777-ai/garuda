#!/usr/bin/env bash
set -euo pipefail

HEADER="// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute."

find . -type f -name "*.go" ! -path "*/vendor/*" ! -path "*/.git/*" | while read -r file; do
    if ! grep -q "SPDX-License-Identifier: Apache-2.0" "$file"; then
        echo "Adding header to $file"
        temp_file=$(mktemp)
        printf "%s\n\n" "$HEADER" > "$temp_file"
        cat "$file" >> "$temp_file"
        mv "$temp_file" "$file"
    fi
done

echo "✅ License headers applied across all Go source files."
