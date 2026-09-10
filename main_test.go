package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIHelpAndVersion(t *testing.T) {
	t.Setenv("UVPIP_UV", filepath.Join(t.TempDir(), "missing"))
	for _, args := range [][]string{nil, {"--help"}, {"-h"}, {"--version"}, {"-V"}} {
		var stdout, stderr bytes.Buffer
		if code := runCLI(args, nil, &stdout, &stderr); code != 0 || stderr.Len() != 0 {
			t.Fatalf("args %v code %d stderr %s", args, code, &stderr)
		}
		if !strings.Contains(stdout.String(), version) {
			t.Fatalf("output %q", &stdout)
		}
		if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
			if !strings.Contains(stdout.String(), "Usage:") {
				t.Fatal("missing usage")
			}
		}
	}
}

func TestCLIForwardsPipAndPip3(t *testing.T) {
	fixtureUV(t)
	before := os.Args
	t.Cleanup(func() { os.Args = before })
	for _, name := range []string{"pip", "pip3", "uvpip"} {
		os.Args = []string{name, "install", "a b"}
		var stdout bytes.Buffer
		if code := runCLI(os.Args[1:], nil, &stdout, io.Discard); code != 0 {
			t.Fatalf("exit %d", code)
		}
		var result struct{ Args []string }
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if strings.Join(result.Args, "|") != "pip|install|a b" {
			t.Fatalf("%s: %v", name, result.Args)
		}
	}
	var stdout bytes.Buffer
	if code := runCLI([]string{"-v", "install", "pkg"}, nil, &stdout, io.Discard); code != 0 || !strings.Contains(stdout.String(), `"-v"`) {
		t.Fatalf("verbose flag swallowed: %d %s", code, &stdout)
	}
	t.Setenv("UVPIP_TEST_MODE", "exit")
	if code := runCLI([]string{"list"}, nil, io.Discard, io.Discard); code != 23 {
		t.Fatalf("child exit = %d", code)
	}
}
