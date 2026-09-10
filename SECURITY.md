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

The revised source described in README is unreleased. Published v1.0.0 binaries
do not acquire source fixes when the branch changes. No maintained backport
schedule or security-support lifetime is promised. Check the changelog and exact
release provenance before deployment; use reviewed source builds when you need
the unreleased behavior.

## Trust Boundaries

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
