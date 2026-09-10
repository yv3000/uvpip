# Dot-source before Go/tools; only process environment is changed.
param([Parameter(Mandatory)][string]$CacheRoot)
$ErrorActionPreference = 'Stop'
if (-not (Test-Path -LiteralPath $CacheRoot -PathType Container)) { throw 'CacheRoot must exist' }
$root = Join-Path ([IO.Path]::GetFullPath($CacheRoot)) 'uvpip-quality'
$locations = @{
    TEMP = 'tmp'; TMP = 'tmp'; TMPDIR = 'tmp'; GOTMPDIR = 'tmp'
    HOME = 'home'; USERPROFILE = 'home'; APPDATA = 'home\AppData\Roaming'; LOCALAPPDATA = 'home\AppData\Local'
    XDG_CACHE_HOME = 'cache'; XDG_CONFIG_HOME = 'config'; XDG_DATA_HOME = 'data'; XDG_STATE_HOME = 'state'
    GOCACHE = 'go-build'; GOMODCACHE = 'go-mod'; GOPATH = 'go'; GOBIN = 'bin'
    GOTELEMETRYDIR = 'telemetry'; STATICCHECK_CACHE = 'staticcheck'; GH_CONFIG_DIR = 'gh'
    DOTNET_CLI_HOME = 'dotnet'; NUGET_PACKAGES = 'nuget'; npm_config_cache = 'npm'
    PIP_CACHE_DIR = 'pip'; UV_CACHE_DIR = 'uv'; PYTHONPYCACHEPREFIX = 'pycache'
}
foreach ($name in $locations.Keys) {
    $path = Join-Path $root $locations[$name]
    [IO.Directory]::CreateDirectory($path) | Out-Null
    [Environment]::SetEnvironmentVariable($name, $path, 'Process')
}
$env:GOENV = 'off'
$env:GOTOOLCHAIN = 'local'
$env:GOCACHEPROG = ''
$env:GOFLAGS = '-mod=readonly'
$env:DOTNET_CLI_TELEMETRY_OPTOUT = '1'
$env:DOTNET_SKIP_FIRST_TIME_EXPERIENCE = '1'
$env:POWERSHELL_TELEMETRY_OPTOUT = '1'
$env:PYTHONDONTWRITEBYTECODE = '1'
$env:PATH = "$env:GOBIN;$env:PATH"
