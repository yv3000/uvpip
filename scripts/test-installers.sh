#!/bin/sh
# Offline only. Pass existing artifact and cache roots; no real home is used.
# sh scripts/test-installers.sh /path/to/test-root /path/to/cache-root
set -eu
: "${1:?Pass an existing test artifact root}"
: "${2:?Pass an existing cache/home/temp root}"
[ -d "$1" ] && [ -d "$2" ]
REPO=$(CDPATH='' cd -- "$(dirname "$0")/.." && pwd)
TEST_ROOT=$1
CACHE_ROOT=$2
export TEMP="$CACHE_ROOT" TMP="$CACHE_ROOT" TMPDIR="$CACHE_ROOT"
export HOME="$CACHE_ROOT" USERPROFILE="$CACHE_ROOT" APPDATA="$CACHE_ROOT" LOCALAPPDATA="$CACHE_ROOT"
export XDG_CACHE_HOME="$CACHE_ROOT" XDG_CONFIG_HOME="$CACHE_ROOT" XDG_DATA_HOME="$CACHE_ROOT"
export GOCACHE="$CACHE_ROOT" GOMODCACHE="$CACHE_ROOT" GOPATH="$CACHE_ROOT"
export UV_CACHE_DIR="$CACHE_ROOT" PIP_CACHE_DIR="$CACHE_ROOT" npm_config_cache="$CACHE_ROOT"
export DOTNET_CLI_HOME="$CACHE_ROOT" NUGET_PACKAGES="$CACHE_ROOT" PYTHONPYCACHEPREFIX="$CACHE_ROOT"
unset ENV BASH_ENV UVPIP_BINARY UVPIP_SHA256 UVPIP_NO_PROFILE
WORK=$(mktemp -d "$TEST_ROOT/uvpip-shell.XXXXXX")
CACHE=$(mktemp -d "$CACHE_ROOT/uvpip-shell.XXXXXX")
trap 'rm -rf "$WORK" "$CACHE"' 0
trap 'exit 1' HUP INT TERM
export TMPDIR="$CACHE" TMP="$CACHE" TEMP="$CACHE"
export HOME="$CACHE/home spaces 'single' \"double\""
export USERPROFILE="$HOME"
mkdir -p "$WORK/mock" "$HOME"
cat > "$WORK/mock/uv" <<'MOCK'
#!/bin/sh
printf 'uv offline\n'
MOCK
cat > "$WORK/mock/uname" <<'MOCK'
#!/bin/sh
case "$1" in -s) printf '%s\n' "${MOCK_OS:-Linux}" ;; -m) echo x86_64 ;; *) exit 1 ;; esac
MOCK
cat > "$WORK/mock/curl" <<'MOCK'
#!/bin/sh
# Production calls: curl -fLSs URL -o OUTPUT. Never access the network.
printf 'partial\n' > "$4"
exit 22
MOCK
cp "$WORK/mock/curl" "$WORK/mock/wget"
cat > "$WORK/fixture" <<'MOCK'
#!/bin/sh
if [ "${1:-}" = --version ]; then echo 'uvpip offline'; exit 0; fi
if [ "${1:-}" = --fail ]; then exit 23; fi
printf '<%s>\n' "$@"
MOCK
chmod +x "$WORK/mock/uv" "$WORK/mock/uname" "$WORK/mock/curl" "$WORK/mock/wget" "$WORK/fixture"
PATH="$WORK/mock:$PATH"
export PATH UVPIP_BINARY="$WORK/fixture"
INSTALL="$REPO/installer/install.sh"
UNINSTALL="$REPO/uninstaller/uninstall.sh"
fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
install() { sh "$INSTALL" > "$WORK/install.log" 2>&1 || fail 'install failed (offline)'; }
uninstall() { sh "$UNINSTALL" > "$WORK/uninstall.log" 2>&1 || fail 'uninstall failed'; }
no_stage() {
    for entry in "$CACHE"/uvpip-install.* "$CACHE"/uvpip-uninstall.*; do
        [ ! -e "$entry" ] || fail 'staging directory leaked'
    done
}

# Fresh/missing profile for each supported shell, including absent optional zprofile.
for shell in bash zsh fish sh; do
    export SHELL="/bin/$shell"
    case "$shell" in
        bash) profile="$HOME/.bashrc" ;;
        zsh) profile="$HOME/.zshrc" ;;
        fish) profile="$HOME/.config/fish/config.fish" ;;
        sh) profile="$HOME/.profile" ;;
    esac
    install
    [ -f "$profile" ] || fail "missing $shell profile"
    [ ! -e "$HOME/.zprofile" ] || fail 'created optional zprofile'
    cp "$profile" "$WORK/installed"
    install
    cmp -s "$profile" "$WORK/installed" || fail 'install not idempotent'
    if grep -q 'export -f' "$profile"; then fail 'nonportable function export'; fi
    if [ "$shell" != fish ]; then
        result=$(sh -c '. "$1"; pip "a b" "single'"'"'quote" "*"' sh "$profile")
        [ "$result" = "<a b>
<single'quote>
<*>" ] || fail 'profile argument forwarding'
    fi
    result=$("$HOME/.uvpip/bin/pip" 'a b' '*')
    [ "$result" = '<a b>
<*>' ] || fail 'shim argument forwarding'
    status=0
    "$HOME/.uvpip/bin/pip3" --fail || status=$?
    [ "$status" = 23 ] || fail 'shim exit status'
    uninstall
    cp "$profile" "$WORK/removed"
    uninstall
    cmp -s "$profile" "$WORK/removed" || fail 'uninstall not idempotent'
    [ ! -e "$HOME/.uvpip/bin/uvpip" ] || fail 'binary not removed'
done

# Preserve unrelated mentions, near-markers, CRLF, and no final newline byte-for-byte.
export SHELL=/bin/bash
printf '# custom .uvpip setting\r\n# --- uvpip start --- extra\r\n' > "$WORK/prefix"
printf '# keep uvpip tail without newline' > "$WORK/suffix"
cat "$WORK/prefix" > "$HOME/.bashrc"
install
cat "$WORK/suffix" >> "$HOME/.bashrc"
cat "$WORK/prefix" "$WORK/suffix" > "$WORK/expected"
printf 'keep' > "$HOME/.uvpip/keep.txt"
uninstall
cmp -s "$HOME/.bashrc" "$WORK/expected" || fail 'unrelated profile bytes changed'
[ "$(cat "$HOME/.uvpip/keep.txt")" = keep ] || fail 'unrelated install file removed'

# Malformed markers fail closed, without truncating or guessing ownership.
for content in '# --- uvpip start ---' '# --- uvpip end ---' \
    '# --- uvpip start ---
# --- uvpip start ---
# --- uvpip end ---'; do
    printf '%s\nkeep this tail' "$content" > "$HOME/.bashrc"
    cp "$HOME/.bashrc" "$WORK/original"
    if sh "$UNINSTALL" > "$WORK/error.log" 2>&1; then fail 'accepted malformed markers'; fi
    cmp -s "$HOME/.bashrc" "$WORK/original" || fail 'malformed profile modified'
    if sh "$INSTALL" > "$WORK/error.log" 2>&1; then fail 'installer accepted malformed markers'; fi
    cmp -s "$HOME/.bashrc" "$WORK/original" || fail 'installer changed malformed profile'
    UVPIP_NO_PROFILE=1 sh "$UNINSTALL" > "$WORK/uninstall.log" 2>&1
done
printf 'keep\n# --- uvpip start ---\nowned\n# --- uvpip end ---\n# --- uvpip start ---\nowned\n# --- uvpip end ---' > "$HOME/.bashrc"
printf 'keep\n' > "$WORK/expected"
uninstall
cmp -s "$HOME/.bashrc" "$WORK/expected" || fail 'exact multi-block removal'
printf '# user config\n' > "$HOME/.bashrc"
cp "$HOME/.bashrc" "$WORK/original"

# Download failure, rejected executable/checksum, and uv setup failure leave no target.
unset UVPIP_BINARY
if sh "$INSTALL" > "$WORK/error.log" 2>&1; then fail 'download failure succeeded'; fi
[ ! -e "$HOME/.uvpip/bin/uvpip" ] || fail 'partial target left'
no_stage
export UVPIP_BINARY="$WORK/fixture" UVPIP_SHA256=bad
if sh "$INSTALL" > "$WORK/error.log" 2>&1; then fail 'checksum mismatch succeeded'; fi
unset UVPIP_SHA256
printf '#!/bin/sh\nexit 9\n' > "$WORK/bad"
export UVPIP_BINARY="$WORK/bad"
if sh "$INSTALL" > "$WORK/error.log" 2>&1; then fail 'invalid binary succeeded'; fi
export UVPIP_BINARY="$WORK/fixture"
# Isolate PATH so a real installed uv cannot satisfy the check.
mkdir "$WORK/no-uv"
for tool in uname curl wget mktemp rm cp chmod sh; do
    ln -s "$(command -v "$tool")" "$WORK/no-uv/$tool"
done
if PATH="$WORK/no-uv" sh "$INSTALL" > "$WORK/error.log" 2>&1; then fail 'uv download failure succeeded'; fi
cmp -s "$HOME/.bashrc" "$WORK/original" || fail 'failed install changed profile'
[ ! -e "$HOME/.uvpip/bin/uvpip" ] || fail 'failed install left target'
[ "$(cat "$HOME/.uvpip/keep.txt")" = keep ] || fail 'failure deleted unrelated file'
no_stage

# Successful checked download and optional checksum use the same staging path.
cat > "$WORK/mock/curl" <<'MOCK'
#!/bin/sh
cp "$MOCK_BINARY" "$4"
MOCK
export MOCK_BINARY="$WORK/fixture"
unset UVPIP_BINARY
if command -v sha256sum >/dev/null 2>&1; then
    HASH=$(sha256sum "$WORK/fixture")
else
    HASH=$(shasum -a 256 "$WORK/fixture")
fi
export UVPIP_SHA256="${HASH%% *}"
install
uninstall
unset UVPIP_SHA256
export UVPIP_BINARY="$WORK/fixture"

# Darwin bash creates both login and interactive profiles; unset SHELL uses .profile.
export MOCK_OS=Darwin
install
[ -f "$HOME/.bash_profile" ] || fail 'missing Darwin login profile'
uninstall
unset MOCK_OS SHELL
install
uninstall

# Isolated opt-out and repository shims use the same relative executable contract.
export UVPIP_NO_PROFILE=1
install
cmp -s "$HOME/.bashrc" "$WORK/original" || fail 'NoProfile changed profile'
for shim in pip pip3; do
    cp "$REPO/bin/$shim.sh" "$HOME/.uvpip/bin/$shim"
    status=0
    sh "$HOME/.uvpip/bin/$shim" --fail || status=$?
    [ "$status" = 23 ] || fail 'repository shim exit status'
done
uninstall
no_stage
printf '%s\n' 'PASS: offline POSIX installer/uninstaller tests'
