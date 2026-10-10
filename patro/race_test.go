package patro

import (
	"sync"
	"testing"

	nepdate "github.com/sanamhub/nepdate-go"
)

// fingerprint reads every name of every day of the coverage into one number.
func fingerprint() int {
	sum := 0
	for d := range Coverage().Days() {
		i := Of(d)
		tithi, _ := i.Tithi()
		sum += len(tithi.Name(nepdate.Nepali))
		if i.IsPublicHoliday() {
			sum++
		}
		for n := range i.EventCount() {
			sum += len(i.Event(n).Name(nepdate.English)) + len(i.Event(n).Name(nepdate.Nepali))
		}
	}
	return sum
}

// TestConcurrentReads runs Of and the name lookups from 8 goroutines over the whole coverage
// and compares with one goroutine (ADR-0004). Under -race it fails if anything writes shared
// state.
func TestConcurrentReads(t *testing.T) {
	want := fingerprint()
	var wg sync.WaitGroup
	errs := make(chan int, 8)
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if got := fingerprint(); got != want {
				errs <- got
			}
		}()
	}
	wg.Wait()
	close(errs)
	for got := range errs {
		t.Errorf("a concurrent run read %d, want %d", got, want)
	}
}
