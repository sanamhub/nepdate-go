// Command data-status reports how long the compiled patro data lasts.
//
//	go run ./tools/data-status
//	go run ./tools/data-status -today=2084-01-01
//
// The scan reads the generated dayInfo table through package patro, not shared/data. Tithi,
// holiday and event end dates are the last day of each kind in the table. A holiday with fewer
// than 183 days left exits 1; tithi and events with fewer than 183 days left warn and exit 0.
// The default today is Nepal time; -today takes a BS date in YYYY-MM-DD form.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	nepdate "github.com/sanamhub/nepdate-go"
	"github.com/sanamhub/nepdate-go/patro"
)

// warningDays is the shortest acceptable runway before a data refresh.
const warningDays = 183

// expiry is one kind of patro data and its last day.
type expiry struct {
	name    string
	end     nepdate.Date
	days    int
	holiday bool
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run reports the three expiries and returns the process exit code.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("data-status", flag.ContinueOnError)
	fs.SetOutput(stderr)
	todayFlag := fs.String("today", "", "BS date in YYYY-MM-DD form; default is today in Nepal")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		writeln(stderr, "data-status: unexpected argument", fs.Arg(0))
		return 2
	}
	today, err := parseToday(*todayFlag)
	if err != nil {
		writeln(stderr, "data-status:", err)
		return 2
	}
	rows, err := expiryRows(today)
	if err != nil {
		writeln(stderr, "data-status:", err)
		return 1
	}

	failed := false
	warned := false
	for _, row := range rows {
		writef(stdout, "%s: ends %s, %d days left\n", row.name, row.end, row.days)
		fail, warn := classify(row.days, row.holiday)
		switch {
		case fail:
			failed = true
			writef(stdout, "data-status: holiday data has %d days left (threshold %d)\n", row.days, warningDays)
		case warn:
			warned = true
			writef(stdout, "data-status: warning: %s data has %d days left (threshold %d)\n", row.name, row.days, warningDays)
		}
	}
	if failed {
		return 1
	}
	if warned {
		writeln(stdout, "data-status: warning")
		return 0
	}
	writeln(stdout, "data-status: ok")
	return 0
}

// writef and writeln report output errors through the exit code, not the report text.
func writef(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}

func writeln(w io.Writer, args ...any) {
	_, _ = fmt.Fprintln(w, args...)
}

// classify reports whether a kind fails or warns. Holidays fail; the other kinds warn.
func classify(days int, holiday bool) (fail, warn bool) {
	if days >= warningDays {
		return false, false
	}
	if holiday {
		return true, false
	}
	return false, true
}

// parseToday parses an explicit BS date, or returns today in Nepal for an empty value.
func parseToday(s string) (nepdate.Date, error) {
	if s == "" {
		return nepdate.Today(nil)
	}
	d, err := nepdate.Parse(s)
	if err != nil {
		return nepdate.Date{}, fmt.Errorf("invalid -today %q: %w", s, err)
	}
	return d, nil
}

// expiryRows scans every day in Coverage and records the last day of each kind.
func expiryRows(today nepdate.Date) ([]expiry, error) {
	var tithi, holiday, event nepdate.Date
	for d := range patro.Coverage().Days() {
		info := patro.Of(d)
		if _, ok := info.Tithi(); ok {
			tithi = d
		}
		if info.IsPublicHoliday() {
			holiday = d
		}
		if info.EventCount() > 0 {
			event = d
		}
	}
	if tithi.IsZero() || holiday.IsZero() || event.IsZero() {
		return nil, errors.New("patro coverage has no tithi, holiday or event day")
	}
	return []expiry{
		{name: "tithi", end: tithi, days: today.DaysUntil(tithi)},
		{name: "holidays", end: holiday, days: today.DaysUntil(holiday), holiday: true},
		{name: "events", end: event, days: today.DaysUntil(event)},
	}, nil
}
