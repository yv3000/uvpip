package main

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidArgs indicates invalid or empty CLI arguments.
	ErrInvalidArgs = errors.New("invalid arguments")

	// ErrUsage indicates CLI argument or flag usage violations (exit status 2).
	ErrUsage = fmt.Errorf("%w: usage error", ErrInvalidArgs)

	// ErrInvalidOverride indicates malformed UVPIP_UV environment variable configuration.
	ErrInvalidOverride = errors.New("invalid uv override")

	// ErrTranslation indicates command translation failure.
	ErrTranslation = errors.New("command translation failed")

	// ErrUVNotFound indicates uv could not be located in PATH or candidate locations.
	ErrUVNotFound = errors.New("uv executable not found")

	// ErrCannotExecute indicates uv binary was found but cannot be started.
	ErrCannotExecute = errors.New("cannot execute uv binary")

	// ErrRecursiveExecution prevents uvpip from calling itself in an infinite loop.
	ErrRecursiveExecution = errors.New("uv resolves to uvpip itself; refusing recursive execution")

	// ErrNotShim indicates an inspected executable is not a recognized uvpip shim wrapper.
	ErrNotShim = errors.New("not a recognized uvpip shim")

	// ErrFlagConflict indicates mutually exclusive or conflicting flags were supplied.
	ErrFlagConflict = fmt.Errorf("%w: conflicting flags", ErrUsage)

	// ErrArgBounds indicates an argument count or length exceeded maximum schema limits.
	ErrArgBounds = fmt.Errorf("%w: argument bounds exceeded", ErrUsage)

	// ErrNullByte indicates an argument contains a forbidden null byte.
	ErrNullByte = fmt.Errorf("%w: forbidden null byte", ErrUsage)

	// ErrEmptyCommand indicates a command was empty or blank.
	ErrEmptyCommand = fmt.Errorf("%w: pip command is empty", ErrUsage)
)
