package main

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
)

// errUsage marks input rejected before uv starts; runCLI maps it to exit status 2.
// It wraps ErrInvalidArgs so callers can inspect failure types programmatically.
var errUsage = ErrUsage

// CLIInputs specifies all user-supplied inputs evaluated at CLI boundaries:
// command arguments and environment configuration overrides.
type CLIInputs struct {
	Args       []string
	UVOverride string
}

// ValidateInputs runs boundary validation on all CLI entry point inputs.
// It returns a typed error wrapping ErrInvalidArgs if arguments are invalid,
// or ErrInvalidOverride if environment configuration overrides are malformed.
func ValidateInputs(inputs CLIInputs) error {
	if len(inputs.Args) > 0 || inputs.UVOverride == "" {
		if err := ValidateArgs(inputs.Args); err != nil {
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

// Schema boundary limits on argument count and argument length.
const (
	MaxArgCount = 512
	MaxArgLen   = 4096
	maxArgCount = MaxArgCount
	maxArgLen   = MaxArgLen
)

// InputValidationPatterns defines schema validation constraints for CLI arguments.
type InputValidationPatterns struct {
	MaxArgCount       int
	MaxArgLen         int
	DisallowNull      bool
	DisallowEmpty     bool
	DisallowFlagClash bool
}

// CanonicalValidationSchema represents the authoritative CLI input boundary schema.
var CanonicalValidationSchema = InputValidationPatterns{
	MaxArgCount:       MaxArgCount,
	MaxArgLen:         MaxArgLen,
	DisallowNull:      true,
	DisallowEmpty:     true,
	DisallowFlagClash: true,
}

// Validate applies schema validation rules to the supplied arguments.
func (p InputValidationPatterns) Validate(args []string) error {
	return ValidateArgs(args)
}

// validateArgBounds verifies schema boundary limits on argument count, argument length, and null-byte injection.
func validateArgBounds(args []string) error {
	if len(args) > MaxArgCount {
		return fmt.Errorf("%w: argument count %d exceeds maximum limit (%d)", ErrArgBounds, len(args), MaxArgCount)
	}
	for i, a := range args {
		if len(a) > MaxArgLen {
			return fmt.Errorf("%w: argument %d length %d exceeds maximum limit (%d)", ErrArgBounds, i, len(a), MaxArgLen)
		}
		if strings.ContainsRune(a, '\x00') {
			return fmt.Errorf("%w: argument %d contains forbidden null byte", ErrNullByte, i)
		}
	}
	return nil
}

// ValidateArgs centralizes argument bounds, blank-command rejection and
// conflicting verbosity checks before translation or subprocess dispatch.
// Values are never echoed: arguments can embed index credentials.
func ValidateArgs(args []string) error {
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return fmt.Errorf("%w: pip command is empty", ErrEmptyCommand)
	}
	if err := validateArgBounds(args); err != nil {
		return err
	}
	return validateTranslationFlags(args)
}

// ValidateArgsWithLogger validates CLI arguments and records structured diagnostics with the given logger.
func ValidateArgsWithLogger(args []string, log *slog.Logger) error {
	err := ValidateArgs(args)
	if err != nil {
		if log != nil {
			log.Info("argument validation failed", "component", "uvpip", "error", err)
		}
		return err
	}
	if log != nil {
		log.Debug("argument validation passed", "component", "uvpip", "arg_count", len(args))
	}
	return nil
}

// validateUVOverride checks a non-empty UVPIP_UV before path lookup, so common
// quoting mistakes get a precise message instead of "executable not found".
func validateUVOverride(path string) error {
	switch {
	case strings.TrimSpace(path) == "":
		return fmt.Errorf("%w: UVPIP_UV is set but blank; unset it or give an absolute path", ErrInvalidOverride)
	case strings.TrimSpace(path) != path:
		return fmt.Errorf("%w: UVPIP_UV has leading or trailing whitespace; check its quoting", ErrInvalidOverride)
	case !filepath.IsAbs(path):
		return fmt.Errorf("%w: UVPIP_UV must be an absolute executable path", ErrInvalidOverride)
	}
	return nil
}
