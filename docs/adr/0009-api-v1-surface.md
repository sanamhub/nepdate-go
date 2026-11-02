# ADR-0009: full API v1 surface

- Status: Accepted
- Date: 2026-11-03
- Deciders: Sanam Pakuwal

## Decision

This ADR completes the Go port against the shared API inventory and supersedes the parts of
ADR-0003 that it changes.

**Dates.** `Date.YearLength` returns the BS year length, 364 to 367, instead of Gregorian
leap-ness. It replaces the old advice to call `YearRange(y).Len()`. `Date.IsToday`,
`Date.IsYesterday` and `Date.IsTomorrow` take a `*time.Location` like `Today`, so the clock is
never read from a hidden default and the old note that they are not ported no longer applies.

**Zero zone.** A nil location means Asia/Kathmandu in both `Today` and `Date.Time`. The old
`Time(nil)` result, midnight UTC, is still available as `Time(time.UTC)`. One nil default in the
package is easier to remember than two.

**Conversion.** `FromYMD` and `Date.YMD` cover the `YYYYMMDD` integer form. The ordered batch
converters are `ToNepaliDates`, `ToEnglishDates` and `ToEnglishDatesFromStrings`; each keeps the
input order and stops at the first error.

**Parsing.** `TryParse` and `TryParseLenient` return `(Date, bool)` for callers that do not need
the error value.

**Serialization.** A `Date` implements both the text hooks and `json.Marshaler` /
`json.Unmarshaler`. Values are JSON strings in `YYYY-MM-DD` form. JSON `null` is rejected with
`ErrOutOfRange`, because the zero `Date` means "no date"; an optional field should be a `*Date`.
This supersedes the "no `MarshalJSON`" sentence in ADR-0003.

**Ranges.** `Range` keeps the inclusive, non-empty model. It gains the factories, containment of
a range, overlap, adjacency, intersection, union, difference, month and fiscal-quarter splits,
the interval and working-day iterators, and both string forms. `Intersect` returns
`(Range, bool)`, since there is no empty range. The set operations that return a slice allocate
that slice; the iterators do not allocate per date. This supersedes the "no set operations"
sentence in ADR-0003.

All helpers stop at the data end instead of advancing past 2199-12-30, so a valid range never
fails because it touches the boundary.

## Consequences

The nil-zone, JSON-null and trailing-backslash changes are breaking for 0.x and are recorded in
the changelog with migration lines. The new surface adds no dependencies and keeps the core
package free of patro imports.
