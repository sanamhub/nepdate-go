package nepdate

import (
	"errors"
	"testing"
)

// checkParseResult holds for every input: no panic, an error is a *ParseError or one of the
// validation sentinels, and a parsed date formats and parses back to itself.
func checkParseResult(t *testing.T, d Date, err error) {
	t.Helper()
	if err != nil {
		var pe *ParseError
		if !errors.As(err, &pe) && !errors.Is(err, ErrOutOfRange) && !errors.Is(err, ErrInvalidMonth) && !errors.Is(err, ErrInvalidDay) {
			t.Fatalf("unexpected error type %T: %v", err, err)
		}
		if !d.IsZero() {
			t.Fatalf("error %v with date %v", err, d)
		}
		return
	}
	back, err := Parse(d.String())
	if err != nil || back != d {
		t.Fatalf("%v does not round-trip: %v, %v", d, back, err)
	}
}

func seedParse(f *testing.F) {
	for _, row := range readTable(f, "shared/spec/vectors/parse.tsv", "\t", 67) {
		f.Add(row["input"])
	}
}

func FuzzParse(f *testing.F) {
	seedParse(f)
	f.Fuzz(func(t *testing.T, s string) {
		d, err := Parse(s)
		checkParseResult(t, d, err)
	})
}

func FuzzParseLenient(f *testing.F) {
	seedParse(f)
	f.Fuzz(func(t *testing.T, s string) {
		d, err := ParseLenient(s)
		checkParseResult(t, d, err)
	})
}
