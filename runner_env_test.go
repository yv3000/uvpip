package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

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

func TestUvCandidatePaths(t *testing.T) {
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

	// Simulated GOOS=windows with mocked APPDATA/LOCALAPPDATA across all host OSes
	t.Run("simulated windows mock", func(t *testing.T) {
		roaming := filepath.Join(home, "AppData", "Roaming")
		local := filepath.Join(home, "AppData", "Local")
		mockEnv := func(k string) string {
			switch k {
			case "APPDATA":
				return roaming
			case "LOCALAPPDATA":
				return local
			default:
				return ""
			}
		}
		winPaths := candidatePathsFor("windows", mockEnv, home)
		if len(winPaths) < 4 {
			t.Fatalf("expected at least 4 windows candidate paths, got %d: %v", len(winPaths), winPaths)
		}
		foundRoaming, foundLocal := false, false
		for _, p := range winPaths {
			if !filepath.IsAbs(p) {
				t.Fatalf("windows candidate path not absolute: %s", p)
			}
			if !strings.HasSuffix(p, ".exe") {
				t.Fatalf("windows candidate path missing .exe suffix: %s", p)
			}
			if strings.Contains(p, "astral") {
				foundRoaming = true
			}
			if strings.Contains(p, "Programs") {
				foundLocal = true
			}
		}
		if !foundRoaming || !foundLocal {
			t.Fatalf("missing mocked roaming or local paths: roaming=%v local=%v in %v", foundRoaming, foundLocal, winPaths)
		}
	})

	// Simulated POSIX fallback check
	t.Run("simulated posix mock", func(t *testing.T) {
		posixPaths := candidatePathsFor("linux", func(string) string { return "" }, home)
		if len(posixPaths) < 2 {
			t.Fatalf("expected at least 2 home candidate paths, got %v", posixPaths)
		}
		if runtime.GOOS != "windows" {
			foundUsrLocal, foundNixOS, foundLinuxbrew := false, false, false
			for _, p := range posixPaths {
				switch p {
				case "/usr/local/bin/uv":
					foundUsrLocal = true
				case "/run/current-system/sw/bin/uv":
					foundNixOS = true
				case "/home/linuxbrew/.linuxbrew/bin/uv":
					foundLinuxbrew = true
				}
			}
			if !foundUsrLocal {
				t.Fatalf("missing /usr/local/bin/uv in posix paths: %v", posixPaths)
			}
			if !foundNixOS {
				t.Fatalf("missing NixOS system profile path in posix paths: %v", posixPaths)
			}
			if !foundLinuxbrew {
				t.Fatalf("missing Linuxbrew path in posix paths: %v", posixPaths)
			}
		}
	})
}

func TestFindUVRejectsSelf(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("UVPIP_UV", self)
	if _, err := findUV(); !errors.Is(err, ErrRecursiveExecution) {
		t.Fatalf("self accepted or wrong error: %v", err)
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
	if _, err := findUV(); !errors.Is(err, ErrRecursiveExecution) {
		t.Fatalf("PATH alias accepted or wrong error: %v", err)
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

func TestBuildEnvKeyCasing(t *testing.T) {
	// Assert uppercase normalization behavior on Windows vs POSIX preservation
	env := []string{"virtual_env=/tmp/myenv", "conda_prefix=/tmp/conda", "uv_system_python=1"}
	res := buildEnv(env)
	if runtime.GOOS == "windows" {
		// On Windows, case-insensitive comparison recognizes virtual_env and conda_prefix without adding duplicate UV_SYSTEM_PYTHON
		if len(res) != len(env) {
			t.Fatalf("expected windows buildEnv to recognize lowercase virtual_env, got %v", res)
		}
	} else {
		// On POSIX, case sensitivity means virtual_env is not VIRTUAL_ENV, so UV_SYSTEM_PYTHON is appended
		foundDefault := false
		for _, e := range res {
			if e == "UV_SYSTEM_PYTHON=1" {
				foundDefault = true
			}
		}
		if !foundDefault {
			t.Fatalf("expected POSIX buildEnv to append UV_SYSTEM_PYTHON=1 for lowercase virtual_env, got %v", res)
		}
	}
}
