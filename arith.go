package nepdate

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
