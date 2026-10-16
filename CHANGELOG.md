# Changelog

## v1.0.2 - 2026-10-15

Feature and reliability release introducing CLI flag conflict validation, structured diagnostics, fallback candidate path expansions, 97%+ test coverage with artifacts, and Windows architecture isolation.

- Add `validateTranslationFlags` rejecting conflicting `-q`/`--quiet` and `-v`/`--verbose` options with exit status 2 (`errUsage`), respecting the `--` separator for subsequent literal operands (`translator.go`, `translator_test.go`).
- Document explicit CLI input validation boundaries, flag conflict contract, and exit codes in `README.md`.
- Add structured slog logging to `runDoctor` with debug trace observability for passing diagnostics and info-level error reporting (`doctor.go`, `doctor_test.go`).
- Add fallback uv candidate discovery paths for Linuxbrew (`/home/linuxbrew/.linuxbrew/bin/uv`) and NixOS (`/run/current-system/sw/bin/uv`) with unit assertions (`runner.go`, `runner_test.go`).
- Add dedicated table-driven test suite `errors_test.go` verifying sentinel errors, achieving 100% test file pairing for all source files.
- Expose global `COVERAGE_THRESHOLD: 90` in `.github/workflows/ci.yml` and pin coverage artifact uploads via `actions/upload-artifact@v7.0.2` with 14-day retention.
- Add isolated Windows architecture download cases (AMD64, ARM64, WoW64) in `scripts/test-installers.ps1`.
- Add pip cache isolation semantics, PowerShell 5.1 and 7+ setup instructions, and expanded external contributor review workflows (`README.md`, `CONTRIBUTING.md`).
- Track community contributor diversification: 3 distinct human contributors active in repository history.

## v1.0.1 - 2026-10-12

Maintenance release introducing CLI input boundary schema defenses, multi-platform candidate path testing, automated GitHub Releases, and threat model documentation.

- Add centralized CLI boundary validation schema (`ValidateInputs`, `CLIInputs`)
  enforcing argument limits, argument length constraints, and null-byte defenses (`validate.go`, `validate_test.go`).
- Add comprehensive cross-platform candidate path and environment casing unit tests (`runner_test.go`).
- Add explicit installer threat model and trust boundaries in `SECURITY.md`, cross-linked in `README.md` and `CONTRIBUTING.md`.
- Add automated multi-platform binary release pipeline in `.github/workflows/release.yml`.
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
