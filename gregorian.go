package nepdate

import (
	"cmp"
	"time"
)

const (
	adMinYear = 1844 // first AD year in adMonthStart
	adMaxYear = 2143 // last AD year with a supported day
)

// gregLen is the length of each Gregorian month in a common year; index 0 is unused.
var gregLen = [13]uint8{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

// FromGregorian returns the BS date of the AD date year-month-day. It returns
// ErrInvalidGregorian if that AD date does not exist (2023-02-29) and ErrOutOfRange if it is
// outside AD 1844-04-11..2143-04-15, checked in that order (ALGORITHM §3). It does not allocate.
func FromGregorian(year int, month time.Month, day int) (Date, error) {
	if month < time.January || month > time.December || day < 1 {
		return Date{}, ErrInvalidGregorian
	}
	n := int(gregLen[month])
	if month == time.February && (year%4 == 0 && year%100 != 0 || year%400 == 0) {
		n = 29
	}
	if day > n {
		return Date{}, ErrInvalidGregorian
	}
	// The year check comes before any arithmetic, so a huge year can't overflow a 32-bit int.
	if year < adMinYear || year > adMaxYear {
		return Date{}, ErrOutOfRange
	}
	s := adMonthStart[(year-adMinYear)*12+int(month)-1] + int32(day) - 1 //nolint:gosec // day is 1..31
	if s < 0 || s > maxSerial {
		return Date{}, ErrOutOfRange
	}
	return fromSerial(s), nil
}

// Gregorian returns the AD date of d, or (0, 0, 0) for the zero Date. It does not allocate.
func (d Date) Gregorian() (year int, month time.Month, day int) {
	if d.IsZero() {
		return 0, 0, 0
	}
	s := toSerial(d)
	k := int((int64(s)*adEstMul + adEstAdd) >> estShift)
	if adMonthStart[k+1] <= s {
		k++
	}
	return adMinYear + k/12, time.Month(k%12 + 1), int(s-adMonthStart[k]) + 1
}

// Weekday returns the day of the week of d (ALGORITHM §7). The zero Date returns Sunday, the
// zero value of time.Weekday. It does not allocate.
func (d Date) Weekday() time.Weekday {
	if d.IsZero() {
		return time.Sunday
	}
	return time.Weekday((toSerial(d) + 4) % 7)
}

// YearDay returns the day of the BS year, 1 for Baishakh 1, or 0 for the zero Date. It does not
// allocate.
func (d Date) YearDay() int {
	if d.IsZero() {
		return 0
	}
	return int(toSerial(d)-bsMonthStart[(int(d.year)-minYear)*12]) + 1
}

// FirstOfMonth returns the first day of d's month, or the zero Date for the zero Date. It does
// not allocate.
func (d Date) FirstOfMonth() Date {
	if d.IsZero() {
		return Date{}
	}
	return Date{d.year, d.month, 1}
}

// LastOfMonth returns the last day of d's month, or the zero Date for the zero Date. It does not
// allocate.
func (d Date) LastOfMonth() Date {
	if d.IsZero() {
		return Date{}
	}
	return Date{d.year, d.month, uint8(d.MonthLength())} //nolint:gosec // 29..32
}

// Compare returns -1 if d is before other, 0 if they are equal and +1 if d is after other. The
// zero Date sorts before every valid date. It does not allocate.
func (d Date) Compare(other Date) int {
	return cmp.Compare(d.key(), other.key())
}

// key packs the fields so that comparing keys compares dates.
func (d Date) key() uint32 {
	return uint32(d.year)<<16 | uint32(d.month)<<8 | uint32(d.day)
}

// Before reports whether d is before other. It does not allocate.
func (d Date) Before(other Date) bool { return d.Compare(other) < 0 }

// After reports whether d is after other. It does not allocate.
func (d Date) After(other Date) bool { return d.Compare(other) > 0 }
