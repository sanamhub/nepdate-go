package nepdate

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestJSON(t *testing.T) {
	d := MustNew(2081, Shrawan, 15)
	b, err := json.Marshal(struct{ D Date }{d})
	if err != nil || string(b) != `{"D":"2081-04-15"}` {
		t.Fatalf("json.Marshal = %s, %v", b, err)
	}
	var back struct{ D Date }
	if err := json.Unmarshal(b, &back); err != nil || back.D != d {
		t.Fatalf("json.Unmarshal = %v, %v", back.D, err)
	}
	m := map[Date]int{d: 1, Min: 2, Max: 3}
	b, err = json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var mback map[Date]int
	if err := json.Unmarshal(b, &mback); err != nil || len(mback) != 3 || mback[d] != 1 || mback[Max] != 3 {
		t.Fatalf("map round trip: %s -> %v, %v", b, mback, err)
	}
}

func TestUnmarshalText(t *testing.T) {
	d := MustNew(2081, Shrawan, 15)
	if err := d.UnmarshalText([]byte("2081-04-33")); !errors.Is(err, ErrInvalidDay) || d != MustNew(2081, Shrawan, 15) {
		t.Errorf("UnmarshalText(2081-04-33) = %v, date %v", err, d)
	}
	var pe *ParseError
	if err := d.UnmarshalText([]byte("2081-04")); !errors.As(err, &pe) || pe.Kind != WrongGroupCount || pe.Input != "2081-04" {
		t.Errorf("UnmarshalText(2081-04) = %v", err)
	}
	if err := d.UnmarshalText([]byte("२०८१-०४-३२")); err != nil || d != MustNew(2081, Shrawan, 32) {
		t.Errorf("UnmarshalText(Devanagari) = %v, %v", d, err)
	}
}

func TestMarshalTextZero(t *testing.T) {
	var z Date
	if b, err := z.MarshalText(); !errors.Is(err, ErrOutOfRange) || b != nil {
		t.Errorf("zero MarshalText = %q, %v", b, err)
	}
	if b, err := z.AppendText([]byte("x")); !errors.Is(err, ErrOutOfRange) || string(b) != "x" {
		t.Errorf("zero AppendText = %q, %v", b, err)
	}
	if b, err := MustNew(2081, 1, 1).AppendText([]byte("x")); err != nil || string(b) != "x2081-01-01" {
		t.Errorf("AppendText = %q, %v", b, err)
	}
}

func FuzzUnmarshalText(f *testing.F) {
	seedParse(f)
	f.Fuzz(func(t *testing.T, s string) {
		var d Date
		err := d.UnmarshalText([]byte(s))
		checkParseResult(t, d, err)
		if err == nil {
			text, err := d.MarshalText()
			var back Date
			if err != nil || back.UnmarshalText(text) != nil || back != d {
				t.Fatalf("%v does not round-trip through text: %q", d, text)
			}
		}
	})
}
