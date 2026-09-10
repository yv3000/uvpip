# Per-user uninstaller. -NoProfile -NoPath disable shell/environment changes.
[CmdletBinding()]
param([switch]$NoProfile, [switch]$NoPath)
$ErrorActionPreference = 'Stop'
if (-not $env:USERPROFILE) { throw 'USERPROFILE must be set' }
$installDir = Join-Path $env:USERPROFILE '.uvpip'
$binDir = Join-Path $installDir 'bin'

function Test-UvpipPath([string]$Entry) {
    [Environment]::ExpandEnvironmentVariables($Entry.Trim().Trim('"')).TrimEnd('\', '/') -ieq $binDir.TrimEnd('\', '/')
}

if (-not $NoProfile -and (Test-Path -LiteralPath $PROFILE)) {
    # Keep existing BOM/encoding, line endings, and all bytes outside exact blocks.
    $reader = New-Object System.IO.StreamReader($PROFILE, [Text.Encoding]::GetEncoding(28591), $true)
    try { $text = $reader.ReadToEnd(); $encoding = $reader.CurrentEncoding } finally { $reader.Dispose() }
    $inside = $false
    foreach ($marker in [regex]::Matches($text, '(?m)^# --- uvpip (start|end) ---\r?$')) {
        $start = $marker.Groups[1].Value -eq 'start'
        if ($start -eq $inside) { throw "Unbalanced uvpip markers in $PROFILE; file left unchanged. Repair manually and retry." }
        $inside = $start
    }
    if ($inside) { throw "Unbalanced uvpip markers in $PROFILE; file left unchanged. Repair manually and retry." }
    $clean = [regex]::Replace($text, '(?ms)^# --- uvpip start ---\r?\n.*?^# --- uvpip end ---\r?(?:\n|\z)', '')
    if ($clean -cne $text) {
        [IO.File]::WriteAllBytes($PROFILE, [byte[]]($encoding.GetPreamble() + $encoding.GetBytes($clean)))
    }
}

if (-not $NoPath) {
    $currentPath = [Environment]::GetEnvironmentVariable('PATH', 'User')
    if (@($currentPath -split ';' | Where-Object { Test-UvpipPath $_ }).Count) {
        $entries = @($currentPath -split ';' | Where-Object { -not (Test-UvpipPath $_) })
        [Environment]::SetEnvironmentVariable('PATH', ($entries -join ';'), 'User')
    }
    $env:PATH = (@($env:PATH -split ';' | Where-Object { -not (Test-UvpipPath $_) }) -join ';')
    $machinePath = [Environment]::GetEnvironmentVariable('PATH', 'Machine')
    if (@($machinePath -split ';' | Where-Object { Test-UvpipPath $_ }).Count) {
        Write-Warning "Legacy System PATH entry remains: $binDir. Remove it manually; this uninstaller never elevates or edits System PATH."
    }
}

# Delete only known install artifacts, never unrelated files in .uvpip.
foreach ($name in 'uvpip.exe', 'pip.cmd', 'pip3.cmd') {
    $path = Join-Path $binDir $name
    if (Test-Path -LiteralPath $path) { Remove-Item -LiteralPath $path -Force }
}
foreach ($path in $binDir, $installDir) {
    if ((Test-Path -LiteralPath $path) -and -not @(Get-ChildItem -LiteralPath $path -Force).Count) {
        [IO.Directory]::Delete($path)
    }
}
Write-Host 'uvpip uninstalled. Restart your terminal. Existing pip and uv were not removed.'
