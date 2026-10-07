package main

import (
	"testing"
)

func FuzzTranslateToUV(f *testing.F) {
	f.Add("install", "requests", "-y")
	f.Add("uninstall", "pkg", "--yes")
	f.Add("upgrade", "mypkg", "")
	f.Add("cache", "dir", "")
	f.Add("", "", "")
	f.Fuzz(func(t *testing.T, cmd, arg1, arg2 string) {
		args := []string{cmd, arg1, arg2}
		out := translateToUV(args)
		if len(out) == 0 {
			t.Fatal("expected non-empty translated arguments")
		}
	})
}

func FuzzValidateUVOverride(f *testing.F) {
	f.Add("/usr/bin/uv")
	f.Add("C:\\uv\\uv.exe")
	f.Add("  /usr/bin/uv  ")
	f.Add("relative/path")
	f.Add("")
	f.Fuzz(func(t *testing.T, path string) {
		_ = validateUVOverride(path)
	})
}

func FuzzBuildEnv(f *testing.F) {
	f.Add("PATH=/bin", "VIRTUAL_ENV=/venv", "UV_SYSTEM_PYTHON=1")
	f.Add("CONDA_PREFIX=/conda", "FOO=BAR", "")
	f.Add("", "", "")
	f.Fuzz(func(t *testing.T, e1, e2, e3 string) {
		env := []string{e1, e2, e3}
		out := buildEnv(env)
		if len(out) < len(env) {
			t.Fatal("buildEnv dropped environment entries")
		}
	})
}
