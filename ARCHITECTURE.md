# Architecture

uvpip is one Go `main` package with no runtime dependencies, plus shell
installers. A `pip ...` call flows through it like this:

```text
pip / pip3 shim or shell function
  -> main.go        runCLI: local commands (--help, --version, doctor) or forward
  -> validate.go    reject blank/oversized/NUL args and conflicting verbosity (exit 2)
  -> translator.go  pip args -> uv args ("uv pip ...", "uv cache ...")
  -> runner.go      find uv, build child env, exec uv, map exit code
  -> uv
```

| File | Role |
| --- | --- |
| `main.go` | Entry point and dispatch. No business logic. |
| `errors.go` | Typed sentinel error taxonomy (`ErrInvalidArgs`, `ErrInvalidOverride`, `ErrUVNotFound`, `ErrRecursiveExecution`, `ErrTranslation`) for structured failure classification. |
| `validate.go` | Boundary schema checks for CLI args (count, length, null bytes) and the `UVPIP_UV` override. Never echoes values. |
| `translator.go` | Validates quiet/verbose conflicts before forwarding; converts pip argv to uv argv. `uninstall -y` stripping, `upgrade` alias, `cache`; respects `--`. |
| `runner.go` | uv discovery (`UVPIP_UV`, then PATH, then per-user dirs; rejects `ErrDot` and self), `UV_SYSTEM_PYTHON` defaulting, exec with untouched streams, exit codes 126/127. |
| `doctor.go` | `uvpip doctor`: probes uv/python versions and recognizes shipped pip shims. |
| `logging.go` | Structured stderr logging: Info failures by default, Debug diagnostics enabled by `UVPIP_DEBUG`; argument values are omitted. |
| `bin/` | Repository pip/pip3 shims (`.sh`, `.cmd`). |
| `installer/`, `uninstaller/` | Per-user install/uninstall for POSIX `sh` and PowerShell; exact marker-block profile edits. |
| `scripts/` | Isolated dev environment, verifiers, and offline installer test suites. |

## Design rules

- One `main` package keeps this small, stateless CLI's dispatch, validation and
  subprocess lifecycle directly reviewable. Files separate responsibilities;
  extra packages or interfaces need an actual reuse or isolation requirement.
  Standard-library process, path and logging APIs keep runtime dependencies zero.
- No shell at runtime: uv gets argv boundaries directly.
- No network at runtime: uvpip never downloads; only the explicit installer may.
- Fail closed: invalid `UVPIP_UV` or malformed profile markers error out instead
  of guessing.
- Typed error taxonomy: failure modes wrap explicit sentinel errors (`errors.go`),
  enabling programmatic inspection and deterministic error unwrapping.
- Explicit schema boundaries: CLI boundaries validate input size and reject
  null-byte injections before argument translation or subprocess execution.
- Secrets stay out of logs: debug records carry counts and paths, never argument
  values or environment contents.

## Subprocess lifecycle

`runUVContext` uses `exec.CommandContext` to cancel the direct uv child. Ordinary
child exit codes propagate; discovery and startup failures map to 127 and 126.
There is no installation-duration timeout. A one-second `WaitDelay` bounds
Go-managed output-pipe draining after child exit or cancellation; if uv exits
successfully but descendants retain those pipes beyond the bound, uvpip reports
the drain failure and returns 1. Nonzero child exit codes are preserved. This does not
terminate a whole process tree or bound a caller's blocked custom writer.
Doctor probes use their own five-second context timeout and one-second drain bound.

## Tests

Each `*.go` file has a `*_test.go` sibling, and `integration_test.go` exercises
a simulated installation layout, doctor/CLI execution and cleanup. `runner_exec_test.go`'s `TestMain` turns
the test binary into an offline fake uv (`fixtureUV`), so subprocess behavior is
tested without uv or network. Installer suites in `scripts/` mock `curl`, `uname`,
and uv. CI enforces a 90% statement-coverage floor; see [CONTRIBUTING.md](CONTRIBUTING.md).

`runner_lifecycle_test.go` checks cancellation after a child readiness handshake
and descendants retaining output pipes, alongside pre-start cancellation and
normal completion. Windows installer tests exercise both cmd shims with spaced
paths, literal arguments, redirected stdin/stdout/stderr and nonzero child exits.
Fuzz targets check typed `ErrInvalidOverride` handling, translation input preservation
and environment-defaulting invariants. Native CI runs on Windows, Linux and macOS, with race
detection on Linux/macOS; these fixtures do not prove all real uv workloads.
