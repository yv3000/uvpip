package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

const version = "1.1.0"

func main() {
	os.Exit(runCLI(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func runCLI(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	log := NewLogger(stderr, os.Getenv("UVPIP_DEBUG"))
	slog.SetDefault(log)

	// No args: print usage
	if len(args) == 0 {
		printUsage(stdout)
		return 0
	}

	// Handle uvpip-native commands
	switch args[0] {
	case "doctor":
		return runDoctor(stdout, log)
	case "--health", "health":
		return runHealthCheck(stdout, log)
	case "--metrics", "metrics":
		return runMetrics(stdout, log)
	case "--version", "-V":
		fmt.Fprintf(stdout, "uvpip v%s (uv-powered pip wrapper)\n", version)
		return 0
	case "--help", "-h":
		printUsage(stdout)
		return 0
	}

	// All other args: validate against canonical schema, translate pip → uv, and run
	if err := CanonicalValidationSchema.ValidateWithLogger(args, log); err != nil {
		fmt.Fprintf(stderr, "[uvpip] %v; run uvpip --help for usage\n", err)
		return 2
	}
	uvArgs := translateToUV(args)
	// Log only the routed uv command and count; argument values may hold credentials.
	log.Debug("translated command", "uv_command", uvArgs[0], "arg_count", len(uvArgs))
	return runUV(uvArgs, stdin, stdout, stderr, log)
}

func printUsage(out io.Writer) {
	fmt.Fprintf(out, `uvpip - transparent pip wrapper powered by uv
version: %s

Usage:
  pip <command> [args]     route supported pip commands through uv
  uvpip doctor            check executable and PATH shim health
  uvpip --health          structured JSON health check for monitoring
  uvpip --metrics         structured runtime telemetry and memory metrics
  uvpip --version         show wrapper version

Examples:
  pip install requests
  pip install -r requirements.txt
  pip uninstall requests
  pip list
  pip freeze
  pip show requests

Unknown commands and flags are passed to uv, which may reject them.
Use python -m pip to bypass the wrapper.
`, version)
}
