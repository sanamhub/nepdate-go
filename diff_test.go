package nepdate

import (
	"math/rand/v2"
	"testing"
)

func TestDiff(t *testing.T) {
	cases := []struct {
		a, b Date
		want Diff
	}{
		// ALGORITHM §8a
		{MustNew(2080, 1, 1), MustNew(2081, 4, 15), Diff{false, 1, 3, 14, 473}},
		{MustNew(2081, 4, 15), MustNew(2080, 1, 1), Diff{true, 1, 3, 14, -473}},
		{MustNew(2081, 4, 32), MustNew(2081, 5, 31), Diff{false, 0, 1, 0, 31}},
		{MustNew(2081, 4, 32), MustNew(2081, 6, 1), Diff{false, 0, 1, 1, 32}},
		{MustNew(2081, 1, 31), MustNew(2081, 2, 30), Diff{false, 0, 0, 30, 30}},
		{MustNew(2081, 3, 15), MustNew(2081, 3, 15), Diff{false, 0, 0, 0, 0}},
		{MustNew(2081, 4, 32), MustNew(2081, 6, 30), Diff{false, 0, 2, 0, 61}},
		{MustNew(2081, 1, 1), MustNew(2082, 1, 1), Diff{false, 1, 0, 0, 366}},
		{MustNew(2081, 4, 15), MustNew(2083, 2, 10), Diff{false, 1, 9, 26, 663}},
		{MustNew(2083, 2, 10), MustNew(2081, 4, 15), Diff{true, 1, 9, 26, -663}},
		{MustNew(2081, 4, 15), MustNew(2081, 4, 15), Diff{false, 0, 0, 0, 0}},
		// ends of the range, and the zero Date (D-01)
		{Min, Max, Diff{false, 298, 11, 29, 109211}},
		{Date{}, Max, Diff{}},
		{Min, Date{}, Diff{}},
	}
	for _, c := range cases {
		if got := c.a.Diff(c.b); got != c.want {
			t.Errorf("%v.Diff(%v) = %+v, want %+v", c.a, c.b, got, c.want)
		}
	}
}

func TestDiffTotalDaysMatchesDaysUntil(t *testing.T) {
	r := rand.New(rand.NewPCG(2081, 4)) //nolint:gosec // a fixed seed is the point: the same 1000 pairs every run
	for range 1000 {
		a := fromSerial(r.Int32N(maxSerial + 1))
		b := fromSerial(r.Int32N(maxSerial + 1))
		diff := a.Diff(b)
		if diff.TotalDays != a.DaysUntil(b) {
			t.Fatalf("%v.Diff(%v).TotalDays = %d, DaysUntil = %d", a, b, diff.TotalDays, a.DaysUntil(b))
		}
		// The breakdown adds back up: months from the earlier date, then days.
		lo, hi := a, b
		if diff.Negative {
			lo, hi = b, a
		}
		mid, err := lo.AddMonths(diff.Years*12+diff.Months, Clamp)
		if err != nil || mid.DaysUntil(hi) != diff.Days || diff.Days < 0 || diff.Days >= 33 {
			t.Fatalf("%v.Diff(%v) = %+v does not add back up", a, b, diff)
		}
	}
}
