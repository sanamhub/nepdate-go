package main

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
)

func validInput() *calendarFile {
	cal := &calendarFile{MinYear: 1901, MaxYear: 2199, MonthLengths: map[string][]int{}}
	for y := 1901; y <= 2199; y++ {
		cal.MonthLengths[strconv.Itoa(y)] = []int{31, 31, 32, 31, 31, 31, 30, 29, 30, 29, 30, 30}
	}
	return cal
}

func TestParseCalendarRejectsLength33(t *testing.T) {
	cal := validInput()
	cal.MonthLengths["2081"][3] = 33
	_, err := parseCalendar(cal)
	if err == nil {
		t.Fatal("month length 33 accepted")
	}
	if msg := err.Error(); !strings.Contains(msg, "2081") || !strings.Contains(msg, "month 4") {
		t.Fatalf("error %q does not name year 2081 and month 4", msg)
	}
}

func TestParseCalendarRejectsMissingYear(t *testing.T) {
	cal := validInput()
	delete(cal.MonthLengths, "2000")
	if _, err := parseCalendar(cal); err == nil || !strings.Contains(err.Error(), "2000") {
		t.Fatalf("missing year 2000: got %v", err)
	}
}

func TestParseCalendarRejectsElevenMonths(t *testing.T) {
	cal := validInput()
	cal.MonthLengths["1950"] = cal.MonthLengths["1950"][:11]
	if _, err := parseCalendar(cal); err == nil || !strings.Contains(err.Error(), "1950") {
		t.Fatalf("11 months in 1950: got %v", err)
	}
}

func TestGenerateIsDeterministic(t *testing.T) {
	first, err := generate("../..")
	if err != nil {
		t.Fatal(err)
	}
	second, err := generate("../..")
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != len(second) {
		t.Fatalf("%d files, then %d", len(first), len(second))
	}
	for i := range first {
		if first[i].path != second[i].path || !bytes.Equal(first[i].data, second[i].data) {
			t.Fatalf("%s differs between two runs", first[i].path)
		}
	}
}

func TestGeneratedHeader(t *testing.T) {
	files, err := generate("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if !bytes.HasPrefix(f.data, []byte(header)) {
			t.Errorf("%s does not start with the generated-code header", f.path)
		}
	}
}

func TestCheckFindsDrift(t *testing.T) {
	files, err := generate("../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := writeFiles(dir, files); err != nil {
		t.Fatal(err)
	}
	if diff, err := check(dir, files); err != nil || len(diff) != 0 {
		t.Fatalf("fresh tree: diff %v, err %v", diff, err)
	}
	changed := append([]byte(nil), files[0].data...)
	changed[len(changed)-2] ^= 1
	if err := writeFiles(dir, []output{{path: files[0].path, data: changed}}); err != nil {
		t.Fatal(err)
	}
	diff, err := check(dir, files)
	if err != nil {
		t.Fatal(err)
	}
	if len(diff) != 1 || diff[0] != files[0].path {
		t.Fatalf("after one changed byte: diff %v, want [%s]", diff, files[0].path)
	}
}

func TestTablesFromData(t *testing.T) {
	files, err := generate("../..")
	if err != nil {
		t.Fatal(err)
	}
	src := string(files[0].data)
	for _, want := range []string{"bsMonthStart = [3589]int32{", "adMonthStart = [3601]int32{", "// 2081\n", "maxSerial     = 109211"} {
		if !strings.Contains(src, want) {
			t.Errorf("calendar_gen.go lacks %q", want)
		}
	}
	if strings.Contains(src, "func init") {
		t.Error("calendar_gen.go has an init function")
	}
}

func TestVerifyEstimateRejectsOffByOne(t *testing.T) {
	cal, err := loadCalendar("../..")
	if err != nil {
		t.Fatal(err)
	}
	bs := bsStarts(cal)
	mul, add, err := searchEstimate(bs[:])
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyEstimate(bs[:], mul, add); err != nil {
		t.Fatalf("searched constants fail verification: %v", err)
	}
	if err := verifyEstimate(bs[:], mul, add-1<<estShift); err == nil {
		t.Error("estimate one month too low accepted")
	}
	if err := verifyEstimate(bs[:], mul, add+1<<estShift); err == nil {
		t.Error("estimate one month too high accepted")
	}
	if err := verifyEstimate(bs[:], mul+400, add); err == nil {
		t.Error("multiplier 400 too high accepted")
	}
}

func TestADStarts(t *testing.T) {
	ad := adStarts()
	if ad[0] != -101 {
		t.Errorf("adMonthStart[0] = %d, want -101", ad[0])
	}
	if ad[3] != -10 || ad[4] != 20 {
		t.Errorf("AD 1844-04 and 1844-05 start at %d and %d, want -10 and 20", ad[3], ad[4])
	}
}
