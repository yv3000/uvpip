package main

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"testing"
)

func TestDebugEnabled(t *testing.T) {
	for _, tt := range []struct {
		value string
		want  bool
	}{
		{"1", true}, {"true", true}, {" TRUE ", true}, {"yes", true}, {"on", true},
		{"", false}, {"0", false}, {"false", false}, {"off", false}, {"debug", false},
	} {
		if got := debugEnabled(tt.value); got != tt.want {
			t.Errorf("debugEnabled(%q) = %v, want %v", tt.value, got, tt.want)
		}
	}
}

func TestNewLogger(t *testing.T) {
	var out bytes.Buffer
	newLogger(&out, "").Debug("hidden")
	if out.Len() != 0 {
		t.Fatalf("disabled logger wrote %q", &out)
	}
	newLogger(&out, "1").Debug("shown", "key", "value")
	if got := out.String(); !strings.Contains(got, "level=DEBUG") || !strings.Contains(got, "component=uvpip") ||
		!strings.Contains(got, "msg=shown") || !strings.Contains(got, "key=value") {
		t.Fatalf("structured record missing fields: %q", got)
	}
}

func TestCLIDebugDiagnostics(t *testing.T) {
	fixtureUV(t)
	secret := "https://user:s3cret-token@example.invalid/simple"
	args := []string{"install", "--index-url", secret, "pkg"}

	// Default: uv's stderr passes through byte-for-byte.
	var stderr bytes.Buffer
	if code := runCLI(args, nil, io.Discard, &stderr); code != 0 || stderr.String() != "child stderr" {
		t.Fatalf("default stderr changed: %d %q", code, &stderr)
	}

	t.Setenv("UVPIP_DEBUG", "1")
	stderr.Reset()
	if code := runCLI(args, nil, io.Discard, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, &stderr)
	}
	got := stderr.String()
	for _, want := range []string{
		`msg="translated command"`, "uv_command=pip arg_count=5",
		`msg="resolved uv"`, "source=UVPIP_UV",
		`msg="prepared environment"`,
		`msg="uv exited"`, "code=0",
		"child stderr",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("debug output missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "s3cret-token") {
		t.Fatalf("debug output leaked argument values:\n%s", got)
	}

	t.Setenv("UVPIP_UV", filepath.Join(t.TempDir(), "missing"))
	stderr.Reset()
	if code := runCLI(args, nil, io.Discard, &stderr); code != 127 || !strings.Contains(stderr.String(), `msg="uv discovery failed"`) {
		t.Fatalf("discovery failure not logged: %d %s", code, &stderr)
	}
}

func TestFailureLoggingWithoutDebug(t *testing.T) {
	t.Setenv("UVPIP_DEBUG", "")
	t.Setenv("UVPIP_UV", filepath.Join(t.TempDir(), "missing"))
	var stderr bytes.Buffer
	code := runCLI([]string{"list"}, nil, io.Discard, &stderr)
	if code != 127 {
		t.Fatalf("exit code %d, want 127", code)
	}
	out := stderr.String()
	if !strings.Contains(out, "level=INFO") || !strings.Contains(out, `msg="uv discovery failed"`) {
		t.Fatalf("discovery failure not logged at info level without UVPIP_DEBUG: %d %s", code, &stderr)
	}
}
