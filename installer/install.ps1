# Per-user installer. Isolation: -NoProfile -NoPath -BinaryPath <local exe>.
[CmdletBinding()]
param(
    [string]$BinaryPath = $env:UVPIP_BINARY,
    [string]$SHA256 = $env:UVPIP_SHA256,
    [switch]$NoProfile,
    [switch]$NoPath
)
$ErrorActionPreference = 'Stop'
if (-not $env:USERPROFILE) { throw 'USERPROFILE must be set' }
$installDir = Join-Path $env:USERPROFILE '.uvpip'
$binDir = Join-Path $installDir 'bin'
$exePath = Join-Path $binDir 'uvpip.exe'

function Test-UvpipPath([string]$Entry) {
    [Environment]::ExpandEnvironmentVariables($Entry.Trim().Trim('"')).TrimEnd('\', '/') -ieq $binDir.TrimEnd('\', '/')
}

function Add-UvpipProfile {
    $text = ''
    $encoding = New-Object System.Text.UTF8Encoding($false)
    $preamble = [byte[]]@()
    if (Test-Path -LiteralPath $PROFILE) {
        # Latin-1 round-trips BOM-less bytes (including UTF-8/ANSI). BOMs select Unicode encoding.
        $reader = New-Object System.IO.StreamReader($PROFILE, [Text.Encoding]::GetEncoding(28591), $true)
        try { $text = $reader.ReadToEnd(); $encoding = $reader.CurrentEncoding } finally { $reader.Dispose() }
        $preamble = $encoding.GetPreamble()
    }
    $markers = [regex]::Matches($text, '(?m)^# --- uvpip (start|end) ---\r?$')
    $inside = $false
    foreach ($marker in $markers) {
        $start = $marker.Groups[1].Value -eq 'start'
        if ($start -eq $inside) { throw "Unbalanced uvpip markers in $PROFILE; repair manually" }
        $inside = $start
    }
    if ($inside) { throw "Unbalanced uvpip markers in $PROFILE; repair manually" }
    if ($markers.Count) { return }
    $block = @'
# --- uvpip start ---
function pip {
    & "$env:USERPROFILE\.uvpip\bin\uvpip.exe" @args
}
function pip3 {
    & "$env:USERPROFILE\.uvpip\bin\uvpip.exe" @args
}
# --- uvpip end ---
'@
    $newline = "`r`n"
    if ($text.Contains("`n") -and -not $text.Contains("`r`n")) { $newline = "`n" }
    if ($text.Length -and -not $text.EndsWith("`n")) { $text += $newline }
    $text += ($block -replace '\r?\n', $newline) + $newline
    [IO.Directory]::CreateDirectory((Split-Path -Parent $PROFILE)) | Out-Null
    [IO.File]::WriteAllBytes($PROFILE, [byte[]]($preamble + $encoding.GetBytes($text)))
}

$stage = Join-Path ([IO.Path]::GetTempPath()) ('uvpip-install-' + [guid]::NewGuid().ToString('N'))
[IO.Directory]::CreateDirectory($stage) | Out-Null
try {
    $candidate = Join-Path $stage 'uvpip.exe'
    if (-not (Test-Path -LiteralPath $exePath)) {
        if ($BinaryPath) {
            Copy-Item -LiteralPath $BinaryPath -Destination $candidate
        } else {
            $arch = $env:PROCESSOR_ARCHITEW6432
            if (-not $arch) { $arch = $env:PROCESSOR_ARCHITECTURE }
            switch ($arch) {
                'AMD64' { $archName = 'amd64' }
                'ARM64' { $archName = 'arm64' }
                default { throw "Unsupported architecture: $arch" }
            }
            Invoke-WebRequest -Uri "https://github.com/yv3000/uvpip/releases/latest/download/uvpip-windows-$archName.exe" -OutFile $candidate -UseBasicParsing
        }
        if ((Get-Item -LiteralPath $candidate).Length -eq 0) { throw 'uvpip binary is empty' }
        if ($SHA256 -and (Get-FileHash -LiteralPath $candidate -Algorithm SHA256).Hash -ine $SHA256) {
            throw 'SHA256 mismatch'
        }
        $checkPath = $candidate
    } else {
        $checkPath = $exePath
    }
    $global:LASTEXITCODE = $null
    & $checkPath --version | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'uvpip executable check failed' }

    # Download the official uv installer completely before executing it.
    if (-not (Get-Command uv -ErrorAction SilentlyContinue)) {
        $uvInstaller = Join-Path $stage 'install-uv.ps1'
        Invoke-WebRequest -Uri 'https://astral.sh/uv/install.ps1' -OutFile $uvInstaller -UseBasicParsing
        if ((Get-Item -LiteralPath $uvInstaller).Length -eq 0) { throw 'uv installer download was empty' }
        & powershell.exe -NoLogo -NoProfile -NonInteractive -ExecutionPolicy Bypass -File $uvInstaller
        if ($LASTEXITCODE -ne 0) { throw 'uv installation failed' }
        $env:PATH = (Join-Path $env:USERPROFILE '.local\bin') + ';' + $env:PATH
    }
    $global:LASTEXITCODE = $null
    & uv --version | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'uv executable check failed' }

    [IO.Directory]::CreateDirectory($binDir) | Out-Null
    if (Test-Path -LiteralPath $candidate) { Move-Item -LiteralPath $candidate -Destination $exePath }
    $shim = "@echo off`r`n`"%~dp0uvpip.exe`" %*`r`nexit /b %errorlevel%`r`n"
    foreach ($name in 'pip.cmd', 'pip3.cmd') {
        [IO.File]::WriteAllText((Join-Path $binDir $name), $shim, [Text.Encoding]::ASCII)
    }
    if (-not $NoProfile) { Add-UvpipProfile }
    if (-not $NoPath) {
        $currentPath = [Environment]::GetEnvironmentVariable('PATH', 'User')
        if (-not (@($currentPath -split ';' | Where-Object { Test-UvpipPath $_ }).Count)) {
            $newPath = $binDir
            if ($currentPath) { $newPath += ';' + $currentPath }
            [Environment]::SetEnvironmentVariable('PATH', $newPath, 'User')
        }
        if (-not (@($env:PATH -split ';' | Where-Object { Test-UvpipPath $_ }).Count)) {
            $env:PATH = "$binDir;$env:PATH"
        }
        $machinePath = [Environment]::GetEnvironmentVariable('PATH', 'Machine')
        if (@($machinePath -split ';' | Where-Object { Test-UvpipPath $_ }).Count) {
            Write-Warning "Legacy System PATH entry found: $binDir. Remove it manually if needed; this installer never elevates or edits System PATH."
        }
    }
    Write-Host 'uvpip installed. Restart your terminal. Existing pip and uv were not removed.'
} finally {
    Remove-Item -LiteralPath $stage -Recurse -Force
}
