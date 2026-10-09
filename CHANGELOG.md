# Changelog

All notable changes to this module are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
