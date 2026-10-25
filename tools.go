//go:build tools

// Package main declares dev-time build and audit tools for Go module tracking.
// These tools are never linked into the uvpip CLI runtime binary.
package main

import (
	_ "github.com/rhysd/actionlint/cmd/actionlint"
	_ "golang.org/x/vuln/cmd/govulncheck"
)
