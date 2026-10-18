package main

import (
	"errors"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrInvalidArgs",
			err:      ErrInvalidArgs,
			expected: "invalid arguments",
		},
		{
			name:     "ErrInvalidOverride",
			err:      ErrInvalidOverride,
			expected: "invalid uv override",
		},
		{
			name:     "ErrTranslation",
			err:      ErrTranslation,
			expected: "command translation failed",
		},
		{
			name:     "ErrUVNotFound",
			err:      ErrUVNotFound,
			expected: "uv executable not found",
		},
		{
			name:     "ErrUsage",
			err:      ErrUsage,
			expected: "invalid arguments: usage error",
		},
		{
			name:     "ErrCannotExecute",
			err:      ErrCannotExecute,
			expected: "cannot execute uv binary",
		},
		{
			name:     "ErrRecursiveExecution",
			err:      ErrRecursiveExecution,
			expected: "uv resolves to uvpip itself; refusing recursive execution",
		},
		{
			name:     "ErrNotShim",
			err:      ErrNotShim,
			expected: "not a recognized uvpip shim",
		},
		{
			name:     "ErrFlagConflict",
			err:      ErrFlagConflict,
			expected: "invalid arguments: usage error: conflicting flags",
		},
		{
			name:     "ErrArgBounds",
			err:      ErrArgBounds,
			expected: "invalid arguments: usage error: argument bounds exceeded",
		},
		{
			name:     "ErrNullByte",
			err:      ErrNullByte,
			expected: "invalid arguments: usage error: forbidden null byte",
		},
		{
			name:     "ErrEmptyCommand",
			err:      ErrEmptyCommand,
			expected: "invalid arguments: usage error: pip command is empty",
		},
	}

	if !errors.Is(ErrUsage, ErrInvalidArgs) {
		t.Errorf("expected ErrUsage to wrap ErrInvalidArgs")
	}
	if !errors.Is(ErrFlagConflict, ErrUsage) || !errors.Is(ErrFlagConflict, ErrInvalidArgs) {
		t.Errorf("expected ErrFlagConflict to wrap ErrUsage and ErrInvalidArgs")
	}
	if !errors.Is(ErrArgBounds, ErrUsage) || !errors.Is(ErrArgBounds, ErrInvalidArgs) {
		t.Errorf("expected ErrArgBounds to wrap ErrUsage and ErrInvalidArgs")
	}
	if !errors.Is(ErrNullByte, ErrUsage) || !errors.Is(ErrNullByte, ErrInvalidArgs) {
		t.Errorf("expected ErrNullByte to wrap ErrUsage and ErrInvalidArgs")
	}
	if !errors.Is(ErrEmptyCommand, ErrUsage) || !errors.Is(ErrEmptyCommand, ErrInvalidArgs) {
		t.Errorf("expected ErrEmptyCommand to wrap ErrUsage and ErrInvalidArgs")
	}

	seen := make(map[string]bool)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				t.Fatalf("expected non-nil error for %s", tc.name)
			}
			if tc.err.Error() != tc.expected {
				t.Errorf("expected message %q, got %q", tc.expected, tc.err.Error())
			}
			if !errors.Is(tc.err, tc.err) {
				t.Errorf("expected errors.Is self-equivalence for %s", tc.name)
			}
			if seen[tc.err.Error()] {
				t.Errorf("duplicate error message detected across sentinel errors: %q", tc.err.Error())
			}
			seen[tc.err.Error()] = true
		})
	}
}
