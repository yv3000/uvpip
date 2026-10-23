package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// quiet is the default (UVPIP_DEBUG unset) logger used by direct runUV tests.
var quiet = slog.New(slog.DiscardHandler)

// The test executable is an offline uv/Python fixture. No shell or network is used.
func TestMain(m *testing.M) {
	if os.Getenv("UVPIP_TEST_CHILD") == "1" {
		switch os.Getenv("UVPIP_TEST_MODE") {
		case "wait-for-cancel":
			runCancellationFixture()
		case "inherit-output", "inherit-output-failure", "hold-output":
			runInheritedPipeFixture()
		case "exit":
			fmt.Fprint(os.Stderr, "child failure")
			os.Exit(23)
		case "version":
			fmt.Println("uv offline-fixture")
		default:
			input, _ := io.ReadAll(os.Stdin)
			_ = json.NewEncoder(os.Stdout).Encode(struct {
				Args                           []string
				Input, System, Venv, PipPython string
			}{os.Args[1:], string(input), os.Getenv("UV_SYSTEM_PYTHON"), os.Getenv("VIRTUAL_ENV"), os.Getenv("PIP_PYTHON")})
			fmt.Fprint(os.Stderr, "child stderr")
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func fixtureUV(t *testing.T) string {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(self)
	if err != nil {
		t.Fatal(err)
	}
	self = filepath.Join(t.TempDir(), "uv-fixture.exe")
	if err := os.WriteFile(self, data, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("UVPIP_UV", self)
	t.Setenv("UVPIP_TEST_CHILD", "1")
	t.Setenv("UVPIP_TEST_MODE", "echo")
	return self
}

func TestRunUV(t *testing.T) {
	fixtureUV(t)
	t.Setenv("VIRTUAL_ENV", "venv with spaces")
	t.Setenv("UV_SYSTEM_PYTHON", "false")
	t.Setenv("PIP_PYTHON", "user value")
	args := []string{"pip", "install", "a b", "a;b", "$(do-not-run)", "x\"y", ""}
	var stdout, stderr bytes.Buffer
	if code := runUV(args, strings.NewReader("stdin unchanged"), &stdout, &stderr, quiet); code != 0 {
		t.Fatalf("exit %d: %s", code, &stderr)
	}
	var got struct {
		Args                           []string
		Input, System, Venv, PipPython string
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Args, args) || got.Input != "stdin unchanged" || got.System != "false" || got.Venv != "venv with spaces" || got.PipPython != "user value" {
		t.Fatalf("subprocess changed arguments/streams/environment: %+v", got)
	}
	if stderr.String() != "child stderr" {
		t.Fatalf("stderr = %q", &stderr)
	}
	t.Setenv("UVPIP_TEST_MODE", "exit")
	stderr.Reset()
	if code := runUV(args, nil, io.Discard, &stderr, quiet); code != 23 || stderr.String() != "child failure" {
		t.Fatalf("exit %d stderr %q", code, &stderr)
	}
}

func TestRunUVFailures(t *testing.T) {
	var stderr bytes.Buffer
	t.Setenv("UVPIP_UV", filepath.Join(t.TempDir(), "missing.exe"))
	if code := runUV(nil, nil, io.Discard, &stderr, quiet); code != 127 || !strings.Contains(stderr.String(), "install uv:") {
		t.Fatalf("code %d: %s", code, &stderr)
	}
	path := filepath.Join(t.TempDir(), "invalid.exe")
	if err := os.WriteFile(path, []byte("not an executable"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("UVPIP_UV", path)
	stderr.Reset()
	if code := runUV(nil, nil, io.Discard, &stderr, quiet); code != 126 || !strings.Contains(stderr.String(), "cannot execute uv") {
		t.Fatalf("code %d: %s", code, &stderr)
	}
}

func TestExitCode(t *testing.T) {
	if code, err := exitCode(nil); code != 0 || err != nil {
		t.Fatalf("success: %d %v", code, err)
	}
	startErr := fmt.Errorf("start: %w", exec.ErrNotFound)
	if code, err := exitCode(startErr); code != 126 || !errors.Is(err, exec.ErrNotFound) {
		t.Fatalf("start failure: %d %v", code, err)
	}
	fixtureUV(t)
	t.Setenv("UVPIP_TEST_MODE", "exit")
	path := os.Getenv("UVPIP_UV")
	if code, err := exitCode(exec.Command(path).Run()); code != 23 || err != nil {
		t.Fatalf("child exit: %d %v", code, err)
	}
}

func TestLookupUV(t *testing.T) {
	for _, tt := range []struct {
		name, match string
		dot         bool
		wantCalls   []string
	}{
		{"PATH first", "uv", false, []string{"uv"}},
		{"fallback", "second", false, []string{"uv", "first", "second"}},
		{"absent", "", false, []string{"uv", "first", "second"}},
		{"unsafe PATH", "", true, []string{"uv"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			path, err := lookupUV(func(name string) (string, error) {
				calls = append(calls, name)
				if tt.dot {
					return "uv", exec.ErrDot
				}
				if name == tt.match {
					return "resolved", nil
				}
				return "", exec.ErrNotFound
			}, []string{"first", "second"})
			if (err == nil) != (tt.match != "") || (err == nil && path != "resolved") {
				t.Fatalf("path %q, err %v", path, err)
			}
			if tt.dot && !errors.Is(err, exec.ErrDot) {
				t.Fatalf("lost ErrDot: %v", err)
			}
			if tt.match == "" && !tt.dot && !errors.Is(err, ErrUVNotFound) {
				t.Fatalf("expected ErrUVNotFound: %v", err)
			}
			if !reflect.DeepEqual(calls, tt.wantCalls) {
				t.Fatalf("calls %v", calls)
			}
		})
	}
}

func TestRunUVInheritedOutputPipes(t *testing.T) {
	testRunUVInheritedOutputPipes(t)
}

// TestOfflineDryRunErrorCodeParity asserts exact error code preservation when running
// in offline / disconnected environments with dry-run flags and invalid invocations.
func TestOfflineDryRunErrorCodeParity(t *testing.T) {
	fixtureUV(t)
	// Mode 'exit' in fixture causes the mock uv child to output "child failure" and exit 23
	t.Setenv("UVPIP_TEST_MODE", "exit")

	var stdout, stderr bytes.Buffer
	// Test simulated offline failed dry-run: pip install --dry-run nonexistent-package
	args := []string{"install", "--dry-run", "nonexistent-package-offline"}
	code := runCLI(args, nil, &stdout, &stderr)

	if code != 23 {
		t.Fatalf("expected exit code 23 faithfully passed through from uv, got %d", code)
	}
	if !strings.Contains(stderr.String(), "child failure") {
		t.Fatalf("expected child failure output passed through to stderr, got %q", stderr.String())
	}

	// Test CLI invalid blank command returning 2 (pip usage error parity)
	var errBuf bytes.Buffer
	blankCode := runCLI([]string{""}, nil, io.Discard, &errBuf)
	if blankCode != 2 {
		t.Fatalf("expected exit code 2 for invalid blank command, got %d", blankCode)
	}

	// Test missing uv returning 127
	t.Setenv("UVPIP_UV", filepath.Join(t.TempDir(), "nonexistent-uv-binary.exe"))
	var missingBuf bytes.Buffer
	missingCode := runCLI([]string{"list"}, nil, io.Discard, &missingBuf)
	if missingCode != 127 {
		t.Fatalf("expected exit code 127 when uv cannot be found, got %d", missingCode)
	}
}
