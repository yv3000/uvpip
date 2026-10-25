package main

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"testing"
)

// TestValidateBoundariesSubcommands asserts every CLI subcommand path enforces
// ValidateInputs boundaries before delegating to the process runner or executing uv.
// Tested subcommands: install, uninstall, freeze, list, doctor.
func TestValidateBoundariesSubcommands(t *testing.T) {
	fixtureUV(t)

	subcommands := []string{"install", "uninstall", "freeze", "list", "doctor"}

	for _, subcmd := range subcommands {
		t.Run(subcmd+"/valid", func(t *testing.T) {
			args := []string{subcmd}
			if subcmd == "install" {
				args = append(args, "requests")
			}
			err := ValidateInputs(CLIInputs{Args: args})
			if err != nil {
				t.Fatalf("expected valid inputs for %s, got: %v", subcmd, err)
			}
		})

		t.Run(subcmd+"/null_byte_injection", func(t *testing.T) {
			var stderr bytes.Buffer
			args := []string{subcmd, "pkg\x00malicious"}
			code := runCLI(args, nil, io.Discard, &stderr)
			if code != 2 {
				t.Fatalf("%s: expected exit code 2 for null byte injection, got %d (stderr: %s)", subcmd, code, stderr.String())
			}
			if !strings.Contains(stderr.String(), "null byte") {
				t.Fatalf("%s: expected null byte rejection in stderr, got: %s", subcmd, stderr.String())
			}
		})

		t.Run(subcmd+"/oversized_argument_length", func(t *testing.T) {
			var stderr bytes.Buffer
			oversized := strings.Repeat("a", 4097)
			args := []string{subcmd, oversized}
			code := runCLI(args, nil, io.Discard, &stderr)
			if code != 2 {
				t.Fatalf("%s: expected exit code 2 for oversized argument, got %d", subcmd, code)
			}
			if !strings.Contains(stderr.String(), "exceeds maximum limit") {
				t.Fatalf("%s: expected length bound error, got: %s", subcmd, stderr.String())
			}
		})
	}
}

// TestValidateBoundariesDoctor asserts doctor input validation handles environment override boundaries.
func TestValidateBoundariesDoctor(t *testing.T) {
	t.Run("valid_doctor_with_valid_override", func(t *testing.T) {
		validPath := filepath.Join(t.TempDir(), "uv.exe")
		t.Setenv("UVPIP_UV", validPath)
		err := ValidateInputs(CLIInputs{Args: []string{"doctor"}, UVOverride: validPath})
		if err != nil {
			t.Fatalf("unexpected validation error: %v", err)
		}
	})

	t.Run("doctor_with_invalid_relative_override", func(t *testing.T) {
		t.Setenv("UVPIP_UV", "relative_uv.exe")
		var stdout, stderr bytes.Buffer
		code := runCLI([]string{"doctor"}, nil, &stdout, &stderr)
		if code != 2 {
			t.Fatalf("expected exit code 2 for invalid override in runCLI, got %d", code)
		}
	})

	t.Run("runDoctor_with_invalid_override", func(t *testing.T) {
		t.Setenv("UVPIP_UV", "relative_uv.exe")
		var out bytes.Buffer
		code := runDoctor(&out, quiet)
		if code != 1 {
			t.Fatalf("expected exit code 1 for invalid override in runDoctor, got %d", code)
		}
		if !strings.Contains(out.String(), "uv override invalid") {
			t.Fatalf("expected override failure in output, got: %s", out.String())
		}
	})
}

// TestValidateBoundariesRunner asserts runUV enforces ValidateInputs before executing child.
func TestValidateBoundariesRunner(t *testing.T) {
	t.Run("runner_rejects_null_byte", func(t *testing.T) {
		var stderr bytes.Buffer
		code := runUV([]string{"pip", "install", "bad\x00pkg"}, nil, io.Discard, &stderr, quiet)
		if code != 2 {
			t.Fatalf("expected exit 2 from runner, got %d", code)
		}
	})

	t.Run("runner_rejects_invalid_override", func(t *testing.T) {
		t.Setenv("UVPIP_UV", "relative.exe")
		var stderr bytes.Buffer
		code := runUV([]string{"pip", "install", "pkg"}, nil, io.Discard, &stderr, quiet)
		if code != 2 {
			t.Fatalf("expected exit 2 from runner for relative override, got %d", code)
		}
	})
}
