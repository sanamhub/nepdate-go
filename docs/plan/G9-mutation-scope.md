# G9: Mutation scope and honest local gates

Read first: ADR-0004, ADR-0006, [`G8-mutation-testing.md`](G8-mutation-testing.md).

Precondition: G8b merged at 92b77d9. The mutation gate covers only package `nepdate`, and the local
`ci` line can read as a pass while golangci-lint and govulncheck are skipped. This phase closes both
gaps without weakening code.

Tool: Gremlins v0.6.0, pinned in `tools/mutation-check`. The root module keeps zero dependencies.
The tool runs with at most four workers and `--test-cpu=1`.

Local tools: golangci-lint v2.14.0 and govulncheck v1.9.0. golangci-lint is pinned in
`.github/workflows/ci.yml`; govulncheck is not pinned there, so G9-01 adds the pin before the local
install.

## Tasks

### G9-00 Plan
This file, the roadmap row and the progress row. Commit before the first tool install or mutation
run.

AC:
- The roadmap lists G9 and its four tasks.
- The plan names the tool versions, packages, baselines and stop point.

### G9-01 Local tools
Install golangci-lint v2.14.0 and govulncheck v1.9.0 with `GOBIN`, `GOPATH` and `GOCACHE` under
`D:\devcache\g9`, never on C:. Add the govulncheck pin to both workflows: the action installs
`@latest`, so the workflow currently does not pin it. Then run `go run ./internal/ci` until the last
line is exactly `ci: 8/8 gates passed`, with no skipped gate.

Fix every lint and vulnerability finding properly. A suppression is allowed only with a reason on
the line, as the existing `//nolint:gosec` comments do; no file-wide or linter-wide `nolint`.

Record whether `-race` can run locally. If a cgo toolchain on D: is cheap, enable it; otherwise say
why it is not.

AC:
- `golangci-lint version` and `govulncheck -version` show the pinned versions.
- The local `ci` last line is exactly `ci: 8/8 gates passed`, with no `skipped` suffix.
- `govulncheck` is pinned in `ci.yml` and `release.yml`, and `golangci-lint` stays at v2.14.0.
- Lint and vulnerability findings are fixed, not blanket-suppressed.
- The progress row records the race result.

Local result on 2026-10-18: golangci-lint v2.14.0 reports 0 issues, govulncheck v1.9.0 finds 0 called
vulnerabilities, and the last line is `ci: 8/8 gates passed`. The install used `GOBIN`, `GOPATH` and
`GOCACHE` under `D:\devcache\g9`.

`-race` does not run on Windows locally: `CGO_ENABLED=0` and there is no C compiler on D:. A full
mingw-w64 install would add a system toolchain outside the D: cache rule, so it was not installed.
WSL Ubuntu has gcc 15.2.0, and `go test -race -shuffle=on ./...` passes there.

### G9-02 Patro mutation
Extend `tools/mutation-check` with `-package=core|patro`, defaulting to `core`. Each package has its
own canonical baseline, `survivors.txt` for `nepdate` and `survivors-patro.txt` for `patro`. The
same exact-count rule and 99.5 % efficacy gate apply, with at most four workers. `-json` evaluates
the report against the selected package's baseline.

Run both packages in CI. The release workflow does not run the mutation gate; a tag is only made
from a `main` commit whose `ci` run, including the `mutation` job, is green. Measure on Windows and
on WSL Ubuntu with an in-distro toolchain, as in G8b-02. Kill every real survivor with a test, or
explain why it is equivalent, unreachable or non-viable. Production code is never weakened to raise
the score.

AC:
- `go run ./tools/mutation-check -package=core` and `-package=patro` both pass.
- `survivors-patro.txt` is committed in `-update` form and has a test that checks that form.
- Windows and Linux raw summaries are recorded, with one baseline per package.
- Every patro survivor has a disposition.
- `ci: 8/8 gates passed`.

Measured on 2026-10-18 with four workers and timeout coefficient 30:

| Run | Package | Killed | Lived | Not covered | Timed out | Test efficacy | Mutation coverage | Survivors | Elapsed |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Windows | core | 471 | 0 | 51 | 0 | 100.00 % | 90.23 % | 51 | 481.5 s |
| Windows | patro | 24 | 0 | 0 | 0 | 100.00 % | 100.00 % | 0 | 14.8 s |
| Linux | core | 471 | 0 | 51 | 0 | 100.00 % | 90.23 % | 51 | 190.4 s |
| Linux | patro | 24 | 0 | 0 | 0 | 100.00 % | 100.00 % | 0 | 3.6 s |

`patro` has no survivor: all 24 hand-written mutants are killed, so `survivors-patro.txt` is the
three-line header with no data rows. No new patro test was needed; `TestAllData`, `TestAnchors`,
`TestEmpty`, `TestCoverage` and `TestPublicHolidays` already cover every mutant.

The first patro run at timeout coefficient 10 had four `TIMED OUT` mutants at `patro.go:37` and
`patro.go:38`. The patro test binary is so fast that ten times its coverage run is under a second.
Coefficient 30 removed them; three repeat runs had no timeout. The core baseline did not change.

### G9-03 Switch-case survivors
The 41 NOT COVERED switch-case mutants in `lenient.go`, `parse.go` and `pattern.go` are not
equivalent: a manual mutation of a representative case fails the suite, but Gremlins never tests
them because Go coverage has no counter for a switch condition. Try a test restructure first, then
a code restructure that keeps the same behaviour, if one is needed to let Gremlins score the case.

Any code restructure must keep the public API, the allocation guarantees and the B5 and B6
benchmarks. Show `benchstat` before and after on the same machine and load; no regression is
allowed. Lower every baseline count the restructure makes scorable, and record the reason for each
one that stays.

AC:
- For each of `lenient.go`, `parse.go` and `pattern.go`, the switch-case count falls or the reason
  is recorded.
- Behaviour is unchanged: the full test suite passes.
- B5 and B6 show no regression in `benchstat`.
- The baselines are updated with `-update`; the gates pass.
- `ci: 8/8 gates passed`.

Result: all 41 switch-case mutants are now scored and killed. The changes are if chains in
`lenientTokens.date`, `foldEqual`, `lenientSeparator`, `appendToken` and `scanStrictLoop`. A small
probe confirmed the cause: the same condition is `NOT COVERED` in a `switch` and `KILLED` as an
`if`. No test was added; the existing suite kills every new mutant.

| Metric | G8b final | G9-03 |
|---|---:|---:|
| killed | 471 | 512 |
| lived | 0 | 0 |
| not covered | 51 | 10 |
| timed out | 0 | 0 |
| test efficacy | 100.00 % | 100.00 % |
| mutations coverage | 90.23 % | 98.08 % |
| survivors | 51 | 10 |
| elapsed | 409.5 s Windows | 578.5 s Windows, 244.9 s Linux |

The remaining 10 are the same dispositions as G8: 4 `arith.go` constants, 3 `time.go` tzdata
fallback, and 1 each in `errors.go`, `names.go` and `fiscal.go` that are non-viable.

`benchstat` over 30 samples per binary, interleaved old and new on the same machine:

| Benchmark | Old | New | Change |
|---|---:|---:|---:|
| B5/library=nepdate | 34.14 ns/date | 32.71 ns/date | ~ (p=0.836) |
| B5/library=nepdate-string | 56.27 ns/date | 56.94 ns/date | ~ (p=0.542) |
| B6/library=nepdate | 9.438 ns/date | 8.684 ns/date | -7.98 % (p=0.040) |

B5 and B6 have no regression. The B6 change is an improvement, not a regression.

### G9-04 Negative controls
For every gate path new in G9, delete a killing test, show the exact failure line, restore the test
and show the gate passes. The new paths are the core baseline after the switch change and the new
patro baseline.

The core control removes `FiscalYear(-2).Label`, the same test that failed the G8b ceiling. A first
attempt removed `TestFoldEqual` and the Kelvin parse case, but the parse vectors cover the same
branch, so all mutants stayed killed; that is redundancy in the suite, not a gate failure. The patro
control removes every `patro` test file.

AC:
- Each negative control records the exact failure line and exit code.
- Restored, the same gate prints `mutation-check: ok`.
- `ci: 8/8 gates passed`.

Core control, `FiscalYear(-2).Label` removed, 510 killed, 12 not covered, exit 1:

```
mutation-check: survivor baseline differs in 2 keys:
  fiscal.go ARITHMETIC_BASE NOT COVERED: 2, baseline 1 (higher count)
  fiscal.go INVERT_NEGATIVES NOT COVERED: 1, new key
run `go run ./tools/mutation-check -update` to accept the new baseline
```

Patro control, every `patro` test file removed, 0 killed, 24 not covered, exit 1:

```
mutation-check: test efficacy 0.00 % is below 99.50 %
survivor baseline differs in 4 keys:
  patro.go ARITHMETIC_BASE NOT COVERED: 8, new key
  patro.go CONDITIONALS_BOUNDARY NOT COVERED: 5, new key
  patro.go CONDITIONALS_NEGATION NOT COVERED: 9, new key
  patro.go INVERT_NEGATIVES NOT COVERED: 2, new key
run `go run ./tools/mutation-check -update` to accept the new baseline
```

The patro control exposed a gate bug: Gremlins sets `mutants_total` to the killed count, so a report
with 0 killed and 24 not covered was rejected as "0 mutants" before the baseline ran. `checkReport`
now counts the mutations in the report instead, and `TestHasMutations` covers an all-not-covered
report.

Restored, `go run ./tools/mutation-check -package=core -json core-switch-win.json` and
`-package=patro -json patro-30.json` both print `mutation-check: ok`.

## Checkpoint

Stop after G9-04. Report the commits, the real local `ci` line, both package summaries on Windows
and Linux, every survivor disposition, the negative controls, the B5/B6 `benchstat` result and any
doubtful point. Do not start a later phase, tag, publish, add a remote or change repository
settings.

## Out of scope

No production behaviour change, no new exported identifier, no dependency in the root module and
no mutation run on `internal/` or `tools/`. The local Windows machine and one WSL Ubuntu
distribution are the only measured platforms.
