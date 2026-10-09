package nepdate

import "testing"

func TestSerialEnds(t *testing.T) {
	if got := toSerial(Min); got != 0 {
		t.Errorf("toSerial(Min) = %d, want 0", got)
	}
	if got := toSerial(Max); got != maxSerial {
		t.Errorf("toSerial(Max) = %d, want %d", got, maxSerial)
	}
	if got := fromSerial(0); got != Min {
		t.Errorf("fromSerial(0) = %v, want Min", got)
	}
	if got := fromSerial(maxSerial); got != Max {
		t.Errorf("fromSerial(maxSerial) = %v, want Max", got)
	}
	// BS 2081-01-01 is serial 65746 (G1-01).
	if got := fromSerial(65746); got != MustNew(2081, Baishakh, 1) {
		t.Errorf("fromSerial(65746) = %v, want 2081-01-01", got)
	}
}

func TestSerialMonthEdges(t *testing.T) {
	// The last and first day of every month: the estimate's one-step correction matters here.
	for k := range len(bsMonthStart) - 1 {
		first, last := bsMonthStart[k], bsMonthStart[k+1]-1
		y, m := minYear+k/12, Month(k%12+1)
		if got := fromSerial(first); got != MustNew(y, m, 1) {
			t.Fatalf("fromSerial(%d) = %v, want %d-%d-01", first, got, y, m)
		}
		if got := fromSerial(last); got != MustNew(y, m, int(last-first)+1) {
			t.Fatalf("fromSerial(%d) = %v, want the last day of %d-%d", last, got, y, m)
		}
		if toSerial(MustNew(y, m, 1)) != first {
			t.Fatalf("toSerial(%d-%d-01) != %d", y, m, first)
		}
	}
}
