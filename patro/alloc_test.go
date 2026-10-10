package patro

import (
	"testing"

	nepdate "github.com/sanamhub/nepdate-go"
)

var (
	sinkInfo   Info
	sinkTithi  Tithi
	sinkBool   bool
	sinkString string
	sinkRange  nepdate.Range
	sinkDate   nepdate.Date
)

// TestAlloc checks that the patro paths of ADR-0002 §5 allocate nothing.
func TestAlloc(t *testing.T) {
	d := nepdate.MustNew(2081, nepdate.Shrawan, 15)
	info := Of(d)
	cases := []struct {
		name string
		f    func()
	}{
		{"Of", func() { sinkInfo = Of(d) }},
		{"TithiName", func() {
			sinkTithi, sinkBool = info.Tithi()
			sinkString = sinkTithi.Name(nepdate.English)
			sinkString = sinkTithi.Name(nepdate.Nepali)
		}},
		{"IsPublicHoliday", func() { sinkBool = info.IsPublicHoliday() }},
		{"RangeEvents", func() {
			for e := range info.Events() {
				sinkString = e.Name(nepdate.English)
				sinkString = e.Name(nepdate.Nepali)
			}
		}},
		{"EventLoop", func() {
			for n := range info.EventCount() {
				sinkString = info.Event(n).Name(nepdate.Nepali)
			}
		}},
		{"Coverage", func() { sinkRange = Coverage() }},
		{"PublicHolidays", func() {
			for h := range PublicHolidays() {
				sinkDate = h
			}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if n := testing.AllocsPerRun(100, c.f); n != 0 {
				t.Errorf("%s allocates %.0f times per call, want 0", c.name, n)
			}
		})
	}
}
