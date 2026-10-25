package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Doctor checks executables, not parent-shell aliases/functions it cannot observe.
func runDoctor(out io.Writer, log *slog.Logger) int {
	if log != nil {
		log.Debug("doctor started", "component", "uvpip")
	}
	// Explicitly validate environment override boundaries at doctor entry point
	if err := ValidateInputs(CLIInputs{Args: []string{"doctor"}, UVOverride: os.Getenv("UVPIP_UV")}); err != nil {
		if log != nil {
			log.Info("doctor input validation failed", "component", "uvpip", "error", err)
		}
		fmt.Fprintf(out, "[!!] uv override invalid: %v\n", err)
		return 1
	}
	fmt.Fprintln(out, "uvpip doctor (PATH executables; shell functions are not visible)")
	code := 0
	for _, name := range []string{"uv", "pip", "pip3", "python"} {
		resolved, err := probeTarget(name, true)
		if err != nil {
			if log != nil {
				log.Info("doctor check failed", "component", "uvpip", "target", name, "error", err)
			}
			fmt.Fprintf(out, "[!!] %s: %v\n", name, err)
			code = 1
		} else {
			if log != nil {
				log.Debug("doctor check passed", "component", "uvpip", "target", name, "resolved", resolved)
			}
			fmt.Fprintf(out, "[OK] %s: %s\n", name, resolved)
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

// probeTarget inspects a single named executable or shim and returns its resolved path or error.
// If probeVersion is true, it also executes --version with a bounded timeout and drain bound.
func probeTarget(name string, probeVersion bool) (string, error) {
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
	if err != nil {
		return "", err
	}
	if name == "pip" || name == "pip3" {
		if !isUvpipShim(path) {
			return path, fmt.Errorf("%w: %s; check PATH precedence", ErrNotShim, path)
		}
		return path, nil
	}
	if probeVersion {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, path, "--version") //nolint:gosec // G204: intentional probe of a resolved executable.
		cmd.WaitDelay = time.Second
		output, runErr := cmd.Output()
		if runErr != nil {
			return path, runErr
		}
		return path + " (" + strings.TrimSpace(string(output)) + ")", nil
	}
	return path, nil
}

// HealthProbe represents an individual component health check outcome.
type HealthProbe struct {
	Component string `json:"component"`
	Status    string `json:"status"`
	Details   string `json:"details,omitempty"`
	Error     string `json:"error,omitempty"`
}

// HealthStatus aggregates overall CLI health check diagnostics.
type HealthStatus struct {
	Status    string        `json:"status"`
	Healthy   bool          `json:"healthy"`
	Timestamp string        `json:"timestamp"`
	Probes    []HealthProbe `json:"probes"`
}

// CheckHealth executes diagnostic probes on PATH executables and shims.
func CheckHealth(log *slog.Logger) HealthStatus {
	healthy := true
	var probes []HealthProbe
	for _, name := range []string{"uv", "pip", "pip3", "python"} {
		resolved, err := probeTarget(name, false)
		if err != nil {
			healthy = false
			probes = append(probes, HealthProbe{
				Component: name,
				Status:    "unhealthy",
				Error:     err.Error(),
			})
			if log != nil {
				log.Info("health probe failed", "component", "uvpip", "target", name, "error", err)
			}
		} else {
			probes = append(probes, HealthProbe{
				Component: name,
				Status:    "healthy",
				Details:   resolved,
			})
			if log != nil {
				log.Debug("health probe passed", "component", "uvpip", "target", name, "path", resolved)
			}
		}
	}
	statusStr := "ok"
	if !healthy {
		statusStr = "degraded"
	}
	return HealthStatus{
		Status:    statusStr,
		Healthy:   healthy,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Probes:    probes,
	}
}

// runHealthCheck prints health status in JSON for automation and returns exit code 0 or 1.
func runHealthCheck(out io.Writer, log *slog.Logger) int {
	status := CheckHealth(log)
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		fmt.Fprintf(out, `{"status":"error","healthy":false,"error":%q}`+"\n", err.Error())
		return 1
	}
	fmt.Fprintln(out, string(data))
	if !status.Healthy {
		return 1
	}
	return 0
}

// MetricsData captures runtime telemetry and memory metrics.
type MetricsData struct {
	GoVersion       string `json:"go_version"`
	GOOS            string `json:"goos"`
	GOARCH          string `json:"goarch"`
	NumGoroutine    int    `json:"num_goroutine"`
	NumCPU          int    `json:"num_cpu"`
	AllocBytes      uint64 `json:"alloc_bytes"`
	TotalAllocBytes uint64 `json:"total_alloc_bytes"`
	SysBytes        uint64 `json:"sys_bytes"`
	HeapAllocBytes  uint64 `json:"heap_alloc_bytes"`
	HeapObjects     uint64 `json:"heap_objects"`
}

// CollectMetrics returns execution telemetry and Go memory statistics.
func CollectMetrics() MetricsData {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return MetricsData{
		GoVersion:       runtime.Version(),
		GOOS:            runtime.GOOS,
		GOARCH:          runtime.GOARCH,
		NumGoroutine:    runtime.NumGoroutine(),
		NumCPU:          runtime.NumCPU(),
		AllocBytes:      m.Alloc,
		TotalAllocBytes: m.TotalAlloc,
		SysBytes:        m.Sys,
		HeapAllocBytes:  m.HeapAlloc,
		HeapObjects:     m.HeapObjects,
	}
}

// runMetrics outputs runtime telemetry as JSON.
func runMetrics(out io.Writer, log *slog.Logger) int {
	metrics := CollectMetrics()
	if log != nil {
		log.Debug("telemetry metrics collected", "component", "uvpip", "heap_objects", metrics.HeapObjects)
	}
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		fmt.Fprintf(out, `{"error":%q}`+"\n", err.Error())
		return 1
	}
	fmt.Fprintln(out, string(data))
	return 0
}
