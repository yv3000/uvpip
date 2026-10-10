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
		{"empty value after command", []string{"install", ""}, true},
		{"no args", nil, false},
		{"empty command", []string{""}, false},
		{"whitespace command", []string{" \t", "requests"}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArgs(tt.args)
			if (err == nil) != tt.ok {
				t.Fatalf("err = %v", err)
			}
			if err != nil && !errors.Is(err, errUsage) {
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
		name    string
		inputs  CLIInputs
		wantErr string
	}{
		{"valid args without override", CLIInputs{Args: []string{"install", "pkg"}}, ""},
		{"valid args with valid override", CLIInputs{Args: []string{"list"}, UVOverride: abs}, ""},
		{"empty args", CLIInputs{Args: nil}, "pip command is empty"},
		{"blank command", CLIInputs{Args: []string{"  "}}, "pip command is empty"},
		{"blank UVPIP_UV", CLIInputs{Args: []string{"list"}, UVOverride: "  "}, "blank"},
		{"whitespace-padded UVPIP_UV", CLIInputs{Args: []string{"list"}, UVOverride: abs + " "}, "whitespace"},
		{"malformed relative UVPIP_UV", CLIInputs{Args: []string{"list"}, UVOverride: "uv"}, "absolute"},
		{"valid override only", CLIInputs{UVOverride: abs}, ""},
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
		})
	}
}
