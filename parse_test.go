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
	return readTable(t, "shared/spec/vectors/parse.tsv", "\t", 57)
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
