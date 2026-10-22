package main

import (
	"io"
	"log/slog"
	"strings"
)

// debugEnabled reports whether a UVPIP_DEBUG value opts in to diagnostics.
func debugEnabled(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

// NewLogger returns a structured (key=value) slog.Logger writing to w.
// In normal operation (UVPIP_DEBUG unset), it logs at LevelInfo so that
// critical execution failures are recorded without emitting noisy diagnostics
// on successful runs. When UVPIP_DEBUG is enabled, it logs at LevelDebug.
// Callers must not log argument values or environment contents: pip arguments
// can embed index credentials.
func NewLogger(w io.Writer, debug string) *slog.Logger {
	level := slog.LevelInfo
	if debugEnabled(debug) {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})).With("component", "uvpip")
}

// newLogger preserves compatibility with internal callers.
func newLogger(w io.Writer, debug string) *slog.Logger {
	return NewLogger(w, debug)
}
