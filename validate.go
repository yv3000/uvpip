package main

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"regexp"
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

var (
	// CommandPattern validates standard CLI commands, subcommands, and flags.
	CommandPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
	// SafeArgPattern prevents null-byte injection across arbitrary CLI argument values.
	SafeArgPattern = regexp.MustCompile(`^[^\x00]+$`)
	// SubcommandPattern enforces standard alphanumeric command tokens.
	SubcommandPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
)

// InputValidationPatterns defines formal schema validation constraints and regex patterns for CLI inputs.
type InputValidationPatterns struct {
	MaxArgCount       int
	MaxArgLen         int
	DisallowNull      bool
	DisallowEmpty     bool
	DisallowFlagClash bool
	CommandPattern    *regexp.Regexp
	SafeArgPattern    *regexp.Regexp
	SubcommandPattern *regexp.Regexp
}

// CanonicalValidationSchema represents the authoritative CLI input boundary schema.
var CanonicalValidationSchema = InputValidationPatterns{
	MaxArgCount:       MaxArgCount,
	MaxArgLen:         MaxArgLen,
	DisallowNull:      true,
	DisallowEmpty:     true,
	DisallowFlagClash: true,
	CommandPattern:    CommandPattern,
	SafeArgPattern:    SafeArgPattern,
	SubcommandPattern: SubcommandPattern,
}

// Validate applies schema validation rules and regex pattern enforcement to the supplied arguments.
func (p InputValidationPatterns) Validate(args []string) error {
	if len(args) == 0 || (p.DisallowEmpty && strings.TrimSpace(args[0]) == "") {
		return fmt.Errorf("%w: pip command is empty", ErrEmptyCommand)
	}
	if err := validateArgBounds(args); err != nil {
		return err
	}
	if p.CommandPattern != nil && !p.CommandPattern.MatchString(args[0]) {
		return fmt.Errorf("%w: command %q contains invalid characters", ErrInvalidArgs, args[0])
	}
	if p.DisallowFlagClash {
		if err := validateTranslationFlags(args); err != nil {
			return err
		}
	}
	return nil
}

// ValidateWithLogger applies canonical schema validation rules and emits structured diagnostics.
func (p InputValidationPatterns) ValidateWithLogger(args []string, log *slog.Logger) error {
	err := p.Validate(args)
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
	return CanonicalValidationSchema.Validate(args)
}

// ValidateArgsWithLogger validates CLI arguments against the canonical schema and records structured diagnostics with the given logger.
func ValidateArgsWithLogger(args []string, log *slog.Logger) error {
	return CanonicalValidationSchema.ValidateWithLogger(args, log)
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
