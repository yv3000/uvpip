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

	// Commands that map directly under "uv pip"
	directPassthrough := map[string]bool{
		"install":    true,
		"list":       true,
		"show":       true,
		"freeze":     true,
		"check":      true,
		"download":   true,
		"wheel":      true,
		"hash":       true,
		"completion": true,
		"debug":      true,
		"inspect":    true,
		"config":     true,
		"cache":      true,
		"index":      true,
		"search":     true,
	}

	if directPassthrough[cmd] {
		// pip install X -> uv pip install X
		return append([]string{"pip"}, args...)
	}

	// Commands that need special handling
	switch cmd {
	case "uninstall", "remove":
		// pip uninstall X -> uv pip uninstall doesn't prompt for confirmation by default — pass -y if not already there
		newArgs := []string{"pip", "uninstall"}
		hasY := false
		for _, a := range args[1:] {
			if a == "-y" || a == "--yes" {
				hasY = true
			}
		}
		if !hasY {
			newArgs = append(newArgs, "-y")
		}
		return append(newArgs, args[1:]...)

	case "upgrade":
		// pip upgrade X -> uv pip install --upgrade X
		return append([]string{"pip", "install", "--upgrade"}, args[1:]...)

	default:
		// Unknown/new pip commands: pass through under uv pip as-is
		return append([]string{"pip"}, args...)
	}
}
