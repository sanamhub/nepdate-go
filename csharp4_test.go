package nepdate

import (
	"errors"
	"testing"
	"time"
)

// Phase-4 rows of the C# catalogue that belong to package nepdate (see TestCSharp in
// csharp_test.go). CalendarDataTests rows are in package patro.

// relativeClock is the fixed "now" of the IsToday rows. D-14: C# reads the clock; here today is
// FromTime of a fixed instant in Nepal, and yesterday and tomorrow are Prev and Next of it.
var relativeClock = time.Date(2024, time.July, 30, 18, 30, 0, 0, time.UTC)

type relative struct{ yesterday, today, tomorrow Date }

func relativeDays(t *testing.T) relative {
	t.Helper()
	today, err := FromTime(relativeClock.In(time.FixedZone("NPT", 20700)))
	if err != nil {
		t.Fatal(err)
	}
	y, _ := today.Prev()
	tm, _ := today.Next()
	return relative{y, today, tm}
}

// isToday, isYesterday and isTomorrow are what C#'s IsToday, IsYesterday and IsTomorrow become
// with an explicit today (D-14).
func (r relative) is(d Date) (today, yesterday, tomorrow bool) {
	return d == r.today, d == r.yesterday, d == r.tomorrow
}

func relativeRow(pick func(relative) Date, wantToday, wantYesterday, wantTomorrow bool) func(*testing.T) {
	return func(t *testing.T) {
		r := relativeDays(t)
		td, yd, tm := r.is(pick(r))
		if td != wantToday || yd != wantYesterday || tm != wantTomorrow {
			t.Errorf("today %v, yesterday %v, tomorrow %v; want %v, %v, %v", td, yd, tm, wantToday, wantYesterday, wantTomorrow)
		}
	}
}

var (
	pickToday     = func(r relative) Date { return r.today }
	pickYesterday = func(r relative) Date { return r.yesterday }
	pickTomorrow  = func(r relative) Date { return r.tomorrow }
	pickPast      = func(relative) Date { return MustNew(2050, 1, 1) }
	pickPlusTwo   = func(r relative) Date { d, _ := r.today.AddDays(2); return d }
	pickMinusTwo  = func(r relative) Date { d, _ := r.today.AddDays(-2); return d }
)

var csharpRows4 = []csharpRow{
	// Extensions/DateTimeExtensionsTests: DateTime.ToNepaliDate is FromTime.
	{id: "DateTimeExtensionsTests.ToNepaliDate_ValidDateTime_ReturnsCorrectNepaliDate", run: fromTimeIs(time.Date(2023, time.August, 30, 0, 0, 0, 0, time.UTC), nd(2080, 5, 13))},
	{id: "DateTimeExtensionsTests.ToNepaliDate_MinValue_ThrowsArgumentOutOfRangeException", run: fromTimeFails(time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC))},
	{id: "DateTimeExtensionsTests.ToNepaliDate_MaxValue_ThrowsArgumentOutOfRangeException", run: fromTimeFails(time.Date(9999, time.December, 31, 23, 59, 59, 999999999, time.UTC))},
	{id: "DateTimeExtensionsTests.ToNepaliDate_DateBeforeMinSupported_ThrowsArgumentOutOfRangeException", run: fromTimeFails(time.Date(1843, time.January, 1, 0, 0, 0, 0, time.UTC))},
	{id: "DateTimeExtensionsTests.ToNepaliDate_DateAfterMaxSupported_ThrowsArgumentOutOfRangeException", run: fromTimeFails(time.Date(2200, time.January, 1, 0, 0, 0, 0, time.UTC))},

	// Core/NepaliDateIsRelativeTests (catalogue phase D-14)
	{id: "NepaliDateIsRelativeTests.IsToday_TodaysDate_ReturnsTrue", run: relativeRow(pickToday, true, false, false)},
	{id: "NepaliDateIsRelativeTests.IsToday_YesterdaysDate_ReturnsFalse", run: relativeRow(pickYesterday, false, true, false)},
	{id: "NepaliDateIsRelativeTests.IsToday_TomorrowsDate_ReturnsFalse", run: relativeRow(pickTomorrow, false, false, true)},
	{id: "NepaliDateIsRelativeTests.IsToday_ArbitraryPastDate_ReturnsFalse", run: relativeRow(pickPast, false, false, false)},
	{id: "NepaliDateIsRelativeTests.IsYesterday_YesterdaysDate_ReturnsTrue", run: relativeRow(pickYesterday, false, true, false)},
	{id: "NepaliDateIsRelativeTests.IsYesterday_TodaysDate_ReturnsFalse", run: relativeRow(pickToday, true, false, false)},
	{id: "NepaliDateIsRelativeTests.IsYesterday_TomorrowsDate_ReturnsFalse", run: relativeRow(pickTomorrow, false, false, true)},
	{id: "NepaliDateIsRelativeTests.IsYesterday_ArbitraryPastDate_ReturnsFalse", run: relativeRow(pickPast, false, false, false)},
	{id: "NepaliDateIsRelativeTests.IsTomorrow_TomorrowsDate_ReturnsTrue", run: relativeRow(pickTomorrow, false, false, true)},
	{id: "NepaliDateIsRelativeTests.IsTomorrow_TodaysDate_ReturnsFalse", run: relativeRow(pickToday, true, false, false)},
	{id: "NepaliDateIsRelativeTests.IsTomorrow_YesterdaysDate_ReturnsFalse", run: relativeRow(pickYesterday, false, true, false)},
	{id: "NepaliDateIsRelativeTests.IsTomorrow_ArbitraryPastDate_ReturnsFalse", run: relativeRow(pickPast, false, false, false)},
	{id: "NepaliDateIsRelativeTests.Today_OnlyIsToday_IsTrue", run: relativeRow(pickToday, true, false, false)},
	{id: "NepaliDateIsRelativeTests.Yesterday_OnlyIsYesterday_IsTrue", run: relativeRow(pickYesterday, false, true, false)},
	{id: "NepaliDateIsRelativeTests.Tomorrow_OnlyIsTomorrow_IsTrue", run: relativeRow(pickTomorrow, false, false, true)},
	{id: "NepaliDateIsRelativeTests.TwoDaysAhead_NoneAreTrue", run: relativeRow(pickPlusTwo, false, false, false)},
	{id: "NepaliDateIsRelativeTests.TwoDaysBehind_NoneAreTrue", run: relativeRow(pickMinusTwo, false, false, false)},
}

func fromTimeFails(tm time.Time) func(*testing.T) {
	return func(t *testing.T) {
		got, err := FromTime(tm)
		if !errors.Is(err, ErrOutOfRange) || !got.IsZero() {
			t.Errorf("FromTime(%v) = %v, %v; want ErrOutOfRange", tm, got, err)
		}
	}
}
