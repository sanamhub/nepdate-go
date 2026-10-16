package benchmarks

import (
	"fmt"
	"testing"
	"time"

	bsdate "github.com/JankariTech/GoBikramSambat"
	dateConverter "github.com/opensource-nepal/go-nepali/dateConverter"
	nepdate "github.com/sanamhub/nepdate-go"
	"github.com/sanamhub/nepdate-go/patro"
	nepcal "github.com/srishanbhattarai/nepcal/nepcal"
	bs "github.com/suprimkhatri77/go-bs"
)

// sink receives the accumulator after each benchmark loop, so the compiler can't remove the work.
var sink int

// nsPerDate reports the elapsed time per single input (one benchmark iteration covers 64 inputs).
func nsPerDate(b *testing.B) {
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/64, "ns/date")
}

func BenchmarkB1(b *testing.B) {
	b.Run("library=nepdate", func(b *testing.B) {
		b.ReportAllocs()
		acc := 0
		for b.Loop() {
			acc = 0
			for i := range benchDates {
				d, _ := nepdate.New(benchY[i], nepdate.Month(benchM[i]), benchD[i])
				y, m, day := d.Gregorian()
				acc += y + int(m) + day
			}
		}
		sink = acc
		nsPerDate(b)
	})
	b.Run("library=gonepali", func(b *testing.B) {
		b.ReportAllocs()
		acc := 0
		for b.Loop() {
			acc = 0
			for i := range benchDates {
				a, _ := dateConverter.NepaliToEnglish(benchY[i], benchM[i], benchD[i])
				acc += a[0] + a[1] + a[2]
			}
		}
		sink = acc
		nsPerDate(b)
	})
	b.Run("library=nepcal", func(b *testing.B) {
		b.ReportAllocs()
		acc := 0
		for b.Loop() {
			acc = 0
			for i := range benchDates {
				t, _ := nepcal.Date(benchY[i], nepcal.Month(benchM[i]), benchD[i])
				g := t.Gregorian()
				acc += g.Year() + int(g.Month()) + g.Day()
			}
		}
		sink = acc
		nsPerDate(b)
	})
	b.Run("library=gobs", func(b *testing.B) {
		b.ReportAllocs()
		acc := 0
		for b.Loop() {
			acc = 0
			for i := range benchDates {
				t, _ := bs.BSToAD(bs.Date{Year: benchY[i], Month: benchM[i], Day: benchD[i]})
				acc += t.Year() + int(t.Month()) + t.Day()
			}
		}
		sink = acc
		nsPerDate(b)
	})
	b.Run("library=gobikramsambat", func(b *testing.B) {
		b.ReportAllocs()
		acc := 0
		for b.Loop() {
			acc = 0
			for i := range benchDates {
				d, _ := bsdate.New(benchD[i], benchM[i], benchY[i])
				g, _ := d.GetGregorianDate()
				acc += g.Year() + int(g.Month()) + g.Day()
			}
		}
		sink = acc
		nsPerDate(b)
	})
}

func BenchmarkB2(b *testing.B) {
	b.Run("library=nepdate", func(b *testing.B) {
		b.ReportAllocs()
		acc := 0
		for b.Loop() {
			acc = 0
			for i := range benchDates {
				d, _ := nepdate.FromGregorian(benchADY[i], time.Month(benchADM[i]), benchADD[i])
				acc += d.Year() + int(d.Month()) + d.Day()
			}
		}
		sink = acc
		nsPerDate(b)
	})
	b.Run("library=gonepali", func(b *testing.B) {
		b.ReportAllocs()
		acc := 0
		for b.Loop() {
			acc = 0
			for i := range benchDates {
				// go-nepali's AD range starts at 1944, one day after the first input's AD date.
				if a, err := dateConverter.EnglishToNepali(benchADY[i], benchADM[i], benchADD[i]); err == nil {
					acc += a[0] + a[1] + a[2]
				}
			}
		}
		sink = acc
		nsPerDate(b)
	})
	b.Run("library=nepcal", func(b *testing.B) {
		b.ReportAllocs()
		acc := 0
		for b.Loop() {
			acc = 0
			for i := range benchDates {
				t, _ := nepcal.FromGregorian(benchAD[i])
				acc += t.Year() + int(t.Month()) + t.Day()
			}
		}
		sink = acc
		nsPerDate(b)
	})
	b.Run("library=gobs", func(b *testing.B) {
		b.ReportAllocs()
		acc := 0
		for b.Loop() {
			acc = 0
			for i := range benchDates {
				d, _ := bs.ADToBS(benchAD[i])
				acc += d.Year + d.Month + d.Day
			}
		}
		sink = acc
		nsPerDate(b)
	})
	b.Run("library=gobikramsambat", func(b *testing.B) {
		b.ReportAllocs()
		acc := 0
		for b.Loop() {
			acc = 0
			for i := range benchDates {
				d, _ := bsdate.NewFromGregorian(benchADD[i], benchADM[i], benchADY[i])
				acc += d.GetYear() + d.GetMonth() + d.GetDay()
			}
		}
		sink = acc
		nsPerDate(b)
	})
}

func BenchmarkB3(b *testing.B) {
	b.Run("library=nepdate", func(b *testing.B) {
		b.ReportAllocs()
		acc := 0
		for b.Loop() {
			acc = 0
			for i := range benchDates {
				d, _ := benchDates[i].AddDays(1000)
				acc += d.Year() + int(d.Month()) + d.Day()
			}
		}
		sink = acc
		nsPerDate(b)
	})
	// These two have no O(1) add-days, so they convert to AD, add 1000 days, convert back.
	b.Run("library=gonepali", func(b *testing.B) {
		b.ReportAllocs()
		acc := 0
		for b.Loop() {
			acc = 0
			for i := range benchDates {
				a, _ := dateConverter.NepaliToEnglish(benchY[i], benchM[i], benchD[i])
				t := time.Date(a[0], time.Month(a[1]), a[2], 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1000)
				a, _ = dateConverter.EnglishToNepali(t.Year(), int(t.Month()), t.Day())
				acc += a[0] + a[1] + a[2]
			}
		}
		sink = acc
		nsPerDate(b)
	})
	b.Run("library=gobs", func(b *testing.B) {
		b.ReportAllocs()
		acc := 0
		for b.Loop() {
			acc = 0
			for i := range benchDates {
				t, _ := bs.BSToAD(bs.Date{Year: benchY[i], Month: benchM[i], Day: benchD[i]})
				d, _ := bs.ADToBS(t.AddDate(0, 0, 1000))
				acc += d.Year + d.Month + d.Day
			}
		}
		sink = acc
		nsPerDate(b)
	})
}

func BenchmarkB4(b *testing.B) {
	b.Run("library=nepdate", func(b *testing.B) {
		b.ReportAllocs()
		acc := 0
		for b.Loop() {
			acc = 0
			for i := range benchDates {
				d, _ := benchDates[i].AddMonths(13, nepdate.Clamp)
				acc += d.Year() + int(d.Month()) + d.Day()
			}
		}
		sink = acc
		nsPerDate(b)
	})
}

func BenchmarkB5(b *testing.B) {
	opts := nepdate.LongOptions{Weekday: true}
	b.Run("library=nepdate", func(b *testing.B) {
		b.ReportAllocs()
		buf := make([]byte, 0, 128)
		acc := 0
		for b.Loop() {
			acc = 0
			for i := range benchDates {
				buf = benchDates[i].AppendLong(buf[:0], opts, nepdate.Nepali)
				acc += len(buf)
			}
		}
		sink = acc
		nsPerDate(b)
	})
	b.Run("library=nepdate-string", func(b *testing.B) {
		b.ReportAllocs()
		acc := 0
		for b.Loop() {
			acc = 0
			for i := range benchDates {
				acc += len(benchDates[i].Long(opts, nepdate.Nepali))
			}
		}
		sink = acc
		nsPerDate(b)
	})
}

func BenchmarkB6(b *testing.B) {
	b.Run("library=nepdate", func(b *testing.B) {
		b.ReportAllocs()
		acc := 0
		for b.Loop() {
			acc = 0
			for i := range benchDates {
				d, _ := nepdate.Parse(benchText[i])
				acc += d.Year() + int(d.Month()) + d.Day()
			}
		}
		sink = acc
		nsPerDate(b)
	})
}

func BenchmarkB7(b *testing.B) {
	b.Run("library=nepdate", func(b *testing.B) {
		b.ReportAllocs()
		acc := 0
		for b.Loop() {
			acc = 0
			for i := range patroDates {
				info := patro.Of(patroDates[i])
				if t, ok := info.Tithi(); ok {
					acc += len(t.Name(nepdate.English))
				}
				if info.IsPublicHoliday() {
					acc++
				}
				for n := 0; n < info.EventCount(); n++ {
					acc += len(info.Event(n).Name(nepdate.English))
				}
			}
		}
		sink = acc
		nsPerDate(b)
	})
}

func BenchmarkB8(b *testing.B) {
	b.Run("library=nepdate", func(b *testing.B) {
		b.ReportAllocs()
		acc := 0
		for b.Loop() {
			acc = 0
			for i := range patroDates {
				r, _ := nepdate.MonthRange(patroDates[i].Year(), patroDates[i].Month())
				for day := range r.Days() {
					info := patro.Of(day)
					if info.IsPublicHoliday() {
						acc++
					}
					for n := 0; n < info.EventCount(); n++ {
						acc += len(info.Event(n).Name(nepdate.English))
					}
				}
			}
		}
		sink = acc
		nsPerDate(b)
	})
}

// agreeWords renders the agrees column value for n disagreeing inputs out of 64.
func agreeWords(n int) string {
	if n == 0 {
		return "yes"
	}
	return fmt.Sprintf("no (%d of 64)", n)
}

// TestAgrees compares each competitor's B1 and B2 on the 64 inputs with nepdate and logs the
// agrees column for docs/benchmarks.md. Disagreement is expected for the differing years listed
// in docs/competitors.md, so this test always passes; it fails only if a competitor crashes.
func TestAgrees(t *testing.T) {
	type conversion struct {
		b1, b2 func(i int) (int, int, int)
	}
	libs := map[string]conversion{
		"gonepali": {
			b1: func(i int) (int, int, int) {
				a, _ := dateConverter.NepaliToEnglish(benchY[i], benchM[i], benchD[i])
				return a[0], a[1], a[2]
			},
			b2: func(i int) (int, int, int) {
				a, err := dateConverter.EnglishToNepali(benchADY[i], benchADM[i], benchADD[i])
				if err != nil || a == nil {
					return -1, -1, -1
				}
				return a[0], a[1], a[2]
			},
		},
		"nepcal": {
			b1: func(i int) (int, int, int) {
				t, _ := nepcal.Date(benchY[i], nepcal.Month(benchM[i]), benchD[i])
				g := t.Gregorian()
				return g.Year(), int(g.Month()), g.Day()
			},
			b2: func(i int) (int, int, int) {
				t, _ := nepcal.FromGregorian(benchAD[i])
				return t.Year(), int(t.Month()), t.Day()
			},
		},
		"gobs": {
			b1: func(i int) (int, int, int) {
				t, _ := bs.BSToAD(bs.Date{Year: benchY[i], Month: benchM[i], Day: benchD[i]})
				return t.Year(), int(t.Month()), t.Day()
			},
			b2: func(i int) (int, int, int) {
				d, _ := bs.ADToBS(benchAD[i])
				return d.Year, d.Month, d.Day
			},
		},
		"gobikramsambat": {
			b1: func(i int) (int, int, int) {
				d, _ := bsdate.New(benchD[i], benchM[i], benchY[i])
				g, _ := d.GetGregorianDate()
				return g.Year(), int(g.Month()), g.Day()
			},
			b2: func(i int) (int, int, int) {
				d, _ := bsdate.NewFromGregorian(benchADD[i], benchADM[i], benchADY[i])
				return d.GetYear(), d.GetMonth(), d.GetDay()
			},
		},
	}
	for name, c := range libs {
		b1, b2 := 0, 0
		for i := range benchDates {
			if y, m, d := c.b1(i); y != benchADY[i] || m != benchADM[i] || d != benchADD[i] {
				b1++
			}
			if y, m, d := c.b2(i); y != benchY[i] || m != benchM[i] || d != benchD[i] {
				b2++
			}
		}
		t.Logf("agree %s B1: %s", name, agreeWords(b1))
		t.Logf("agree %s B2: %s", name, agreeWords(b2))
	}
}
