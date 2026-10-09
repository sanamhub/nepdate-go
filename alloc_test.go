package nepdate

import (
	"testing"
	"time"
)

// Package-level sinks keep the compiler from removing the measured calls.
var (
	sinkDate    Date
	sinkInt     int
	sinkErr     error
	sinkString  string
	sinkMonth   time.Month
	sinkWeekday time.Weekday
)

// TestAlloc checks that the paths of ADR-0002 §5 allocate nothing.
func TestAlloc(t *testing.T) {
	d := MustNew(2081, Shrawan, 15)
	e := MustNew(2082, Baishakh, 1)
	cases := []struct {
		name string
		f    func()
	}{
		{"New", func() { sinkDate, sinkErr = New(2081, Shrawan, 15) }},
		{"FromGregorian", func() { sinkDate, sinkErr = FromGregorian(2024, time.July, 30) }},
		{"Gregorian", func() { sinkInt, sinkMonth, sinkInt = d.Gregorian() }},
		{"Weekday", func() { sinkWeekday = d.Weekday() }},
		{"YearDay", func() { sinkInt = d.YearDay() }},
		{"MonthLength", func() { sinkInt = d.MonthLength() }},
		{"Compare", func() { sinkInt = d.Compare(e) }},
		{"MonthName", func() { sinkString = Shrawan.Name(Nepali) }},
		{"WeekdayName", func() { sinkString = WeekdayName(time.Tuesday, Nepali) }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if n := testing.AllocsPerRun(100, c.f); n != 0 {
				t.Errorf("%s allocates %.0f times per call, want 0", c.name, n)
			}
		})
	}
}
