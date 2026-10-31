package patro

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	nepdate "github.com/sanamhub/nepdate-go"
)

type jsonPair struct{ Np, En string }

type jsonDay struct {
	Tithi   *jsonPair  `json:"tithi"`
	Holiday bool       `json:"holiday"`
	Events  []jsonPair `json:"events"`
}

// loadData reads shared/data/patro/*.json directly, so the tests don't trust the generator.
func loadData(t *testing.T) map[nepdate.Date]jsonDay {
	t.Helper()
	paths, err := filepath.Glob("../shared/data/patro/*.json")
	if err != nil || len(paths) == 0 {
		t.Fatalf("patro data: %v", err)
	}
	out := map[nepdate.Date]jsonDay{}
	for _, p := range paths {
		raw, err := os.ReadFile(p) //nolint:gosec // test reads the repository's data
		if err != nil {
			t.Fatal(err)
		}
		var f struct{ Days map[string]jsonDay }
		if err := json.Unmarshal(raw, &f); err != nil {
			t.Fatal(err)
		}
		for date, day := range f.Days {
			d, err := nepdate.Parse(date)
			if err != nil {
				t.Fatal(err)
			}
			out[d] = day
		}
	}
	return out
}

func names(i Info, lang nepdate.Lang) []string {
	var out []string
	for e := range i.Events() {
		out = append(out, e.Name(lang))
	}
	return out
}

// TestAllData compares Of with the JSON for every day of the coverage, in both languages.
func TestAllData(t *testing.T) {
	data := loadData(t)
	for d := range Coverage().Days() {
		want, i := data[d], Of(d)
		tithi, ok := i.Tithi()
		if (want.Tithi != nil) != ok || ok && (tithi.Name(nepdate.English) != want.Tithi.En || tithi.Name(nepdate.Nepali) != want.Tithi.Np) {
			t.Fatalf("%v: tithi %q, %v; want %+v", d, tithi.Name(nepdate.English), ok, want.Tithi)
		}
		if i.IsPublicHoliday() != want.Holiday {
			t.Fatalf("%v: holiday %v, want %v", d, i.IsPublicHoliday(), want.Holiday)
		}
		var en, np []string
		for _, e := range want.Events {
			en, np = append(en, e.En), append(np, e.Np)
		}
		if !slices.Equal(names(i, nepdate.English), en) || !slices.Equal(names(i, nepdate.Nepali), np) || i.EventCount() != len(en) {
			t.Fatalf("%v: events %q, want %q", d, names(i, nepdate.English), en)
		}
		for n := range i.EventCount() {
			if i.Event(n).Name(nepdate.Nepali) != np[n] {
				t.Fatalf("%v: Event(%d) = %q", d, n, i.Event(n).Name(nepdate.Nepali))
			}
		}
	}
}

// TestAnchors is the table of PATRO §4.
func TestAnchors(t *testing.T) {
	cases := []struct {
		d       nepdate.Date
		tithi   string
		holiday bool
		events  []string
	}{
		{nepdate.MustNew(2081, 1, 1), "Panchami", true, []string{"New Year", "Mesh Sankranti", "Bisket Jatra"}},
		{nepdate.MustNew(2081, 1, 2), "Shashthi", false, nil},
		{nepdate.MustNew(2081, 2, 15), "Panchami", true, []string{"Republic Day"}},
		{nepdate.MustNew(2080, 11, 1), "Chaturthi", false, []string{"Mangala Chauthi Brata", "People's War Day", "Tilakund Chauthi", "Kumbha Sankranti", "World Radio Day"}},
		{nepdate.MustNew(2081, 4, 15), "Dashami", false, []string{"Kheer Khane Din", "World Day Against Human Trafficking", "International Friendship Day"}},
	}
	for _, c := range cases {
		i := Of(c.d)
		tithi, ok := i.Tithi()
		if !ok || tithi.Name(nepdate.English) != c.tithi || tithi.Name(nepdate.Nepali) == "" {
			t.Errorf("%v: tithi %q, %v; want %q", c.d, tithi.Name(nepdate.English), ok, c.tithi)
		}
		if i.IsPublicHoliday() != c.holiday {
			t.Errorf("%v: holiday %v", c.d, i.IsPublicHoliday())
		}
		if got := names(i, nepdate.English); !slices.Equal(got, c.events) {
			t.Errorf("%v: events %q, want %q", c.d, got, c.events)
		}
		if got := names(i, nepdate.Nepali); len(got) != len(c.events) {
			t.Errorf("%v: %d Nepali event names", c.d, len(got))
		}
	}
	// 2081-01-01 in Nepali, from the data file.
	if got := names(Of(nepdate.MustNew(2081, 1, 1)), nepdate.Nepali); !slices.Equal(got, []string{"नयाँ वर्ष", "मेष संक्रान्ति", "बिस्का: जात्रा"}) {
		t.Errorf("2081-01-01 Nepali events %q", got)
	}
}

func TestEmpty(t *testing.T) {
	for _, d := range []nepdate.Date{{}, nepdate.MustNew(2070, 6, 15), nepdate.MustNew(2090, 1, 1), nepdate.Min, nepdate.Max} {
		i := Of(d)
		if _, ok := i.Tithi(); ok || i.IsPublicHoliday() || i.EventCount() != 0 || len(names(i, nepdate.English)) != 0 {
			t.Errorf("Of(%v) is not empty", d)
		}
	}
	if (Info{}).Event(0).Name(nepdate.English) != "" {
		t.Error("Info{}.Event(0) has a name")
	}
	i := Of(nepdate.MustNew(2081, 1, 1))
	if i.Event(-1).Name(nepdate.English) != "" || i.Event(3).Name(nepdate.Nepali) != "" {
		t.Error("an out-of-range Event has a name")
	}
	tithi, _ := i.Tithi()
	if tithi.Name(nepdate.Lang(2)) != "" || i.Event(0).Name(nepdate.Lang(9)) != "" || (Tithi{}).Name(nepdate.English) != "" {
		t.Error("an unknown Lang or zero Tithi has a name")
	}
}

// TestSummary compares the compiled tables and coverage with
// shared/spec/vectors/patro-summary.json, so dropped or duplicated data fails here.
func TestSummary(t *testing.T) {
	sum := loadSummary(t)
	start, err := nepdate.Parse(sum.FirstDay)
	if err != nil {
		t.Fatal(err)
	}
	end, err := nepdate.Parse(sum.LastDay)
	if err != nil {
		t.Fatal(err)
	}
	tithiStart, err := nepdate.Parse(sum.TithiFirstDay)
	if err != nil {
		t.Fatal(err)
	}
	tithiEnd, err := nepdate.Parse(sum.TithiLastDay)
	if err != nil {
		t.Fatal(err)
	}
	if r := Coverage(); r.Start() != start || r.End() != end || r.Len() != sum.Days {
		t.Errorf("Coverage() = %v..%v (%d days), want %v..%v (%d days)",
			r.Start(), r.End(), r.Len(), start, end, sum.Days)
	}
	if got := nepdate.Min.DaysUntil(start); got != sum.FirstSerial {
		t.Errorf("first coverage serial %d, want %d", got, sum.FirstSerial)
	}
	if got := nepdate.Min.DaysUntil(end); got != sum.LastSerial {
		t.Errorf("last coverage serial %d, want %d", got, sum.LastSerial)
	}

	daysWithTithi, holidays, eventDays, events := 0, 0, 0, 0
	var firstTithi, lastTithi nepdate.Date
	for i, rec := range dayInfo {
		if rec&tithiMask != 0 {
			daysWithTithi++
			d, err := start.AddDays(i)
			if err != nil {
				t.Fatal(err)
			}
			if firstTithi.IsZero() {
				firstTithi = d
			}
			lastTithi = d
		}
		if rec&holidayBit != 0 {
			holidays++
		}
		if n := int((rec >> countShift) & countMask); n > 0 {
			eventDays++
			events += n
		}
	}
	eventIDs := make(map[uint16]bool, len(eventName))
	for _, id := range eventName {
		eventIDs[id] = true
	}
	checks := []struct {
		name string
		got  int
		want int
	}{
		{"dayInfo entries", len(dayInfo), sum.Days},
		{"eventName entries", len(eventName), sum.Events},
		{"nameOff entries", len(nameOff), 2*sum.Names + 1},
		{"nameText bytes", len(nameText), sum.NameBytes},
		{"tithi names", tithiNames, sum.TithiNames},
		{"tithi days", daysWithTithi, sum.TithiDays},
		{"holidays", holidays, sum.Holidays},
		{"event days", eventDays, sum.EventDays},
		{"events", events, sum.Events},
		{"event names", len(eventIDs), sum.EventNames},
		{"names", (len(nameOff) - 1) / 2, sum.Names},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d from the summary", c.name, c.got, c.want)
		}
	}
	if firstTithi != tithiStart || lastTithi != tithiEnd {
		t.Errorf("tithi range %v..%v, want %v..%v", firstTithi, lastTithi, tithiStart, tithiEnd)
	}
	if got := nepdate.Min.DaysUntil(firstTithi); got != sum.TithiFirstSerial {
		t.Errorf("first tithi serial %d, want %d", got, sum.TithiFirstSerial)
	}
}

func TestPublicHolidays(t *testing.T) {
	data := loadData(t)
	sum := loadSummary(t)
	var got []nepdate.Date
	for d := range PublicHolidays() {
		got = append(got, d)
	}
	if len(got) != sum.Holidays || !slices.IsSortedFunc(got, nepdate.Date.Compare) {
		t.Fatalf("%d holidays, want %d, sorted %v", len(got), sum.Holidays, slices.IsSortedFunc(got, nepdate.Date.Compare))
	}
	for _, d := range got {
		if !data[d].Holiday {
			t.Errorf("%v is not a holiday in the data", d)
		}
	}
	n := 0
	for range PublicHolidays() {
		n++
		if n == 2 {
			break
		}
	}
	if n != 2 {
		t.Error("break did not stop PublicHolidays")
	}
}
