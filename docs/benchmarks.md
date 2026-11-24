# Benchmarks (Go)

Time and allocation of the Go port on operations B1 to B8 of
[`shared/spec/BENCHMARK.md`](../shared/spec/BENCHMARK.md), against the one competitor that passes
[`docs/competitors.md`](competitors.md), `github.com/opensource-nepal/go-nepali` v0.4.1.

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
measured no allocation.

The comparison with C# NepDate, the original library, and the target check from `spec/BENCHMARK.md` section 6 are in the hub: [performance page](https://nepdate.com/docs/project/performance/).

| ID | nepdate ns/op | std ns | B/op | allocs/op | gonepali ns/op | agrees |
|---|---:|---:|---:|---:|---:|---|
| B1 BS to AD | 5.24 | 0.05 | 0 | 0 | 111.85 | no (3 of 64) |
| B2 AD to BS | 4.20 | 0.03 | 0 | 0 | 49.05 | no (4 of 64) |
| B3 add 1000 days | 4.03 | 0.01 | 0 | 0 | 212.85 | n/a (convert) |
| B4 add 13 months | 3.64 | 0.12 | 0 | 0 | | |
| B5 long, Nepali, weekday | 24.79 | 0.14 | 0 | 0 | | |
| B5-string `Long` | 42.92 | 0.70 | 4,144 | 64 | | |
| B6 strict parse | 7.84 | 0.04 | 0 | 0 | | |
| B7 patro info | 7.11 | 0.03 | 0 | 0 | | |
| B8 patro month grid | 255.55 | 2.83 | 0 | 0 | | |

The B3 gonepali row converts to AD, adds 1000 days with `time.AddDate`, and converts back, because
go-nepali has no O(1) add-days. Its agrees column is not measured; `TestAgrees` covers B1 and B2.

## Agreement

The agrees counts come from `go test -run TestAgrees -v .` in `benchmarks/`, which compares
gonepali's B1 and B2 against `nepdate` on the 64 inputs. It differs on 3 B1 dates and 4 B2 dates:
BS 2086-07-23, 2088-01-04 and 2089-06-10, plus the rejected AD 1943-04-14 on B2. The differing
years are listed in [`docs/competitors.md`](competitors.md).
