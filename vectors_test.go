package nepdate

import (
	"strconv"
	"testing"
	"time"
)

func checkConversionRow(t *testing.T, row map[string]string) {
	t.Helper()
	d := bsDate(t, row["bs"])
	y, m, day := ymd(t, row["ad"])
	gy, gm, gd := d.Gregorian()
	if gy != y || int(gm) != m || gd != day {
		t.Errorf("%s: Gregorian() = %d-%02d-%02d, want %s", row["bs"], gy, gm, gd, row["ad"])
	}
	back, err := FromGregorian(y, time.Month(m), day)
	if err != nil || back != d {
		t.Errorf("FromGregorian(%s) = %v, %v; want %s", row["ad"], back, err, row["bs"])
	}
	if got, want := d.Weekday(), weekdayByName[row["weekday"]]; got != want {
		t.Errorf("%s: Weekday() = %v, want %s", row["bs"], got, row["weekday"])
	}
	if got := strconv.Itoa(d.MonthLength()); got != row["month_length"] {
		t.Errorf("%s: MonthLength() = %s, want %s", row["bs"], got, row["month_length"])
	}
}

func TestMonthBoundaryVectors(t *testing.T) {
	rows := readTable(t, "shared/spec/vectors/month-boundaries.csv", ",", 7176)
	for _, row := range rows {
		checkConversionRow(t, row)
	}
}

func TestGoldenDatesConversion(t *testing.T) {
	rows := readTable(t, "shared/spec/vectors/csharp-golden/dates.tsv", "\t", 1018)
	for _, row := range rows {
		checkConversionRow(t, row)
		d := bsDate(t, row["bs"])
		if got := strconv.Itoa(d.YearDay()); got != row["day_of_year"] {
			t.Errorf("%s: YearDay() = %s, want %s", row["bs"], got, row["day_of_year"])
		}
	}
}
