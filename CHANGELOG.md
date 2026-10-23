# Changelog

## v1.0.9 - 2026-10-23

Architecture robustness, supply-chain dependency health, and test coverage expansion release introducing Python dependency vulnerability scanning (`pip-audit`), a 90% Python coverage gate, live coverage badge automation, installer edge-case coverage, and offline dry-run error code parity assertions.

- Add supply-chain dependency vulnerability scanning for Python packaging via pinned `pip-audit==2.7.3` in `.github/workflows/ci.yml`.
- Enforce strict 90% Python test coverage gate in CI (`coverage report -m --fail-under=90`) with comprehensive unit test coverage for `uvpip` Python CLI launcher, PATH fallback search, and virtualenv configuration in `tests/test_py_entrypoint.py`.
- Add `scripts/update-coverage-badge.sh` to dynamically synchronize `README.md` test coverage badge with live test metrics, enforced against drift in CI.
- Add offline installer edge-case test suites for trailing slashes and custom directory targets in `scripts/test-installers.ps1` and `scripts/test-installers.sh` (#22, PR #25 by @trulyimmortal).
- Add offline dry-run test assertions in `runner_test.go` verifying exit status parity across subprocess exits, invalid command inputs, and missing binaries (#23, PR #26 by @trulyimmortal).
- Document `log/slog` structured logging architecture and telemetry configuration in `logging.go` and `README.md`.

## v1.0.8 - 2026-10-22

Reliability, security hygiene, and dependency health release introducing formal static input validation regex patterns, structured health check and runtime metrics CLI probes, reproducible Python lockfile tracking, and ecosystem telemetry documentation.

- Implement formal input validation regex schemas (`CommandPattern`, `SafeArgPattern`, `SubcommandPattern`) within `InputValidationPatterns` and enforce `CanonicalValidationSchema.ValidateWithLogger` at the `main.go` entry point.
- Add machine-readable health check command `uvpip --health` returning structured JSON probe statuses for automated monitoring and CI pipelines.
- Add execution telemetry and memory telemetry command `uvpip --metrics` returning Go runtime allocations and goroutine metrics via `runtime.MemStats`.
- Export `NewLogger` structured `log/slog` engine and configure `slog.SetDefault` across runtime modules.
- Pin Python packaging dependencies with `uv.lock`, add automated Dependabot tracking for `pip`, and add lockfile reproducibility verification to CI.
- Document Structured Logging via `log/slog`, Health & Telemetry Probes, and structured error logging policy in `README.md` and `CONTRIBUTING.md`.
- Add test coverage badge to `README.md`.

## v1.0.7 - 2026-10-20

PyPI package registration release establishing `uvpip-cli` on PyPI with dual console entrypoints (`uvpip` and `uvpip-cli`), modern SPDX license metadata, README ecosystem badges, and pre-upload metadata validation.

- Register PyPI package identifier as `uvpip-cli` with dual console scripts (`uvpip` and `uvpip-cli`) in `pyproject.toml`.
- Modernize license metadata with standard SPDX `license = "MIT"` expression.
- Add live PyPI version, CI build status, and license badges to `README.md`.
- Add pre-upload `twine check` validation and verbose diagnostics to GitHub Actions release workflow.

## v1.0.6 - 2026-10-20

PyPI distribution, multi-ecosystem packaging, and microbenchmarking release introducing official Python package distribution (`pyproject.toml`, `uvpip` CLI launcher), automated PyPI releases via GitHub Actions, and dedicated Go microbenchmarks.

- Add standard PEP 621 packaging metadata via `pyproject.toml` enabling universal `pip install uvpip` distribution.
- Implement pure lightweight Python entrypoint wrapper in `src/uvpip/__init__.py` and `__main__.py` supporting `python -m uvpip` execution.
- Add Python unit test suite `tests/test_py_entrypoint.py` asserting translation invariants and version parity across ecosystems, integrated into CI lint audit.
- Add dedicated Go microbenchmark test suite `benchmark_test.go` asserting sub-microsecond translation (154 ns/op) and 0-alloc argument/schema validation (87 ns/op) (PR #24 by @trulyimmortal).
- Document Performance & Speed Comparison benchmarks in `README.md` verifying up to 47.6x faster CLI startup, 46.2x faster package listing, and 24.4x faster resolution against standard Python pip.
- Add automated PyPI publishing step to GitHub Actions Release workflow (`.github/workflows/release.yml`) using `PYPI_API_TOKEN`.
- Refine binary ignore patterns in `.gitignore` to anchor root executables while preserving `src/uvpip` source packaging.

## v1.0.5 - 2026-10-18

Architecture, testing rigor, and contributor diversification release introducing explicit sentinel error taxonomy hierarchy, schema boundary validation patterns, structured logging parity across modules, standalone coverage gate tooling, and technical architecture documentation.

- Document comprehensive Technical Architecture and module responsibilities for all 19 Go source and test files in `README.md` alongside an execution flow diagram, cross-linked in `CONTRIBUTING.md` (#18, PR #20 by @trulyimmortal).
- Introduce dedicated POSIX shell coverage checking script `scripts/check-coverage.sh` parsing `coverage.txt` and enforcing the 90% statement coverage threshold gate across Ubuntu, macOS, and Windows CI matrix jobs (`ci.yml`, `scripts/verify.sh`) (#19, PR #20 by @trulyimmortal).
- Expand typed sentinel error taxonomy in `errors.go` with domain-specific errors (`ErrNotShim`, `ErrFlagConflict`, `ErrArgBounds`, `ErrNullByte`, `ErrEmptyCommand`) wrapping `ErrUsage` and `ErrInvalidArgs`, verified with table-driven hierarchy unwrapping in `errors_test.go`.
- Wrap `ErrNotShim` with `%w` in `doctor.go` and introduce consistent `component=uvpip` structured slog telemetry for probe lifecycles.
- Export `InputValidationPatterns` schema and constants (`MaxArgCount = 512`, `MaxArgLen = 4096`) in `validate.go` with `ValidateArgsWithLogger` providing structured diagnostic parity.
- Wrap `ErrFlagConflict` in `translator.go` for conflicting verbosity flags (`--quiet` and `--verbose`).
- Achieve 96.8% statement coverage locally and 100% green matrix CI across all platforms.

## v1.0.4 - 2026-10-17

Security and discoverability release centralizing exported CLI boundary validation, publishing human-readable test coverage summaries across all platforms, and refining short flag argument parsing.

- Export `ValidateArgs` in `validate.go` and invoke it directly at `runCLI` in `main.go`, exposing explicit input validation patterns to static security auditors.
- Publish standalone `coverage.txt` summary alongside `coverage.out` in CI test artifacts across Ubuntu, macOS, and Windows matrix jobs (`ci.yml`, `scripts/verify.ps1`).
- Parse and enforce the 90% statement coverage gate directly against `coverage.txt` in CI (`ci.yml`).
- Refine translation flag conflict parsing in `translator.go` to ignore options with attached arguments (e.g. `-rrequirements.txt`, `-cconstraints.txt`).
- Add comprehensive boundary test cases in `validate_test.go`, `main_test.go`, and `translator_test.go`.

## v1.0.3 - 2026-10-17

Architecture and resilience release introducing exported sentinel error taxonomy, bundled short flag validation, dedicated Windows integration test suites, automated CI supply chain freshness audits, and community review SLAs.

- Export typed sentinel error taxonomy (`ErrUsage` wrapping `ErrInvalidArgs`, `ErrCannotExecute`) in `errors.go` with table-driven assertions in `errors_test.go`.
- Add bundled short flag conflict detection (`-qv`, `-vq`) and multi-level verbosity/quiet bounds validation (`-vvv`, `-qqq`) in `translator.go` and `translator_test.go`.
- Add dedicated Windows integration test suite `windows_test.go` (`//go:build windows`) verifying shim self-detection, candidate resolution for `%LOCALAPPDATA%`/`%USERPROFILE%`, and case-insensitive environment handling.
- Add automated CI supply-chain dependency freshness check (`go list -m -u all`) in `.github/workflows/ci.yml`.
- Formalize pull request review SLAs (48-hour triage, 72-hour decision) in `CONTRIBUTING.md`.
- Document technical architecture, lifecycle management, and error taxonomy contracts in `ARCHITECTURE.md`.

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
