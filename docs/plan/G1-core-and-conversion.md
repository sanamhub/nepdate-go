# G1: Core types and BS ↔ AD conversion

Read first: ADR-0002 (all of it: the table layout and the hot-path rules are decided there),
ADR-0003, ADR-0004, `docs/API.md` (Date, names, errors), `shared/spec/ALGORITHM.md` §1–7,
`shared/spec/PARITY.md`, `shared/spec/PORT-PLAN.md` phase 1 and "C# test catalogue".

## Tasks

### G1-01 Generated calendar tables
Extend `internal/gen` to write into `calendar_gen.go` (ADR-0002 §2):

- `bsMonthStart [3589]int32`: entry `k = (y-1901)*12 + (m-1)` is the serial (ALGORITHM §2) of BS
  `y-m-01`; entry 3588 is 109212.
- `adMonthStart [3601]int32`: entry `k = (y-1844)*12 + (m-1)` is the serial of AD `y-m-01`, for
  AD 1844-01 to 2144-01 (entries before 1844-04 are negative: `adMonthStart[0] == -101`).
- The month-estimate constants `bsEstMul`, `bsEstAdd`, `adEstMul`, `adEstAdd` with
  `estShift = 20` (ADR-0002 §3). The generator searches for them and **verifies** them: for every
  serial 0..=109211, `true month index - ((s*mul + add) >> 20)` is 0 or 1. If no constants pass,
  it exits 1 with a message; it never writes unverified constants.
- `const epochUnixDays = -45920`, `minYear = 1901`, `maxYear = 2199`, `maxSerial = 109211`.

One comment per year row (`// 2081`). Arrays are plain composite literals of integer constants so
the compiler lays them out as static data (ADR-0002 §4); no `init()` function in the package.
AC: `bsMonthStart[3588] == 109212`; `bsMonthStart[(2081-1901)*12] == 65746` (BS 2081-01-01 is
AD 2024-04-13); `adMonthStart[3600]` is the serial of AD 2144-01-01; `go run ./internal/gen
-check` exits 0; a unit test in `internal/gen` feeds constants that are off by one and gets an
error.

### G1-02 `Month`, `Lang`, errors (`names.go`, `errors.go`)
Per API. `Month.String()` returns the English name; `Name`, `ShortName` and `String` return `""`
for a `Month` outside 1..12. Name tables are package-level `[12]string` arrays of constants.
`ParseError.Error()` builds its message by concatenation, without `fmt`.
AC: `errors.Is(err, ErrInvalidDay)` works for errors returned by `New`; `Month(13).String() == ""`;
`Month(0).Name(Nepali) == ""`.

### G1-03 `Date`, `New`, accessors, validation (`date.go`)
`Date` is `struct{ year uint16; month, day uint8 }` (ADR-0002 §1). `New` checks year
1901..=2199, then month 1..=12, then day 1..=`bsMonthStart[k+1]-bsMonthStart[k]` (ALGORITHM §3
precedence). The zero `Date` follows D-01: `IsZero()` true, accessors 0, `MonthLength()` 0.
AC: `New(2081, Shrawan, 32)` is valid; `New(2081, Shrawan, 33)` returns `ErrInvalidDay`;
`New(2081, 13, 1)` returns `ErrInvalidMonth`; `New(1900, 13, 40)` and `New(2200, 1, 1)` return
`ErrOutOfRange`; `Date{}.IsZero()` and `Date{}.Year() == 0`; `MustNew(2081, Shrawan, 33)` panics;
`unsafe.Sizeof(Date{}) == 4` (in a test file; the library itself never imports `unsafe`).

### G1-04 Serial conversion (`serial.go`, unexported)
`toSerial(Date) int32` is `bsMonthStart[k] + day - 1`. `fromSerial(int32) Date` (precondition
`0 <= s <= maxSerial`) is the estimate plus one correction of ADR-0002 §3. No loop, no binary
search, no closure.
AC: `toSerial(Min) == 0`, `toSerial(Max) == 109211`; `fromSerial(toSerial(d)) == d` for every
date (covered by G1-07); `go build -gcflags=-m` reports `can inline toSerial` and
`can inline fromSerial` (paste the lines in the PR).

### G1-05 Gregorian (`gregorian.go`)
- `FromGregorian(y, m, d)`: Gregorian validity first (`gregLen [13]uint8` plus the leap rule,
  ALGORITHM §3) → `ErrInvalidGregorian`; then `y` outside 1844..=2143 → `ErrOutOfRange` before any
  arithmetic (a huge `y` can't overflow a 32-bit `int`); then
  `s = adMonthStart[(y-1844)*12+m-1] + d - 1`, outside 0..=109211 → `ErrOutOfRange`; then
  `fromSerial(s)`.
- `Gregorian()`: AD month estimate on `adMonthStart` plus one correction (ADR-0002 §3).
- `Weekday()` = `time.Weekday((s + 4) % 7)`, `YearDay()` = `s - bsMonthStart[(y-1901)*12] + 1`,
  `FirstOfMonth`, `LastOfMonth`, `Compare`, `Before`, `After` (compare `year`, then `month`, then
  `day`; no serial needed).

AC: AD 1844-04-11 is BS 1901-01-01 and AD 2143-04-15 is BS 2199-12-30; AD 1844-04-10 and
2143-04-16 return `ErrOutOfRange`; AD 2023-02-30 and 2023-02-29 return `ErrInvalidGregorian`;
`FromGregorian(math.MaxInt, 1, 1)` returns `ErrOutOfRange`; the five rows of ALGORITHM §9 match.

### G1-06 Reference oracle (`reference_test.go`)
The literal pseudocode of ALGORITHM §4 (`to_serial` month loop, `from_serial` binary search over
year starts) and §5 (Hinnant `days_from_civil`, `civil_from_days`), line for line, in a test
file. These are the spec; the fast path in G1-04 and G1-05 must equal them.
AC: the file has a comment pointing to ALGORITHM §4 and §5 above each function; the reviewer can
put the spec next to it and match lines one to one.

### G1-07 Tests
- `vectors_test.go`: `shared/spec/vectors/month-boundaries.csv` (7176 rows: BS, AD, weekday,
  length) and `shared/spec/vectors/csharp-golden/dates.tsv` (1018 rows) columns `ad`, `weekday`,
  `day_of_year`, `month_length`. The test fails if a file has fewer rows than stated here.
- `exhaustive_test.go`: for every serial 0..=109211, `fromSerial` equals the reference
  `from_serial`, `toSerial` of the result gives the serial back, `Gregorian()` equals the reference
  `civil_from_days`, and `FromGregorian` of that AD date gives the same `Date`.
- `alloc_test.go`: `testing.AllocsPerRun(100, f) == 0`, one subtest each, for `New`,
  `FromGregorian`, `Gregorian`, `Weekday`, `YearDay`, `MonthLength`, `Compare`, `Month.Name`,
  `WeekdayName`.
- `race_test.go`: `TestConcurrentReads` (ADR-0004) over `New`, `FromGregorian`, `Gregorian`.
- `csharp_test.go`: `TestCSharp` with every catalogue row tagged `1` (ADR-0004 naming rule).
  Expected sources (the catalogue's `phase` column decides): `NepaliDateConstructionTests`,
  `NepaliDatePropertiesTests`, `DictionaryIntegrityTests`, `NepaliDateComparableTests`, and the
  construction, comparison, `DayOfWeek` and min/max rows of `OptimizationVerificationTests`.
  `GetHashCode_*` rows become "equal dates are equal map keys". `IsLeapYear_*` is D-12.
  If `shared/spec/vectors/csharp-tests/catalog.tsv` does not exist yet (hub H5), do the rest of
  the phase and report this AC as blocked at the checkpoint.

## Definition of Done
Roadmap DoD + milestone M1.

## Checkpoint (agent stops here for review)

The reviewer checks:

- [ ] `go run ./internal/ci` ends with `ci: 5/5 gates passed` (paste the line).
- [ ] `go run ./internal/gen -check` exits 0, and `calendar_gen.go` has no `func init`.
- [ ] `reference_test.go` matches ALGORITHM §4 and §5 line for line.
- [ ] `go test -run Exhaustive -v .` passes; note how long it takes (expect under 1 s).
- [ ] `go test -run Alloc -v .`: every subtest reports 0.
- [ ] `go test -v -run TestCSharp . | grep -c -- '--- '` equals the number of catalogue rows tagged
      `1` (or the AC is reported as blocked on H5).
- [ ] `go build -gcflags=-m . 2>&1 | grep -E 'toSerial|fromSerial|escapes'`: both functions inline,
      and nothing in `date.go`, `serial.go` or `gregorian.go` escapes to the heap.
- [ ] No exported identifier beyond `docs/API.md` (`go doc -all .` against the API file).
