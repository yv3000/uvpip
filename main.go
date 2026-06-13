package main

import (
	"fmt"
	"os"
)

const version = "1.0.0"

func main() {
	args := os.Args[1:]

	// No args: print usage
	if len(args) == 0 {
		printUsage()
		os.Exit(0)
	}

	// Handle uvpip-native commands
	switch args[0] {
	case "doctor":
		runDoctor()
		return
	case "--version", "-V", "-v":
		fmt.Printf("uvpip v%s (uv-powered pip wrapper)\n", version)
		return
	case "--help", "-h":
		printUsage()
		return
	}

	// All other args: translate pip → uv and run
	uvArgs := translateToUV(args)
	exitCode := runUV(uvArgs)
	os.Exit(exitCode)
}

func printUsage() {
	fmt.Println("uvpip — transparent pip wrapper powered by uv")
	fmt.Printf("version: %s\n\n", version)
	fmt.Println("Usage:")
	fmt.Println("  pip <command> [args]    — all pip commands work as normal")
	fmt.Println("  uvpip doctor            — check system health")
	fmt.Println("  uvpip --version         — show version")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  pip install requests")
	fmt.Println("  pip install -r requirements.txt")
	fmt.Println("  pip uninstall requests")
	fmt.Println("  pip list")
	fmt.Println("  pip freeze")
	fmt.Println("  pip show requests")
	fmt.Println("")
	fmt.Println("uv runs under the hood — same commands, 10-100x faster.")
}
