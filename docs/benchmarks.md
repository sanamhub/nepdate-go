# Benchmarks (Go)

Time and allocation of the Go port on operations B1 to B8 of
[`shared/spec/BENCHMARK.md`](../shared/spec/BENCHMARK.md), against the four competitor Go libraries
of [`docs/competitors.md`](../competitors.md) and the C# NepDate baseline.

## Machine

| Item | Value |
|---|---|
| CPU | 12th Gen Intel Core i5-12500H (4 performance + 8 efficiency cores, 16 threads, base 3.1 GHz) |
| RAM | 31.7 GB |
| OS | Windows 11 Home 10.0.26300 |
| Power plan | Ultimate Performance |
| Affinity | logical CPU 2 (mask `0x4`), set with `start /affinity 4` |
| Machine load | other agents were running; CPU load sampled every 5 s during the run ranged 1 to 71 % (median about 22 %), so the numbers are noisier than the quiet-machine protocol of BENCHMARK section 1 |

## Versions

| Tool | Version |
|---|---|
| Go | go1.27.1 windows/amd64 |
| benchstat | golang.org/x/perf/cmd/benchstat |
| go-nepali | v0.4.1 |
| nepcal | v1.3.0 |
| go-bs | v0.7.0 |
| GoBikramSambat | v0.0.0-20200507040640-7d5f1ba6f8da |

## Command

From `cmd` in `benchmarks/`:

```
go test -c -o bench.test.exe .
start "" /high /affinity 4 /b /wait bench.test.exe -test.run=^$ -test.bench=. -test.benchtime=300ms -test.count=10 > bench.txt
benchstat -col /library bench.txt
```

On this machine the command was run through a one-line batch file because Git Bash mangles the
quoting, and the batch file writes the caret as `^^$` so cmd passes the literal `^$`. The raw
output is published as [`go-bench.txt`](go-bench.txt) for later `benchstat` comparison.

## Results

Time is `benchstat`'s median of the `ns/date` metric (one benchmark iteration covers the 64
inputs; B8's operation is one whole month grid, per BENCHMARK section 3). "0" means the benchmark
measured no allocation. The C# column is the .NET 10 column of the hub's
`docs/benchmarks/dotnet.md` (commit `1d7ca17`), never measured here. The numbers carry the machine
load recorded above; several rows have wide run to run spread (B5, B7 and B8).

| ID | nepdate ns/op | B/op | allocs/op | gonepali ns/op | agrees | nepcal ns/op | agrees | gobs ns/op | agrees | gobikramsambat ns/op | agrees | C# ns/op | C# B/op |
|---|---:|---:|---:|---:|---|---:|---|---:|---|---:|---|---:|---:|
| B1 BS to AD | 6.21 | 0 | 0 | 172.9 | no (3 of 64) | 1627.5 | no (2 of 64) | 379.85 | no (1 of 64) | 215.75 | no (4 of 64) | 13.29 | 0 |
| B2 AD to BS | 5.03 | 0 | 0 | 56.23 | no (4 of 64) | 6113.5 | no (2 of 64) | 349.3 | no (1 of 64) | 207.55 | no (4 of 64) | 18.64 | 0 |
| B3 add 1000 days | 4.67 | 0 | 0 | 282.5 | n/a (convert) | | | 759.7 | n/a (convert) | | | 17.50 | 0 |
| B4 add 13 months | 4.14 | 0 | 0 | | | | | | | | | 18.85 | 0 |
| B5 long, Nepali, weekday | 30.95 | 0 | 0 | | | | | | | | | 153.02 | 435 |
| B5-string `Long` | 73.98 | 4144 | 64 | | | | | | | | | 153.02 | 435 |
| B6 strict parse | 34.80 | 0 | 0 | | | | | | | | | 21.55 | 0 |
| B7 patro info | 11.83 | 0 | 0 | | | | | | | | | 47.93 | 51 |
| B8 patro month grid | 398.1 | 0 | 0 | | | | | | | | | 3,132.66 | 1,429 |

The B3 gonepali and gobs rows convert to AD, add 1000 days with `time.AddDate`, and convert back,
because neither has an O(1) add-days. Their agrees column is not measured; `TestAgrees` only
covers B1 and B2.

## Targets

- BENCHMARK section 6 (release blocker): B1, B2 and B7 within 1.5 times C# NepDate. B1 is 0.47
  times, B2 is 0.27 times, B7 is 0.25 times C#, all with 0 B/op. Passes.
- ADR-0002 section 7 (reported, not blocking): every B operation faster than C# NepDate. B6
  misses: 34.80 ns against C#'s 21.55 ns. A 2 s CPU profile of B6 (`-test.cpuprofile`) puts 68.8 %
  of the samples in `scanStrict` and 7.4 % in the inlined `devaAt`: the parser checks for
  Devanagari digits at every byte even though the B6 inputs are ASCII. The machine was loaded
  during the run, so the miss is reported, not optimised away.

## Agreement

The agrees counts come from `go test -run TestAgrees .`, which compares each competitor's B1 and
B2 against `nepdate` on the 64 inputs. The differing years are listed in
[`docs/competitors.md`](../competitors.md); the counts here match that table (for example gobs
differs in one of its 64 dates, and gobikramsambat in four).
