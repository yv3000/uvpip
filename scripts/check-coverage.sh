#!/bin/sh
# scripts/check-coverage.sh - Parses go tool cover output and enforces coverage threshold.
set -eu

COVERAGE_FILE="${1:-coverage.txt}"
THRESHOLD="${2:-${COVERAGE_THRESHOLD:-90}}"

if [ ! -f "$COVERAGE_FILE" ]; then
	if [ -f "coverage.out" ]; then
		go tool cover -func=coverage.out > coverage.txt
		COVERAGE_FILE="coverage.txt"
	else
		printf 'Error: coverage file "%s" not found\n' "$COVERAGE_FILE" >&2
		exit 1
	fi
fi

# Extract total statement coverage percentage from the summary line
TOTAL=$(awk '$1 == "total:" { sub("%", "", $3); print $3 }' "$COVERAGE_FILE")

if [ -z "$TOTAL" ]; then
	printf 'Error: could not parse total coverage percentage from %s\n' "$COVERAGE_FILE" >&2
	exit 1
fi

printf 'Total statement coverage: %s%% (Threshold: %s%%)\n' "$TOTAL" "$THRESHOLD"

awk -v total="$TOTAL" -v min="$THRESHOLD" 'BEGIN {
	if (total !~ /^[0-9]+([.][0-9]+)?$/ || total + 0 < min + 0) {
		printf "Coverage %s%% is missing, invalid or below threshold %s%%\n", total, min > "/dev/stderr"
		exit 1
	}
}'

printf 'Coverage check PASSED (%s%% >= %s%%)\n' "$TOTAL" "$THRESHOLD"
