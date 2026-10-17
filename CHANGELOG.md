# Changelog

All notable changes to this module are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Strict parsing of the `YYYY/MM/DD` text form uses a fast path, so `Parse` and `ParseLenient` are
  faster on that form. Other forms and all errors are unchanged.

## [0.1.0] - 2026-10-17

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
  `Event`, `Coverage` and `PublicHolidays`, for BS 2077-01-01 to 2084-08-20. Package `nepdate`
  doesn't import it.
- Text marshalling as `YYYY-MM-DD` (`MarshalText`, `AppendText`, `UnmarshalText`), so a `Date`
  works with `encoding/json`, including as a map key.
- `FromTime`, `Date.Time` and `Today`.
