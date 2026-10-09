package nepdate

import (
	"testing"
	"time"
)

func TestMonthNames(t *testing.T) {
	want := [12][3]string{
		{"Baishakh", "Bai", "बैशाख"},
		{"Jestha", "Jes", "जेठ"},
		{"Ashad", "Asa", "असार"},
		{"Shrawan", "Shr", "साउन"},
		{"Bhadra", "Bha", "भदौ"},
		{"Ashoj", "Aso", "असोज"},
		{"Kartik", "Kar", "कार्तिक"},
		{"Mangsir", "Man", "मंसिर"},
		{"Poush", "Pou", "पुष"},
		{"Magh", "Mag", "माघ"},
		{"Falgun", "Fal", "फागुन"},
		{"Chaitra", "Cha", "चैत"},
	}
	for i, w := range want {
		m := Month(i + 1)
		if got := m.Name(English); got != w[0] {
			t.Errorf("Month(%d).Name(English) = %q, want %q", m, got, w[0])
		}
		if got := m.String(); got != w[0] {
			t.Errorf("Month(%d).String() = %q, want %q", m, got, w[0])
		}
		if got := m.ShortName(English); got != w[1] {
			t.Errorf("Month(%d).ShortName(English) = %q, want %q", m, got, w[1])
		}
		if got := m.Name(Nepali); got != w[2] {
			t.Errorf("Month(%d).Name(Nepali) = %q, want %q", m, got, w[2])
		}
		// FORMATTING §1: Nepali has no month abbreviation, so the short name is the full name.
		if got := m.ShortName(Nepali); got != w[2] {
			t.Errorf("Month(%d).ShortName(Nepali) = %q, want %q", m, got, w[2])
		}
	}
	if Baishakh != 1 || Chaitra != 12 || Shrawan != 4 {
		t.Error("Month constants are not 1-based")
	}
}

func TestMonthNamesOutOfRange(t *testing.T) {
	for _, m := range []Month{0, 13, 255} {
		if m.String() != "" || m.Name(English) != "" || m.Name(Nepali) != "" || m.ShortName(English) != "" || m.ShortName(Nepali) != "" {
			t.Errorf("Month(%d) has a name", m)
		}
	}
	if Baishakh.Name(Lang(2)) != "" || Baishakh.ShortName(Lang(7)) != "" {
		t.Error("unknown Lang returned a name")
	}
}

func TestWeekdayNames(t *testing.T) {
	want := [7][4]string{
		{"Sunday", "Sun", "आइतवार", "आइत"},
		{"Monday", "Mon", "सोमवार", "सोम"},
		{"Tuesday", "Tue", "मङ्गलवार", "मङ्गल"},
		{"Wednesday", "Wed", "बुधवार", "बुध"},
		{"Thursday", "Thu", "बिहिवार", "बिहि"},
		{"Friday", "Fri", "शुक्रवार", "शुक्र"},
		{"Saturday", "Sat", "शनिवार", "शनि"},
	}
	for i, w := range want {
		wd := time.Weekday(i)
		got := [4]string{WeekdayName(wd, English), WeekdayShortName(wd, English), WeekdayName(wd, Nepali), WeekdayShortName(wd, Nepali)}
		if got != w {
			t.Errorf("weekday %d: got %q, want %q", i, got, w)
		}
	}
	for _, wd := range []time.Weekday{-1, 7} {
		if WeekdayName(wd, English) != "" || WeekdayShortName(wd, Nepali) != "" {
			t.Errorf("weekday %d has a name", wd)
		}
	}
	if WeekdayName(time.Monday, Lang(9)) != "" {
		t.Error("unknown Lang returned a weekday name")
	}
}
