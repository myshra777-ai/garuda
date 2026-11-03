# ===========================================================================
# GARUDA ZERO-DEPENDENCY WINDOWS POWERSHELL SETUP (setup.ps1)
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
#      automatically downloads the Go version named in go.mod into
#      $HOME\.garuda\toolchain\go and compiles all 4 binaries in 1 command
#      with CGO_ENABLED=0)
#   4. Runs migrations, seeds owner account, indexes repo, installs Active MCP & Talon
#   5. Starts Garuda API Gateway, runs Doctor & Guard, and opens http://localhost:8080/login
#
# Login: pass -Password, or set GARUDA_PASSWORD, to choose your own. Otherwise a
# random password is generated once and saved to $HOME\.garuda\owner-login.txt.
# No default password is ever used.
#
# NOTE: Keep this file strictly 7-bit ASCII so Windows PowerShell 5.1 (CP1252
# default encoding when no UTF-8 BOM is present) never misinterprets multi-byte
# UTF-8 sequences (0x93/0x94) as smart-quote string delimiters.
# ===========================================================================

param(
    [string]$Workspace = "",
    [string]$Email = "",
    [string]$Password = "",
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
        # Stream child output to the host instead of the caller's pipeline. This
        # prevents assignments from capturing the full installer log as an exit code.
        cmd.exe /c "$CommandLine" 2>&1 | ForEach-Object { Write-Host $_ }
        $commandExit = $LASTEXITCODE
        return $commandExit
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

# Owner password: -Password > GARUDA_PASSWORD > saved from an earlier run (same email) > new random value.
# garuda init does not overwrite an existing user's password, so a re-run must reuse the saved one.
$LoginFile = Join-Path $GarudaHome "owner-login.txt"
$PasswordSource = "you"
if (-not $Password -and $env:GARUDA_PASSWORD) { $Password = $env:GARUDA_PASSWORD }
if (-not $Password -and (Test-Path $LoginFile)) {
    $savedLogin = @(Get-Content $LoginFile)
    if ($savedLogin.Count -ge 2 -and $savedLogin[0] -eq $Email -and $savedLogin[1].Length -ge 8) {
        $Password = $savedLogin[1]
        $PasswordSource = "saved"
    }
}
if (-not $Password) {
    $pwBytes = New-Object byte[] 12
    $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    $rng.GetBytes($pwBytes)
    $rng.Dispose()
    $Password = (($pwBytes | ForEach-Object { $_.ToString("x2") }) -join "")
    Set-Content -Path $LoginFile -Value @($Email, $Password) -Encoding ASCII
    # Limit the file to the current user (best effort; harmless if icacls is unavailable).
    $null = Invoke-QuietCmd "icacls `"$LoginFile`" /inheritance:r /grant:r `"$($env:USERNAME):(R,W)`""
    $PasswordSource = "generated"
}
if ($Password.Length -lt 8) {
    throw "The owner password must be at least 8 characters."
}
# Hand the password to garuda init through the environment (not the command line,
# which other local users can read in the process list).
$env:GARUDA_PASSWORD = $Password

$DbUrl = if ($env:DATABASE_URL) { $env:DATABASE_URL } else { "postgres://test:test@127.0.0.1:5433/garuda_test?sslmode=disable" }
$env:PORT = "$Port"
$env:DATABASE_URL = $DbUrl
$env:CONTROL_DATABASE_URL = $DbUrl
if (-not $env:GARUDA_CONTROL_TOKEN) {
    # No well-known default: generate a random control token. garuda init saves it to .env,
    # so later runs reuse the same value.
    $tokBytes = New-Object byte[] 24
    $tokRng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    $tokRng.GetBytes($tokBytes)
    $tokRng.Dispose()
    $env:GARUDA_CONTROL_TOKEN = "gct_" + (($tokBytes | ForEach-Object { $_.ToString("x2") }) -join "")
}
$env:GARUDA_WORKSPACE = $Workspace
$env:GARUDA_TALON_ENFORCE = "true"
$env:CGO_ENABLED = "0"

Write-Host "===================================================================" -ForegroundColor Cyan
$GarudaVersion = "dev"
if (Test-Path "version.txt") { $GarudaVersion = (Get-Content "version.txt" -TotalCount 1).Trim() }
Write-Host "[GARUDA] $GarudaVersion ZERO-DEPENDENCY WINDOWS SETUP & TALON HARNESS" -ForegroundColor Cyan
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

function Test-PostgresServerReady {
    # A TCP handshake can succeed while PostgreSQL is still starting or recovering.
    $pgIsReady = Join-Path (Join-Path (Join-Path $RuntimeDir "postgres") "pgsql\bin") "pg_isready.exe"
    if (-not (Test-Path $pgIsReady)) {
        $command = Get-Command "pg_isready.exe" -ErrorAction SilentlyContinue
        if ($command) { $pgIsReady = $command.Source }
    }
    if ($pgIsReady -and (Test-Path $pgIsReady)) {
        $readyCode = Invoke-QuietCmd "`"$pgIsReady`" -h 127.0.0.1 -p 5433 -t 1"
        return ($readyCode -eq 0)
    }

    # Fallback only proves that a listener accepts TCP. garuda init still performs
    # the authoritative PostgreSQL protocol/credential check before proceeding.
    return (Test-TcpPort "127.0.0.1" 5433)
}

# Microsoft Visual C++ 2015-2022 runtime (x64). The portable PostgreSQL needs it.
function Test-VCRuntime {
    $keys = @(
        "HKLM:\SOFTWARE\Microsoft\VisualStudio\14.0\VC\Runtimes\x64",
        "HKLM:\SOFTWARE\WOW6432Node\Microsoft\VisualStudio\14.0\VC\Runtimes\x64"
    )
    foreach ($k in $keys) {
        $v = Get-ItemProperty -Path $k -ErrorAction SilentlyContinue
        if ($v -and $v.Installed -eq 1) { return $true }
    }
    return (Test-Path (Join-Path $env:SystemRoot "System32\vcruntime140_1.dll"))
}

function Ensure-VCRuntime {
    if (Test-VCRuntime) { return }
    Write-Host "  [*] Windows needs the Microsoft Visual C++ runtime (a small, standard Microsoft download)." -ForegroundColor Cyan
    Write-Host "  [*] Installing it now. Windows will ask for permission: please click Yes."
    $vcExe = Join-Path $RuntimeDir "vc_redist.x64.exe"
    try {
        Download-FileFast "https://aka.ms/vs/17/release/vc_redist.x64.exe" $vcExe
        $proc = Start-Process -FilePath $vcExe -ArgumentList "/install", "/quiet", "/norestart" -Verb RunAs -Wait -PassThru
        # 0 = installed, 3010 = installed (restart suggested), 1638 = a newer version is already present
        if (@(0, 3010, 1638) -contains $proc.ExitCode) {
            Write-Host "  [OK] Visual C++ runtime installed."
        } else {
            Write-Host "  [!] The Visual C++ installer returned code $($proc.ExitCode)." -ForegroundColor Yellow
        }
    } catch {
        Write-Host "  [!] Could not install the Visual C++ runtime automatically: $($_.Exception.Message)" -ForegroundColor Yellow
        Write-Host "      Install it by hand from https://aka.ms/vs/17/release/vc_redist.x64.exe and re-run this script."
    } finally {
        Remove-Item $vcExe -Force -ErrorAction SilentlyContinue
    }
}

# 1. Stop any stale garuda-api process on $Port
Write-Host ""
Write-Host "[1/7] Checking port $Port..." -ForegroundColor Yellow
Get-Process -Name "garuda-api" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
Start-Sleep -Milliseconds 500
if (Test-TcpPort "127.0.0.1" $Port) {
    throw "Port $Port is already in use by another program (not a Garuda API). Re-run with a free port, for example: .\setup.ps1 -Port 8081"
}

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
            if (Test-PostgresServerReady) { break }
            Start-Sleep -Seconds 1
        }
    }

    if (-not (Test-TcpPort "127.0.0.1" 5433)) {
        Write-Host "  [*] Bootstrapping Portable Zero-Admin PostgreSQL 16..." -ForegroundColor Cyan
        $PgDir = Join-Path $RuntimeDir "postgres"
        $PgBinDir = Join-Path $PgDir "pgsql\bin"
        $PgDataDir = Join-Path $DataDir "pg"
        $PgLog = Join-Path $DataDir "postgres.log"

        Ensure-VCRuntime

        if (-not (Test-Path (Join-Path $PgBinDir "pg_ctl.exe"))) {
            $PgZip = Join-Path $RuntimeDir "postgresql-16-win64.zip"
            $PgUrl = "https://get.enterprisedb.com/postgresql/postgresql-16.4-1-windows-x64-binaries.zip"
            Write-Host "  [*] Downloading Portable PostgreSQL 16 (about 320 MB, one-time setup)..."
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
                $vcHint = ""
                if ($initCode -eq -1073741515) {
                    $vcHint = " Windows reports a missing DLL: install the Microsoft Visual C++ Redistributable (https://aka.ms/vs/17/release/vc_redist.x64.exe) and re-run."
                }
                throw "Failed to initialize local PostgreSQL data directory at $PgDataDir (initdb exit code $initCode).$vcHint"
            }
        }

        $stalePid = Join-Path $PgDataDir "postmaster.pid"
        $pgCtlExe = Join-Path $PgBinDir "pg_ctl.exe"
        $existingPgStatus = Invoke-QuietCmd "`"$pgCtlExe`" -D `"$PgDataDir`" status"
        if (($existingPgStatus -eq 0) -and (-not (Test-TcpPort "127.0.0.1" 5433))) {
            if (Test-Path $PgLog) {
                Write-Host "  --- Last 100 lines of PostgreSQL log ---" -ForegroundColor Red
                Get-Content $PgLog -Tail 100 | ForEach-Object { Write-Host $_ }
            }
            throw "PostgreSQL reports that the cluster is running but port 5433 is not accepting connections. The installer will not delete a live postmaster.pid; inspect $PgLog."
        }
        if (($existingPgStatus -ne 0) -and (Test-Path $stalePid)) {
            Remove-Item -Path $stalePid -Force -ErrorAction SilentlyContinue
        }

        Write-Host "  [*] Starting local PostgreSQL server on port 5433..."
        $pgCtlExe = Join-Path $PgBinDir "pg_ctl.exe"
        $postgresExe = Join-Path $PgBinDir "postgres.exe"
        # Start the server directly and detached. Going through "pg_ctl start" inside
        # cmd.exe blocked forever on a real Windows PC even though the server was
        # already accepting connections, so the script never moved on.
        $pgArgs = @("-D", "`"$PgDataDir`"", "-p", "5433", "-c", "listen_addresses=127.0.0.1")
        Start-Process -FilePath $postgresExe -ArgumentList $pgArgs -WindowStyle Hidden -RedirectStandardOutput "$PgLog.out" -RedirectStandardError $PgLog
        Write-Host -NoNewline "  [*] Waiting for PostgreSQL to accept connections"
        for ($i = 0; $i -lt 90; $i++) {
            if (Test-PostgresServerReady) { break }
            Write-Host -NoNewline "."
            Start-Sleep -Seconds 1
        }
        Write-Host ""
        $createDbExe = Join-Path $PgBinDir "createdb.exe"
        $null = Invoke-QuietCmd "`"$createDbExe`" -h 127.0.0.1 -p 5433 -U test garuda_test"
    }
}

$postgresReady = $false
for ($i = 0; $i -lt 50; $i++) {
    if (Test-PostgresServerReady) { $postgresReady = $true; break }
    Start-Sleep -Milliseconds 500
}
if (-not $postgresReady) {
    $pgLog = Join-Path $DataDir "postgres.log"
    if (Test-Path $pgLog) {
        Write-Host "  --- Last 100 lines of PostgreSQL log ---" -ForegroundColor Red
        Get-Content $pgLog -Tail 100 | ForEach-Object { Write-Host $_ }
    }
    throw "PostgreSQL did not become ready on 127.0.0.1:5433. Check $pgLog for details."
}
Write-Host "  [OK] PostgreSQL listener probe passed on 127.0.0.1:5433; garuda init verifies database connectivity" -ForegroundColor Green

# 3. Ensure Go Toolchain & Build Garuda Binaries in 1 Command (Zero manual install required)
Write-Host "[3/7] Preparing Garuda binaries (garuda.exe, garuda-api.exe, garuda-mcp.exe, garuda-talon.exe)..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path "bin" | Out-Null

$PortableGoBin = Join-Path $ToolchainDir "go\bin"
if (Test-Path (Join-Path $PortableGoBin "go.exe")) {
    $env:PATH = "$PortableGoBin;$env:PATH"
}

function Test-GoNewEnough {
    # Go 1.21+ can fetch the exact toolchain named in go.mod by itself; older Go cannot build this repo.
    $goCmd = Get-Command go -ErrorAction SilentlyContinue
    if (-not $goCmd) { return $false }
    $oldPref = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        $goOut = "$(& $goCmd.Source version 2>$null)"
        if ($goOut -match 'go(\d+)\.(\d+)') {
            return (([int]$Matches[1] -gt 1) -or ([int]$Matches[2] -ge 21))
        }
    } catch {} finally {
        $ErrorActionPreference = $oldPref
    }
    return $false
}

if (-not (Test-GoNewEnough)) {
    # Download the Go version named in go.mod (toolchain line first, then the go line) so Go never has to
    # download a second toolchain while compiling.
    $goCandidates = @()
    if (Test-Path "go.mod") {
        $goModText = Get-Content "go.mod" -Raw
        $tcMatch = [regex]::Match($goModText, '(?m)^toolchain\s+go(\d+\.\d+(?:\.\d+)?)')
        if ($tcMatch.Success) { $goCandidates += $tcMatch.Groups[1].Value }
        $goMatch = [regex]::Match($goModText, '(?m)^go\s+(\d+\.\d+(?:\.\d+)?)')
        if ($goMatch.Success) { $goCandidates += $goMatch.Groups[1].Value }
    }
    $goCandidates += "1.26.0"
    $goCandidates = @($goCandidates | Select-Object -Unique)

    Write-Host "  [*] Go 1.21 or newer not found -- downloading Portable Go into $ToolchainDir\go (no Admin required)..." -ForegroundColor Cyan
    $Arch = if ($env:PROCESSOR_ARCHITECTURE -match "ARM64") { "arm64" } else { "amd64" }
    $goDownloaded = $false
    foreach ($goVer in $goCandidates) {
        $GoZip = Join-Path $ToolchainDir "go$goVer.windows-$Arch.zip"
        $GoUrl = "https://go.dev/dl/go$goVer.windows-$Arch.zip"
        try {
            Write-Host "  [*] Trying Go $goVer ..."
            Download-FileFast $GoUrl $GoZip
            $goDownloaded = $true
            break
        } catch {
            Write-Host "  [WARN] Go $goVer was not downloadable: $($_.Exception.Message)" -ForegroundColor Yellow
            Remove-Item $GoZip -Force -ErrorAction SilentlyContinue
        }
    }
    if (-not $goDownloaded) {
        throw "Could not download Portable Go from go.dev. Check your internet/proxy, or install Go from https://go.dev/dl/ and re-run."
    }
    Write-Host "  [*] Extracting Portable Go toolchain..."
    Remove-Item (Join-Path $ToolchainDir "go") -Recurse -Force -ErrorAction SilentlyContinue
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

# Windows can refuse to run a program it does not trust (Smart App Control, Windows Defender
# Application Control / Device Guard, or a work or school policy). Garuda's programs are not
# code-signed yet. Detect that now and explain it, instead of failing later with a log dump.
$smokeCode = Invoke-QuietCmd "`".\bin\garuda.exe`" --version"
if ($smokeCode -eq 4551) {
    Write-Host ""
    Write-Host "Windows blocked the Garuda program that was just built (error 4551)." -ForegroundColor Red
    Write-Host "This is Windows security, not a Garuda bug: Smart App Control, Device Guard or an"
    Write-Host "organization policy refuses to run programs that are not code-signed."
    Write-Host ""
    Write-Host "Choose one:"
    Write-Host "  1. Use WSL2 (no security change): run 'wsl --install -d Ubuntu', restart, then follow the Linux"
    Write-Host "     steps in INSTALL.md inside Ubuntu. Open http://localhost:8080/login in your Windows browser."
    Write-Host "  2. Personal PC only: Windows Security > App and browser control > Smart App Control settings > Off."
    Write-Host "     Windows cannot turn Smart App Control back on without reinstalling Windows."
    Write-Host "  3. Work or school PC: ask your IT team to allow $RepoRoot\bin."
    throw "Windows blocked bin\garuda.exe by policy (exit code 4551). See the options above."
}

# 4. Run 1-Command Initialization (Migrations, .env, Owner Account, AST Scan, Active MCP & Talon Hooks)
Write-Host "[4/7] Running migrations, provisioning account, indexing workspace, and activating Garuda Talon..." -ForegroundColor Yellow
$initExit = Invoke-LoggedCmd "`".\bin\garuda.exe`" init --workspace `"$Workspace`" --email `"$Email`" --skip-scan"
if ($initExit -ne 0) {
    $pgLog = Join-Path $DataDir "postgres.log"
    if (Test-Path $pgLog) {
        Write-Host "  --- Last 100 lines of PostgreSQL log ---" -ForegroundColor Red
        Get-Content $pgLog -Tail 100 | ForEach-Object { Write-Host $_ }
    }
    throw "garuda.exe init failed with exit code $initExit. PostgreSQL logs (if present) were printed above; log path: $pgLog."
}

# 5. Start Garuda API Gateway in Background
Write-Host "[5/7] Starting Garuda API Gateway on http://localhost:$Port ..." -ForegroundColor Yellow
$ApiLog = Join-Path $DataDir "garuda-api.log"
$ApiErrLog = Join-Path $DataDir "garuda-api.err.log"
Start-Process -FilePath (Join-Path $RepoRoot "bin\garuda-api.exe") -WorkingDirectory $RepoRoot -RedirectStandardOutput $ApiLog -RedirectStandardError $ApiErrLog -WindowStyle Hidden

# Probe 127.0.0.1, not "localhost": Windows can resolve localhost to IPv6 (::1) while
# the API listens on IPv4 only. Skip any system proxy for this local request.
function Test-ApiHealthy([int]$PortNumber) {
    try {
        $req = [System.Net.WebRequest]::Create("http://127.0.0.1:$PortNumber/health")
        $req.Proxy = $null
        $req.Timeout = 2000
        $resp = $req.GetResponse()
        $ok = ([int]$resp.StatusCode -eq 200)
        $resp.Close()
        return $ok
    } catch {
        return $false
    }
}

$apiHealthy = $false
for ($i = 0; $i -lt 40; $i++) {
    if (Test-ApiHealthy $Port) { $apiHealthy = $true; break }
    Start-Sleep -Milliseconds 500
}
if (-not $apiHealthy) {
    if (Test-Path $ApiErrLog) {
        Write-Host "  --- API stderr ---" -ForegroundColor Red
        Get-Content $ApiErrLog -Tail 40 | ForEach-Object { Write-Host $_ }
    }
    if (Test-Path $ApiLog) {
        Write-Host "  --- API stdout ---" -ForegroundColor Red
        Get-Content $ApiLog -Tail 40 | ForEach-Object { Write-Host $_ }
    }
    throw "Garuda API did not become healthy at http://127.0.0.1:$Port/health. Review logs under $DataDir."
}

# 5b. Index the installed repository in a separate CLI process rather than
# recursively invoking the root command from within init.
Write-Host "[5b/7] Indexing repository source into workspace '$Workspace'..." -ForegroundColor Yellow
$analyzeExit = Invoke-LoggedCmd "`".\bin\garuda.exe`" analyze `"$RepoRoot`" --workspace `"$Workspace`" --save"
if ($analyzeExit -ne 0) {
    $pgLog = Join-Path $DataDir "postgres.log"
    if (Test-Path $pgLog) {
        Write-Host "  --- Last 100 lines of PostgreSQL log ---" -ForegroundColor Red
        Get-Content $pgLog -Tail 100 | ForEach-Object { Write-Host $_ }
    }
    throw "Initial repository analysis failed with exit code $analyzeExit. PostgreSQL logs (if present) were printed above; log path: $pgLog."
}

# 6. Run Architectural Summary & Talon Guard (without redundant re-index)
Write-Host "[6/7] Running Architectural Summary & Talon Guard..." -ForegroundColor Yellow
$summaryExit = Invoke-LoggedCmd "`".\bin\garuda.exe`" summary"
if ($summaryExit -ne 0) { throw "Garuda summary failed with exit code $summaryExit." }
$guardExit = Invoke-LoggedCmd "`".\bin\garuda.exe`" talon guard --workspace `"$Workspace`" --policies policies --auto-index=false"
if ($guardExit -ne 0) { Write-Host "  [WARN] Talon guard returned exit code $guardExit. Review findings above; policy blocks are not hidden." -ForegroundColor Yellow }

# 7. Run Garuda Doctor & Open Browser
Write-Host "[7/7] Running Garuda System Doctor..." -ForegroundColor Yellow
$doctorExit = Invoke-LoggedCmd "`".\bin\garuda.exe`" doctor"
if ($doctorExit -ne 0) { Write-Host "  [WARN] System Doctor returned exit code $doctorExit. Review diagnostics above." -ForegroundColor Yellow }

Write-Host ""
Write-Host "===================================================================" -ForegroundColor Green
Write-Host "[OK] GARUDA $GarudaVersion + TALON SETUP COMPLETED ON WINDOWS (review warnings above)" -ForegroundColor Green
Write-Host "   * Web Dashboard:  http://localhost:$Port/login"
if ($PasswordSource -eq "you") {
    Write-Host "   * Login:          $Email (the password you chose)"
} else {
    Write-Host "   * Login:          $Email / $Password"
    Write-Host "                     (saved in $LoginFile; set GARUDA_PASSWORD to choose your own)"
}
Write-Host "   * System Doctor:  .\bin\garuda.exe doctor"
Write-Host "   * Talon Guard:    .\bin\garuda.exe talon guard --workspace $Workspace"
Write-Host "===================================================================" -ForegroundColor Green

try {
    Start-Process "http://localhost:$Port/login"
} catch {}
