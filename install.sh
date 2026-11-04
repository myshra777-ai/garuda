#!/usr/bin/env bash
# ===========================================================================
# GARUDA ONE-COMMAND UNIVERSAL INSTALLER (install.sh)
# ===========================================================================
# Supports:
#   - macOS (Apple Silicon M1/M2/M3/M4 & Intel x86_64)
#   - Linux (Ubuntu, Debian, Fedora, Arch, RHEL, WSL2)
#   - Chromebook / ChromiumOS (Crostini Linux container)
#   - Android (Termux)
#   - Windows Git Bash / MSYS2 / Cygwin (delegates to setup.ps1)
#
# Run in 1 command (even if Go, PostgreSQL, GCC, or Docker are missing):
#   curl -fsSL https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.sh | bash
#
# Needs only `curl` (or `git`) plus `tar`. Everything else (Go, PostgreSQL) is
# downloaded into ~/.garuda by setup.sh, with no sudo.
#   Android (Termux):   pkg install -y curl git && curl -fsSL <url above> | bash
#   Debian/Ubuntu/ChromeOS Crostini:  sudo apt-get install -y curl git xz-utils unzip
#   macOS:              xcode-select --install   (provides git and curl)
#
# Existing installs are updated safely: the installer refuses to touch a
# directory that is not a Garuda checkout, and refuses to discard local edits
# or local commits.
# ===========================================================================

set -euo pipefail

INSTALL_DIR="${GARUDA_INSTALL_DIR:-$HOME/garuda}"

if ! command -v git >/dev/null 2>&1 && ! command -v curl >/dev/null 2>&1; then
    echo "[ERROR] Neither git nor curl was found. Install one of them first:" >&2
    echo "        Android Termux: pkg install -y curl git" >&2
    echo "        Debian/Ubuntu/ChromeOS: sudo apt-get install -y curl git" >&2
    echo "        macOS: xcode-select --install" >&2
    exit 1
fi

echo "==================================================================="
echo "GARUDA UNIVERSAL 1-COMMAND INSTALLER"
echo "==================================================================="

if [[ ! -f "$INSTALL_DIR/setup.sh" ]]; then
    if [[ -e "$INSTALL_DIR" || -L "$INSTALL_DIR" ]]; then
        if [[ ! -d "$INSTALL_DIR" ]]; then
            echo "[ERROR] Install target $INSTALL_DIR exists and is not a directory. Choose a fresh GARUDA_INSTALL_DIR; no files were changed." >&2
            exit 1
        fi
        if [[ -n "$(find "$INSTALL_DIR" -mindepth 1 -maxdepth 1 -print -quit 2>/dev/null)" ]]; then
            echo "[ERROR] Install target $INSTALL_DIR already exists and is not a Garuda installation. Back up its contents and choose a fresh GARUDA_INSTALL_DIR; the installer will not delete it." >&2
            exit 1
        fi
        rmdir "$INSTALL_DIR"
    fi
    mkdir -p "$(dirname "$INSTALL_DIR")"
    if command -v git >/dev/null 2>&1; then
        echo "[*] Cloning Garuda into $INSTALL_DIR (shallow clone --depth 1 for fast setup)..."
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
            # Either the user has real local commits, or upstream history was
            # rewritten (force-pushed / re-rooted) since they cloned. We cannot
            # tell which from a shallow fetch, and we never discard commits, so
            # move the old checkout aside intact and take a fresh clone. Data
            # in ~/.garuda lives outside this directory and is reused.
            backup="${INSTALL_DIR}.old-$(date +%Y%m%d-%H%M%S)"
            echo "[!] Local commits exist that are not in origin/main ($local_ahead in $INSTALL_DIR): upstream history was rewritten, or you committed locally."
            echo "[*] Moving it aside, unchanged, to $backup and installing a fresh copy."
            mv "$INSTALL_DIR" "$backup" || {
                echo "[ERROR] Could not move $INSTALL_DIR aside. Nothing was changed." >&2
                exit 1
            }
            if ! git clone --depth 1 --single-branch --branch main https://github.com/myshra777-ai/garuda.git "$INSTALL_DIR"; then
                echo "[ERROR] Fresh clone failed. Restoring your original checkout." >&2
                rmdir "$INSTALL_DIR" 2>/dev/null || true
                mv "$backup" "$INSTALL_DIR"
                exit 1
            fi
            echo "[OK] Fresh copy installed. Your previous checkout is kept at $backup (delete it when you no longer need it)."
        elif ! git -C "$INSTALL_DIR" reset --hard origin/main; then
            echo "[ERROR] Update failed. Inspect $INSTALL_DIR before retrying." >&2
            exit 1
        fi
    fi
fi

cd "$INSTALL_DIR"
chmod +x setup.sh
exec ./setup.sh "$@"
