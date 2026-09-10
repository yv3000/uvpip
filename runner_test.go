package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// The test executable is an offline uv/Python fixture. No shell or network is used.
func TestMain(m *testing.M) {
	if os.Getenv("UVPIP_TEST_CHILD") == "1" {
		switch os.Getenv("UVPIP_TEST_MODE") {
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
	if code := runUV(args, strings.NewReader("stdin unchanged"), &stdout, &stderr); code != 0 {
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
	if code := runUV(args, nil, io.Discard, &stderr); code != 23 || stderr.String() != "child failure" {
		t.Fatalf("exit %d stderr %q", code, &stderr)
	}
}

func TestRunUVFailures(t *testing.T) {
	var stderr bytes.Buffer
	t.Setenv("UVPIP_UV", filepath.Join(t.TempDir(), "missing.exe"))
	if code := runUV(nil, nil, io.Discard, &stderr); code != 127 || !strings.Contains(stderr.String(), "install uv:") {
		t.Fatalf("code %d: %s", code, &stderr)
	}
	path := filepath.Join(t.TempDir(), "invalid.exe")
	if err := os.WriteFile(path, []byte("not an executable"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("UVPIP_UV", path)
	stderr.Reset()
	if code := runUV(nil, nil, io.Discard, &stderr); code != 126 || !strings.Contains(stderr.String(), "cannot execute uv") {
		t.Fatalf("code %d: %s", code, &stderr)
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
			if !reflect.DeepEqual(calls, tt.wantCalls) {
				t.Fatalf("calls %v", calls)
			}
		})
	}
}

func TestFindUV(t *testing.T) {
	self := fixtureUV(t)
	if got, err := findUV(); err != nil || got != self {
		t.Fatalf("%q %v", got, err)
	}
	t.Setenv("UVPIP_UV", "relative.exe")
	if _, err := findUV(); err == nil || !strings.Contains(err.Error(), "absolute") {
		t.Fatalf("%v", err)
	}
	t.Setenv("UVPIP_UV", t.TempDir())
	if _, err := findUV(); err == nil {
		t.Fatal("accepted directory as executable")
	}
	dir := t.TempDir()
	name := "uv"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	data, err := os.ReadFile(self)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("UVPIP_UV", "")
	t.Setenv("PATH", dir)
	if got, err := findUV(); err != nil || got != path {
		t.Fatalf("PATH lookup: %q %v", got, err)
	}
}

func TestCandidatePaths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", filepath.Join(home, "roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "local"))
	paths := uvCandidatePaths()
	if len(paths) < 2 || !strings.HasPrefix(paths[0], home) {
		t.Fatalf("paths %v", paths)
	}
	for _, p := range paths {
		if !filepath.IsAbs(p) {
			t.Fatalf("relative candidate %q", p)
		}
	}
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	t.Setenv("APPDATA", "")
	t.Setenv("LOCALAPPDATA", "")
	for _, p := range uvCandidatePaths() {
		if !filepath.IsAbs(p) {
			t.Fatalf("relative candidate without home %q", p)
		}
	}
	for _, key := range []string{"HOME", "USERPROFILE", "APPDATA", "LOCALAPPDATA"} {
		t.Setenv(key, ".")
	}
	for _, p := range uvCandidatePaths() {
		if !filepath.IsAbs(p) {
			t.Fatalf("relative home candidate %q", p)
		}
	}
}

func TestFindUVRejectsSelf(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("UVPIP_UV", self)
	if _, err := findUV(); err == nil || !strings.Contains(err.Error(), "recursive") {
		t.Fatalf("self accepted: %v", err)
	}
	if runtime.GOOS == "windows" {
		return // Windows locks hardlinks to the running executable against cleanup.
	}
	dir := t.TempDir()
	name := "uv"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	alias := filepath.Join(dir, name)
	if err := os.Link(self, alias); err != nil {
		t.Fatal(err)
	}
	t.Setenv("UVPIP_UV", "")
	t.Setenv("PATH", dir)
	if _, err := findUV(); err == nil || !strings.Contains(err.Error(), "recursive") {
		t.Fatalf("PATH alias accepted: %v", err)
	}
}

func TestBuildEnv(t *testing.T) {
	for _, tt := range []struct {
		name     string
		in, want []string
	}{
		{"system", []string{"PATH=a", "TOKEN=a=b"}, []string{"PATH=a", "TOKEN=a=b", "UV_SYSTEM_PYTHON=1"}},
		{"venv", []string{"VIRTUAL_ENV=somewhere"}, []string{"VIRTUAL_ENV=somewhere"}},
		{"conda", []string{"CONDA_PREFIX=somewhere"}, []string{"CONDA_PREFIX=somewhere"}},
		{"empty conda", []string{"CONDA_PREFIX="}, []string{"CONDA_PREFIX=", "UV_SYSTEM_PYTHON=1"}},
		{"empty venv", []string{"VIRTUAL_ENV="}, []string{"VIRTUAL_ENV=", "UV_SYSTEM_PYTHON=1"}},
		{"explicit", []string{"UV_SYSTEM_PYTHON=false", "PIP_PYTHON=python"}, []string{"UV_SYSTEM_PYTHON=false", "PIP_PYTHON=python"}},
		{"explicit with venv", []string{"UV_SYSTEM_PYTHON=1", "VIRTUAL_ENV=venv"}, []string{"UV_SYSTEM_PYTHON=1", "VIRTUAL_ENV=venv"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := buildEnv(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v want %v", got, tt.want)
			}
			got[0] = "mutated"
			if tt.in[0] == "mutated" {
				t.Fatal("input aliased")
			}
		})
	}
	want := []string{"virtual_env=venv", "UV_SYSTEM_PYTHON=1"}
	if runtime.GOOS == "windows" {
		want = want[:1]
	}
	if got := buildEnv([]string{"virtual_env=venv"}); !reflect.DeepEqual(got, want) {
		t.Fatalf("case rules: %v", got)
	}
}
