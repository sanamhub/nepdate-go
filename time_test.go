package nepdate

import (
	"errors"
	"testing"
	"time"
)

var npt = time.FixedZone("NPT", 5*3600+45*60)

func TestFromTime(t *testing.T) {
	utc := time.Date(2024, time.July, 30, 18, 30, 0, 0, time.UTC)
	if got, err := FromTime(utc); err != nil || got != MustNew(2081, Shrawan, 15) {
		t.Errorf("FromTime(UTC) = %v, %v; want 2081-04-15", got, err)
	}
	// The same instant is already 31 July 00:15 in Nepal.
	if got, err := FromTime(utc.In(npt)); err != nil || got != MustNew(2081, Shrawan, 16) {
		t.Errorf("FromTime(NPT) = %v, %v; want 2081-04-16", got, err)
	}
	if _, err := FromTime(time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC)); !errors.Is(err, ErrOutOfRange) {
		t.Errorf("FromTime(1800) error = %v", err)
	}
	if _, err := FromTime(time.Time{}); !errors.Is(err, ErrOutOfRange) {
		t.Errorf("FromTime(zero time) error = %v", err)
	}
}

func TestTime(t *testing.T) {
	d := MustNew(2081, Shrawan, 15)
	if got := d.Time(nil); !got.Equal(time.Date(2024, time.July, 30, 0, 0, 0, 0, time.UTC)) || got.Location() != time.UTC {
		t.Errorf("Time(nil) = %v", got)
	}
	if got := d.Time(npt); !got.Equal(time.Date(2024, time.July, 30, 0, 0, 0, 0, npt)) {
		t.Errorf("Time(NPT) = %v", got)
	}
	if !(Date{}).Time(nil).IsZero() {
		t.Error("zero Date gives a non-zero time")
	}
	if back, err := FromTime(d.Time(npt)); err != nil || back != d {
		t.Errorf("FromTime(Time(NPT)) = %v, %v", back, err)
	}
}

func TestToday(t *testing.T) {
	for _, loc := range []*time.Location{nil, time.UTC, npt} {
		check := loc
		if check == nil {
			check = kathmandu()
		}
		before, _ := FromTime(time.Now().In(check))
		got, err := Today(loc)
		after, _ := FromTime(time.Now().In(check))
		if err != nil || got != before && got != after {
			t.Errorf("Today(%v) = %v, %v; want %v", loc, got, err, before)
		}
	}
	if name := kathmandu().String(); name != "Asia/Kathmandu" && name != "NPT" {
		t.Errorf("default zone %q", name)
	}
	_, offset := time.Date(2024, 7, 30, 12, 0, 0, 0, kathmandu()).Zone()
	if offset != 20700 {
		t.Errorf("default zone offset %d, want 20700", offset)
	}
}
