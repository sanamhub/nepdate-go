package main

import (
	"fmt"
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

// TestPatroCounts checks the tables against shared/spec/vectors/patro-summary.json.
func TestPatroCounts(t *testing.T) {
	cal, files := loadTestPatro(t)
	p, err := buildPatro(cal, files)
	if err != nil {
		t.Fatal(err)
	}
	sum, err := loadPatroSummary("../..")
	if err != nil {
		t.Fatal(err)
	}
	if err := checkPatroSummary(cal, p, sum); err != nil {
		t.Fatal(err)
	}
	// Tithi names come first, so every tithi id is below the number of tithi names.
	if p.tithiNames >= 63 {
		t.Errorf("%d tithi names", p.tithiNames)
	}
}

// TestPatroSummaryMismatch checks that a summary count that disagrees with the data fails, naming
// the field and both values.
func TestPatroSummaryMismatch(t *testing.T) {
	cal, files := loadTestPatro(t)
	p, err := buildPatro(cal, files)
	if err != nil {
		t.Fatal(err)
	}
	sum, err := loadPatroSummary("../..")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		field string
		edit  func(*patroSummary)
	}{
		{"names", func(s *patroSummary) { s.Names++ }},
		{"days", func(s *patroSummary) { s.Days-- }},
		{"holidays", func(s *patroSummary) { s.Holidays++ }},
		{"name_bytes", func(s *patroSummary) { s.NameBytes-- }},
		{"tithi_first_day", func(s *patroSummary) { s.TithiFirstDay = "2000-01-01" }},
	}
	for _, c := range cases {
		bad := *sum
		c.edit(&bad)
		err := checkPatroSummary(cal, p, &bad)
		if err == nil || !strings.Contains(err.Error(), c.field) || !strings.Contains(err.Error(), "summary says") || !strings.Contains(err.Error(), "the data has") {
			t.Errorf("summary with a wrong %s gave %v", c.field, err)
		}
	}
}

// TestPatroSummaryMissing checks that a shared/ without the summary is told to resync.
func TestPatroSummaryMissing(t *testing.T) {
	if _, err := loadPatroSummary(t.TempDir()); err == nil || !strings.Contains(err.Error(), "resync") {
		t.Fatalf("a missing summary gave %v", err)
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
	sum, err := loadPatroSummary("../..")
	if err != nil {
		t.Fatal(err)
	}
	var src string
	for _, f := range files {
		if f.path == "patro/patro_gen.go" {
			src = string(f.data)
		}
	}
	want := []string{
		"package patro",
		fmt.Sprintf("dayInfo = [%d]uint32{", sum.Days),
		fmt.Sprintf("eventName = [%d]uint16{", sum.Events),
		fmt.Sprintf("nameOff = [%d]uint32{", 2*sum.Names+1),
		"// 2081-01-01\n",
		fmt.Sprintf("covFirst = %d", sum.FirstSerial),
	}
	for _, w := range want {
		if !strings.Contains(src, w) {
			t.Errorf("patro_gen.go lacks %q", w)
		}
	}
	if strings.Contains(src, "func init") {
		t.Error("patro_gen.go has an init function")
	}
}
