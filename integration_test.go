package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestInstallDoctorUninstallLifecycle exercises the full installation lifecycle:
// pre-install failure, simulated installation, doctor and CLI execution pass,
// and uninstallation resulting in clean diagnostic failure.
func TestInstallDoctorUninstallLifecycle(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("UVPIP_TEST_CHILD", "1")
	t.Setenv("UVPIP_TEST_MODE", "version")

	// Prepare mock python in an external tool directory
	toolsDir := filepath.Join(home, "tools")
	if err := os.MkdirAll(toolsDir, 0700); err != nil {
		t.Fatal(err)
	}

	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	selfBytes, err := os.ReadFile(self)
	if err != nil {
		t.Fatal(err)
	}

	exeExt := ""
	shimExt := ""
	if runtime.GOOS == "windows" {
		exeExt = ".exe"
		shimExt = ".cmd"
	}

	pythonPath := filepath.Join(toolsDir, "python"+exeExt)
	if err := os.WriteFile(pythonPath, selfBytes, 0700); err != nil {
		t.Fatal(err)
	}

	// Prepare mock uv binary
	uvPath := filepath.Join(toolsDir, "uv"+exeExt)
	if err := os.WriteFile(uvPath, selfBytes, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("UVPIP_UV", uvPath)
	t.Setenv("PATH", toolsDir)

	// Stage 1: Pre-install state.
	// Neither pip nor pip3 shim exists in PATH, so doctor must fail.
	var preDocOut bytes.Buffer
	if code := runDoctor(&preDocOut); code != 1 || !strings.Contains(preDocOut.String(), "[!!]") {
		t.Fatalf("expected pre-install doctor failure, got code %d output:\n%s", code, preDocOut.String())
	}

	// Stage 2: Simulate installer layout in $HOME/.uvpip/bin.
	installBin := filepath.Join(home, ".uvpip", "bin")
	if err := os.MkdirAll(installBin, 0700); err != nil {
		t.Fatal(err)
	}

	installedUvpip := filepath.Join(installBin, "uvpip"+exeExt)
	if err := os.WriteFile(installedUvpip, selfBytes, 0700); err != nil {
		t.Fatal(err)
	}

	shimContent := "#!/bin/sh\nexec \"$(dirname \"$0\")/uvpip\" \"$@\"\n"
	if runtime.GOOS == "windows" {
		shimContent = "@echo off\r\n\"%~dp0uvpip.exe\" %*\r\n"
	}

	for _, name := range []string{"pip", "pip3"} {
		shimPath := filepath.Join(installBin, name+shimExt)
		if err := os.WriteFile(shimPath, []byte(shimContent), 0700); err != nil {
			t.Fatal(err)
		}
	}

	// Prepend installed bin directory to PATH
	t.Setenv("PATH", installBin+string(os.PathListSeparator)+toolsDir)

	// Stage 3: Doctor check in installed state.
	var postInstallDoc bytes.Buffer
	if code := runCLI([]string{"doctor"}, nil, &postInstallDoc, &postInstallDoc); code != 0 {
		t.Fatalf("doctor failed in installed state with code %d:\n%s", code, postInstallDoc.String())
	}
	for _, check := range []string{"[OK] uv:", "[OK] pip:", "[OK] pip3:", "[OK] python:"} {
		if !strings.Contains(postInstallDoc.String(), check) {
			t.Fatalf("doctor output missing %q:\n%s", check, postInstallDoc.String())
		}
	}

	// Stage 4: Run CLI commands through the installed wrapper.
	t.Setenv("UVPIP_TEST_MODE", "echo")
	var stdout, stderr bytes.Buffer
	if code := runCLI([]string{"install", "requests"}, nil, &stdout, &stderr); code != 0 {
		t.Fatalf("CLI run failed with exit %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "install") || !strings.Contains(stdout.String(), "requests") {
		t.Fatalf("unexpected forwarded args in stdout: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := runCLI([]string{"uninstall", "-y", "pkg"}, nil, &stdout, &stderr); code != 0 {
		t.Fatalf("CLI uninstall failed with exit %d: %s", code, stderr.String())
	}
	if strings.Contains(stdout.String(), "-y") {
		t.Fatalf("expected -y flag stripped by translator: %s", stdout.String())
	}

	// Stage 5: Simulate uninstaller cleanup.
	for _, name := range []string{"pip", "pip3"} {
		if err := os.Remove(filepath.Join(installBin, name+shimExt)); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Remove(installedUvpip); err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(installBin)
	_ = os.Remove(filepath.Join(home, ".uvpip"))

	// Stage 6: Doctor check after uninstallation.
	t.Setenv("UVPIP_TEST_MODE", "version")
	var postUninstallDoc bytes.Buffer
	if code := runDoctor(&postUninstallDoc); code != 1 || !strings.Contains(postUninstallDoc.String(), "[!!]") {
		t.Fatalf("expected doctor failure post-uninstall, got code %d output:\n%s", code, postUninstallDoc.String())
	}
}
