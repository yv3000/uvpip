# Contributing

Keep changes focused and include a runnable regression check for changed behavior.
Do not add vendored environments, caches, secrets, generated binaries, or coverage
profiles to commits. Tests must not modify your real profile, registry PATH, pip,
or uv installation. Use the offline installer suites rather than a live install.

## Toolchain

Verification is pinned to:

- Go **1.26.8** (a supported release listed by [go.dev](https://go.dev/dl/), not
  a claim that it is the newest Go minor version).
- Staticcheck: `honnef.co/go/tools/cmd/staticcheck@v0.8.1`.
- govulncheck: `golang.org/x/vuln/cmd/govulncheck@v1.8.0` (requires Go 1.26).
- actionlint: `github.com/rhysd/actionlint/cmd/actionlint@v1.7.7`.
- Gitleaks: `github.com/zricethezav/gitleaks/v8@v8.30.1` (redacted source secret scan).
- [ShellCheck 0.11.0](https://github.com/koalaman/shellcheck/releases/tag/v0.11.0).
- Windows PowerShell **5.1** for Windows installer tests; PowerShell 7 (`pwsh`)
  for syntax parsing on Linux/macOS. Linux race tests also require a C compiler.

`staticcheck.conf` inherits upstream default checks; `.shellcheckrc` selects
POSIX shell without globally suppressing warnings. The module language minimum
is Go **1.26.0**; verification uses **1.26.8**. The runtime has no third-party Go
dependencies, so there is no `go.sum`. `GOTOOLCHAIN=local` prevents an
unexpected automatic compiler download. Change tool pins through review, not
by substituting `@latest` in CI. Dependabot covers Go modules and Actions, not
the versions embedded in `go install` commands or ShellCheck download hashes.

## Windows Verification

Run from the repository root. These two parent directories must already exist.
On this workspace, all development caches/home/temp belong under the manager
root and installer test artifacts under the testing root:

```powershell
$cache = 'X:\ALL FOLDER\MANAGER of all packages and etc'
$tests = 'X:\ALL FOLDER\TESTING FOR EVERYTHING'
. .\scripts\env.ps1 -CacheRoot $cache
$env:GOTOOLCHAIN = 'local'
$goBin = Join-Path $env:LOCALAPPDATA 'go1.26.8\go\bin'
if (-not (Test-Path -LiteralPath (Join-Path $goBin 'go.exe') -PathType Leaf)) { throw 'Install Go 1.26.8 under the managed LOCALAPPDATA first' }
$env:PATH = "$goBin;$env:GOBIN;$env:PATH"
go install honnef.co/go/tools/cmd/staticcheck@v0.8.1
if ($LASTEXITCODE -ne 0) { throw 'Staticcheck installation failed' }
go install golang.org/x/vuln/cmd/govulncheck@v1.8.0
if ($LASTEXITCODE -ne 0) { throw 'govulncheck installation failed' }
go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.7
if ($LASTEXITCODE -ne 0) { throw 'actionlint installation failed' }
go install github.com/zricethezav/gitleaks/v8@v8.30.1
if ($LASTEXITCODE -ne 0) { throw 'Gitleaks installation failed' }
powershell.exe -NoLogo -NoProfile -NonInteractive -ExecutionPolicy Bypass -File scripts/verify.ps1 -TestRoot $tests -CacheRoot $cache
if ($LASTEXITCODE -ne 0) { throw 'Verification failed' }
```

Install Go and ShellCheck before these commands, using reviewed archives and
checksums. Keep tool installations beneath the manager root on this machine and
make them available on process PATH; do not change persistent PATH just for tests.
`go install` writes the verification tools to the environment helper's `GOBIN`.
Close the development shell afterward: the helper deliberately redirects HOME
and other environment variables and is **not** part of live installation.
Resolve the Go path **after** calling the helper: on this machine, `LOCALAPPDATA`
then points to `uvpip-quality\home\AppData\Local` beneath the manager root, not
your normal Windows profile.

The verifier calls `env.ps1` itself, checks every native exit code, and retains
the build and coverage report in a unique directory under the managed TEMP. It
prints that path for inspection. It does not overwrite the checkout's preexisting
ignored `uvpip.exe`. The installer suite runs in a child **powershell.exe**, not
`pwsh`: its `Add-Type -OutputType ConsoleApplication` fixture needs PowerShell 5.1.

## Required Fresh Clone Verification

Before release, verify the exact committed source in a new clone, not just the
working tree. Before source changes are committed, an archive of the Git index
can provide an independent staged-source check, but it is **not a fresh clone**:
it excludes unstaged/untracked changes and has no commit provenance. It does not
satisfy this requirement. After committing, use the local-clone route below;
after pushing the reviewed commit to `main`, use the remote route and require
green native CI for that commit.

Each Windows block is standalone and may start from any directory in a dedicated
Windows PowerShell 5.1 session. Both existing parent roots are explicit, and every
clone goes into a new `uvpip-clean-<GUID>` directory beneath the dedicated X: testing
root. Never clone into the source checkout, manager root, or a real user profile.
Go 1.26.8 must already be installed at the managed path shown; reviewed ShellCheck
0.11.0 and Git must be on the process PATH. Go analysis tools are installed with
the documented pins after cloning.

The bootstrap deliberately dot-sources `env.ps1` from the **trusted, reviewed
source checkout before any Git command**. This redirects home, temp, and tool
caches before cloning; the clone's own helper is then loaded for verification.
This tests an independent checkout, not an independent tool/cache installation.
Do not bootstrap with an unreviewed downloaded script. Close the session afterward.

### Committed Local Main

Commit all intended source changes first. This clones the exact local `main`
branch without sharing hardlinked Git objects; it neither commits nor pushes.

```powershell
$ErrorActionPreference = 'Stop'
$SourceRoot = 'X:\ALL FOLDER\PROJECTTTTTT\uvpip'
$CacheRoot = 'X:\ALL FOLDER\MANAGER of all packages and etc'
$TestRoot = 'X:\ALL FOLDER\TESTING FOR EVERYTHING'
if (-not (Test-Path -LiteralPath $CacheRoot -PathType Container) -or -not (Test-Path -LiteralPath $TestRoot -PathType Container)) { throw 'CacheRoot and TestRoot must already exist' }
. (Join-Path $SourceRoot 'scripts\env.ps1') -CacheRoot $CacheRoot
$goBin = Join-Path $env:LOCALAPPDATA 'go1.26.8\go\bin'
if (-not (Test-Path -LiteralPath (Join-Path $goBin 'go.exe') -PathType Leaf)) { throw 'Managed Go 1.26.8 is missing' }
$env:PATH = "$goBin;$env:PATH"
Get-Command shellcheck -ErrorAction Stop | Out-Null
$status = git -C $SourceRoot status --porcelain
if ($LASTEXITCODE -ne 0) { throw 'Source status failed' }
if ($status) { throw 'Commit the intended changes before fresh-clone verification' }
$SourceCommit = git -C $SourceRoot rev-parse refs/heads/main
if ($LASTEXITCODE -ne 0) { throw 'Cannot resolve local main' }
$CloneRoot = Join-Path $TestRoot ('uvpip-clean-' + [guid]::NewGuid().ToString('N'))
git clone --no-hardlinks --single-branch --branch main -- $SourceRoot $CloneRoot
if ($LASTEXITCODE -ne 0) { throw 'Local clone failed' }
Set-Location -LiteralPath $CloneRoot
$CloneCommit = git rev-parse HEAD
if ($LASTEXITCODE -ne 0) { throw 'Cannot resolve clone commit' }
if ($CloneCommit -ne $SourceCommit) { throw 'Clone does not match the selected source commit' }
Write-Host "Verifying local main commit: $CloneCommit"
. .\scripts\env.ps1 -CacheRoot $CacheRoot
go install honnef.co/go/tools/cmd/staticcheck@v0.8.1
if ($LASTEXITCODE -ne 0) { throw 'Staticcheck installation failed' }
go install golang.org/x/vuln/cmd/govulncheck@v1.8.0
if ($LASTEXITCODE -ne 0) { throw 'govulncheck installation failed' }
go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.7
if ($LASTEXITCODE -ne 0) { throw 'actionlint installation failed' }
go install github.com/zricethezav/gitleaks/v8@v8.30.1
if ($LASTEXITCODE -ne 0) { throw 'Gitleaks installation failed' }
powershell.exe -NoLogo -NoProfile -NonInteractive -ExecutionPolicy Bypass -File scripts/verify.ps1 -TestRoot $TestRoot -CacheRoot $CacheRoot
if ($LASTEXITCODE -ne 0) { throw 'Fresh-clone verification failed' }
```

### Pushed Remote Main

Run only once the revised source is pushed to `main`. This uses the real public
repository URL; cloning it before the push cannot verify local changes. Record
the printed commit and compare it with the reviewed source commit and CI run.

```powershell
$ErrorActionPreference = 'Stop'
$SourceRoot = 'X:\ALL FOLDER\PROJECTTTTTT\uvpip'
$CacheRoot = 'X:\ALL FOLDER\MANAGER of all packages and etc'
$TestRoot = 'X:\ALL FOLDER\TESTING FOR EVERYTHING'
if (-not (Test-Path -LiteralPath $CacheRoot -PathType Container) -or -not (Test-Path -LiteralPath $TestRoot -PathType Container)) { throw 'CacheRoot and TestRoot must already exist' }
. (Join-Path $SourceRoot 'scripts\env.ps1') -CacheRoot $CacheRoot
$goBin = Join-Path $env:LOCALAPPDATA 'go1.26.8\go\bin'
if (-not (Test-Path -LiteralPath (Join-Path $goBin 'go.exe') -PathType Leaf)) { throw 'Managed Go 1.26.8 is missing' }
$env:PATH = "$goBin;$env:PATH"
Get-Command shellcheck -ErrorAction Stop | Out-Null
$CloneRoot = Join-Path $TestRoot ('uvpip-clean-' + [guid]::NewGuid().ToString('N'))
git clone --single-branch --branch main -- https://github.com/yv3000/uvpip.git $CloneRoot
if ($LASTEXITCODE -ne 0) { throw 'Remote clone failed' }
Set-Location -LiteralPath $CloneRoot
$CloneCommit = git rev-parse HEAD
if ($LASTEXITCODE -ne 0) { throw 'Cannot resolve clone commit' }
Write-Host "Verifying remote main commit: $CloneCommit"
. .\scripts\env.ps1 -CacheRoot $CacheRoot
go install honnef.co/go/tools/cmd/staticcheck@v0.8.1
if ($LASTEXITCODE -ne 0) { throw 'Staticcheck installation failed' }
go install golang.org/x/vuln/cmd/govulncheck@v1.8.0
if ($LASTEXITCODE -ne 0) { throw 'govulncheck installation failed' }
go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.7
if ($LASTEXITCODE -ne 0) { throw 'actionlint installation failed' }
powershell.exe -NoLogo -NoProfile -NonInteractive -ExecutionPolicy Bypass -File scripts/verify.ps1 -TestRoot $TestRoot -CacheRoot $CacheRoot
if ($LASTEXITCODE -ne 0) { throw 'Fresh-clone verification failed' }
```

The verifier prints its newly generated artifact directory. Use that output,
not a path copied from an earlier run, when inspecting build/coverage artifacts.

## Linux / macOS Verification

Use existing, absolute `CACHE_ROOT` and `TEST_ROOT` directories outside the
checkout. In a dedicated development shell, redirect Go installation paths first:

```sh
: "${CACHE_ROOT:?Set an existing cache root}"
: "${TEST_ROOT:?Set an existing test root}"
test -d "$CACHE_ROOT" && test -d "$TEST_ROOT" || exit 1
export HOME="$CACHE_ROOT/uvpip-dev/home"
export TMPDIR="$CACHE_ROOT/uvpip-dev/tmp" TMP="$CACHE_ROOT/uvpip-dev/tmp" TEMP="$CACHE_ROOT/uvpip-dev/tmp"
export GOPATH="$CACHE_ROOT/uvpip-dev/go" GOBIN="$CACHE_ROOT/uvpip-dev/bin"
export GOCACHE="$CACHE_ROOT/uvpip-dev/go-build" GOMODCACHE="$CACHE_ROOT/uvpip-dev/go-mod" GOTMPDIR="$TMPDIR"
export GOTOOLCHAIN=local
export PATH="$GOBIN:$PATH"
mkdir -p "$HOME" "$TMPDIR" "$GOPATH" "$GOBIN" "$GOCACHE" "$GOMODCACHE"
go install honnef.co/go/tools/cmd/staticcheck@v0.8.1 || exit 1
go install golang.org/x/vuln/cmd/govulncheck@v1.8.0 || exit 1
go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.7 || exit 1
go install github.com/zricethezav/gitleaks/v8@v8.30.1 || exit 1
sh scripts/verify.sh "$TEST_ROOT" "$CACHE_ROOT"
```

Go, ShellCheck, and `pwsh` must be available on PATH. The script sets its own
development environment, runs Linux race tests, and retains build/coverage
artifacts beneath the managed temporary directory. Do not treat Git Bash or WSL
as a substitute for native Windows installer tests.

## Formatting And Focused Tests

From the repository root, after the isolated environment setup above, format
before committing and run the basic Go tests. PowerShell does not expand native
command wildcards, so pass the root Go files explicitly:

```powershell
$goFiles = @(Get-ChildItem -LiteralPath . -Filter '*.go' -File | ForEach-Object { $_.FullName })
gofmt -w @goFiles
if ($LASTEXITCODE -ne 0) { throw 'Formatting failed' }
go test ./...
if ($LASTEXITCODE -ne 0) { throw 'Tests failed' }
```

On POSIX shells:

```sh
gofmt -w ./*.go || exit 1
go test ./... || exit 1
```

To generate and inspect coverage separately from the full verifier:

```powershell
$coverage = Join-Path $env:TEMP ('uvpip-coverage-' + [guid]::NewGuid().ToString('N') + '.out')
go test -count=1 "-coverprofile=$coverage" ./...
if ($LASTEXITCODE -ne 0) { throw 'Coverage tests failed' }
go tool cover "-func=$coverage"
if ($LASTEXITCODE -ne 0) { throw 'Coverage inspection failed' }
Write-Host "Coverage profile: $coverage"
```

```sh
coverage=$(mktemp "$TMPDIR/uvpip-coverage.XXXXXX") || exit 1
go test -count=1 -coverprofile="$coverage" ./... || exit 1
go tool cover -func="$coverage" || exit 1
printf 'Coverage profile: %s\n' "$coverage"
```

These focused commands do not replace the verifier or enforce its coverage gate.

## Checks And Scope

The verifiers run a non-mutating `gofmt -l` check on the root Go package,
`go vet ./...`, `go build -o <dedicated-output> ./...`,
`go test -count=1 -coverprofile=<managed-temp>/coverage.out ./...`, and
`go tool cover -func=<profile>`. The total statement coverage must be at least
**90% on each host**; a missing report or failed command is an error. Add new
package directories to the formatting check if the current single-package layout
changes. Plain `go build ./...` is the standard build command but emits an ignored
binary in this checkout; verification always uses `-o` to avoid overwriting it.

They also run `staticcheck ./...`, `govulncheck ./...`, `gitleaks dir --redact --no-banner .`,
`actionlint .github/workflows/ci.yml`, ShellCheck on all current POSIX installer,
uninstaller, shim, and developer scripts, PowerShell parsing, and the native
offline installer suite. `govulncheck` may access the public vulnerability
database, and tool installation uses Go's module proxy/checksum infrastructure.
Installer tests themselves mock downloads and uv setup and do not require network.

To run just installer checks after setting up the isolated environment:

```powershell
powershell.exe -NoLogo -NoProfile -NonInteractive -ExecutionPolicy Bypass -File scripts/check-powershell.ps1
if ($LASTEXITCODE -ne 0) { throw 'PowerShell parse failed' }
powershell.exe -NoLogo -NoProfile -NonInteractive -ExecutionPolicy Bypass -File scripts/test-installers.ps1 -TestRoot $tests -CacheRoot $cache
if ($LASTEXITCODE -ne 0) { throw 'Installer tests failed' }
```

```sh
sh scripts/test-installers.sh "$TEST_ROOT" "$CACHE_ROOT"
```

Inspect the verifier output and the CI run for the exact commit being evaluated.
Git Bash installer checks and cross-builds are not native Linux/macOS execution
evidence. A passing scan reports only what its current database and rules detect.

CI is configured for Ubuntu 24.04, Windows Server 2022, and macOS 14 runners. It
has read-only repository permissions, disables checkout credential persistence and setup-go
caching, pins Actions by full commit SHA, and does not publish or request secrets.
PowerShell parsing is not a full PowerShell static analyzer. Mocked installer
tests do not prove live downloads, every shell's startup rules, all architectures,
or every uv/package combination. Coverage is only one gate, not a compatibility
or security guarantee.

## Manual Release

There is intentionally no automatic publishing workflow.

1. Review the exact source commit, version constant, changelog, security changes,
   and dependency/tool pins. Never relabel old binaries as the revised source.
2. Complete the required fresh-clone verification and require green CI on the
   same source commit on all three operating systems. Manually exercise the real
   binary and installer in disposable native environments, including active venvs
   and supported interactive shells. Record untested platforms explicitly.
3. Build all six targets with Go 1.26.8 and `CGO_ENABLED=0`: `windows/amd64`,
   `windows/arm64`, `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`.
   Use `GOOS`, `GOARCH`, and `go build -trimpath -o <release-directory>/<asset> ./...`.
   Asset names must be `uvpip-<os>-<arch>`, with `.exe` for Windows. Cross-building
   is not native execution testing. Keep the release directory outside the checkout.
4. Compute SHA-256 for each final asset (`Get-FileHash -Algorithm SHA256` on
   Windows, `sha256sum` or `shasum -a 256` on POSIX). Independently review the
   source-to-asset mapping, checksums, and release notes before publication.
5. A maintainer manually publishes a new tag/release with the six reviewed assets
   and a checksum manifest. Do not replace historical v1.0.0 assets. Document any
   actual signing separately; a checksum alone is not a signature.
6. Download the published assets into a disposable environment, verify against
   the reviewed manifest, and smoke-test installation. Only then remove or update
   the README's unreleased-source warning. No release is created by verification.
