# uvpip

Type `pip`, run [uv](https://github.com/astral-sh/uv).

uvpip is a small Go wrapper that translates common pip commands to uv. Optional
per-user `pip` and `pip3` shims and shell functions let you keep familiar commands
without replacing Python's pip installation. Speed and compatibility depend on
uv, your packages, network, and environment; uvpip is not a complete pip emulator.

> **Source versus releases:** This README describes the revised, unreleased
> source. The published [v1.0.0 release](https://github.com/yv3000/uvpip/releases/tag/v1.0.0)
> predates these changes. Running a revised installer without a local binary
> still downloads the latest **published** binary, not a build of this source.
> Use the local-build route below to get the revised runtime. The source still
> reports version `1.0.0`; record the source commit, not just the version string.

## Requirements

- A working [uv installation](https://docs.astral.sh/uv/getting-started/installation/)
  and a Python interpreter supported by that version of uv.
- For source builds and project verification: Go **1.26.8** (module language
  minimum **1.26.0**). No third-party Go runtime dependencies are required.
- Installers target Windows, macOS, and Linux on amd64 and arm64. Native
  Linux/macOS verification is pending CI on the pushed revised source; Git Bash
  tests are not native Linux/macOS execution.
- Windows installation uses PowerShell. Windows installer tests explicitly use
  Windows PowerShell **5.1**. POSIX installation uses `sh` and standard utilities;
  downloading requires `curl` or `wget`.

uvpip itself does not install Python. The explicit installer can install uv if
it is absent; the revised runtime never downloads or installs anything on its own.
Package operations delegated to uv can, of course, access the network and modify
the selected Python environment.

## Install From Source

Review a checkout before executing its scripts. Build with Go 1.26.8 from the
repository root, and choose an output path outside the checkout. Do not overwrite
an existing local binary. Install uv separately first if you want to avoid the
installer downloading and executing the official uv setup script.

### Windows

In a normal PowerShell session, with `$buildDir` set to an existing directory
reserved for build output:

```powershell
$binary = Join-Path $buildDir 'uvpip-source.exe'
go build -o $binary ./...
if ($LASTEXITCODE -ne 0) { throw 'Build failed' }
$hash = (Get-FileHash -LiteralPath $binary -Algorithm SHA256).Hash
& .\installer\install.ps1 -BinaryPath $binary -SHA256 $hash
```

This installs into `$env:USERPROFILE\.uvpip\bin`, adds `pip.cmd`/`pip3.cmd`,
adds functions to the current PowerShell host's `$PROFILE`, and updates **User
PATH**. It never elevates or edits **System PATH**. Different PowerShell hosts
can have different profiles. Use `-NoProfile` and/or `-NoPath` to opt out of the
corresponding integration. `UVPIP_BINARY` and `UVPIP_SHA256` supply defaults for
`-BinaryPath` and `-SHA256`.

### macOS / Linux

With `BUILD_DIR` set to an existing directory reserved for build output:

```sh
go build -o "$BUILD_DIR/uvpip-source" ./... &&
UVPIP_BINARY="$BUILD_DIR/uvpip-source" sh installer/install.sh
```

This installs into `$HOME/.uvpip/bin`. Based on `$SHELL`, it adds a marked block
to bash, zsh, fish, or `.profile` configuration. The block sets PATH and defines
`pip`/`pip3` functions. Set `UVPIP_NO_PROFILE=1` to skip profile integration and
manage PATH yourself. Set `UVPIP_SHA256` to an independently checked SHA-256
digest to validate a new candidate binary before execution.

### Installation Notes

- Restart your terminal after install or uninstall. Existing sessions can retain
  functions or cached command locations. No blanket promise of shell interception
  applies to hosts that do not load the modified profile.
- Installers check a candidate's `--version`, use private staging directories,
  and can validate a caller-supplied checksum. They do **not** authenticate release
  signatures. Hashing a local build detects a changed copy, not malicious source.
- Without a local binary, installers download the platform asset from
  `releases/latest/download`. That URL is mutable. Review the exact release and
  obtain a trusted checksum, or build reviewed source yourself.
- If uv is absent from command lookup, explicit installation downloads its
  official installer fully before executing it. A local uvpip binary alone does
  not make installation offline; uv must already be available.
- Re-running installation preserves an existing uvpip executable. It is not an
  automatic upgrade and a supplied checksum is not rechecked against an existing
  install. To change versions, preserve anything you need, uninstall, then install
  the reviewed replacement.
- Development environment helpers redirect HOME, caches, and temporary paths.
  **Do not source `scripts/env.ps1` in a live installation session.** Use a fresh
  normal shell for a real install, or deliberately use a sandbox for testing.

## Usage

```sh
uvpip --version
uvpip install requests
uvpip install -r requirements.txt
uvpip uninstall requests
uvpip list
uvpip freeze
uvpip doctor
```

After shell integration, substitute `pip` or `pip3` for `uvpip`.

### Command Translation

- Most arguments are forwarded as `uv pip <arguments>`, including `install`,
  `list`, `show`, `freeze`, and `check`.
- `uninstall` and the `remove` alias become `uv pip uninstall`. `-y` and `--yes`
  are stripped before `--` because uv does not use pip's confirmation prompt.
  Arguments after `--` are preserved.
- `upgrade <arguments>` is a uvpip convenience alias for
  `uv pip install --upgrade <arguments>`, not a standard pip command.
- `cache <arguments>` becomes `uv cache <arguments>`. Forwarding `cache list`
  does not guarantee your uv version implements it.
- No arguments, leading `--help`/`-h`, leading `--version`/`-V`, and `doctor` are
  handled locally. `-v` is forwarded to uv, not treated as wrapper version.
- Unknown commands and flags go to uv. Unsupported operations fail with uv's
  error; uvpip never retries with real pip. See uv's
  [pip compatibility documentation](https://docs.astral.sh/uv/pip/compatibility/).

The child process receives argument boundaries directly, without a runtime shell,
and inherits stdin, stdout, and stderr without output rewriting. A normal uv exit
code is returned unchanged. Failure to locate uv returns **127**; failure to start
the resolved executable returns **126**. Signal termination has no portable child
exit code and returns **1**.

### Environment Selection

uvpip preserves the user environment, including explicit `UV_SYSTEM_PYTHON`.
When neither `VIRTUAL_ENV` nor `CONDA_PREFIX` is nonempty and there is no
`UV_SYSTEM_PYTHON` entry, it defaults `UV_SYSTEM_PYTHON=1`. Activated virtual and
Conda environments therefore follow uv's selection unless explicitly overridden.
uvpip does not translate `PIP_PYTHON` into a uv interpreter override. Use
uv-supported interpreter options when exact selection matters.

`UVPIP_UV` optionally selects an **absolute** executable path. Otherwise uvpip
searches PATH, then known per-user/platform installation locations. It refuses
Go's `ErrDot` current-directory lookup and excludes relative fallback candidates.
It rejects uv resolving to its own executable using `os.SameFile`, including
hardlinks or symlinks to that same file, but not separate identical copies.
An invalid explicit override fails rather than falling back to a different uv.
Only point the override and PATH at trusted executables.

Shell functions can keep intercepting `pip` after virtual environment activation
when the profile was loaded. Noninteractive processes, aliases, other profiles,
and venv PATH precedence can bypass them. Call `uvpip` explicitly in automation.
To deliberately bypass the wrapper, use the desired interpreter's pip:

```sh
python -m pip install requests
```

### Diagnostics

`uvpip doctor` checks uv, pip, pip3, and Python (with `python3` fallback), recognizes
the shipped text shims, and returns **1** on a failed check. uv/Python version
probes have a five-second timeout. This is an executable/PATH check, not proof
that a parent shell function or alias is active; a working interactive function
can coexist with a failed PATH-shim check. Inspect `Get-Command pip -All` in
PowerShell or `type pip` in your interactive POSIX shell when investigating that.

## Uninstall

Run the reviewed script from the same checkout in a normal shell:

```powershell
& .\uninstaller\uninstall.ps1
```

```sh
sh uninstaller/uninstall.sh
```

Uninstall removes known uvpip binaries/shims and complete, exactly marked profile
blocks, while retaining unrelated files. Windows removes matching User PATH
entries, not System PATH. Legacy System PATH entries require manual review.
Malformed profile markers cause failure rather than guessed deletion; repair
them manually and retry. Windows accepts `-NoProfile`/`-NoPath`; POSIX accepts
`UVPIP_NO_PROFILE=1`. Python's pip and uv are not uninstalled. Restart the shell.

## Development And Security

See [CONTRIBUTING.md](CONTRIBUTING.md) for required fresh-clone verification,
formatting/tests, pinned tools, the coverage gate, and the manual release checklist.
CI does not publish binaries or use release secrets.

See [SECURITY.md](SECURITY.md) for the trust model and reporting guidance, and
[CHANGELOG.md](CHANGELOG.md) for released versus unreleased changes.

## License

[MIT](LICENSE), copyright (c) 2026 YV.
