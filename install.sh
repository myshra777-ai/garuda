#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════════════
# 🦅 GARUDA v1.6 ONE-COMMAND UNIVERSAL INSTALLER (install.sh)
# ═══════════════════════════════════════════════════════════════════════════
# Supports:
#   • macOS (Apple Silicon M1/M2/M3/M4 & Intel)
#   • Linux (Ubuntu, Debian, Fedora, Arch, RHEL, WSL2)
#   • Chromebook / ChromiumOS (Crostini Linux container)
#   • Android (Termux)
#
# Run in 1 command (even if Git, Go, PostgreSQL, GCC, or Docker are missing):
#   curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash
# ═══════════════════════════════════════════════════════════════════════════

set -euo pipefail

INSTALL_DIR="${GARUDA_INSTALL_DIR:-$HOME/garuda}"

echo "═══════════════════════════════════════════════════════════════════"
echo "🦅 GARUDA v1.6 UNIVERSAL 1-COMMAND INSTALLER"
echo "═══════════════════════════════════════════════════════════════════"

if [[ ! -f "$INSTALL_DIR/setup.sh" ]]; then
    if command -v git >/dev/null 2>&1; then
        echo "⬇️  Cloning Garuda into $INSTALL_DIR..."
        git clone https://github.com/myshra777-ai/garuda.git "$INSTALL_DIR"
    else
        echo "⬇️  Git not found — downloading Garuda archive via curl..."
        mkdir -p "$INSTALL_DIR"
        curl -fsSL "https://github.com/myshra777-ai/garuda/archive/refs/heads/main.tar.gz" | tar -xz -C "$INSTALL_DIR" --strip-components=1
    fi
else
    echo "✓ Found existing Garuda directory at $INSTALL_DIR"
fi

cd "$INSTALL_DIR"
chmod +x setup.sh
exec ./setup.sh "$@"
