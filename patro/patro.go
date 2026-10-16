// Package patro is the Nepali almanac (पात्रो) for package nepdate: the tithi (lunar day), the
// public holiday flag and the events of a BS date.
//
// Data covers BS 2077-01-01 to 2084-08-20 (see Coverage); dates outside it have no tithi, no
// holiday and no events. The tables are compiled in, so no lookup reads a file or allocates.
// Package nepdate never imports this package, so a program that only converts dates links none
// of this data.
package patro

import (
	"iter"

	nepdate "github.com/sanamhub/nepdate-go"
)

// Info is the almanac entry of one date. The zero Info is empty.
type Info struct {
	rec uint32
}

// Tithi is a lunar day name. The zero Tithi has no name.
type Tithi struct {
	ref uint16 // name id + 1
}

// Event is a named event or festival. The zero Event has no name.
type Event struct {
	ref uint16 // name id + 1
}

// Of returns the almanac entry of d. It never fails: the zero Date and dates outside Coverage
// return the empty Info. It does not allocate.
func Of(d nepdate.Date) Info {
	if d.IsZero() {
		return Info{}
	}
	i := nepdate.Min.DaysUntil(d) - covFirst
	if i < 0 || i >= len(dayInfo) {
		return Info{}
	}
	return Info{dayInfo[i]}
}

// Tithi returns the tithi of the day, and false if the day has none. It does not allocate.
func (i Info) Tithi() (Tithi, bool) {
	id := i.rec & tithiMask
	return Tithi{uint16(id)}, id != 0
}

// IsPublicHoliday reports whether the day is a gazetted public holiday. It does not allocate.
func (i Info) IsPublicHoliday() bool {
	return i.rec&holidayBit != 0
}

// EventCount returns the number of events of the day. It does not allocate.
func (i Info) EventCount() int {
	return int(i.rec >> countShift & countMask)
}

// Event returns the n-th event of the day in file order, or the zero Event (Name "") if n is
// out of range. Looping over EventCount and Event needs no closure, unlike Events. It does not
// allocate.
func (i Info) Event(n int) Event {
	if n < 0 || n >= i.EventCount() {
		return Event{}
	}
	return Event{eventName[int(i.rec>>firstShift)+n] + 1}
}

// Events returns the events of the day in file order. Ranging over it directly does not
// allocate.
func (i Info) Events() iter.Seq[Event] {
	return func(yield func(Event) bool) {
		for n := range i.EventCount() {
			if !yield(i.Event(n)) {
				return
			}
		}
	}
}

// Name returns the tithi name in lang, or "" for the zero Tithi or an unknown lang. It does not
// allocate.
func (t Tithi) Name(lang nepdate.Lang) string {
	return name(t.ref, lang)
}

// Name returns the event name in lang, or "" for the zero Event or an unknown lang. It does not
// allocate.
func (e Event) Name(lang nepdate.Lang) string {
	return name(e.ref, lang)
}

// name slices name ref-1 in lang out of nameText; it never allocates.
func name(ref uint16, lang nepdate.Lang) string {
	if ref == 0 || lang > nepdate.Nepali {
		return ""
	}
	k := 2*(int(ref)-1) + int(lang)
	return nameText[nameOff[k]:nameOff[k+1]]
}

// Coverage returns the smallest range that contains every day with almanac data, BS 2077-01-01
// to 2084-08-20 with today's data. It does not allocate.
func Coverage() nepdate.Range {
	first, _ := nepdate.Min.AddDays(covFirst)
	last, _ := nepdate.Min.AddDays(covLast)
	r, _ := nepdate.NewRange(first, last)
	return r
}

// PublicHolidays returns every public holiday in ascending order. Ranging over it directly does
// not allocate.
func PublicHolidays() iter.Seq[nepdate.Date] {
	return func(yield func(nepdate.Date) bool) {
		for i, rec := range dayInfo {
			if rec&holidayBit == 0 {
				continue
			}
			d, _ := nepdate.Min.AddDays(covFirst + i)
			if !yield(d) {
				return
			}
		}
	}
}
