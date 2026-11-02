package nepdate

import (
	"sync"
	"time"
)

// FromTime returns the BS date of t's calendar date in t's own location. FromTime(t) and
// FromTime(t.UTC()) can differ by a day: Nepal is UTC+05:45, so 2024-07-30 18:30 UTC is already
// 2024-07-31 in Kathmandu (BS 2081-04-16, against 2081-04-15 in UTC). Convert t with In first
// to get the date somewhere else. It returns ErrOutOfRange outside AD 1844-04-11..2143-04-15.
// It does not allocate.
func FromTime(t time.Time) (Date, error) {
	return FromGregorian(t.Date())
}

// Time returns midnight at the start of d in loc. A nil loc means Asia/Kathmandu, the same
// default as Today, or the fixed zone NPT (UTC+05:45) when the system has no time zone database.
// Pass time.UTC for midnight UTC. The zero Date returns the zero time.Time.
func (d Date) Time(loc *time.Location) time.Time {
	if d.IsZero() {
		return time.Time{}
	}
	if loc == nil {
		loc = kathmandu()
	}
	y, m, day := d.Gregorian()
	return time.Date(y, m, day, 0, 0, 0, 0, loc)
}

// Today returns today's BS date in loc. A nil loc means Asia/Kathmandu, or the fixed zone
// NPT (UTC+05:45) when the system has no time zone database; the library doesn't embed one.
// A server clock in UTC is 5 h 45 min behind Nepal, so pass the zone you mean. It returns
// ErrOutOfRange if the clock is outside the supported range.
func Today(loc *time.Location) (Date, error) {
	return todayAt(time.Now(), loc)
}

// todayAt is Today with the clock supplied, so the relative and current range helpers have a
// deterministic test path.
func todayAt(now time.Time, loc *time.Location) (Date, error) {
	if loc == nil {
		loc = kathmandu()
	}
	return FromTime(now.In(loc))
}

// IsToday reports whether d is today in loc. A nil loc means Asia/Kathmandu, like Today. It
// returns false for the zero Date and when the clock is outside the supported range.
func (d Date) IsToday(loc *time.Location) bool {
	return d.isTodayAt(time.Now(), loc)
}

func (d Date) isTodayAt(now time.Time, loc *time.Location) bool {
	today, err := todayAt(now, loc)
	return err == nil && d == today
}

// IsYesterday reports whether d is the day before today in loc. A nil loc means Asia/Kathmandu,
// like Today. It returns false for the zero Date and when the clock is outside the supported
// range.
func (d Date) IsYesterday(loc *time.Location) bool {
	return d.isYesterdayAt(time.Now(), loc)
}

func (d Date) isYesterdayAt(now time.Time, loc *time.Location) bool {
	today, err := todayAt(now, loc)
	if err != nil {
		return false
	}
	yesterday, ok := today.Prev()
	return ok && d == yesterday
}

// IsTomorrow reports whether d is the day after today in loc. A nil loc means Asia/Kathmandu,
// like Today. It returns false for the zero Date and when the clock is outside the supported
// range.
func (d Date) IsTomorrow(loc *time.Location) bool {
	return d.isTomorrowAt(time.Now(), loc)
}

func (d Date) isTomorrowAt(now time.Time, loc *time.Location) bool {
	today, err := todayAt(now, loc)
	if err != nil {
		return false
	}
	tomorrow, ok := today.Next()
	return ok && d == tomorrow
}

// kathmandu loads the zone once: time.LoadLocation reads the zone database and allocates on
// every call (ADR-0002 §6, the one sync.OnceValue of the package).
var kathmandu = sync.OnceValue(func() *time.Location {
	if loc, err := time.LoadLocation("Asia/Kathmandu"); err == nil {
		return loc
	}
	return time.FixedZone("NPT", 5*3600+45*60)
})
