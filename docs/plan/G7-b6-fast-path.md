# G7: B6 strict-parse fast path

Read first: ADR-0002 sections 5 to 7, the reviewer decision in
[`B6-strict-parse-options.md`](B6-strict-parse-options.md), and [`docs/benchmarks.md`](../benchmarks.md).

Precondition: G6 merged on `main` (753dcdf). The reviewer chose Option A: a canonical 10-byte ASCII
fast path in `scanStrict`, with an exhaustive differential test and a differential fuzz target.

## G7-01 Step 0: old against old

The quiet-machine rerun in the options file is not available because the machine is busy, so the
noise floor is measured directly. Two builds of the old code were made from `benchmarks/` before any
change; their SHA-256 hashes are identical (`5488E66F...`). They ran interleaved, A B A B, 10
rounds each, pinned with `/high /affinity 4`, with `-test.run=NoTest -test.bench=BenchmarkB6
-test.benchtime=300ms -test.count=1`, and compared with `benchstat`.

| Binary | ns/date | B/op | allocs/op |
|---|---:|---:|---:|
| old-a | 42.25 ± 18 % | 0 | 0 |
| old-b | 41.77 ± 30 % | 0 | 0 |

The median difference is 1.1 % (p=0.796, n=10), so two runs of the same binary are
indistinguishable. CPU load was sampled every second: 69.1 % minimum, 89 % median, 100 % maximum.

AC:
- The two old binaries have the same hash.
- The interleaved table shows the noise floor and the load next to it.

## G7-02 Option A fast path

At the top of `scanStrict`, accept exactly 10 bytes with digits at 0 to 3, 5 to 6 and 8 to 9, and a
strict separator at 4 and 7. Build the year, month and day with direct digit arithmetic, then call
`newFromInts`, so the range, month and day checks and their error order stay exactly as they are.
Anything else runs the existing loop, which moves to `scanStrictLoop` and stays the reference. Put
the fast path in `scanStrict` so `ParseLenient` benefits too.

Tests, written before the performance numbers count:

- an exhaustive differential test over all 109,212 dates from 1901-01-01 to 2199-12-30, with each of
  the seven strict separators, comparing `scanCanonical` with `scanStrictLoop`;
- `FuzzScanCanonical`, a differential target over arbitrary input that compares result, error and
  error kind, and requires every canonical-shaped input to take the fast path;
- a negative control: break one digit check and show both tests fail.

The change lands only if the interleaved old versus new median gain is at least 30 %, with 0 B/op
and 0 allocs/op. Below 30 %, revert the code and report the numbers instead. Add a CHANGELOG
`[Unreleased]` line only when it lands.

AC:
- The two differential tests pass.
- Breaking one digit check makes both tests fail.
- `go test -run Alloc` still reports 0 for every parse path.
- The interleaved old versus new table shows at least a 30 % gain, with the load next to it.
- `go run ./internal/ci` ends with `ci: 8/8 gates passed`.
- `CHANGELOG.md` has an `[Unreleased]` line only if the change lands.
