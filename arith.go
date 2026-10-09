package nepdate

// bsMonths is the number of BS months in the range, 1901-01 to 2199-12.
const bsMonths = (maxYear - minYear + 1) * 12

// AddDays returns d moved by n days. It returns ErrOutOfRange if the result is outside the
// supported range or d is the zero Date.
func (d Date) AddDays(n int) (Date, error) {
	// The bound check comes first, so a huge n can't overflow a 32-bit int.
	if d.IsZero() || n < -maxSerial || n > maxSerial {
		return Date{}, ErrOutOfRange
	}
	s := toSerial(d) + int32(n) //nolint:gosec // n is within ±maxSerial
	if s < 0 || s > maxSerial {
		return Date{}, ErrOutOfRange
	}
	return fromSerial(s), nil
}

// Next returns the day after d, or (Date{}, false) at Max and for the zero Date.
func (d Date) Next() (Date, bool) {
	next, err := d.AddDays(1)
	return next, err == nil
}

// Prev returns the day before d, or (Date{}, false) at Min and for the zero Date.
func (d Date) Prev() (Date, bool) {
	prev, err := d.AddDays(-1)
	return prev, err == nil
}

// DaysUntil returns the signed number of days from d to other (PARITY D-07): positive when
// other is later. It returns 0 if either date is the zero Date.
func (d Date) DaysUntil(other Date) int {
	if d.IsZero() || other.IsZero() {
		return 0
	}
	return int(toSerial(other) - toSerial(d))
}

// Overflow says what AddMonths and AddYears do when the day doesn't exist in the target month
// (ALGORITHM §8).
type Overflow uint8

// The overflow modes.
const (
	// Clamp moves the day back to the last day of the target month, like C# AddMonths(n).
	Clamp Overflow = iota
	// Spill carries the extra days into the next month, like C# AddMonths(n, awayFromMonthEnd:
	// true). The spec calls this mode Overflow.
	Spill
)

// AddMonths returns d moved by n months, with o deciding what happens to a day past the end of
// the target month. It returns ErrOutOfRange if the result is outside the supported range or d
// is the zero Date.
func (d Date) AddMonths(n int, o Overflow) (Date, error) {
	// The bound check comes first, so a huge n can't overflow a 32-bit int.
	if d.IsZero() || n < -bsMonths || n > bsMonths {
		return Date{}, ErrOutOfRange
	}
	t := monthIndex(int(d.year), Month(d.month)) + n
	if t < 0 || t >= bsMonths {
		return Date{}, ErrOutOfRange
	}
	length := int(bsMonthStart[t+1] - bsMonthStart[t])
	day := int(d.day)
	if day > length {
		if o != Spill {
			day = length
		} else {
			t++
			day -= length
			if t >= bsMonths {
				return Date{}, ErrOutOfRange
			}
		}
	}
	return Date{
		year:  uint16(minYear + t/12), //nolint:gosec // t is 0..3587
		month: uint8(t%12 + 1),        //nolint:gosec // 1..12
		day:   uint8(day),             //nolint:gosec // 1..32
	}, nil
}

// AddYears returns d moved by n years, which is AddMonths(12*n, o). It returns ErrOutOfRange
// like AddMonths.
func (d Date) AddYears(n int, o Overflow) (Date, error) {
	// Checked before the multiplication, so 12*n can't overflow a 32-bit int.
	if n < -(maxYear-minYear) || n > maxYear-minYear {
		return Date{}, ErrOutOfRange
	}
	return d.AddMonths(12*n, o)
}
