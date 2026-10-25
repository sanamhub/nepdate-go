package patro

import (
	"bufio"
	"os"
	"slices"
	"strings"
	"testing"

	nepdate "github.com/sanamhub/nepdate-go"
)

// TestCSharp ports the CalendarDataTests rows of the C# catalogue
// (shared/spec/vectors/csharp-tests/catalog.tsv). One subtest per row, "." and "#" replaced by
// "_" (ADR-0004). Rows about C#'s internal arrays (CalendarOffsets, TithiData, EventData,
// HolidayData) have no counterpart in this layout: the generator tests in internal/gen check
// the same facts on the Go tables.

type csharpRow struct {
	id   string
	skip string
	run  func(t *testing.T)
}

const internalLayout = "n/a: internal layout; internal/gen checks the generated tables instead"

var csharpRows = []csharpRow{
	{id: "CalendarDataTests.CalendarStrings_Pool_IsNonEmpty", run: func(t *testing.T) {
		check(t, len(nameText) > 0 && len(nameOff) > 1, "the name table is not empty")
	}},
	{id: "CalendarDataTests.CalendarStrings_Pool_AllEntriesHaveNonEmptyNpAndEn", run: everyNameNonEmpty},
	{id: "CalendarDataTests.CalendarOffsets_MonthStart_HasExpectedLength", skip: internalLayout},
	{id: "CalendarDataTests.CalendarOffsets_MonthStart_IsStrictlyMonotonic", skip: internalLayout},
	{id: "CalendarDataTests.CalendarOffsets_TotalDays_MatchesExpectedCount", skip: internalLayout},
	{id: "CalendarDataTests.TithiData_DayOffsets_IsStrictlySorted", skip: internalLayout},
	{id: "CalendarDataTests.TithiData_AllPoolIds_AreInRange", skip: internalLayout},
	{id: "CalendarDataTests.TithiData_EntryCount_Matches2190", run: func(t *testing.T) {
		n := 0
		for d := range Coverage().Days() {
			if _, ok := Of(d).Tithi(); ok {
				n++
			}
		}
		check(t, n == 2190, "2190 days have a tithi")
	}},
	{id: "CalendarDataTests.EventData_DayOffsets_IsStrictlySorted", skip: internalLayout},
	{id: "CalendarDataTests.EventData_SliceStart_IsNonDecreasing", skip: internalLayout},
	{id: "CalendarDataTests.EventData_SliceStart_SentinelEqualsEntriesLength", skip: internalLayout},
	{id: "CalendarDataTests.EventData_AllEntries_AreInRange", skip: internalLayout},
	{id: "CalendarDataTests.EventData_DayCount_Matches1671", run: func(t *testing.T) {
		n := 0
		for d := range Coverage().Days() {
			if Of(d).EventCount() > 0 {
				n++
			}
		}
		check(t, n == 1671, "1671 days have events")
	}},
	{id: "CalendarDataTests.HolidayData_DayOffsets_IsStrictlySorted", run: func(t *testing.T) {
		check(t, slices.IsSortedFunc(slices.Collect(PublicHolidays()), nepdate.Date.Compare), "holidays ascend")
	}},
	{id: "CalendarDataTests.HolidayData_Count_Matches294", run: func(t *testing.T) {
		check(t, len(slices.Collect(PublicHolidays())) == 294, "294 holidays")
	}},
	{id: "CalendarDataTests.CalendarOffsets_GetOffset_BaseYear_Month1_Day1_ReturnsZero", skip: internalLayout},
	{id: "CalendarDataTests.CalendarOffsets_GetOffset_OutOfRange_ReturnsNegativeOne", skip: internalLayout},
	{id: "CalendarDataTests.CalendarOffsets_GetOffset_AllOffsetsAreUniqueSentinel", skip: internalLayout},
	{id: "CalendarDataTests.NepaliDate_2081_01_01_CalendarData", run: func(t *testing.T) {
		i := Of(nepdate.MustNew(2081, 1, 1))
		tithi, _ := i.Tithi()
		check(t, tithi.Name(nepdate.Nepali) == "पञ्चमी" && tithi.Name(nepdate.English) == "Panchami", "tithi Panchami")
		check(t, i.IsPublicHoliday(), "holiday")
		check(t, slices.Contains(eventNames(i, nepdate.Nepali), "नयाँ वर्ष") && slices.Contains(eventNames(i, nepdate.English), "New Year"), "New Year")
	}},
	{id: "CalendarDataTests.NepaliDate_2081_01_02_Tithi_IsShashthi", run: func(t *testing.T) {
		tithi, _ := Of(nepdate.MustNew(2081, 1, 2)).Tithi()
		check(t, tithi.Name(nepdate.Nepali) == "षष्ठी" && tithi.Name(nepdate.English) == "Shashthi", "tithi Shashthi")
	}},
	{id: "CalendarDataTests.NepaliDate_2081_01_05_IsPublicHoliday", run: func(t *testing.T) {
		i := Of(nepdate.MustNew(2081, 1, 5))
		tithi, _ := i.Tithi()
		check(t, i.IsPublicHoliday() && tithi.Name(nepdate.Nepali) == "नवमी" && tithi.Name(nepdate.English) == "Navami", "Ram Navami holiday")
	}},
	{id: "CalendarDataTests.NepaliDate_2081_02_15_IsPublicHoliday_GanatantraDiwas", run: func(t *testing.T) {
		i := Of(nepdate.MustNew(2081, 2, 15))
		check(t, i.IsPublicHoliday() && slices.Contains(eventNames(i, nepdate.Nepali), "गणतन्त्र दिवस") && slices.Contains(eventNames(i, nepdate.English), "Republic Day"), "Republic Day holiday")
	}},
	{id: "CalendarDataTests.NepaliDate_2080_11_01_CalendarData", run: func(t *testing.T) {
		i := Of(nepdate.MustNew(2080, 11, 1))
		tithi, _ := i.Tithi()
		check(t, tithi.Name(nepdate.Nepali) == "चतुर्थी" && tithi.Name(nepdate.English) == "Chaturthi" && !i.IsPublicHoliday(), "Chaturthi, no holiday")
		check(t, slices.Contains(eventNames(i, nepdate.Nepali), "जनयुद्ध दिवस") && slices.Contains(eventNames(i, nepdate.English), "People's War Day"), "People's War Day")
	}},
	{id: "CalendarDataTests.NepaliDate_BeforeRange_AllPropertiesReturnDefaults", run: emptyOn(nepdate.MustNew(2000, 6, 15))},
	{id: "CalendarDataTests.NepaliDate_AfterRange_AllPropertiesReturnDefaults", run: emptyOn(nepdate.MustNew(2090, 1, 1))},
	{id: "CalendarDataTests.NepaliDate_OutsideTithiRange_TithiIsEmpty", run: func(t *testing.T) {
		tithi, ok := Of(nepdate.MustNew(2070, 6, 15)).Tithi()
		check(t, !ok && tithi.Name(nepdate.English) == "" && tithi.Name(nepdate.Nepali) == "", "no tithi")
	}},
	{id: "CalendarDataTests.NepaliDate_EventsNp_And_EventsEn_AlwaysHaveSameLength", run: func(t *testing.T) {
		for _, d := range []nepdate.Date{nepdate.MustNew(2081, 1, 1), nepdate.MustNew(2081, 1, 5), nepdate.MustNew(2081, 2, 10), nepdate.MustNew(2080, 11, 1), nepdate.MustNew(2075, 1, 1)} {
			i := Of(d)
			check(t, len(eventNames(i, nepdate.Nepali)) == len(eventNames(i, nepdate.English)), "same number of names in both languages")
		}
	}},
	// C#'s GetCalendarInfo and its flat properties are one Info here, so this checks that Of
	// gives the same Info on repeated calls.
	{id: "CalendarDataTests.GetCalendarInfo_TithiMatchesFlatProperties", run: func(t *testing.T) {
		first, second := Of(nepdate.MustNew(2081, 1, 1)), Of(nepdate.MustNew(2081, 1, 1))
		check(t, first == second, "Of is stable")
	}},
	{id: "CalendarDataTests.AllDaysInScrapedRange_OffsetsAreInBounds", skip: internalLayout},
	{id: "CalendarDataTests.AllDaysInScrapedRange_NoExceptionOnPropertyAccess", run: func(*testing.T) {
		r, _ := nepdate.NewRange(nepdate.MustNew(2078, 1, 1), nepdate.MustNew(2082, 12, 1).LastOfMonth())
		for d := range r.Days() {
			i := Of(d)
			tithi, _ := i.Tithi()
			_ = tithi.Name(nepdate.Nepali)
			_ = i.IsPublicHoliday()
			_ = eventNames(i, nepdate.English)
		}
	}},
	{id: "CalendarDataTests.AllTithiEntries_ResolveToNonEmptyStrings", run: func(t *testing.T) {
		for d := range Coverage().Days() {
			if tithi, ok := Of(d).Tithi(); ok && (tithi.Name(nepdate.English) == "" || tithi.Name(nepdate.Nepali) == "") {
				t.Fatalf("%v: empty tithi name", d)
			}
		}
	}},
	{id: "CalendarDataTests.AllHolidayOffsets_AreInBounds", skip: internalLayout},
	{id: "CalendarDataTests.AllEventOffsets_AreInBounds", skip: internalLayout},
	{id: "CalendarDataTests.HolidayCount_ViaApi_Matches294", run: func(t *testing.T) {
		r, _ := nepdate.NewRange(nepdate.MustNew(2001, 1, 1), nepdate.MustNew(2089, 12, 1).LastOfMonth())
		n := 0
		for d := range r.Days() {
			if Of(d).IsPublicHoliday() {
				n++
			}
		}
		check(t, n == 294, "294 holidays in 2001..2089")
	}},
	// C# returns empty arrays, never null; ranging over Events of a day without events yields
	// nothing.
	{id: "CalendarDataTests.CalendarInfo_EventArrays_AreNeverNull", run: func(t *testing.T) {
		check(t, Of(nepdate.MustNew(2081, 1, 1)).EventCount() > 0 && Of(nepdate.MustNew(2001, 1, 2)).EventCount() == 0, "events or none")
	}},
}

func TestCSharp(t *testing.T) {
	for _, row := range csharpRows {
		t.Run(strings.NewReplacer(".", "_", "#", "_").Replace(row.id), func(t *testing.T) {
			if row.skip != "" {
				t.Skip(row.skip)
			}
			row.run(t)
		})
	}
}

// TestCSharpCatalogueCovered checks csharpRows against every CalendarDataTests row of the
// catalogue, once each.
func TestCSharpCatalogueCovered(t *testing.T) {
	f, err := os.Open("../shared/spec/vectors/csharp-tests/catalog.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	want := map[string]bool{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		id, _, _ := strings.Cut(sc.Text(), "\t")
		if strings.HasPrefix(id, "CalendarDataTests.") {
			want[id] = true
		}
	}
	seen := map[string]bool{}
	for _, r := range csharpRows {
		if seen[r.id] || !want[r.id] || (r.skip == "") == (r.run == nil) {
			t.Errorf("row %s: duplicate, not in the catalogue, or not exactly one of skip and run", r.id)
		}
		seen[r.id] = true
	}
	if len(want) != 35 || len(seen) != len(want) {
		t.Errorf("%d rows ported, catalogue has %d (35 expected)", len(seen), len(want))
	}
}

func check(t *testing.T, ok bool, what string) {
	t.Helper()
	if !ok {
		t.Errorf("failed: %s", what)
	}
}

func eventNames(i Info, lang nepdate.Lang) []string {
	out := make([]string, 0, i.EventCount())
	for n := range i.EventCount() {
		out = append(out, i.Event(n).Name(lang))
	}
	return out
}

func everyNameNonEmpty(t *testing.T) {
	for k := 0; k+1 < len(nameOff); k++ {
		if nameOff[k+1] <= nameOff[k] {
			t.Fatalf("name %d in language %d is empty", k/2, k%2)
		}
	}
}

func emptyOn(d nepdate.Date) func(*testing.T) {
	return func(t *testing.T) {
		i := Of(d)
		_, ok := i.Tithi()
		check(t, !ok && !i.IsPublicHoliday() && i.EventCount() == 0 && i == Info{}, "empty Info")
	}
}
