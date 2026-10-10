package nepdate

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"
)

// TestCSharp ports C# NepDate's own tests (shared/spec/vectors/csharp-tests/catalog.tsv, hub
// H5). One subtest per catalogue row, named after the row id with "." and "#" replaced by "_"
// (ADR-0004). A row that is not ported names its PARITY deviation or says n/a.
//
// A row marked "pending" waits for the task that adds its API; the marker comes out in that
// task (hub PORT-PLAN, "C# test catalogue").

type csharpRow struct {
	id   string
	skip string
	run  func(t *testing.T)
}

// csharpPhases lists the catalogue phases this package covers so far.
var csharpPhases = []string{"1", "2", "3", "4", "D-14"}

// patroRowPrefix marks the catalogue rows that package patro ports in its own TestCSharp.
const patroRowPrefix = "CalendarDataTests."

var csharpRows = []csharpRow{
	// Abilities/NepaliDateComparableTests: _d1 = 2081-01-01, _d2 = 2081-06-15, _d3 = 2082-01-01.
	{id: "NepaliDateComparableTests.CompareTo_Object_Earlier_ReturnsNegative", run: func(t *testing.T) {
		check(t, cmpD1().Compare(cmpD2()) < 0, "d1 < d2")
	}},
	{id: "NepaliDateComparableTests.CompareTo_Object_Same_ReturnsZero", run: func(t *testing.T) {
		check(t, cmpD1().Compare(MustNew(2081, 1, 1)) == 0, "d1 == 2081-01-01")
	}},
	{id: "NepaliDateComparableTests.CompareTo_Object_Later_ReturnsPositive", run: func(t *testing.T) {
		check(t, cmpD3().Compare(cmpD1()) > 0, "d3 > d1")
	}},
	// C# null maps to the zero Date in Go (D-01), which sorts before every valid date.
	{id: "NepaliDateComparableTests.CompareTo_Object_Null_ReturnsPositive", run: func(t *testing.T) {
		check(t, cmpD1().Compare(Date{}) > 0, "d1 > zero Date")
	}},
	{id: "NepaliDateComparableTests.CompareTo_WrongType_ThrowsArgumentException", skip: "n/a: Compare takes a Date, so the Go compiler rejects other types"},
	{id: "NepaliDateComparableTests.CompareTo_Generic_Earlier_ReturnsNegative", run: func(t *testing.T) {
		check(t, cmpD1().Compare(cmpD2()) < 0, "d1 < d2")
	}},
	{id: "NepaliDateComparableTests.CompareTo_Generic_Same_ReturnsZero", run: func(t *testing.T) {
		check(t, cmpD1().Compare(MustNew(2081, 1, 1)) == 0, "d1 == 2081-01-01")
	}},
	{id: "NepaliDateComparableTests.CompareTo_Generic_Later_ReturnsPositive", run: func(t *testing.T) {
		check(t, cmpD3().Compare(cmpD1()) > 0, "d3 > d1")
	}},
	{id: "NepaliDateComparableTests.SortedSet_OrdersCorrectly", run: sortsCorrectly},
	{id: "NepaliDateComparableTests.ArraySort_ViaIComparable_OrdersCorrectly", run: sortsCorrectly},
	{id: "NepaliDateComparableTests.List_Sort_OrdersCorrectly", run: sortsCorrectly},

	// Core/DictionaryIntegrityTests
	{id: "DictionaryIntegrityTests.AllNepaliMonths_RoundtripToEnglishAndBack_ProducesOriginalDate", run: func(t *testing.T) {
		for y := 1901; y <= 2199; y++ {
			for m := Baishakh; m <= Chaitra; m++ {
				d := MustNew(y, m, 1)
				if back, err := FromGregorian(d.Gregorian()); err != nil || back != d {
					t.Fatalf("%v round-trips to %v, %v", d, back, err)
				}
			}
		}
	}},
	{id: "DictionaryIntegrityTests.AllNepaliDays_RoundtripToEnglishAndBack_ProducesOriginalDate", run: func(t *testing.T) {
		total := 0
		for y := 1901; y <= 2199; y++ {
			for m := Baishakh; m <= Chaitra; m++ {
				for day := 1; day <= MustNew(y, m, 1).MonthLength(); day++ {
					d := MustNew(y, m, day)
					if back, err := FromGregorian(d.Gregorian()); err != nil || back != d {
						t.Fatalf("%v round-trips to %v, %v", d, back, err)
					}
					total++
				}
			}
		}
		check(t, total > 100000, "more than 100,000 days tested")
	}},
	{id: "DictionaryIntegrityTests.AllEnglishDays_RoundtripToNepaliAndBack_ProducesOriginalDate", run: func(t *testing.T) {
		day, end := adTime(Min), adTime(Max)
		total := 0
		for ; !day.After(end); day = day.AddDate(0, 0, 1) {
			bs, err := FromGregorian(day.Date())
			if err != nil {
				t.Fatalf("FromGregorian(%v): %v", day, err)
			}
			if !adTime(bs).Equal(day) {
				t.Fatalf("%v round-trips to %v", day, adTime(bs))
			}
			total++
		}
		check(t, total > 100000, "more than 100,000 days tested")
	}},
	{id: "DictionaryIntegrityTests.AllMonths_MonthEndDay_IsWithinValidRange", run: func(t *testing.T) {
		for y := 1901; y <= 2199; y++ {
			for m := Baishakh; m <= Chaitra; m++ {
				if n := MustNew(y, m, 1).MonthLength(); n < 29 || n > 32 {
					t.Fatalf("%d-%d has %d days", y, m, n)
				}
			}
		}
	}},
	{id: "DictionaryIntegrityTests.ConsecutiveNepaliDays_ProduceConsecutiveEnglishDays", run: func(t *testing.T) {
		var prev time.Time
		for y := 1901; y <= 2199; y++ {
			for m := Baishakh; m <= Chaitra; m++ {
				for day := 1; day <= MustNew(y, m, 1).MonthLength(); day++ {
					cur := adTime(MustNew(y, m, day))
					if !prev.IsZero() && !cur.Equal(prev.AddDate(0, 0, 1)) {
						t.Fatalf("%d-%d-%d is %v, the day before was %v", y, m, day, cur, prev)
					}
					prev = cur
				}
			}
		}
	}},
	{id: "DictionaryIntegrityTests.KnownReferenceDates_ConvertCorrectly#1", run: knownReference(MustNew(2080, 1, 1), time.Date(2023, 4, 14, 0, 0, 0, 0, time.UTC))},
	{id: "DictionaryIntegrityTests.KnownReferenceDates_ConvertCorrectly#2", run: knownReference(MustNew(2000, 1, 1), time.Date(1943, 4, 14, 0, 0, 0, 0, time.UTC))},
	{id: "DictionaryIntegrityTests.KnownReferenceDates_ConvertCorrectly#3", run: knownReference(MustNew(2081, 1, 1), time.Date(2024, 4, 13, 0, 0, 0, 0, time.UTC))},
	{id: "DictionaryIntegrityTests.KnownReferenceDates_ConvertCorrectly#4", run: knownReference(MustNew(1901, 1, 1), time.Date(1844, 4, 11, 0, 0, 0, 0, time.UTC))},
	{id: "DictionaryIntegrityTests.BoundaryDates_ConvertCorrectly", run: func(t *testing.T) {
		check(t, Min.Year() == 1901 && Min.Month() == 1 && Min.Day() == 1, "Min is 1901-01-01")
		check(t, Max.Year() == 2199 && Max.Month() == 12, "Max is in 2199-12")
		for _, d := range []Date{Min, Max} {
			back, err := FromGregorian(d.Gregorian())
			check(t, err == nil && back == d, "boundary round-trips")
		}
	}},
	{id: "DictionaryIntegrityTests.TotalDaysInRange_MatchesEnglishDateSpan", run: func(t *testing.T) {
		total := 0
		for y := 1901; y <= 2199; y++ {
			for m := Baishakh; m <= Chaitra; m++ {
				total += MustNew(y, m, 1).MonthLength()
			}
		}
		// time.Duration tops out at 292 years, so count days through Unix seconds.
		span := int((adTime(Max).Unix()-adTime(Min).Unix())/86400) + 1
		check(t, total == span, "sum of month lengths equals the AD span")
	}},

	// Core/NepaliDateConstructionTests
	{id: "NepaliDateConstructionTests.Constructor_ValidNepaliDate_CreatesInstance", run: func(t *testing.T) {
		d, err := New(2080, 5, 15)
		check(t, err == nil && d.Year() == 2080 && d.Month() == 5 && d.Day() == 15, "New(2080, 5, 15)")
	}},
	{id: "NepaliDateConstructionTests.Constructor_MinValue_CreatesInstance", run: func(t *testing.T) {
		check(t, Min.Year() == 1901 && Min.Month() == 1 && Min.Day() == 1, "Min is 1901-01-01")
	}},
	{id: "NepaliDateConstructionTests.Constructor_MaxValue_CreatesInstance", run: func(t *testing.T) {
		check(t, Max.Year() == 2199 && Max.Month() == 12 && Max.Day() > 0, "Max is in 2199-12")
	}},
	{id: "NepaliDateConstructionTests.Constructor_InvalidNepaliDate_ThrowsException#1", run: newFails(2080, 0, 1, ErrInvalidMonth)},
	{id: "NepaliDateConstructionTests.Constructor_InvalidNepaliDate_ThrowsException#2", run: newFails(2080, 13, 1, ErrInvalidMonth)},
	{id: "NepaliDateConstructionTests.Constructor_InvalidNepaliDate_ThrowsException#3", run: newFails(2080, 1, 0, ErrInvalidDay)},
	{id: "NepaliDateConstructionTests.Constructor_InvalidNepaliDate_ThrowsException#4", run: newFails(2080, 1, 33, ErrInvalidDay)},
	{id: "NepaliDateConstructionTests.Constructor_ValidStringFormats_CreatesInstance#1", run: parsesTo("2080/05/15", nd(2080, 5, 15))},
	{id: "NepaliDateConstructionTests.Constructor_ValidStringFormats_CreatesInstance#2", run: parsesTo("2080-05-15", nd(2080, 5, 15))},
	{id: "NepaliDateConstructionTests.Constructor_ValidStringFormats_CreatesInstance#3", run: parsesTo("2080.05.15", nd(2080, 5, 15))},
	{id: "NepaliDateConstructionTests.Constructor_ValidStringFormats_CreatesInstance#4", run: parsesTo("2080_05_15", nd(2080, 5, 15))},
	{id: "NepaliDateConstructionTests.Constructor_ValidStringFormats_CreatesInstance#5", run: parsesTo(`2080\05\15`, nd(2080, 5, 15))},
	{id: "NepaliDateConstructionTests.Constructor_ValidStringFormats_CreatesInstance#6", run: parsesTo("2080 05 15", nd(2080, 5, 15))},
	{id: "NepaliDateConstructionTests.Constructor_AutoAdjustedFormats_CreatesInstance#1", skip: "D-06: autoAdjust guesses the field order"},
	{id: "NepaliDateConstructionTests.Constructor_AutoAdjustedFormats_CreatesInstance#2", skip: "D-06: autoAdjust guesses the field order"},
	{id: "NepaliDateConstructionTests.Constructor_AutoAdjustedFormats_CreatesInstance#3", run: parsesTo("2080/05/15", nd(2080, 5, 15))},
	{id: "NepaliDateConstructionTests.Constructor_AutoAdjustedFormats_CreatesInstance#4", skip: "D-06: 2-digit years are not expanded"},
	{id: "NepaliDateConstructionTests.Constructor_AutoAdjust_MonthOverflowWithBoundaryDay_SwapsAsDocumented#1", skip: "D-06: autoAdjust swaps month and day"},
	{id: "NepaliDateConstructionTests.Constructor_AutoAdjust_MonthOverflowWithBoundaryDay_SwapsAsDocumented#2", skip: "D-06: autoAdjust swaps month and day"},
	{id: "NepaliDateConstructionTests.TryParse_AutoAdjust_MonthOverflowWithBoundaryDay_SwapsAsDocumented", skip: "D-06: autoAdjust swaps month and day"},
	{id: "NepaliDateConstructionTests.Constructor_InvalidStringFormats_ThrowsException#1", run: parseFailsSyntax("")},
	{id: "NepaliDateConstructionTests.Constructor_InvalidStringFormats_ThrowsException#2", run: parseFailsSyntax("invalid")},
	{id: "NepaliDateConstructionTests.Constructor_InvalidStringFormats_ThrowsException#3", run: parseFailsSyntax("2080/5")},
	{id: "NepaliDateConstructionTests.Constructor_ValidFormatInvalidDate_ThrowsException#1", run: parseFailsWith("2080/5/40", ErrInvalidDay)},
	{id: "NepaliDateConstructionTests.Constructor_EnglishDate_ConvertsCorrectly", run: func(t *testing.T) {
		d, err := FromGregorian(2023, time.August, 30)
		check(t, err == nil && d == MustNew(2080, 5, 13), "AD 2023-08-30 is BS 2080-05-13")
	}},
	{id: "NepaliDateConstructionTests.TryParse_ValidString_ReturnsTrueWithCorrectDate", run: parsesTo("2080/05/15", nd(2080, 5, 15))},
	{id: "NepaliDateConstructionTests.TryParse_InvalidString_ReturnsFalseWithDefault#1", run: parseFailsSyntax("")},
	{id: "NepaliDateConstructionTests.TryParse_InvalidString_ReturnsFalseWithDefault#2", run: parseFailsSyntax("invalid")},
	{id: "NepaliDateConstructionTests.TryParse_InvalidString_ReturnsFalseWithDefault#3", run: parseFailsSyntax("2080/5")},

	// Core/NepaliDatePropertiesTests: the date is 2080-05-15 unless the row says otherwise.
	{id: "NepaliDatePropertiesTests.EnglishDate_Conversion_ReturnsCorrectDate", run: func(t *testing.T) {
		y, m, d := MustNew(2080, 5, 15).Gregorian()
		check(t, y == 2023 && m == time.September && d == 1, "2080-05-15 is AD 2023-09-01")
	}},
	{id: "NepaliDatePropertiesTests.DayOfWeek_ReturnsCorrectValue", run: func(t *testing.T) {
		check(t, MustNew(2080, 5, 15).Weekday() == time.Friday, "2080-05-15 is a Friday")
	}},
	{id: "NepaliDatePropertiesTests.MonthEndDay_ReturnsCorrectValue", run: func(t *testing.T) {
		check(t, MustNew(2080, 5, 15).MonthLength() == 31, "Bhadra 2080 has 31 days")
	}},
	{id: "NepaliDatePropertiesTests.MonthName_ReturnsCorrectEnum", run: func(t *testing.T) {
		check(t, MustNew(2080, 5, 15).Month() == Bhadra, "month 5 is Bhadra")
	}},
	{id: "NepaliDatePropertiesTests.Today_ReturnsCurrentNepaliDate", run: func(t *testing.T) {
		// C# compares NepaliDate.Today with new NepaliDate(DateTime.Today); the clock may
		// cross midnight between the two reads, so either side matches.
		before, _ := FromTime(time.Now().In(kathmandu()))
		today, err := Today(nil)
		after, _ := FromTime(time.Now().In(kathmandu()))
		check(t, err == nil && (today == before || today == after), "Today(nil) is today in Kathmandu")
	}},
	{id: "NepaliDatePropertiesTests.DayOfYear_FirstDayOfYear_Returns1", run: func(t *testing.T) {
		check(t, MustNew(2080, 1, 1).YearDay() == 1, "Baishakh 1 is day 1")
	}},
	{id: "NepaliDatePropertiesTests.DayOfYear_FirstDayOfMonth4_EqualsSum_Of_Months1To3_Plus1", run: func(t *testing.T) {
		want := MustNew(2080, 1, 1).MonthLength() + MustNew(2080, 2, 1).MonthLength() + MustNew(2080, 3, 1).MonthLength() + 1
		check(t, MustNew(2080, 4, 1).YearDay() == want, "Shrawan 1 follows the first three months")
	}},
	// C# null maps to the zero Date in Go (D-01).
	{id: "NepaliDatePropertiesTests.Equals_NullObject_ReturnsFalse", run: func(t *testing.T) {
		check(t, MustNew(2080, 5, 15) != Date{}, "a date is not the zero Date")
	}},
	{id: "NepaliDatePropertiesTests.Equals_DifferentType_ReturnsFalse", skip: "n/a: == on a Date only accepts a Date, so the Go compiler rejects other types"},

	// Core/OptimizationVerificationTests, phase 1 rows
	{id: "OptimizationVerificationTests.Constructor_AllSeparators_ParseCorrectly#1", run: parsesTo("2080/05/15", nd(2080, 5, 15))},
	{id: "OptimizationVerificationTests.Constructor_AllSeparators_ParseCorrectly#2", run: parsesTo("2080-05-15", nd(2080, 5, 15))},
	{id: "OptimizationVerificationTests.Constructor_AllSeparators_ParseCorrectly#3", run: parsesTo("2080.05.15", nd(2080, 5, 15))},
	{id: "OptimizationVerificationTests.Constructor_AllSeparators_ParseCorrectly#4", run: parsesTo("2080_05_15", nd(2080, 5, 15))},
	{id: "OptimizationVerificationTests.Constructor_AllSeparators_ParseCorrectly#5", run: parsesTo(`2080\05\15`, nd(2080, 5, 15))},
	{id: "OptimizationVerificationTests.Constructor_AllSeparators_ParseCorrectly#6", run: parsesTo("2080 05 15", nd(2080, 5, 15))},
	{id: "OptimizationVerificationTests.Constructor_AllSeparators_ParseCorrectly#7", run: parsesTo("2080।05।15", nd(2080, 5, 15))},
	{id: "OptimizationVerificationTests.Constructor_AllSeparators_ParseCorrectly#8", run: parsesTo("2080|05|15", nd(2080, 5, 15))},
	{id: "OptimizationVerificationTests.Constructor_InvalidStrings_ThrowsException#1", run: parseFailsSyntax("")},
	{id: "OptimizationVerificationTests.Constructor_InvalidStrings_ThrowsException#2", run: parseFailsSyntax("invalid")},
	{id: "OptimizationVerificationTests.Constructor_InvalidStrings_ThrowsException#3", run: parseFailsSyntax("2080/05")},
	{id: "OptimizationVerificationTests.Constructor_InvalidStrings_ThrowsException#4", run: parseFailsSyntax("2080")},
	{id: "OptimizationVerificationTests.Constructor_InvalidStrings_ThrowsException#5", run: parseFailsSyntax("2080/05/15/99")},
	{id: "OptimizationVerificationTests.Constructor_InvalidStrings_ThrowsException#6", run: parseFailsSyntax("abc/def/ghi")},
	{id: "OptimizationVerificationTests.Constructor_NullString_ThrowsException", skip: "n/a: a Go string is never nil"},
	{id: "OptimizationVerificationTests.Constructor_WithoutLeadingZeros_ParsesCorrectly#1", run: parsesTo("2080/5/15", nd(2080, 5, 15))},
	{id: "OptimizationVerificationTests.Constructor_WithoutLeadingZeros_ParsesCorrectly#2", run: parsesTo("2080/05/5", nd(2080, 5, 5))},
	{id: "OptimizationVerificationTests.Constructor_WithoutLeadingZeros_ParsesCorrectly#3", run: parsesTo("2080/1/1", nd(2080, 1, 1))},
	{id: "OptimizationVerificationTests.Operator_Equality_SameDates_ReturnsTrue", run: func(t *testing.T) {
		a, b := MustNew(2080, 5, 15), MustNew(2080, 5, 15)
		check(t, a == b, "a == b")
	}},
	{id: "OptimizationVerificationTests.Operator_Inequality_DifferentDates_ReturnsTrue", run: func(t *testing.T) {
		a, b := MustNew(2080, 5, 15), MustNew(2080, 5, 16)
		check(t, a != b, "a != b")
	}},
	{id: "OptimizationVerificationTests.Operator_LessThan_EarlierDate_ReturnsTrue", run: func(t *testing.T) {
		e, l := MustNew(2080, 5, 15), MustNew(2080, 5, 16)
		check(t, e.Before(l) && !l.Before(e) && !e.Before(e), "Before")
	}},
	{id: "OptimizationVerificationTests.Operator_LessThanOrEqual_SameAndEarlier_ReturnsTrue", run: func(t *testing.T) {
		e, l := MustNew(2080, 5, 15), MustNew(2080, 5, 16)
		check(t, !e.After(l) && !e.After(e) && l.After(e), "not After")
	}},
	{id: "OptimizationVerificationTests.Operator_GreaterThan_LaterDate_ReturnsTrue", run: func(t *testing.T) {
		e, l := MustNew(2080, 5, 15), MustNew(2080, 5, 16)
		check(t, l.After(e) && !e.After(l) && !l.After(l), "After")
	}},
	{id: "OptimizationVerificationTests.Operator_GreaterThanOrEqual_SameAndLater_ReturnsTrue", run: func(t *testing.T) {
		e, l := MustNew(2080, 5, 15), MustNew(2080, 5, 16)
		check(t, !l.Before(e) && !l.Before(l) && e.Before(l), "not Before")
	}},
	{id: "OptimizationVerificationTests.Operator_MinMaxBoundary_CorrectOrdering", run: func(t *testing.T) {
		check(t, Min.Before(Max) && Max.After(Min) && Min.Compare(Max) <= 0 && Max.Compare(Min) >= 0, "Min < Max")
	}},
	{id: "OptimizationVerificationTests.CompareTo_EarlierDate_ReturnsNegative", run: func(t *testing.T) {
		check(t, MustNew(2080, 5, 15).Compare(MustNew(2080, 5, 16)) < 0, "earlier < later")
	}},
	{id: "OptimizationVerificationTests.CompareTo_SameDate_ReturnsZero", run: func(t *testing.T) {
		d := MustNew(2080, 5, 15)
		same := d
		check(t, d.Compare(same) == 0, "d == d")
	}},
	{id: "OptimizationVerificationTests.CompareTo_LaterDate_ReturnsPositive", run: func(t *testing.T) {
		check(t, MustNew(2080, 5, 16).Compare(MustNew(2080, 5, 15)) > 0, "later > earlier")
	}},
	{id: "OptimizationVerificationTests.CompareTo_DifferentYears_CorrectOrdering", run: func(t *testing.T) {
		check(t, MustNew(2080, 1, 1).Compare(MustNew(2081, 1, 1)) < 0, "2080 < 2081")
	}},
	{id: "OptimizationVerificationTests.CompareTo_DifferentMonths_CorrectOrdering", run: func(t *testing.T) {
		check(t, MustNew(2080, 1, 1).Compare(MustNew(2080, 2, 1)) < 0, "month 1 < month 2")
	}},
	{id: "OptimizationVerificationTests.Equals_SameDates_ReturnsTrue", run: func(t *testing.T) {
		a, b := MustNew(2080, 5, 15), MustNew(2080, 5, 15)
		check(t, a == b, "equal")
	}},
	{id: "OptimizationVerificationTests.Equals_DifferentDates_ReturnsFalse", run: func(t *testing.T) {
		check(t, MustNew(2080, 5, 15) != MustNew(2080, 5, 16), "not equal")
	}},
	// GetHashCode rows become "equal dates are equal map keys" (G1-07).
	{id: "OptimizationVerificationTests.GetHashCode_EqualDates_SameHashCode", run: func(t *testing.T) {
		m := map[Date]int{MustNew(2080, 5, 15): 7}
		check(t, m[MustNew(2080, 5, 15)] == 7, "equal dates are the same map key")
	}},
	{id: "OptimizationVerificationTests.GetHashCode_StableAcrossMultipleCalls", run: func(t *testing.T) {
		d := MustNew(2080, 5, 15)
		m := map[Date]int{d: 1}
		m[d]++
		check(t, len(m) == 1 && m[d] == 2, "a date is one stable map key")
	}},
	{id: "OptimizationVerificationTests.IsLeapYear_ChecksEnglishYearCorrectly#1", skip: "D-12: IsLeapYear is not ported"},
	{id: "OptimizationVerificationTests.IsLeapYear_ChecksEnglishYearCorrectly#2", skip: "D-12: IsLeapYear is not ported"},
	{id: "OptimizationVerificationTests.IsLeapYear_ChecksEnglishYearCorrectly#3", skip: "D-12: IsLeapYear is not ported"},
	{id: "OptimizationVerificationTests.Constructor_InvalidYearMonthDay_ThrowsException#1", run: newFails(2080, 0, 15, ErrInvalidMonth)},
	{id: "OptimizationVerificationTests.Constructor_InvalidYearMonthDay_ThrowsException#2", run: newFails(2080, 13, 15, ErrInvalidMonth)},
	{id: "OptimizationVerificationTests.Constructor_InvalidYearMonthDay_ThrowsException#3", run: newFails(2080, 5, 0, ErrInvalidDay)},
	{id: "OptimizationVerificationTests.Constructor_InvalidYearMonthDay_ThrowsException#4", run: newFails(2080, 5, 33, ErrInvalidDay)},
	{id: "OptimizationVerificationTests.Constructor_InvalidYearMonthDay_ThrowsException#5", run: newFails(1900, 1, 1, ErrOutOfRange)},
	{id: "OptimizationVerificationTests.Constructor_InvalidYearMonthDay_ThrowsException#6", run: newFails(2200, 1, 1, ErrOutOfRange)},
	{id: "OptimizationVerificationTests.Constructor_MinValue_IsValid", run: func(t *testing.T) {
		check(t, Min.Year() == 1901 && Min.Month() == 1 && Min.Day() == 1, "Min is 1901-01-01")
	}},
	{id: "OptimizationVerificationTests.Constructor_MaxValue_IsValid", run: func(t *testing.T) {
		check(t, Max.Year() == 2199 && Max.Month() == 12 && Max.Day() >= 29 && Max.Day() <= 32, "Max is the last day of 2199-12")
	}},
	{id: "OptimizationVerificationTests.EnglishDate_MinValue_ConvertsAndRoundtrips", run: func(t *testing.T) {
		back, err := FromGregorian(Min.Gregorian())
		check(t, err == nil && back == Min, "Min round-trips")
	}},
	{id: "OptimizationVerificationTests.EnglishDate_MaxValue_ConvertsAndRoundtrips", run: func(t *testing.T) {
		back, err := FromGregorian(Max.Gregorian())
		check(t, err == nil && back == Max, "Max round-trips")
	}},
	{id: "OptimizationVerificationTests.EnglishDate_JustBeforeMin_ThrowsException", run: func(t *testing.T) {
		_, err := FromGregorian(1844, time.April, 10)
		check(t, errors.Is(err, ErrOutOfRange), "AD 1844-04-10 is out of range")
	}},
	{id: "OptimizationVerificationTests.EnglishDate_JustAfterMax_ThrowsException", run: func(t *testing.T) {
		_, err := FromGregorian(2143, time.April, 16)
		check(t, errors.Is(err, ErrOutOfRange), "AD 2143-04-16 is out of range")
	}},
	{id: "OptimizationVerificationTests.MonthEndDay_CreatingDateWithExactEndDay_Succeeds", run: func(t *testing.T) {
		for y := 2070; y <= 2090; y++ {
			for m := Baishakh; m <= Chaitra; m++ {
				n := MustNew(y, m, 1).MonthLength()
				last, err := New(y, m, n)
				check(t, err == nil && last.Day() == n, "the last day is valid")
				_, err = New(y, m, n+1)
				check(t, errors.Is(err, ErrInvalidDay), "the day after the last is invalid")
			}
		}
	}},
	{id: "OptimizationVerificationTests.DayOfWeek_MatchesEnglishDateDayOfWeek", run: func(t *testing.T) {
		for y := 1910; y <= 2190; y += 30 {
			for m := Baishakh; m <= Chaitra; m += 3 {
				d := MustNew(y, m, 15)
				check(t, d.Weekday() == adTime(d).Weekday(), "weekday matches the AD date")
			}
		}
	}},
	// C#'s IsDefault is IsZero in Go (D-01).
	{id: "OptimizationVerificationTests.IsDefault_DefaultStruct_ReturnsTrue", run: func(t *testing.T) {
		var d Date
		check(t, d.IsZero(), "zero Date IsZero")
	}},
	{id: "OptimizationVerificationTests.IsDefault_ValidDate_ReturnsFalse", run: func(t *testing.T) {
		check(t, !MustNew(2080, 5, 15).IsZero(), "a valid date is not zero")
	}},
}

// allCSharpRows is every ported catalogue row, phase by phase.
func allCSharpRows() []csharpRow {
	return slices.Concat(csharpRows, csharpRows2, csharpRows3, csharpRows4)
}

func TestCSharp(t *testing.T) {
	for _, row := range allCSharpRows() {
		t.Run(csharpTestName(row.id), func(t *testing.T) {
			if row.skip != "" {
				t.Skip(row.skip)
			}
			row.run(t)
		})
	}
}

// TestCSharpCatalogueCovered checks that csharpRows has every catalogue row of the phases in
// csharpPhases exactly once, and nothing else.
func TestCSharpCatalogueCovered(t *testing.T) {
	rows := readTable(t, "shared/spec/vectors/csharp-tests/catalog.tsv", "\t", 542)
	want := map[string]bool{}
	for _, r := range rows {
		if slices.Contains(csharpPhases, r["phase"]) && !strings.HasPrefix(r["id"], patroRowPrefix) {
			want[r["id"]] = true
		}
	}
	seen := map[string]bool{}
	for _, r := range allCSharpRows() {
		if seen[r.id] {
			t.Errorf("row %s appears twice", r.id)
		}
		seen[r.id] = true
		if !want[r.id] {
			t.Errorf("row %s is not in the catalogue for phases %v", r.id, csharpPhases)
		}
		if (r.skip == "") == (r.run == nil) {
			t.Errorf("row %s needs exactly one of skip and run", r.id)
		}
	}
	for id := range want {
		if !seen[id] {
			t.Errorf("catalogue row %s has no test", id)
		}
	}
}

func csharpTestName(id string) string {
	return strings.NewReplacer(".", "_", "#", "_").Replace(id)
}

func check(t *testing.T, ok bool, what string) {
	t.Helper()
	if !ok {
		t.Errorf("failed: %s", what)
	}
}

// adTime is d's AD date as midnight UTC, for comparisons with the time package.
func adTime(d Date) time.Time {
	y, m, day := d.Gregorian()
	return time.Date(y, m, day, 0, 0, 0, 0, time.UTC)
}

func cmpD1() Date { return MustNew(2081, 1, 1) }
func cmpD2() Date { return MustNew(2081, 6, 15) }
func cmpD3() Date { return MustNew(2082, 1, 1) }

func sortsCorrectly(t *testing.T) {
	dates := []Date{cmpD3(), cmpD1(), cmpD2()}
	slices.SortFunc(dates, Date.Compare)
	check(t, slices.Equal(dates, []Date{cmpD1(), cmpD2(), cmpD3()}), "sorted d1, d2, d3")
}

func knownReference(d Date, ad time.Time) func(*testing.T) {
	return func(t *testing.T) {
		check(t, adTime(d).Equal(ad), "BS to AD")
		back, err := FromGregorian(ad.Date())
		check(t, err == nil && back == d, "AD to BS")
	}
}

func newFails(y int, m Month, d int, want error) func(*testing.T) {
	return func(t *testing.T) {
		_, err := New(y, m, d)
		if !errors.Is(err, want) {
			t.Errorf("New(%d, %d, %d) error = %v, want %v", y, m, d, err, want)
		}
	}
}

func parsesTo(input string, want Date) func(*testing.T) {
	return func(t *testing.T) {
		got, err := Parse(input)
		if err != nil || got != want {
			t.Errorf("Parse(%q) = %v, %v; want %v", input, got, err, want)
		}
	}
}

// parseFailsSyntax expects a *ParseError, C#'s InvalidNepaliDateFormatException or a false
// TryParse, with the zero Date (C#'s default).
func parseFailsSyntax(input string) func(*testing.T) {
	return func(t *testing.T) {
		got, err := Parse(input)
		var pe *ParseError
		if !errors.As(err, &pe) || !got.IsZero() {
			t.Errorf("Parse(%q) = %v, %v; want a *ParseError", input, got, err)
		}
	}
}

func parseFailsWith(input string, want error) func(*testing.T) {
	return func(t *testing.T) {
		got, err := Parse(input)
		if !errors.Is(err, want) || !got.IsZero() {
			t.Errorf("Parse(%q) = %v, %v; want %v", input, got, err, want)
		}
	}
}
