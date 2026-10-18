# G8: Mutation testing

Read first: ADR-0004, ADR-0006, [`G7-b6-fast-path.md`](G7-b6-fast-path.md).

Precondition: G7 merged on `main` (bebe42e). Mutation testing measures the tests, not the library:
a surviving mutant is a test gap, and the answer is a new test or a written equivalence argument.
Production code is never weakened to raise the score.

Tool: Gremlins v0.6.0, package `github.com/go-gremlins/gremlins/cmd/gremlins`. The version is
pinned in the `gremlinsVersion` constant of `tools/mutation-check`, which runs
`go run github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0`. The root module keeps zero
dependencies: the download is a build cache entry, not a `go.mod` requirement.

## Tasks

### G8-01 Plan
This file, the roadmap row and the progress row. Commit before the first mutation run.

AC:
- The plan lists the tool, its pin and every later task.
- The root module's `go.mod` and the absent `go.sum` are unchanged.

### G8-02 Baseline
`tools/mutation-check` runs Gremlins on the hand-written core files with
`--workers=4 --test-cpu=1 --timeout-coefficient=10`, and writes machine-readable JSON. The score is
the JSON `test_efficacy`, which is `KILLED / (KILLED + LIVED)`. Survivors are `LIVED`, `NOT COVERED`
and `TIMED OUT` mutants, grouped by file. Generated tables and subdirectories are excluded.

AC:
- The run uses at most 4 workers and writes machine-readable JSON.
- The baseline score, the survivor count and the survivor table by file are recorded.
- Every survivor has a file, line, mutant type and status.

Baseline on 2026-10-18, after a first run at the default timeout coefficient 3 produced 24
timeouts:

| Metric | Value |
|---|---:|
| killed | 469 |
| lived | 0 |
| not covered | 53 |
| timed out | 0 |
| test efficacy | 100.00 % |
| mutations coverage | 89.85 % |
| survivors | 53 |
| elapsed | 436.7 s |

| File | Lived | Not covered | Timed out | Survivors |
|---|---:|---:|---:|---:|
| lenient.go | 0 | 21 | 0 | 21 |
| pattern.go | 0 | 11 | 0 | 11 |
| parse.go | 0 | 9 | 0 | 9 |
| arith.go | 0 | 4 | 0 | 4 |
| fiscal.go | 0 | 3 | 0 | 3 |
| time.go | 0 | 3 | 0 | 3 |
| errors.go | 0 | 1 | 0 | 1 |
| names.go | 0 | 1 | 0 | 1 |

The survivor details are reproduced in the final report. The first run's 24 timeouts disappeared
when the coefficient rose from 3 to 10, so the clean baseline is used.

### G8-03 Kill or classify
Kill every survivor with a test, or argue in the report why the mutant is equivalent. An
equivalent mutant gets a comment in the test that covers the line, not a code change that hides
the mutation.

AC:
- Every survivor from G8-02 has a disposition: killed by a named test, or equivalent with a reason.
- The final score and survivor count are recorded.
- No production code is weakened: no check, error path or boundary is removed.

Final run on 2026-10-18, after adding two tests. The first is
`FiscalYear(-2).Label(English) == "-2/01"`, which exercises the sign branch at `fiscal.go:104`.
The second is a Kelvin-sign month spelling, `15 \u212Aartik 2080`, which exercises the `foldEqual`
branch. Gremlins still maps that branch to a switch-case condition, so the Kelvin test does not
change the JSON status.

| Metric | Baseline | Final |
|---|---:|---:|
| killed | 469 | 471 |
| lived | 0 | 0 |
| not covered | 53 | 51 |
| timed out | 0 | 0 |
| test efficacy | 100.00 % | 100.00 % |
| mutations coverage | 89.85 % | 90.23 % |
| survivors | 53 | 51 |

The fiscal test killed the two `fiscal.go:104` mutants. The remaining 51 are dispositioned here:

| File | Count | Disposition |
|---|---:|---|
| `lenient.go` | 21 | Switch and case conditions in `date`, `foldEqual` and `lenientSeparator`. The bodies are exercised by the parse vectors and the Kelvin test, and a manual `case nMonths == 1 && nNums != 2` mutation fails the suite. Gremlins cannot attribute coverage to the condition expression, so it never tests these mutants. Not equivalent, but unscorable by this tool. |
| `pattern.go` | 11 | Switch cases in `appendToken`. The format golden tests exercise every case, and a manual `case c == 'y' && n > 4` mutation fails the suite. Same coverage artifact. |
| `parse.go` | 9 | Switch cases in `scanStrictLoop`. The parse vectors exercise ASCII, Devanagari and danda digits, and a manual `case c > '0' && c <= '9'` mutation fails the suite. Same coverage artifact. |
| `arith.go` | 4 | The compile-time constant `bsMonths`. A manual `(maxYear - minYear - 1) * 12` mutation fails `TestDiff` and `TestDiffTotalDaysMatchesDaysUntil`. Constant expressions have no coverage counter, so Gremlins does not score it. |
| `time.go` | 3 | The `time.FixedZone("NPT", 5*3600+45*60)` fallback, reached only when the zone database is missing. Equivalent on a machine with tzdata and not reachable in the test environment without removing it. |
| `errors.go` | 1 | `InvalidCharacter ParseErrorKind = iota + 1`. The `iota - 1` mutation is a negative value in a `uint8` constant and does not compile, so it is not viable. |
| `names.go` | 1 | `English Lang = iota`. The `-iota` mutation makes `Nepali` negative in a `uint8` and does not compile, so it is not viable. |
| `fiscal.go` | 1 | `Q1 Quarter = 1 + iota`. The `1 - iota` mutation makes Q3 and Q4 negative in a `uint8` and does not compile, so it is not viable. |

The score is already 100.00 % before the tests, so the two tests do not change the score; they
remove two uncovered mutants. The remaining survivors are compile-time constants, non-viable
constants, a platform fallback, and switch-case conditions that Gremlins cannot map to a coverage
counter. A manual mutation of a representative switch case from each file fails the suite, which is
the evidence that the tests kill the behavior even though Gremlins cannot score it.

### G8-04 Gate
`tools/mutation-check` enforces:

- `test_efficacy` as an unrounded float against 99.5 %, just below the final 100.00 %.
- Every `(file, mutator type, status)` count in `tools/mutation-check/survivors.txt` exactly. A
  higher, lower, new or missing count fails, and the output names the key. Line and column are not
  in the key, so an edit elsewhere in the file does not churn the baseline.
- `-update` rewrites the sorted baseline. It is the only way to accept a new survivor or record a
  kill.

There is no separate survivor total. The sum of the baseline counts is the only survivor number,
so a new survivor cannot hide behind a killed one.

The K7 lesson: a rounded percentage can hide a regression. 95.4 % must fail a 96 % gate, and one
extra survivor must fail even when the rounded percentage stays the same. The unit test
`TestScoreIsNotRounded` checks 95.9 % against 96 %.

The `mutation` job in `.github/workflows/ci.yml` runs
`go run ./tools/mutation-check -workers=4` with a 40-minute timeout. The runner rejects a worker
count above 4, and Gremlins also gets `--test-cpu=1`.

Negative control: removing the `FiscalYear(-2).Label` assertion turns two killed `fiscal.go:104`
mutants into survivors. The score stays at 100.00 %, because a not-covered mutant is not a lived
mutant, but the baseline check fails on two keys:

```
mutation-check: survivor baseline differs in 2 keys:
  fiscal.go ARITHMETIC_BASE NOT COVERED: 2, baseline 1 (higher count)
  fiscal.go INVERT_NEGATIVES NOT COVERED: 1, new key
run `go run ./tools/mutation-check -update` to accept the new baseline
```

The report for that run was 469 killed, 0 lived, 53 not covered, 0 timed out, efficacy 100.00 %.
Restored, `go run ./tools/mutation-check -json E:/temp/g8/final.json` prints `mutation-check: ok`.
A doctored copy of the report with `test_efficacy` set to 95.9 prints
`mutation-check: test efficacy 95.90 % is below 99.50 %` and exits 1. The two checks cover
different regressions: the score catches a covered mutant that stops being killed, and the baseline
catches a test that stops covering a line or a mutant type that starts surviving.

The `mutation` job is a separate job in `ci.yml`, next to `ci`, `consumer`, `lint` and
`govulncheck`. The eight gates inside `internal/ci` are unchanged, so the `ci:` line stays 8/8.

AC:
- The gate fails below the score threshold and on any higher, lower, new or missing baseline line.
- `-update` rewrites the baseline; no `-max-survivors` flag remains.
- The CI job runs the gate with at most 4 workers.
- Negative control: delete one test that kills a mutant, show the per-file failure, restore the
  test and show it passes.
- `go run ./internal/ci` still ends with `ci: 8/8 gates passed`.

## G8b: Gate hardening

G8 merged at cce831f. The single total ceiling let a newly uncovered mutant hide behind a newly
killed one, and the Linux survivor set was unmeasured.

### G8b-01 Per-file survivor baseline
`survivors.txt` has one line per file, mutator type and status, sorted, with the count last. The
header says how to regenerate it. The gate uses `-update` to rewrite the file; `-max-survivors` is
gone, so the file is the only source of the survivor total.

Tests cover a higher count, a new key, a lower count, a missing key, an exact match and the update
path. The negative control above is the higher and new key case.

AC:
- The gate fails on a higher, lower, new or missing key and prints the key and both counts.
- `-update` rewrites the file sorted, with line numbers absent from the key.
- The exact baseline passes, and the score gate still fails below 99.5 %.
- `ci: 8/8 gates passed`.

### G8b-02 Linux baseline
GitHub CI cannot run because the repository has no remote. The Linux run is a clean clone of
`feat/G8b-mutation-gate` at 0eafbed on WSL Ubuntu 26.04, Go 1.27.1 linux/amd64, four workers. The
Go toolchain, module cache and build cache are inside the WSL filesystem (`/home/sanam/g8b-go` and
`/home/sanam/g8b-cache`), not on a Windows drive. The clone is `/home/sanam/g8b-clone`.

| Run | Killed | Lived | Not covered | Timed out | Test efficacy | Mutation coverage | Survivors | Elapsed |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Windows | 471 | 0 | 51 | 0 | 100.00 % | 90.23 % | 51 | 409.5 s |
| Linux | 471 | 0 | 51 | 0 | 100.00 % | 90.23 % | 51 | 207.6 s |

The 13 baseline keys and all counts are identical on both platforms. File names are normalized to
slash form in the runner, and `Asia/Kathmandu` loads from the system zone database on both, so the
fallback at `time.go:47` stays uncovered. One baseline serves both platforms.

The first WSL attempt put the clone, toolchain and module cache on `/mnt/d`. Each mutant rebuild
read them over 9p and the run was too slow to finish. Moving them inside the distro made the same
run finish in 207.6 s. The table above uses the in-distro run.

AC:
- The Linux run is a clean clone of the branch, four workers.
- The two raw summaries are recorded.
- The survivor set is compared with Windows and the baseline is platform-neutral, or a difference
  is explained.

## Out of scope

No tags, publishing, remote or repository settings changes. The gate runs in CI, not inside
`internal/ci`, because a mutation run takes minutes.
