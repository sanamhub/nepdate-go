# G5: Docs, benchmarks, release v0.1.0

Read first: `shared/spec/BENCHMARK.md` (all of it), ADR-0002 §7 and §8 (targets and the
prototype numbers to compare with), ADR-0005, `docs/competitors.md`, `shared/spec/PORT-PLAN.md`
phase 5.

Preconditions: hub H6 merged, so the hub's `docs/benchmarks/dotnet.md` exists (the C# column).
If it doesn't, do G5-01, G5-03 and the Go side of G5-02, and report the C# column as blocked at
the checkpoint.

## Tasks

### G5-01 Docs and examples
Doc comment on every exported identifier (starts with its name, says what fails, and says
"It does not allocate." on every function ADR-0002 §5 lists). `example_test.go` with
`ExampleNew`, `ExampleDate_Format`, `ExampleDate_AppendLong`, `ExampleParseLenient`,
`ExampleFromTime`, and `patro/example_test.go` with `ExampleOf`: these are the README snippets,
so they compile and run. Each has an `// Output:` comment. README gets a quick start, a one-line
explanation of "patro" (hub ADR-0004) and a link to pkg.go.dev.
AC: `go test -run Example ./...` passes; each Go block in the README quick start is the body of one
of these functions, character for character; `golangci-lint` `revive` reports no missing doc
comment.

### G5-02 Benchmarks (`benchmarks/`, separate module)

Module and inputs:
- `benchmarks/go.mod`: `module github.com/sanamhub/nepdate-go/benchmarks`, `replace
  github.com/sanamhub/nepdate-go => ..`, and the four modules of `docs/competitors.md` pinned
  exactly. The root module still has no `go.sum`.
- `inputs_test.go` builds the inputs once, outside any timed loop, through the public API:
  the 64 BS dates of BENCHMARK §2 (`s0` = serial of 2000-01-01, `s1` of 2090-12-30, step
  `(s1 - s0) / 63` = 527, so date 0 is 2000-01-01 and date 63 is 2090-11-24), their AD
  equivalents, their `String()` texts, and the 64 patro dates of BENCHMARK §3 (step
  `(67077 - 64285) / 63` = 44 over the coverage). A test checks these four facts.

Benchmarks (`bench_test.go`): one function per operation (`BenchmarkB1` ... `BenchmarkB8`), one
sub-benchmark per library named `library=<name>` (`b.Run("library=nepdate", ...)`,
`library=gonepali`, `library=nepcal`, `library=gobs`, `library=gobikramsambat`), so `benchstat
-col /library` puts the libraries side by side:

| ID | Body for this library, run over all 64 inputs per iteration | Competitors |
|---|---|---|
| B1 | `d, _ := nepdate.New(y, m, d)`; `y2, m2, d2 := d.Gregorian()`; add all three to an accumulator | all four (calls in `docs/competitors.md`) |
| B2 | `d, _ := nepdate.FromGregorian(y, m, d)`; add `Year()+int(Month())+Day()` | all four |
| B3 | `d.AddDays(1000)` | none has it in O(1); go-nepali and go-bs via convert, `time.AddDate`, convert back, marked as such |
| B4 | `d.AddMonths(13, nepdate.Clamp)` | none |
| B5 | `buf = d.AppendLong(buf[:0], nepdate.LongOptions{Weekday: true}, nepdate.Nepali)` with `buf` made once with capacity 128; a second sub-benchmark `library=nepdate-string` uses `d.Long(...)` | none |
| B6 | `nepdate.Parse(text)` | none |
| B7 | `i := patro.Of(d)`; tithi name length, holiday flag, and the length of every event's English name | none |
| B8 | for each of the 64 dates, `patro.Of` for every day of its month (`nepdate.MonthRange(...).Days()`), reading the holiday flag and every event name length | none |

Rules (BENCHMARK §4):
- `for b.Loop()`, `b.ReportAllocs()`, and `b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/64, "ns/date")`
  at the end; the accumulator is written to a package-level `var sink int` after the loop so the
  work can't be removed.
- "agrees": `TestAgrees` runs each competitor's B1 and B2 on the 64 dates and compares with
  `nepdate`; the table prints `yes` or `no (n of 64)`.

Running on the reference machine (BENCHMARK §1), from `cmd` in `benchmarks/`, with other
applications closed:

```
go test -c -o bench.test.exe .
start "" /high /affinity 4 /b /wait bench.test.exe -test.run=^$ -test.bench=. -test.benchtime=300ms -test.count=10 > bench.txt
benchstat -col /library bench.txt
```

`/affinity 4` pins to logical CPU 2, a performance core (on this machine CPUs 0–7 are the four
performance cores with hyper-threading; CPU 8 and up are efficiency cores, measured about 2.5
times slower in ADR-0002 §8). `/high` keeps other processes off that CPU: without it the
prototype runs showed up to ±75 % variation, with it ±2 %. The Go runtime sets `GOMAXPROCS` to 1
under a one-CPU affinity mask, so benchmark names lose their `-16` suffix; that is expected.

Publish `docs/benchmarks.md` in the BENCHMARK §5 layout: the machine table, Go version,
competitor versions, the exact commands above, the date, and one table with columns ID,
nepdate ns/date, B/op, allocs/op, each competitor's ns/date and agrees, and the C# NepDate ns/op
and bytes/op copied from the hub's `docs/benchmarks/dotnet.md` (never measured here). Also
publish `bench.txt` as `docs/benchmarks/go-bench.txt` so `benchstat` can compare later runs.

AC:
- [ ] B1 to B8 exist for `nepdate`; B1 and B2 exist for every competitor that builds.
- [ ] `go test -run TestAgrees .` passes and its result fills the agrees column.
- [ ] For `nepdate`, `B/op` and `allocs/op` are 0 for B1, B2, B3, B4, B5 (`AppendLong`), B6, B7
      and B8. `B5` `library=nepdate-string` shows exactly 64 allocs/op (one per date, the returned string).
- [ ] BENCHMARK §6 (release blocker): B1, B2 and B7 within 1.5 times the C# NepDate ns/op.
- [ ] ADR-0002 §7 (report, not a blocker): every B-operation faster than C# NepDate.
- [ ] A missed target is reported to the human with the numbers and a CPU profile
      (`-test.cpuprofile`), not guess-optimised.

### G5-03 Release workflow + runbook
`.github/workflows/release.yml` per ADR-0005 and `docs/runbooks/release.md`.
`release.yml` also has `workflow_dispatch` with input `version`; a dispatched run does preflight
and CI and skips the release and proxy jobs.
AC (checked after G5-04, in the same PR): dispatched with `version: v0.1.0`, preflight and CI pass
and the release job shows as skipped.

### G5-04 v0.1.0
Move `[Unreleased]` to `## [0.1.0] - <date>`. **Agent stops here.** The human tags and approves.

## Checkpoint (agent stops here for review)

The reviewer checks:

- [ ] `go run ./internal/ci` ends with `ci: 6/6 gates passed`.
- [ ] `go test -run Example -v ./...` lists every README example and passes.
- [ ] `docs/benchmarks.md` has the machine table, the exact commands, versions, the date and all
      eight rows; the C# column cites the hub file and its commit.
- [ ] Re-run `benchstat docs/benchmarks/go-bench.txt` locally: the numbers in the table are the
      benchstat medians, not single runs.
- [ ] The BENCHMARK §6 release blocker holds, or the miss is reported.
- [ ] `CHANGELOG.md` has `## [0.1.0] - YYYY-MM-DD` and `go run ./internal/ci release-check v0.1.0`
      passes.
