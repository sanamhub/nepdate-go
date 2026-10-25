package nepdate

import (
	"errors"
	"slices"
	"testing"
	"time"
)

// C# catalogue rows for arithmetic, fiscal years and ranges (see TestCSharp in csharp_test.go).

const (
	d02       = "D-02: AddMonths takes whole months only"
	d06       = "D-06: autoAdjust guesses the field order or expands 2-digit years"
	d09       = "D-09: Range has no set operations, splitting, working days or intervals"
	d09String = "D-09: Range has no String method; format Start and End instead"
)

var csharpArithRows = []csharpRow{
	// Core/FiscalYearTests (D-08 shape: a FiscalYear value instead of static methods)
	{id: "FiscalYearTests.GetFiscalYearStartAndEndDate_ValidYear_ReturnsCorrectDates", run: fyBounds(2080, nd(2080, 4, 1), lastOf(2081, 3))},
	{id: "FiscalYearTests.GetFiscalYearStartAndEndDate_InvalidYear_ThrowsException#1", run: fyOutOfRange(1843)},
	{id: "FiscalYearTests.GetFiscalYearStartAndEndDate_InvalidYear_ThrowsException#2", run: fyOutOfRange(2200)},
	{id: "FiscalYearTests.GetFiscalYearStartDate_ValidYear_ReturnsCorrectDate", run: fyBounds(2080, nd(2080, 4, 1), Date{})},
	{id: "FiscalYearTests.GetFiscalYearEndDate_ValidYear_ReturnsCorrectDate", run: fyBounds(2080, Date{}, lastOf(2081, 3))},
	{id: "FiscalYearTests.FiscalYearStartAndEndDate_OnNepaliDateInstance_ReturnsCorrectDates", run: fyBounds(nd(2080, 6, 15).FiscalYear(), nd(2080, 4, 1), lastOf(2081, 3))},
	{id: "FiscalYearTests.FiscalYearStartDate_OnNepaliDateInstance_ReturnsCorrectDate", run: fyBounds(nd(2080, 6, 15).FiscalYear(), nd(2080, 4, 1), Date{})},
	{id: "FiscalYearTests.FiscalYearEndDate_OnNepaliDateInstance_ReturnsCorrectDate", run: fyBounds(nd(2080, 6, 15).FiscalYear(), Date{}, lastOf(2081, 3))},
	// C# passes a month; its quarter is that month's Quarter (D-08).
	{id: "FiscalYearTests.GetFiscalYearQuarterStartAndEndDate_ValidYearAndMonth_ReturnsCorrectDates", run: quarterBounds(2080, nd(2080, 5, 1).Quarter(), nd(2080, 4, 1), lastOf(2080, 6))},
	{id: "FiscalYearTests.GetFiscalYearQuarterStartDate_ValidYearAndMonth_ReturnsCorrectDate", run: quarterBounds(2080, nd(2080, 8, 1).Quarter(), nd(2080, 7, 1), Date{})},
	{id: "FiscalYearTests.GetFiscalYearQuarterEndDate_ValidYearAndMonth_ReturnsCorrectDate", run: quarterBounds(2080, nd(2080, 11, 1).Quarter(), Date{}, lastOf(2080, 12))},

	// Core/FiscalYearInstanceMethodTests (D-08 shape: yearOffset is FiscalYear arithmetic,
	// FiscalYearQuarters.Current is the date's own Quarter)
	{id: "FiscalYearInstanceMethodTests.FiscalYearStartDate_DateInShrawan_ReturnsSameyearShrawan1", run: fyBounds(nd(2080, 4, 15).FiscalYear(), nd(2080, 4, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearStartDate_DateInChaitra_ReturnsSameYearShrawan1", run: fyBounds(nd(2080, 12, 15).FiscalYear(), nd(2080, 4, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearStartDate_DateInBasikhakh_ReturnsPreviousYearShrawan1", run: fyBounds(nd(2080, 1, 15).FiscalYear(), nd(2079, 4, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearStartDate_DateInAshadh_ReturnsPreviousYearShrawan1", run: fyBounds(nd(2080, 3, 10).FiscalYear(), nd(2079, 4, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearStartDate_WithPositiveYearOffset_ShiftsForward", run: fyBounds(nd(2080, 4, 1).FiscalYear()+1, nd(2081, 4, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearStartDate_WithNegativeYearOffset_ShiftsBackward", run: fyBounds(nd(2080, 4, 1).FiscalYear()-1, nd(2079, 4, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearEndDate_DateInShrawan_ReturnsNextYearAshad", run: fyBounds(nd(2080, 4, 15).FiscalYear(), Date{}, lastOf(2081, 3))},
	{id: "FiscalYearInstanceMethodTests.FiscalYearEndDate_DateInChaitra_ReturnsNextYearAshadh", run: fyBounds(nd(2080, 12, 15).FiscalYear(), Date{}, lastOf(2081, 3))},
	{id: "FiscalYearInstanceMethodTests.FiscalYearEndDate_DateInBasikhakh_ReturnsSameYearAshadh", run: fyBounds(nd(2080, 1, 15).FiscalYear(), Date{}, lastOf(2080, 3))},
	{id: "FiscalYearInstanceMethodTests.FiscalYearEndDate_WithPositiveYearOffset_ShiftsForward", run: fyBounds(nd(2080, 4, 1).FiscalYear()+1, Date{}, lastOf(2082, 3))},
	{id: "FiscalYearInstanceMethodTests.FiscalYearStartAndEndDate_DateInQ2_ReturnsBothBoundaries", run: fyBounds(nd(2080, 8, 15).FiscalYear(), nd(2080, 4, 1), lastOf(2081, 3))},
	{id: "FiscalYearInstanceMethodTests.FiscalYearStartAndEndDate_DateInQ4_ReturnsPreviousFiscalYearBoundaries", run: fyBounds(nd(2081, 2, 10).FiscalYear(), nd(2080, 4, 1), lastOf(2081, 3))},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterStartDate_ExplicitFirst_ReturnsShrawan1", run: quarterBounds(nd(2080, 5, 15).FiscalYear(), Q1, nd(2080, 4, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterStartDate_ExplicitSecond_ReturnsKartik1", run: quarterBounds(nd(2080, 5, 15).FiscalYear(), Q2, nd(2080, 7, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterStartDate_ExplicitThird_ReturnsMagh1", run: quarterBounds(nd(2080, 5, 15).FiscalYear(), Q3, nd(2080, 10, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterStartDate_ExplicitFourth_ReturnsNextYearBasikhakh1", run: quarterBounds(nd(2080, 5, 15).FiscalYear(), Q4, nd(2081, 1, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterEndDate_ExplicitFirst_ReturnsLastDayOfAshoj", run: quarterBounds(nd(2080, 5, 15).FiscalYear(), Q1, Date{}, lastOf(2080, 6))},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterEndDate_ExplicitSecond_ReturnsLastDayOfPoush", run: quarterBounds(nd(2080, 5, 15).FiscalYear(), Q2, Date{}, lastOf(2080, 9))},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterEndDate_ExplicitThird_ReturnsLastDayOfChaitra", run: quarterBounds(nd(2080, 5, 15).FiscalYear(), Q3, Date{}, lastOf(2080, 12))},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterEndDate_ExplicitFourth_ReturnsNextYearLastDayOfAshadh", run: quarterBounds(nd(2080, 5, 15).FiscalYear(), Q4, Date{}, lastOf(2081, 3))},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterStartAndEndDate_ExplicitSecond_ReturnsBothBoundaries", run: quarterBounds(nd(2080, 5, 15).FiscalYear(), Q2, nd(2080, 7, 1), lastOf(2080, 9))},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterStartAndEndDate_ExplicitThird_ReturnsBothBoundaries", run: quarterBounds(nd(2080, 5, 15).FiscalYear(), Q3, nd(2080, 10, 1), lastOf(2080, 12))},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterStartDate_Current_DateInQ1_ReturnsShrawan1", run: quarterBounds(nd(2080, 5, 15).FiscalYear(), nd(2080, 5, 15).Quarter(), nd(2080, 4, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterEndDate_Current_DateInQ1_ReturnsLastDayOfAshoj", run: quarterBounds(nd(2080, 5, 15).FiscalYear(), nd(2080, 5, 15).Quarter(), Date{}, lastOf(2080, 6))},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterStartDate_Current_DateInQ2_ReturnsKartik1", run: quarterBounds(nd(2080, 8, 15).FiscalYear(), nd(2080, 8, 15).Quarter(), nd(2080, 7, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterStartDate_Current_DateInQ3_ReturnsMagh1", run: quarterBounds(nd(2080, 11, 15).FiscalYear(), nd(2080, 11, 15).Quarter(), nd(2080, 10, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterEndDate_Current_DateInQ3_ReturnsLastDayOfChaitra", run: quarterBounds(nd(2080, 11, 15).FiscalYear(), nd(2080, 11, 15).Quarter(), Date{}, lastOf(2080, 12))},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterStartDate_ExplicitFirstWithOffset1_ReturnsNextYearShrawan1", run: quarterBounds(nd(2080, 5, 15).FiscalYear()+1, Q1, nd(2081, 4, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterStartDate_ExplicitFirstWithNegativeOffset_ReturnsPreviousYearShrawan1", run: quarterBounds(nd(2080, 5, 15).FiscalYear()-1, Q1, nd(2079, 4, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearStartDate_AllFiscalMonths_Return2080Shrawan1#1", run: fyBounds(nd(2080, 4, 1).FiscalYear(), nd(2080, 4, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearStartDate_AllFiscalMonths_Return2080Shrawan1#2", run: fyBounds(nd(2080, 6, 1).FiscalYear(), nd(2080, 4, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearStartDate_AllFiscalMonths_Return2080Shrawan1#3", run: fyBounds(nd(2080, 7, 1).FiscalYear(), nd(2080, 4, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearStartDate_AllFiscalMonths_Return2080Shrawan1#4", run: fyBounds(nd(2080, 12, 1).FiscalYear(), nd(2080, 4, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearStartDate_Q4CalendarMonths_ReturnPreviousYearShrawan1#1", run: fyBounds(nd(2080, 1, 1).FiscalYear(), nd(2079, 4, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearStartDate_Q4CalendarMonths_ReturnPreviousYearShrawan1#2", run: fyBounds(nd(2080, 2, 1).FiscalYear(), nd(2079, 4, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearStartDate_Q4CalendarMonths_ReturnPreviousYearShrawan1#3", run: fyBounds(nd(2080, 3, 1).FiscalYear(), nd(2079, 4, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterStartDate_FromQ4Date_ExplicitFirst_ReturnsFY2080Q1Start", run: quarterBounds(nd(2081, 2, 10).FiscalYear(), Q1, nd(2080, 4, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterStartDate_FromQ4Date_ExplicitSecond_ReturnsFY2080Q2Start", run: quarterBounds(nd(2081, 2, 10).FiscalYear(), Q2, nd(2080, 7, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterStartDate_FromQ4Date_ExplicitThird_ReturnsFY2080Q3Start", run: quarterBounds(nd(2081, 2, 10).FiscalYear(), Q3, nd(2080, 10, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterStartDate_FromQ4Date_ExplicitFourth_ReturnsFY2080Q4Start", run: quarterBounds(nd(2081, 2, 10).FiscalYear(), Q4, nd(2081, 1, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterStartDate_FromQ4Date_Current_ReturnsFY2080Q4Start", run: quarterBounds(nd(2081, 2, 10).FiscalYear(), nd(2081, 2, 10).Quarter(), nd(2081, 1, 1), Date{})},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterEndDate_FromQ4Date_ExplicitFirst_ReturnsFY2080Q1End", run: quarterBounds(nd(2081, 2, 10).FiscalYear(), Q1, Date{}, lastOf(2080, 6))},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterEndDate_FromQ4Date_ExplicitThird_ReturnsFY2080Q3End", run: quarterBounds(nd(2081, 2, 10).FiscalYear(), Q3, Date{}, lastOf(2080, 12))},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterEndDate_FromQ4Date_ExplicitFourth_ReturnsFY2080Q4End", run: quarterBounds(nd(2081, 2, 10).FiscalYear(), Q4, Date{}, lastOf(2081, 3))},
	{id: "FiscalYearInstanceMethodTests.FiscalYearQuarterEndDate_FromQ4Date_Current_ReturnsFY2080Q4End", run: quarterBounds(nd(2081, 2, 10).FiscalYear(), nd(2081, 2, 10).Quarter(), Date{}, lastOf(2081, 3))},

	// Core/NepaliDateArithmeticEdgeCaseTests
	{id: "NepaliDateArithmeticEdgeCaseTests.AddDays_Zero_ReturnsSameDate", run: addDays(nd(2080, 5, 15), 0, nd(2080, 5, 15))},
	{id: "NepaliDateArithmeticEdgeCaseTests.AddDays_PositiveAcrossMonthBoundary_IsCorrect", run: addDays(nd(2080, 5, 31), 1, nd(2080, 6, 1))},
	{id: "NepaliDateArithmeticEdgeCaseTests.AddDays_NegativeAcrossMonthBoundary_IsCorrect", run: addDays(nd(2080, 6, 1), -1, nd(2080, 5, 31))},
	{id: "NepaliDateArithmeticEdgeCaseTests.AddDays_PositiveAcrossYearBoundary_IsCorrect", run: addDays(lastOf(2080, 12), 1, nd(2081, 1, 1))},
	{id: "NepaliDateArithmeticEdgeCaseTests.AddDays_NegativeAcrossYearBoundary_IsCorrect", run: addDays(nd(2081, 1, 1), -1, lastOf(2080, 12))},
	{id: "NepaliDateArithmeticEdgeCaseTests.AddDays_LargePositiveValue_IsCorrect", run: func(t *testing.T) {
		start := nd(2080, 1, 1)
		got, err := start.AddDays(365)
		check(t, err == nil && (got.Year() == 2080 || got.Year() == 2081) && !got.Before(start), "365 days later is in 2080 or 2081")
	}},
	{id: "NepaliDateArithmeticEdgeCaseTests.AddMonths_PositiveCrossesYearBoundary_IsCorrect", run: addMonths(nd(2080, 11, 15), 5, Clamp, nd(2081, 4, 15))},
	{id: "NepaliDateArithmeticEdgeCaseTests.AddMonths_NegativeCrossesYearBoundary_IsCorrect", run: addMonths(nd(2080, 3, 15), -5, Clamp, nd(2079, 10, 15))},
	{id: "NepaliDateArithmeticEdgeCaseTests.AddMonths_PositiveMoreThan12_CrossesMultipleYears", run: addMonths(nd(2080, 5, 15), 14, Clamp, nd(2081, 7, 15))},
	{id: "NepaliDateArithmeticEdgeCaseTests.AddMonths_Exactly12_SameMonthNextYear", run: addMonths(nd(2080, 6, 15), 12, Clamp, nd(2081, 6, 15))},
	{id: "NepaliDateArithmeticEdgeCaseTests.AddMonths_DayCappedAtMonthEnd_WhenResultMonthIsShorter", run: addMonths(nd(2081, 4, 32), 5, Clamp, lastOf(2081, 9))},
	{id: "NepaliDateArithmeticEdgeCaseTests.AddMonths_Negative_SameDayPreservedWhenFits", run: addMonths(nd(2080, 8, 10), -1, Clamp, nd(2080, 7, 10))},
	{id: "NepaliDateArithmeticEdgeCaseTests.AddMonths_HalfMonth_AddsApproximately15Days", skip: d02},
	{id: "NepaliDateArithmeticEdgeCaseTests.AddMonths_NegativeHalfMonth_SubtractsApproximately15Days", skip: d02},
	{id: "NepaliDateArithmeticEdgeCaseTests.AddMonths_OneAndHalfMonths_AddsApproximately46Days", skip: d02},
	{id: "NepaliDateArithmeticEdgeCaseTests.AddMonths_FractionalValue_ReturnsDateFurtherThanZeroMonths", skip: d02},
	{id: "NepaliDateArithmeticEdgeCaseTests.Constructor_DateTimeWithMidnightTime_IgnoresTime", run: fromTimeIs(time.Date(2023, time.September, 1, 0, 0, 0, 0, time.UTC), nd(2080, 5, 15))},
	{id: "NepaliDateArithmeticEdgeCaseTests.Constructor_DateTimeWithNoonTime_IgnoresTime", run: fromTimeIs(time.Date(2023, time.September, 1, 12, 30, 45, 0, time.UTC), nd(2080, 5, 15))},
	{id: "NepaliDateArithmeticEdgeCaseTests.Constructor_DateTimeWithEndOfDayTime_IgnoresTime", run: fromTimeIs(time.Date(2023, time.September, 1, 23, 59, 59, 0, time.UTC), nd(2080, 5, 15))},
	{id: "NepaliDateArithmeticEdgeCaseTests.Constructor_TwoDateTimesOnSameEnglishDateDifferentTimes_ProduceSameNepaliDate", run: func(t *testing.T) {
		morning, err1 := FromTime(time.Date(2023, time.January, 1, 6, 0, 0, 0, time.UTC))
		evening, err2 := FromTime(time.Date(2023, time.January, 1, 22, 0, 0, 0, time.UTC))
		check(t, err1 == nil && err2 == nil && morning == evening, "same AD date, same BS date")
	}},
	{id: "NepaliDateArithmeticEdgeCaseTests.TryParse_AutoAdjust_DayYearMonthFormat_ReturnsCorrectDate", skip: d06},
	{id: "NepaliDateArithmeticEdgeCaseTests.TryParse_AutoAdjust_MonthDayYearInvertedOrder_ReturnsCorrectDate", skip: d06},
	{id: "NepaliDateArithmeticEdgeCaseTests.TryParse_AutoAdjust_TwoDigitYear_ExpandsToCurrentMillennium", skip: d06},
	{id: "NepaliDateArithmeticEdgeCaseTests.TryParse_AutoAdjust_InvalidDateAfterAdjustment_ReturnsFalse", skip: d06},
	{id: "NepaliDateArithmeticEdgeCaseTests.TryParse_AutoAdjust_ValidStandardFormat_ReturnsCorrectDate", run: parsesTo("2080/05/15", nd(2080, 5, 15))},

	// Core/NepaliDateManipulationTests
	{id: "NepaliDateManipulationTests.MonthEndDate_ReturnsCorrectDate", run: func(t *testing.T) {
		check(t, nd(2080, 5, 15).LastOfMonth() == nd(2080, 5, 31), "Bhadra 2080 ends on day 31")
	}},
	{id: "NepaliDateManipulationTests.AddMonths_ReturnsCorrectDate#1", run: addMonths(nd(2080, 5, 15), 1, Clamp, nd(2080, 6, 15))},
	{id: "NepaliDateManipulationTests.AddMonths_ReturnsCorrectDate#2", run: addMonths(nd(2080, 5, 15), -1, Clamp, nd(2080, 4, 15))},
	{id: "NepaliDateManipulationTests.AddMonths_ReturnsCorrectDate#3", run: addMonths(nd(2080, 5, 15), 12, Clamp, nd(2081, 5, 15))},
	{id: "NepaliDateManipulationTests.AddMonths_ReturnsCorrectDate#4", run: addMonths(nd(2081, 4, 32), 5, Clamp, nd(2081, 9, 29))},
	{id: "NepaliDateManipulationTests.AddMonths_ReturnsCorrectDate#5", run: addMonths(nd(2081, 4, 32), 5, Spill, nd(2081, 10, 3))},
	{id: "NepaliDateManipulationTests.AddDays_ReturnsCorrectDate#1", run: addDays(nd(2080, 5, 15), 5, nd(2080, 5, 20))},
	{id: "NepaliDateManipulationTests.AddDays_ReturnsCorrectDate#2", run: addDays(nd(2080, 5, 15), -5, nd(2080, 5, 10))},
	{id: "NepaliDateManipulationTests.AddDays_ReturnsCorrectDate#3", run: addDays(nd(2080, 5, 31), 1, nd(2080, 6, 1))},
	{id: "NepaliDateManipulationTests.AddDays_ReturnsCorrectDate#4", run: addDays(nd(2080, 12, 30), 5, nd(2081, 1, 5))},
	// C#'s TimeSpan subtraction is DaysUntil in Go (D-07).
	{id: "NepaliDateManipulationTests.Subtract_TwoDates_ReturnsCorrectTimeSpan", run: func(t *testing.T) {
		a, b := nd(2080, 5, 1), nd(2080, 5, 16)
		check(t, a.DaysUntil(b) == 15 && b.DaysUntil(a) == -15, "15 days apart, signed")
	}},
	{id: "NepaliDateManipulationTests.Subtract_SameDate_ReturnsZero", run: daysUntil(nd(2080, 5, 15), nd(2080, 5, 15), 0)},
	{id: "OptimizationVerificationTests.Subtraction_SameDate_ReturnsZeroDays", run: daysUntil(nd(2080, 5, 15), nd(2080, 5, 15), 0)},
	{id: "OptimizationVerificationTests.Subtraction_AdjacentDays_ReturnsOneDay", run: func(t *testing.T) {
		a, b := nd(2080, 5, 15), nd(2080, 5, 16)
		check(t, a.DaysUntil(b) == 1 && b.DaysUntil(a) == -1, "adjacent days are 1 apart, signed")
	}},
	{id: "OptimizationVerificationTests.Subtraction_CrossMonth_CorrectDays", run: daysUntil(nd(2080, 4, 32), nd(2080, 5, 1), 1)},
	{id: "OptimizationVerificationTests.Subtraction_CrossYear_CorrectDays", run: daysUntil(lastOf(2080, 12), nd(2081, 1, 1), 1)},

	// Core/NepaliDateRangeTests
	{id: "NepaliDateRangeTests.Constructor_ValidDates_CreatesRange", run: func(t *testing.T) {
		r, ok := NewRange(nd(2080, 1, 1), nd(2080, 1, 15))
		check(t, ok && r.Start() == nd(2080, 1, 1) && r.End() == nd(2080, 1, 15) && r.Len() == 15, "2080-01-01..15 has 15 days")
	}},
	// C# returns an empty range; Go reports false and the zero Range, which is empty.
	{id: "NepaliDateRangeTests.Constructor_EndBeforeStart_CreatesEmptyRange", run: func(t *testing.T) {
		r, ok := NewRange(nd(2080, 1, 15), nd(2080, 1, 1))
		check(t, !ok && r.Len() == 0, "end before start gives no range")
	}},
	{id: "NepaliDateRangeTests.SingleDay_CreatesRangeWithOneDay", run: func(t *testing.T) {
		r, ok := NewRange(nd(2080, 1, 1), nd(2080, 1, 1))
		check(t, ok && r.Start() == r.End() && r.Len() == 1, "a one-day range")
	}},
	{id: "NepaliDateRangeTests.FromDayCount_CreatesCorrectRange", skip: d09},
	{id: "NepaliDateRangeTests.ForMonth_CreatesCorrectRange", run: func(t *testing.T) {
		r, err := MonthRange(2080, 1)
		check(t, err == nil && r.Start() == nd(2080, 1, 1) && r.End() == lastOf(2080, 1), "Baishakh 2080")
	}},
	{id: "NepaliDateRangeTests.ForFiscalYear_CreatesCorrectRange", run: func(t *testing.T) {
		r, err := FiscalYear(2080).Range()
		check(t, err == nil && r.Start() == nd(2080, 4, 1) && r.End() == lastOf(2081, 3), "FY 2080/81")
	}},
	{id: "NepaliDateRangeTests.Contains_DateInRange_ReturnsTrue", run: contains(nd(2080, 1, 1), nd(2080, 1, 15), nd(2080, 1, 10), true)},
	{id: "NepaliDateRangeTests.Contains_DateOutsideRange_ReturnsFalse", run: contains(nd(2080, 1, 1), nd(2080, 1, 15), nd(2080, 1, 20), false)},
	{id: "NepaliDateRangeTests.Contains_RangeFullyContained_ReturnsTrue", skip: d09},
	{id: "NepaliDateRangeTests.Overlaps_RangesOverlap_ReturnsTrue", skip: d09},
	{id: "NepaliDateRangeTests.Overlaps_RangesDontOverlap_ReturnsFalse", skip: d09},
	{id: "NepaliDateRangeTests.IsAdjacentTo_AdjacentRanges_ReturnsTrue", skip: d09},
	{id: "NepaliDateRangeTests.Intersect_OverlappingRanges_ReturnsIntersection", skip: d09},
	{id: "NepaliDateRangeTests.Union_OverlappingRanges_ReturnsUnion", skip: d09},
	{id: "NepaliDateRangeTests.Except_RemovingRangeFromMiddle_ReturnsTwoRanges", skip: d09},
	{id: "NepaliDateRangeTests.SplitByMonth_RangeAcrossMonths_SplitsCorrectly", skip: d09},
	{id: "NepaliDateRangeTests.WorkingDays_ReturnsCorrectDays", skip: d09},
	{id: "NepaliDateRangeTests.Enumeration_EnumeratesAllDaysInRange", run: func(t *testing.T) {
		r, _ := NewRange(nd(2080, 1, 1), nd(2080, 1, 10))
		var want []Date
		for i := 1; i <= 10; i++ {
			want = append(want, nd(2080, 1, i))
		}
		check(t, slices.Equal(slices.Collect(r.Days()), want), "Days yields 2080-01-01..10")
	}},
	{id: "NepaliDateRangeTests.Demo_RangeOperations", skip: d09},

	// Core/NepaliDateRangeExtendedTests
	{id: "NepaliDateRangeExtendedTests.ForCalendarYear_StartsAtBaishakhOne_EndsAtChaitraEnd", run: func(t *testing.T) {
		r, err := YearRange(2080)
		check(t, err == nil && r.Start() == nd(2080, 1, 1) && r.End() == lastOf(2080, 12), "BS year 2080")
	}},
	{id: "NepaliDateRangeExtendedTests.ForCalendarYear_LengthIsReasonable", run: func(t *testing.T) {
		r, err := YearRange(2080)
		check(t, err == nil && r.Len() >= 365 && r.Len() <= 366, "BS 2080 has 365 or 366 days")
	}},
	// Current* rows read the clock in C#. D-14: the same structure is checked for a fixed date
	// (BS 2081-04-15) standing in for today.
	{id: "NepaliDateRangeExtendedTests.CurrentMonth_StartIsFirstDayOfCurrentMonth", run: func(t *testing.T) {
		r, err := MonthRange(fixedToday().Year(), fixedToday().Month())
		check(t, err == nil && r.Start() == fixedToday().FirstOfMonth(), "the month starts on day 1")
	}},
	{id: "NepaliDateRangeExtendedTests.CurrentMonth_EndIsLastDayOfCurrentMonth", run: func(t *testing.T) {
		r, err := MonthRange(fixedToday().Year(), fixedToday().Month())
		check(t, err == nil && r.End() == fixedToday().LastOfMonth(), "the month ends on its last day")
	}},
	{id: "NepaliDateRangeExtendedTests.CurrentMonth_ContainsToday", run: func(t *testing.T) {
		r, err := MonthRange(fixedToday().Year(), fixedToday().Month())
		check(t, err == nil && r.Contains(fixedToday()), "the month contains the date")
	}},
	{id: "NepaliDateRangeExtendedTests.CurrentFiscalYear_ContainsToday", run: func(t *testing.T) {
		r, err := fixedToday().FiscalYear().Range()
		check(t, err == nil && r.Contains(fixedToday()), "the fiscal year contains the date")
	}},
	{id: "NepaliDateRangeExtendedTests.CurrentFiscalYear_StartIsShrawan", run: func(t *testing.T) {
		r, err := fixedToday().FiscalYear().Range()
		check(t, err == nil && r.Start().Month() == Shrawan && r.Start().Day() == 1, "the fiscal year starts on Shrawan 1")
	}},
	{id: "NepaliDateRangeExtendedTests.CurrentFiscalYear_EndIsAshadh", run: func(t *testing.T) {
		r, err := fixedToday().FiscalYear().Range()
		check(t, err == nil && r.End().Month() == Ashad && r.End() == r.End().LastOfMonth(), "the fiscal year ends on the last day of Ashad")
	}},
	{id: "NepaliDateRangeExtendedTests.CurrentCalendarYear_ContainsToday", run: func(t *testing.T) {
		r, err := YearRange(fixedToday().Year())
		check(t, err == nil && r.Contains(fixedToday()), "the BS year contains the date")
	}},
	{id: "NepaliDateRangeExtendedTests.CurrentCalendarYear_StartIsBaishakh1", run: func(t *testing.T) {
		r, err := YearRange(fixedToday().Year())
		check(t, err == nil && r.Start().Month() == Baishakh && r.Start().Day() == 1, "the BS year starts on Baishakh 1")
	}},
	{id: "NepaliDateRangeExtendedTests.SplitByFiscalQuarter_EmptyRange_ReturnsEmptyArray", skip: d09},
	{id: "NepaliDateRangeExtendedTests.SplitByFiscalQuarter_RangeWithinOneQuarter_ReturnsSingleRange", skip: d09},
	{id: "NepaliDateRangeExtendedTests.SplitByFiscalQuarter_RangeSpanningTwoQuarters_ReturnsTwoRanges", skip: d09},
	{id: "NepaliDateRangeExtendedTests.SplitByFiscalQuarter_FullFiscalYear_ReturnsFourRanges", skip: d09},
	{id: "NepaliDateRangeExtendedTests.SplitByFiscalQuarter_RangeInQ4Months_SingleRange", skip: d09},
	{id: "NepaliDateRangeExtendedTests.DatesWithInterval_Zero_ThrowsArgumentOutOfRangeException", skip: d09},
	{id: "NepaliDateRangeExtendedTests.DatesWithInterval_NegativeInterval_ThrowsArgumentOutOfRangeException", skip: d09},
	{id: "NepaliDateRangeExtendedTests.DatesWithInterval_IntervalOfOne_YieldsEveryDay", skip: d09},
	{id: "NepaliDateRangeExtendedTests.DatesWithInterval_IntervalOfTwo_YieldsOddDays", skip: d09},
	{id: "NepaliDateRangeExtendedTests.DatesWithInterval_IntervalLargerThanRange_YieldsOnlyStartDate", skip: d09},
	{id: "NepaliDateRangeExtendedTests.DatesWithInterval_EmptyRange_YieldsNothing", skip: d09},
	{id: "NepaliDateRangeExtendedTests.WeekendDays_IncludeSundayTrue_ReturnsBothSaturdayAndSunday", skip: d09},
	{id: "NepaliDateRangeExtendedTests.WeekendDays_IncludeSundayFalse_ReturnsOnlySaturday", skip: d09},
	{id: "NepaliDateRangeExtendedTests.WeekendDays_RangeWithNoWeekend_ReturnsEmpty", skip: d09},
	{id: "NepaliDateRangeExtendedTests.WeekendDays_EmptyRange_ReturnsEmpty", skip: d09},
	{id: "NepaliDateRangeExtendedTests.WorkingDays_ExcludeSundayFalse_ExcludesOnlySaturday", skip: d09},
	{id: "NepaliDateRangeExtendedTests.WorkingDays_ExcludeSundayTrue_ExcludesBothSatAndSun", skip: d09},
	{id: "NepaliDateRangeExtendedTests.WorkingDays_EmptyRange_ReturnsEmpty", skip: d09},
	{id: "NepaliDateRangeExtendedTests.FromDayCount_ZeroDays_ThrowsArgumentOutOfRangeException", skip: d09},
	{id: "NepaliDateRangeExtendedTests.FromDayCount_NegativeDays_ThrowsArgumentOutOfRangeException", skip: d09},
	{id: "NepaliDateRangeExtendedTests.Intersect_NonOverlappingRanges_ReturnsEmptyRange", skip: d09},
	{id: "NepaliDateRangeExtendedTests.Intersect_AdjacentRanges_ReturnsEmptyRange", skip: d09},
	{id: "NepaliDateRangeExtendedTests.Intersect_OneRangeFullyInsideOther_ReturnsInnerRange", skip: d09},
	{id: "NepaliDateRangeExtendedTests.Union_LeftRangeIsEmpty_ReturnsRightRange", skip: d09},
	{id: "NepaliDateRangeExtendedTests.Union_RightRangeIsEmpty_ReturnsLeftRange", skip: d09},
	{id: "NepaliDateRangeExtendedTests.Union_NonOverlappingRanges_SpansEntireGap", skip: d09},
	{id: "NepaliDateRangeExtendedTests.Contains_ExactStartDate_ReturnsTrue", run: contains(nd(2080, 1, 5), nd(2080, 1, 20), nd(2080, 1, 5), true)},
	{id: "NepaliDateRangeExtendedTests.Contains_ExactEndDate_ReturnsTrue", run: contains(nd(2080, 1, 5), nd(2080, 1, 20), nd(2080, 1, 20), true)},
	{id: "NepaliDateRangeExtendedTests.Contains_OneDayBeforeStart_ReturnsFalse", run: contains(nd(2080, 1, 5), nd(2080, 1, 20), nd(2080, 1, 4), false)},
	{id: "NepaliDateRangeExtendedTests.Contains_OneDayAfterEnd_ReturnsFalse", run: contains(nd(2080, 1, 5), nd(2080, 1, 20), nd(2080, 1, 21), false)},
	{id: "NepaliDateRangeExtendedTests.Contains_EmptyRange_ContainsNothing", run: func(t *testing.T) {
		r, _ := NewRange(nd(2080, 5, 10), nd(2080, 5, 1))
		check(t, !r.Contains(nd(2080, 5, 5)), "the empty range contains nothing")
	}},
	{id: "NepaliDateRangeExtendedTests.IsAdjacentTo_GapBetweenRanges_ReturnsFalse", skip: d09},
	{id: "NepaliDateRangeExtendedTests.IsAdjacentTo_OverlappingRanges_ReturnsFalse", skip: d09},
	{id: "NepaliDateRangeExtendedTests.IsAdjacentTo_EmptyRange_ReturnsFalse", skip: d09},
	{id: "NepaliDateRangeExtendedTests.Except_NonOverlappingExcludeRange_ReturnsOriginalRange", skip: d09},
	{id: "NepaliDateRangeExtendedTests.Except_ExcludeRangeFullyContainsThis_ReturnsEmptyArray", skip: d09},
	{id: "NepaliDateRangeExtendedTests.Except_ExcludeOverlapsStart_ReturnsRightPart", skip: d09},
	{id: "NepaliDateRangeExtendedTests.Except_ExcludeOverlapsEnd_ReturnsLeftPart", skip: d09},
	{id: "NepaliDateRangeExtendedTests.GetEnumerator_ThreeDayRange_YieldsExactlyThreeDates", run: func(t *testing.T) {
		r, _ := NewRange(nd(2080, 5, 13), nd(2080, 5, 15))
		check(t, slices.Equal(slices.Collect(r.Days()), []Date{nd(2080, 5, 13), nd(2080, 5, 14), nd(2080, 5, 15)}), "three days")
	}},
	{id: "NepaliDateRangeExtendedTests.GetEnumerator_EmptyRange_YieldsNoDates", run: func(t *testing.T) {
		r, _ := NewRange(nd(2080, 5, 10), nd(2080, 5, 1))
		check(t, len(slices.Collect(r.Days())) == 0, "no days")
	}},
	{id: "NepaliDateRangeExtendedTests.GetEnumerator_SingleDayRange_YieldsOneDate", run: func(t *testing.T) {
		r, _ := NewRange(nd(2080, 5, 15), nd(2080, 5, 15))
		check(t, slices.Equal(slices.Collect(r.Days()), []Date{nd(2080, 5, 15)}), "one day")
	}},
	{id: "NepaliDateRangeExtendedTests.ToString_DefaultFormat_IsStartDashEnd", skip: d09String},
	{id: "NepaliDateRangeExtendedTests.ToString_EmptyRange_ReturnsEmptyRangeString", skip: d09String},
	{id: "NepaliDateRangeExtendedTests.ToString_WithDateFormat_UsesSpecifiedFormat", skip: d09String},
	{id: "NepaliDateRangeExtendedTests.ToString_WithDateFormat_EmptyRange_ReturnsEmptyRangeString", skip: d09String},
	{id: "NepaliDateRangeExtendedTests.Equals_IdenticalRanges_ReturnsTrue", run: rangesEqual(nd(2080, 1, 1), nd(2080, 1, 15), nd(2080, 1, 1), nd(2080, 1, 15), true)},
	{id: "NepaliDateRangeExtendedTests.Equals_DifferentStart_ReturnsFalse", run: rangesEqual(nd(2080, 1, 1), nd(2080, 1, 15), nd(2080, 1, 2), nd(2080, 1, 15), false)},
	{id: "NepaliDateRangeExtendedTests.Equals_DifferentEnd_ReturnsFalse", run: rangesEqual(nd(2080, 1, 1), nd(2080, 1, 15), nd(2080, 1, 1), nd(2080, 1, 16), false)},
	{id: "NepaliDateRangeExtendedTests.Equals_BothEmpty_ReturnsTrue", run: rangesEqual(nd(2080, 5, 10), nd(2080, 5, 1), nd(2079, 3, 10), nd(2079, 3, 1), true)},
	{id: "NepaliDateRangeExtendedTests.Equals_ObjectOverload_SameRange_ReturnsTrue", run: rangesEqual(nd(2080, 1, 1), nd(2080, 1, 15), nd(2080, 1, 1), nd(2080, 1, 15), true)},
	{id: "NepaliDateRangeExtendedTests.Equals_ObjectOverload_WrongType_ReturnsFalse", skip: "n/a: == on a Range only accepts a Range, so the Go compiler rejects other types"},
	// GetHashCode becomes "equal ranges are equal map keys".
	{id: "NepaliDateRangeExtendedTests.GetHashCode_EqualRanges_SameHash", run: func(t *testing.T) {
		a, _ := NewRange(nd(2080, 1, 1), nd(2080, 1, 15))
		b, _ := NewRange(nd(2080, 1, 1), nd(2080, 1, 15))
		m := map[Range]int{a: 1}
		check(t, m[b] == 1, "equal ranges are the same map key")
	}},
	{id: "NepaliDateRangeExtendedTests.OperatorEquals_IdenticalRanges_ReturnsTrue", run: rangesEqual(nd(2080, 1, 1), nd(2080, 1, 15), nd(2080, 1, 1), nd(2080, 1, 15), true)},
	{id: "NepaliDateRangeExtendedTests.OperatorNotEquals_DifferentRanges_ReturnsTrue", run: rangesEqual(nd(2080, 1, 1), nd(2080, 1, 15), nd(2080, 1, 1), nd(2080, 1, 16), false)},
}

// nd is MustNew for short catalogue rows.
func nd(y int, m Month, d int) Date { return MustNew(y, m, d) }

// lastOf is the last day of BS month y-m.
func lastOf(y int, m Month) Date { return MustNew(y, m, 1).LastOfMonth() }

// fixedToday stands in for the clock in the C# Current* rows (D-14).
func fixedToday() Date { return MustNew(2081, Shrawan, 15) }

// fyBounds checks the start and end of f; a zero want is not checked.
func fyBounds(f FiscalYear, start, end Date) func(*testing.T) {
	return func(t *testing.T) {
		if !start.IsZero() {
			got, err := f.Start()
			check(t, err == nil && got == start, "fiscal year start")
		}
		if !end.IsZero() {
			got, err := f.End()
			check(t, err == nil && got == end, "fiscal year end")
		}
	}
}

func fyOutOfRange(f FiscalYear) func(*testing.T) {
	return func(t *testing.T) {
		_, err1 := f.Start()
		_, err2 := f.End()
		check(t, errors.Is(err1, ErrOutOfRange) && errors.Is(err2, ErrOutOfRange), "start and end are out of range")
	}
}

// quarterBounds checks the start and end of quarter q of f; a zero want is not checked.
func quarterBounds(f FiscalYear, q Quarter, start, end Date) func(*testing.T) {
	return func(t *testing.T) {
		r, err := f.QuarterRange(q)
		check(t, err == nil, "QuarterRange succeeds")
		check(t, start.IsZero() || r.Start() == start, "quarter start")
		check(t, end.IsZero() || r.End() == end, "quarter end")
	}
}

func addDays(d Date, n int, want Date) func(*testing.T) {
	return func(t *testing.T) {
		got, err := d.AddDays(n)
		if err != nil || got != want {
			t.Errorf("%v.AddDays(%d) = %v, %v; want %v", d, n, got, err, want)
		}
	}
}

func addMonths(d Date, n int, o Overflow, want Date) func(*testing.T) {
	return func(t *testing.T) {
		got, err := d.AddMonths(n, o)
		if err != nil || got != want {
			t.Errorf("%v.AddMonths(%d, %d) = %v, %v; want %v", d, n, o, got, err, want)
		}
	}
}

func daysUntil(a, b Date, want int) func(*testing.T) {
	return func(t *testing.T) {
		if got := a.DaysUntil(b); got != want {
			t.Errorf("%v.DaysUntil(%v) = %d, want %d", a, b, got, want)
		}
	}
}

func contains(start, end, d Date, want bool) func(*testing.T) {
	return func(t *testing.T) {
		r, ok := NewRange(start, end)
		check(t, ok && r.Contains(d) == want, "Contains")
	}
}

// rangesEqual builds two ranges (a failed NewRange gives the empty zero Range, as C# gives an
// empty range) and compares them with ==.
func rangesEqual(s1, e1, s2, e2 Date, want bool) func(*testing.T) {
	return func(t *testing.T) {
		a, _ := NewRange(s1, e1)
		b, _ := NewRange(s2, e2)
		check(t, (a == b) == want, "range equality")
	}
}

func fromTimeIs(tm time.Time, want Date) func(*testing.T) {
	return func(t *testing.T) {
		got, err := FromTime(tm)
		if err != nil || got != want {
			t.Errorf("FromTime(%v) = %v, %v; want %v", tm, got, err, want)
		}
	}
}
