package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// parseCoverageTotal extracts the total statement percentage from coverage summary text.
func parseCoverageTotal(content string) (float64, error) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[0] == "total:" {
			pctStr := strings.TrimSuffix(fields[len(fields)-1], "%")
			val, err := strconv.ParseFloat(pctStr, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid coverage float: %w", err)
			}
			return val, nil
		}
	}
	return 0, fmt.Errorf("no total coverage line found")
}

// assertCoverageGate validates that total statement coverage meets or exceeds threshold.
func assertCoverageGate(content string, threshold float64) error {
	total, err := parseCoverageTotal(content)
	if err != nil {
		return err
	}
	if total < threshold {
		return fmt.Errorf("coverage %.1f%% below minimum threshold %.1f%%", total, threshold)
	}
	return nil
}

// TestCoverageGate asserts that the coverage threshold gate passes for valid coverage
// and deterministically rejects synthetic sub-threshold coverage samples.
func TestCoverageGate(t *testing.T) {
	const threshold = 90.0

	t.Run("passes_above_threshold", func(t *testing.T) {
		sample := `github.com/yv3000/uvpip/main.go:16:		runCLI		100.0%
total:						(statements)	96.5%`
		if err := assertCoverageGate(sample, threshold); err != nil {
			t.Fatalf("expected sample to pass gate, got: %v", err)
		}
	})

	t.Run("fails_below_threshold", func(t *testing.T) {
		lowSample := `github.com/yv3000/uvpip/main.go:16:		runCLI		50.0%
total:						(statements)	54.2%`
		err := assertCoverageGate(lowSample, threshold)
		if err == nil {
			t.Fatal("expected sub-threshold sample to fail gate, but passed")
		}
		if !strings.Contains(err.Error(), "below minimum threshold") {
			t.Fatalf("expected below threshold error, got: %v", err)
		}
	})

	t.Run("fails_on_malformed_sample", func(t *testing.T) {
		err := assertCoverageGate("no summary here", threshold)
		if err == nil {
			t.Fatal("expected malformed sample to fail")
		}
	})
}
