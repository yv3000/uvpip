# Security

## Reporting

Do not post credentials, private index URLs, or a weaponized vulnerability report
in a public issue. If GitHub private vulnerability reporting is enabled for this
repository, use [Report a vulnerability](https://github.com/yv3000/uvpip/security/advisories/new).
Availability has not been confirmed: the unauthenticated endpoint requires login.
If the option is unavailable after signing in, open a minimal public issue asking
the maintainer to enable a private reporting channel, **without exploit details**.
No private email address or response-time guarantee is advertised here.

Include the source commit or release, OS/architecture, Go and uv versions when
relevant, a minimal reproduction, impact, and whether installer/profile changes
are involved. Redact tokens and sensitive paths. Do not attach real environment
dumps or package-index credentials.

## Release Status

Official multi-platform binaries and Python packages are published via [GitHub Releases](https://github.com/yv3000/uvpip/releases) and [PyPI](https://pypi.org/project/uvpip-cli/) for each tagged version (current: **v1.1.0**). Legacy release binaries do not acquire source fixes retrospectively. No maintained backport schedule or long-term security-support lifetime for older versions is promised. Check [CHANGELOG.md](CHANGELOG.md) for exact release provenance and migration notes; use reviewed source builds when you need unreleased behavior directly from `main`.

## Threat Model: Installer Scripts and Trust Boundaries

### Installer Scripts Threat Model

`installer/install.sh` and `installer/install.ps1` automate downloading and setting up uvpip for interactive user shells. Their threat model and security boundaries are defined as follows:

- **Unauthenticated Remote Downloads**: When executed without a pre-existing local binary, installer scripts fetch release assets from GitHub Releases over HTTPS (`--proto https --tlsv1.2`). While transport encryption is enforced, release assets and mutable `releases/latest/download` URLs are not signed with GPG or cryptographic keys.
- **Optional SHA-256 Verification**: Both installers support optional digest verification via `-SHA256 <hash>` (PowerShell) or `UVPIP_SHA256=<hash>` (POSIX sh). When supplied, the script asserts the SHA-256 hash before executing or moving the candidate binary. If the checksum mismatches, installation aborts immediately. Checksums are not automatically pulled from remote manifests; operators must supply an independently validated digest.
- **Recommendation to Build From Source**: For high-assurance, enterprise, or air-gapped installations, operators should not pipe remote installer scripts into a shell. Instead, build uvpip directly from reviewed source using `go build ./...` with pinned dependencies, and supply the resulting executable to the installer via `-BinaryPath` / `UVPIP_BINARY`.
- **Upstream uv Bootstrapping**: If `uv` is not found on PATH or candidate paths, explicit installation downloads and executes Astral's official `uv` installer. That upstream bootstrap script has its own external security boundaries and network requirements.
- **Automated Offline Invariant Verification**: CI asserts in `scripts/test-installers.sh` that running `install.sh` with `UVPIP_BINARY` and `UVPIP_NO_PROFILE=1` succeeds purely offline without invoking network utilities (`curl`/`wget` are intercepted with fatal trap binaries), printing `offline install verified: no network calls`.
- **Input Validation Boundary Patterns**: Canonical regex patterns are exported via `input_validation_patterns` in both Go (`validate.go`) and Python (`src/uvpip/__init__.py`), enforcing schema validation against null-byte injection (`^[^\x00]+$`) and validating command token structure (`^[a-zA-Z0-9_\-]+$`) across all CLI boundaries.

### Runtime Trust Boundaries

- The revised runtime starts a resolved uv executable with an argument array,
  not a shell. It does not download uv, elevate privileges, or retry failures
  using Python's pip. Missing uv fails with exit 127.
- PATH, known uv installation locations, and an absolute `UVPIP_UV` override are
  trusted inputs. Keep them under trusted ownership. The runtime refuses
  current-directory `ErrDot` resolution and excludes relative fallback candidates.
  Neither check is a sandbox or executable-signature verification.
- The runtime uses `os.SameFile` to reject uv resolving to its own executable,
  including hardlinks or symlinks to the same underlying file. It does not compare
  contents: a separate byte-identical copy is not rejected by this guard. This
  limits same-file recursion, not all possible wrapper loops or malicious code.
- The child inherits the user's environment and streams. uv remains responsible
  for package resolution, network access, interpreter selection, package builds,
  and installation. Untrusted packages and build backends can execute code.
- Installers deliberately execute the candidate binary for `--version`. Inspect
  source/assets before use. Optional `UVPIP_SHA256`/`-SHA256` checks happen before
  a new candidate is executed; the checksum must come from a trusted source.
  Existing installed binaries are not revalidated against a newly supplied hash.
- Default downloads use HTTPS and a mutable latest-release URL. Installers do not
  verify signatures or automatically retrieve a trusted checksum manifest. If uv
  is missing, explicit setup downloads and runs Astral's official installer. That
  upstream installer has its own behavior and trust requirements.
- Integration writes per-user files/profile blocks and, on Windows, User PATH.
  It never requests elevation or edits System PATH. Legacy System PATH entries
  are reported for manual review. Profile markers are an ownership convention,
  not a security boundary against another process running as the same user.
- Uninstall removes only known artifacts and complete exact profile blocks.
  Malformed markers fail closed rather than guessing. Back up valuable profiles;
  installation/uninstallation is not a transactional filesystem operation.

Use `-NoProfile -NoPath` on Windows or `UVPIP_NO_PROFILE=1` on POSIX for deliberate
sandbox integration, with test HOME/profile/temp paths. Do not run live installers
as a substitute for the offline tests. CI runs with read-only permissions, no
release secrets, SHA-pinned Actions, and dependency vulnerability checks.
CI also runs a pinned, redacted Gitleaks scan of the current source tree.
`govulncheck` is not a secret scanner and does not audit uv or Python packages.
uvpip's runtime has no third-party Go module dependencies; that does not remove risks
in the Go toolchain, external tools, installers, or delegated package operations.
Review diffs for secrets before sharing or committing them; never rely on an
ignored file or a clean scan to protect a credential that has already leaked.
