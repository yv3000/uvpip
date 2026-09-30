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

// newLogger returns a structured (key=value) debug logger writing to w when
// UVPIP_DEBUG is enabled. Otherwise it discards records, so uv's own stderr
// stream reaches the user unchanged. Callers must not log argument values or
// environment contents: pip arguments can embed index credentials.
func newLogger(w io.Writer, debug string) *slog.Logger {
	if !debugEnabled(debug) {
		return slog.New(slog.DiscardHandler)
	}
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug})).With("component", "uvpip")
}
