#!/usr/bin/env bash
# install.sh — uvpip installer for macOS and Linux
# Run with: curl -fsSL https://raw.githubusercontent.com/yv3000/uvpip/main/installer/install.sh | sh

set -e

INSTALL_DIR="$HOME/.uvpip"
BIN_DIR="$INSTALL_DIR/bin"
RELEASE_BASE="https://github.com/yv3000/uvpip/releases/latest/download"

ok()  { printf "  [OK] %s\n" "$1"; }
wrn() { printf "  [!!] %s\n" "$1"; }
err() { printf "  [ERR] %s\n" "$1" >&2; exit 1; }
nfo() { printf "  [->] %s\n" "$1"; }

echo ""
echo "  uvpip installer for macOS / Linux"
echo "  ----------------------------------------"
echo ""

# ─── Step 1: Detect OS and architecture ──────────────────────────────────────
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Darwin)  OS_NAME="darwin" ;;
    Linux)   OS_NAME="linux" ;;
    *)       err "Unsupported OS: $OS" ;;
esac

case "$ARCH" in
    x86_64)          ARCH_NAME="amd64" ;;
    aarch64|arm64)   ARCH_NAME="arm64" ;;
    *)               err "Unsupported architecture: $ARCH" ;;
esac

BINARY_NAME="uvpip-${OS_NAME}-${ARCH_NAME}"
ok "Detected: $OS_NAME / $ARCH_NAME -> $BINARY_NAME"

# ─── Step 2: Check if already installed ──────────────────────────────────────
if [ -f "$BIN_DIR/uvpip" ]; then
    wrn "uvpip is already installed at $BIN_DIR/uvpip"
    nfo "To reinstall, run the uninstaller first:"
    nfo "curl -fsSL https://raw.githubusercontent.com/yv3000/uvpip/main/uninstaller/uninstall.sh | sh"
    echo ""
    exit 0
fi

# ─── Step 3: Check / install uv ──────────────────────────────────────────────
UV_PATH=""
if command -v uv >/dev/null 2>&1; then
    UV_VER="$(uv --version 2>/dev/null || echo 'unknown')"
    ok "uv already installed: $UV_VER"
    UV_PATH="$(command -v uv)"
else
    nfo "uv not found. Installing uv automatically..."
    if curl -fsSL https://astral.sh/uv/install.sh | sh; then
        # Refresh PATH in current shell
        export PATH="$HOME/.local/bin:$HOME/.cargo/bin:$PATH"
        if command -v uv >/dev/null 2>&1; then
            UV_VER="$(uv --version 2>/dev/null || echo 'unknown')"
            ok "uv installed: $UV_VER"
            UV_PATH="$(command -v uv)"
        else
            err "uv installed but not found in PATH. Try: export PATH=\"\$HOME/.local/bin:\$PATH\" then re-run."
        fi
    else
        err "Failed to install uv. Install manually: https://docs.astral.sh/uv/getting-started/installation/"
    fi
fi

# ─── Step 4: Check pip / Python ──────────────────────────────────────────────
if command -v pip >/dev/null 2>&1; then
    PIP_VER="$(pip --version 2>/dev/null || echo 'unknown')"
    ok "pip already available: $PIP_VER"
elif command -v pip3 >/dev/null 2>&1; then
    PIP_VER="$(pip3 --version 2>/dev/null || echo 'unknown')"
    ok "pip3 available: $PIP_VER"
elif command -v python3 >/dev/null 2>&1; then
    ok "Python3 found but pip missing. uvpip will use uv directly."
else
    wrn "Neither pip nor Python found. uvpip will use uv directly."
fi

# ─── Step 5: Create install directory ────────────────────────────────────────
mkdir -p "$BIN_DIR"
ok "Created $INSTALL_DIR"

# ─── Step 6: Download uvpip binary ───────────────────────────────────────────
DOWNLOAD_URL="$RELEASE_BASE/$BINARY_NAME"
nfo "Downloading $BINARY_NAME from GitHub releases..."

if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$DOWNLOAD_URL" -o "$BIN_DIR/uvpip"
elif command -v wget >/dev/null 2>&1; then
    wget -q "$DOWNLOAD_URL" -O "$BIN_DIR/uvpip"
else
    err "Neither curl nor wget found. Install one and retry."
fi

chmod +x "$BIN_DIR/uvpip"
ok "Downloaded uvpip binary to $BIN_DIR/uvpip"

# ─── Step 7: Create pip and pip3 shim scripts ────────────────────────────────
cat > "$BIN_DIR/pip" << 'SHIMEOF'
#!/usr/bin/env sh
exec "$(dirname "$0")/uvpip" "$@"
SHIMEOF

cat > "$BIN_DIR/pip3" << 'SHIMEOF'
#!/usr/bin/env sh
exec "$(dirname "$0")/uvpip" "$@"
SHIMEOF

chmod +x "$BIN_DIR/pip" "$BIN_DIR/pip3"
ok "Created pip and pip3 shim scripts"

# ─── Step 8: Add to PATH in shell config ─────────────────────────────────────
SHELL_NAME="$(basename "${SHELL:-sh}")"
PATH_EXPORT="export PATH=\"$BIN_DIR:\$PATH\""

add_to_shell_config() {
    local config_file="$1"
    if [ -f "$config_file" ] || [ "$2" = "force" ]; then
        if ! grep -q "uvpip" "$config_file" 2>/dev/null; then
            echo "" >> "$config_file"
            echo "# uvpip — transparent pip -> uv wrapper" >> "$config_file"
            echo "$PATH_EXPORT" >> "$config_file"
            ok "Added PATH entry to $config_file"
            return 0
        else
            wrn "PATH entry already in $config_file"
            return 0
        fi
    fi
    return 1
}

PATH_ADDED=false

case "$SHELL_NAME" in
    zsh)
        add_to_shell_config "$HOME/.zshrc" force && PATH_ADDED=true
        add_to_shell_config "$HOME/.zprofile"
        ;;
    bash)
        if [ "$OS_NAME" = "darwin" ]; then
            add_to_shell_config "$HOME/.bash_profile" force && PATH_ADDED=true
        fi
        add_to_shell_config "$HOME/.bashrc" force && PATH_ADDED=true
        ;;
    fish)
        FISH_CONFIG="$HOME/.config/fish/config.fish"
        mkdir -p "$(dirname "$FISH_CONFIG")"
        if ! grep -q "uvpip" "$FISH_CONFIG" 2>/dev/null; then
            echo "" >> "$FISH_CONFIG"
            echo "# uvpip" >> "$FISH_CONFIG"
            echo "fish_add_path $BIN_DIR" >> "$FISH_CONFIG"
            ok "Added PATH entry to $FISH_CONFIG"
            PATH_ADDED=true
        fi
        ;;
    *)
        # Fallback: try common config files
        add_to_shell_config "$HOME/.profile" force && PATH_ADDED=true
        ;;
esac

if [ "$PATH_ADDED" = false ]; then
    wrn "Could not detect shell config file."
    nfo "Add this line manually to your shell config:"
    nfo "$PATH_EXPORT"
fi

# Refresh in current session
export PATH="$BIN_DIR:$PATH"
ok "Refreshed current session PATH"

# ─── Step 9: Verify install ───────────────────────────────────────────────────
if "$BIN_DIR/uvpip" --version >/dev/null 2>&1; then
    VER="$("$BIN_DIR/uvpip" --version 2>/dev/null)"
    ok "uvpip verified: $VER"
else
    wrn "uvpip binary ran but could not verify version"
fi

# ─── Done ─────────────────────────────────────────────────────────────────────
echo ""
echo "  ----------------------------------------"
echo ""
ok "uvpip installed successfully"
ok "pip and pip3 now route through uv (10-100x faster)"
echo ""
nfo "IMPORTANT: Restart your terminal (or run: source ~/.zshrc) for PATH changes to take effect."
nfo "Then run:  pip install requests"
nfo "Or run:   uvpip doctor    to verify everything is working."
echo ""
