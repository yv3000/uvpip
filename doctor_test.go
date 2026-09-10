package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestIsUvpipShim(t *testing.T) {
	for _, name := range []string{"pip.sh", "pip3.sh", "pip.cmd", "pip3.cmd"} {
		if !isUvpipShim(filepath.Join("bin", name)) {
			t.Errorf("shipped shim not recognized: %s", name)
		}
	}
	path := filepath.Join(t.TempDir(), "shim")
	if isUvpipShim(path) || isUvpipShim(filepath.Dir(path)) {
		t.Fatal("accepted missing file/directory")
	}
	for _, content := range []string{"mentions uvpip only", "\x00uvpip", strings.Repeat("uvpip", 1000)} {
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		if isUvpipShim(path) {
			t.Fatalf("accepted non-shim: %.20q", content)
		}
	}
}

func TestDoctor(t *testing.T) {
	self := fixtureUV(t)
	t.Setenv("UVPIP_TEST_MODE", "version")
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	python := "python3"
	shimSuffix := ""
	if runtime.GOOS == "windows" {
		python += ".exe"
		shimSuffix = ".cmd"
	}
	data, err := os.ReadFile(self)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, python), data, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"pip", "pip3"} {
		content := "#!/bin/sh\nexec \"$(dirname \"$0\")/uvpip\" \"$@\"\n"
		if runtime.GOOS == "windows" {
			content = "@echo off\r\n\"%~dp0uvpip.exe\" %*\r\n"
		}
		if err := os.WriteFile(filepath.Join(dir, name+shimSuffix), []byte(content), 0700); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	if code := runCLI([]string{"doctor"}, nil, &out, &out); code != 0 || strings.Contains(out.String(), "[!!]") {
		t.Fatalf("code %d: %s", code, &out)
	}
	for _, name := range []string{"pip", "pip3"} {
		if err := os.WriteFile(filepath.Join(dir, name+shimSuffix), []byte("real pip"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	out.Reset()
	if code := runDoctor(&out); code != 1 || strings.Count(out.String(), "not a recognized") != 2 {
		t.Fatalf("code %d: %s", code, &out)
	}
	t.Setenv("UVPIP_TEST_MODE", "exit")
	out.Reset()
	if code := runDoctor(&out); code != 1 || strings.Count(out.String(), "[!!]") != 4 {
		t.Fatalf("failed executable accepted %d: %s", code, &out)
	}
	t.Setenv("PATH", t.TempDir())
	t.Setenv("UVPIP_UV", filepath.Join(dir, "missing"))
	out.Reset()
	if code := runDoctor(&out); code != 1 || strings.Count(out.String(), "[!!]") != 4 {
		t.Fatalf("missing dependencies accepted %d: %s", code, &out)
	}
}
