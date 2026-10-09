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
	sinkBytes   []byte
)

// TestAlloc checks that the paths of ADR-0002 §5 allocate nothing.
func TestAlloc(t *testing.T) {
	d := MustNew(2081, Shrawan, 15)
	e := MustNew(2082, Baishakh, 1)
	month, _ := MonthRange(2081, Shrawan)
	buf := make([]byte, 0, 128)
	text := []byte("2081-04-15")
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
		{"AppendShort", func() { sinkBytes = d.AppendShort(buf[:0], DMY, Dash, false, Nepali) }},
		{"AppendLongEnglish", func() { sinkBytes = d.AppendLong(buf[:0], LongOptions{Weekday: true}, English) }},
		{"AppendLongNepali", func() { sinkBytes = d.AppendLong(buf[:0], LongOptions{Weekday: true}, Nepali) }},
		{"AppendFormatPattern", func() { sinkBytes = d.AppendFormat(buf[:0], "yyyy-MM-dd", English) }},
		{"AppendFormatD", func() { sinkBytes = d.AppendFormat(buf[:0], "D", Nepali) }},
		{"AppendDigits", func() { sinkBytes = Nepali.AppendDigits(buf[:0], 2081, 4) }},
		{"ParseASCII", func() { sinkDate, sinkErr = Parse("2081/04/15") }},
		{"ParseDevanagari", func() { sinkDate, sinkErr = Parse("२०८१/०४/१५") }},
		{"ParseLenient", func() { sinkDate, sinkErr = ParseLenient("15 Shrawan 2080") }},
		{"MonthShortName", func() { sinkString = Ashoj.ShortName(English) }},
		{"AppendText", func() { sinkBytes, sinkErr = d.AppendText(buf[:0]) }},
		{"UnmarshalText", func() { sinkErr = sinkDate.UnmarshalText(text) }},
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

// TestAllocOne checks the methods that allocate exactly the value they return (ADR-0002 §5),
// and that a failed parse allocates at most its *ParseError.
func TestAllocOne(t *testing.T) {
	d := MustNew(2081, Shrawan, 15)
	cases := []struct {
		name string
		max  float64
		f    func()
	}{
		{"String", 1, func() { sinkString = d.String() }},
		{"Short", 1, func() { sinkString = d.Short(YMD, Slash, true, Nepali) }},
		{"Long", 1, func() { sinkString = d.Long(LongOptions{Weekday: true}, Nepali) }},
		{"Format", 1, func() { sinkString = d.Format("MMMM d, yyyy", English) }},
		{"FailedParse", 1, func() { sinkDate, sinkErr = Parse("2081/04/1x") }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			n := testing.AllocsPerRun(100, c.f)
			if c.name == "FailedParse" && n > c.max || c.name != "FailedParse" && n != c.max {
				t.Errorf("%s allocates %.0f times per call, want %.0f", c.name, n, c.max)
			}
		})
	}
}
