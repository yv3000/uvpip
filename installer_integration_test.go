package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// TestInstallerIntegration executes scripts/test-installers.sh on POSIX platforms
// within isolated temporary directories to assert installer correctness end-to-end.
func TestInstallerIntegration(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping POSIX installer shell script integration test on Windows")
	}

	testRoot := t.TempDir()
	cacheRoot := t.TempDir()

	scriptPath := filepath.Join("scripts", "test-installers.sh")
	if _, err := os.Stat(scriptPath); err != nil {
		t.Fatalf("installer test script not found: %v", err)
	}

	cmd := exec.Command("sh", scriptPath, testRoot, cacheRoot)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("test-installers.sh failed: %v\nOutput:\n%s", err, string(output))
	}
}
