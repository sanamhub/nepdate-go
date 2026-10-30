# Changelog

All notable changes to this module are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- The published-package check now also runs weekly and daily, and reports a failure through the
  `published-check` issue.

## [0.1.0-beta.1] - 2026-10-21

First public pre-release. The API may still change before 0.1.0.

### Added

- `Date`, `New`, `MustNew`, `Min`, `Max` and the accessors `Year`, `Month`, `Day`, `IsZero`,
  `MonthLength`, `YearDay`, `Weekday`, `FirstOfMonth`, `LastOfMonth`.
- BS to AD and back: `FromGregorian` and `Date.Gregorian`, for BS 1901-01-01 to 2199-12-30.
- `Compare`, `Before`, `After`.
- `Month` and `Lang` with `Name`, `ShortName` and `String`; `WeekdayName`, `WeekdayShortName`.
- Errors `ErrOutOfRange`, `ErrInvalidMonth`, `ErrInvalidDay`, `ErrInvalidGregorian` and
  `ParseError`.
- Arithmetic: `AddDays`, `AddMonths` and `AddYears` with `Clamp` or `Spill`, `Next`, `Prev`,
  `DaysUntil` and `Diff`.
- Fiscal years: `FiscalYear`, `FiscalYearOf`, `Quarter`, with `Start`, `End`, `Range`,
  `QuarterRange` and `Label`.
- `Range` with `NewRange`, `MonthRange`, `YearRange`, `Len`, `Contains` and `Days`.
- Formatting: `String`, `Short`, `Long` and `Format`, each with an `Append` twin that doesn't
  allocate; `DateOrder`, `Separator`, `LongOptions`; `Lang.AppendDigits` for Devanagari digits.
- Parsing: strict `Parse` (ASCII or Devanagari digits) and `ParseLenient` for month names, era
  markers and day-first dates.
- Package `patro`: `Of`, `Info` with `Tithi`, `IsPublicHoliday`, `Events`, `EventCount` and
  `Event`, `Coverage` and `PublicHolidays`, for BS 2077-01-01 to 2084-08-20. Tithi names end
  earlier, at BS 2083-12-30 (AD 2027-04-13). Package `nepdate` doesn't import it.
- Text marshalling as `YYYY-MM-DD` (`MarshalText`, `AppendText`, `UnmarshalText`), so a `Date`
  works with `encoding/json`, including as a map key.
- `FromTime`, `Date.Time` and `Today`.
- Runnable examples for every exported function and type, including arithmetic, fiscal years,
  ranges, JSON marshalling and the patro queries.
- Go 1.24 or later. CI tests the minimum supported version, Go 1.24.13 (ADR-0008).
