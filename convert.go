package nepdate

import "time"

// FromYMD returns the date of an integer in YYYYMMDD form, so 20810415 is 2081-04-15. It returns
// New's errors: ErrOutOfRange for a year outside 1901..2199, ErrInvalidMonth for a month outside
// 1..12 and ErrInvalidDay for a day outside the month. It does not allocate.
func FromYMD(n int) (Date, error) {
	year := n / 10000
	month := n / 100 % 100
	day := n % 100
	return newFromInts(year, month, day)
}

// YMD returns d as YYYYMMDD, so 2081-04-15 is 20810415, or 0 for the zero Date. It does not
// allocate.
func (d Date) YMD() int {
	if d.IsZero() {
		return 0
	}
	return int(d.year)*10000 + int(d.month)*100 + int(d.day)
}

// ToNepaliDates converts every time to its BS date in the time's own location, keeping the input
// order. It stops at the first time outside AD 1844-04-11..2143-04-15 and returns nil with that
// error. It allocates the result slice and nothing else.
func ToNepaliDates(times []time.Time) ([]Date, error) {
	if len(times) == 0 {
		return nil, nil
	}
	dates := make([]Date, len(times))
	for i, tm := range times {
		d, err := FromTime(tm)
		if err != nil {
			return nil, err
		}
		dates[i] = d
	}
	return dates, nil
}

// ToEnglishDates converts every date to midnight UTC on its Gregorian day, keeping the input
// order. It returns ErrOutOfRange for the zero Date. It allocates the result slice and nothing
// else.
func ToEnglishDates(dates []Date) ([]time.Time, error) {
	if len(dates) == 0 {
		return nil, nil
	}
	times := make([]time.Time, len(dates))
	for i, d := range dates {
		if d.IsZero() {
			return nil, ErrOutOfRange
		}
		year, month, day := d.Gregorian()
		times[i] = time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	}
	return times, nil
}

// ToEnglishDatesFromStrings parses every string with Parse and converts it to midnight UTC on its
// Gregorian day, keeping the input order. It stops at the first invalid string and returns nil
// with that error. It allocates the result slice and, on a failed parse, the parse error.
func ToEnglishDatesFromStrings(texts []string) ([]time.Time, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	times := make([]time.Time, len(texts))
	for i, text := range texts {
		d, err := Parse(text)
		if err != nil {
			return nil, err
		}
		year, month, day := d.Gregorian()
		times[i] = time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	}
	return times, nil
}
