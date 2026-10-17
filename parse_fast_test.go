package nepdate

import (
	"errors"
	"testing"
)

// strictSeparators are the seven ASCII separators PARSING §1 accepts.
var strictSeparators = [...]byte{'-', '/', '.', '_', '\\', ' ', '|'}

// canonicalText builds the 10-byte YYYY<sep>MM<sep>DD form. The date range has four-digit years
// and two-digit months and days, so every date has exactly this shape.
func canonicalText(d Date, sep byte) [10]byte {
	y, m, day := d.Year(), int(d.Month()), d.Day()
	return [10]byte{
		digitByte(y / 1000),
		digitByte(y / 100 % 10),
		digitByte(y / 10 % 10),
		digitByte(y % 10),
		sep,
		digitByte(m / 10),
		digitByte(m % 10),
		sep,
		digitByte(day / 10),
		digitByte(day % 10),
	}
}

// digitByte returns the ASCII byte of a decimal digit.
func digitByte(n int) byte {
	return byte('0' + n) //nolint:gosec // callers pass 0..9
}

// sameError reports whether two parse errors are the same error. The scanners return package
// sentinels, so errors.Is is exact here; nil matches only nil.
func sameError(a, b error) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return errors.Is(a, b) && errors.Is(b, a)
}

// canonicalShape is the test oracle for the shape the fast path must accept, including invalid
// dates such as 0000/00/00, whose validation error must match the general loop's.
func canonicalShape(s string) bool {
	if len(s) != 10 {
		return false
	}
	if !isStrictSeparator(s[4]) || !isStrictSeparator(s[7]) {
		return false
	}
	for _, i := range [...]int{0, 1, 2, 3, 5, 6, 8, 9} {
		if s[i]-'0' > 9 {
			return false
		}
	}
	return true
}

// TestScanCanonicalMatchesLoop checks every date and every strict separator. It also requires the
// fast path to accept each canonical text; a declined canonical text would silently narrow the
// path and leave ParseLenient on the slower route.
func TestScanCanonicalMatchesLoop(t *testing.T) {
	for s := int32(0); s <= maxSerial; s++ {
		d := fromSerial(s)
		for _, sep := range strictSeparators {
			text := canonicalText(d, sep)
			fast, ok, fastErr := scanCanonical(text[:])
			if !ok {
				t.Fatalf("scanCanonical(%q) declined a canonical text", text[:])
			}
			slow, kind, slowErr := scanStrictLoop(text[:])
			if kind != 0 {
				t.Fatalf("scanStrictLoop(%q) kind = %v, want 0", text[:], kind)
			}
			if fast != slow {
				t.Fatalf("scanCanonical(%q) = %v, scanStrictLoop = %v", text[:], fast, slow)
			}
			if !sameError(fastErr, slowErr) {
				t.Fatalf("scanCanonical(%q) error = %v, scanStrictLoop error = %v", text[:], fastErr, slowErr)
			}
		}
	}
}

// FuzzScanCanonical compares the fast path with the general loop on arbitrary input, including
// invalid dates and non-canonical shapes, and checks that scanStrict dispatches to the same
// result either way.
func FuzzScanCanonical(f *testing.F) {
	for _, s := range []string{
		"2081/04/15", "2081-04-15", "2081.04.15", "2081_04_15", "2081\\04\\15", "2081 04 15", "2081|04|15",
		"2081/09/15", "2081/9/15", "2081/09/5", "2081/04/1", "2081/04/150",
		"0000/00/00", "9999/99/99", "2081/0a/15", "2081/0:/15", "2081/04/1x", "२०८१/०४/१५", "",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		fast, ok, fastErr := scanCanonical(s)
		slow, kind, slowErr := scanStrictLoop(s)
		if canonicalShape(s) && !ok {
			t.Fatalf("scanCanonical(%q) declined a canonical text", s)
		}
		if ok {
			if fast != slow {
				t.Fatalf("scanCanonical(%q) = %v, scanStrictLoop = %v", s, fast, slow)
			}
			if kind != 0 {
				t.Fatalf("scanStrictLoop(%q) kind = %v for a canonical text", s, kind)
			}
			if !sameError(fastErr, slowErr) {
				t.Fatalf("scanCanonical(%q) error = %v, scanStrictLoop error = %v", s, fastErr, slowErr)
			}
		}
		got, gotKind, gotErr := scanStrict(s)
		if got != slow || gotKind != kind || !sameError(gotErr, slowErr) {
			t.Fatalf("scanStrict(%q) = %v, %v, %v; scanStrictLoop = %v, %v, %v",
				s, got, gotKind, gotErr, slow, kind, slowErr)
		}
	})
}
