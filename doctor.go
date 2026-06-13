package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// runDoctor prints a health report for the uvpip installation.
func runDoctor() {
	fmt.Println("")
	fmt.Println("  uvpip doctor")
	fmt.Println("  ----------------------------------------")
	fmt.Println("")

	allOK := true

	// Check 1: uvpip binary
	selfPath, _ := os.Executable()
	checkLine("uvpip binary", selfPath, true)

	// Check 2: uv
	uvPath, uvErr := findUV()
	if uvErr != nil {
		checkLine("uv", "NOT FOUND — run installer or: curl -fsSL https://astral.sh/uv/install.sh | sh", false)
		allOK = false
	} else {
		uvVersion := ""
		out, err := exec.Command(uvPath, "--version").Output()
		if err == nil {
			uvVersion = strings.TrimSpace(string(out))
		}
		checkLine("uv", fmt.Sprintf("found at %s  (%s)", uvPath, uvVersion), true)
	}

	// Check 3: pip shim active
	pipPath, pipErr := exec.LookPath("pip")
	if pipErr != nil {
		checkLine("pip shim", "pip not found in PATH", false)
		allOK = false
	} else {
		isOurShim := isUvpipShim(pipPath)
		if isOurShim {
			checkLine("pip shim", fmt.Sprintf("active — %s", pipPath), true)
		} else {
			checkLine("pip shim", fmt.Sprintf("WARNING — pip resolves to real pip at %s (uvpip is not first in PATH)", pipPath), false)
			allOK = false
		}
	}

	// Check 4: pip3 shim active
	pip3Path, pip3Err := exec.LookPath("pip3")
	if pip3Err != nil {
		checkLine("pip3 shim", "pip3 not found in PATH", false)
	} else {
		isOurShim := isUvpipShim(pip3Path)
		if isOurShim {
			checkLine("pip3 shim", fmt.Sprintf("active — %s", pip3Path), true)
		} else {
			checkLine("pip3 shim", fmt.Sprintf("WARNING — pip3 resolves to real pip3 at %s", pip3Path), false)
			allOK = false
		}
	}

	// Check 5: PATH order
	pathEnv := os.Getenv("PATH")
	var sep string
	if runtime.GOOS == "windows" {
		sep = ";"
	} else {
		sep = ":"
	}
	entries := strings.Split(pathEnv, sep)
	uvpipBinIndex := -1
	realPipIndex := -1
	for i, e := range entries {
		if strings.Contains(strings.ToLower(e), ".uvpip") || strings.Contains(strings.ToLower(e), "uvpip") {
			if uvpipBinIndex == -1 {
				uvpipBinIndex = i
			}
		}
		// Common real pip locations
		if strings.Contains(strings.ToLower(e), "python") ||
			strings.Contains(strings.ToLower(e), "site-packages") ||
			strings.Contains(strings.ToLower(e), "scripts") {
			if realPipIndex == -1 {
				realPipIndex = i
			}
		}
	}
	if uvpipBinIndex != -1 && (realPipIndex == -1 || uvpipBinIndex < realPipIndex) {
		checkLine("PATH order", "uvpip bin is before Python/pip in PATH", true)
	} else if uvpipBinIndex == -1 {
		checkLine("PATH order", "uvpip bin directory not found in PATH — restart your terminal", false)
		allOK = false
	} else {
		checkLine("PATH order", "WARNING — a Python/pip path appears before uvpip in PATH", false)
		allOK = false
	}

	// Check 6: Python
	pythonPath, pythonErr := exec.LookPath("python")
	if pythonErr != nil {
		pythonPath, pythonErr = exec.LookPath("python3")
	}
	if pythonErr != nil {
		checkLine("Python", "not found in PATH", false)
	} else {
		out, err := exec.Command(pythonPath, "--version").Output()
		if err == nil {
			checkLine("Python", fmt.Sprintf("found (%s) at %s", strings.TrimSpace(string(out)), pythonPath), true)
		} else {
			checkLine("Python", fmt.Sprintf("found at %s", pythonPath), true)
		}
	}

	fmt.Println("")
	fmt.Println("  ----------------------------------------")
	if allOK {
		fmt.Println("  All checks passed. uvpip is working correctly.")
	} else {
		fmt.Println("  Some checks failed. Restart your terminal, then run 'uvpip doctor' again.")
		fmt.Println("  If issues persist, re-run the installer.")
	}
	fmt.Println("")
}

func checkLine(label, detail string, ok bool) {
	status := "[OK] "
	if !ok {
		status = "[!!] "
	}
	fmt.Printf("  %s %-18s %s\n", status, label, detail)
}

// isUvpipShim returns true if the binary at path is our uvpip shim.
// On Windows: checks if pip.cmd references uvpip.exe
// On Unix: checks if the file references uvpip
func isUvpipShim(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	content := strings.ToLower(string(data))
	return strings.Contains(content, "uvpip")
}
