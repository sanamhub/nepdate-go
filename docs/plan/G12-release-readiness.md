# G12: Release readiness without a release

Read first: ADR-0005, ADR-0008, [`docs/runbooks/release.md`](../runbooks/release.md),
[`CHANGELOG.md`](../../CHANGELOG.md) and [`shared/spec/PORT-PLAN.md`](../../shared/spec/PORT-PLAN.md).

Precondition: G11 merged at da62113, with the review commits through f26720a. ADR-0008 records Go
1.24.13 as the supported minimum, and the example gate requires an output comment. The local line is
`ci: 9/9 gates passed, 1 reduced (tests without -race)`. The machine is shared, so this phase runs
no benchmarks.

The roadmap ended at G11. G12 closes the M3 gaps that need no tag, remote or benchmark run: data
expiry, release preflight, a committed `go doc` preview and a draft 0.1.0 note. The version stays
0.1.0 and no tag is created.

## Tasks

### G12-00 Plan
This file, the roadmap row and the progress rows. Commit before any tool, workflow or generated
file change.

AC:
- The roadmap diagram and phase table name G12 and its four work tasks.
- The plan names the data thresholds, the release preflight checks, the godoc generation version
  and the draft-note boundary.
- `docs/plan/PROGRESS.md` has a row for every G12 task before its work lands.
- No code, test, workflow, generated documentation or changelog change in this commit.

### G12-01 Data expiry
Add `tools/data-status`, standard library only. It scans the compiled patro tables through the
public package, not `shared/data`, and finds the last day with tithi, with a public holiday and with
an event. It prints one line per kind with the end date and days left from `--today`, a BS date in
`YYYY-MM-DD` form. With no flag it uses `nepdate.Today(nil)`, which is today in Nepal.

The threshold is 183 days. A holiday below 183 days exits 1. Tithi and event values below 183 days
print a warning and exit 0. Exactly 183 days passes. Missing or invalid input exits 2, and a missing
kind exits 1.

Unit tests cover 183 and 182 days for holiday and warning cases, the exact generated-table end
dates, the explicit and default-clock paths, and the pass and fail exit codes. `release.yml` runs
the tool in preflight after the release check. `ci.yml` does not run it.

AC:
- `go run ./tools/data-status -today=2084-01-01` exits 0 and prints the three end dates and days
  left.
- `go run ./tools/data-status -today=2084-08-01` exits 1 and names the holiday expiry.
- Unit tests at 183 and 182 days pass, with 182 as the negative control.
- The generated-table ends are asserted as tithi 2083-12-30, holidays 2084-08-18 and events
  2084-08-20.
- `release.yml` has the step and `ci.yml` has none.
- `go test ./tools/data-status` and `go vet ./tools/data-status` pass.
- `go list -m all` still lists only the root module and `go.sum` is absent.

Result on 2026-10-18: the scan finds tithi ending 2083/12/30, holidays 2084/08/18 and events
2084/08/20. `-today=2084-01-01` exits 0 with tithi at -1 days (warning) and holidays at 233 and
events at 235. `-today=2084-08-01` exits 1 with holidays at 17 days and tithi at -217 and events at
19 days. `TestClassify` covers 183 and 182 days, and `TestExpiryRows` asserts the three end dates.
`release.yml` runs the tool in preflight; `ci.yml` was not changed.

### G12-02 Release dry run
Extend `go run ./internal/ci release-check vX.Y.Z` to check, in addition to the CHANGELOG section
from ADR-0005:

- README contains `go get` with the module path in `go.mod`, and its `Go X.Y or later` line matches
  the `go` line in `go.mod`.
- The Go 1.24 patch in ADR-0008 matches the patch in the `ci.yml` matrix, and its minor version
  matches `go.mod`.
- `shared/VERSION` is a 40-character lowercase hexadecimal hub SHA.
- `shared/data/NOTICE` names `data/calendar/`, `data/patro/`, the MIT License and NepDate.
- `go list -m all` prints only the module path.

The command prints one line per passing check and exits 1 with every failed check named. The
existing `release.yml` preflight call becomes the full release readiness check. No new public API
and no dependency are added.

Tests build a complete temporary repository and make one negative control per check: missing
CHANGELOG section, wrong README import path, wrong README go line, mismatched ADR patch, non-SHA
version, incomplete NOTICE and an extra module from `go list -m all`. The passing fixture is also
tested.

AC:
- `go run ./internal/ci release-check v0.1.0` exits 0 on the branch and prints the seven passing
  check lines.
- Each of the seven negative controls exits 1 and names its check.
- `release.yml` still calls the command in preflight.
- `go test ./internal/ci` and `go vet ./internal/ci` pass.
- `go list -m all` still lists only the root module and `go.sum` is absent.

Result on 2026-10-18: `go run ./internal/ci release-check v0.1.0` prints seven passing lines and
`release-check: ok`. The seven negative controls in `TestReleaseReadinessNegativeControls` pass:
CHANGELOG section, README import path, README go line, ADR-0008 pin, shared version, NOTICE and
module dependency. `release.yml` names the step `Release readiness` and keeps the command in
preflight. The runbook lists the checks.

### G12-03 pkg.go.dev preview
Generate `docs/godoc/nepdate.txt` and `docs/godoc/patro.txt` with `go doc -all` under Go 1.27.1,
the gate toolchain, and commit them. Add a short `docs/godoc/README.md` that says they are generated
and names the command.

Add `TestGoDocPreview` in the root package. When `runtime.Version()` is exactly `go1.27.1`, it runs
`go doc -all .` and `go doc -all ./patro` and compares the exact bytes with the committed files. On
any other Go version it calls `t.Skipf` and prints the required version and the running version.
`internal/ci` adds `pkg.go.dev preview (needs Go 1.27.1)` to the reduced list when `go env GOVERSION`
is not exactly `go1.27.1`, so the summary never reads as a full pass.

AC:
- Both text files are committed and contain the full `go doc -all` output.
- Under Go 1.27.1, `go test -run TestGoDocPreview .` passes.
- Under Go 1.24.13, `go test -run TestGoDocPreview -v .` prints a skip that names both versions.
- A changed doc comment makes the test fail with the stale file name, and regenerating restores it.
- Under a Go version other than 1.27.1, `go run ./internal/ci` adds the reduced reason; under
  1.27.1 the line stays `1 reduced (tests without -race)`.
- `TestGoDocReduced` covers 1.24.13, 1.27.0, 1.27.1 and 1.27.2.

Result on 2026-10-18: `docs/godoc/nepdate.txt` is 18,482 bytes and `patro.txt` is 2,479 bytes, both
made with Go 1.27.1. `TestGoDocPreview` compares both byte for byte and passes. Under Go 1.24.13 it
prints `godoc_test.go:20: pkg.go.dev preview needs go1.27.1, this is go1.24.13` and `--- SKIP`.
Under Go 1.27.2 the full run ends with
`ci: 9/9 gates passed, 2 reduced (pkg.go.dev preview (needs go1.27.1), tests without -race)`.

Control: a temporary sentence in the `doc.go` package comment made the test fail with
`docs/godoc/nepdate.txt is stale; regenerate with ...` and the first differing line. Removing the
sentence restored the pass. The test cache does not hash comments, so gate 3 now runs
`go test -count=1` and a doc-only change cannot hide behind a cached pass.

### G12-04 Draft release notes
Create `docs/release/0.1.0-notes.md` as a draft candidate only. It merges the existing
`[0.1.0]` section with the `[Unreleased]` items, uses `## [0.1.0] - 2026-10-18`, and starts with a
note that `CHANGELOG.md` and `go.mod` are unchanged and no tag exists. Do not edit `CHANGELOG.md`,
`go.mod`, `README.md`, the version or any tag.

AC:
- The draft file exists with the candidate date and both the initial 0.1.0 features and the
  [Unreleased] additions and change.
- The draft says it is not the changelog and that no tag was made.
- `CHANGELOG.md`, `go.mod` and the git version are unchanged.
- `go run ./internal/ci` passes and the docs gate accepts the new page.
- No tag, remote or repository setting changes.

Result on 2026-10-18: `docs/release/0.1.0-notes.md` holds the candidate notes under
`## [0.1.0] - 2026-10-18`. It merges the initial 0.1.0 features with the two `[Unreleased]` additions
and the parse fast-path change, and says that `CHANGELOG.md`, `go.mod` and the tag are untouched.
The docs gate accepted the page and the full run kept the line
`ci: 9/9 gates passed, 1 reduced (tests without -race)`.

## Checkpoint

Stop after G12-04. Report the commits, the real local `ci` line verbatim, the `data-status` output
for one passing and one failing date, the negative controls, the Go 1.24 skip line and every
doubtful point. Do not tag, publish, add a remote, change repository settings or edit the released
changelog section. Do not start a later phase.

## Out of scope

No public API or behaviour change, no new root-module dependency and no benchmark run. No tag,
GitHub release, proxy request or settings change. GitHub CI is not measured because this checkout
has no remote.
