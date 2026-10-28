# ===========================================================================
# GARUDA v1.9.0 ZERO-DEPENDENCY WINDOWS POWERSHELL SETUP (setup.ps1)
# ===========================================================================
# Usage (from Windows PowerShell 5.1 or PowerShell 7 -- no Admin, Go,
# PostgreSQL, GCC, or Docker required):
#   powershell -NoProfile -ExecutionPolicy Bypass -File .\setup.ps1
#
# What this script manages automatically:
#   1. Frees port 8080 if a previous Garuda instance is running
#   2. Ensures PostgreSQL is running on 127.0.0.1:5433 (uses Docker if running,
#      or automatically downloads & starts Portable Zero-Admin PostgreSQL 16 in
#      $HOME\.garuda\runtime\postgres + cluster in $HOME\.garuda\data\pg)
#   3. Ensures Garuda binaries exist in .\bin\ (uses Go if installed, or
#      automatically downloads Portable Go 1.26 into $HOME\.garuda\toolchain\go
#      and compiles all 4 binaries in 1 command with CGO_ENABLED=0)
#   4. Runs migrations, seeds owner account, indexes repo, installs Active MCP & Talon
#   5. Starts Garuda API Gateway, runs Doctor & Guard, and opens http://localhost:8080/login
#
# NOTE: Keep this file strictly 7-bit ASCII so Windows PowerShell 5.1 (CP1252
# default encoding when no UTF-8 BOM is present) never misinterprets multi-byte
# UTF-8 sequences (0x93/0x94) as smart-quote string delimiters.
# ===========================================================================

param(
    [string]$Workspace = "",
    [string]$Email = "",
    [string]$Password = "garuda2026",
    [int]$Port = 8080
)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

function Invoke-QuietCmd([string]$CommandLine) {
    $oldPref = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        cmd.exe /c "$CommandLine >nul 2>&1"
        return $LASTEXITCODE
    } finally {
        $ErrorActionPreference = $oldPref
    }
}

function Invoke-LoggedCmd([string]$CommandLine) {
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

$RepoRoot = if ($PSScriptRoot) { $PSScriptRoot } elseif ($MyInvocation.MyCommand.Definition) { Split-Path -Parent $MyInvocation.MyCommand.Definition } else { (Get-Location).Path }
Set-Location $RepoRoot

$GarudaHome = Join-Path $HOME ".garuda"
$ToolchainDir = Join-Path $GarudaHome "toolchain"
$RuntimeDir = Join-Path $GarudaHome "runtime"
$DataDir = Join-Path $GarudaHome "data"
New-Item -ItemType Directory -Force -Path $ToolchainDir, $RuntimeDir, $DataDir | Out-Null

# Load existing .env if present
if (Test-Path ".env") {
    Get-Content ".env" | ForEach-Object {
        if ($_ -match "^\s*([^#][^=]+)=(.*)$") {
            [System.Environment]::SetEnvironmentVariable($matches[1].Trim(), $matches[2].Trim(), "Process")
        }
    }
}

if (-not $Email) {
    if ($env:GARUDA_EMAIL) {
        $Email = $env:GARUDA_EMAIL
    } elseif (Get-Command git -ErrorAction SilentlyContinue) {
        $oldPref = $ErrorActionPreference
        $ErrorActionPreference = "Continue"
        try {
            $gitEmail = (& git config user.email 2>$null)
            if ($LASTEXITCODE -eq 0 -and $gitEmail -and ("$gitEmail").Contains("@")) {
                $Email = ("$gitEmail").Trim()
            }
        } catch {} finally {
            $ErrorActionPreference = $oldPref
        }
    }
    if (-not $Email) {
        $user = if ($env:USERNAME) { $env:USERNAME.ToLower() -replace '\s+','.' } else { "admin" }
        $Email = "$user@garuda.local"
    }
}

if (-not $Workspace) {
    if ($env:GARUDA_WORKSPACE) {
        $Workspace = $env:GARUDA_WORKSPACE
    } else {
        $Workspace = Split-Path -Leaf $RepoRoot
    }
}

$DbUrl = if ($env:DATABASE_URL) { $env:DATABASE_URL } else { "postgres://test:test@127.0.0.1:5433/garuda_test?sslmode=disable" }
$env:PORT = "$Port"
$env:DATABASE_URL = $DbUrl
$env:CONTROL_DATABASE_URL = $DbUrl
$env:GARUDA_CONTROL_TOKEN = if ($env:GARUDA_CONTROL_TOKEN) { $env:GARUDA_CONTROL_TOKEN } else { "devtoken-changeme-please" }
$env:GARUDA_WORKSPACE = $Workspace
$env:GARUDA_TALON_ENFORCE = "true"
$env:CGO_ENABLED = "0"

Write-Host "===================================================================" -ForegroundColor Cyan
Write-Host "[GARUDA] v1.9.0 ZERO-DEPENDENCY WINDOWS SETUP & TALON HARNESS" -ForegroundColor Cyan
Write-Host "===================================================================" -ForegroundColor Cyan
Write-Host "  * Detected Owner:     $Email"
Write-Host "  * Target Workspace:   $Workspace"
Write-Host "  * Gateway Port:       :$Port"
Write-Host "  * Pure-Go Build Mode: CGO_ENABLED=0 (No GCC/MinGW required)"

function Test-TcpPort([string]$HostName, [int]$PortNumber) {
    try {
        $client = New-Object System.Net.Sockets.TcpClient
        $iar = $client.BeginConnect($HostName, $PortNumber, $null, $null)
        $wait = $iar.AsyncWaitHandle.WaitOne(800, $false)
        if ($wait -and $client.Connected) {
            $client.Close()
            return $true
        }
        $client.Close()
    } catch {}
    return $false
}

# 1. Stop any stale garuda-api process on $Port
Write-Host ""
Write-Host "[1/7] Checking port $Port..." -ForegroundColor Yellow
Get-Process -Name "garuda-api" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue

# 2. Ensure PostgreSQL is running on port 5433 (Docker or Portable Zero-Admin PostgreSQL)
Write-Host "[2/7] Verifying PostgreSQL database on 127.0.0.1:5433..." -ForegroundColor Yellow
if (-not (Test-TcpPort "127.0.0.1" 5433)) {
    $dockerReady = $false
    if (Get-Command docker -ErrorAction SilentlyContinue) {
        $exitCode = Invoke-QuietCmd "docker info"
        if ($exitCode -eq 0) {
            $dockerReady = $true
        }
    }

    if ($dockerReady) {
        Write-Host "  [*] Docker detected -- starting container 'garuda-postgres' on :5433..."
        $startCode = Invoke-QuietCmd "docker start garuda-postgres"
        if ($startCode -ne 0) {
            $null = Invoke-QuietCmd "docker run -d --name garuda-postgres -e POSTGRES_USER=test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=garuda_test -p 5433:5432 postgres:16-alpine"
        }
        for ($i = 0; $i -lt 25; $i++) {
            if (Test-TcpPort "127.0.0.1" 5433) { break }
            Start-Sleep -Seconds 1
        }
    }

    if (-not (Test-TcpPort "127.0.0.1" 5433)) {
        Write-Host "  [*] Bootstrapping Portable Zero-Admin PostgreSQL 16..." -ForegroundColor Cyan
        $PgDir = Join-Path $RuntimeDir "postgres"
        $PgBinDir = Join-Path $PgDir "pgsql\bin"
        $PgDataDir = Join-Path $DataDir "pg"
        $PgLog = Join-Path $DataDir "postgres.log"

        if (-not (Test-Path (Join-Path $PgBinDir "pg_ctl.exe"))) {
            $PgZip = Join-Path $RuntimeDir "postgresql-16-win64.zip"
            $PgUrl = "https://get.enterprisedb.com/postgresql/postgresql-16.4-1-windows-x64-binaries.zip"
            Write-Host "  [*] Downloading Portable PostgreSQL 16 (~55 MB, one-time setup)..."
            Download-FileFast $PgUrl $PgZip
            Write-Host "  [*] Extracting Portable PostgreSQL to $PgDir..."
            Extract-ArchiveFast $PgZip $PgDir
            Remove-Item $PgZip -Force -ErrorAction SilentlyContinue
        }

        $env:PATH = "$PgBinDir;$env:PATH"
        if (-not (Test-Path (Join-Path $PgDataDir "PG_VERSION"))) {
            if (Test-Path $PgDataDir) {
                Remove-Item -Path $PgDataDir -Recurse -Force -ErrorAction SilentlyContinue
            }
            Write-Host "  [*] Initializing local PostgreSQL cluster in $PgDataDir..."
            $initDbExe = Join-Path $PgBinDir "initdb.exe"
            $initCode = Invoke-QuietCmd "`"$initDbExe`" -D `"$PgDataDir`" -U test -A trust -E UTF8 --no-locale"
            if ($initCode -ne 0) {
                throw "Failed to initialize local PostgreSQL data directory at $PgDataDir"
            }
        }

        $stalePid = Join-Path $PgDataDir "postmaster.pid"
        if (Test-Path $stalePid) {
            Remove-Item -Path $stalePid -Force -ErrorAction SilentlyContinue
        }

        Write-Host "  [*] Starting local PostgreSQL server on port 5433..."
        $pgCtlExe = Join-Path $PgBinDir "pg_ctl.exe"
        $null = Invoke-QuietCmd "`"$pgCtlExe`" -D `"$PgDataDir`" -l `"$PgLog`" -o `"-p 5433`" start"
        for ($i = 0; $i -lt 25; $i++) {
            if (Test-TcpPort "127.0.0.1" 5433) { break }
            Start-Sleep -Seconds 1
        }
        $pgIsReadyExe = Join-Path $PgBinDir "pg_isready.exe"
        for ($i = 0; $i -lt 25; $i++) {
            if ((Invoke-QuietCmd "`"$pgIsReadyExe`" -h 127.0.0.1 -p 5433") -eq 0) { break }
            Start-Sleep -Milliseconds 400
        }
        $psqlExe = Join-Path $PgBinDir "psql.exe"
        $createDbExe = Join-Path $PgBinDir "createdb.exe"
        $null = Invoke-QuietCmd "`"$psqlExe`" -h 127.0.0.1 -p 5433 -U test -d postgres -c `"ALTER ROLE test WITH SUPERUSER LOGIN PASSWORD 'test';`""
        $null = Invoke-QuietCmd "`"$createDbExe`" -h 127.0.0.1 -p 5433 -U test garuda_test"
        $null = Invoke-QuietCmd "`"$psqlExe`" -h 127.0.0.1 -p 5433 -U test -d postgres -c `"CREATE DATABASE garuda_test OWNER test;`""
    }
}

if (-not (Test-TcpPort "127.0.0.1" 5433)) {
    throw "PostgreSQL could not be started on 127.0.0.1:5433. Check $DataDir\postgres.log for details."
}
Write-Host "  [OK] PostgreSQL is online and ready on 127.0.0.1:5433" -ForegroundColor Green

# 3. Ensure Go Toolchain & Build Garuda Binaries in 1 Command (Zero manual install required)
Write-Host "[3/7] Preparing Garuda binaries (garuda.exe, garuda-api.exe, garuda-mcp.exe, garuda-talon.exe)..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path "bin" | Out-Null

$PortableGoBin = Join-Path $ToolchainDir "go\bin"
if (Test-Path (Join-Path $PortableGoBin "go.exe")) {
    $env:PATH = "$PortableGoBin;$env:PATH"
}

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "  [*] Go not installed -- downloading Portable Go 1.26 into $ToolchainDir\go (no Admin required)..." -ForegroundColor Cyan
    $Arch = if ($env:PROCESSOR_ARCHITECTURE -match "ARM64") { "arm64" } else { "amd64" }
    $GoZip = Join-Path $ToolchainDir "go1.26.0.windows-$Arch.zip"
    $GoUrl = "https://go.dev/dl/go1.26.0.windows-$Arch.zip"
    Download-FileFast $GoUrl $GoZip
    Write-Host "  [*] Extracting Portable Go toolchain..."
    Extract-ArchiveFast $GoZip $ToolchainDir
    Remove-Item $GoZip -Force -ErrorAction SilentlyContinue
    $env:PATH = "$PortableGoBin;$env:PATH"
}

Write-Host "  [*] Compiling all 4 Garuda binaries in 1 pass with CGO_ENABLED=0..."
$bCode = Invoke-LoggedCmd "go build -o bin\ .\cmd\garuda .\cmd\garuda-api .\cmd\garuda-mcp .\cmd\garuda-talon"
if ($bCode -ne 0 -or (-not (Test-Path "bin\garuda.exe"))) {
    throw "Failed to compile Garuda binaries into .\bin\"
}
Write-Host "  [OK] Built bin\garuda.exe, bin\garuda-api.exe, bin\garuda-mcp.exe, bin\garuda-talon.exe" -ForegroundColor Green

# 4. Run 1-Command Initialization (Migrations, .env, Owner Account, AST Scan, Active MCP & Talon Hooks)
Write-Host "[4/7] Running migrations, provisioning account, indexing workspace, and activating Garuda Talon..." -ForegroundColor Yellow
$initExit = Invoke-LoggedCmd "`".\bin\garuda.exe`" init --workspace `"$Workspace`" --email `"$Email`" --password `"$Password`""
if ($initExit -ne 0) {
    throw "garuda.exe init failed with exit code $initExit"
}

# 5. Start Garuda API Gateway in Background
Write-Host "[5/7] Starting Garuda API Gateway on http://localhost:$Port ..." -ForegroundColor Yellow
$ApiLog = Join-Path $DataDir "garuda-api.log"
$ApiErrLog = Join-Path $DataDir "garuda-api.err.log"
Start-Process -FilePath (Join-Path $RepoRoot "bin\garuda-api.exe") -WorkingDirectory $RepoRoot -RedirectStandardOutput $ApiLog -RedirectStandardError $ApiErrLog -WindowStyle Hidden

for ($i = 0; $i -lt 25; $i++) {
    try {
        $resp = Invoke-WebRequest -Uri "http://localhost:$Port/health" -UseBasicParsing -TimeoutSec 2
        if ($resp.StatusCode -eq 200) { break }
    } catch {}
    Start-Sleep -Milliseconds 400
}

# 6. Run Architectural Summary & Talon Guard (without redundant re-index)
Write-Host "[6/7] Running Architectural Summary & Talon Guard..." -ForegroundColor Yellow
$null = Invoke-LoggedCmd "`".\bin\garuda.exe`" summary"
$null = Invoke-LoggedCmd "`".\bin\garuda.exe`" talon guard --workspace `"$Workspace`" --policies policies --auto-index=false"

# 7. Run Garuda Doctor & Open Browser
Write-Host "[7/7] Running Garuda System Doctor..." -ForegroundColor Yellow
$null = Invoke-LoggedCmd "`".\bin\garuda.exe`" doctor"

Write-Host ""
Write-Host "===================================================================" -ForegroundColor Green
Write-Host "[OK] GARUDA v1.9.0 + TALON IS 100% CONFIGURED & ONLINE ON WINDOWS!" -ForegroundColor Green
Write-Host "   * Web Dashboard:  http://localhost:$Port/login"
Write-Host "   * Email / Pass:   $Email / $Password"
Write-Host "   * System Doctor:  .\bin\garuda.exe doctor"
Write-Host "   * Talon Guard:    .\bin\garuda.exe talon guard --workspace $Workspace"
Write-Host "===================================================================" -ForegroundColor Green

try {
    Start-Process "http://localhost:$Port/login"
} catch {}
