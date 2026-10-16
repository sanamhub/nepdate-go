package nepdate

import "time"

// Lang selects the language of names and digits in formatted output.
type Lang uint8

// The supported languages. English uses ASCII digits, Nepali uses Devanagari digits.
const (
	English Lang = iota
	Nepali
)

// Month is a Bikram Sambat month, Baishakh = 1 to Chaitra = 12.
type Month uint8

// The twelve BS months.
const (
	Baishakh Month = 1 + iota
	Jestha
	Ashad
	Shrawan
	Bhadra
	Ashoj
	Kartik
	Mangsir
	Poush
	Magh
	Falgun
	Chaitra
)

// Names from shared/spec/FORMATTING.md §1. Index 0 is Baishakh or Sunday.
var (
	monthNames = [2][12]string{
		{"Baishakh", "Jestha", "Ashad", "Shrawan", "Bhadra", "Ashoj", "Kartik", "Mangsir", "Poush", "Magh", "Falgun", "Chaitra"},
		{"बैशाख", "जेठ", "असार", "साउन", "भदौ", "असोज", "कार्तिक", "मंसिर", "पुष", "माघ", "फागुन", "चैत"},
	}
	// Ashad and Ashoj are Asa and Aso, not C#'s shared "Ash" (PARITY D-04).
	monthShortNames = [12]string{"Bai", "Jes", "Asa", "Shr", "Bha", "Aso", "Kar", "Man", "Pou", "Mag", "Fal", "Cha"}

	weekdayNames = [2][7]string{
		{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"},
		{"आइतवार", "सोमवार", "मङ्गलवार", "बुधवार", "बिहिवार", "शुक्रवार", "शनिवार"},
	}
	weekdayShortNames = [2][7]string{
		{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},
		{"आइत", "सोम", "मङ्गल", "बुध", "बिहि", "शुक्र", "शनि"},
	}
)

// Name returns the month's full name in lang, or "" if m is not 1..12 or lang is unknown. It
// does not allocate.
func (m Month) Name(lang Lang) string {
	if m < Baishakh || m > Chaitra || lang > Nepali {
		return ""
	}
	return monthNames[lang][m-1]
}

// ShortName returns the three-letter English abbreviation, or the full Nepali name for Nepali,
// which has no customary abbreviation. Ashad is "Asa" and Ashoj is "Aso" (PARITY D-04). It
// returns "" if m is not 1..12 or lang is unknown. It does not allocate.
func (m Month) ShortName(lang Lang) string {
	if m < Baishakh || m > Chaitra || lang > Nepali {
		return ""
	}
	if lang == Nepali {
		return monthNames[Nepali][m-1]
	}
	return monthShortNames[m-1]
}

// String returns the English name of the month, or "" if m is not 1..12.
func (m Month) String() string {
	return m.Name(English)
}

// WeekdayName returns the full name of w in lang, or "" if w is not Sunday..Saturday or lang is
// unknown. It does not allocate.
func WeekdayName(w time.Weekday, lang Lang) string {
	if w < time.Sunday || w > time.Saturday || lang > Nepali {
		return ""
	}
	return weekdayNames[lang][w]
}

// WeekdayShortName returns the short name of w in lang ("Tue", "मङ्गल"), or "" if w is not
// Sunday..Saturday or lang is unknown. It does not allocate.
func WeekdayShortName(w time.Weekday, lang Lang) string {
	if w < time.Sunday || w > time.Saturday || lang > Nepali {
		return ""
	}
	return weekdayShortNames[lang][w]
}
