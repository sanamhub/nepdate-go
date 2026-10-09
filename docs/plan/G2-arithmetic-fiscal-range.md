# G2: Arithmetic, diff, fiscal year, range

Read first: `shared/spec/ALGORITHM.md` §8–8b, `docs/API.md` (arithmetic, fiscal, range),
`shared/spec/PARITY.md` D-02, D-03, D-07, D-08, D-09, `shared/spec/PORT-PLAN.md` phase 2.

Golden files are in `shared/spec/vectors/csharp-golden/`; `ERR` in a cell means the call returns
`ErrOutOfRange`. `add-days.tsv` and `add-months.tsv` include negative amounts.

## Tasks

### G2-01 `AddDays`, `Next`, `Prev`, `DaysUntil`
Through the serial: `fromSerial(toSerial(d) + n)` (ADR-0002 §3), so `AddDays` is two table reads
and one estimate, with no loop. `AddDays` returns `ErrOutOfRange` when `n < -109211` or `n > 109211` before
adding, so a huge `n` can't overflow a 32-bit `int` (386). `Next` and `Prev` return
`(Date{}, false)` at the ends.
AC: golden `add-days.tsv` (9162 rows) passes; `Max.AddDays(1)` and `Min.AddDays(-1)` return
`ErrOutOfRange`; `AddDays(math.MaxInt)` and `AddDays(math.MinInt)` return `ErrOutOfRange`;
`Min.DaysUntil(Max) == 109211` and `Max.DaysUntil(Min) == -109211`.
`FuzzAddDays(serial, n)`: for a valid date `d`, if `d.AddDays(n)` succeeds, adding `-n` gives `d`
back (ADR-0004).

### G2-02 `AddMonths`, `AddYears` with `Clamp` / `Spill`
ALGORITHM §8. The target month's length is `bsMonthStart[t+1] - bsMonthStart[t]`; no serial
conversion is needed. `AddYears` returns `ErrOutOfRange` when `|n| > 299` before multiplying by 12, and
`AddMonths` when `|n| > 3588`, so neither overflows a 32-bit `int`.
AC: golden `add-months.tsv` (10180 rows; column `clamp` with `Clamp`, column `overflow` with
`Spill`); the §8 examples (2081-04-32 +2 → Clamp 2081-06-30, Spill 2081-07-02; +5 → Clamp
2081-09-29, Spill 2081-10-03); `AddYears(math.MaxInt, Clamp)` returns `ErrOutOfRange`.

### G2-03 `Diff` (ALGORITHM §8a)
`Years`, `Months` and `Days` are magnitudes; `Negative` and `TotalDays` carry the sign.
The six ALGORITHM §8a examples, plus these extra rows computed from `shared/data/` with the §8a
pseudocode:

| a | b | Negative | Years | Months | Days | TotalDays |
|---|---|---|---|---|---|---|
| 2081-04-32 | 2081-06-30 | false | 0 | 2 | 0 | 61 |
| 2081-01-01 | 2082-01-01 | false | 1 | 0 | 0 | 366 |
| 2081-04-15 | 2083-02-10 | false | 1 | 9 | 26 | 663 |
| 2083-02-10 | 2081-04-15 | true | 1 | 9 | 26 | -663 |
| 2081-04-15 | 2081-04-15 | false | 0 | 0 | 0 | 0 |

AC: the §8a examples and the table above; `a.Diff(b).TotalDays == a.DaysUntil(b)` for 1000 random pairs from a
fixed seed.

### G2-04 `FiscalYear`, `Quarter` (ALGORITHM §8b)
AC: golden `dates.tsv` columns `fy_start`, `fy_end`, `quarter_start`, `quarter_end` (the quarter of
the row's date in its own fiscal year); `FiscalYear(2082).Label(English)` is `2082/83` and
`Label(Nepali)` is `२०८२/८३`; `FiscalYear(1900).Start()` and `FiscalYear(2199).End()` return
`ErrOutOfRange`.

### G2-05 `Range`
`NewRange`, `MonthRange`, `YearRange`, `Len`, `Contains`, `Days() iter.Seq[Date]` (stops when the
caller breaks). `Days` walks the serial range and calls `fromSerial` per step; it allocates
nothing, including the iterator itself, when the caller ranges over it directly.
AC: for every year 1901..2199, `YearRange(y).Len()` equals the sum of its 12 `MonthLength`s and is
in 364..367; ranging over `MonthRange(2081, Shrawan).Days()` yields 32 dates and stops after 3 when
the loop breaks at the third; `NewRange(Max, Min)` returns `false`; `YearRange(2200)` returns
`ErrOutOfRange`.

### G2-06 Allocation, race and C# catalogue tests
- `alloc_test.go` gains subtests (each `== 0`) for `AddDays`, `AddMonths` (both modes),
  `AddYears`, `DaysUntil`, `Diff`, `Next`, `Prev`, `FiscalYearOf`, `FiscalYear.Start`,
  `FiscalYear.QuarterRange`, `MonthRange`, `Range.Contains`, and a full `for d := range
  r.Days()` over `MonthRange(2081, Shrawan)`.
- `TestConcurrentReads` gains `AddDays`, `AddMonths` and `Diff`.
- `TestCSharp` gains every catalogue row tagged `2`. Expected sources:
  `NepaliDateArithmeticEdgeCaseTests`, `NepaliDateManipulationTests`, `FiscalYearTests`,
  `FiscalYearInstanceMethodTests` (D-08 shape: same dates, different calls),
  `NepaliDateRangeTests`, `NepaliDateRangeExtendedTests` and the subtraction rows of
  `OptimizationVerificationTests`. Rows for `Except`, `Union`, `Intersect`, `IsAdjacentTo`,
  `SplitByFiscalQuarter`, `WorkingDays`, `WeekendDays`, `DatesWithInterval` and `FromDayCount` are
  D-09; fractional `AddMonths`/`AddDays` rows are D-02/D-03; `Current*` rows (clock-dependent)
  are D-14 and get a `Today`-free equivalent where the assertion is about structure (for example
  "a fiscal year starts on Shrawan 1").

AC: every subtest reports 0 allocations; `go test -v -run TestCSharp .` lists every row tagged
`2` once.

## Checkpoint (agent stops here for review)

The reviewer checks:

- [ ] `go run ./internal/ci` ends with `ci: 5/5 gates passed`.
- [ ] Golden `add-days.tsv` (9162 rows) and `add-months.tsv` (10180 rows, both columns) pass, and
      the test fails if either file has fewer rows.
- [ ] The overflow guards: `AddDays(math.MaxInt)`, `AddYears(math.MaxInt, Clamp)` and the
      `GOARCH=386` CI job pass.
- [ ] `go test -run '^$' -fuzz '^FuzzAddDays$' -fuzztime 30s .` finds nothing (run locally).
- [ ] `go test -run Alloc -v .`: every new subtest reports 0.
- [ ] `TestCSharp` row count for phase `2` matches the catalogue; every skipped row names its
      D-ID.
- [ ] `Diff` matches the ALGORITHM §8a examples and the G2-03 table, and contains no loop over
      days.
