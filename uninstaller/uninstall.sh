#!/usr/bin/env bash
# uninstall.sh — uvpip uninstaller for macOS and Linux
# Run with: curl -fsSL https://raw.githubusercontent.com/yv3000/uvpip/main/uninstaller/uninstall.sh | sh

set -e

INSTALL_DIR="$HOME/.uvpip"
BIN_DIR="$INSTALL_DIR/bin"

ok()  { printf "  [OK] %s\n" "$1"; }
wrn() { printf "  [!!] %s\n" "$1"; }
err() { printf "  [ERR] %s\n" "$1" >&2; exit 1; }
nfo() { printf "  [->] %s\n" "$1"; }

echo ""
echo "  uvpip uninstaller for macOS / Linux"
echo "  ----------------------------------------"
echo ""

# ─── Step 1: Check if installed ──────────────────────────────────────────────
if [ ! -d "$INSTALL_DIR" ]; then
    wrn "uvpip does not appear to be installed (no directory at $INSTALL_DIR)"
    exit 0
fi

# ─── Step 2: Remove PATH entry from shell configs ────────────────────────────
remove_from_config() {
    local config_file="$1"
    if [ -f "$config_file" ] && grep -q "uvpip" "$config_file" 2>/dev/null; then
        # Remove uvpip comment and PATH lines
        if sed --version 2>/dev/null | grep -q GNU; then
            # GNU sed (Linux)
            sed -i '/# uvpip/d' "$config_file"
            sed -i '/\.uvpip/d' "$config_file"
        else
            # BSD sed (macOS)
            sed -i '' '/# uvpip/d' "$config_file"
            sed -i '' '/\.uvpip/d' "$config_file"
        fi
        ok "Removed uvpip PATH entry from $config_file"
    fi
}

remove_from_config "$HOME/.zshrc"
remove_from_config "$HOME/.zprofile"
remove_from_config "$HOME/.bashrc"
remove_from_config "$HOME/.bash_profile"
remove_from_config "$HOME/.profile"

FISH_CONFIG="$HOME/.config/fish/config.fish"
if [ -f "$FISH_CONFIG" ] && grep -q "uvpip" "$FISH_CONFIG" 2>/dev/null; then
    if sed --version 2>/dev/null | grep -q GNU; then
        sed -i '/uvpip/d' "$FISH_CONFIG"
    else
        sed -i '' '/uvpip/d' "$FISH_CONFIG"
    fi
    ok "Removed uvpip PATH entry from $FISH_CONFIG"
fi

# ─── Step 3: Delete install directory ────────────────────────────────────────
rm -rf "$INSTALL_DIR"
ok "Deleted $INSTALL_DIR"

# ─── Done ─────────────────────────────────────────────────────────────────────
echo ""
echo "  ----------------------------------------"
echo ""
ok "uvpip uninstalled"
ok "Original pip restored"
nfo "uv itself was NOT removed. You may still use it directly."
nfo "Restart your terminal for changes to take full effect."
echo ""
