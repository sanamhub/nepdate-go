# G10: Fuzz automation and API compatibility gate

Read first: ADR-0004, ADR-0006, [`docs/API.md`](../API.md), `shared/spec/PARSING.md` and
`shared/spec/FORMATTING.md`.

Precondition: G9 merged at 357cbb4. The local line is
`ci: 8/8 gates passed, 1 reduced (tests without -race)`. The race detector stays a CI-only check,
by the G9 review decision. The machine is shared, so this phase runs no benchmarks.

Tool: `golang.org/x/exp/cmd/apidiff` pinned at
`v0.0.0-20261014143641-f0251eb4aa3e`. The gate fetches it with `go run` and a version suffix, so the
root module keeps zero dependencies and no `go.sum`. The snapshots are
`testdata/api/nepdate.snapshot` and `testdata/api/patro.snapshot`, made by `apidiff -w` in package
mode. Module mode would also report the command packages under `tools/`, which are not API.

Fuzz runs use the standard `go test -fuzz` driver. Local default: 30 s per target, `-parallel=4`,
one target at a time, so at most four fuzz workers. Workflows use the same tool with a fixed longer
budget. A crasher is written by Go under `testdata/fuzz/<target>/` and must stay there, because the
next plain `go test` runs it as a regression test.

## Tasks

### G10-00 Plan
This file, the roadmap row and the progress rows. Commit before the first fuzz run or tool install.

AC:
- The roadmap lists G10 as a follow-up phase and names its four tasks.
- The plan names the fuzz budgets, the apidiff pin, the snapshot path and the stop point.
- `docs/plan/PROGRESS.md` has a row for every G10 task before its work lands.
- No production code, test or workflow changes in this commit.

### G10-01 Fuzz inventory
Record every `Fuzz*` target, the property it checks and its seed corpus. Add a target for every
public parse or format entry point that has none: `String`, `Short` and `AppendShort` for all
orders, `Long` and `AppendLong`, `Format` and `AppendFormat`, `MarshalText`, `AppendText` and
`AppendDigits`. Each new target compares against a test-only reference written from the spec, or
against an invariant that crosses a format and a parse direction. A target that calls the same
function twice is not acceptable.

The inventory is the table in this file. It must name the package, the entry point, the seed count
and the property. Existing targets stay; a widened target keeps its name.

AC:
- The plan lists all 6 existing targets with property and seed corpus.
- The plan lists each new target, its entry point and its independent oracle or invariant.
- `go test -run '^Fuzz' ./...` passes and executes every seed.
- `go test ./...` passes with no production behaviour change.
- The only production change is a real bug fix, and then the phase stops with the finding first.

### G10-02 Fuzz runs
Add `tools/fuzz-check`. It discovers targets from `go test -list '^Fuzz'`, runs each with
`-fuzz=^<name>$ -fuzztime=<budget> -parallel=4`, and prints one line per target with its exec count,
new corpus entries and the result. It rejects a parallel value above 4. It lists
`testdata/fuzz/` before and after a target: a new file is a crasher, is left in place and fails the
run at that target. A normal failure also stops the run.

Run every target for 30 s locally and record the table. Add a weekly workflow, and the same job in
`release.yml`, with Go 1.27.1 on `ubuntu-latest`, SHA-pinned actions, `permissions: contents: read`
and a fixed longer budget. Do not add fuzzing to `ci.yml` or to the PR path.

AC:
- `go run ./tools/fuzz-check -time=30s` runs every target and ends with a summary line.
- Every target reports a nonzero exec count and zero new corpus entries and crashers locally.
- A synthetic crasher in a temporary copy is detected and the tool exits 1, with the corpus path.
- `.github/workflows/fuzz.yml` is scheduled weekly and has `workflow_dispatch`; it does not run on
  `pull_request`.
- `release.yml` has the same fuzz job and the release job needs it.
- The workflows use Go 1.27.1 and SHA-pinned actions.
- The local run, per-target numbers and any finding are recorded in this file.
- `ci: 9/9 gates passed, 1 reduced (tests without -race)` after G10-03.

Local run on 2026-10-18, Windows amd64, Go 1.27.1, 30 s per target and `-parallel=4`. The run
followed an earlier 30 s run, so the fuzz cache already held 52 entries; "total corpus" is the size
at the end and "new this run" is what this run added. No target wrote a file under `testdata/fuzz`.

| Target | Execs | Total corpus | New this run | Crashers |
|---|---:|---:|---:|---:|
| `FuzzAddDays` | 534,944 | 8 | 0 | 0 |
| `FuzzAppendDigits` | 718,486 | 26 | 0 | 0 |
| `FuzzFormatPattern` | 787,325 | 86 | 13 | 0 |
| `FuzzFormatRoundTrip` | 646,189 | 15 | 0 | 0 |
| `FuzzLong` | 598,076 | 11 | 0 | 0 |
| `FuzzParse` | 982,338 | 81 | 0 | 0 |
| `FuzzParseLenient` | 587,441 | 146 | 33 | 0 |
| `FuzzScanCanonical` | 718,458 | 48 | 2 | 0 |
| `FuzzText` | 545,926 | 9 | 1 | 0 |
| `FuzzUnmarshalText` | 509,976 | 82 | 3 | 0 |
| **Total** | **6,629,159** | | **52** | **0** |

The first 30 s run, before the cache existed, reported 7,329,831 execs and 186 new entries. It also
exposed a reporting bug: the tool summed the repeated cumulative `new interesting` value and printed
44 new entries for `FuzzAddDays` against 4 actual. `parseStats` now keeps the last value, and
`TestParseStatsKeepsCumulativeCountOnce` covers the repeated line. The rerun above uses the fixed
tool.

`go test -list` found 10 targets with no duplicate. `actionlint` accepted `ci.yml`, `fuzz.yml` and
`release.yml`. Fuzzing is not in `ci.yml`; `fuzz.yml` runs weekly and on dispatch, and `release.yml`
runs the same 3-minute budget before `release`.

### G10-03 API compatibility
Commit `testdata/api/nepdate.snapshot` and `testdata/api/patro.snapshot`, the apidiff export data
of the current exported surface of each package. Add `tools/api-check`:

- `go run ./tools/api-check` compares each snapshot with the current package. A removed or changed
  identifier fails.
- An addition fails until the snapshot is updated, and every exported identifier in both packages
  must appear in `docs/API.md`. An addition that is missing from that page fails even after
  `-update`.
- `go run ./tools/api-check -update` rewrites both snapshots from the current packages.

Add gate 9 to `internal/ci`. Write ADR-0007 for the new gate and update the ADR index; do not edit
ADR-0004 or ADR-0006. Add an `api` job to `ci.yml` and `release.yml`, and make `release` need it.

AC:
- Both snapshots are committed `-update` output and their byte sizes are recorded.
- `go run ./tools/api-check` exits 0 on the branch.
- Renaming an exported function fails, naming the old and new identifiers.
- Adding an exported function without `docs/API.md` fails, naming the identifier, both with and
  without `-update`.
- `go run ./internal/ci` ends with `ci: 9/9 gates passed` in CI, and locally with
  `1 reduced (tests without -race)`.
- `go list -m all` lists only the root module, `go.sum` is absent and no dependency is added.
- ADR-0007 records the gate, its tool pin and why `internal/` and `tools/` are excluded.

Result on 2026-10-18: `nepdate.snapshot` is 7,486 bytes and `patro.snapshot` is 6,387 bytes, 13,873
bytes together. Package mode reports "no changes" for both. The gate finds 126 distinct exported
names in the two packages and every one is in `docs/API.md`. `go test ./internal/ci` and
`go test ./tools/api-check` pass, `golangci-lint` reports 0 issues, and `actionlint` accepts all
three workflows. ADR-0007 records gate 9 and the package-mode reason.

### G10-04 Negative controls
Run one control for each new failure path:

- plant a panic in a parse path, run the local 30 s fuzz budget and show the failing fuzz target;
- rename an exported function and show the API gate failure;
- add an exported function with a doc comment but no `docs/API.md` entry, run `-update`, and show
  the failure.

Remove every planted change and crasher after its control. Restored, the same commands pass.

AC:
- Each control records the exact command, failure line and exit code.
- The planted panic is found within the 30 s budget and its corpus file is named.
- The renamed function failure names the removed and added identifiers.
- The undocumented addition fails after `-update`, so the documentation check, not the snapshot,
  is the failing step.
- Restored, `go run ./tools/fuzz-check -time=30s` and `go run ./tools/api-check` pass.
- `git status` is clean after the restores.

## Fuzz inventory

Measured from the branch at G10-01; seed counts are the `f.Add` calls or table rows read by that
target.

| Target | Package | Entry point | Seeds | Property |
|---|---|---|---:|---|
| `FuzzParse` | `nepdate` | `Parse` | 67 | No panic; an error is `*ParseError` or a validation sentinel; success formats with `String` and parses back to the same date. |
| `FuzzParseLenient` | `nepdate` | `ParseLenient` | 67 | Same typed-error and `String` round-trip invariant as `FuzzParse`. |
| `FuzzUnmarshalText` | `nepdate` | `UnmarshalText`, `MarshalText` | 67 | Same typed-error rule, then a text round trip through `MarshalText` and `UnmarshalText`. |
| `FuzzAddDays` | `nepdate` | `AddDays` | 4 | Adding `n` then `-n` returns the original date; every error is `ErrOutOfRange`. |
| `FuzzFormatRoundTrip` | `nepdate` | `Short`, `AppendShort` | 10 | Widened in G10-01: all 6 orders, 6 separators, both pad values and both languages; a test-only formatter is the oracle; YMD output parses back. |
| `FuzzScanCanonical` | `nepdate` | unexported `scanStrict` path | 19 | The fast path agrees with the reference loop on date and error, and dispatch returns the same result. |

New targets:

| Target | Package | Entry point | Seeds | Independent oracle or invariant |
|---|---|---|---:|---|
| `FuzzLong` | `nepdate` | `Long`, `AppendLong` | 4 | A test-only long formatter built from the FORMATTING §1 names and §4 layout. |
| `FuzzFormatPattern` | `nepdate` | `Format`, `AppendFormat` | 6 | A test-only pattern scanner built from FORMATTING §5; `AppendFormat` keeps a prefix. |
| `FuzzText` | `nepdate` | `String`, `MarshalText`, `AppendText` | 4 | A `fmt.Sprintf` oracle; the text parse round trip is an independent direction. |
| `FuzzAppendDigits` | `nepdate` | `Lang.AppendDigits` | 5 | A test-only decimal formatter from the ADR-0002 digit rule; zero padding and sign are explicit. |

`Month.Name`, `Month.ShortName`, `WeekdayName` and `WeekdayShortName` are table lookups with no
parse or scan path. The format oracles check their output for every date, and the static tables are
covered by the golden vectors, so no separate target is added. `patro` has no parse or format entry
point.

The six existing targets and four new targets hold 253 seeds: 67 each for `FuzzParse`,
`FuzzParseLenient` and `FuzzUnmarshalText`, 19 for `FuzzScanCanonical`, 10 for
`FuzzFormatRoundTrip`, 5 for `FuzzAppendDigits`, 6 for `FuzzFormatPattern`, and 4 each for
`FuzzAddDays`, `FuzzLong` and `FuzzText`. `go test -run '^Fuzz' .` passed all 253 seeds on
2026-10-18. No `testdata/fuzz` corpus is committed at the start of the phase.

## Checkpoint

Stop after G10-04. Report the commits, the real local `ci` line verbatim, fuzz numbers per target,
the API snapshot size, the negative controls, and every doubtful point. Do not tag, publish, add a
remote or change repository settings. Do not start a later phase.

## Out of scope

No public API or behaviour change, no dependency in the root module and no benchmark run. The fuzz
workflow runs weekly and on release, not on every pull request. GitHub CI is not measured because
this checkout has no remote.
