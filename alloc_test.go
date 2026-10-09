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
	sinkBool    bool
	sinkDiff    Diff
	sinkFY      FiscalYear
	sinkRange   Range
)

// TestAlloc checks that the paths of ADR-0002 §5 allocate nothing.
func TestAlloc(t *testing.T) {
	d := MustNew(2081, Shrawan, 15)
	e := MustNew(2082, Baishakh, 1)
	month, _ := MonthRange(2081, Shrawan)
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
		{"AddDays", func() { sinkDate, sinkErr = d.AddDays(1000) }},
		{"AddMonthsClamp", func() { sinkDate, sinkErr = d.AddMonths(13, Clamp) }},
		{"AddMonthsSpill", func() { sinkDate, sinkErr = d.AddMonths(13, Spill) }},
		{"AddYears", func() { sinkDate, sinkErr = d.AddYears(-3, Clamp) }},
		{"DaysUntil", func() { sinkInt = d.DaysUntil(e) }},
		{"Diff", func() { sinkDiff = d.Diff(e) }},
		{"Next", func() { sinkDate, sinkBool = d.Next() }},
		{"Prev", func() { sinkDate, sinkBool = d.Prev() }},
		{"FiscalYearOf", func() { sinkFY = FiscalYearOf(d) }},
		{"FiscalYearStart", func() { sinkDate, sinkErr = FiscalYear(2081).Start() }},
		{"FiscalYearQuarterRange", func() { sinkRange, sinkErr = FiscalYear(2081).QuarterRange(Q4) }},
		{"MonthRange", func() { sinkRange, sinkErr = MonthRange(2081, Shrawan) }},
		{"RangeContains", func() { sinkBool = month.Contains(d) }},
		{"RangeDays", func() {
			for day := range month.Days() {
				sinkDate = day
			}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if n := testing.AllocsPerRun(100, c.f); n != 0 {
				t.Errorf("%s allocates %.0f times per call, want 0", c.name, n)
			}
		})
	}
}
