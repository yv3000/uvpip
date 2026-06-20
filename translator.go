package main

import "strings"

// translateToUV converts pip CLI args into the equivalent uv CLI args.
// uv pip subcommand is designed to be pip-compatible so most flags pass through 1:1.
// We just prepend "pip" to route through "uv pip <args>".
func translateToUV(args []string) []string {
	if len(args) == 0 {
		return []string{"pip"}
	}

	cmd := strings.ToLower(args[0])

	// Commands that map directly under "uv pip" — verified against
	// `uv pip --help`: compile, sync, install, uninstall, freeze, list, show, tree, check.
	// (uninstall is handled separately below because it needs arg filtering.)
	directPassthrough := map[string]bool{
		"install": true,
		"list":    true,
		"show":    true,
		"freeze":  true,
		"check":   true,
		"compile": true,
		"sync":    true,
		"tree":    true,
	}

	if directPassthrough[cmd] {
		// pip install X -> uv pip install X
		return append([]string{"pip"}, args...)
	}

	// "cache" is a top-level uv command, not a "uv pip" subcommand.
	if cmd == "cache" {
		return append([]string{"cache"}, args[1:]...)
	}

	// Commands that need special handling
	switch cmd {
	case "uninstall", "remove":
		filteredArgs := []string{}
		for _, a := range args[1:] {
			if a == "-y" || a == "--yes" {
				continue
			}
			filteredArgs = append(filteredArgs, a)
		}
		return append([]string{"pip", "uninstall"}, filteredArgs...)

	case "upgrade":
		// pip upgrade X -> uv pip install --upgrade X
		return append([]string{"pip", "install", "--upgrade"}, args[1:]...)

	default:
		// Unknown/new pip commands: pass through under uv pip as-is
		return append([]string{"pip"}, args...)
	}
}
