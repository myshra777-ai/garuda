#!/usr/bin/env bash
# ===========================================================================
# GARUDA v1.9.2 MULTI-OS RELEASE & PACKAGE MATRIX BUILDER
# ===========================================================================
# Cross-compiles 100% standalone zero-CGO (CGO_ENABLED=0) release archives
# and OS-native package manifests for:
#   1. Windows 10/11 (x64 / amd64)             -> garuda-v1.9.2-windows-amd64.zip
#   2. Windows 11 ARM64 (Snapdragon)           -> garuda-v1.9.2-windows-arm64.zip
#   3. macOS Apple Silicon (M1/M2/M3/M4 arm64) -> garuda-v1.9.2-darwin-arm64.tar.gz
#   4. macOS Intel (x86_64 / amd64)            -> garuda-v1.9.2-darwin-amd64.tar.gz
#   5. Linux x86_64 (Ubuntu/Debian/RHEL/WSL2)  -> garuda-v1.9.2-linux-amd64.tar.gz + .deb
#   6. Linux ARM64 & Chromebook (Crostini)     -> garuda-v1.9.2-linux-arm64.tar.gz + .deb
#   7. Android Termux (ARM64)                  -> garuda-v1.9.2-android-arm64-termux.tar.gz
#   8. FreeBSD x86_64                          -> garuda-v1.9.2-freebsd-amd64.tar.gz
#   9. Source & Universal Installer Bundle     -> garuda-v1.9.2-universal-source.tar.gz
#
# NOTE: dist/ is git-ignored and NEVER committed into Git so the repository
# and universal source tarball stay lean (~5-8 MB instead of 255 MB).
# ===========================================================================

set -euo pipefail

cd "$(dirname "$0")/.."
DEFAULT_VER="$(tr -d '[:space:]' < version.txt 2>/dev/null || echo 'v1.9.2')"
[[ -z "$DEFAULT_VER" ]] && DEFAULT_VER="v1.9.2"
VERSION="${GARUDA_VERSION:-$DEFAULT_VER}"
VERSION_NO_V="${VERSION#v}"
OUT_DIR="${1:-dist}"
rm -rf "$OUT_DIR"
mkdir -p "$OUT_DIR"

if [[ -x "$HOME/.garuda/toolchain/go/bin/go" ]]; then
    export PATH="$HOME/.garuda/toolchain/go/bin:$PATH"
fi

LDFLAGS="-s -w -X main.version=${VERSION}"

TARGETS=(
    "windows/amd64/garuda-${VERSION}-windows-amd64"
    "windows/arm64/garuda-${VERSION}-windows-arm64"
    "darwin/arm64/garuda-${VERSION}-darwin-arm64"
    "darwin/amd64/garuda-${VERSION}-darwin-amd64"
    "linux/amd64/garuda-${VERSION}-linux-amd64"
    "linux/arm64/garuda-${VERSION}-linux-arm64"
    "linux/arm64/garuda-${VERSION}-android-arm64-termux"
    "freebsd/amd64/garuda-${VERSION}-freebsd-amd64"
)

echo "==================================================================="
echo "Building Garuda ${VERSION} Multi-OS Release Matrix in ${OUT_DIR}/"
echo "==================================================================="

build_deb_package() {
    local arch_go="$1"
    local stage_dir="$2"
    local deb_arch="amd64"
    [[ "$arch_go" == "arm64" ]] && deb_arch="arm64"

    if ! command -v dpkg-deb >/dev/null 2>&1; then
        return 0
    fi

    local deb_root="${OUT_DIR}/deb-build-${deb_arch}"
    mkdir -p "${deb_root}/DEBIAN"
    mkdir -p "${deb_root}/usr/local/bin"
    mkdir -p "${deb_root}/usr/share/garuda/migrations"
    mkdir -p "${deb_root}/usr/share/garuda/policies"
    mkdir -p "${deb_root}/usr/share/doc/garuda"

    cp "${stage_dir}/bin/garuda" "${stage_dir}/bin/garuda-api" "${stage_dir}/bin/garuda-mcp" "${stage_dir}/bin/garuda-talon" "${deb_root}/usr/local/bin/"
    chmod 755 "${deb_root}/usr/local/bin/"*
    cp -r migrations/* "${deb_root}/usr/share/garuda/migrations/" 2>/dev/null || true
    cp -r policies/* "${deb_root}/usr/share/garuda/policies/" 2>/dev/null || true
    cp README.md INSTALL.md PLAYBOOK.md CHANGELOG.md RELEASE_NOTES.md "${deb_root}/usr/share/doc/garuda/" 2>/dev/null || true

    cat <<EOF > "${deb_root}/DEBIAN/control"
Package: garuda
Version: ${VERSION_NO_V}
Section: devel
Priority: optional
Architecture: ${deb_arch}
Maintainer: Rohit Mishra <myshra777@gmail.com>
Homepage: https://github.com/myshra777-ai/garuda
Description: Garuda — Local-First AI Code Change-Control, Financial Audit & 35-Tool MCP Engine
 Deterministic compiler-grade AST analysis, active AI agent harness (Garuda Talon),
 human-in-the-loop 0.0-1.0 formula weighting, 3-tier surface criticality,
 continuous financial data discrepancy auditing (48 Cr vs 4.8 Cr), and RFC 6962 Merkle proofs.
EOF

    dpkg-deb --build --root-owner-group "${deb_root}" "${OUT_DIR}/garuda_${VERSION_NO_V}_${deb_arch}.deb" >/dev/null 2>&1 || true
    rm -rf "${deb_root}"
    if [[ -f "${OUT_DIR}/garuda_${VERSION_NO_V}_${deb_arch}.deb" ]]; then
        echo "  [OK] Packaged Debian/Ubuntu installer: ${OUT_DIR}/garuda_${VERSION_NO_V}_${deb_arch}.deb"
    fi
}

for entry in "${TARGETS[@]}"; do
    IFS='/' read -r GOOS GOARCH PKG_NAME <<< "$entry"
    EXT=""
    [[ "$GOOS" == "windows" ]] && EXT=".exe"

    STAGE_DIR="${OUT_DIR}/${PKG_NAME}"
    mkdir -p "${STAGE_DIR}/bin" "${STAGE_DIR}/docs" "${STAGE_DIR}/scripts"

    echo "[*] Cross-compiling ${PKG_NAME} (${GOOS}/${GOARCH}, CGO_ENABLED=0)..."
    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -trimpath -ldflags="$LDFLAGS" -o "${STAGE_DIR}/bin/garuda${EXT}" ./cmd/garuda
    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -trimpath -ldflags="$LDFLAGS" -o "${STAGE_DIR}/bin/garuda-api${EXT}" ./cmd/garuda-api
    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -trimpath -ldflags="$LDFLAGS" -o "${STAGE_DIR}/bin/garuda-mcp${EXT}" ./cmd/garuda-mcp
    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -trimpath -ldflags="$LDFLAGS" -o "${STAGE_DIR}/bin/garuda-talon${EXT}" ./cmd/garuda-talon
    echo "prebuilt" > "${STAGE_DIR}/bin/.build_stamp"

    cp -r migrations policies cmd internal "${STAGE_DIR}/"
    cp go.mod go.sum version.txt setup.sh setup.ps1 install.sh install.ps1 README.md INSTALL.md PLAYBOOK.md CHANGELOG.md EVIDENCE.md LICENSE "${STAGE_DIR}/" 2>/dev/null || true
    [[ -f RELEASE_NOTES.md ]] && cp RELEASE_NOTES.md "${STAGE_DIR}/"
    cp docs/*.md "${STAGE_DIR}/docs/" 2>/dev/null || true
    cp scripts/mcp_verify.py "${STAGE_DIR}/scripts/" 2>/dev/null || true

    if [[ "$GOOS" == "windows" ]]; then
        if command -v zip >/dev/null 2>&1; then
            (cd "$OUT_DIR" && zip -qr "${PKG_NAME}.zip" "${PKG_NAME}")
        elif command -v python3 >/dev/null 2>&1; then
            python3 -c "import shutil; shutil.make_archive('${OUT_DIR}/${PKG_NAME}', 'zip', '${OUT_DIR}', '${PKG_NAME}')"
        else
            tar -czf "${OUT_DIR}/${PKG_NAME}.tar.gz" -C "$OUT_DIR" "${PKG_NAME}"
        fi
        echo "  [OK] Packaged ${OUT_DIR}/${PKG_NAME}.zip"
    else
        tar -czf "${OUT_DIR}/${PKG_NAME}.tar.gz" -C "$OUT_DIR" "${PKG_NAME}"
        echo "  [OK] Packaged ${OUT_DIR}/${PKG_NAME}.tar.gz"
        if [[ "$GOOS" == "linux" && "$PKG_NAME" != *"android"* ]]; then
            build_deb_package "$GOARCH" "$STAGE_DIR"
        fi
    fi

    rm -rf "$STAGE_DIR"
done

# Package lean universal source & 1-command installer archive (strictly excluding dist/, bin/, .git/, .garuda/)
STAGE_SRC="${OUT_DIR}/garuda-${VERSION}-universal"
mkdir -p "$STAGE_SRC"
tar --exclude='./.git' \
    --exclude="./${OUT_DIR}" \
    --exclude='./dist' \
    --exclude='./bin' \
    --exclude='./.garuda' \
    --exclude='./refund-api/refund-api' \
    --exclude='./refund-api/bin' \
    --exclude='./install_garuda_*.sh' \
    --exclude='./install_garuda_*.bash' \
    -cf - . | (cd "$STAGE_SRC" && tar -xf -)
tar -czf "${OUT_DIR}/garuda-${VERSION}-universal-source.tar.gz" -C "$OUT_DIR" "garuda-${VERSION}-universal"
rm -rf "$STAGE_SRC"
echo "  [OK] Packaged lean ${OUT_DIR}/garuda-${VERSION}-universal-source.tar.gz"

# Copy 1-click installer scripts & release notes into dist/ for direct GitHub Release attachment
cp install.sh install.ps1 setup.sh setup.ps1 "${OUT_DIR}/" 2>/dev/null || true
[[ -f RELEASE_NOTES.md ]] && cp RELEASE_NOTES.md "${OUT_DIR}/RELEASE_NOTES_${VERSION}.md"

# Generate SHA256SUMS.txt
(
    cd "$OUT_DIR"
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum *.tar.gz *.zip *.deb *.sh *.ps1 2>/dev/null > SHA256SUMS.txt || true
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 *.tar.gz *.zip *.deb *.sh *.ps1 2>/dev/null > SHA256SUMS.txt || true
    fi
)
echo "  [OK] Generated ${OUT_DIR}/SHA256SUMS.txt"
echo "==================================================================="
echo "[SUCCESS] All Garuda ${VERSION} multi-OS release packages built in ${OUT_DIR}/"
echo "NOTE: dist/ is git-ignored; attach dist/* to GitHub Releases via 'gh release upload' or CI."
echo "==================================================================="
