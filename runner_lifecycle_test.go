package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func runInheritedPipeFixture() {
	if os.Getenv("UVPIP_TEST_MODE") == "hold-output" {
		time.Sleep(8 * time.Second)
		return
	}
	child := exec.Command(os.Args[0])
	child.Env = append(os.Environ(), "UVPIP_TEST_MODE=hold-output")
	child.Stdout, child.Stderr = os.Stdout, os.Stderr
	if err := child.Start(); err != nil {
		panic(err)
	}
	if err := os.WriteFile(os.Getenv("UVPIP_TEST_PID"), []byte(strconv.Itoa(child.Process.Pid)), 0600); err != nil {
		panic(err)
	}
	_ = child.Process.Release()
	if os.Getenv("UVPIP_TEST_MODE") == "inherit-output-failure" {
		fmt.Fprint(os.Stderr, "child failure")
		os.Exit(23)
	}
}

func testRunUVInheritedOutputPipes(t *testing.T) {
	for _, tt := range []struct {
		name, mode string
		wantCode   int
	}{
		{"ordinary output", "version", exitOK},
		{"descendant retains pipes", "inherit-output", exitGeneral},
		{"child failure retains pipes", "inherit-output-failure", 23},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fixtureUV(t)
			pidFile := filepath.Join(t.TempDir(), "descendant.pid")
			t.Setenv("UVPIP_TEST_MODE", tt.mode)
			t.Setenv("UVPIP_TEST_PID", pidFile)
			t.Cleanup(func() {
				data, err := os.ReadFile(pidFile)
				if err != nil {
					return
				}
				pid, err := strconv.Atoi(string(data))
				if err != nil {
					t.Error(err)
					return
				}
				if child, err := os.FindProcess(pid); err == nil {
					_ = child.Kill()
					_, _ = child.Wait()
				}
			})
			var out, stderr bytes.Buffer
			started := time.Now()
			code := runUVContext(context.Background(), nil, nil, &out, &stderr, quiet)
			if code != tt.wantCode {
				t.Fatalf("exit=%d want=%d stderr=%s", code, tt.wantCode, &stderr)
			}
			if time.Since(started) >= 10*time.Second {
				t.Fatal("runner waited for descendant pipes")
			}
			if tt.wantCode == exitGeneral && !strings.Contains(stderr.String(), "output pipes") {
				t.Fatalf("missing output-pipe diagnostic: %s", &stderr)
			}
			if tt.wantCode == 23 && stderr.String() != "child failure" {
				t.Fatalf("child failure diagnostic changed: %q", &stderr)
			}
		})
	}
}

// The readiness output proves that cancellation happens after process startup.
func runCancellationFixture() {
	fmt.Fprint(os.Stdout, "ready")
	time.Sleep(10 * time.Second) // Bound the fixture even if cancellation regresses.
}

type cancelOnReady struct {
	cancel context.CancelFunc
	ready  bool
}

func (w *cancelOnReady) Write(p []byte) (int, error) {
	if len(p) > 0 {
		w.ready = true
		w.cancel()
	}
	return len(p), nil
}

func TestRunUVRunningChildCancellation(t *testing.T) {
	for _, tt := range []struct {
		name        string
		beforeStart bool
	}{
		{"before startup", true},
		{"after readiness", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fixtureUV(t)
			t.Setenv("UVPIP_TEST_MODE", "wait-for-cancel")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			out := &cancelOnReady{cancel: cancel}
			if tt.beforeStart {
				cancel()
			}
			started := time.Now()
			code := runUVContext(ctx, []string{"pip", "list"}, nil, out, io.Discard, quiet)
			if code == exitOK || out.ready == tt.beforeStart {
				t.Fatalf("exit=%d ready=%v beforeStart=%v", code, out.ready, tt.beforeStart)
			}
			if !tt.beforeStart && code == exitCannotExecute {
				t.Fatal("running child was incorrectly reported as a startup failure")
			}
			if elapsed := time.Since(started); elapsed >= 5*time.Second {
				t.Fatalf("cancellation did not promptly reap child: %s", elapsed)
			}
		})
	}
}
