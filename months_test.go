package nepdate

import (
	"errors"
	"math"
	"strconv"
	"testing"
)

func TestAddMonthsGolden(t *testing.T) {
	rows := readTable(t, "shared/spec/vectors/csharp-golden/add-months.tsv", "\t", 10180)
	for _, row := range rows {
		d := bsDate(t, row["bs"])
		n, err := strconv.Atoi(row["months"])
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []struct {
			o   Overflow
			col string
		}{{Clamp, "clamp"}, {Spill, "overflow"}} {
			got, err := d.AddMonths(n, mode.o)
			if row[mode.col] == "ERR" {
				if !errors.Is(err, ErrOutOfRange) || !got.IsZero() {
					t.Errorf("%s + %d months (%s) = %v, %v; want ErrOutOfRange", row["bs"], n, mode.col, got, err)
				}
				continue
			}
			if want := bsDate(t, row[mode.col]); err != nil || got != want {
				t.Errorf("%s + %d months (%s) = %v, %v; want %v", row["bs"], n, mode.col, got, err, want)
			}
		}
	}
}

func TestAddMonthsExamples(t *testing.T) {
	d := MustNew(2081, Shrawan, 32)
	cases := []struct {
		n    int
		o    Overflow
		want Date
	}{
		{2, Clamp, MustNew(2081, Ashoj, 30)},
		{2, Spill, MustNew(2081, Kartik, 2)},
		{5, Clamp, MustNew(2081, Poush, 29)},
		{5, Spill, MustNew(2081, Magh, 3)},
	}
	for _, c := range cases {
		if got, err := d.AddMonths(c.n, c.o); err != nil || got != c.want {
			t.Errorf("2081-04-32 + %d months (mode %d) = %v, %v; want %v", c.n, c.o, got, err, c.want)
		}
	}
}

func TestAddMonthsBounds(t *testing.T) {
	for _, c := range []struct {
		name string
		f    func() (Date, error)
	}{
		{"AddYears(MaxInt)", func() (Date, error) { return Min.AddYears(math.MaxInt, Clamp) }},
		{"AddYears(MinInt)", func() (Date, error) { return Max.AddYears(math.MinInt, Clamp) }},
		{"AddYears(300)", func() (Date, error) { return Min.AddYears(300, Clamp) }},
		{"AddMonths(MaxInt)", func() (Date, error) { return Min.AddMonths(math.MaxInt, Spill) }},
		{"AddMonths(MinInt)", func() (Date, error) { return Max.AddMonths(math.MinInt, Clamp) }},
		{"AddMonths(3588)", func() (Date, error) { return Min.AddMonths(3588, Clamp) }},
		{"Spill into 2200", func() (Date, error) { return MustNew(2199, Baishakh, 31).AddMonths(11, Spill) }},
		{"zero AddMonths", func() (Date, error) { return Date{}.AddMonths(0, Clamp) }},
		{"zero AddYears", func() (Date, error) { return Date{}.AddYears(1, Clamp) }},
	} {
		if got, err := c.f(); !errors.Is(err, ErrOutOfRange) || !got.IsZero() {
			t.Errorf("%s = %v, %v; want ErrOutOfRange", c.name, got, err)
		}
	}
	if got, err := Min.AddMonths(3587, Clamp); err != nil || got != MustNew(2199, Chaitra, 1) {
		t.Errorf("Min + 3587 months = %v, %v", got, err)
	}
	if got, err := Min.AddYears(298, Clamp); err != nil || got != MustNew(2199, Baishakh, 1) {
		t.Errorf("Min + 298 years = %v, %v", got, err)
	}
	// Shrawan 2078 has 32 days and Shrawan 2079 has 31.
	if got, err := MustNew(2078, Shrawan, 32).AddYears(1, Spill); err != nil || got != MustNew(2079, Bhadra, 1) {
		t.Errorf("2078-04-32 + 1 year (Spill) = %v, %v; want 2079-05-01", got, err)
	}
	if got, err := MustNew(2078, Shrawan, 32).AddYears(1, Clamp); err != nil || got != MustNew(2079, Shrawan, 31) {
		t.Errorf("2078-04-32 + 1 year (Clamp) = %v, %v; want 2079-04-31", got, err)
	}
}
