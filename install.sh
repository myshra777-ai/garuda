#!/usr/bin/env bash
# ===========================================================================
# GARUDA v1.8.4 ONE-COMMAND UNIVERSAL INSTALLER (install.sh)
# ===========================================================================
# Supports:
#   - macOS (Apple Silicon M1/M2/M3/M4 & Intel x86_64)
#   - Linux (Ubuntu, Debian, Fedora, Arch, RHEL, WSL2)
#   - Chromebook / ChromiumOS (Crostini Linux container)
#   - Android (Termux)
#   - Windows Git Bash / MSYS2 / Cygwin (delegates to setup.ps1)
#
# Run in 1 command (even if Git, Go, PostgreSQL, GCC, or Docker are missing):
#   curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash
# ===========================================================================

set -euo pipefail

INSTALL_DIR="${GARUDA_INSTALL_DIR:-$HOME/garuda}"

echo "==================================================================="
echo "GARUDA v1.8.4 UNIVERSAL 1-COMMAND INSTALLER"
echo "==================================================================="

if [[ ! -f "$INSTALL_DIR/setup.sh" ]]; then
    if command -v git >/dev/null 2>&1; then
        echo "[*] Cloning Garuda into $INSTALL_DIR (shallow clone --depth 1 for fast setup)..."
        rm -rf "$INSTALL_DIR"
        git clone --depth 1 --single-branch --branch main https://github.com/myshra777-ai/garuda.git "$INSTALL_DIR"
    else
        echo "[*] Git not found -- downloading Garuda archive via curl..."
        mkdir -p "$INSTALL_DIR"
        curl -fsSL "https://github.com/myshra777-ai/garuda/archive/refs/heads/main.tar.gz" | tar -xz -C "$INSTALL_DIR" --strip-components=1
    fi
else
    echo "[OK] Found existing Garuda directory at $INSTALL_DIR"
    if [[ -d "$INSTALL_DIR/.git" ]] && command -v git >/dev/null 2>&1; then
        echo "[*] Checking existing checkout before update..."
        if ! git -C "$INSTALL_DIR" rev-parse --verify HEAD >/dev/null 2>&1; then
            echo "[ERROR] Existing Git checkout at $INSTALL_DIR has no valid HEAD. No update was applied." >&2
            exit 1
        fi
        if [[ -n "$(git -C "$INSTALL_DIR" status --porcelain --untracked-files=no 2>/dev/null)" ]]; then
            echo "[ERROR] Tracked local changes found in $INSTALL_DIR. Commit or back up your work before updating; no files were changed." >&2
            exit 1
        fi
        if ! git -C "$INSTALL_DIR" fetch --depth 1 origin main; then
            echo "[ERROR] Could not fetch origin/main. Existing installation was left unchanged." >&2
            exit 1
        fi
        local_ahead="$(git -C "$INSTALL_DIR" rev-list --count origin/main..HEAD 2>/dev/null)" || {
            echo "[ERROR] Could not compare local commits with origin/main. Existing installation was left unchanged." >&2
            exit 1
        }
        if [[ "$local_ahead" =~ ^[0-9]+$ ]] && (( local_ahead > 0 )); then
            echo "[ERROR] Local commits exist that are not in origin/main. The installer will not discard them." >&2
            exit 1
        fi
        if ! git -C "$INSTALL_DIR" reset --hard origin/main; then
            echo "[ERROR] Update failed. Inspect $INSTALL_DIR before retrying." >&2
            exit 1
        fi
    fi
fi

cd "$INSTALL_DIR"
chmod +x setup.sh
exec ./setup.sh "$@"
