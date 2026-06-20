package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// runUV executes: uv <args> with full stdio inheritance.
// Returns the process exit code.
// stdio is ALWAYS inherited — this is critical for interactive tools to work.
func runUV(args []string) int {
	uvPath, err := findUV()
	if err != nil {
		// uv not found — try to auto-install it
		fmt.Fprintln(os.Stderr, "[uvpip] uv not found. Attempting auto-install...")
		if installErr := autoInstallUV(); installErr != nil {
			fmt.Fprintf(os.Stderr, "[uvpip] auto-install failed: %v\n", installErr)
			fmt.Fprintln(os.Stderr, "[uvpip] Please install uv manually: https://docs.astral.sh/uv/getting-started/installation/")
			return 1
		}
		uvPath, err = findUV()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[uvpip] uv still not found after install: %v\n", err)
			return 1
		}
	}

	cmd := exec.Command(uvPath, args...)

	// CRITICAL: inherit all three stdio streams
	// This ensures interactive output, progress bars, prompts all work correctly
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment overrides so scaffold tools and child processes
	// that call pip internally also route through uv
	cmd.Env = buildEnv(uvPath)

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return 1
	}
	return 0
}

// findUV searches for the uv binary on the current system.
func findUV() (string, error) {
	// First try: PATH lookup
	if path, err := exec.LookPath("uv"); err == nil {
		return path, nil
	}

	// Second try: common install locations per OS
	candidates := uvCandidatePaths()
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("uv binary not found")
}

// uvCandidatePaths returns platform-specific locations where uv might be installed.
func uvCandidatePaths() []string {
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "windows":
		appdata := os.Getenv("APPDATA")
		localappdata := os.Getenv("LOCALAPPDATA")
		return []string{
			home + `\.cargo\bin\uv.exe`,
			home + `\.local\bin\uv.exe`,
			appdata + `\astral\uv\bin\uv.exe`,
			localappdata + `\Programs\uv\uv.exe`,
		}
	case "darwin", "linux":
		return []string{
			home + "/.local/bin/uv",
			home + "/.cargo/bin/uv",
			"/usr/local/bin/uv",
			"/opt/homebrew/bin/uv",
			"/usr/bin/uv",
		}
	default:
		return []string{home + "/.local/bin/uv"}
	}
}

// autoInstallUV installs uv using the official installer for the current OS.
func autoInstallUV() error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass",
			"-Command", "irm https://astral.sh/uv/install.ps1 | iex")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	case "darwin", "linux":
		cmd := exec.Command("sh", "-c", "curl -fsSL https://astral.sh/uv/install.sh | sh")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// buildEnv constructs the environment for the uv subprocess.
// Sets overrides so child processes that call pip also use uv.
func buildEnv(uvPath string) []string {
	env := os.Environ()

	// Only set UV_SYSTEM_PYTHON if no virtualenv is currently active.
	// Detect an active venv via the VIRTUAL_ENV environment variable, which
	// is set by every standard venv/virtualenv activation script.
	_, venvActive := os.LookupEnv("VIRTUAL_ENV")

	overrides := map[string]string{
		// Points any child "pip" call at our uvpip binary so nested installs also go through uv
		"PIP_PYTHON": uvPath,
	}

	if !venvActive {
		// No venv active: let uv target the system interpreter so `pip install X`
		// works even when the user hasn't created a venv (matches plain pip's behavior).
		overrides["UV_SYSTEM_PYTHON"] = "1"
	}
	// If a venv IS active, do not set UV_SYSTEM_PYTHON at all — uv will
	// correctly detect and use the active venv on its own, exactly like
	// plain `uv pip install` does today.

	// Build final env — overrides win over existing values
	result := []string{}
	for _, e := range env {
		key := strings.SplitN(e, "=", 2)[0]
		if _, overridden := overrides[key]; !overridden {
			result = append(result, e)
		}
	}
	for k, v := range overrides {
		result = append(result, k+"="+v)
	}
	return result
}
