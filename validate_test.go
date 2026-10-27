package main

import (
	"bytes"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateArgs(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		ok   bool
	}{
		{"command", []string{"install", "requests"}, true},
		{"verbosity conflict", []string{"install", "-q", "--verbose"}, false},
		{"literal verbosity operand", []string{"install", "-q", "--", "--verbose"}, true},
		{"empty value after command", []string{"install", ""}, true},
		{"no args", nil, false},
		{"empty command", []string{""}, false},
		{"whitespace command", []string{" \t", "requests"}, false},
		{"512 args", append([]string{"install"}, make([]string, 511)...), true},
		{"513 args", append([]string{"install"}, make([]string, 512)...), false},
		{"4096 bytes", []string{"install", strings.Repeat("é", 2048)}, true},
		{"4097 bytes", []string{"install", strings.Repeat("é", 2048) + "a"}, false},
		{"null byte", []string{"install", "private\x00value"}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateArgs(tt.args)
			if (err == nil) != tt.ok {
				t.Fatalf("err = %v", err)
			}
			if err != nil && (!errors.Is(err, ErrUsage) || !errors.Is(err, ErrInvalidArgs)) {
				t.Fatalf("not a usage error: %v", err)
			}
		})
	}
}

func TestValidateUVOverride(t *testing.T) {
	abs, err := filepath.Abs("uv")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name, path, want string
	}{
		{"absolute", abs, ""},
		{"blank", "  ", "blank"},
		{"trailing space", abs + " ", "whitespace"},
		{"leading tab", "\t" + abs, "whitespace"},
		{"relative", "uv", "absolute"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUVOverride(tt.path)
			if tt.want == "" {
				if err != nil {
					t.Fatalf("rejected valid path: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %v, want mention of %q", err, tt.want)
			}
		})
	}
}

func TestCLIRejectsBlankCommand(t *testing.T) {
	fixtureUV(t)
	var stdout, stderr bytes.Buffer
	if code := runCLI([]string{" ", "secret-token"}, nil, &stdout, &stderr); code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("uv ran: %s", &stdout)
	}
	if !strings.Contains(stderr.String(), "pip command is empty") || strings.Contains(stderr.String(), "secret-token") {
		t.Fatalf("stderr = %q", &stderr)
	}
}

func TestRunUVRejectsBlankOverride(t *testing.T) {
	t.Setenv("UVPIP_UV", "   ")
	var stderr bytes.Buffer
	if code := runUV([]string{"pip"}, nil, io.Discard, &stderr, quiet); code != 127 || !strings.Contains(stderr.String(), "blank") {
		t.Fatalf("code %d: %s", code, &stderr)
	}
}

func TestValidateInputs(t *testing.T) {
	abs, err := filepath.Abs("uv")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name      string
		inputs    CLIInputs
		wantTyped error
		wantErr   string
	}{
		{"valid args without override", CLIInputs{Args: []string{"install", "pkg"}}, nil, ""},
		{"valid args with valid override", CLIInputs{Args: []string{"list"}, UVOverride: abs}, nil, ""},
		{"empty args", CLIInputs{Args: nil}, ErrInvalidArgs, "pip command is empty"},
		{"blank command", CLIInputs{Args: []string{"  "}}, ErrInvalidArgs, "pip command is empty"},
		{"blank UVPIP_UV", CLIInputs{Args: []string{"list"}, UVOverride: "  "}, ErrInvalidOverride, "blank"},
		{"whitespace-padded UVPIP_UV", CLIInputs{Args: []string{"list"}, UVOverride: abs + " "}, ErrInvalidOverride, "whitespace"},
		{"malformed relative UVPIP_UV", CLIInputs{Args: []string{"list"}, UVOverride: "uv"}, ErrInvalidOverride, "absolute"},
		{"valid override only", CLIInputs{UVOverride: abs}, nil, ""},
		{"null byte in args", CLIInputs{Args: []string{"install", "pkg\x00name"}}, ErrInvalidArgs, "forbidden null byte"},
		{"oversized single argument", CLIInputs{Args: []string{"install", strings.Repeat("a", 4097)}}, ErrInvalidArgs, "exceeds maximum limit"},
		{"oversized argument count", CLIInputs{Args: append([]string{"install"}, make([]string, 513)...)}, ErrInvalidArgs, "exceeds maximum limit"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInputs(tt.inputs)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("err = %v, want mention of %q", err, tt.wantErr)
			}
			if tt.wantTyped != nil && !errors.Is(err, tt.wantTyped) {
				t.Fatalf("err %v does not wrap typed error %v", err, tt.wantTyped)
			}
		})
	}
}

func TestValidateArgsWithLoggerAndSchema(t *testing.T) {
	var logBuf bytes.Buffer
	log := newLogger(&logBuf, "1")

	// Valid arguments
	err := ValidateArgsWithLogger([]string{"install", "requests"}, log)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if !strings.Contains(logBuf.String(), "argument validation passed") || !strings.Contains(logBuf.String(), "component=uvpip") {
		t.Errorf("expected passing structured log with component=uvpip, got %s", logBuf.String())
	}

	// Schema Validate method
	if err := CanonicalValidationSchema.Validate([]string{"install", "requests"}); err != nil {
		t.Fatalf("CanonicalValidationSchema.Validate rejected valid args: %v", err)
	}

	// Specific sentinel error unwrapping
	logBuf.Reset()
	errEmpty := ValidateArgsWithLogger([]string{""}, log)
	if !errors.Is(errEmpty, ErrEmptyCommand) || !errors.Is(errEmpty, ErrUsage) {
		t.Errorf("expected ErrEmptyCommand wrapping, got %v", errEmpty)
	}
	if !strings.Contains(logBuf.String(), "argument validation failed") || !strings.Contains(logBuf.String(), "component=uvpip") {
		t.Errorf("expected failure structured log with component=uvpip, got %s", logBuf.String())
	}

	errBounds := ValidateArgs(append([]string{"install"}, make([]string, 513)...))
	if !errors.Is(errBounds, ErrArgBounds) {
		t.Errorf("expected ErrArgBounds wrapping, got %v", errBounds)
	}

	errNull := ValidateArgs([]string{"install", "pkg\x00name"})
	if !errors.Is(errNull, ErrNullByte) {
		t.Errorf("expected ErrNullByte wrapping, got %v", errNull)
	}

	errConflict := ValidateArgs([]string{"install", "-q", "-v"})
	if !errors.Is(errConflict, ErrFlagConflict) {
		t.Errorf("expected ErrFlagConflict wrapping, got %v", errConflict)
	}
}

func TestInputValidationPatternsAndRegex(t *testing.T) {
	if CommandPattern == nil || SafeArgPattern == nil || SubcommandPattern == nil {
		t.Fatal("expected non-nil regex validation patterns")
	}

	for _, tt := range []struct {
		name    string
		command string
		want    bool
	}{
		{"alphanumeric", "install", true},
		{"hyphenated", "download-extra", true},
		{"flag", "--upgrade", true},
		{"underscore", "show_all", true},
		{"semicolon injection", "install;rm", false},
		{"pipe injection", "install|sh", false},
		{"dollar expansion", "install$VAR", false},
		{"backtick injection", "install`whoami`", false},
	} {
		t.Run("command_pattern_"+tt.name, func(t *testing.T) {
			got := CommandPattern.MatchString(tt.command)
			if got != tt.want {
				t.Errorf("CommandPattern.MatchString(%q) = %v; want %v", tt.command, got, tt.want)
			}
		})
	}

	for _, tt := range []struct {
		name string
		arg  string
		want bool
	}{
		{"standard text", "requests==2.31.0", true},
		{"path with slashes", "/tmp/requirements.txt", true},
		{"null byte embedded", "pkg\x00extra", false},
		{"null byte prefix", "\x00test", false},
	} {
		t.Run("safe_arg_pattern_"+tt.name, func(t *testing.T) {
			got := SafeArgPattern.MatchString(tt.arg)
			if got != tt.want {
				t.Errorf("SafeArgPattern.MatchString(%q) = %v; want %v", tt.arg, got, tt.want)
			}
		})
	}

	// Assert schema rejects malicious command token
	badCmdErr := CanonicalValidationSchema.Validate([]string{"install;cat", "requests"})
	if badCmdErr == nil || !errors.Is(badCmdErr, ErrInvalidArgs) {
		t.Errorf("expected ErrInvalidArgs on invalid command token, got %v", badCmdErr)
	}

	// Assert input_validation_patterns slice exports valid regex patterns
	if len(input_validation_patterns) != 3 || len(CanonicalValidationPatterns) != 3 {
		t.Fatalf("expected 3 validation patterns, got %d", len(input_validation_patterns))
	}
	for _, p := range input_validation_patterns {
		if p == "" {
			t.Error("expected non-empty regex pattern")
		}
	}
}
