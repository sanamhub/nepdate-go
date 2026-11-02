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
| Machine load before | 5, 3 and 3 percent, sampled 2 s apart |
| Machine load after | 0, 1 and 2 percent, sampled 2 s apart |

The machine had the owner's normal desktop applications open. No other build, test suite or
benchmark ran during the run.

## Command

From `cmd` in `benchmarks/`:

```
go test -c -o bench.test.exe .
start "" /high /affinity 4 /b /wait bench.test.exe -test.run=^$ -test.bench=. -test.benchtime=300ms -test.count=10 > go-bench-2026-11-02.txt
```

The raw output is [`go-bench-2026-11-02.txt`](benchmarks/go-bench-2026-11-02.txt), and the console
log with the load samples is [`go-bench-2026-11-02.log`](benchmarks/go-bench-2026-11-02.log).

## Results

Time is the median of the 10 runs of the `ns/date` metric (one benchmark iteration covers the 64
inputs; B8's operation is one whole month grid, per BENCHMARK section 3). "0" means the benchmark
measured no allocation. The C# column is the .NET 10 column of the hub's
[`docs/benchmarks/dotnet.md`](https://github.com/sanamhub/nepdate/blob/main/docs/benchmarks/dotnet.md),
measured on the same machine on 2026-11-02.

| ID | nepdate ns/op | std ns | B/op | allocs/op | gonepali ns/op | agrees | C# ns/op | C# B/op |
|---|---:|---:|---:|---:|---:|---|---:|---:|
| B1 BS to AD | 5.14 | 0.08 | 0 | 0 | 110.35 | no (3 of 64) | 13.50 | 0 |
| B2 AD to BS | 4.09 | 0.02 | 0 | 0 | 48.27 | no (4 of 64) | 20.10 | 0 |
| B3 add 1000 days | 3.97 | 0.02 | 0 | 0 | 210.40 | n/a (convert) | 19.05 | 0 |
| B4 add 13 months | 3.56 | 0.04 | 0 | 0 | | | 20.19 | 0 |
| B5 long, Nepali, weekday | 24.21 | 0.18 | 0 | 0 | | | 150.41 | 435 |
| B5-string `Long` | 42.01 | 1.01 | 4,144 | 64 | | | 150.41 | 435 |
| B6 strict parse | 7.63 | 0.06 | 0 | 0 | | | 22.42 | 0 |
| B7 patro info | 6.91 | 0.04 | 0 | 0 | | | 48.21 | 51 |
| B8 patro month grid | 246.05 | 3.79 | 0 | 0 | | | 3,112.76 | 1,429 |

The B3 gonepali row converts to AD, adds 1000 days with `time.AddDate`, and converts back, because
go-nepali has no O(1) add-days. Its agrees column is not measured; `TestAgrees` covers B1 and B2.

## Targets

- BENCHMARK section 6 (release blocker): B1, B2 and B7 within 1.5 times C# NepDate. B1 is 0.38
  times, B2 is 0.20 times, B7 is 0.14 times C#, all with 0 B/op. Passes.
- ADR-0002 section 7 (reported, not blocking): every B operation faster than C# NepDate. B5-string
  allocates but is faster than C# B5. B8 is 246.05 ns against 3,112.76 ns. Passes.

## Agreement

The agrees counts come from `go test -run TestAgrees -v .` in `benchmarks/`, which compares
gonepali's B1 and B2 against `nepdate` on the 64 inputs. It differs on 3 B1 dates and 4 B2 dates:
BS 2086-07-23, 2088-01-04 and 2089-06-10, plus the rejected AD 1943-04-14 on B2. The differing
years are listed in [`docs/competitors.md`](competitors.md).
