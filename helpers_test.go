package nepdate

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// readTable reads a CSV or TSV vector file and returns its rows without the header, as maps
// from column name to cell. It fails the test unless the file has exactly wantRows rows, so a
// truncated, regenerated or grown vector file can't pass unnoticed.
func readTable(t testing.TB, path string, sep string, wantRows int) []map[string]string {
	t.Helper()
	f, err := os.Open(path) //nolint:gosec // test reads vector files of the repository
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<16), 1<<20)
	var header []string
	var rows []map[string]string
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		cells := strings.Split(line, sep)
		if header == nil {
			header = cells
			continue
		}
		row := make(map[string]string, len(header))
		for i, h := range header {
			if i < len(cells) {
				row[h] = cells[i]
			}
		}
		rows = append(rows, row)
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if len(rows) != wantRows {
		t.Fatalf("%s: %d rows, want exactly %d", path, len(rows), wantRows)
	}
	return rows
}

// ymd splits "YYYY-MM-DD" into numbers.
func ymd(t testing.TB, s string) (y, m, d int) {
	t.Helper()
	parts := strings.Split(s, "-")
	if len(parts) != 3 {
		t.Fatalf("bad date %q", s)
	}
	var n [3]int
	for i, p := range parts {
		v, err := strconv.Atoi(p)
		if err != nil {
			t.Fatalf("bad date %q", s)
		}
		n[i] = v
	}
	return n[0], n[1], n[2]
}

// bsDate parses a "YYYY-MM-DD" BS date from a vector file.
func bsDate(t testing.TB, s string) Date {
	t.Helper()
	y, m, d := ymd(t, s)
	date, err := New(y, Month(m), d) //nolint:gosec // New rejects a month outside 1..12
	if err != nil {
		t.Fatalf("vector date %s: %v", s, err)
	}
	return date
}

var weekdayByName = map[string]time.Weekday{
	"Sunday": time.Sunday, "Monday": time.Monday, "Tuesday": time.Tuesday, "Wednesday": time.Wednesday,
	"Thursday": time.Thursday, "Friday": time.Friday, "Saturday": time.Saturday,
}

// benchmarkDates are the 64 dates of shared/spec/BENCHMARK.md §2.
func benchmarkDates() [64]Date {
	s0 := toSerial(MustNew(2000, Baishakh, 1))
	s1 := toSerial(MustNew(2090, Chaitra, 30))
	var ds [64]Date
	for i := range ds {
		ds[i] = fromSerial(s0 + int32(i)*((s1-s0)/63))
	}
	return ds
}
