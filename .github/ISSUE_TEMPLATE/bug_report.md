---
name: Bug Report
about: Create a report to help us improve uvpip
title: 'fix: '
labels: ['bug', 'triage']
assignees: ''
---

> Please review our [Contributing Guidelines](CONTRIBUTING.md) before submitting.

## Bug Description
A clear and concise description of the bug.

## Steps to Reproduce
1. Command executed: `uvpip ...`
2. Environment: `Python version`, `uv version`, `OS` (Windows / Linux / macOS)
3. Observed behavior vs Expected behavior

## Terminal Output / Logs
```text
# Run with UVPIP_DEBUG=1 to capture diagnostic telemetry
UVPIP_DEBUG=1 uvpip <command>
```

## Relevant Invariants
- [ ] Confirmed zero external runtime dependencies (`go list -deps`)
- [ ] Tested offline without touching live package indices where applicable
