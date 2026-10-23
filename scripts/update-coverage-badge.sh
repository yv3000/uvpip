#!/bin/sh
# scripts/update-coverage-badge.sh - Updates README.md coverage badge from coverage.txt
set -eu

COVERAGE_FILE="${1:-coverage.txt}"
README_FILE="${2:-README.md}"

if [ ! -f "$COVERAGE_FILE" ]; then
	if [ -f "coverage.out" ]; then
		go tool cover -func=coverage.out > "$COVERAGE_FILE"
	else
		printf 'Error: coverage file "%s" not found\n' "$COVERAGE_FILE" >&2
		exit 1
	fi
fi

TOTAL=$(awk '$1 == "total:" { sub("%", "", $3); print $3 }' "$COVERAGE_FILE")

if [ -z "$TOTAL" ]; then
	printf 'Error: could not parse total coverage percentage from %s\n' "$COVERAGE_FILE" >&2
	exit 1
fi

ROUNDED=$(awk -v total="$TOTAL" 'BEGIN { printf "%.0f", total }')

printf 'Parsed coverage: %s%% (badge target: %s%%)\n' "$TOTAL" "$ROUNDED"

tmp_file="${README_FILE}.tmp.$$"
sed -E "s|(https://img.shields.io/badge/Coverage-)[0-9]+%25(-brightgreen\.svg)|\1${ROUNDED}%25\2|g" "$README_FILE" > "$tmp_file"
mv "$tmp_file" "$README_FILE"

printf 'Updated %s coverage badge to %s%%\n' "$README_FILE" "$ROUNDED"
