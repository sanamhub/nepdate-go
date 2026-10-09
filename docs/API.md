# `nepdate` (Go): exported API contract (v0.1)

The complete exported surface for v0.1. Implement exactly this; adding an exported identifier
needs this file changed in the same PR, approved by the human. Behaviour: `shared/spec/`
(`ALGORITHM.md`, `FORMATTING.md`, `PARSING.md`, `PATRO.md`, `PARITY.md`). Conventions: ADR-0003.

```go
package nepdate // import "github.com/sanamhub/nepdate-go"

// ---------- Date ----------
// Date is a Bikram Sambat calendar date. It is a small comparable value: use == and as a map key.
// The zero value is "no date" (D-01): IsZero reports true, accessors return 0, and every
// operation that returns an error returns ErrOutOfRange.
type Date struct{ /* year uint16; month, day uint8 (unexported) */ }

var (
    Min = Date{ /* 1901-01-01 */ }
    Max = Date{ /* 2199-12-30 */ }
)

func New(year int, month Month, day int) (Date, error)       // ErrOutOfRange, ErrInvalidMonth, ErrInvalidDay
func MustNew(year int, month Month, day int) Date             // panics; for constants and tests only
func FromGregorian(year int, month time.Month, day int) (Date, error) // ErrInvalidGregorian, ErrOutOfRange
func FromTime(t time.Time) (Date, error)                      // uses t.Date() in t's own location
func Today(loc *time.Location) (Date, error)                  // today in loc; nil means Asia/Kathmandu, or time.FixedZone("NPT", 20700) if tzdata lacks it
func Parse(s string) (Date, error)                            // strict, PARSING §1; *ParseError, or the New sentinels
func ParseLenient(s string) (Date, error)                     // PARSING §2; *ParseError, or the New sentinels

func (d Date) IsZero() bool
func (d Date) Year() int
func (d Date) Month() Month
func (d Date) Day() int
func (d Date) Weekday() time.Weekday                          // reuses the standard type; Sunday = 0
func (d Date) YearDay() int                                   // 1-based
func (d Date) MonthLength() int                               // 29..32
func (d Date) FirstOfMonth() Date
func (d Date) LastOfMonth() Date
func (d Date) Gregorian() (year int, month time.Month, day int)
func (d Date) Time(loc *time.Location) time.Time              // midnight in loc; nil means time.UTC

func (d Date) AddDays(n int) (Date, error)
func (d Date) AddMonths(n int, o Overflow) (Date, error)
func (d Date) AddYears(n int, o Overflow) (Date, error)
func (d Date) DaysUntil(other Date) int                       // other - d, signed (D-07)
func (d Date) Diff(other Date) Diff                           // ALGORITHM §8a
func (d Date) Next() (Date, bool)                             // (Date{}, false) at Max
func (d Date) Prev() (Date, bool)                             // (Date{}, false) at Min
func (d Date) Compare(other Date) int                         // -1, 0, +1 (cmp.Compare style)
func (d Date) Before(other Date) bool
func (d Date) After(other Date) bool

func (d Date) FiscalYear() FiscalYear
func (d Date) Quarter() Quarter

// Formatting. Each string method has an Append twin that writes into b and does not allocate.
func (d Date) String() string                                 // "YYYY/MM/DD"
func (d Date) Short(o DateOrder, sep Separator, pad bool, lang Lang) string
func (d Date) AppendShort(b []byte, o DateOrder, sep Separator, pad bool, lang Lang) []byte
func (d Date) Long(opts LongOptions, lang Lang) string
func (d Date) AppendLong(b []byte, opts LongOptions, lang Lang) []byte
func (d Date) Format(pattern string, lang Lang) string        // FORMATTING §5
func (d Date) AppendFormat(b []byte, pattern string, lang Lang) []byte

// Text encoding as "YYYY-MM-DD" (D-11). Covers encoding/json v1 and v2, YAML, map keys.
func (d Date) MarshalText() ([]byte, error)
func (d Date) AppendText(b []byte) ([]byte, error)            // encoding.TextAppender
func (d *Date) UnmarshalText(b []byte) error                  // strict parse

// ---------- names ----------
type Lang uint8
const ( English Lang = iota; Nepali )
func (l Lang) AppendDigits(b []byte, n int, minWidth int) []byte   // ASCII or Devanagari digits

type Month uint8
const ( Baishakh Month = 1 + iota; Jestha; Ashad; Shrawan; Bhadra; Ashoj; Kartik; Mangsir; Poush; Magh; Falgun; Chaitra )
func (m Month) Name(lang Lang) string                         // FORMATTING §1; "" outside 1..12, same for ShortName, String
func (m Month) ShortName(lang Lang) string                    // D-04: Asa / Aso
func (m Month) String() string                                // English name

func WeekdayName(w time.Weekday, lang Lang) string
func WeekdayShortName(w time.Weekday, lang Lang) string      // both return "" outside Sunday..Saturday

// ---------- errors (sentinels; test with errors.Is) ----------
var (
    ErrOutOfRange       = errors.New("nepdate: date outside BS 1901-01-01..2199-12-30")
    ErrInvalidMonth     = errors.New("nepdate: month must be 1..12")
    ErrInvalidDay       = errors.New("nepdate: day outside the month")
    ErrInvalidGregorian = errors.New("nepdate: invalid Gregorian date")
)
type ParseErrorKind uint8
const ( InvalidCharacter ParseErrorKind = iota + 1; WrongGroupCount; NumberTooLong; Ambiguous; Unrecognized )
type ParseError struct { Input string; Kind ParseErrorKind }
func (e *ParseError) Error() string

// ---------- arithmetic ----------
type Overflow uint8
const ( Clamp Overflow = iota; Spill )                        // ALGORITHM §8 (Clamp = C# default)
type Diff struct { Negative bool; Years, Months, Days int; TotalDays int }

// ---------- format options ----------
type DateOrder uint8
const ( YMD DateOrder = iota; YDM; MYD; MDY; DYM; DMY )
type Separator uint8
const ( Slash Separator = iota; Backslash; Dot; Underscore; Dash; Space )
type LongOptions struct { Weekday, OmitYear, NoPad bool }     // zero value = C# default (year shown, padded)

// ---------- fiscal ----------
type FiscalYear int                                           // starting BS year: FY 2082/83 = 2082
func FiscalYearOf(d Date) FiscalYear
func (f FiscalYear) Start() (Date, error)
func (f FiscalYear) End() (Date, error)
func (f FiscalYear) Range() (Range, error)
func (f FiscalYear) QuarterRange(q Quarter) (Range, error)
func (f FiscalYear) Label(lang Lang) string                   // "2082/83" / "२०८२/८३"
type Quarter uint8
const ( Q1 Quarter = 1 + iota; Q2; Q3; Q4 )
func (q Quarter) Months() [3]Month

// ---------- range ----------
type Range struct{ /* start, end Date (inclusive, start <= end) */ }
func NewRange(start, end Date) (Range, bool)                  // false if start > end or a zero Date
func MonthRange(year int, month Month) (Range, error)
func YearRange(year int) (Range, error)
func (r Range) Start() Date
func (r Range) End() Date
func (r Range) Len() int
func (r Range) Contains(d Date) bool
func (r Range) Days() iter.Seq[Date]                          // D-09: no set operations
```

```go
package patro // import "github.com/sanamhub/nepdate-go/patro"

// Separate package (hub ADR-0004, ADR-0001 §1). It imports nepdate; nepdate never imports it.

type Info struct{ /* unexported */ }
func Of(d nepdate.Date) Info                                  // never fails; empty outside coverage and for the zero Date
func (i Info) Tithi() (Tithi, bool)
func (i Info) IsPublicHoliday() bool
func (i Info) Events() iter.Seq[Event]                        // file order, no allocation
func (i Info) EventCount() int
func (i Info) Event(n int) Event                              // n-th event in file order; the zero Event (Name "") if n is out of range
type Tithi struct{ /* index into static table */ }
func (t Tithi) Name(lang nepdate.Lang) string
type Event struct{ /* index into static table */ }
func (e Event) Name(lang nepdate.Lang) string
func Coverage() nepdate.Range
func PublicHolidays() iter.Seq[nepdate.Date]                  // ascending
```

Notes:
- `Spill` is the neutral name for the spec's `Overflow` mode (C# `awayFromMonthEnd: true`); `Overflow`
  is the type, so the constant can't share the name.
- `Weekday` reuses `time.Weekday`, which already counts from Sunday = 0, like the spec.
- `FromTime` uses the calendar date in `t`'s own location. `FromTime(t.UTC())` and `FromTime(t)` can
  differ by a day in Nepal (UTC+05:45). The doc comment must say so.
