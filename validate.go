package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// errUsage marks input rejected before uv starts; runCLI maps it to exit status 2.
var errUsage = errors.New("usage error")

// CLIInputs specifies all user-supplied inputs evaluated at CLI boundaries:
// command arguments and environment configuration overrides.
type CLIInputs struct {
	Args       []string
	UVOverride string
}

// ValidateInputs runs boundary validation on all CLI entry point inputs.
// It returns an error wrapping errUsage if arguments are invalid, or an error
// describing invalid environment configuration overrides.
func ValidateInputs(inputs CLIInputs) error {
	if len(inputs.Args) > 0 || inputs.UVOverride == "" {
		if err := validateArgs(inputs.Args); err != nil {
			return err
		}
	}
	if inputs.UVOverride != "" {
		if err := validateUVOverride(inputs.UVOverride); err != nil {
			return err
		}
	}
	return nil
}

// validateArgs rejects a blank pip command, which uv would only report as an
// unrecognized empty subcommand. Values are never echoed: arguments can embed
// index credentials.
func validateArgs(args []string) error {
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return fmt.Errorf("%w: pip command is empty", errUsage)
	}
	return nil
}

// validateUVOverride checks a non-empty UVPIP_UV before path lookup, so common
// quoting mistakes get a precise message instead of "executable not found".
func validateUVOverride(path string) error {
	switch {
	case strings.TrimSpace(path) == "":
		return fmt.Errorf("UVPIP_UV is set but blank; unset it or give an absolute path")
	case strings.TrimSpace(path) != path:
		return fmt.Errorf("UVPIP_UV has leading or trailing whitespace; check its quoting")
	case !filepath.IsAbs(path):
		return fmt.Errorf("UVPIP_UV must be an absolute executable path")
	}
	return nil
}
