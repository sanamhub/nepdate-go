package nepdate

import (
	"errors"
	"strings"
	"testing"
)

var parseKinds = map[string]ParseErrorKind{
	"InvalidCharacter": InvalidCharacter,
	"WrongGroupCount":  WrongGroupCount,
	"NumberTooLong":    NumberTooLong,
	"Ambiguous":        Ambiguous,
	"Unrecognized":     Unrecognized,
}

var parseSentinels = map[string]error{
	"OutOfRange":   ErrOutOfRange,
	"InvalidMonth": ErrInvalidMonth,
	"InvalidDay":   ErrInvalidDay,
}

// checkParse compares a parse result with a parse.tsv cell: a date, or ERR:<Kind>.
func checkParse(t *testing.T, what, input string, got Date, err error, cell string) {
	t.Helper()
	kind, isErr := strings.CutPrefix(cell, "ERR:")
	if !isErr {
		if want := bsDate(t, cell); err != nil || got != want {
			t.Errorf("%s(%q) = %v, %v; want %s", what, input, got, err, cell)
		}
		return
	}
	if !got.IsZero() {
		t.Errorf("%s(%q) returned %v with an error", what, input, got)
	}
	if k, ok := parseKinds[kind]; ok {
		var pe *ParseError
		if !errors.As(err, &pe) || pe.Kind != k || pe.Input != input {
			t.Errorf("%s(%q) error = %v, want *ParseError %s", what, input, err, kind)
		}
		return
	}
	sentinel, ok := parseSentinels[kind]
	if !ok {
		t.Fatalf("unknown error kind %q in parse.tsv", kind)
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("%s(%q) error = %v, want %v", what, input, err, sentinel)
	}
}

func parseRows(t *testing.T) []map[string]string {
	return readTable(t, "shared/spec/vectors/parse.tsv", "\t", 67)
}

func TestParseStrict(t *testing.T) {
	for _, row := range parseRows(t) {
		got, err := Parse(row["input"])
		checkParse(t, "Parse", row["input"], got, err, row["strict"])
	}
}

func TestParseStrictExtra(t *testing.T) {
	cases := []struct{ input, want string }{
		{"२०८०/०५/१५", "2080-05-15"},
		{"२०80-0५-१5", "2080-05-15"}, // mixed digit systems (D-05)
		{"2080/257/01", "ERR:InvalidMonth"},
		{"9999/9999/9999", "ERR:OutOfRange"},
		{"2080/0/15", "ERR:InvalidMonth"},
		{"2080/05/0", "ERR:InvalidDay"},
		{"0000/05/15", "ERR:OutOfRange"},
		{"2080/05/1५a", "ERR:InvalidCharacter"},
		{"2080/05/15\xe0", "ERR:InvalidCharacter"}, // truncated UTF-8
		{"2080/05/15\xe0\xa5", "ERR:InvalidCharacter"},
		{"2080/05/15\xe0\xa5\xb0", "ERR:InvalidCharacter"}, // U+0970, not a digit
		{"2080/05/15 1", "ERR:WrongGroupCount"},
		{"00002080/05/15", "ERR:NumberTooLong"},
	}
	for _, c := range cases {
		got, err := Parse(c.input)
		checkParse(t, "Parse", c.input, got, err, c.want)
	}
}

func TestParseLenient(t *testing.T) {
	for _, row := range parseRows(t) {
		got, err := ParseLenient(row["input"])
		checkParse(t, "ParseLenient", row["input"], got, err, row["lenient"])
	}
}

func TestParseLenientExtra(t *testing.T) {
	cases := []struct{ input, want string }{
		{"15 Ashwin 2080", "2080-06-15"},
		{"15 वैशाख 2080", "2080-01-15"},
		{"15 जेष्ठ 2080", "2080-02-15"},
		{"15 श्रावण 2080", "2080-04-15"},
		{"Baisakh 15, 2080", "2080-01-15"},
		{"15 \u212Aartik 2080", "2080-07-15"}, // Kelvin sign folds to k, PARSING §2
		{"2080 baisakh 15", "2080-01-15"},
		{"मिति २०८० साउन १५ गते", "2080-04-15"},
		{"2080-04-15 वि.सं", "2080-04-15"},
		{"2080/04/15 v.s.", "2080-04-15"},
		{"15 Shrawan 2080 Bhadra", "ERR:Unrecognized"},
		{"15 Shrawan", "ERR:Unrecognized"},
		{"15 Shrawan 2080 1", "ERR:Unrecognized"},
		{"2080 Shrawan 2081", "ERR:Ambiguous"},
		{"1 2 3 4 5 6 7 8 9", "ERR:Unrecognized"},
		{"15 Shrawan 12345", "ERR:NumberTooLong"},
		{"15/13/2080", "ERR:InvalidMonth"},
		{"32 Bhadra 2081", "ERR:InvalidDay"},
		{"१५.०४.२०८०", "2080-04-15"},
		{"15Shrawan2080", "ERR:Unrecognized"}, // one word (PARSING §2)
		{"   ", "ERR:Unrecognized"},
	}
	for _, c := range cases {
		got, err := ParseLenient(c.input)
		checkParse(t, "ParseLenient", c.input, got, err, c.want)
	}
}

// TestLenientMonthNames checks every spelling of PARSING §3, in three letter cases for Latin.
func TestLenientMonthNames(t *testing.T) {
	names := [12][]string{
		{"baishakh", "baisakh", "baishak", "vaishakh", "vaisakh", "bai", "बैशाख", "वैशाख", "बैसाख"},
		{"jestha", "jeth", "jyestha", "jes", "जेठ", "जेष्ठ", "ज्येष्ठ"},
		{"ashad", "ashadh", "asar", "asadh", "asa", "असार", "आषाढ", "असाढ"},
		{"shrawan", "shravan", "srawan", "sawan", "saun", "shr", "साउन", "श्रावण", "सावन"},
		{"bhadra", "bhadau", "bha", "भदौ", "भाद्र", "भाद्रपद"},
		{"ashoj", "asoj", "ashwin", "aso", "असोज", "आश्विन", "असौज"},
		{"kartik", "kattik", "kar", "कार्तिक", "कात्तिक"},
		{"mangsir", "mansir", "margashirsha", "man", "मंसिर", "मङ्सिर", "मार्गशीर्ष"},
		{"poush", "paush", "push", "pus", "pou", "पुष", "पौष", "पुस"},
		{"magh", "mag", "माघ"},
		{"falgun", "phalgun", "fagun", "fal", "फागुन", "फाल्गुन"},
		{"chaitra", "chait", "cha", "चैत", "चैत्र"},
	}
	for i, list := range names {
		want := MustNew(2080, Month(i+1), 1)
		for _, name := range list {
			spellings := []string{name}
			if name[0] < 0x80 { // Latin: also upper case and title case
				spellings = append(spellings, strings.ToUpper(name), strings.ToUpper(name[:1])+name[1:])
			}
			for _, n := range spellings {
				if got, err := ParseLenient("1 " + n + " 2080"); err != nil || got != want {
					t.Errorf("ParseLenient(%q) = %v, %v; want %v", "1 "+n+" 2080", got, err, want)
				}
			}
		}
	}
}
