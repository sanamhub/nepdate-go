package nepdate

import (
	"testing"
	"time"
)

// TestExhaustive compares the fast path with the ALGORITHM §4 and §5 reference for every day
// of the range.
func TestExhaustive(t *testing.T) {
	for s := 0; s <= maxSerial; s++ {
		ry, rm, rd := refFromSerial(s)
		d := fromSerial(int32(s))
		if d.Year() != ry || int(d.Month()) != rm || d.Day() != rd {
			t.Fatalf("fromSerial(%d) = %d-%d-%d, reference %d-%d-%d", s, d.Year(), d.Month(), d.Day(), ry, rm, rd)
		}
		if got := toSerial(d); int(got) != s || refToSerial(ry, rm, rd) != s {
			t.Fatalf("toSerial(%v) = %d, want %d", d, got, s)
		}
		ay, am, ad := refCivilFromDays(s + refEpochUnixDays)
		gy, gm, gd := d.Gregorian()
		if gy != ay || int(gm) != am || gd != ad {
			t.Fatalf("serial %d: Gregorian() = %d-%d-%d, reference %d-%d-%d", s, gy, gm, gd, ay, am, ad)
		}
		back, err := FromGregorian(ay, time.Month(am), ad)
		if err != nil || back != d {
			t.Fatalf("FromGregorian(%d-%d-%d) = %v, %v; want %v", ay, am, ad, back, err, d)
		}
		if got := int(d.Weekday()); got != (s+4)%7 {
			t.Fatalf("serial %d: Weekday() = %d, want %d", s, got, (s+4)%7)
		}
	}
}
