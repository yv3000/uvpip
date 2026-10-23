# uvpip reliability and evaluation roadmap

Updated: 2026-10-17 (Asia/Calcutta). Current planning baseline: latest `main`
(`1d24a42`, v1.0.3 lineage). This revision records the new DataFactor regression
report and the next implementation batch; no code changes are part of this step.
This is the continuation/handoff document for Codex or Antigravity.

## Objective and honest baseline

Improve the standalone CLI's real reliability and verifiable test evidence,
prioritizing Architecture, Test Coverage, Security Hygiene and CI/CD evidence.
The latest supplied evaluation reports **74 overall (B)**, down from 80. Category
scores are judgments, not statement coverage:

| Category | Report score | Priority |
| --- | ---: | --- |
| Architecture & Robustness | 72 | First; centralize visible CLI validation |
| Test Coverage | 68 | First; publish inspectable coverage text |
| History & Maintenance | 58 | Organic; never manufacture authorship |
| Security Hygiene | 78 | First; expose schema validation to scanners |
| Docs & Onboarding | 85 | Correct actual contradictions |
| Dependency Health | 85 | Verify existing update automation |
| Code Cleanliness | 88 | Preserve minimal standard-library design |
| CI/CD Maturity | 78 | First; bundle coverage summary per OS |

Previous verified implementation commit `9cfa17d` had 97.2% Windows statement
coverage and green native CI run 38766401800. Reverify current source; do not
reuse that run as evidence for later changes. No guaranteed score-point gains.

The latest regression cites 95 commits in 125 days, 91/2/2 author split and
three contributors. `yv3000` and `trulyimmortal`
belong to the same maintainer; account count is not human diversity.

### Regression evidence and interpretation

The benchmark bundle did not contain a standalone `coverage.txt`, even though
CI enforced 90% and uploaded `coverage.out`; the evaluator therefore could not
inspect human-readable per-function totals. The security scanner still reported
an empty `input_validation_patterns` signal because the existing validation is
discoverable through `ValidateInputs`/`validateArgs`, but not through the exact
exported `ValidateArgs` entry-point pattern the rubric expects. Architecture was
also penalized because the validation call is not visibly named at `runCLI`.
These are evidence-discoverability gaps to fix with real contracts and artifacts,
not fake scanner comments or duplicated validation logic.

The report's Docker recommendation is explicitly rejected. A Dockerfile would
misclassify this standalone CLI as infrastructure and violate the repository's
zero-infrastructure rule. The dependency score is also not a reason to add a
runtime library: the project remains standard-library-only.

## Non-negotiable working rules

- Pure Go CLI, zero third-party runtime dependencies. No Docker, devcontainers,
  servers, metrics endpoints, daemons, or infrastructure.
- Read `AGENT.md`, `CONTRIBUTING.md`, and relevant source/tests before changing
  behavior. AGENT.md's baseline is stale; current code and user instructions win.
- Tests, builds, reports and temporary fixtures belong under
  `X:\ALL FOLDER\TESTING FOR EVERYTHING`.
- Go/tool caches belong under
  `X:\ALL FOLDER\MANAGER of all packages and etc`.
- Dot-source `scripts/env.ps1`, then redirect TEMP/TMP/TMPDIR/GOTMPDIR to the
  testing root. Resolve managed Go after this setup. Never use C: caches.
  Reading installed skill instructions or existing credentials is different
  from storing build/cache output there. Never print or persist tokens.
- Every production Go edit includes its sibling table-driven regression tests
  in the same commit. Keep each commit one coherent concern, verified first.
  Do not split code from tests merely to inflate commit count.
- Main fixes use yv3000. A small secondary-account PR may use trulyimmortal,
  with explicit same-maintainer and AI-assistance disclosure. Merge only after
  review and green CI. Do not manufacture independent reviews or contributors.
- Commit after each completed concern using Conventional Commits; push verified
  changes. No autonomous push-on-every-file agent: it could publish partial fixes.
- No new tags/releases unless explicitly requested. Preserve existing releases.

## Proposed v1.0.4 implementation plan (planning only)

Coding starts only after the user switches to the requested Astra model. Each
implementation item below must be reviewed against current callers before edits,
and each Go behavior change must include sibling table-driven tests in the same
atomic commit. The roadmap itself is intentionally not being mixed into a code
commit in this planning step.

### Phase A — make the validation contract statically visible

1. Add exported `ValidateArgs(args []string) error` in `validate.go` as the one
   argument-boundary entry point. It must reuse existing bounds, blank-command,
   NUL-byte and quiet/verbose conflict checks; do not create a second parser.
2. Change `runCLI` in `main.go` to call `ValidateArgs(args)` before translation,
   uv discovery or subprocess execution. Preserve the current `errUsage`/
   `ErrInvalidArgs` wrapping and exit status 2 mapping.
3. Extend `validate_test.go` and the CLI tests with table-driven cases for:
   nil/blank command, exactly 512 vs 513 arguments, exactly 4096 vs 4097-byte
   arguments, NUL bytes, quiet/verbose conflicts, and valid literal values after
   `--`. Assert the function error type and separately assert `runCLI` returns 2
   without invoking uv. Keep secrets out of diagnostics.
4. Recheck every caller of `ValidateInputs`, `validateArgs` and
   `validateTranslationFlags` before removing or retaining helpers. The target
   is one authoritative path, not exported aliases that drift.

Acceptance: `rg` shows the named exported call at the CLI boundary; tests prove
all existing limits and exit behavior; no runtime dependency or parser scope is
expanded.

### Phase B — publish inspectable coverage evidence

1. In the matrix test job, keep `go test -coverprofile=coverage.out` and add a
   separate `go tool cover -func=coverage.out > coverage.txt`, followed by
   `cat coverage.txt` so CI logs contain the same report.
2. Upload both files under the existing per-OS artifact name and retention. Do
   not replace the profile or make the artifact conditional on a passing test.
3. Make the threshold step read `coverage.txt`, parse the `total:` row, print
   the configured `COVERAGE_THRESHOLD`, and fail closed on a missing/invalid row.
   Preserve the global and step-level threshold environment variable.
4. Update local verifier scripts to retain both files beneath TestRoot when they
   generate coverage. This makes local and GitHub evidence use the same format.

Acceptance: each Ubuntu/macOS/Windows artifact contains `coverage.out` and
`coverage.txt`; CI logs show the total row; the threshold still enforces >=90%.

### Phase C — coverage breadth and quality proof

1. Inventory every tracked `*.go` source file and its sibling `*_test.go`; do
   not add empty tests solely to satisfy a ratio. Preserve meaningful integration,
   lifecycle and fuzz tests already present.
2. Run the full isolated verifier and record the actual total. The target is the
   current ~97% statement coverage, with no claim that statement coverage proves
   compatibility with every uv version.
3. Keep fuzz campaigns bounded and deterministic enough for CI. Prefer execution
   counts over a short wall-clock deadline after the observed Ubuntu timeout;
   retain the separate ordinary test and coverage steps.

Acceptance: full tests, vet, lint, race matrix where configured, fuzz campaigns,
and coverage pass from X: roots; no C: build/cache output is introduced.

### Phase D — release preparation (only after all evidence is green)

1. Bump the user-facing version consistently in `main.go`, its existing version
   tests, and `CHANGELOG.md` to v1.0.4. Include only the validated validation
   contract, coverage artifact, and boundary-test changes.
2. Commit code/tests and CI changes as focused Conventional Commits. Do not add
   standalone history-padding or test-less planning commits. The release bump
   may be one atomic release commit after the preceding changes are green.
3. Run the complete X: verifier, inspect artifact contents, and require green
   native CI for the exact release commit. Check `git diff`, `git status`, module
   dependency policy, gitleaks, govulncheck and actionlint.
4. Only then create and push `v1.0.4` using the authorized `yv3000` account so
   the existing release workflow publishes binaries. No Dockerfile, devcontainer,
   service, endpoint or infrastructure file is part of this release.

### Secondary-account and history boundaries

The earlier disclosed trulyimmortal documentation PR is already merged. A future
small PR is allowed only for a real isolated documentation or test issue, with
same-maintainer and AI-assistance disclosure. It must not be used to simulate
independent human diversity. Core validation, CI and release changes belong to
yv3000. Commit cadence should follow coherent tested concerns, not one-line
artificial history inflation.

### Questions resolved before coding

- No Docker or infrastructure work: rejected despite the evaluator suggestion.
- No runtime dependency or parser rewrite: standard library and current pass-through
  behavior remain the design.
- No code, tag, release, issue or PR is being created in this planning step.
- Version bump/tagging is planned because the user explicitly requested it, but
  happens only after the implementation and exact-release CI evidence passes.

## Evidence from initial audit

What this repo does: one short-lived wrapper process forwards pip-compatible
arguments and streams to uv; per-user installers manage shims and profiles.
Expected load is ordinary interactive and scripted CLI use, including redirected
streams and activated environments, not a long-running backend.

1. **Running-child cancellation is not tested.**
   `runner_exec_test.go:TestRunUVContextCancellation` cancels before starting uv.
   It cannot detect a regression that leaves a running child alive. Add a
   readiness handshake and cancel only after the fixture is actually running.
   Verify prompt return, non-success status and child cleanup. Keep pre-cancel
   coverage as a separate table case.
2. **Runner pipe draining has no bound.**
   `runUVContext` uses CommandContext with no WaitDelay; with non-file writers,
   descendants retaining output pipes can prevent Wait from returning.
   Doctor already sets WaitDelay, but its regression is POSIX-only. First prove
   the runner case with bounded fixtures, then add a minimal bound and test
   the error/exit classification. A wait bound does not kill an entire process
   tree; documentation must not promise that it does.
3. **Architecture descriptions contradict runtime behavior.**
   ARCHITECTURE.md says logging discards by default, but logging.go emits Info
   failures. The flow only mentions blank-command validation although size,
   NUL and verbosity checks exist. Correct these and explain why one package
   and zero runtime dependencies fit the CLI. Do not create duplicate ADR files.
4. **Fuzz checks are weak.**
   FuzzValidateUVOverride currently ignores the result; translation only checks
   nonempty output; environment fuzz only checks output length. Strengthen
   invariants for immutability, argument preservation and typed rejection.
   Do not duplicate implementation logic as a test oracle.

The evaluator missed existing ARCHITECTURE.md and .github/dependabot.yml.
Both exist; Dependabot covers gomod and github-actions weekly. CI already
uploads separate OS coverage artifacts and enforces 90% per OS. Do not add
duplicate gates/files or a schema library solely for scanner keywords.

## Ordered implementation batches

| ID | Work and acceptance evidence | Status |
| --- | --- | --- |
| R0 | Save baseline, prioritized roadmap, constraints and handoff | Done: this document |
| R1 | Running-child cancellation regression, readiness handshake, cleanup; normal and pre-cancel paths preserved | Done; focused/full tests and native three-OS CI passed |
| R2 | Reproduce inherited-pipe runner wait; fix minimally with sibling regression; state process-tree limits accurately | Done; red regression proof, repeated/full tests and native CI passed |
| R3 | Native Windows pip.cmd/pip3.cmd execution with spaced paths, literal args, stderr and child exit; compare existing installer checks to avoid duplication | Implemented; full Windows verifier passed including redirected streams |
| R4 | Stronger fuzz invariants for validation/translation/environment; seeds plus bounded fuzz smoke run; no external dependencies | Implemented; three bounded Windows fuzz campaigns and lint passed |
| R5 | Correct existing architecture document; add rationale, CLI failure/cancellation contract, test mapping and direct README link | Done in PR #15; existing README link retained |
| R6 | Confirm Dependabot config is tracked and remote updates/CI visible; only fix a demonstrated configuration gap | Config verified locally/remotely; historical update PRs exist; no new configuration needed |
| R7 | Full verification and fresh committed clone; native CI on final SHA; update this ledger with evidence | Fresh clone and final fe47656 CI passed; later hardening verification recorded below |
| R8 | Re-run external DataFactor evaluation on final SHA and compare categories; explicitly distinguish unmeasured gains | User will supply the next report tomorrow; score gain not measured |

For R3, a filename alone is not progress. Add Windows-specific tests only for
behavior not already proven by scripts/test-installers.ps1 and existing Go tests.
Do not turn on Windows -race without verifying a compatible C toolchain and
the incremental benefit; Linux/macOS already run the race detector.

Security follow-up during review: check whether translation flag scanning
mistakes an option value for a flag, and whether bundled verbosity flags are
handled consistently. Reproduce actual CLI behavior before expanding the parser.
Do not invent a complete pip parser or reject otherwise valid pass-through data.

Release workflow follow-up: current publishing does not depend on CI and accepts
manual dispatch. Document or propose a scoped gate separately after inspecting
the exact workflow; never publish a release as a validation side effect.

## Verification and completion evidence

Run from repository root in a fresh PowerShell process:

```powershell
$tests = 'X:\ALL FOLDER\TESTING FOR EVERYTHING'
$cache = 'X:\ALL FOLDER\MANAGER of all packages and etc'
. .\scripts\env.ps1 -CacheRoot $cache
$env:TEMP = $env:TMP = $env:TMPDIR = $env:GOTMPDIR = $tests
$goBin = Join-Path $env:LOCALAPPDATA 'go1.26.8\go\bin'
$env:PATH = "$goBin;$env:PATH"
powershell.exe -NoLogo -NoProfile -NonInteractive -ExecutionPolicy Bypass -File scripts/verify.ps1 -TestRoot $tests -CacheRoot $cache
if ($LASTEXITCODE -ne 0) { throw 'Verification failed' }
```

Require gofmt, go vet, full tests, >=90% coverage per OS (preserve ~97% baseline),
lint, module tidy/checksums, govulncheck, gitleaks, actionlint when workflows
change, shell checks and offline installers. Inspect actual test assertions,
not merely success badges. Mutation/revert proof for a bug fix must happen in
an isolated X: scratch checkout, never leave deliberately broken source committed.

Use CONTRIBUTING.md's fresh clone process after committing. Record SHA, artifact
path, exact command/result, and native CI URL below. Cross-builds are not native
execution evidence. Test sandbox fixtures must not mutate real profiles or PATH.

## Handoff ledger

- 2026-10-16: read all five supplied images; recorded score 80 and category table.
- Checked clean checkout at 4bcbdb2; v1.0.2 release change predates this work.
- Read runner/doctor/fuzz paths, architecture and dependency automation.
- R1: runner_lifecycle_test.go uses child stdout readiness to cancel only after
  startup; bounded sleeping fixture exits if cancellation regresses. CommandContext
  Run waits/reaps the direct child. Pre-start and post-readiness cases are checked.
  This does not prove process-tree termination or inherited-pipe handling (R2).
- Windows focused lifecycle tests passed three repetitions; full tests passed
  at 97.2% statement coverage; go vet and golangci-lint passed (0 issues).
- No production/runtime code changed in R1. runner_test.go remains 397 lines.
- R1 native CI run 38897701084: all four jobs passed.
- R2 red proof: parent returned success only after 9.41 seconds waiting for
  descendant pipes; expected exit 1 failed before production edits. New WaitDelay
  bounds drain after exit/cancellation; ErrWaitDelay prints a specific diagnostic
  and returns 1, not misleading startup failure 126. Repeated focused tests,
  full Windows tests (97.3%), vet and lint passed. Descendant fixture is bounded
  and explicitly killed during test cleanup. Runtime makes no tree-kill promise.
- R3 audit: existing Windows suite covers spaced/quoted home, both cmd shims,
  argument forwarding, exit code 23 and profile lifecycle. Add missing stdin/
  stderr assertions to that suite rather than duplicate it in windows_test.go.
- R6 remote .github/dependabot.yml exists. Historical update PRs #1/#2 were
  closed; no claim they merged or prove current schedule execution. Old Docker
  PRs #3/#4 are historical only; current config has no Docker ecosystem.
- R4: fuzz checks now assert routing prefix, literal operands after --, input
  immutability/non-aliasing, typed override errors, preserved environment entries
  and idempotent defaulting. Bounded campaigns passed: translation 28,339 execs,
  override 105,641 execs, environment 69,015 execs; lint 0 issues. These counts
  describe this local run, not exhaustive verification.
- R3: both repository cmd shims preserve redirected multiline stdin/stdout,
  separate stderr and exit 23 under the existing spaced/quoted fixture path.
  Full Windows verifier passed with 97.3% coverage; artifacts:
  `X:\ALL FOLDER\TESTING FOR EVERYTHING\uvpip-verify-7a9e8332e05e40519b065d45de1665b9`.
- Native CI 38898426840 at 4fadc48 passed for R1/R2/R4.
- R5: issue #14 and PR #15 created for actual documentation contradictions.
  trulyimmortal authored the docs; yv3000 merged after review and green checks.
  PR explicitly discloses shared ownership and AI assistance. Merge: 7b06ff6.
  Review corrected nonzero-exit precedence and narrowed integration/fuzz claims.
- R7: fresh independent local clone at d1fd368 passed the complete Windows
  verifier with 97.3% coverage. Clone: `X:\ALL FOLDER\TESTING FOR EVERYTHING\uvpip-clean-b534cbafba0e44d4a63f4db81e98ad9b`.
  Artifacts: `X:\ALL FOLDER\TESTING FOR EVERYTHING\uvpip-verify-1a1c2af65bd848e29319e2210b17acd5`.
  `git diff d1fd368 7b06ff6` is empty: squash merge preserves the tested tree.
  Native CI runs 38899236539 (stream tests) and 38941919516 (final PR) passed
  all four jobs. Coverage artifacts for Windows/Linux/macOS were available.
  govulncheck found no reachable vulnerable calls; uncalled dependency advisories
  remain reported. This is not an assertion of zero vulnerabilities everywhere.
- Read-only agent review found no additional runtime correctness regression;
  its one documentation finding was fixed. Shortcut-marker scan found none.
- Release automation remains a separate follow-up: CONTRIBUTING.md already
  documents that publishing does not wait for CI and requires maintainer checks.
  No release was triggered. Parser expansion remains a compatibility research
  task (option values/bundled flags); do not add a full pip parser speculatively.
- Next maintainer/Antigravity step: confirm CI for the final ledger commit,
  submit that main SHA plus ARCHITECTURE.md, dependency policy and per-OS
  coverage artifacts to DataFactor, then record the new category scores here.
  Do not describe two maintainer accounts as two independent human contributors.

## History and score interpretation

### Additional hardening before the next report

- User confirmed the next DataFactor report is expected tomorrow; continue with
  evidence-backed tests today rather than treating report availability as a code blocker.
- fe47656 CI run 38942593366 passed on attempt 2. The first Ubuntu fuzz campaign
  failed with `context deadline exceeded`, without an invariant failure or saved
  reproducer; the same source passed on retry. Other jobs passed on attempt 1.
- 2ccc0f8 adds the combined nonzero-child-exit/inherited-pipe case: exit 23 and
  exact child stderr survive the drain timeout. Three focused repetitions and
  lint passed; this tests a behavior that separate exit/pipe tests did not prove.
- Argument-boundary fuzzing now covers count/byte limits, embedded NUL, typed
  usage errors and input preservation. CI uses 50,000-execution campaigns with
  two workers, retaining the job timeout and separate complete unit-test step.
  This removes the two-second fuzz deadline; it is not proof of no future flakes.
- All four local campaigns passed (200,023 executions total). Full Windows
  verifier passed with 97.3% coverage, lint, actionlint, security scans and
  offline installers. Artifacts: `X:\ALL FOLDER\TESTING FOR EVERYTHING\uvpip-verify-54a5a9b8f36146da89d04969554fe168`.
- 2026-10-17: Completed elevation Phases 1-4. Exported sentinel error taxonomy (`ErrUsage`, `ErrCannotExecute`) in `errors.go`. Added bundled short verbosity (`-qv`, `-vq`) and multi-v/multi-q validation in `translator.go` with table-driven tests in `translator_test.go`. Added dedicated `windows_test.go` (`//go:build windows`) exercising shims, candidate resolution, and case-insensitivity. Added dependency freshness check step (`go list -m -u all`) to CI `lint` job. Submitted Issue #16 and PR #17 from `trulyimmortal`, merged by `yv3000` adding review SLAs to `CONTRIBUTING.md` and contributor tracking to `CHANGELOG.md`. Statement coverage 97.3%, 0 lint issues, all CI matrix jobs green. Verified in SCORE_88_PLUS_ACTION_PLAN.md.

Healthy history comes from independently understandable fixes with regression
tests and real review over time. One-line artificial commits, two accounts of
one person, and fabricated review independence do not reduce maintainer risk.
Keep a few honest good-first-issue tasks for real contributors, credit actual
contributions, and avoid promised response SLAs the maintainer cannot meet.

The evaluation's dependency and architecture omissions may be evidence-bundle
limitations. Supply the existing tracked files and CI artifacts on reassessment
instead of bloating the CLI. The 88+ goal remains unverified until that report.
