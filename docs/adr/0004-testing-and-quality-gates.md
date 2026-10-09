# ADR-0004: Testing and quality gates

- Status: Accepted
- Date: 2026-09-25 (revised 2026-10-07: patro rename, allocation and benchmark gates)
- Deciders: Sanam Pakuwal

## Decision

`go run ./internal/ci` runs, in order, stopping at the first failure:

1. `gofmt -l .` is empty; `go vet ./...`.
2. `golangci-lint run` (v2 config, `.golangci.yml`: `version: "2"`, `linters.default: standard`
   plus `staticcheck`, `revive`, `errorlint`, `gocritic`, `perfsprint`, `prealloc`, `gosec`,
   `misspell`, `unparam`, `exhaustive`; formatters `gofumpt`, `goimports`). Skipped with a warning
   when not installed locally; CI always runs it.
3. `go test -race -shuffle=on ./...`, then `go test -cover . ./patro`: the program reads each
   package's `coverage: N% of statements` line and fails if `nepdate` or `patro` is below
   90.0 % (org target 80 %). The race detector needs cgo and does not run on `GOARCH=386`, so
   when `go env CGO_ENABLED` is not `1` or `go env GOARCH` is not `amd64` or `arm64`, the program
   prints a warning and runs the same command without `-race`. (On the reference Windows machine
   `CGO_ENABLED` is `0` because no C compiler is installed, so `-race` runs only in CI on Linux
   and macOS.)
4. `go run ./internal/gen -check` (generated tables match `shared/data/`, ADR-0001 §5).
5. `govulncheck ./...` (CI via `golang/govulncheck-action`). Skipped with a warning when not
   installed locally.
6. Core isolation (hub ADR-0004), added in G4: `go list -deps .` must not list
   `github.com/sanamhub/nepdate-go/patro`; `go build -o <temp dir>/coreonly ./internal/coreonly`
   (a `package main` that imports only `nepdate`) produces a binary whose bytes don't contain
   `World Day Against Human Trafficking`; and `go build -o <temp dir>/withpatro ./internal/withpatro`
   (imports `patro` and prints the events of 2081-04-15) produces a binary whose bytes do contain
   it, so the check can't pass by accident.

Tests:

| Kind | What | Where |
|---|---|---|
| Vectors | every file in `shared/spec/vectors/`, read from disk in `_test.go` files. `go test` runs in the package directory, so the root package reads `shared/spec/vectors/...` and `patro` reads `../shared/spec/vectors/...` | G1 to G4 |
| C# catalogue | one test per row of `shared/spec/vectors/csharp-tests/catalog.tsv` tagged with the phase, named after the row `id` (rule below), or a comment naming the deviation or `n/a` | G1 to G4 |
| Exhaustive | every serial 0..=109211: the fast path equals the literal ALGORITHM §4 and §5 reference functions in `reference_test.go`, and BS → AD → BS round-trips | G1 |
| Fuzz | `testing.F` for `Parse`, `ParseLenient`, `UnmarshalText` (never panic; an error is a `*ParseError` or one of the three validation sentinels; a parsed date formats and re-parses to itself), and `FuzzAddDays` (adding `n` then `-n` gives the date back). Corpus under `testdata/fuzz/` | G2, G3, G4 |
| Allocations | `testing.AllocsPerRun(100, f) == 0` for every path ADR-0002 §5 lists, one subtest per path, in `alloc_test.go` of each package | G1 to G4 |
| Race | `TestConcurrentReads`: 8 goroutines run conversions, formatting, parsing and `patro.Of` over all 64 benchmark dates for 1000 rounds and compare with a single-goroutine run. Run under `-race` in CI. The tables are never written after start-up, so this guards against a future cache or lazy initialisation | G1 (core), G4 (patro) |
| Examples | `Example*` functions with `// Output:` for every README snippet | G5 |
| Benchmarks | B1 to B8 of `shared/spec/BENCHMARK.md`, `for b.Loop()` with `b.ReportAllocs()`; compared with `benchstat`. Not run per PR | G5 |

C# catalogue test names: Go test names can't contain `.` or `#`, so the row id
`NepaliDateConstructionTests.Constructor_ValidDate#3` becomes the subtest
`TestCSharp/NepaliDateConstructionTests_Constructor_ValidDate_3`: one table-driven
`TestCSharp` per package, one subtest per row, `.` and `#` replaced by `_`. A row that is not
ported has a table entry with `skip: "D-xx"` or `skip: "n/a: <reason>"`, so `go test -v -run
TestCSharp` lists every row id exactly once. `internal/ci` does not parse it; the reviewer checks the
count at each checkpoint.

CI: `ubuntu-latest`, `windows-latest`, `macos-latest` × Go 1.26 and 1.27; also `GOARCH=386` on
Linux, because `int` is 32 bits there. Actions pinned by SHA, `permissions: contents: read`.
