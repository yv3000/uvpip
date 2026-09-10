package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// runUV never invokes a shell or downloads code. Streams pass through unchanged.
func runUV(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	uvPath, err := findUV()
	if err != nil {
		fmt.Fprintf(stderr, "[uvpip] %v; install uv: https://docs.astral.sh/uv/getting-started/installation/\n", err)
		return 127
	}
	cmd := exec.Command(uvPath, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdin, stdout, stderr
	cmd.Env = buildEnv(os.Environ())
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if code := exitErr.ExitCode(); code >= 0 {
				return code
			}
			return 1 // A signal has no portable process exit code.
		}
		fmt.Fprintf(stderr, "[uvpip] cannot execute uv at %q: %v\n", uvPath, err)
		return 126
	}
	return 0
}

func findUV() (string, error) {
	path := os.Getenv("UVPIP_UV")
	var err error
	if path != "" {
		if !filepath.IsAbs(path) {
			return "", fmt.Errorf("UVPIP_UV must be an absolute executable path")
		}
		path, err = exec.LookPath(path)
		if err != nil {
			return "", fmt.Errorf("UVPIP_UV: %w", err)
		}
	} else {
		path, err = lookupUV(exec.LookPath, uvCandidatePaths())
		if err != nil {
			return "", err
		}
	}
	self, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate uvpip executable: %w", err)
	}
	selfInfo, err := os.Stat(self)
	if err != nil {
		return "", fmt.Errorf("inspect uvpip executable: %w", err)
	}
	uvInfo, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("inspect uv executable: %w", err)
	}
	if os.SameFile(selfInfo, uvInfo) {
		return "", fmt.Errorf("uv resolves to uvpip itself; refusing recursive execution")
	}
	return path, nil
}

func lookupUV(lookPath func(string) (string, error), candidates []string) (string, error) {
	if path, err := lookPath("uv"); err == nil {
		return path, nil
	} else if errors.Is(err, exec.ErrDot) {
		return "", fmt.Errorf("refusing uv from the current directory: %w", err)
	}
	for _, candidate := range candidates {
		if path, err := lookPath(candidate); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("uv executable not found")
}

func uvCandidatePaths() []string {
	home, _ := os.UserHomeDir()
	var paths []string
	name := "uv"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if home != "" {
		paths = append(paths, filepath.Join(home, ".local", "bin", name), filepath.Join(home, ".cargo", "bin", name))
	}
	if runtime.GOOS == "windows" {
		if dir := os.Getenv("APPDATA"); dir != "" {
			paths = append(paths, filepath.Join(dir, "astral", "uv", "bin", name))
		}
		if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
			paths = append(paths, filepath.Join(dir, "Programs", "uv", name))
		}
	} else {
		paths = append(paths, "/usr/local/bin/uv", "/opt/homebrew/bin/uv", "/usr/bin/uv")
	}
	absPaths := paths[:0]
	for _, path := range paths {
		if filepath.IsAbs(path) {
			absPaths = append(absPaths, path)
		}
	}
	return absPaths
}

// Respect explicit uv settings; otherwise match pip's activated-venv/system choice.
func buildEnv(env []string) []string {
	venv, configured := false, false
	for _, entry := range env {
		key, value, _ := strings.Cut(entry, "=")
		if runtime.GOOS == "windows" {
			key = strings.ToUpper(key)
		}
		if (key == "VIRTUAL_ENV" || key == "CONDA_PREFIX") && value != "" {
			venv = true
		}
		if key == "UV_SYSTEM_PYTHON" {
			configured = true
		}
	}
	result := append([]string{}, env...)
	if !venv && !configured {
		result = append(result, "UV_SYSTEM_PYTHON=1")
	}
	return result
}
