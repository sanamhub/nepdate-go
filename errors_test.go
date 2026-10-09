package nepdate

import (
	"errors"
	"testing"
)

func TestParseErrorMessage(t *testing.T) {
	cases := []struct {
		kind ParseErrorKind
		want string
	}{
		{InvalidCharacter, `nepdate: cannot parse "2081-0x-15": invalid character`},
		{WrongGroupCount, `nepdate: cannot parse "2081-0x-15": wrong number of groups`},
		{NumberTooLong, `nepdate: cannot parse "2081-0x-15": number longer than 4 digits`},
		{Ambiguous, `nepdate: cannot parse "2081-0x-15": ambiguous`},
		{Unrecognized, `nepdate: cannot parse "2081-0x-15": unrecognized`},
		{ParseErrorKind(0), `nepdate: cannot parse "2081-0x-15"`},
	}
	for _, c := range cases {
		e := &ParseError{Input: "2081-0x-15", Kind: c.kind}
		if got := e.Error(); got != c.want {
			t.Errorf("kind %d: %q, want %q", c.kind, got, c.want)
		}
	}
	var err error = &ParseError{Input: "x", Kind: Ambiguous}
	var pe *ParseError
	if !errors.As(err, &pe) || pe.Kind != Ambiguous {
		t.Error("errors.As does not find *ParseError")
	}
}

func TestSentinelsDistinct(t *testing.T) {
	all := []error{ErrOutOfRange, ErrInvalidMonth, ErrInvalidDay, ErrInvalidGregorian}
	for i, a := range all {
		for j, b := range all {
			if (i == j) != errors.Is(a, b) {
				t.Errorf("errors.Is(%v, %v) = %v", a, b, i != j)
			}
		}
	}
}
