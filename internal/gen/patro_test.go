package main

import (
	"strings"
	"testing"
)

func loadTestPatro(t *testing.T) (*calendar, map[int]*patroFile) {
	t.Helper()
	cal, err := loadCalendar("../..")
	if err != nil {
		t.Fatal(err)
	}
	files, err := loadPatroFiles("../..")
	if err != nil {
		t.Fatal(err)
	}
	return cal, files
}

// TestPatroCounts checks the PATRO §3 counts from the generated tables.
func TestPatroCounts(t *testing.T) {
	cal, files := loadTestPatro(t)
	p, err := buildPatro(cal, files)
	if err != nil {
		t.Fatal(err)
	}
	tithi, holidays, eventDays, events := 0, 0, 0, 0
	for _, rec := range p.dayInfo {
		if rec&tithiMask != 0 {
			tithi++
		}
		if rec&holidayBit != 0 {
			holidays++
		}
		if n := (rec >> countShift) & countMask; n > 0 {
			eventDays++
			events += int(n)
		}
	}
	checks := []struct {
		name      string
		got, want int
	}{
		{"T (tithi days)", tithi, 2190},
		{"H (holidays)", holidays, 294},
		{"E (event days)", eventDays, 1671},
		{"events", events, 3032},
		{"eventName entries", len(p.eventName), 3032},
		{"name pairs", len(p.pairs), 753},
		{"nameOff entries", len(p.nameOff), 1507},
		{"dayInfo entries", len(p.dayInfo), 2793},
		{"covFirst", int(p.covFirst), 64285},
		{"covLast", int(p.covLast), 67077},
		{"nameText bytes", len(p.nameText), 59800},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}
	// Tithi names come first, so every tithi id is below the number of tithi names.
	if p.tithiNames >= 63 {
		t.Errorf("%d tithi names", p.tithiNames)
	}
}

func TestPatroRejectsTithiGap(t *testing.T) {
	cal, files := loadTestPatro(t)
	day := files[2081].Days["2081-05-10"]
	day.Tithi = nil
	files[2081].Days["2081-05-10"] = day
	_, err := buildPatro(cal, files)
	if err == nil || !strings.Contains(err.Error(), "2081-05-10") {
		t.Fatalf("a missing tithi day gave %v", err)
	}
}

func TestPatroRejectsSixteenEvents(t *testing.T) {
	cal, files := loadTestPatro(t)
	day := files[2081].Days["2081-01-01"]
	for len(day.Events) < 16 {
		day.Events = append(day.Events, namePair{Np: "क", En: "x"})
	}
	files[2081].Days["2081-01-01"] = day
	_, err := buildPatro(cal, files)
	if err == nil || !strings.Contains(err.Error(), "16 events") {
		t.Fatalf("16 events on one day gave %v", err)
	}
}

func TestPatroRejectsTooManyTithiNames(t *testing.T) {
	cal, files := loadTestPatro(t)
	n := 0
	for date, day := range files[2081].Days {
		if day.Tithi != nil && n < 70 {
			day.Tithi = &namePair{Np: "ति" + date, En: "t" + date}
			files[2081].Days[date] = day
			n++
		}
	}
	if _, err := buildPatro(cal, files); err == nil || !strings.Contains(err.Error(), "tithi names") {
		t.Fatalf("70 distinct tithi names gave %v", err)
	}
}

func TestPatroSource(t *testing.T) {
	files, err := generate("../..")
	if err != nil {
		t.Fatal(err)
	}
	var src string
	for _, f := range files {
		if f.path == "patro/patro_gen.go" {
			src = string(f.data)
		}
	}
	for _, want := range []string{"package patro", "dayInfo = [2793]uint32{", "eventName = [3032]uint16{", "nameOff = [1507]uint32{", "// 2081-01-01\n", "covFirst = 64285"} {
		if !strings.Contains(src, want) {
			t.Errorf("patro_gen.go lacks %q", want)
		}
	}
	if strings.Contains(src, "func init") {
		t.Error("patro_gen.go has an init function")
	}
}
