package nepdate

// The literal algorithm of shared/spec/ALGORITHM.md, as the oracle for the fast path. Each
// function follows its pseudocode line for line; the inputs come straight from
// shared/data/calendar/bs-calendar.json, not from the generated tables.

import (
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"testing"
)

const (
	refMinYear = 1901
	refMaxYear = 2199
	refYears   = 299
)

type refTables struct {
	packed    [refYears]uint32
	yearStart [refYears + 1]int
}

// ALGORITHM §1 and §2: LEN from the JSON, then PACKED and YEAR_START.
var loadRef = sync.OnceValue(func() *refTables {
	raw, err := os.ReadFile("shared/data/calendar/bs-calendar.json")
	if err != nil {
		panic(err)
	}
	var file struct {
		MonthLengths map[string][12]int `json:"month_lengths"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		panic(err)
	}
	t := &refTables{}
	for i := range refYears {
		lens := file.MonthLengths[strconv.Itoa(refMinYear+i)]
		sum := 0
		for m := 1; m <= 12; m++ {
			t.packed[i] += uint32(lens[m-1]-29) << (2 * (m - 1))
			sum += lens[m-1]
		}
		t.yearStart[i+1] = t.yearStart[i] + sum
	}
	return t
})

// ALGORITHM §2: month_length(y, m) = 29 + ((PACKED[y - MIN_YEAR] >> (2*(m-1))) & 3)
func refMonthLength(y, m int) int {
	return 29 + int((loadRef().packed[y-refMinYear]>>(2*(m-1)))&3)
}

// ALGORITHM §4: to_serial
func refToSerial(y, m, d int) int {
	s := loadRef().yearStart[y-refMinYear]
	for k := 1; k <= m-1; k++ {
		s += refMonthLength(y, k)
	}
	return s + (d - 1)
}

// ALGORITHM §4: from_serial (precondition 0 <= s <= MAX_SERIAL)
func refFromSerial(s int) (y, m, d int) {
	ys := loadRef().yearStart[:]
	// i = largest i such that YEAR_START[i] <= s, by binary search over YEAR_START[0..YEARS]
	lo, hi := 0, refYears
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if ys[mid] <= s {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	i := lo
	y = refMinYear + i
	r := s - ys[i] // 0-based day of year
	m = 1
	for r >= refMonthLength(y, m) {
		r -= refMonthLength(y, m)
		m++
	}
	return y, m, r + 1
}

// ALGORITHM §5: days_from_civil (Howard Hinnant), proleptic Gregorian to days since 1970-01-01
func refDaysFromCivil(y, m, d int) int {
	if m <= 2 {
		y--
	}
	era := y / 400              // y >= 0 in our range
	yoe := y - era*400          // [0, 399]
	mp := (m + 9) % 12          // March = 0 .. February = 11
	doy := (153*mp+2)/5 + d - 1 // [0, 365]
	doe := yoe*365 + yoe/4 - yoe/100 + doy
	return era*146097 + doe - 719468
}

// ALGORITHM §5: civil_from_days (Howard Hinnant), days since 1970-01-01 to (y, m, d)
func refCivilFromDays(z int) (y, m, d int) {
	z += 719468
	era := z / 146097
	doe := z - era*146097
	yoe := (doe - doe/1460 + doe/36524 - doe/146096) / 365
	y = yoe + era*400
	doy := doe - (365*yoe + yoe/4 - yoe/100)
	mp := (5*doy + 2) / 153
	d = doy - (153*mp+2)/5 + 1
	if mp < 10 {
		m = mp + 3
	} else {
		m = mp - 9
	}
	if m <= 2 {
		y++
	}
	return y, m, d
}

// ALGORITHM §2: EPOCH_UNIX_DAYS = days_from_civil(1844, 4, 11)
var refEpochUnixDays = refDaysFromCivil(1844, 4, 11)

// TestReferenceOracle checks the oracle itself against ALGORITHM §9, so a mistake copying the
// pseudocode can't hide behind a matching mistake in the fast path.
func TestReferenceOracle(t *testing.T) {
	rows := [][6]int{
		{1901, 1, 1, 1844, 4, 11},
		{2081, 1, 1, 2024, 4, 13},
		{2082, 1, 1, 2025, 4, 14},
		{2083, 1, 1, 2026, 4, 14},
		{2199, 12, 30, 2143, 4, 15},
	}
	for _, r := range rows {
		s := refToSerial(r[0], r[1], r[2])
		if y, m, d := refFromSerial(s); y != r[0] || m != r[1] || d != r[2] {
			t.Errorf("refFromSerial(%d) = %d-%d-%d, want %v", s, y, m, d, r[:3])
		}
		if y, m, d := refCivilFromDays(s + refEpochUnixDays); y != r[3] || m != r[4] || d != r[5] {
			t.Errorf("refCivilFromDays for %v = %d-%d-%d, want %v", r[:3], y, m, d, r[3:])
		}
		if got := refDaysFromCivil(r[3], r[4], r[5]) - refEpochUnixDays; got != s {
			t.Errorf("refDaysFromCivil(%v) - epoch = %d, want %d", r[3:], got, s)
		}
	}
	if refEpochUnixDays != -45920 {
		t.Errorf("EPOCH_UNIX_DAYS = %d, want -45920", refEpochUnixDays)
	}
}
