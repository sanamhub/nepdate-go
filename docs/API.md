# `nepdate` (Go): exported API contract

The complete exported surface of `github.com/sanamhub/nepdate-go` and its patro package.
Behaviour is fixed by `shared/spec/` (`ALGORITHM.md`, `FORMATTING.md`, `PARSING.md`, `PATRO.md`,
`PARITY.md`). Conventions are in ADR-0003.

```go
import (
	nepdate "github.com/sanamhub/nepdate-go"
	"github.com/sanamhub/nepdate-go/patro"
)
```

Go 1.24 or later. The core package has no dependencies. The zero `Date` is "no date": `IsZero`
reports true, its accessors return zero values, and every operation that can fail returns
`ErrOutOfRange`. A valid date is BS 1901-01-01 to 2199-12-30, which is AD 1844-04-11 to
2143-04-15.

## 1. Dates, identity and comparison

```go
type Date struct{ /* year uint16; month, day uint8 */ }

var (
	Min = Date{1901, 1, 1}
	Max = Date{2199, 12, 30}
)

func New(year int, month Month, day int) (Date, error)
func MustNew(year int, month Month, day int) Date
func FromGregorian(year int, month time.Month, day int) (Date, error)
func FromTime(t time.Time) (Date, error)
func Today(loc *time.Location) (Date, error)

func (d Date) IsZero() bool
func (d Date) Year() int
func (d Date) Month() Month
func (d Date) Day() int
func (d Date) Gregorian() (year int, month time.Month, day int)
func (d Date) Time(loc *time.Location) time.Time
func (d Date) Weekday() time.Weekday
func (d Date) YearDay() int
func (d Date) MonthLength() int
func (d Date) YearLength() int
func (d Date) FirstOfMonth() Date
func (d Date) LastOfMonth() Date
func (d Date) IsToday(loc *time.Location) bool
func (d Date) IsYesterday(loc *time.Location) bool
func (d Date) IsTomorrow(loc *time.Location) bool
func (d Date) FiscalYear() FiscalYear
func (d Date) Quarter() Quarter
func (d Date) Compare(other Date) int
func (d Date) Before(other Date) bool
func (d Date) After(other Date) bool
```

`New` validates the year, then the month, then the day. `MustNew` panics with `New`'s error and
is meant for constants and tests. `FromGregorian` rejects a nonexistent AD date with
`ErrInvalidGregorian` and an out-of-range one with `ErrOutOfRange`. `FromTime` uses `t.Date()` in
`t`'s own location, so `FromTime(t)` and `FromTime(t.UTC())` can differ by a day. `Today` reads
the clock in `loc`; nil means Asia/Kathmandu, or the fixed zone NPT (UTC+05:45) when the system
has no time zone database.

`YearDay` is the 1-based day of the BS year. `MonthLength` is 29 to 32. `YearLength` is the sum
of the twelve BS month lengths, 364 to 367; it is not Gregorian leap-ness. `Weekday` reuses
`time.Weekday` with Sunday as 0. `Compare` returns -1, 0 or +1 and matches `cmp.Compare`.
`Time` returns midnight at the start of the date in `loc`; nil means Asia/Kathmandu, the same
default as `Today`, and an explicit `time.UTC` gives midnight UTC.

`Date` is comparable: `==`, `!=`, `<`, `<=`, `>`, `>=` and map keys work. The field order is not
part of the contract.

```go
d, err := nepdate.New(2081, nepdate.Shrawan, 15)
if err == nil {
	fmt.Println(d, d.Weekday(), d.YearLength())
}
// 2081/04/15 Tuesday 366
```

## 2. Conversion and serialization

```go
func FromYMD(n int) (Date, error)
func (d Date) YMD() int

func ToNepaliDates(times []time.Time) ([]Date, error)
func ToEnglishDates(dates []Date) ([]time.Time, error)
func ToEnglishDatesFromStrings(texts []string) ([]time.Time, error)

func (d Date) MarshalText() ([]byte, error)
func (d Date) AppendText(b []byte) ([]byte, error)
func (d *Date) UnmarshalText(b []byte) error
func (d Date) MarshalJSON() ([]byte, error)
func (d *Date) UnmarshalJSON(data []byte) error
```

The integer form is `YYYYMMDD`: `FromYMD(20810415)` is 2081-04-15, and `YMD` returns it.
`FromYMD` uses `New`'s errors.

The three batch functions keep the input order. `ToNepaliDates` converts each time in its own
location. `ToEnglishDates` and `ToEnglishDatesFromStrings` return midnight UTC on each Gregorian
day. A batch stops at the first error and returns nil with that error; one bad string fails the
whole call.

The canonical text form is `YYYY-MM-DD`. A `Date` implements `encoding.TextMarshaler`,
`encoding.TextAppender` and `encoding.TextUnmarshaler`, so it works in JSON, YAML and map keys.
It also has explicit JSON methods: values marshal as a quoted string, `null` is rejected with
`ErrOutOfRange`, and a non-string JSON value is rejected with a `*ParseError`. An optional field
should be a `*Date`; nil marshals as `null`.

```go
d, _ := nepdate.FromYMD(20810415)
payload, _ := json.Marshal(struct{ D nepdate.Date }{d})
fmt.Println(string(payload)) // {"D":"2081-04-15"}
fmt.Println(d.YMD())         // 20810415
```

## 3. Arithmetic

```go
type Overflow uint8
const (
	Clamp Overflow = iota
	Spill
)

type Diff struct {
	Negative bool
	Years, Months, Days int
	TotalDays int
}

func (d Date) AddDays(n int) (Date, error)
func (d Date) AddMonths(n int, o Overflow) (Date, error)
func (d Date) AddYears(n int, o Overflow) (Date, error)
func (d Date) DaysUntil(other Date) int
func (d Date) Diff(other Date) Diff
func (d Date) Next() (Date, bool)
func (d Date) Prev() (Date, bool)
```

`AddMonths` takes whole months. `Clamp` moves a day past the target month end back to its last
day; `Spill` carries the extra days into the following month. `AddYears(n, o)` is
`AddMonths(12*n, o)`. Negative month or year counts subtract by the same rules.

`DaysUntil(other)` is `other - d`, positive when `other` is later. It is the named day
difference and the opposite direction from C# `Subtract`; Go has no operator to make the
distinction. `Diff` returns calendar years, months and days plus the signed total, which equals
`DaysUntil`.

`Next` and `Prev` return `(Date{}, false)` at `Max` and `Min`.

```go
d := nepdate.MustNew(2081, nepdate.Shrawan, 32)
clamped, _ := d.AddMonths(2, nepdate.Clamp)
spilled, _ := d.AddMonths(2, nepdate.Spill)
fmt.Println(clamped, spilled) // 2081/06/30 2081/07/02
```

## 4. Formatting

```go
type DateOrder uint8
const (
	YMD DateOrder = iota
	YDM
	MYD
	MDY
	DYM
	DMY
)

type Separator uint8
const (
	Slash Separator = iota
	Backslash
	Dot
	Underscore
	Dash
	Space
)

type LongOptions struct {
	Weekday  bool
	OmitYear bool
	NoPad    bool
}

func (d Date) String() string
func (d Date) Short(o DateOrder, sep Separator, pad bool, lang Lang) string
func (d Date) AppendShort(b []byte, o DateOrder, sep Separator, pad bool, lang Lang) []byte
func (d Date) Long(opts LongOptions, lang Lang) string
func (d Date) AppendLong(b []byte, opts LongOptions, lang Lang) []byte
func (d Date) Format(pattern string, lang Lang) string
func (d Date) AppendFormat(b []byte, pattern string, lang Lang) []byte
```

`String` is `YYYY/MM/DD`. `Short` writes the three fields in the chosen order with the chosen
separator; `pad` keeps the year at four digits and the month and day at two. `Long` writes
`[Weekday, ]Month day[, year]`. `Format` implements the standard patterns `""`, `G`, `g`, `d`,
`D` and `s`, and the custom pattern language: runs of `y`, `M` and `d`, `\x` escapes, and
`'...'` literals. A pattern ending in `\` writes a literal backslash. `MMM` uses `Asa` and `Aso`
for Ashad and Ashoj.

Every string form has an `Append` twin that writes into `b` without allocating when `b` has
room.

```go
d := nepdate.MustNew(2081, nepdate.Shrawan, 15)
fmt.Println(d.Short(nepdate.DMY, nepdate.Dash, false, nepdate.English)) // 15-4-2081
fmt.Println(d.Long(nepdate.LongOptions{Weekday: true}, nepdate.Nepali))
// मङ्गलवार, साउन १५, २०८१
```

## 5. Parsing

```go
type ParseErrorKind uint8
const (
	InvalidCharacter ParseErrorKind = iota + 1
	WrongGroupCount
	NumberTooLong
	Ambiguous
	Unrecognized
)

type ParseError struct {
	Input string
	Kind  ParseErrorKind
}
func (e *ParseError) Error() string

func Parse(s string) (Date, error)
func ParseLenient(s string) (Date, error)
func TryParse(s string) (Date, bool)
func TryParseLenient(s string) (Date, bool)
```

`Parse` accepts year, month and day in that order, ASCII or Devanagari digits (mixed is
allowed), at most four digits per group, and the separators `-`, `/`, `.`, `_`, `\`, space, `|`
and `।`. `ParseLenient` also accepts a closed list of month names, era markers, `गते` and
`मिति`, and a four-digit year first or last. It never guesses, so a two-digit year is
`Ambiguous`.

`TryParse` and `TryParseLenient` return `(Date{}, false)` instead of an error. A syntax failure
is a `*ParseError`; a text that parses but names an invalid date returns `ErrOutOfRange`,
`ErrInvalidMonth` or `ErrInvalidDay`.

```go
d, ok := nepdate.TryParseLenient("15 Shrawan 2081")
fmt.Println(d, ok) // 2081/04/15 true
```

## 6. Ranges

```go
type Range struct{ /* start, end Date */ }

func NewRange(start, end Date) (Range, bool)
func SingleDay(d Date) Range
func FromDayCount(start Date, days int) (Range, error)
func MonthRange(year int, month Month) (Range, error)
func YearRange(year int) (Range, error)
func CurrentMonth(loc *time.Location) (Range, error)
func CurrentFiscalYear(loc *time.Location) (Range, error)
func CurrentCalendarYear(loc *time.Location) (Range, error)

func (r Range) Start() Date
func (r Range) End() Date
func (r Range) Len() int
func (r Range) Contains(d Date) bool
func (r Range) ContainsRange(other Range) bool
func (r Range) Overlaps(other Range) bool
func (r Range) IsAdjacentTo(other Range) bool
func (r Range) Intersect(other Range) (Range, bool)
func (r Range) Union(other Range) Range
func (r Range) Except(other Range) []Range
func (r Range) SplitByMonth() []Range
func (r Range) SplitByFiscalQuarter() []Range
func (r Range) DatesWithInterval(interval int) (iter.Seq[Date], error)
func (r Range) WorkingDays(excludeSunday bool) iter.Seq[Date]
func (r Range) WeekendDays(includeSunday bool) iter.Seq[Date]
func (r Range) Days() iter.Seq[Date]
func (r Range) String() string
func (r Range) Format(o DateOrder, sep Separator, lang Lang) string
```

A range is inclusive and cannot be empty: `NewRange` returns false when `start > end` or either
bound is the zero `Date`. `FromDayCount` requires at least one day. `CurrentMonth`,
`CurrentFiscalYear` and `CurrentCalendarYear` read the clock in `loc`, where nil means
Asia/Kathmandu.

`ContainsRange` accepts the zero Range as the contained range. `Overlaps`, `IsAdjacentTo` and
`Intersect` return false or no range instead of an empty range. `Union` is the smallest covering
range. `Except` returns zero, one or two ranges.

The first and last pieces of both splits are clipped to the range. A range that reaches
2199-12-30 yields its last piece and its last date; no helper throws or wraps because it touches
the data end. `DatesWithInterval` rejects an interval below 1 with `ErrOutOfRange`; check the
error before ranging over the sequence. A nil sequence is returned with that error.

`Range` is comparable, so `==`, `!=` and map keys use the two bounds. `String` is
`start - end` with the default date form. `Format` writes both dates with the same order,
separator and language.

```go
r, _ := nepdate.NewRange(
	nepdate.MustNew(2081, nepdate.Shrawan, 10),
	nepdate.MustNew(2081, nepdate.Shrawan, 20),
)
fmt.Println(r.Len(), r.Contains(nepdate.MustNew(2081, nepdate.Shrawan, 15)))
// 11 true
fmt.Println(r.Format(nepdate.YMD, nepdate.Dash, nepdate.English))
// 2081-04-10 - 2081-04-20
```

## 7. Fiscal year

```go
type FiscalYear int
func FiscalYearOf(d Date) FiscalYear
func (f FiscalYear) Start() (Date, error)
func (f FiscalYear) End() (Date, error)
func (f FiscalYear) Range() (Range, error)
func (f FiscalYear) QuarterRange(q Quarter) (Range, error)
func (f FiscalYear) Label(lang Lang) string

type Quarter uint8
const (
	Q1 Quarter = 1 + iota
	Q2
	Q3
	Q4
)
func (q Quarter) Months() [3]Month
```

A fiscal year starts on 1 Shrawan (month 4) and ends on the last day of Ashad (month 3) of the
next BS year. It is named by its starting year, so FY 2082/83 is `FiscalYear(2082)`. Q1 is
Shrawan to Ashoj, Q2 Kartik to Poush, Q3 Magh to Chaitra and Q4 Baishakh to Ashad of the next
year. `Date.FiscalYear()` and `Date.Quarter()` return the containing values.

`Label` writes `2082/83` in English or `२०८२/८३` in Nepali. An out-of-range start, end or
quarter returns `ErrOutOfRange`.

```go
fy := nepdate.MustNew(2081, nepdate.Shrawan, 15).FiscalYear()
r, _ := fy.QuarterRange(nepdate.Q1)
fmt.Println(fy.Label(nepdate.Nepali), r.Start(), r.End())
// २०८१/८२ 2081/04/01 2081/06/30
```

## 8. Patro

```go
package patro

type Info struct{ /* unexported */ }
func Of(d nepdate.Date) Info
func (i Info) Tithi() (Tithi, bool)
func (i Info) IsPublicHoliday() bool
func (i Info) Events() iter.Seq[Event]
func (i Info) EventCount() int
func (i Info) Event(n int) Event

type Tithi struct{ /* unexported */ }
func (t Tithi) Name(lang nepdate.Lang) string

type Event struct{ /* unexported */ }
func (e Event) Name(lang nepdate.Lang) string

func Coverage() nepdate.Range
func PublicHolidays() iter.Seq[nepdate.Date]
```

The patro package is separate and imports `nepdate`; the core never imports it. `Of` never fails:
outside the data range, and for the zero `Date`, it returns an empty `Info`. Tithi, holiday and
event names are empty when absent, and `Events` is in file order. The data covers BS 2077-01-01
to 2084-08-20; tithi names end at BS 2083-12-30. `Coverage` and `PublicHolidays` describe that
data.

```go
i := patro.Of(nepdate.MustNew(2081, nepdate.Baishakh, 1))
t, _ := i.Tithi()
fmt.Println(t.Name(nepdate.Nepali), i.IsPublicHoliday())
```

## 9. Names, languages, errors and options

```go
type Lang uint8
const (
	English Lang = iota
	Nepali
)
func (l Lang) AppendDigits(b []byte, n, minWidth int) []byte

type Month uint8
const (
	Baishakh Month = 1 + iota
	Jestha
	Ashad
	Shrawan
	Bhadra
	Ashoj
	Kartik
	Mangsir
	Poush
	Magh
	Falgun
	Chaitra
)
func (m Month) Name(lang Lang) string
func (m Month) ShortName(lang Lang) string
func (m Month) String() string

func WeekdayName(w time.Weekday, lang Lang) string
func WeekdayShortName(w time.Weekday, lang Lang) string

var (
	ErrOutOfRange       error
	ErrInvalidMonth     error
	ErrInvalidDay       error
	ErrInvalidGregorian error
)
```

`English` writes ASCII digits and names; `Nepali` writes Devanagari digits and names. Month names
and weekday names are empty outside their valid ranges. English month abbreviations are `Bai`,
`Jes`, `Asa`, `Shr`, `Bha`, `Aso`, `Kar`, `Man`, `Pou`, `Mag`, `Fal` and `Cha`.

The four sentinel errors are tested with `errors.Is`. `ErrOutOfRange` also covers the zero
`Date` and any operation whose result leaves the supported range. `ErrInvalidGregorian` covers a
nonexistent AD date such as 2023-02-29.

```go
fmt.Println(nepdate.Shrawan.Name(nepdate.Nepali))
// साउन
fmt.Println(nepdate.Ashad.ShortName(nepdate.English))
// Asa
```

## Zero values

| Value | Meaning |
|---|---|
| `Date{}` | no date, `IsZero() == true`; accessors return 0 or `""`, error-returning operations return `ErrOutOfRange`, relative predicates return false |
| `Range{}` | empty range; `Len() == 0`, `String() == ""`, the set operations return the zero of their result type |
| `Month(0)` | no month; names are `""` |
| `Separator` and `DateOrder` values outside their constants | `Short` and `Format` return `""` |
