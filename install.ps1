# ═══════════════════════════════════════════════════════════════════════════
# 🦅 GARUDA v1.6 ONE-COMMAND WINDOWS POWERSHELL BOOTSTRAPPER (install.ps1)
# ═══════════════════════════════════════════════════════════════════════════
# Run from any Windows PowerShell window (even if Git, Go, Postgres, or Docker
# are NOT installed):
#
#   Set-ExecutionPolicy Bypass -Scope Process -Force; irm https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.ps1 | iex
# ═══════════════════════════════════════════════════════════════════════════

$ErrorActionPreference = "Stop"
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$InstallDir = if ($env:GARUDA_INSTALL_DIR) { $env:GARUDA_INSTALL_DIR } else { Join-Path $HOME "garuda" }

Write-Host "===================================================================" -ForegroundColor Cyan
Write-Host "🦅 GARUDA 1-COMMAND WINDOWS INSTALLER" -ForegroundColor Cyan
Write-Host "===================================================================" -ForegroundColor Cyan

if (-not (Test-Path (Join-Path $InstallDir "setup.ps1"))) {
    if (Get-Command git -ErrorAction SilentlyContinue) {
        Write-Host "⬇️ Cloning Garuda into $InstallDir..." -ForegroundColor Yellow
        git clone https://github.com/myshra777-ai/garuda.git $InstallDir
    } else {
        Write-Host "⬇️ Git not installed — downloading Garuda release archive directly..." -ForegroundColor Yellow
        $TmpZip = Join-Path $env:TEMP "garuda-main.zip"
        $TmpExtract = Join-Path $env:TEMP "garuda-extract"
        Invoke-WebRequest -Uri "https://github.com/myshra777-ai/garuda/archive/refs/heads/main.zip" -OutFile $TmpZip -UseBasicParsing
        if (Test-Path $TmpExtract) { Remove-Item $TmpExtract -Recurse -Force }
        Expand-Archive -Path $TmpZip -DestinationPath $TmpExtract -Force
        if (Test-Path $InstallDir) { Remove-Item $InstallDir -Recurse -Force }
        Move-Item -Path (Join-Path $TmpExtract "garuda-main") -Destination $InstallDir -Force
        Remove-Item $TmpZip -Force -ErrorAction SilentlyContinue
    }
} else {
    Write-Host "✓ Found existing Garuda installation at $InstallDir" -ForegroundColor Green
}

Set-Location $InstallDir
& powershell -ExecutionPolicy Bypass -File (Join-Path $InstallDir "setup.ps1")
