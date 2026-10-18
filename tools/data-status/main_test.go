package main

import (
	"bytes"
	"strings"
	"testing"

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
	d, err := parseToday("2084-08-01")
	if err != nil || d.String() != "2084/08/01" {
		t.Fatalf("parseToday = %v, %v; want 2084/08/01", d, err)
	}
	if _, err := parseToday("2084-08-32"); err == nil {
		t.Error("invalid date accepted")
	}
	if _, err := parseToday(""); err != nil {
		t.Errorf("default clock failed: %v", err)
	}
}

func TestRunPassesWithWarnings(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"-today=2084-01-01"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("run = %d, want 0; stderr:\n%s", code, errOut.String())
	}
	got := out.String()
	for _, want := range []string{
		"tithi: ends 2083/12/30, -1 days left",
		"holidays: ends 2084/08/18, 233 days left",
		"events: ends 2084/08/20, 235 days left",
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
	code := run([]string{"-today=2084-08-01"}, &out, &errOut)
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
