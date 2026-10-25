#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════════════
# 🦅 GARUDA MULTI-OS STANDALONE BINARY MATRIX BUILDER
# ═══════════════════════════════════════════════════════════════════════════
# Cross-compiles 100% standalone zero-CGO (CGO_ENABLED=0) release bundles for:
#   • windows/amd64  (Windows 10/11 x64)
#   • windows/arm64  (Windows on Snapdragon / ARM64)
#   • darwin/arm64   (macOS Apple Silicon M1/M2/M3/M4)
#   • darwin/amd64   (macOS Intel)
#   • linux/amd64    (Linux x86_64, WSL2, x86 Chromebooks)
#   • linux/arm64    (Linux ARM64, ARM Chromebooks, Android Termux)
# ═══════════════════════════════════════════════════════════════════════════

set -euo pipefail

cd "$(dirname "$0")/.."
OUT_DIR="${1:-dist}"
mkdir -p "$OUT_DIR"

TARGETS=(
    "windows/amd64"
    "windows/arm64"
    "darwin/arm64"
    "darwin/amd64"
    "linux/amd64"
    "linux/arm64"
)

for target in "${TARGETS[@]}"; do
    GOOS="${target%/*}"
    GOARCH="${target#*/}"
    EXT=""
    [[ "$GOOS" == "windows" ]] && EXT=".exe"

    STAGE_DIR="${OUT_DIR}/garuda-${GOOS}-${GOARCH}"
    mkdir -p "${STAGE_DIR}/bin"

    echo "🔨 Building standalone bundle for ${GOOS}/${GOARCH} (CGO_ENABLED=0)..."
    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="-s -w" -o "${STAGE_DIR}/bin/garuda${EXT}" ./cmd/garuda
    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="-s -w" -o "${STAGE_DIR}/bin/garuda-api${EXT}" ./cmd/garuda-api
    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="-s -w" -o "${STAGE_DIR}/bin/garuda-mcp${EXT}" ./cmd/garuda-mcp
    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="-s -w" -o "${STAGE_DIR}/bin/garuda-talon${EXT}" ./cmd/garuda-talon

    cp -r migrations policies setup.sh setup.ps1 install.sh install.ps1 README.md INSTALL.md PLAYBOOK.md "${STAGE_DIR}/"

    if [[ "$GOOS" == "windows" ]] && command -v zip >/dev/null 2>&1; then
        (cd "$OUT_DIR" && zip -qr "garuda-${GOOS}-${GOARCH}.zip" "garuda-${GOOS}-${GOARCH}")
    else
        tar -czf "${OUT_DIR}/garuda-${GOOS}-${GOARCH}.tar.gz" -C "$OUT_DIR" "garuda-${GOOS}-${GOARCH}"
    fi
    rm -rf "$STAGE_DIR"
    echo "  ✓ Packaged ${OUT_DIR}/garuda-${GOOS}-${GOARCH}"
done

echo "🎉 All 6 cross-platform standalone release bundles built in ${OUT_DIR}/"
