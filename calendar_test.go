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
	if got, want := int(adMonthStart[3600]), daysFromCivilRef(2144, 1, 1)-epochUnixDays; got != want {
		t.Errorf("adMonthStart[3600] = %d, want %d", got, want)
	}
	if maxSerial != 109211 || minYear != 1901 || maxYear != 2199 || epochUnixDays != -45920 {
		t.Error("calendar constants changed")
	}
}

// daysFromCivilRef is ALGORITHM §5 days_from_civil, used here until reference_test.go (G1-06)
// holds the full oracle.
func daysFromCivilRef(y, m, d int) int {
	if m <= 2 {
		y--
	}
	era := y / 400
	yoe := y - era*400
	mp := (m + 9) % 12
	doy := (153*mp+2)/5 + d - 1
	doe := yoe*365 + yoe/4 - yoe/100 + doy
	return era*146097 + doe - 719468
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
