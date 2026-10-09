# ADR-0003: API conventions, errors and parity mapping

- Status: Accepted
- Date: 2026-09-25
- Deciders: Sanam Pakuwal

## Decision

**Conventions** (full surface in `docs/API.md`):
- Constructors return `(Date, error)`. `Must*` exists only for constants and tests.
- Methods take and return values, never pointers (except `UnmarshalText`).
- Reuse the standard library where the meaning is identical: `time.Weekday`, `time.Month` for AD,
  `iter.Seq`, `encoding.TextMarshaler`/`TextAppender`. No `MarshalJSON`: text marshalling already
  covers JSON v1 and v2 as a string.
- Every string-returning formatter has an `Append*` twin (the `time.AppendFormat` pattern).
- `FromTime` uses `t.Date()` in `t`'s own location. `Today(loc)` needs an explicit location, because
  a server in UTC is 5 h 45 min behind Nepal. The library never imports `time/tzdata` (450 KB); the
  docs show `time.FixedZone("NPT", 5*3600+45*60)` for callers without a zone database.

**Errors:** exported sentinels checked with `errors.Is`. Parse failures in the syntax step
(PARSING §4 first five kinds) are `*ParseError` with a `Kind`; a parsed date that fails
validation returns the same sentinel as `New` (`ErrOutOfRange`, `ErrInvalidMonth`,
`ErrInvalidDay`). Sentinels are package-level, so returning one does not allocate; a
`*ParseError` costs one allocation, on the failure path only. No `fmt.Errorf` in library code.

**Parity** (`shared/spec/PARITY.md`, applies unchanged):

| ID | Go |
|---|---|
| D-01 | zero `Date` is "no date": `IsZero()`, operations return `ErrOutOfRange` |
| D-02, D-03 | `AddMonths(n int, Overflow)`, `AddDays(n int)`; integers only |
| D-04 | `Month.ShortName` returns `Asa` / `Aso` |
| D-05, D-06 | `Parse` (strict), `ParseLenient` |
| D-07 | `DaysUntil(other) int` |
| D-08 | `FiscalYear` type with `Start`, `End`, `QuarterRange` |
| D-09 | `Range` with `Len`, `Contains`, `Days() iter.Seq[Date]` |
| D-10 | not ported |
| D-11 | text marshalling as `"YYYY-MM-DD"` |
| D-12 | `Date.MonthLength`; year length via `YearRange(y).Len()` |
| D-13 | package `github.com/sanamhub/nepdate-go/patro` (hub ADR-0004, ADR-0001 §1) |
| D-14 | not ported; `Today(loc)` and `AddDays(±1)` |
| D-15 | `Lang` parameter on every formatter |

A Go-only deviation is not allowed; propose it to the hub.
