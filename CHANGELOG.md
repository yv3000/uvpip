# Changelog

## Unreleased

These changes describe revised source, not the downloadable v1.0.0 binaries.
The source version constant remains `1.0.0`; identify builds by source commit.

- Add centralized CLI boundary validation schema (`ValidateInputs`, `CLIInputs`)
  coordinating command arguments and path overrides with table-driven assertions (`validate.go`).
- Add structured Info-level failure audit trail logging in runner for binary discovery
  and process execution failures outside debug mode (`runner.go`, `logging.go`).
- Add granular shell installer test coverage for missing download tools and existing binary
  reuse branches (`scripts/test-installers.sh`).
- Document external contribution channels, commit pairing rules, and PR review
  policy in `CONTRIBUTING.md`.
- Reject a blank pip command with exit status 2 before uv starts, and report
  blank or whitespace-padded `UVPIP_UV` values precisely (`validate.go`).
- Add offline Go integration test (`integration_test.go`) exercising the full
  install, doctor diagnostics, CLI translation, and uninstall flow.
- Remove runtime uv downloads and implicit recovery; missing uv returns 127,
  launch failures return 126, and normal child exit codes pass through.
- Support an absolute `UVPIP_UV` override, reject current-directory `ErrDot`
  executable resolution, and exclude relative fallback candidates.
- Reject uv resolving to uvpip's own executable with `os.SameFile`, including
  hardlinks/symlinks to the same file, not separate byte-identical copies.
- Preserve explicit uv environment settings and respect nonempty `CONDA_PREFIX`
  as well as `VIRTUAL_ENV` when deciding whether to default `UV_SYSTEM_PYTHON=1`.
- Set the Go language minimum to 1.26.0 and pin verification to Go 1.26.8. Remove
  the empty `go.sum`; the runtime has no third-party Go dependencies.
- Pin govulncheck and actionlint as Go `tool` directives, so `go.sum` now records
  checksums for development tools only (never linked into the binary), CI runs
  `go mod tidy -diff` and `go mod verify`, and Dependabot updates the tools.
- Replace standalone Staticcheck with golangci-lint v2.14.0 (`.golangci.yml`:
  standard linters plus gosec, errorlint, gocritic, misspell, and others).
- Split CI into directly named steps: format, module checks, vet, build, test
  with coverage (race on Linux/macOS), coverage gate, govulncheck, installer
  tests, golangci-lint, actionlint, ShellCheck, PowerShell parsing, and Gitleaks.
- Add opt-in `UVPIP_DEBUG=1` structured (`log/slog`) stderr diagnostics for uv
  discovery, environment defaulting, and exit status. Argument values and
  environment contents are never logged. Default output is unchanged.
- Add `.env.example` and a README table documenting every supported variable.
- Remove the legacy `v1` tree: 1,529 files and 313,277 lines.
- Forward `-v` to uv, reserve `--version`/`-V` for wrapper version, and preserve
  uninstall arguments after `--`.
- Report diagnostic failures with nonzero status, recognize shipped text shims,
  and bound uv/Python version probes. Parent-shell functions are not inspected.
- Revise per-user installers/uninstallers with staging, optional supplied
  checksums, local binary inputs, integration opt-outs, and exact profile-block
  handling. Do not elevate or edit Windows System PATH; retain unrelated files.
- Add offline installer checks and isolated verification entry points, pinned
  lint/vulnerability tools, a three-OS CI matrix, Linux race tests, a 90% minimum
  statement-coverage gate, and Dependabot for Go modules and Actions.
- Replace unverified speed/compatibility claims with explicit behavior, trust
  boundaries, source-versus-release guidance, required fresh-clone verification,
  and a manual release checklist.

## v1.0.0 - 2026-06-20

- Initial published release with six downloadable assets: Windows, macOS, and
  Linux, each for amd64 and arm64.
- Introduced pip-to-uv command translation and per-user shims/profile integration.

Release record: <https://github.com/yv3000/uvpip/releases/tag/v1.0.0>.
