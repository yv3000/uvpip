package main

import (
	"fmt"
	"io"
	"os"
)

const version = "1.0.0"

func main() {
	os.Exit(runCLI(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func runCLI(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// No args: print usage
	if len(args) == 0 {
		printUsage(stdout)
		return 0
	}

	// Handle uvpip-native commands
	switch args[0] {
	case "doctor":
		return runDoctor(stdout)
	case "--version", "-V":
		fmt.Fprintf(stdout, "uvpip v%s (uv-powered pip wrapper)\n", version)
		return 0
	case "--help", "-h":
		printUsage(stdout)
		return 0
	}

	// All other args: translate pip → uv and run
	log := newLogger(stderr, os.Getenv("UVPIP_DEBUG"))
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
