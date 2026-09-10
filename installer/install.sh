#!/bin/sh
# uvpip installer for macOS and Linux. Local/offline binary: UVPIP_BINARY=/path/to/uvpip.
set -eu

: "${HOME:?HOME must be set}"
BIN_DIR="$HOME/.uvpip/bin"
RELEASE_BASE="https://github.com/yv3000/uvpip/releases/latest/download"
err() { printf 'uvpip: %s\n' "$*" >&2; exit 1; }
download() {
    if command -v curl >/dev/null 2>&1; then
        curl -fLSs "$1" -o "$2" || return 1
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$1" -O "$2" || return 1
    else
        err 'curl or wget is required to download files'
    fi
    [ -s "$2" ]
}

case "$(uname -s)" in
    Darwin) OS_NAME=darwin ;;
    Linux) OS_NAME=linux ;;
    *) err 'Unsupported OS' ;;
esac
case "$(uname -m)" in
    x86_64) ARCH_NAME=amd64 ;;
    aarch64|arm64) ARCH_NAME=arm64 ;;
    *) err 'Unsupported architecture' ;;
esac

# Only this private staging directory is removed on failure.
STAGE=$(mktemp -d "${TMPDIR:-/tmp}/uvpip-install.XXXXXX")
trap 'rm -rf "$STAGE"' 0
trap 'exit 1' HUP INT TERM
if [ ! -f "$BIN_DIR/uvpip" ]; then
    if [ -n "${UVPIP_BINARY:-}" ]; then
        cp "$UVPIP_BINARY" "$STAGE/uvpip" || err 'Could not copy UVPIP_BINARY'
    else
        download "$RELEASE_BASE/uvpip-$OS_NAME-$ARCH_NAME" "$STAGE/uvpip" || err 'uvpip download failed'
    fi
    [ -s "$STAGE/uvpip" ] || err 'uvpip binary is empty'
    # Optional caller-supplied checksum, not a claim of release signature verification.
    if [ -n "${UVPIP_SHA256:-}" ]; then
        if command -v sha256sum >/dev/null 2>&1; then
            HASH=$(sha256sum "$STAGE/uvpip")
        elif command -v shasum >/dev/null 2>&1; then
            HASH=$(shasum -a 256 "$STAGE/uvpip")
        else
            err 'sha256sum or shasum is required for UVPIP_SHA256'
        fi
        HASH=${HASH%% *}
        [ "$HASH" = "$(printf '%s' "$UVPIP_SHA256" | tr 'A-F' 'a-f')" ] || err 'SHA256 mismatch'
    fi
    chmod +x "$STAGE/uvpip"
    "$STAGE/uvpip" --version >/dev/null 2>&1 || err 'uvpip executable check failed'
else
    "$BIN_DIR/uvpip" --version >/dev/null 2>&1 || err 'Existing uvpip executable check failed'
fi

# Keep explicit uv setup, but never execute a partial/failed download.
if ! command -v uv >/dev/null 2>&1; then
    download https://astral.sh/uv/install.sh "$STAGE/install-uv.sh" || err 'uv installer download failed'
    sh "$STAGE/install-uv.sh" || err 'uv installation failed'
    PATH="$HOME/.local/bin:$HOME/.cargo/bin:$PATH"
    export PATH
fi
uv --version >/dev/null 2>&1 || err 'uv executable check failed'

mkdir -p "$BIN_DIR"
if [ -f "$STAGE/uvpip" ]; then
    mv "$STAGE/uvpip" "$BIN_DIR/uvpip"
fi
for shim in pip pip3; do
    cat > "$STAGE/$shim" <<'SHIM'
#!/bin/sh
exec "$(dirname "$0")/uvpip" "$@"
SHIM
    chmod +x "$STAGE/$shim"
    mv "$STAGE/$shim" "$BIN_DIR/$shim"
done

add_to_config() {
    config=$1
    if [ -f "$config" ]; then
        # Ambiguous/nested/unmatched markers must never trigger a rewrite.
        awk '{ sub(/\r$/, "") }
            $0 == "# --- uvpip start ---" { if (inside) exit 1; inside=1 }
            $0 == "# --- uvpip end ---" { if (!inside) exit 1; inside=0 }
            END { if (inside) exit 1 }' "$config" || err "Unbalanced uvpip markers in $config; repair them manually"
        if awk '{ sub(/\r$/, "") } $0 == "# --- uvpip start ---" { found=1 } END { exit !found }' "$config"; then
            return 0
        fi
    fi
    mkdir -p "$(dirname "$config")"
    if [ -s "$config" ] && [ -n "$(tail -c 1 "$config")" ]; then printf '\n' >> "$config"; fi
    # Runtime HOME references avoid embedding shell-sensitive installation paths.
    if [ "$SHELL_NAME" = fish ]; then
        cat >> "$config" <<'PROFILE'
# --- uvpip start ---
fish_add_path "$HOME/.uvpip/bin"
function pip
    "$HOME/.uvpip/bin/uvpip" $argv
end
function pip3
    "$HOME/.uvpip/bin/uvpip" $argv
end
# --- uvpip end ---
PROFILE
    else
        cat >> "$config" <<'PROFILE'
# --- uvpip start ---
export PATH="$HOME/.uvpip/bin:$PATH"
pip() {
    "$HOME/.uvpip/bin/uvpip" "$@"
}
pip3() {
    "$HOME/.uvpip/bin/uvpip" "$@"
}
# --- uvpip end ---
PROFILE
    fi
}

if [ "${UVPIP_NO_PROFILE:-0}" != 1 ]; then
    SHELL_NAME=${SHELL:-sh}
    SHELL_NAME=${SHELL_NAME##*/}
    case "$SHELL_NAME" in
        zsh)
            add_to_config "$HOME/.zshrc"
            if [ -f "$HOME/.zprofile" ]; then add_to_config "$HOME/.zprofile"; fi
            ;;
        bash)
            if [ "$OS_NAME" = darwin ]; then add_to_config "$HOME/.bash_profile"; fi
            add_to_config "$HOME/.bashrc"
            ;;
        fish) add_to_config "$HOME/.config/fish/config.fish" ;;
        *) add_to_config "$HOME/.profile" ;;
    esac
fi
printf '%s\n' 'uvpip installed. Restart your terminal to use it. Existing pip and uv were not removed.'
