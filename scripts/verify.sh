#!/bin/sh
# Development only. Artifacts and caches stay outside the checkout.
set -eu
export LC_ALL=C
: "${1:?Pass an existing test artifact root}"
: "${2:?Pass an existing cache/home/temp root}"
[ -d "$1" ] && [ -d "$2" ]
TEST_ROOT=$(CDPATH='' cd -- "$1" && pwd)
CACHE_ROOT=$(CDPATH='' cd -- "$2" && pwd)
REPO=$(CDPATH='' cd -- "$(dirname "$0")/.." && pwd)
export HOME="$CACHE_ROOT/uvpip-dev/home" USERPROFILE="$CACHE_ROOT/uvpip-dev/home"
export TMPDIR="$CACHE_ROOT/uvpip-dev/tmp" TMP="$CACHE_ROOT/uvpip-dev/tmp" TEMP="$CACHE_ROOT/uvpip-dev/tmp"
export GOPATH="$CACHE_ROOT/uvpip-dev/go" GOBIN="$CACHE_ROOT/uvpip-dev/bin"
export GOCACHE="$CACHE_ROOT/uvpip-dev/go-build" GOMODCACHE="$CACHE_ROOT/uvpip-dev/go-mod" GOTMPDIR="$TMPDIR"
export XDG_CACHE_HOME="$CACHE_ROOT/uvpip-dev/cache" XDG_CONFIG_HOME="$CACHE_ROOT/uvpip-dev/config" XDG_DATA_HOME="$CACHE_ROOT/uvpip-dev/data"
export APPDATA="$XDG_CONFIG_HOME" LOCALAPPDATA="$XDG_CACHE_HOME"
export UV_CACHE_DIR="$XDG_CACHE_HOME/uv" PIP_CACHE_DIR="$XDG_CACHE_HOME/pip" npm_config_cache="$XDG_CACHE_HOME/npm"
export DOTNET_CLI_HOME="$HOME" NUGET_PACKAGES="$XDG_CACHE_HOME/nuget" PYTHONPYCACHEPREFIX="$XDG_CACHE_HOME/pycache"
export GOTOOLCHAIN=local GOFLAGS=-mod=readonly GOENV=off
export GOTELEMETRYDIR="$CACHE_ROOT/uvpip-dev/telemetry" STATICCHECK_CACHE="$CACHE_ROOT/uvpip-dev/staticcheck"
export PATH="$GOBIN:$PATH"
mkdir -p "$HOME" "$TMPDIR" "$GOPATH" "$GOBIN" "$GOCACHE" "$GOMODCACHE" "$XDG_CACHE_HOME" "$XDG_CONFIG_HOME" "$XDG_DATA_HOME"
ARTIFACTS=$(mktemp -d "$TMPDIR/uvpip-verify.XXXXXX")
printf 'Verification artifacts: %s\n' "$ARTIFACTS"
cd "$REPO"
[ "$(go env GOVERSION)" = go1.26.8 ] || { printf 'Go 1.26.8 is required\n' >&2; exit 1; }
UNFORMATTED=$(gofmt -l ./*.go)
[ -z "$UNFORMATTED" ] || { printf 'Run gofmt on:\n%s\n' "$UNFORMATTED" >&2; exit 1; }
go vet ./...
go build -o "$ARTIFACTS/uvpip-build" ./...
go test -count=1 -coverprofile="$ARTIFACTS/coverage.out" ./...
go tool cover -func="$ARTIFACTS/coverage.out" > "$ARTIFACTS/coverage.txt"
cat "$ARTIFACTS/coverage.txt"
awk '$1 == "total:" && $2 == "(statements)" && $3 ~ /^[0-9]+([.][0-9]+)?%$/ { found=1; coverage=$3+0 }
     END { if (!found || coverage < 90) { print "Total statement coverage must be at least 90%"; exit 1 } }' "$ARTIFACTS/coverage.txt"
if [ "$(uname -s)" = Linux ]; then go test -race -count=1 ./...; fi
staticcheck ./...
govulncheck ./...
gitleaks dir --redact --no-banner .
actionlint .github/workflows/ci.yml
shellcheck installer/*.sh uninstaller/*.sh scripts/*.sh bin/*.sh
pwsh -NoLogo -NoProfile -NonInteractive -File scripts/check-powershell.ps1
sh scripts/test-installers.sh "$TEST_ROOT" "$CACHE_ROOT"
printf '%s\n' 'PASS: POSIX verification (90% minimum statement coverage)'
