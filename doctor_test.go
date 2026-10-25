package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
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
	if code := runDoctor(&out, quiet); code != 1 || strings.Count(out.String(), "not a recognized") != 2 {
		t.Fatalf("code %d: %s", code, &out)
	}
	t.Setenv("UVPIP_TEST_MODE", "exit")
	out.Reset()
	if code := runDoctor(&out, quiet); code != 1 || strings.Count(out.String(), "[!!]") != 4 {
		t.Fatalf("failed executable accepted %d: %s", code, &out)
	}
	t.Setenv("PATH", t.TempDir())
	t.Setenv("UVPIP_UV", filepath.Join(dir, "missing"))
	out.Reset()
	if code := runDoctor(&out, quiet); code != 1 || strings.Count(out.String(), "[!!]") != 4 {
		t.Fatalf("missing dependencies accepted %d: %s", code, &out)
	}
}

func TestDoctorStructuredLogging(t *testing.T) {
	fixtureUV(t)
	t.Setenv("UVPIP_TEST_MODE", "version")
	t.Setenv("UVPIP_DEBUG", "1")
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	t.Setenv("UVPIP_UV", filepath.Join(dir, "missing"))

	var stdout, stderr bytes.Buffer
	code := runCLI([]string{"doctor"}, nil, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit 1 for failed doctor run, got %d", code)
	}
	got := stderr.String()
	if !strings.Contains(got, `msg="doctor check failed"`) || !strings.Contains(got, "target=uv") || !strings.Contains(got, "component=uvpip") {
		t.Fatalf("expected structured log for doctor failure in stderr with component=uvpip, got: %q", got)
	}
}

func TestDoctorShimErrorWrapping(t *testing.T) {
	dir := t.TempDir()
	fakePip := filepath.Join(dir, "pip")
	if runtime.GOOS == "windows" {
		fakePip += ".cmd"
	}
	if err := os.WriteFile(fakePip, []byte("echo not a shim"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	var out bytes.Buffer
	var logBuf bytes.Buffer
	log := newLogger(&logBuf, "1")
	code := runDoctor(&out, log)
	if code != 1 {
		t.Fatalf("expected doctor failure, got code %d", code)
	}
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "component=uvpip") {
		t.Errorf("expected component=uvpip in logs, got %s", logOutput)
	}
	if !strings.Contains(out.String(), "not a recognized uvpip shim") {
		t.Errorf("expected shim error message in output, got %s", out.String())
	}
}

func TestDoctorInheritedOutputPipes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell fixture; normal doctor probes are tested on every OS")
	}
	for _, tt := range []struct {
		name     string
		spawn    string
		wantCode int
	}{
		{"closed pipes", "", 0},
		{"inherited pipes", "/bin/sleep 3 &\necho $! >> \"$UVPIP_DOCTOR_CHILD_PID\"\n", 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			pidFile := filepath.Join(dir, "children")
			t.Setenv("UVPIP_DOCTOR_CHILD_PID", pidFile)
			t.Cleanup(func() {
				data, _ := os.ReadFile(pidFile)
				for _, field := range strings.Fields(string(data)) {
					pid, err := strconv.Atoi(field)
					if err != nil {
						t.Error(err)
						continue
					}
					if child, err := os.FindProcess(pid); err == nil {
						_ = child.Kill() // The bounded fixture may already have exited.
					}
				}
			})
			for _, name := range []string{"uv", "python", "pip", "pip3"} {
				content := "#!/bin/sh\n" + tt.spawn + "printf 'offline fixture\\n'\n"
				if name == "pip" || name == "pip3" {
					content = "#!/bin/sh\nexec \"$(dirname \"$0\")/uvpip\" \"$@\"\n"
				}
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0700); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("UVPIP_UV", filepath.Join(dir, "uv"))
			t.Setenv("PATH", dir)
			var out bytes.Buffer
			started := time.Now()
			if code := runDoctor(&out, quiet); code != tt.wantCode {
				t.Fatalf("exit %d, want %d: %s", code, tt.wantCode, &out)
			}
			if elapsed := time.Since(started); elapsed >= 5*time.Second {
				t.Fatalf("inherited pipes delayed doctor for %s", elapsed)
			}
			if tt.wantCode != 0 && strings.Count(out.String(), "[!!]") != 2 {
				t.Fatalf("expected both version probes to reject inherited pipes: %s", &out)
			}
		})
	}
}

func TestHealthCheckAndMetrics(t *testing.T) {
	// Test CheckHealth
	status := CheckHealth(nil)
	if status.Status != "ok" && status.Status != "degraded" {
		t.Errorf("unexpected status %q", status.Status)
	}
	if len(status.Probes) != 4 {
		t.Errorf("expected 4 probes, got %d", len(status.Probes))
	}
	if status.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}

	// Test runHealthCheck
	var healthOut bytes.Buffer
	healthCode := runHealthCheck(&healthOut, nil)
	if healthCode != 0 && healthCode != 1 {
		t.Errorf("unexpected health exit code %d", healthCode)
	}
	var parsedHealth HealthStatus
	if err := json.Unmarshal(healthOut.Bytes(), &parsedHealth); err != nil {
		t.Fatalf("failed to parse health JSON output: %v; output: %s", err, healthOut.String())
	}
	if parsedHealth.Status == "" {
		t.Error("expected parsed status")
	}

	// Test CollectMetrics
	metrics := CollectMetrics()
	if metrics.GoVersion == "" || metrics.GOOS == "" || metrics.NumCPU <= 0 {
		t.Errorf("invalid metrics: %+v", metrics)
	}

	// Test runMetrics
	var metricsOut bytes.Buffer
	metricsCode := runMetrics(&metricsOut, nil)
	if metricsCode != 0 {
		t.Errorf("unexpected metrics exit code %d", metricsCode)
	}
	var parsedMetrics MetricsData
	if err := json.Unmarshal(metricsOut.Bytes(), &parsedMetrics); err != nil {
		t.Fatalf("failed to parse metrics JSON output: %v; output: %s", err, metricsOut.String())
	}
	if parsedMetrics.NumGoroutine <= 0 {
		t.Errorf("expected positive goroutines, got %d", parsedMetrics.NumGoroutine)
	}
}

func TestRunHealthCheck_Degraded(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("UVPIP_UV", "")
	var out bytes.Buffer
	code := runHealthCheck(&out, quiet)
	if code != 1 {
		t.Fatalf("expected code 1 for degraded health, got %d", code)
	}
	if !strings.Contains(out.String(), `"status": "degraded"`) {
		t.Fatalf("expected degraded in output: %s", out.String())
	}
}

func TestRunMetrics_Logging(t *testing.T) {
	var out bytes.Buffer
	var logBuf bytes.Buffer
	log := NewLogger(&logBuf, "1")
	code := runMetrics(&out, log)
	if code != 0 {
		t.Fatalf("expected code 0, got %d", code)
	}
	if !strings.Contains(logBuf.String(), "telemetry metrics collected") {
		t.Fatalf("expected telemetry log in debug output: %s", logBuf.String())
	}
}

