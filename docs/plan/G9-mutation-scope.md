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

### G9-02 Patro mutation
Extend `tools/mutation-check` with `-package=core|patro`, defaulting to `core`. Each package has its
own canonical baseline, `survivors.txt` for `nepdate` and `survivors-patro.txt` for `patro`. The
same exact-count rule and 99.5 % efficacy gate apply, with at most four workers. `-json` evaluates
the report against the selected package's baseline.

Run both packages in CI and in the release workflow. Measure on Windows and on WSL Ubuntu with an
in-distro toolchain, as in G8b-02. Kill every real survivor with a test, or explain why it is
equivalent, unreachable or non-viable. Production code is never weakened to raise the score.

AC:
- `go run ./tools/mutation-check -package=core` and `-package=patro` both pass.
- `survivors-patro.txt` is committed in `-update` form and has a test that checks that form.
- Windows and Linux raw summaries are recorded, with one baseline per package.
- Every patro survivor has a disposition.
- `ci: 8/8 gates passed`.

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

### G9-04 Negative controls
For every gate path new in G9, delete a killing test, show the exact failure line, restore the test
and show the gate passes. Cover patro and each file changed by G9-03. Repeat the core control if
G9-02 or G9-03 changes the core baseline.

AC:
- Each negative control records the exact failure line and exit code.
- Restored, the same gate prints `mutation-check: ok` or the package-specific ok line.
- `ci: 8/8 gates passed`.

## Checkpoint

Stop after G9-04. Report the commits, the real local `ci` line, both package summaries on Windows
and Linux, every survivor disposition, the negative controls, the B5/B6 `benchstat` result and any
doubtful point. Do not start a later phase, tag, publish, add a remote or change repository
settings.

## Out of scope

No production behaviour change, no new exported identifier, no dependency in the root module and
no mutation run on `internal/` or `tools/`. The local Windows machine and one WSL Ubuntu
distribution are the only measured platforms.
