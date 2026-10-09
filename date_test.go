package nepdate

import (
	"errors"
	"testing"
	"unsafe"
)

func TestNewValidation(t *testing.T) {
	cases := []struct {
		y   int
		m   Month
		d   int
		err error
	}{
		{2081, Shrawan, 32, nil},
		{2081, Shrawan, 33, ErrInvalidDay},
		{2081, 13, 1, ErrInvalidMonth},
		{2081, 0, 1, ErrInvalidMonth},
		{2081, Baishakh, 0, ErrInvalidDay},
		{2081, Baishakh, -1, ErrInvalidDay},
		{1900, 13, 40, ErrOutOfRange}, // year is checked first (ALGORITHM §3)
		{2200, 1, 1, ErrOutOfRange},
		{2199, 13, 0, ErrInvalidMonth},
		{1901, Baishakh, 1, nil},
		{2199, Chaitra, 30, nil},
		{2199, Chaitra, 31, ErrInvalidDay},
	}
	for _, c := range cases {
		d, err := New(c.y, c.m, c.d)
		if !errors.Is(err, c.err) || (c.err == nil) != (err == nil) {
			t.Errorf("New(%d, %d, %d) error = %v, want %v", c.y, c.m, c.d, err, c.err)
			continue
		}
		if err != nil {
			if !d.IsZero() {
				t.Errorf("New(%d, %d, %d) returned %v with an error", c.y, c.m, c.d, d)
			}
			continue
		}
		if d.Year() != c.y || d.Month() != c.m || d.Day() != c.d || d.IsZero() {
			t.Errorf("New(%d, %d, %d) = %d-%d-%d", c.y, c.m, c.d, d.Year(), d.Month(), d.Day())
		}
	}
}

func TestZeroDate(t *testing.T) {
	var d Date
	if !d.IsZero() || d.Year() != 0 || d.Month() != 0 || d.Day() != 0 || d.MonthLength() != 0 {
		t.Errorf("zero Date: IsZero %v, %d-%d-%d, length %d", d.IsZero(), d.Year(), d.Month(), d.Day(), d.MonthLength())
	}
}

func TestMinMax(t *testing.T) {
	if Min != MustNew(1901, Baishakh, 1) || Max != MustNew(2199, Chaitra, 30) {
		t.Errorf("Min = %v, Max = %v", Min, Max)
	}
}

func TestMonthLength(t *testing.T) {
	if got := MustNew(2081, Shrawan, 1).MonthLength(); got != 32 {
		t.Errorf("2081 Shrawan has %d days, want 32", got)
	}
	if got := MustNew(2081, Bhadra, 1).MonthLength(); got != 31 {
		t.Errorf("2081 Bhadra has %d days, want 31", got)
	}
}

func TestMustNewPanics(t *testing.T) {
	defer func() {
		r := recover()
		err, ok := r.(error)
		if !ok || !errors.Is(err, ErrInvalidDay) {
			t.Errorf("MustNew(2081, Shrawan, 33) panicked with %v, want ErrInvalidDay", r)
		}
	}()
	MustNew(2081, Shrawan, 33)
}

func TestDateSize(t *testing.T) {
	if s := unsafe.Sizeof(Date{}); s != 4 {
		t.Errorf("Date is %d bytes, want 4 (ADR-0002 §1)", s)
	}
}

func TestDateIsMapKey(t *testing.T) {
	m := map[Date]int{MustNew(2081, Shrawan, 15): 1}
	if m[MustNew(2081, Shrawan, 15)] != 1 {
		t.Error("equal dates are not the same map key")
	}
}
