package nepdate

import (
	"sync"
	"testing"
	"time"
)

type concurrentResult struct {
	newDate, fromAD Date
	adY, adD        int
	adM             time.Month
	plus1000        Date
	plus13          Date
	diff            Diff
}

func concurrentRound(dates *[64]Date) [64]concurrentResult {
	var out [64]concurrentResult
	for i, d := range dates {
		r := &out[i]
		r.newDate, _ = New(d.Year(), d.Month(), d.Day())
		r.adY, r.adM, r.adD = d.Gregorian()
		r.fromAD, _ = FromGregorian(r.adY, r.adM, r.adD)
		r.plus1000, _ = d.AddDays(1000)
		r.plus13, _ = d.AddMonths(13, Clamp)
		r.diff = d.Diff(dates[(i+17)%len(dates)])
	}
	return out
}

// TestConcurrentReads runs the conversions from 8 goroutines and compares with one goroutine
// (ADR-0004). Under -race it fails if anything writes shared state.
func TestConcurrentReads(t *testing.T) {
	dates := benchmarkDates()
	want := concurrentRound(&dates)
	var wg sync.WaitGroup
	errs := make(chan string, 8)
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 1000 {
				if concurrentRound(&dates) != want {
					errs <- "a concurrent round differs from the single-goroutine run"
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}
