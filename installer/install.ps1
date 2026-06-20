# install.ps1 - uvpip Windows installer
# Run with: irm https://raw.githubusercontent.com/yv3000/uvpip/main/installer/install.ps1 | iex
# No admin required. Uses User-level PATH only (with System PATH fallback via UAC prompt).

$ErrorActionPreference = "Stop"

$installDir = Join-Path $env:USERPROFILE ".uvpip"
$binDir     = Join-Path $installDir "bin"
$exePath    = Join-Path $binDir "uvpip.exe"
$releaseBase = "https://github.com/yv3000/uvpip/releases/latest/download"

function Write-OK  { param([string]$msg) Write-Host "  [OK] $msg" -ForegroundColor Green }
function Write-WRN { param([string]$msg) Write-Host "  [!!] $msg" -ForegroundColor Yellow }
function Write-ERR { param([string]$msg) Write-Host "  [ERR] $msg" -ForegroundColor Red }
function Write-NFO { param([string]$msg) Write-Host "  [->] $msg" -ForegroundColor Cyan }

Write-Host ""
Write-Host "  uvpip installer for Windows" -ForegroundColor White
Write-Host "  ----------------------------------------" -ForegroundColor DarkGray
Write-Host ""

# --- Step 1: Detect architecture ---------------------------------------------
$arch = $env:PROCESSOR_ARCHITECTURE
$binaryName = "uvpip-windows-amd64.exe"
if ($arch -eq "ARM64") {
    $binaryName = "uvpip-windows-arm64.exe"
}
Write-OK "Architecture: $arch -> $binaryName"

# --- Step 2: Check if already installed --------------------------------------
if (Test-Path $exePath) {
    Write-WRN "uvpip is already installed at $exePath"
    Write-NFO "To reinstall, run the uninstaller first:"
    Write-NFO "irm https://raw.githubusercontent.com/yv3000/uvpip/main/uninstaller/uninstall.ps1 | iex"
    Write-Host ""
    exit 0
}

# --- Step 3: Check / install uv ----------------------------------------------
$uvFound = $false
try {
    $uvVer = & uv --version 2>$null
    if ($LASTEXITCODE -eq 0 -and $uvVer) {
        Write-OK "uv already installed: $uvVer"
        $uvFound = $true
    }
} catch {}

if (-not $uvFound) {
    Write-NFO "uv not found. Installing uv automatically..."
    try {
        Invoke-RestMethod https://astral.sh/uv/install.ps1 | Invoke-Expression
        # Refresh PATH in current session so uv is available
        $uvBinDir = Join-Path $env:USERPROFILE ".local\bin"
        if (Test-Path $uvBinDir) {
            $env:PATH = "$uvBinDir;$env:PATH"
        }
        $uvVer = & uv --version 2>$null
        if ($LASTEXITCODE -eq 0 -and $uvVer) {
            Write-OK "uv installed: $uvVer"
            $uvFound = $true
        } else {
            throw "uv did not respond after install"
        }
    } catch {
        Write-ERR "Failed to install uv: $($_.Exception.Message)"
        Write-NFO "Install uv manually from: https://docs.astral.sh/uv/getting-started/installation/"
        exit 1
    }
}

# --- Step 4: Check / install pip (Python) ------------------------------------
$pipFound = $false
try {
    $pipVer = & pip --version 2>$null
    if ($LASTEXITCODE -eq 0 -and $pipVer) {
        Write-OK "pip already available: $pipVer"
        $pipFound = $true
    }
} catch {}

if (-not $pipFound) {
    Write-NFO "pip not found. Checking for Python..."
    try {
        $pyVer = & python --version 2>$null
        if ($LASTEXITCODE -eq 0 -and $pyVer) {
            Write-OK "Python found: $pyVer"
            Write-NFO "pip not installed but Python is present. uvpip will use uv directly."
        } else {
            Write-WRN "Python not found. uvpip will work via uv but standard pip fallback unavailable."
        }
    } catch {
        Write-WRN "Could not detect Python. uvpip will use uv directly."
    }
}

# --- Step 5: Create install directory ----------------------------------------
try {
    New-Item -ItemType Directory -Force -Path $binDir | Out-Null
    Write-OK "Created $installDir"
} catch {
    Write-ERR "Failed to create install directory: $($_.Exception.Message)"
    exit 1
}

# --- Step 6: Download uvpip binary -------------------------------------------
$downloadUrl = "$releaseBase/$binaryName"
Write-NFO "Downloading $binaryName from GitHub releases..."
try {
    Invoke-WebRequest -Uri $downloadUrl -OutFile $exePath -UseBasicParsing
    Write-OK "Downloaded uvpip.exe to $exePath"
} catch {
    Write-ERR "Download failed: $($_.Exception.Message)"
    Write-NFO "URL tried: $downloadUrl"
    Write-NFO "Make sure a GitHub release exists with that binary name."
    if (Test-Path $installDir) { Remove-Item -Path $installDir -Recurse -Force -ErrorAction SilentlyContinue }
    exit 1
}

# --- Step 7: Write pip.cmd and pip3.cmd shims --------------------------------
$shimContent = "@echo off`r`n`"$exePath`" %*"
try {
    Set-Content -Path (Join-Path $binDir "pip.cmd")  -Value $shimContent -Encoding ASCII
    Set-Content -Path (Join-Path $binDir "pip3.cmd") -Value $shimContent -Encoding ASCII
    Write-OK "Created pip.cmd and pip3.cmd shims"
} catch {
    Write-ERR "Failed to create shim files: $($_.Exception.Message)"
    exit 1
}

# --- Step 8: Prepend to User PATH --------------------------------------------
try {
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($null -eq $currentPath) { $currentPath = "" }
    if ($currentPath -notlike "*$binDir*") {
        [Environment]::SetEnvironmentVariable("PATH", "$binDir;$currentPath", "User")
        Write-OK "Added $binDir to User PATH"
    } else {
        Write-WRN "$binDir already in User PATH"
    }
} catch {
    Write-ERR "Failed to update User PATH: $($_.Exception.Message)"
}

# --- Step 9: Prepend to System PATH (requires admin - opens UAC prompt) ------
try {
    $machinePath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
    if ($null -eq $machinePath) { $machinePath = "" }
    if ($machinePath -notlike "*$binDir*") {
        $escapedBinDir = $binDir -replace "'", "''"
        Start-Process powershell -Verb RunAs -Wait -ArgumentList `
            "-ExecutionPolicy Bypass -Command `"[Environment]::SetEnvironmentVariable('PATH', '$escapedBinDir;' + [Environment]::GetEnvironmentVariable('PATH','Machine'), 'Machine')`""
        Write-OK "Added $binDir to System PATH (admin)"
    } else {
        Write-OK "Already in System PATH"
    }
} catch {
    Write-WRN "Could not update System PATH automatically."
    Write-NFO "Run this manually as Administrator:"
    Write-NFO "[Environment]::SetEnvironmentVariable('PATH', '$binDir;' + [Environment]::GetEnvironmentVariable('PATH','Machine'), 'Machine')"
}

# --- Step 10: Refresh current session PATH -----------------------------------
try {
    if ($env:PATH -notlike "*$binDir*") {
        $env:PATH = "$binDir;$env:PATH"
    }
    Write-OK "Refreshed current session PATH"
} catch {
    Write-WRN "Could not refresh session PATH. Restart your terminal."
}

# --- Step 11: Verify install -------------------------------------------------
try {
    $testOutput = & "$exePath" --version 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-OK "uvpip verified: $testOutput"
    } else {
        Write-WRN "uvpip binary ran but returned non-zero exit code"
    }
} catch {
    Write-WRN "Could not verify uvpip binary: $($_.Exception.Message)"
}

# --- Done ---------------------------------------------------------------------
Write-Host ""
Write-Host "  ----------------------------------------" -ForegroundColor DarkGray
Write-Host ""
Write-OK "uvpip installed successfully"
Write-OK "pip and pip3 now run through uv (10-100x faster)"
Write-Host ""
Write-NFO "IMPORTANT: Restart your terminal for PATH changes to take full effect."
Write-NFO "Then run: pip install requests"
Write-NFO "Or run:   uvpip doctor    to verify everything is working."
Write-Host ""
