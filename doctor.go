package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Doctor checks executables, not parent-shell aliases/functions it cannot observe.
func runDoctor(out io.Writer, log *slog.Logger) int {
	if log != nil {
		log.Debug("doctor started", "component", "uvpip")
	}
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
					err = fmt.Errorf("%w: %s; check PATH precedence", ErrNotShim, path)
				}
			} else {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				// path comes from findUV/exec.LookPath; probing its version is the check itself.
				cmd := exec.CommandContext(ctx, path, "--version") //nolint:gosec // G204: intentional probe of a resolved executable.
				// Descendants can inherit output pipes after the probed process exits.
				cmd.WaitDelay = time.Second
				output, runErr := cmd.Output()
				cancel()
				err = runErr
				if err == nil {
					path += " (" + strings.TrimSpace(string(output)) + ")"
				}
			}
		}
		if err != nil {
			if log != nil {
				log.Info("doctor check failed", "component", "uvpip", "target", name, "error", err)
			}
			fmt.Fprintf(out, "[!!] %s: %v\n", name, err)
			code = 1
		} else {
			if log != nil {
				log.Debug("doctor check passed", "component", "uvpip", "target", name, "resolved", path)
			}
			fmt.Fprintf(out, "[OK] %s: %s\n", name, path)
		}
	}
	if code != 0 {
		fmt.Fprintln(out, "Check your installation and restart the shell. Use python -m pip to bypass uvpip.")
	}
	if log != nil {
		log.Debug("doctor completed", "component", "uvpip", "exit_code", code)
	}
	return code
}

func isUvpipShim(path string) bool {
	file, err := os.Open(path) //nolint:gosec // G304: bounded read-only inspection of a PATH-resolved shim.
	if err != nil {
		return false
	}
	defer func() { _ = file.Close() }() // Read-only; a close error cannot change the result.
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
