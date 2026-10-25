# ═══════════════════════════════════════════════════════════════════════════
# 🦅 GARUDA v1.6 ZERO-DEPENDENCY WINDOWS POWERSHELL SETUP (setup.ps1)
# ═══════════════════════════════════════════════════════════════════════════
# Usage (from Windows PowerShell — no Admin, Go, Postgres, GCC, or Docker required):
#   powershell -ExecutionPolicy Bypass -File .\setup.ps1
#
# What this script manages automatically:
#   1. Frees port 8080 if a previous Garuda instance is running
#   2. Ensures PostgreSQL is running on :5433 (uses Docker if running, or
#      automatically downloads & starts Portable Zero-Admin PostgreSQL 16 in
#      $HOME\.garuda\runtime\postgres + cluster in $HOME\.garuda\data\pg)
#   3. Ensures Garuda binaries exist in .\bin\ (uses prebuilt binaries if present,
#      or automatically downloads Portable Go 1.24 into $HOME\.garuda\toolchain\go
#      and compiles with CGO_ENABLED=0 — zero GCC/C compiler required!)
#   4. Runs migrations, seeds owner account, installs Active MCP & Talon hooks
#   5. Starts Garuda API Gateway and opens http://localhost:8080/login
# ═══════════════════════════════════════════════════════════════════════════

param(
    [string]$Workspace = "",
    [string]$Email = "",
    [string]$Password = "garuda2026",
    [int]$Port = 8080
)

$ErrorActionPreference = "Stop"
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$RepoRoot = Split-Path -Parent $MyInvocation.MyCommand.Definition
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
    } else {
        try {
            $gitEmail = (git config user.email 2>$null)
            if ($gitEmail -and $gitEmail.Contains("@")) { $Email = $gitEmail.Trim() }
        } catch {}
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

$DbUrl = if ($env:DATABASE_URL) { $env:DATABASE_URL } else { "postgres://test:test@localhost:5433/garuda_test?sslmode=disable" }
$env:PORT = "$Port"
$env:DATABASE_URL = $DbUrl
$env:CONTROL_DATABASE_URL = $DbUrl
$env:GARUDA_CONTROL_TOKEN = if ($env:GARUDA_CONTROL_TOKEN) { $env:GARUDA_CONTROL_TOKEN } else { "devtoken-changeme-please" }
$env:GARUDA_WORKSPACE = $Workspace
$env:GARUDA_TALON_ENFORCE = "true"
$env:CGO_ENABLED = "0"

Write-Host "===================================================================" -ForegroundColor Cyan
Write-Host "🦅 GARUDA v1.6 ZERO-DEPENDENCY WINDOWS SETUP & TALON HARNESS" -ForegroundColor Cyan
Write-Host "===================================================================" -ForegroundColor Cyan
Write-Host "  • Detected Owner:     $Email"
Write-Host "  • Target Workspace:   $Workspace"
Write-Host "  • Gateway Port:       :$Port"
Write-Host "  • Pure-Go Build Mode: CGO_ENABLED=0 (No GCC/MinGW required)"

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
Write-Host "`n[1/6] Checking port $Port..." -ForegroundColor Yellow
Get-Process -Name "garuda-api" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue

# 2. Ensure PostgreSQL is running on port 5433 (Docker or Portable Zero-Admin PostgreSQL)
Write-Host "[2/6] Verifying PostgreSQL database on :5433..." -ForegroundColor Yellow
if (-not (Test-TcpPort "127.0.0.1" 5433)) {
    $dockerReady = $false
    if (Get-Command docker -ErrorAction SilentlyContinue) {
        try {
            docker info *> $null
            if ($LASTEXITCODE -eq 0) { $dockerReady = $true }
        } catch {}
    }

    if ($dockerReady) {
        Write-Host "  ⏳ Docker detected — starting container 'garuda-postgres' on :5433..."
        docker start garuda-postgres *> $null
        if ($LASTEXITCODE -ne 0) {
            docker run -d --name garuda-postgres `
                -e POSTGRES_USER=test `
                -e POSTGRES_PASSWORD=test `
                -e POSTGRES_DB=garuda_test `
                -p 5433:5432 `
                postgres:16-alpine *> $null
        }
        for ($i = 0; $i -lt 25; $i++) {
            if (Test-TcpPort "127.0.0.1" 5433) { break }
            Start-Sleep -Seconds 1
        }
    } else {
        Write-Host "  ⏳ No Docker daemon detected — bootstrapping Portable Zero-Admin PostgreSQL 16..." -ForegroundColor Cyan
        $PgDir = Join-Path $RuntimeDir "postgres"
        $PgBinDir = Join-Path $PgDir "pgsql\bin"
        $PgDataDir = Join-Path $DataDir "pg"
        $PgLog = Join-Path $DataDir "postgres.log"

        if (-not (Test-Path (Join-Path $PgBinDir "pg_ctl.exe"))) {
            $PgZip = Join-Path $RuntimeDir "postgresql-16-win64.zip"
            $PgUrl = "https://get.enterprisedb.com/postgresql/postgresql-16.4-1-windows-x64-binaries.zip"
            Write-Host "  ⬇️ Downloading Portable PostgreSQL 16 (~55 MB, one-time setup)..."
            Invoke-WebRequest -Uri $PgUrl -OutFile $PgZip -UseBasicParsing
            Write-Host "  📦 Extracting Portable PostgreSQL to $PgDir..."
            Expand-Archive -Path $PgZip -DestinationPath $PgDir -Force
            Remove-Item $PgZip -Force -ErrorAction SilentlyContinue
        }

        $env:PATH = "$PgBinDir;$env:PATH"
        if (-not (Test-Path (Join-Path $PgDataDir "PG_VERSION"))) {
            Write-Host "  🛠️ Initializing local PostgreSQL cluster in $PgDataDir..."
            & (Join-Path $PgBinDir "initdb.exe") -D $PgDataDir -U test -A trust -E UTF8 --no-locale *> $null
        }

        Write-Host "  🚀 Starting local PostgreSQL server on port 5433..."
        & (Join-Path $PgBinDir "pg_ctl.exe") -D $PgDataDir -l $PgLog -o "-p 5433" start *> $null
        for ($i = 0; $i -lt 20; $i++) {
            if (Test-TcpPort "127.0.0.1" 5433) { break }
            Start-Sleep -Seconds 1
        }
        & (Join-Path $PgBinDir "createdb.exe") -h 127.0.0.1 -p 5433 -U test garuda_test 2>$null
    }
}
Write-Host "  ✓ PostgreSQL is online and ready on :5433" -ForegroundColor Green

# 3. Ensure Go Toolchain or Prebuilt Binaries (Zero manual install required)
Write-Host "[3/6] Preparing Garuda binaries (garuda.exe, garuda-api.exe, garuda-mcp.exe, garuda-talon.exe)..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path "bin" | Out-Null

$NeedBuild = (-not (Test-Path "bin\garuda.exe")) -or (-not (Test-Path "bin\garuda-api.exe")) -or (-not (Test-Path "bin\garuda-mcp.exe")) -or (-not (Test-Path "bin\garuda-talon.exe"))

# Check if Go is installed or in $HOME\.garuda\toolchain\go\bin
$PortableGoBin = Join-Path $ToolchainDir "go\bin"
if (Test-Path (Join-Path $PortableGoBin "go.exe")) {
    $env:PATH = "$PortableGoBin;$env:PATH"
}

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "  ⏳ Go not installed — downloading Portable Go 1.24 into $ToolchainDir\go (no Admin required)..." -ForegroundColor Cyan
    $Arch = if ($env:PROCESSOR_ARCHITECTURE -match "ARM64") { "arm64" } else { "amd64" }
    $GoZip = Join-Path $ToolchainDir "go1.24.0.windows-$Arch.zip"
    $GoUrl = "https://go.dev/dl/go1.24.0.windows-$Arch.zip"
    Invoke-WebRequest -Uri $GoUrl -OutFile $GoZip -UseBasicParsing
    Write-Host "  📦 Extracting Portable Go toolchain..."
    Expand-Archive -Path $GoZip -DestinationPath $ToolchainDir -Force
    Remove-Item $GoZip -Force -ErrorAction SilentlyContinue
    $env:PATH = "$PortableGoBin;$env:PATH"
}

Write-Host "  🔨 Compiling Garuda binaries with CGO_ENABLED=0 (pure Go, no C compiler needed)..."
go build -o bin\garuda.exe .\cmd\garuda
go build -o bin\garuda-api.exe .\cmd\garuda-api
go build -o bin\garuda-mcp.exe .\cmd\garuda-mcp
go build -o bin\garuda-talon.exe .\cmd\garuda-talon
Write-Host "  ✓ Built bin\garuda.exe, bin\garuda-api.exe, bin\garuda-mcp.exe, bin\garuda-talon.exe" -ForegroundColor Green

# 4. Run 1-Command Initialization (Migrations, .env, Owner Account, Active MCP & Talon Hooks)
Write-Host "[4/6] Running migrations, provisioning account, and activating Garuda Talon..." -ForegroundColor Yellow
.\bin\garuda.exe init --workspace $Workspace --email $Email --password $Password

# 5. Start Garuda API Gateway in Background
Write-Host "[5/6] Starting Garuda API Gateway on http://localhost:$Port ..." -ForegroundColor Yellow
$ApiLog = Join-Path $DataDir "garuda-api.log"
$ApiErrLog = Join-Path $DataDir "garuda-api.err.log"
Start-Process -FilePath ".\bin\garuda-api.exe" -WorkingDirectory $RepoRoot -RedirectStandardOutput $ApiLog -RedirectStandardError $ApiErrLog -WindowStyle Hidden

for ($i = 0; $i -lt 20; $i++) {
    try {
        $resp = Invoke-WebRequest -Uri "http://localhost:$Port/health" -UseBasicParsing -TimeoutSec 2
        if ($resp.StatusCode -eq 200) { break }
    } catch {}
    Start-Sleep -Milliseconds 500
}

# 6. Run Garuda Doctor & Open Browser
Write-Host "[6/6] Running Garuda System Doctor..." -ForegroundColor Yellow
.\bin\garuda.exe doctor

Write-Host "`n===================================================================" -ForegroundColor Green
Write-Host "🚀 GARUDA v1.6 + TALON IS 100% CONFIGURED & ONLINE ON WINDOWS!" -ForegroundColor Green
Write-Host "   • Web Dashboard:  http://localhost:$Port/login"
Write-Host "   • Email / Pass:   $Email / $Password"
Write-Host "   • System Doctor:  .\bin\garuda.exe doctor"
Write-Host "   • Talon Guard:    .\bin\garuda.exe talon guard --workspace $Workspace"
Write-Host "===================================================================" -ForegroundColor Green

try {
    Start-Process "http://localhost:$Port/login"
} catch {}
