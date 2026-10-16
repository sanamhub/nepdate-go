package nepdate

// Date is a Bikram Sambat calendar date. It is a small comparable value: use == and as a map
// key. The zero value is "no date" (PARITY D-01): IsZero reports true, accessors return 0, and
// every operation that returns an error returns ErrOutOfRange.
type Date struct {
	year       uint16
	month, day uint8
}

// Min and Max are the first and last supported dates, BS 1901-01-01 and 2199-12-30.
var (
	Min = Date{1901, 1, 1}
	Max = Date{2199, 12, 30}
)

// New returns the date year-month-day. It checks the year first, then the month, then the day
// (ALGORITHM §3), and returns ErrOutOfRange for a year outside 1901..2199, ErrInvalidMonth for a
// month outside 1..12 and ErrInvalidDay for a day outside the month. It does not allocate.
func New(year int, month Month, day int) (Date, error) {
	if year < minYear || year > maxYear {
		return Date{}, ErrOutOfRange
	}
	if month < Baishakh || month > Chaitra {
		return Date{}, ErrInvalidMonth
	}
	if day < 1 || day > monthLen(year, month) {
		return Date{}, ErrInvalidDay
	}
	return Date{uint16(year), uint8(month), uint8(day)}, nil //nolint:gosec // all three checked above
}

// MustNew is New for constants and tests. It panics with New's error.
func MustNew(year int, month Month, day int) Date {
	d, err := New(year, month, day)
	if err != nil {
		panic(err)
	}
	return d
}

// monthIndex is k = (y-1901)*12 + m-1, the index into bsMonthStart. The date must be valid.
func monthIndex(year int, month Month) int {
	return (year-minYear)*12 + int(month) - 1
}

// monthLen is the length of a valid BS month.
func monthLen(year int, month Month) int {
	k := monthIndex(year, month)
	return int(bsMonthStart[k+1] - bsMonthStart[k])
}

// IsZero reports whether d is the zero Date, which is "no date". It does not allocate.
func (d Date) IsZero() bool { return d.year == 0 }

// Year returns the BS year, 1901..2199, or 0 for the zero Date. It does not allocate.
func (d Date) Year() int { return int(d.year) }

// Month returns the BS month, or 0 for the zero Date. It does not allocate.
func (d Date) Month() Month { return Month(d.month) }

// Day returns the day of the month, 1..32, or 0 for the zero Date. It does not allocate.
func (d Date) Day() int { return int(d.day) }

// MonthLength returns the number of days in d's month, 29..32, or 0 for the zero Date. It does
// not allocate.
func (d Date) MonthLength() int {
	if d.IsZero() {
		return 0
	}
	return monthLen(int(d.year), Month(d.month))
}
