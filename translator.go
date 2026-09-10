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

	// "cache" is a top-level uv command, not a "uv pip" subcommand.
	if cmd == "cache" {
		return append([]string{"cache"}, args[1:]...)
	}

	// Commands that need special handling
	switch cmd {
	case "uninstall", "remove":
		filteredArgs := []string{}
		options := true
		for _, a := range args[1:] {
			if a == "--" {
				options = false
			}
			if options && (a == "-y" || a == "--yes") {
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
