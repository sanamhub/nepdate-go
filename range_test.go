package nepdate

import (
	"errors"
	"testing"
)

func TestYearRangeLengths(t *testing.T) {
	for y := 1901; y <= 2199; y++ {
		r, err := YearRange(y)
		if err != nil {
			t.Fatalf("YearRange(%d): %v", y, err)
		}
		sum := 0
		for m := Baishakh; m <= Chaitra; m++ {
			sum += MustNew(y, m, 1).MonthLength()
		}
		if r.Len() != sum || r.Len() < 364 || r.Len() > 367 {
			t.Errorf("YearRange(%d).Len() = %d, months sum to %d", y, r.Len(), sum)
		}
		if r.Start() != MustNew(y, Baishakh, 1) || r.End() != MustNew(y, Chaitra, 1).LastOfMonth() {
			t.Errorf("YearRange(%d) = %v..%v", y, r.Start(), r.End())
		}
	}
	for _, y := range []int{1900, 2200} {
		if _, err := YearRange(y); !errors.Is(err, ErrOutOfRange) {
			t.Errorf("YearRange(%d) error = %v", y, err)
		}
	}
}

func TestMonthRange(t *testing.T) {
	r, err := MonthRange(2081, Shrawan)
	if err != nil {
		t.Fatal(err)
	}
	var days []Date
	for d := range r.Days() {
		days = append(days, d)
	}
	if len(days) != 32 || days[0] != MustNew(2081, Shrawan, 1) || days[31] != MustNew(2081, Shrawan, 32) || r.Len() != 32 {
		t.Errorf("MonthRange(2081, Shrawan) yields %d days", len(days))
	}
	n := 0
	for range r.Days() {
		n++
		if n == 3 {
			break
		}
	}
	if n != 3 {
		t.Errorf("break after 3 counted %d", n)
	}
	if _, err := MonthRange(2081, 13); !errors.Is(err, ErrInvalidMonth) {
		t.Errorf("MonthRange(2081, 13) error = %v", err)
	}
	if _, err := MonthRange(2200, 1); !errors.Is(err, ErrOutOfRange) {
		t.Errorf("MonthRange(2200, 1) error = %v", err)
	}
}

func TestNewRange(t *testing.T) {
	if _, ok := NewRange(Max, Min); ok {
		t.Error("NewRange(Max, Min) succeeded")
	}
	if _, ok := NewRange(Date{}, Max); ok {
		t.Error("NewRange with a zero start succeeded")
	}
	if _, ok := NewRange(Min, Date{}); ok {
		t.Error("NewRange with a zero end succeeded")
	}
	r, ok := NewRange(Min, Max)
	if !ok || r.Len() != 109212 || !r.Contains(Min) || !r.Contains(Max) || r.Contains(Date{}) {
		t.Errorf("NewRange(Min, Max) = %v, %v", r, ok)
	}
	one, ok := NewRange(Min, Min)
	if !ok || one.Len() != 1 {
		t.Errorf("one-day range: Len %d, ok %v", one.Len(), ok)
	}
	s, _ := MonthRange(2081, Shrawan)
	if s.Contains(MustNew(2081, Ashad, 31)) || s.Contains(MustNew(2081, Bhadra, 1)) || !s.Contains(MustNew(2081, Shrawan, 15)) {
		t.Error("Contains is wrong at the month edges")
	}
}

func TestZeroRange(t *testing.T) {
	var r Range
	if r.Len() != 0 || r.Contains(Min) || !r.Start().IsZero() || !r.End().IsZero() {
		t.Error("zero Range is not empty")
	}
	for range r.Days() {
		t.Fatal("zero Range yields a day")
	}
}
