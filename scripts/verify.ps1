# Development only. Run in a fresh powershell.exe -NoProfile process.
[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$TestRoot,
    [Parameter(Mandatory)][string]$CacheRoot
)
$ErrorActionPreference = 'Stop'
if (-not (Test-Path -LiteralPath $TestRoot -PathType Container) -or -not (Test-Path -LiteralPath $CacheRoot -PathType Container)) {
    throw 'TestRoot and CacheRoot must already exist'
}
$TestRoot = (Resolve-Path -LiteralPath $TestRoot).Path
$CacheRoot = (Resolve-Path -LiteralPath $CacheRoot).Path
. (Join-Path $PSScriptRoot 'env.ps1') -CacheRoot $CacheRoot
$env:GOTOOLCHAIN = 'local'
$env:GOFLAGS = '-mod=readonly'
$env:PATH = "$env:GOBIN;$env:PATH"
$artifacts = Join-Path $env:TEMP ('uvpip-verify-' + [guid]::NewGuid().ToString('N'))
[IO.Directory]::CreateDirectory($artifacts) | Out-Null
Write-Host "Verification artifacts: $artifacts"
Push-Location (Split-Path -Parent $PSScriptRoot)
try {
    $goVersion = & go env GOVERSION
    if ($LASTEXITCODE -ne 0) { throw 'go env failed' }
    if ($goVersion -ne 'go1.26.8') { throw "Expected Go 1.26.8, found $goVersion" }
    $goFiles = @(Get-ChildItem -LiteralPath . -Filter '*.go' -File | ForEach-Object { $_.FullName })
    if (-not $goFiles.Count) { throw 'No project Go files found' }
    $unformatted = & gofmt -l @goFiles
    if ($LASTEXITCODE -ne 0) { throw 'gofmt failed' }
    if ($unformatted) { throw "Run gofmt on: $($unformatted -join ', ')" }
    & go vet ./...
    if ($LASTEXITCODE -ne 0) { throw 'go vet failed' }
    & go build -o (Join-Path $artifacts 'uvpip-build.exe') ./...
    if ($LASTEXITCODE -ne 0) { throw 'go build failed' }
    $coverage = Join-Path $artifacts 'coverage.out'
    & go test -count=1 "-coverprofile=$coverage" ./...
    if ($LASTEXITCODE -ne 0) { throw 'go test failed' }
    $report = & go tool cover "-func=$coverage"
    if ($LASTEXITCODE -ne 0) { throw 'go tool cover failed' }
    $report | ForEach-Object { Write-Host $_ }
    $total = [regex]::Match(($report -join "`n"), '(?m)^total:\s+\(statements\)\s+([0-9]+(?:\.[0-9]+)?)%\s*$')
    if (-not $total.Success) { throw 'Missing total statement coverage' }
    $percent = [double]::Parse($total.Groups[1].Value, [Globalization.CultureInfo]::InvariantCulture)
    if ($percent -lt 90) { throw "Statement coverage $percent% is below 90%" }
    & staticcheck ./...
    if ($LASTEXITCODE -ne 0) { throw 'staticcheck failed' }
    & govulncheck ./...
    if ($LASTEXITCODE -ne 0) { throw 'govulncheck failed' }
    & gitleaks dir --redact --no-banner .
    if ($LASTEXITCODE -ne 0) { throw 'Secret scan failed' }
    & actionlint .github/workflows/ci.yml
    if ($LASTEXITCODE -ne 0) { throw 'actionlint failed' }
    $shellFiles = @(Get-ChildItem -Path installer, uninstaller, scripts, bin -Filter '*.sh' -File | ForEach-Object { $_.FullName })
    & shellcheck @shellFiles
    if ($LASTEXITCODE -ne 0) { throw 'shellcheck failed' }
    & powershell.exe -NoLogo -NoProfile -NonInteractive -ExecutionPolicy Bypass -File scripts/check-powershell.ps1
    if ($LASTEXITCODE -ne 0) { throw 'PowerShell syntax validation failed' }
    # The fixture compiler uses ConsoleApplication, which requires Windows PowerShell 5.1.
    & powershell.exe -NoLogo -NoProfile -NonInteractive -ExecutionPolicy Bypass -File scripts/test-installers.ps1 -TestRoot $TestRoot -CacheRoot $CacheRoot
    if ($LASTEXITCODE -ne 0) { throw 'Offline Windows installer tests failed' }
    Write-Host 'PASS: Windows verification (90% minimum statement coverage)'
} finally {
    Pop-Location
}
