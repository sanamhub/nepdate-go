package nepdate

import "testing"

func TestCalendarTables(t *testing.T) {
	if got := bsMonthStart[3588]; got != 109212 {
		t.Errorf("bsMonthStart[3588] = %d, want 109212", got)
	}
	// BS 2081-01-01 is AD 2024-04-13.
	if got := bsMonthStart[(2081-1901)*12]; got != 65746 {
		t.Errorf("bsMonthStart[2081-01] = %d, want 65746", got)
	}
	if got := adMonthStart[0]; got != -101 {
		t.Errorf("adMonthStart[0] = %d, want -101", got)
	}
	// AD 2144-01-01 is 109212 + (days from 2143-04-16 to 2144-01-01) = 109212 + 260.
	if got, want := int(adMonthStart[3600]), refDaysFromCivil(2144, 1, 1)-refEpochUnixDays; got != want {
		t.Errorf("adMonthStart[3600] = %d, want %d", got, want)
	}
	if refEpochUnixDays != epochUnixDays || refToSerial(2199, 12, 30) != maxSerial {
		t.Error("generated constants disagree with the reference")
	}
	if maxSerial != 109211 || minYear != 1901 || maxYear != 2199 || epochUnixDays != -45920 {
		t.Error("calendar constants changed")
	}
}

// TestEstimateConstants repeats the generator's check (ADR-0002 §3): for every serial the
// month estimate is the true month index or one below it.
func TestEstimateConstants(t *testing.T) {
	bk, ak := 0, 3 // serial 0 is in BS month 0 and AD month 3 (1844-04)
	for s := int32(0); s <= maxSerial; s++ {
		for bsMonthStart[bk+1] <= s {
			bk++
		}
		for adMonthStart[ak+1] <= s {
			ak++
		}
		if d := bk - int((int64(s)*bsEstMul+bsEstAdd)>>estShift); d != 0 && d != 1 {
			t.Fatalf("serial %d: BS estimate off by %d", s, d)
		}
		if d := ak - int((int64(s)*adEstMul+adEstAdd)>>estShift); d != 0 && d != 1 {
			t.Fatalf("serial %d: AD estimate off by %d", s, d)
		}
	}
}
