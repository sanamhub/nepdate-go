package nepdate

import (
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The name tables of names.go and lenient.go are typed by hand. These tests read the same
// tables from shared/spec, so a spec change that isn't copied into the code fails here.

// specSection returns the lines of the "## <n>." section of a spec file.
func specSection(t *testing.T, file string, n int) []string {
	t.Helper()
	raw, err := os.ReadFile("shared/spec/" + file) //nolint:gosec // reads the repository's spec
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	in := false
	for _, line := range strings.Split(strings.ReplaceAll(string(raw), "\r", ""), "\n") {
		if strings.HasPrefix(line, "## ") {
			in = strings.HasPrefix(line, "## "+strconv.Itoa(n)+".")
			continue
		}
		if in {
			out = append(out, line)
		}
	}
	if len(out) == 0 {
		t.Fatalf("%s: no section %d", file, n)
	}
	return out
}

// tableRows returns the cells of the markdown table rows whose first cell is a number.
func tableRows(lines []string) [][]string {
	var rows [][]string
	for _, line := range lines {
		if !strings.HasPrefix(line, "|") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		for i, c := range cells {
			cells[i] = strings.Trim(strings.TrimSpace(c), "*")
		}
		if _, err := strconv.Atoi(cells[0]); err == nil {
			rows = append(rows, cells)
		}
	}
	return rows
}

func TestSpecMonthAndWeekdayNames(t *testing.T) {
	rows := tableRows(specSection(t, "FORMATTING.md", 1))
	var months, weekdays [][]string
	for _, r := range rows {
		switch len(r) {
		case 4:
			months = append(months, r)
		case 5:
			weekdays = append(weekdays, r)
		}
	}
	if len(months) != 12 || len(weekdays) != 7 {
		t.Fatalf("FORMATTING §1: %d month rows and %d weekday rows, want 12 and 7", len(months), len(weekdays))
	}
	for i, r := range months {
		m := Month(i + 1)
		got := []string{m.Name(English), m.ShortName(English), m.Name(Nepali)}
		if !slices.Equal(got, r[1:]) {
			t.Errorf("month %d: code %q, spec %q", i+1, got, r[1:])
		}
	}
	for i, r := range weekdays {
		w := time.Weekday(i)
		got := []string{
			WeekdayName(w, English), WeekdayShortName(w, English),
			WeekdayName(w, Nepali), WeekdayShortName(w, Nepali),
		}
		if !slices.Equal(got, r[1:]) {
			t.Errorf("weekday %d: code %q, spec %q", i, got, r[1:])
		}
	}
}

func TestSpecLenientMonthWords(t *testing.T) {
	rows := tableRows(specSection(t, "PARSING.md", 3))
	if len(rows) != 12 {
		t.Fatalf("PARSING §3: %d rows, want 12", len(rows))
	}
	want := map[string]Month{}
	for i, r := range rows {
		for _, tok := range strings.Split(r[1], ",") {
			want[strings.TrimSpace(tok)] = Month(i + 1)
		}
	}
	got := map[string]Month{}
	for _, w := range monthWords {
		if _, dup := got[w.name]; dup {
			t.Errorf("monthWords lists %q twice", w.name)
		}
		got[w.name] = w.month
	}
	for tok, m := range want {
		if got[tok] != m {
			t.Errorf("spec maps %q to %v, code maps it to %v", tok, m, got[tok])
		}
	}
	for tok := range got {
		if _, ok := want[tok]; !ok {
			t.Errorf("code accepts %q, which PARSING §3 doesn't list", tok)
		}
	}
}

func TestSpecLenientIgnoredWords(t *testing.T) {
	text := strings.Join(specSection(t, "PARSING.md", 2), "\n")
	quoted := regexp.MustCompile("`([^`]+)`")
	between := func(from, to string) []string {
		start := strings.Index(text, from)
		if start < 0 {
			t.Fatalf("PARSING §2: can't find %q", from)
		}
		end := strings.Index(text[start+1:], to)
		if end < 0 {
			t.Fatalf("PARSING §2: nothing ends the list after %q", from)
		}
		matches := quoted.FindAllStringSubmatch(text[start:start+1+end], -1)
		out := make([]string, 0, len(matches))
		for _, m := range matches {
			out = append(out, strings.ToLower(m[1]))
		}
		return out
	}
	want := append(between("era markers:", "\n   - Drop tokens `"), between("Drop tokens `गते`", "\n")...)
	got := ignoredWords[:]
	slices.Sort(want)
	sorted := slices.Clone(got)
	slices.Sort(sorted)
	if !slices.Equal(sorted, want) {
		t.Errorf("ignoredWords %q, PARSING §2 lists %q", sorted, want)
	}
}

func TestFoldEqual(t *testing.T) {
	cases := []struct {
		token, entry string
		want         bool
	}{
		{"SHRAWAN", "shrawan", true},
		{"ShRaWaN", "shrawan", true},
		{"Kartik", "kartik", true},    // Kelvin sign folds to k
		{"ſhrawan", "shrawan", false}, // long s does not fold
		{"ı", "i", false},             // dotless i does not fold
		{"Bſ", "bs", false},
		{"shrawa", "shrawan", false},
		{"shrawann", "shrawan", false},
		{"", "", true},
		{"\xe2\x84", "k", false}, // a cut-off Kelvin sign is two other bytes
		{"श्रावण", "श्रावण", true},
	}
	for _, c := range cases {
		if got := foldEqual(c.token, c.entry); got != c.want {
			t.Errorf("foldEqual(%q, %q) = %v, want %v", c.token, c.entry, got, c.want)
		}
	}
}
