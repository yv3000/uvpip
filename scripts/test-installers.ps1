# Run with powershell.exe -NoProfile -File scripts/test-installers.ps1 -TestRoot <existing dir> -CacheRoot <existing dir>
[CmdletBinding()]
param([Parameter(Mandatory)][string]$TestRoot, [Parameter(Mandatory)][string]$CacheRoot)
$ErrorActionPreference = 'Stop'
if (-not (Test-Path -LiteralPath $TestRoot -PathType Container) -or -not (Test-Path -LiteralPath $CacheRoot -PathType Container)) {
    throw 'TestRoot and CacheRoot must already exist'
}
$repo = Split-Path -Parent $PSScriptRoot
$savedEnv = @{}
foreach ($name in 'TEMP', 'TMP', 'TMPDIR', 'HOME', 'USERPROFILE', 'APPDATA', 'LOCALAPPDATA',
    'XDG_CACHE_HOME', 'XDG_CONFIG_HOME', 'XDG_DATA_HOME', 'GOCACHE', 'GOMODCACHE', 'GOPATH',
    'DOTNET_CLI_HOME', 'NUGET_PACKAGES', 'npm_config_cache', 'PIP_CACHE_DIR', 'UV_CACHE_DIR', 'PYTHONPYCACHEPREFIX', 'PATH') {
    $savedEnv[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
    if ($name -ne 'PATH') { [Environment]::SetEnvironmentVariable($name, $CacheRoot, 'Process') }
}
$work = Join-Path $TestRoot ('uvpip-powershell-' + [guid]::NewGuid().ToString('N'))
$cache = Join-Path $CacheRoot ('uvpip-powershell-' + [guid]::NewGuid().ToString('N'))
[IO.Directory]::CreateDirectory($work) | Out-Null
[IO.Directory]::CreateDirectory($cache) | Out-Null
$env:TEMP = $env:TMP = $env:TMPDIR = $cache
$env:USERPROFILE = $env:HOME = Join-Path $cache "home spaces 'quote' [literal]"
[IO.Directory]::CreateDirectory($env:USERPROFILE) | Out-Null
$PROFILE = Join-Path $env:USERPROFILE 'profiles\test.ps1'
$installer = Join-Path $repo 'installer\install.ps1'
$uninstaller = Join-Path $repo 'uninstaller\uninstall.ps1'
$userPathBefore = [Environment]::GetEnvironmentVariable('PATH', 'User')
$machinePathBefore = [Environment]::GetEnvironmentVariable('PATH', 'Machine')
function Assert($Condition, [string]$Message) { if (-not $Condition) { throw "FAIL: $Message" } }
function Bytes([string]$Path) { [Convert]::ToBase64String([IO.File]::ReadAllBytes($Path)) }
function Expect-Failure([scriptblock]$Action) {
    $failed = $false
    try { & $Action 6>$null } catch { $failed = $true }
    Assert $failed 'expected terminating failure'
}
function Assert-NoStage {
    Assert (@(Get-ChildItem -LiteralPath $cache -Directory -Filter 'uvpip-install-*').Count -eq 0) 'staging leaked'
}
# All downloads and uv setup are mocked. No network or persistent environment writes.
function uv { $global:LASTEXITCODE = 0; 'uv offline' }
function Invoke-WebRequest {
    param($Uri, $OutFile, [switch]$UseBasicParsing)
    [IO.File]::WriteAllText($OutFile, 'partial')
    throw 'Offline simulated download failure'
}
try {
    foreach ($file in $installer, $uninstaller) {
        $tokens = $null; $errors = $null
        $ast = [Management.Automation.Language.Parser]::ParseFile($file, [ref]$tokens, [ref]$errors)
        Assert ($errors.Count -eq 0) "parse failed: $file"
        # Exercise exactly the production matcher, without registry mutation.
        $matcher = $ast.Find({ param($node) $node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq 'Test-UvpipPath' }, $true)
        . ([scriptblock]::Create($matcher.Extent.Text))
        $binDir = Join-Path $env:USERPROFILE '.uvpip\bin'
        Assert (Test-UvpipPath $binDir.ToUpperInvariant()) 'case insensitive exact PATH match'
        Assert (Test-UvpipPath ('"' + $binDir + '\"')) 'quoted PATH with trailing separator'
        Assert (-not (Test-UvpipPath ($binDir + '-other'))) 'PATH substring match'
        $entries = @('first', $binDir, '', ($binDir + '-other'), 'last') | Where-Object { -not (Test-UvpipPath $_) }
        Assert (($entries -join ';') -eq "first;;$binDir-other;last") 'PATH entry preservation'
    }
    # Built using Windows PowerShell's existing compiler; all compiler temp files are redirected.
    $fixture = Join-Path $work 'fixture.exe'
    Add-Type -OutputAssembly $fixture -OutputType ConsoleApplication -TypeDefinition @'
public static class InstallerFixture {
    public static int Main(string[] args) {
        if (args.Length == 1 && args[0] == "--version") { System.Console.WriteLine("uvpip offline"); return 0; }
        if (args.Length == 1 && args[0] == "--fail") return 23;
        foreach (string arg in args) System.Console.WriteLine("<" + arg + ">");
        return 0;
    }
}
'@
    & $installer -BinaryPath $fixture -NoPath 6>$null
    Assert (Test-Path -LiteralPath $PROFILE) 'missing profile not created'
    $installedProfile = Bytes $PROFILE
    & $installer -BinaryPath $fixture -NoPath 6>$null
    Assert ((Bytes $PROFILE) -eq $installedProfile) 'install idempotence'
    $contents = [IO.File]::ReadAllText($PROFILE)
    Assert ($contents.Contains('$env:USERPROFILE')) 'runtime home reference missing'
    & {
        . $PROFILE
        Assert ((pip 'a b' "single'quote" '*' ) -join "`n" -ceq "<a b>`n<single'quote>`n<*>") 'profile argument forwarding'
        pip3 --fail
        Assert ($LASTEXITCODE -eq 23) 'profile exit status'
    }
    foreach ($name in 'pip.cmd', 'pip3.cmd') {
        $shim = Join-Path $binDir $name
        $output = & $shim 'a b' '*'
        Assert (($output -join "`n") -ceq "<a b>`n<*>") 'cmd argument forwarding'
        & $shim --fail
        Assert ($LASTEXITCODE -eq 23) 'cmd exit status'
        Copy-Item -LiteralPath (Join-Path $repo "bin\$name") -Destination $shim -Force
        & $shim --fail
        Assert ($LASTEXITCODE -eq 23) 'repository cmd exit status'
    }
    & $uninstaller -NoPath 6>$null
    $removedProfile = Bytes $PROFILE
    & $uninstaller -NoPath 6>$null
    Assert ((Bytes $PROFILE) -eq $removedProfile) 'uninstall idempotence'

    # BOM-less UTF-8, BOM UTF-8, UTF-16 LE/BE, and ANSI retain unrelated bytes.
    foreach ($encoding in @((New-Object Text.UTF8Encoding($false)), (New-Object Text.UTF8Encoding($true)),
        [Text.Encoding]::Unicode, [Text.Encoding]::BigEndianUnicode, [Text.Encoding]::GetEncoding(1252))) {
        $prefix = '# keep .uvpip ' + [char]0xe9 + "`r`n# --- uvpip start --- extra`r`n"
        $suffix = '# keep tail without newline'
        [IO.File]::WriteAllText($PROFILE, $prefix, $encoding)
        & $installer -BinaryPath $fixture -NoPath 6>$null
        [IO.File]::AppendAllText($PROFILE, $suffix, $encoding)
        $expected = Join-Path $work 'expected'
        [IO.File]::WriteAllText($expected, $prefix + $suffix, $encoding)
        $keep = Join-Path $env:USERPROFILE '.uvpip\keep.txt'
        [IO.File]::WriteAllText($keep, 'keep')
        & $uninstaller -NoPath 6>$null
        Assert ((Bytes $PROFILE) -eq (Bytes $expected)) 'profile encoding/bytes preservation'
        Assert ([IO.File]::ReadAllText($keep) -eq 'keep') 'unrelated file removed'
    }
    foreach ($bad in @('# --- uvpip start ---', '# --- uvpip end ---', "# --- uvpip start ---`n# --- uvpip start ---`n# --- uvpip end ---")) {
        [IO.File]::WriteAllText($PROFILE, $bad + "`nkeep tail")
        $before = Bytes $PROFILE
        Expect-Failure { & $uninstaller -NoPath }
        Assert ((Bytes $PROFILE) -eq $before) 'malformed profile changed'
        Expect-Failure { & $installer -BinaryPath $fixture -NoPath }
        Assert ((Bytes $PROFILE) -eq $before) 'installer changed malformed profile'
        & $uninstaller -NoProfile -NoPath 6>$null
    }
    # Adjacent complete blocks, LF endings, and an end marker at EOF.
    [IO.File]::WriteAllText($PROFILE, "keep`n# --- uvpip start ---`nowned`n# --- uvpip end ---`n# --- uvpip start ---`nowned`n# --- uvpip end ---")
    & $uninstaller -NoPath 6>$null
    Assert ([IO.File]::ReadAllText($PROFILE) -ceq "keep`n") 'exact multi-block removal'
    [IO.File]::WriteAllText($PROFILE, '# keep unchanged')
    $before = Bytes $PROFILE
    Expect-Failure { & $installer -BinaryPath '' -SHA256 '' -NoPath }
    Expect-Failure { & $installer -BinaryPath $fixture -SHA256 'bad' -NoPath }
    $badBinary = Join-Path $work 'bad.exe'
    [IO.File]::WriteAllText($badBinary, 'not executable')
    Expect-Failure { & $installer -BinaryPath $badBinary -NoPath }
    # Force uv absent; a failed download must never invoke its partial installer.
    function Get-Command {
        param($Name, $ErrorAction)
        if ($Name -ne 'uv') { Microsoft.PowerShell.Core\Get-Command $Name -ErrorAction $ErrorAction }
    }
    Expect-Failure { & $installer -BinaryPath $fixture -NoPath }
    Remove-Item Function:\Get-Command
    Assert ((Bytes $PROFILE) -eq $before) 'failed download changed profile'
    Assert (-not (Test-Path -LiteralPath (Join-Path $binDir 'uvpip.exe'))) 'failed install left executable'
    Assert ([IO.File]::ReadAllText($keep) -eq 'keep') 'failed install removed unrelated data'
    Assert-NoStage
    function Invoke-WebRequest {
        param($Uri, $OutFile, [switch]$UseBasicParsing)
        Copy-Item -LiteralPath $fixture -Destination $OutFile
    }
    $hash = (Get-FileHash -LiteralPath $fixture -Algorithm SHA256).Hash
    & $installer -BinaryPath '' -SHA256 $hash -NoProfile -NoPath 6>$null
    & $uninstaller -NoProfile -NoPath 6>$null
    Assert ((Bytes $PROFILE) -eq $before) 'NoProfile changed profile'
    Assert-NoStage
    Assert ([Environment]::GetEnvironmentVariable('PATH', 'User') -ceq $userPathBefore) 'User PATH changed'
    Assert ([Environment]::GetEnvironmentVariable('PATH', 'Machine') -ceq $machinePathBefore) 'Machine PATH changed'
    Write-Host 'PASS: offline Windows installer/uninstaller tests'
} finally {
    Remove-Item -LiteralPath $work, $cache -Recurse -Force
    foreach ($name in $savedEnv.Keys) { [Environment]::SetEnvironmentVariable($name, $savedEnv[$name], 'Process') }
}
