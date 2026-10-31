package benchmarks

import (
	"testing"
	"time"

	nepdate "github.com/sanamhub/nepdate-go"
	"github.com/sanamhub/nepdate-go/patro"
)

// The inputs every benchmark runs over, from shared/spec/BENCHMARK.md sections 2 and 3. They are
// built once, outside any timed loop, through the public API.

var (
	benchDates [64]nepdate.Date // the 64 BS dates, 2000-01-01 to 2090-11-24, step 527 serials
	benchY     [64]int
	benchM     [64]int
	benchD     [64]int
	benchADY   [64]int // their AD equivalents
	benchADM   [64]int
	benchADD   [64]int
	benchAD    [64]time.Time
	benchText  [64]string       // their String() texts, "YYYY/MM/DD"
	patroDates [64]nepdate.Date // the 64 patro dates over the coverage, step 44 serials
)

func init() {
	d0, err := nepdate.New(2000, nepdate.Baishakh, 1)
	if err != nil {
		panic(err)
	}
	d1, err := nepdate.New(2090, nepdate.Chaitra, 30)
	if err != nil {
		panic(err)
	}
	s0 := nepdate.Min.DaysUntil(d0)
	s1 := nepdate.Min.DaysUntil(d1)
	step := (s1 - s0) / 63
	for i := range benchDates {
		d, err := nepdate.Min.AddDays(s0 + i*step)
		if err != nil {
			panic(err)
		}
		benchDates[i] = d
		benchY[i], benchM[i], benchD[i] = d.Year(), int(d.Month()), d.Day()
		y, m, day := d.Gregorian()
		benchADY[i], benchADM[i], benchADD[i] = y, int(m), day
		benchAD[i] = time.Date(y, m, day, 0, 0, 0, 0, time.UTC)
		benchText[i] = d.String()
	}
	// The patro coverage, from the generated tables (shared/spec/PATRO.md section 2).
	cov := patro.Coverage()
	c0 := nepdate.Min.DaysUntil(cov.Start())
	c1 := nepdate.Min.DaysUntil(cov.End())
	for i := range patroDates {
		d, err := nepdate.Min.AddDays(c0 + i*((c1-c0)/63))
		if err != nil {
			panic(err)
		}
		patroDates[i] = d
	}
}

// TestInputs checks the four facts BENCHMARK section 2 and 3 state about the inputs.
func TestInputs(t *testing.T) {
	d0, _ := nepdate.New(2000, nepdate.Baishakh, 1)
	d63, _ := nepdate.New(2090, nepdate.Falgun, 24)
	if benchDates[0] != d0 || benchDates[63] != d63 {
		t.Fatalf("benchDates endpoints: got %s and %s, want %s and %s", benchDates[0], benchDates[63], d0, d63)
	}
	for i := 0; i < len(benchDates)-1; i++ {
		if got := benchDates[i].DaysUntil(benchDates[i+1]); got != 527 {
			t.Fatalf("step between index %d and %d is %d, want 527", i, i+1, got)
		}
	}
	for i := range benchDates {
		d, err := nepdate.FromGregorian(benchADY[i], time.Month(benchADM[i]), benchADD[i])
		if err != nil || d != benchDates[i] {
			t.Fatalf("AD equivalent %d: FromGregorian(%d, %d, %d) = %s, %v; want %s", i, benchADY[i], benchADM[i], benchADD[i], d, err, benchDates[i])
		}
		d2, err := nepdate.Parse(benchText[i])
		if err != nil || d2 != benchDates[i] {
			t.Fatalf("text %d: Parse(%q) = %s, %v; want %s", i, benchText[i], d2, err, benchDates[i])
		}
	}
	p0 := patro.Coverage().Start()
	if patroDates[0] != p0 {
		t.Fatalf("patroDates[0] = %s, want %s", patroDates[0], p0)
	}
	for i := 0; i < len(patroDates)-1; i++ {
		if got := patroDates[i].DaysUntil(patroDates[i+1]); got != 44 {
			t.Fatalf("patro step between index %d and %d is %d, want 44", i, i+1, got)
		}
	}
	cov := patro.Coverage()
	for i, d := range patroDates {
		if !cov.Contains(d) {
			t.Fatalf("patroDates[%d] = %s is outside patro coverage %s..%s", i, d, cov.Start(), cov.End())
		}
	}
}
