# Benchmarks (Go)

Time and allocation of the Go port on operations B1 to B8 of
[`shared/spec/BENCHMARK.md`](../shared/spec/BENCHMARK.md), against the one competitor that passes
[`docs/competitors.md`](competitors.md), `github.com/opensource-nepal/go-nepali` v0.4.1, and the C#
NepDate baseline.

## Machine

| Item | Value |
|---|---|
| CPU | 12th Gen Intel Core i5-12500H (4 performance + 8 efficiency cores, 16 threads, base 3.1 GHz) |
| RAM | 31.7 GB |
| OS | Windows 11 Home 10.0.26300 |
| Power plan | Ultimate Performance |
| Affinity | logical CPU 2 (mask `0x4`), set with `start /affinity 4` |
| Go | go1.27.1 windows/amd64 |
| Idle gate before the run | 9.2 percent average, 12.2 peak (30 one-second samples) |
| Load during the run | 16.2 percent average, 22.6 peak over 42 one-second samples; 10.6 and 17.0 on the CPUs other than CPU 2 |

The machine was idle: no other build, test suite or benchmark ran during the run. The load during
the run includes the benchmark's own pinned core and the sampler. The samples are in
[`go-bench-2026-11-16-cpu-samples.txt`](../benchmarks/go-bench-2026-11-16-cpu-samples.txt).

## Command

From `cmd` in `benchmarks/`:

```
go test -c -o bench.test.exe .
start "" /high /affinity 4 /b /wait bench.test.exe -test.run=^$ -test.bench=. -test.benchtime=300ms -test.count=10 > go-bench-2026-11-16.txt
```

The raw output is [`go-bench-2026-11-16.txt`](../benchmarks/go-bench-2026-11-16.txt), and the console
log with the load samples is [`go-bench-2026-11-16.log`](../benchmarks/go-bench-2026-11-16.log).

## Results

Time is the median of the 10 runs of the `ns/date` metric (one benchmark iteration covers the 64
inputs; B8's operation is one whole month grid, per BENCHMARK section 3). "0" means the benchmark
measured no allocation. The C# column is the .NET 10 column of the hub's
[`docs/benchmarks/dotnet.md`](https://github.com/sanamhub/nepdate/blob/main/docs/benchmarks/dotnet.md),
measured on the same machine on 2026-11-20.

| ID | nepdate ns/op | std ns | B/op | allocs/op | gonepali ns/op | agrees | C# ns/op | C# B/op |
|---|---:|---:|---:|---:|---:|---|---:|---:|
| B1 BS to AD | 5.24 | 0.05 | 0 | 0 | 111.85 | no (3 of 64) | 5.60 | 0 |
| B2 AD to BS | 4.20 | 0.03 | 0 | 0 | 49.05 | no (4 of 64) | 5.90 | 0 |
| B3 add 1000 days | 4.03 | 0.01 | 0 | 0 | 212.85 | n/a (convert) | 2.79 | 0 |
| B4 add 13 months | 3.64 | 0.12 | 0 | 0 | | | 2.45 | 0 |
| B5 long, Nepali, weekday | 24.79 | 0.14 | 0 | 0 | | | 29.96 | 69 |
| B5-string `Long` | 42.92 | 0.70 | 4,144 | 64 | | | 29.96 | 69 |
| B6 strict parse | 7.84 | 0.04 | 0 | 0 | | | 7.28 | 0 |
| B7 patro info | 7.11 | 0.03 | 0 | 0 | | | 16.46 | 51 |
| B8 patro month grid | 255.55 | 2.83 | 0 | 0 | | | 633.13 | 1,429 |

The B3 gonepali row converts to AD, adds 1000 days with `time.AddDate`, and converts back, because
go-nepali has no O(1) add-days. Its agrees column is not measured; `TestAgrees` covers B1 and B2.

## Targets

- BENCHMARK section 6 (release blocker): B1, B2 and B7 within 1.5 times C# NepDate. B1 is 0.94
  times, B2 is 0.71 times, B7 is 0.43 times C#, all with 0 B/op. Passes.
- ADR-0002 section 7 (reported, not blocking): every B operation faster than C# NepDate. B5-string
  allocates and is slower than C# B5. B8 is 255.55 ns against 633.13 ns. Not met: B3, B4, B6 and
  B5-string are slower.

## Agreement

The agrees counts come from `go test -run TestAgrees -v .` in `benchmarks/`, which compares
gonepali's B1 and B2 against `nepdate` on the 64 inputs. It differs on 3 B1 dates and 4 B2 dates:
BS 2086-07-23, 2088-01-04 and 2089-06-10, plus the rejected AD 1943-04-14 on B2. The differing
years are listed in [`docs/competitors.md`](competitors.md).
