package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	nepdate "github.com/sanamhub/nepdate-go"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		name    string
		days    int
		holiday bool
		fail    bool
		warn    bool
	}{
		{"holiday at threshold", 183, true, false, false},
		{"holiday below threshold", 182, true, true, false},
		{"tithi at threshold", 183, false, false, false},
		{"tithi below threshold", 182, false, false, true},
		{"event below threshold", 0, false, false, true},
		{"expired holiday", -1, true, true, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fail, warn := classify(c.days, c.holiday)
			if fail != c.fail || warn != c.warn {
				t.Errorf("classify(%d, %v) = fail=%v warn=%v, want fail=%v warn=%v",
					c.days, c.holiday, fail, warn, c.fail, c.warn)
			}
		})
	}
}

func TestExpiryRows(t *testing.T) {
	today := nepdate.MustNew(2084, nepdate.Baishakh, 1)
	rows, err := expiryRows(today)
	if err != nil {
		t.Fatal(err)
	}
	want := []struct {
		name string
		end  string
		days int
	}{
		{"tithi", "2083/12/30", -1},
		{"holidays", "2084/08/18", 233},
		{"events", "2084/08/20", 235},
	}
	if len(rows) != len(want) {
		t.Fatalf("got %d rows, want %d", len(rows), len(want))
	}
	for i, w := range want {
		if rows[i].name != w.name || rows[i].end.String() != w.end || rows[i].days != w.days {
			t.Errorf("row %d = %s %s %d days, want %s %s %d days",
				i, rows[i].name, rows[i].end, rows[i].days, w.name, w.end, w.days)
		}
	}
}

func TestParseToday(t *testing.T) {
	d, err := parseToday("2027-11-16", time.Time{})
	if err != nil || d.String() != "2084/08/01" {
		t.Fatalf("parseToday = %v, %v; want 2084/08/01", d, err)
	}
	for _, bad := range []string{"2027-11-31", "2084/08/01", "2084-08-01x"} {
		if _, err := parseToday(bad, time.Time{}); err == nil {
			t.Errorf("parseToday(%q) accepted", bad)
		}
	}
}

// TestParseTodayDefaultIsUTC pins the default to the UTC calendar date. At 20:00 UTC it is
// already the next day in Nepal (+05:45), and the gate must still use the UTC day, the rule every
// port's data gate shares.
func TestParseTodayDefaultIsUTC(t *testing.T) {
	now := time.Date(2027, time.June, 3, 20, 0, 0, 0, time.UTC)
	if nepal := now.In(time.FixedZone("NPT", 5*3600+45*60)); nepal.Day() != 4 {
		t.Fatalf("test premise: Nepal day is %d, want 4", nepal.Day())
	}
	d, err := parseToday("", now.In(time.FixedZone("NPT", 5*3600+45*60)))
	if err != nil {
		t.Fatal(err)
	}
	want, err := parseToday("2027-06-03", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if d != want {
		t.Errorf("default today = %s, want %s (the UTC day 2027-06-03)", d, want)
	}
}

// TestHolidayBoundary checks the 183-day rule on the real tables with AD dates, the same dates the
// Rust and Python gates use: 2027-06-03 leaves 183 holiday days, 2027-06-04 leaves 182.
func TestHolidayBoundary(t *testing.T) {
	for _, c := range []struct {
		today string
		code  int
		want  string
	}{
		{"2027-06-03", 0, "holidays: ends 2084/08/18 (AD 2027-12-03), 183 days left"},
		{"2027-06-04", 1, "data-status: holiday data has 182 days left (threshold 183)"},
	} {
		var out, errOut bytes.Buffer
		if code := run([]string{"-today=" + c.today}, &out, &errOut); code != c.code {
			t.Errorf("-today=%s: run = %d, want %d\n%s%s", c.today, code, c.code, out.String(), errOut.String())
		}
		if !strings.Contains(out.String(), c.want) {
			t.Errorf("-today=%s: output lacks %q:\n%s", c.today, c.want, out.String())
		}
	}
}

func TestRunPassesWithWarnings(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"-today=2027-04-14"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("run = %d, want 0; stderr:\n%s", code, errOut.String())
	}
	got := out.String()
	for _, want := range []string{
		"tithi: ends 2083/12/30 (AD 2027-04-13), -1 days left",
		"holidays: ends 2084/08/18 (AD 2027-12-03), 233 days left",
		"events: ends 2084/08/20 (AD 2027-12-05), 235 days left",
		"data-status: warning: tithi data has -1 days left (threshold 183)",
		"data-status: warning\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output lacks %q:\n%s", want, got)
		}
	}
}

func TestRunFailsOnHolidayExpiry(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"-today=2027-11-16"}, &out, &errOut)
	if code != 1 {
		t.Fatalf("run = %d, want 1; stdout:\n%s\nstderr:\n%s", code, out.String(), errOut.String())
	}
	if !strings.Contains(out.String(), "data-status: holiday data has 17 days left (threshold 183)") {
		t.Errorf("output lacks the holiday failure:\n%s", out.String())
	}
}

func TestRunRejectsExtraArgument(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"extra"}, &out, &errOut); code != 2 {
		t.Errorf("run = %d, want 2", code)
	}
}
