# Parse only: never execute installers or profiles during syntax validation.
[CmdletBinding()]
param()
$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot
$failed = $false
foreach ($directory in 'installer', 'uninstaller', 'scripts', 'bin') {
    foreach ($file in Get-ChildItem -LiteralPath (Join-Path $repo $directory) -Filter '*.ps1' -File) {
        $tokens = $null
        $parseErrors = $null
        $null = [Management.Automation.Language.Parser]::ParseFile($file.FullName, [ref]$tokens, [ref]$parseErrors)
        foreach ($parseError in $parseErrors) {
            [Console]::Error.WriteLine('{0}:{1}: {2}', $file.FullName, $parseError.Extent.StartLineNumber, $parseError.Message)
            $failed = $true
        }
    }
}
if ($failed) { throw 'PowerShell syntax validation failed' }
Write-Host 'PASS: PowerShell syntax'
