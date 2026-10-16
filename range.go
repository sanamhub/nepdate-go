package nepdate

import "iter"

// Range is an inclusive span of dates, start <= end. The zero Range is empty. It has no set
// operations (PARITY D-09): range over Days and filter instead.
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

// Days returns an iterator over the days of r in order. It does not allocate when ranged over
// directly (ADR-0002 §6).
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
