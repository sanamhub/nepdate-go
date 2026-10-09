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
	for _, row := range readTable(f, "shared/spec/vectors/parse.tsv", "\t", 63) {
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

// FuzzFormatRoundTrip formats a date with Short in Y-M-D order and every separator, padding and
// language, and parses it back.
func FuzzFormatRoundTrip(f *testing.F) {
	f.Add(int32(65746+105), uint8(0), uint8(0), true, uint8(0))
	f.Add(int32(0), uint8(0), uint8(5), false, uint8(1))
	f.Add(int32(maxSerial), uint8(0), uint8(1), true, uint8(1))
	f.Fuzz(func(t *testing.T, serial int32, order, sep uint8, pad bool, lang uint8) {
		if serial < 0 || serial > maxSerial {
			return
		}
		d := fromSerial(serial)
		o, sp, l := DateOrder(order%6), Separator(sep%6), Lang(lang%2)
		if o != YMD {
			return
		}
		s := d.Short(o, sp, pad, l)
		back, err := Parse(s)
		if err != nil || back != d {
			t.Fatalf("Parse(%q) = %v, %v; want %v", s, back, err, d)
		}
	})
}
