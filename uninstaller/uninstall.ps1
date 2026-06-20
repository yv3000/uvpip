
# uninstall.ps1 - uvpip Windows uninstaller
# Run with: iex (irm https://raw.githubusercontent.com/yv3000/uvpip/main/uninstaller/uninstall.ps1)

$ErrorActionPreference = "Stop"

$installDir = Join-Path $env:USERPROFILE ".uvpip"
$binDir     = Join-Path $installDir "bin"

function Write-OK  { param([string]$msg) Write-Host "  [OK] $msg" -ForegroundColor Green }
function Write-WRN { param([string]$msg) Write-Host "  [!!] $msg" -ForegroundColor Yellow }
function Write-ERR { param([string]$msg) Write-Host "  [ERR] $msg" -ForegroundColor Red }
function Write-NFO { param([string]$msg) Write-Host "  [->] $msg" -ForegroundColor Cyan }

Write-Host ""
Write-Host "  uvpip uninstaller for Windows" -ForegroundColor White
Write-Host "  ----------------------------------------" -ForegroundColor DarkGray
Write-Host ""

# --- Step 1: Check if installed ----------------------------------------------
if (-not (Test-Path $installDir)) {
    Write-WRN "uvpip does not appear to be installed (no folder at $installDir)"
    return
}

# --- Step 2: Remove from User PATH -------------------------------------------
try {
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($currentPath) {
        $entries = $currentPath -split ';' | Where-Object { $_ -ne $binDir -and $_ -ne "" }
        [Environment]::SetEnvironmentVariable("PATH", ($entries -join ';'), "User")
        Write-OK "Removed $binDir from User PATH"
    } else {
        Write-WRN "User PATH was empty"
    }
} catch {
    Write-ERR "Failed to update User PATH: $($_.Exception.Message)"
}

# --- Step 2.5: Remove from PowerShell profile --------------------------------
try {
    if (Test-Path -Path $PROFILE) {
        $profileContent = Get-Content -Path $PROFILE -ErrorAction SilentlyContinue
        if ($profileContent -match "# --- uvpip start ---") {
            $newProfile = @()
            $insideBlock = $false
            foreach ($line in $profileContent) {
                if ($line -match "# --- uvpip start ---") {
                    $insideBlock = $true
                    continue
                }
                if ($line -match "# --- uvpip end ---") {
                    $insideBlock = $false
                    continue
                }
                if ($insideBlock) {
                    continue
                }
                $newProfile += $line
            }
            Set-Content -Path $PROFILE -Value $newProfile
            Write-OK "Removed shell functions from `$PROFILE"
        }
    }
} catch {
    Write-WRN "Failed to clean `$PROFILE: $($_.Exception.Message)"
}

# --- Step 3: Remove from System PATH (UAC) -----------------------------------
try {
    $machinePath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
    if ($machinePath -like "*$binDir*") {
        $escapedBinDir = $binDir -replace "'", "''"
        Start-Process powershell -Verb RunAs -Wait -ArgumentList `
            "-ExecutionPolicy Bypass -Command `"`$e = [Environment]::GetEnvironmentVariable('PATH','Machine') -split ';' | Where-Object { `$_ -ne '$escapedBinDir' -and `$_ -ne '' }; [Environment]::SetEnvironmentVariable('PATH', (`$e -join ';'), 'Machine')`""
        Write-OK "Removed $binDir from System PATH (admin)"
    } else {
        Write-OK "Not present in System PATH"
    }
} catch {
    Write-WRN "Could not update System PATH: $($_.Exception.Message)"
    Write-NFO "Remove $binDir manually from System PATH if needed."
}

# --- Step 4: Delete install directory ----------------------------------------
try {
    Remove-Item -Path $installDir -Recurse -Force
    Write-OK "Deleted $installDir"
} catch {
    Write-ERR "Failed to delete $installDir : $($_.Exception.Message)"
    Write-NFO "Close any terminal windows using uvpip, then run this script again."
    return
}

# --- Step 5: Refresh current session -----------------------------------------
try {
    $sessionEntries = $env:PATH -split ';' | Where-Object { $_ -ne $binDir -and $_ -ne "" }
    $env:PATH = $sessionEntries -join ';'
    Write-OK "Refreshed current session PATH"
} catch {
    Write-WRN "Could not refresh session PATH. Restart your terminal."
}

# --- Done ---------------------------------------------------------------------
Write-Host ""
Write-Host "  ----------------------------------------" -ForegroundColor DarkGray
Write-Host ""
Write-OK "uvpip uninstalled"
Write-OK "Original pip restored (real pip.exe is now first in PATH again)"
Write-NFO "uv itself was NOT removed (you may still use it directly)."
Write-NFO "Restart your terminal for changes to take full effect."
Write-Host ""
