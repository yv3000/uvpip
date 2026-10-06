# Architecture

uvpip is one Go `main` package with no runtime dependencies, plus shell
installers. A `pip ...` call flows through it like this:

```text
pip / pip3 shim or shell function
  -> main.go        runCLI: local commands (--help, --version, doctor) or forward
  -> validate.go    reject blank commands (exit 2) before anything runs
  -> translator.go  pip args -> uv args ("uv pip ...", "uv cache ...")
  -> runner.go      find uv, build child env, exec uv, map exit code
  -> uv
```

| File | Role |
| --- | --- |
| `main.go` | Entry point and dispatch. No business logic. |
| `validate.go` | Boundary checks for CLI args and the `UVPIP_UV` override. Never echoes values. |
| `translator.go` | Pure function: pip argv to uv argv. `uninstall -y` stripping, `upgrade` alias, `cache`. |
| `runner.go` | uv discovery (`UVPIP_UV`, then PATH, then per-user dirs; rejects `ErrDot` and self), `UV_SYSTEM_PYTHON` defaulting, exec with untouched streams, exit codes 126/127. |
| `doctor.go` | `uvpip doctor`: probes uv/python versions and recognizes shipped pip shims. |
| `logging.go` | Opt-in `UVPIP_DEBUG` slog handler; discards by default. |
| `bin/` | Repository pip/pip3 shims (`.sh`, `.cmd`). |
| `installer/`, `uninstaller/` | Per-user install/uninstall for POSIX `sh` and PowerShell; exact marker-block profile edits. |
| `scripts/` | Isolated dev environment, verifiers, and offline installer test suites. |

## Design rules

- No shell at runtime: uv gets argv boundaries directly.
- No network at runtime: uvpip never downloads; only the explicit installer may.
- Fail closed: invalid `UVPIP_UV` or malformed profile markers error out instead
  of guessing.
- Secrets stay out of logs: debug records carry counts and paths, never argument
  values or environment contents.

## Tests

Each `*.go` file has a `*_test.go` sibling, and `integration_test.go` exercises
the complete install-doctor-uninstall lifecycle. `runner_test.go`'s `TestMain` turns
the test binary into an offline fake uv (`fixtureUV`), so subprocess behavior is
tested without uv or network. Installer suites in `scripts/` mock `curl`, `uname`,
and uv. CI enforces a 90% statement-coverage floor; see [CONTRIBUTING.md](CONTRIBUTING.md).
