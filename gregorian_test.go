package nepdate

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestFromGregorian(t *testing.T) {
	cases := []struct {
		y    int
		m    time.Month
		d    int
		want Date
		err  error
	}{
		{1844, time.April, 11, Min, nil},
		{2143, time.April, 15, Max, nil},
		{1844, time.April, 10, Date{}, ErrOutOfRange},
		{2143, time.April, 16, Date{}, ErrOutOfRange},
		{1844, time.January, 1, Date{}, ErrOutOfRange},
		{2143, time.December, 31, Date{}, ErrOutOfRange},
		{2023, time.February, 30, Date{}, ErrInvalidGregorian},
		{2023, time.February, 29, Date{}, ErrInvalidGregorian},
		{2024, time.February, 29, MustNew(2080, Falgun, 17), nil},
		{1900, time.February, 29, Date{}, ErrInvalidGregorian}, // not a leap year
		{2000, time.February, 29, MustNew(2056, Falgun, 17), nil},
		{2024, 13, 1, Date{}, ErrInvalidGregorian},
		{2024, 0, 1, Date{}, ErrInvalidGregorian},
		{2024, time.April, 31, Date{}, ErrInvalidGregorian},
		{2024, time.April, 0, Date{}, ErrInvalidGregorian},
		{math.MaxInt, time.January, 1, Date{}, ErrOutOfRange},
		{math.MinInt, time.January, 1, Date{}, ErrOutOfRange},
		{1, time.January, 1, Date{}, ErrOutOfRange},
	}
	for _, c := range cases {
		got, err := FromGregorian(c.y, c.m, c.d)
		if got != c.want || !errors.Is(err, c.err) || (err == nil) != (c.err == nil) {
			t.Errorf("FromGregorian(%d, %d, %d) = %v, %v; want %v, %v", c.y, c.m, c.d, got, err, c.want, c.err)
		}
	}
}

// TestReferenceValues is the table of ALGORITHM §9.
func TestReferenceValues(t *testing.T) {
	cases := []struct {
		bs      Date
		y       int
		m       time.Month
		d       int
		weekday time.Weekday
	}{
		{MustNew(1901, Baishakh, 1), 1844, time.April, 11, time.Thursday},
		{MustNew(2081, Baishakh, 1), 2024, time.April, 13, time.Saturday},
		{MustNew(2082, Baishakh, 1), 2025, time.April, 14, time.Monday},
		{MustNew(2083, Baishakh, 1), 2026, time.April, 14, time.Tuesday},
		{MustNew(2199, Chaitra, 30), 2143, time.April, 15, time.Monday},
	}
	for _, c := range cases {
		y, m, d := c.bs.Gregorian()
		if y != c.y || m != c.m || d != c.d {
			t.Errorf("%v.Gregorian() = %d-%d-%d, want %d-%d-%d", c.bs, y, m, d, c.y, c.m, c.d)
		}
		if got, err := FromGregorian(c.y, c.m, c.d); err != nil || got != c.bs {
			t.Errorf("FromGregorian(%d-%d-%d) = %v, %v; want %v", c.y, c.m, c.d, got, err, c.bs)
		}
		if got := c.bs.Weekday(); got != c.weekday {
			t.Errorf("%v.Weekday() = %v, want %v", c.bs, got, c.weekday)
		}
	}
}

func TestYearDayAndMonthEnds(t *testing.T) {
	d := MustNew(2081, Shrawan, 15)
	// Baishakh 31 + Jestha 32 + Ashad 31 days in 2081, then 15.
	if got := d.YearDay(); got != 31+32+31+15 {
		t.Errorf("YearDay = %d, want %d", got, 31+32+31+15)
	}
	if got := Min.YearDay(); got != 1 {
		t.Errorf("Min.YearDay() = %d", got)
	}
	if got := d.FirstOfMonth(); got != MustNew(2081, Shrawan, 1) {
		t.Errorf("FirstOfMonth = %v", got)
	}
	if got := d.LastOfMonth(); got != MustNew(2081, Shrawan, 32) {
		t.Errorf("LastOfMonth = %v", got)
	}
}

func TestZeroDateDerived(t *testing.T) {
	var z Date
	if y, m, d := z.Gregorian(); y != 0 || m != 0 || d != 0 {
		t.Errorf("zero Gregorian() = %d-%d-%d", y, m, d)
	}
	if z.YearDay() != 0 || z.Weekday() != 0 || !z.FirstOfMonth().IsZero() || !z.LastOfMonth().IsZero() {
		t.Error("zero Date derived values are not zero")
	}
	if z.Compare(Min) != -1 || !z.Before(Min) {
		t.Error("zero Date does not sort before Min")
	}
}

func TestCompare(t *testing.T) {
	a := MustNew(2080, Chaitra, 30)
	b := MustNew(2081, Baishakh, 1)
	c := MustNew(2081, Jestha, 1)
	e := MustNew(2081, Jestha, 2)
	ordered := []Date{Min, a, b, c, e, Max}
	for i, x := range ordered {
		for j, y := range ordered {
			want := 0
			if i < j {
				want = -1
			} else if i > j {
				want = 1
			}
			if got := x.Compare(y); got != want {
				t.Errorf("%v.Compare(%v) = %d, want %d", x, y, got, want)
			}
			if x.Before(y) != (want < 0) || x.After(y) != (want > 0) {
				t.Errorf("Before/After disagree with Compare for %v, %v", x, y)
			}
		}
	}
}
