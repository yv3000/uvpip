package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Doctor checks executables, not parent-shell aliases/functions it cannot observe.
func runDoctor(out io.Writer) int {
	fmt.Fprintln(out, "uvpip doctor (PATH executables; shell functions are not visible)")
	code := 0
	for _, name := range []string{"uv", "pip", "pip3", "python"} {
		var path string
		var err error
		if name == "uv" {
			path, err = findUV()
		} else {
			path, err = exec.LookPath(name)
			if name == "python" && err != nil {
				path, err = exec.LookPath("python3")
			}
		}
		if err == nil {
			if name == "pip" || name == "pip3" {
				if !isUvpipShim(path) {
					err = fmt.Errorf("%s is not a recognized uvpip shim; check PATH precedence", path)
				}
			} else {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				output, runErr := exec.CommandContext(ctx, path, "--version").Output()
				cancel()
				err = runErr
				if err == nil {
					path += " (" + strings.TrimSpace(string(output)) + ")"
				}
			}
		}
		if err != nil {
			fmt.Fprintf(out, "[!!] %s: %v\n", name, err)
			code = 1
		} else {
			fmt.Fprintf(out, "[OK] %s: %s\n", name, path)
		}
	}
	if code != 0 {
		fmt.Fprintln(out, "Check your installation and restart the shell. Use python -m pip to bypass uvpip.")
	}
	return code
}

func isUvpipShim(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 4097))
	if err != nil || len(data) > 4096 || strings.ContainsRune(string(data), 0) {
		return false
	}
	content := strings.ToLower(string(data))
	// Recognize our shipped text shims, not arbitrary binaries containing our name.
	return strings.Contains(content, `exec "$(dirname "$0")/uvpip" "$@"`) ||
		strings.Contains(content, `exec "$home/.uvpip/bin/uvpip" "$@"`) ||
		strings.Contains(content, `"%~dp0uvpip.exe" %*`) ||
		strings.Contains(content, `"%userprofile%\.uvpip\bin\uvpip.exe" %*`)
}
