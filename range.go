package nepdate

import (
	"iter"
	"time"
)

// Range is an inclusive span of dates, start <= end. The zero Range is empty. Equality is the
// equality of the two bounds, so == works and a Range can be a map key.
type Range struct {
	start, end Date
}

// NewRange returns the range start..end. It returns false if start is after end or either date
// is the zero Date. It does not allocate.
func NewRange(start, end Date) (Range, bool) {
	if start.IsZero() || end.IsZero() || start.After(end) {
		return Range{}, false
	}
	return Range{start, end}, true
}

// SingleDay returns the range that holds only d, or the zero Range for the zero Date. It does
// not allocate.
func SingleDay(d Date) Range {
	if d.IsZero() {
		return Range{}
	}
	return Range{d, d}
}

// FromDayCount returns the range of days dates starting at start. It returns ErrOutOfRange if
// days is below 1, start is the zero Date, or the last day would be after Max. It does not
// allocate.
func FromDayCount(start Date, days int) (Range, error) {
	if start.IsZero() || days < 1 || days > int(maxSerial)+1 {
		return Range{}, ErrOutOfRange
	}
	end, err := start.AddDays(days - 1)
	if err != nil {
		return Range{}, err
	}
	return Range{start, end}, nil
}

// MonthRange returns every day of the given BS month. It returns New's errors for an invalid
// year or month. It does not allocate.
func MonthRange(year int, month Month) (Range, error) {
	first, err := New(year, month, 1)
	if err != nil {
		return Range{}, err
	}
	return Range{first, first.LastOfMonth()}, nil
}

// YearRange returns every day of the BS year, Baishakh 1 to the last day of Chaitra. It returns
// ErrOutOfRange for a year outside 1901..2199. It does not allocate.
func YearRange(year int) (Range, error) {
	if year < minYear || year > maxYear {
		return Range{}, ErrOutOfRange
	}
	return Range{
		Date{uint16(year), uint8(Baishakh), 1},                             //nolint:gosec // checked above
		Date{uint16(year), uint8(Chaitra), uint8(monthLen(year, Chaitra))}, //nolint:gosec // checked above
	}, nil
}

// CurrentMonth returns the days of the BS month that contains today in loc. A nil loc means
// Asia/Kathmandu, like Today. It returns Today's error when the clock is outside the supported
// range.
func CurrentMonth(loc *time.Location) (Range, error) {
	return currentMonthAt(time.Now(), loc)
}

func currentMonthAt(now time.Time, loc *time.Location) (Range, error) {
	today, err := todayAt(now, loc)
	if err != nil {
		return Range{}, err
	}
	return MonthRange(today.Year(), today.Month())
}

// CurrentFiscalYear returns the days of the fiscal year that contains today in loc. A nil loc
// means Asia/Kathmandu, like Today. It returns Today's error when the clock is outside the
// supported range, and ErrOutOfRange when the fiscal year ends after Max.
func CurrentFiscalYear(loc *time.Location) (Range, error) {
	return currentFiscalYearAt(time.Now(), loc)
}

func currentFiscalYearAt(now time.Time, loc *time.Location) (Range, error) {
	today, err := todayAt(now, loc)
	if err != nil {
		return Range{}, err
	}
	return FiscalYearOf(today).Range()
}

// CurrentCalendarYear returns the days of the BS calendar year that contains today in loc. A nil
// loc means Asia/Kathmandu, like Today. It returns Today's error when the clock is outside the
// supported range.
func CurrentCalendarYear(loc *time.Location) (Range, error) {
	return currentCalendarYearAt(time.Now(), loc)
}

func currentCalendarYearAt(now time.Time, loc *time.Location) (Range, error) {
	today, err := todayAt(now, loc)
	if err != nil {
		return Range{}, err
	}
	return YearRange(today.Year())
}

// Start returns the first day of r, or the zero Date for the zero Range. It does not allocate.
func (r Range) Start() Date { return r.start }

// End returns the last day of r, or the zero Date for the zero Range. It does not allocate.
func (r Range) End() Date { return r.end }

// Len returns the number of days in r, counting both ends, or 0 for the zero Range. It does not
// allocate.
func (r Range) Len() int {
	if r.start.IsZero() {
		return 0
	}
	return r.start.DaysUntil(r.end) + 1
}

// Contains reports whether d is in r. The zero Date is in no range. It does not allocate.
func (r Range) Contains(d Date) bool {
	return !d.IsZero() && !r.start.IsZero() && !d.Before(r.start) && !d.After(r.end)
}

// ContainsRange reports whether r fully contains other. A non-empty range contains the zero
// Range, and the zero Range contains nothing. It does not allocate.
func (r Range) ContainsRange(other Range) bool {
	if r.start.IsZero() {
		return false
	}
	if other.start.IsZero() {
		return true
	}
	return !other.start.Before(r.start) && !other.end.After(r.end)
}

// Overlaps reports whether r and other share a date. The zero Range overlaps nothing. It does
// not allocate.
func (r Range) Overlaps(other Range) bool {
	if r.start.IsZero() || other.start.IsZero() {
		return false
	}
	return !r.start.After(other.end) && !other.start.After(r.end)
}

// IsAdjacentTo reports whether one range starts the day after the other ends. It returns false
// at Min and Max, so a valid range never fails because it touches the data end (PARITY D-16). It
// does not allocate.
func (r Range) IsAdjacentTo(other Range) bool {
	if r.start.IsZero() || other.start.IsZero() {
		return false
	}
	if next, ok := r.end.Next(); ok && next == other.start {
		return true
	}
	if next, ok := other.end.Next(); ok && next == r.start {
		return true
	}
	return false
}

// Intersect returns the dates in both r and other. It returns false when they do not overlap, so
// a caller has no empty range to inspect (PARITY D-01, D-16). It does not allocate.
func (r Range) Intersect(other Range) (Range, bool) {
	if r.start.IsZero() || other.start.IsZero() {
		return Range{}, false
	}
	start, end := r.start, r.end
	if other.start.After(start) {
		start = other.start
	}
	if other.end.Before(end) {
		end = other.end
	}
	if start.After(end) {
		return Range{}, false
	}
	return Range{start, end}, true
}

// Union returns the smallest range that covers both r and other. The zero Range adds nothing. It
// does not allocate.
func (r Range) Union(other Range) Range {
	if r.start.IsZero() {
		return other
	}
	if other.start.IsZero() {
		return r
	}
	start, end := r.start, r.end
	if other.start.Before(start) {
		start = other.start
	}
	if other.end.After(end) {
		end = other.end
	}
	return Range{start, end}
}

// Except returns the parts of r that are not in other, in date order: none, one or two ranges.
// It does not allocate when the result is nil or one range.
func (r Range) Except(other Range) []Range {
	if r.start.IsZero() {
		return nil
	}
	if other.start.IsZero() || !r.Overlaps(other) {
		return []Range{r}
	}
	var out []Range
	if r.start.Before(other.start) {
		end, _ := other.start.Prev()
		out = append(out, Range{r.start, end})
	}
	if other.end.Before(r.end) {
		start, _ := other.end.Next()
		out = append(out, Range{start, r.end})
	}
	return out
}

// SplitByMonth returns one range per BS month that r touches, in date order. The first and last
// pieces are clipped to r, and a range that reaches Max still returns its last piece (PARITY
// D-16). The result allocates one slice.
func (r Range) SplitByMonth() []Range {
	if r.start.IsZero() {
		return nil
	}
	var out []Range
	for current := r.start; ; {
		end := current.LastOfMonth()
		if end.After(r.end) {
			end = r.end
		}
		out = append(out, Range{current, end})
		if !end.Before(r.end) {
			return out
		}
		next, ok := end.Next()
		if !ok {
			return out
		}
		current = next
	}
}

// SplitByFiscalQuarter returns one range per fiscal quarter that r touches, in date order. The
// first and last pieces are clipped to r, and a range that reaches Max still returns its last
// piece (PARITY D-16). The result allocates one slice.
func (r Range) SplitByFiscalQuarter() []Range {
	if r.start.IsZero() {
		return nil
	}
	var out []Range
	for current := r.start; ; {
		end := quarterEnd(current)
		if end.After(r.end) {
			end = r.end
		}
		out = append(out, Range{current, end})
		if !end.Before(r.end) {
			return out
		}
		next, ok := end.Next()
		if !ok {
			return out
		}
		current = next
	}
}

// quarterEnd returns the last day of the fiscal quarter that contains d. A quarter that ends
// after Max is clipped to Max, so a split never needs a date outside the supported range.
func quarterEnd(d Date) Date {
	fy := FiscalYearOf(d)
	q := d.Quarter()
	months := q.Months()
	year := int(fy)
	if q == Q4 {
		year++
	}
	if year > maxYear {
		return Max
	}
	if year < minYear {
		return Min
	}
	return Date{uint16(year), uint8(months[2]), uint8(monthLen(year, months[2]))} //nolint:gosec // checked above
}

// DatesWithInterval returns every interval-th day of r, starting at Start and including End when
// the steps land on it. It returns ErrOutOfRange when interval is below 1. A nil sequence and an
// error are returned together, so check the error before ranging. It does not allocate while the
// sequence is ranged over.
func (r Range) DatesWithInterval(interval int) (iter.Seq[Date], error) {
	if interval < 1 {
		return nil, ErrOutOfRange
	}
	empty := func(func(Date) bool) {}
	if r.start.IsZero() {
		return empty, nil
	}
	if interval > int(maxSerial)+1 {
		interval = int(maxSerial) + 1
	}
	return func(yield func(Date) bool) {
		for s, e, step := toSerial(r.start), toSerial(r.end), int32(interval); s <= e; s += step { //nolint:gosec // interval is at most maxSerial+1
			if !yield(fromSerial(s)) {
				return
			}
		}
	}, nil
}

// WorkingDays returns every day of r except Saturday, and except Sunday when excludeSunday is
// true. A range that reaches Max yields its last date (PARITY D-16). It does not allocate while
// ranged over.
func (r Range) WorkingDays(excludeSunday bool) iter.Seq[Date] {
	return func(yield func(Date) bool) {
		if r.start.IsZero() {
			return
		}
		for s, e := toSerial(r.start), toSerial(r.end); s <= e; s++ {
			d := fromSerial(s)
			weekday := d.Weekday()
			if weekday == time.Saturday || excludeSunday && weekday == time.Sunday {
				continue
			}
			if !yield(d) {
				return
			}
		}
	}
}

// WeekendDays returns every Saturday in r, and every Sunday when includeSunday is true. It does
// not allocate while ranged over.
func (r Range) WeekendDays(includeSunday bool) iter.Seq[Date] {
	return func(yield func(Date) bool) {
		if r.start.IsZero() {
			return
		}
		for s, e := toSerial(r.start), toSerial(r.end); s <= e; s++ {
			d := fromSerial(s)
			weekday := d.Weekday()
			if weekday != time.Saturday && (!includeSunday || weekday != time.Sunday) {
				continue
			}
			if !yield(d) {
				return
			}
		}
	}
}

// Days returns an iterator over the days of r in order. It yields every date through Max when the
// range ends there (PARITY D-16). It does not allocate when ranged over directly (ADR-0002 §6).
func (r Range) Days() iter.Seq[Date] {
	return func(yield func(Date) bool) {
		if r.start.IsZero() {
			return
		}
		for s, e := toSerial(r.start), toSerial(r.end); s <= e; s++ {
			if !yield(fromSerial(s)) {
				return
			}
		}
	}
}

// String returns r as "start - end" with the default date form, or "" for the zero Range.
func (r Range) String() string {
	if r.start.IsZero() {
		return ""
	}
	return r.start.String() + " - " + r.end.String()
}

// Format returns r as its two dates in order o, separated by sep, in lang's digits, with " - "
// between them. It returns "" for the zero Range or an unknown order or separator.
func (r Range) Format(o DateOrder, sep Separator, lang Lang) string {
	if r.start.IsZero() || int(o) >= len(orderFields) || int(sep) >= len(separatorBytes) {
		return ""
	}
	return r.start.Short(o, sep, true, lang) + " - " + r.end.Short(o, sep, true, lang)
}
