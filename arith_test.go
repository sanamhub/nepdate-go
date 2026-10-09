package nepdate

import (
	"errors"
	"math"
	"strconv"
	"testing"
)

func TestAddDaysGolden(t *testing.T) {
	rows := readTable(t, "shared/spec/vectors/csharp-golden/add-days.tsv", "\t", 9162)
	for _, row := range rows {
		d := bsDate(t, row["bs"])
		n, err := strconv.Atoi(row["days"])
		if err != nil {
			t.Fatal(err)
		}
		got, err := d.AddDays(n)
		if row["result"] == "ERR" {
			if !errors.Is(err, ErrOutOfRange) {
				t.Errorf("%s + %d days = %v, %v; want ErrOutOfRange", row["bs"], n, got, err)
			}
			continue
		}
		if want := bsDate(t, row["result"]); err != nil || got != want {
			t.Errorf("%s + %d days = %v, %v; want %v", row["bs"], n, got, err, want)
		}
	}
}

func TestAddDaysEdges(t *testing.T) {
	for _, c := range []struct {
		d Date
		n int
	}{
		{Max, 1},
		{Min, -1},
		{Min, math.MaxInt},
		{Min, math.MinInt},
		{Max, math.MinInt},
		{Min, 109212},
		{Max, -109212},
		{Date{}, 0},
		{Date{}, 1},
	} {
		if got, err := c.d.AddDays(c.n); !errors.Is(err, ErrOutOfRange) || !got.IsZero() {
			t.Errorf("%v.AddDays(%d) = %v, %v; want ErrOutOfRange", c.d, c.n, got, err)
		}
	}
	if got, err := Min.AddDays(109211); err != nil || got != Max {
		t.Errorf("Min.AddDays(109211) = %v, %v", got, err)
	}
	if got, err := Max.AddDays(-109211); err != nil || got != Min {
		t.Errorf("Max.AddDays(-109211) = %v, %v", got, err)
	}
}

func TestDaysUntil(t *testing.T) {
	if got := Min.DaysUntil(Max); got != 109211 {
		t.Errorf("Min.DaysUntil(Max) = %d", got)
	}
	if got := Max.DaysUntil(Min); got != -109211 {
		t.Errorf("Max.DaysUntil(Min) = %d", got)
	}
	if got := Min.DaysUntil(Date{}); got != 0 {
		t.Errorf("DaysUntil(zero Date) = %d, want 0 (D-01)", got)
	}
	if got := (Date{}).DaysUntil(Max); got != 0 {
		t.Errorf("zero Date DaysUntil = %d, want 0 (D-01)", got)
	}
}

func TestNextPrev(t *testing.T) {
	d := MustNew(2081, Shrawan, 32)
	if n, ok := d.Next(); !ok || n != MustNew(2081, Bhadra, 1) {
		t.Errorf("Next = %v, %v", n, ok)
	}
	if p, ok := MustNew(2081, Bhadra, 1).Prev(); !ok || p != d {
		t.Errorf("Prev = %v, %v", p, ok)
	}
	for _, c := range []struct {
		name string
		f    func() (Date, bool)
	}{
		{"Max.Next", Max.Next}, {"Min.Prev", Min.Prev}, {"zero Next", Date{}.Next}, {"zero Prev", Date{}.Prev},
	} {
		if got, ok := c.f(); ok || !got.IsZero() {
			t.Errorf("%s = %v, %v; want zero, false", c.name, got, ok)
		}
	}
}

func FuzzAddDays(f *testing.F) {
	f.Add(int32(0), 1)
	f.Add(int32(65746), -1000)
	f.Add(int32(maxSerial), 1)
	f.Add(int32(100), math.MaxInt)
	f.Fuzz(func(t *testing.T, s int32, n int) {
		if s < 0 || s > maxSerial {
			return
		}
		d := fromSerial(s)
		got, err := d.AddDays(n)
		if err != nil {
			if !errors.Is(err, ErrOutOfRange) {
				t.Fatalf("%v.AddDays(%d): %v", d, n, err)
			}
			return
		}
		back, err := got.AddDays(-n)
		if err != nil || back != d {
			t.Fatalf("%v + %d - %d = %v, %v", d, n, n, back, err)
		}
	})
}
