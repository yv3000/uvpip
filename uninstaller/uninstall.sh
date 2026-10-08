#!/bin/sh
# Remove uvpip-owned files and complete, exact profile blocks only.
set -eu
: "${HOME:?HOME must be set}"
BIN_DIR="$HOME/.uvpip/bin"
STAGE=$(mktemp -d "${TMPDIR:-/tmp}/uvpip-uninstall.XXXXXX")
trap 'rm -rf "$STAGE"' 0
trap 'exit 1' HUP INT TERM

remove_from_config() {
    config=$1
    [ -f "$config" ] || return 0
    awk '{ sub(/\r$/, "") }
        $0 == "# --- uvpip start ---" { if (inside) exit 1; inside=1 }
        $0 == "# --- uvpip end ---" { if (!inside) exit 1; inside=0 }
        END { if (inside) exit 1 }' "$config" || {
            printf 'uvpip: Unbalanced markers in %s; file left unchanged. Repair manually and retry.\n' "$config" >&2
            exit 1
        }
    # read/printf retain CRLF and a missing final newline outside the block.
    inside=0
    CR=$(printf '\r')
    while :; do
        line=
        newline=1
        IFS= read -r line || newline=0
        [ "$newline" = 1 ] || [ -n "$line" ] || break
        case "${line%"$CR"}" in
            '# --- uvpip start ---') inside=1 ;;
            '# --- uvpip end ---') inside=0 ;;
            *)
                if [ "$inside" = 0 ]; then
                    printf '%s' "$line"
                    if [ "$newline" = 1 ]; then printf '\n'; fi
                fi
                ;;
        esac
        [ "$newline" = 1 ] || break
    done < "$config" > "$STAGE/profile"
    if ! cmp -s "$config" "$STAGE/profile"; then
        cat "$STAGE/profile" > "$config"
    fi
}

if [ "${UVPIP_NO_PROFILE:-0}" != 1 ]; then
    for config in "$HOME/.zshrc" "$HOME/.zprofile" "$HOME/.bashrc" \
        "$HOME/.bash_profile" "$HOME/.profile" "$HOME/.config/fish/config.fish"; do
        remove_from_config "$config"
    done
fi
# Do not recursively delete a directory that may contain unrelated user files.
rm -f "$BIN_DIR/uvpip" "$BIN_DIR/pip" "$BIN_DIR/pip3"
rmdir "$BIN_DIR" "$HOME/.uvpip" 2>/dev/null || :
printf '%s\n' 'uvpip uninstalled. Restart your terminal. Existing pip and uv were not removed.'
