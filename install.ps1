# ===========================================================================
# GARUDA v1.8.4 ONE-COMMAND WINDOWS POWERSHELL BOOTSTRAPPER (install.ps1)
# ===========================================================================
# Run from any Windows PowerShell 5.1 or PowerShell 7 window (no Admin, Git,
# Go, PostgreSQL, GCC, or Docker required):
#
#   Set-ExecutionPolicy Bypass -Scope Process -Force; irm https://raw.githubusercontent.com/myshra777-ai/garuda/main/install.ps1 | iex
#
# NOTE: Keep this file strictly 7-bit ASCII so Windows PowerShell 5.1 (CP1252
# default encoding when no UTF-8 BOM is present) never misinterprets multi-byte
# UTF-8 sequences (0x93/0x94) as smart-quote string delimiters.
# ===========================================================================

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

function Invoke-QuietCmd([string]$CommandLine) {
    $oldPref = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        cmd.exe /c "$CommandLine"
        return $LASTEXITCODE
    } finally {
        $ErrorActionPreference = $oldPref
    }
}

function Download-FileFast([string]$Url, [string]$OutFile) {
    $curlExe = Join-Path $env:SystemRoot "System32\curl.exe"
    if (Test-Path $curlExe) {
        $oldPref = $ErrorActionPreference
        $ErrorActionPreference = "Continue"
        try {
            & $curlExe -fSL --retry 3 -o $OutFile $Url
            if ($LASTEXITCODE -eq 0 -and (Test-Path $OutFile)) {
                return
            }
        } finally {
            $ErrorActionPreference = $oldPref
        }
    }
    Invoke-WebRequest -Uri $Url -OutFile $OutFile -UseBasicParsing
}

function Extract-ArchiveFast([string]$ZipPath, [string]$DestDir) {
    New-Item -ItemType Directory -Force -Path $DestDir | Out-Null
    $tarExe = Join-Path $env:SystemRoot "System32\tar.exe"
    if (Test-Path $tarExe) {
        $oldPref = $ErrorActionPreference
        $ErrorActionPreference = "Continue"
        try {
            & $tarExe -xf $ZipPath -C $DestDir 2>$null
            if ($LASTEXITCODE -eq 0) {
                return
            }
        } finally {
            $ErrorActionPreference = $oldPref
        }
    }
    Expand-Archive -Path $ZipPath -DestinationPath $DestDir -Force
}

$InstallDir = if ($env:GARUDA_INSTALL_DIR) { $env:GARUDA_INSTALL_DIR } else { Join-Path $HOME "garuda" }

Write-Host "===================================================================" -ForegroundColor Cyan
Write-Host "[GARUDA] v1.8.4 1-COMMAND WINDOWS INSTALLER" -ForegroundColor Cyan
Write-Host "===================================================================" -ForegroundColor Cyan

if (-not (Test-Path (Join-Path $InstallDir "setup.ps1"))) {
    if (Get-Command git -ErrorAction SilentlyContinue) {
        Write-Host "[*] Cloning Garuda into $InstallDir (shallow clone for fast setup)..." -ForegroundColor Yellow
        if (Test-Path $InstallDir) {
            Remove-Item -Path $InstallDir -Recurse -Force -ErrorAction SilentlyContinue
        }
        $code = Invoke-QuietCmd "git clone --depth 1 https://github.com/myshra777-ai/garuda.git `"$InstallDir`""
        if ($code -ne 0 -or (-not (Test-Path (Join-Path $InstallDir "setup.ps1")))) {
            throw "Failed to clone Garuda repository into $InstallDir"
        }
    } else {
        Write-Host "[*] Git not installed -- downloading Garuda release archive directly..." -ForegroundColor Yellow
        $TmpZip = Join-Path $env:TEMP "garuda-main.zip"
        $TmpExtract = Join-Path $env:TEMP "garuda-extract"
        Download-FileFast "https://github.com/myshra777-ai/garuda/archive/refs/heads/main.zip" $TmpZip
        if (Test-Path $TmpExtract) { Remove-Item $TmpExtract -Recurse -Force -ErrorAction SilentlyContinue }
        Extract-ArchiveFast $TmpZip $TmpExtract
        if (Test-Path $InstallDir) { Remove-Item $InstallDir -Recurse -Force -ErrorAction SilentlyContinue }
        Move-Item -Path (Join-Path $TmpExtract "garuda-main") -Destination $InstallDir -Force
        Remove-Item $TmpZip -Force -ErrorAction SilentlyContinue
    }
} else {
    Write-Host "[OK] Found existing Garuda installation at $InstallDir" -ForegroundColor Green
    if ((Test-Path (Join-Path $InstallDir ".git")) -and (Get-Command git -ErrorAction SilentlyContinue)) {
        Write-Host "[*] Pulling latest updates in $InstallDir..." -ForegroundColor Yellow
        $null = Invoke-QuietCmd "git -C `"$InstallDir`" fetch --depth 1 origin main >nul 2>&1"
        $null = Invoke-QuietCmd "git -C `"$InstallDir`" reset --hard origin/main >nul 2>&1"
    } else {
        try {
            Download-FileFast "https://raw.githubusercontent.com/myshra777-ai/garuda/main/setup.ps1" (Join-Path $InstallDir "setup.ps1")
        } catch {}
    }
}

Set-Location $InstallDir
& powershell -NoProfile -ExecutionPolicy Bypass -File (Join-Path $InstallDir "setup.ps1")
