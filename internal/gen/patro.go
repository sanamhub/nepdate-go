package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
)

// Bit layout of one dayInfo record (ADR-0002 §2).
const (
	tithiMask  = 1<<6 - 1 // bits 0-5: tithi name id + 1, 0 = no tithi
	holidayBit = 1 << 6   // bit 6: public holiday
	countShift = 7        // bits 7-10: number of events
	countMask  = 1<<4 - 1
	firstShift = 11 // bits 11-22: index of the day's first event in eventName
	firstMask  = 1<<12 - 1

	maxTithiNames = tithiMask - 1 // ids 0..62 fit as id+1
	maxDayEvents  = countMask
	maxEvents     = firstMask
)

// namePair is one name in both languages, as in shared/data/patro/*.json.
type namePair struct {
	Np string `json:"np"`
	En string `json:"en"`
}

type patroDay struct {
	Tithi   *namePair  `json:"tithi"`
	Holiday bool       `json:"holiday"`
	Events  []namePair `json:"events"`
}

type patroFile struct {
	Year int                 `json:"year"`
	Days map[string]patroDay `json:"days"`
}

// patroTables is the content of patro/patro_gen.go.
type patroTables struct {
	covFirst, covLast int32
	tithiNames        int
	dayInfo           []uint32
	eventName         []uint16
	pairs             []namePair
	nameOff           []uint32
	nameText          string
}

// patroSummaryPath is the hub's count file, vendored under shared/spec/vectors (PATRO §3). Every
// port reads its counts from here instead of pinning its own.
const patroSummaryPath = "shared/spec/vectors/patro-summary.json"

// patroSummarySchema is the only summary layout this generator reads.
const patroSummarySchema = 1

// patroSummary is shared/spec/vectors/patro-summary.json.
type patroSummary struct {
	Schema           int    `json:"schema"`
	FirstYear        int    `json:"first_year"`
	LastYear         int    `json:"last_year"`
	FirstDay         string `json:"first_day"`
	LastDay          string `json:"last_day"`
	FirstSerial      int32  `json:"first_serial"`
	LastSerial       int32  `json:"last_serial"`
	Days             int    `json:"days"`
	TithiFirstDay    string `json:"tithi_first_day"`
	TithiLastDay     string `json:"tithi_last_day"`
	TithiFirstSerial int32  `json:"tithi_first_serial"`
	TithiDays        int    `json:"tithi_days"`
	TithiNames       int    `json:"tithi_names"`
	Holidays         int    `json:"holidays"`
	EventDays        int    `json:"event_days"`
	Events           int    `json:"events"`
	EventNames       int    `json:"event_names"`
	Names            int    `json:"names"`
	NameBytes        int    `json:"name_bytes"`
}

var patroFileName = regexp.MustCompile(`^[0-9]{4}\.json$`)

// loadPatroSummary reads the vendored counts. A missing file means shared/ predates the summary,
// so it is a resync, not a code change.
func loadPatroSummary(root string) (*patroSummary, error) {
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(patroSummaryPath))) //nolint:gosec // dev tool reading the repository
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%s is missing: shared/ predates the patro summary, resync shared/", patroSummaryPath)
	}
	if err != nil {
		return nil, err
	}
	var s patroSummary
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("%s: %w", patroSummaryPath, err)
	}
	if s.Schema != patroSummarySchema {
		return nil, fmt.Errorf("%s: schema %d, this generator reads schema %d", patroSummaryPath, s.Schema, patroSummarySchema)
	}
	return &s, nil
}

// checkPatroSummary compares every count derived from the data with the vendored summary, naming
// the field and both values on a disagreement (PATRO §3). A stale or half-synced shared/ fails
// here instead of producing a table that disagrees with every other port.
func checkPatroSummary(cal *calendar, p *patroTables, sum *patroSummary) error {
	bs := bsStarts(cal)
	tithiDays, holidays, eventDays, events := 0, 0, 0, 0
	tithiFirst, tithiLast := int32(-1), int32(-1)
	for i, rec := range p.dayInfo {
		if rec&tithiMask != 0 {
			tithiDays++
			s := p.covFirst + int32(i) //nolint:gosec // one record per coverage day
			if tithiFirst < 0 {
				tithiFirst = s
			}
			tithiLast = s
		}
		if rec&holidayBit != 0 {
			holidays++
		}
		if n := int(rec>>countShift) & countMask; n > 0 {
			eventDays++
			events += n
		}
	}
	eventIDs := make(map[uint16]bool, len(p.eventName))
	for _, id := range p.eventName {
		eventIDs[id] = true
	}
	firstDay, lastDay := dateOf(&bs, p.covFirst), dateOf(&bs, p.covLast)
	tithiFirstDay, tithiLastDay := "", ""
	if tithiFirst >= 0 {
		tithiFirstDay, tithiLastDay = dateOf(&bs, tithiFirst), dateOf(&bs, tithiLast)
	}
	checks := []struct {
		field      string
		file, data string
	}{
		{"first_year", strconv.Itoa(sum.FirstYear), firstDay[:4]},
		{"last_year", strconv.Itoa(sum.LastYear), lastDay[:4]},
		{"first_day", sum.FirstDay, firstDay},
		{"last_day", sum.LastDay, lastDay},
		{"first_serial", strconv.FormatInt(int64(sum.FirstSerial), 10), strconv.FormatInt(int64(p.covFirst), 10)},
		{"last_serial", strconv.FormatInt(int64(sum.LastSerial), 10), strconv.FormatInt(int64(p.covLast), 10)},
		{"days", strconv.Itoa(sum.Days), strconv.Itoa(len(p.dayInfo))},
		{"tithi_first_day", sum.TithiFirstDay, tithiFirstDay},
		{"tithi_last_day", sum.TithiLastDay, tithiLastDay},
		{"tithi_first_serial", strconv.FormatInt(int64(sum.TithiFirstSerial), 10), strconv.FormatInt(int64(tithiFirst), 10)},
		{"tithi_days", strconv.Itoa(sum.TithiDays), strconv.Itoa(tithiDays)},
		{"tithi_names", strconv.Itoa(sum.TithiNames), strconv.Itoa(p.tithiNames)},
		{"holidays", strconv.Itoa(sum.Holidays), strconv.Itoa(holidays)},
		{"event_days", strconv.Itoa(sum.EventDays), strconv.Itoa(eventDays)},
		{"events", strconv.Itoa(sum.Events), strconv.Itoa(events)},
		{"event_names", strconv.Itoa(sum.EventNames), strconv.Itoa(len(eventIDs))},
		{"names", strconv.Itoa(sum.Names), strconv.Itoa(len(p.pairs))},
		{"name_bytes", strconv.Itoa(sum.NameBytes), strconv.Itoa(len(p.nameText))},
	}
	for _, c := range checks {
		if c.file != c.data {
			return fmt.Errorf("%s: %s: summary says %s, the data has %s", patroSummaryPath, c.field, c.file, c.data)
		}
	}
	if len(p.eventName) != sum.Events {
		return fmt.Errorf("%s: events: summary says %d, eventName holds %d", patroSummaryPath, sum.Events, len(p.eventName))
	}
	return nil
}

func loadPatroFiles(root string) (map[int]*patroFile, error) {
	dir := filepath.Join(root, "shared", "data", "patro")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := map[int]*patroFile{}
	for _, e := range entries {
		if !patroFileName.MatchString(e.Name()) {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name())) //nolint:gosec // dev tool reading the repository
		if err != nil {
			return nil, err
		}
		var f patroFile
		if err := json.Unmarshal(raw, &f); err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		files[f.Year] = &f
	}
	if len(files) == 0 {
		return nil, errors.New("no patro files in shared/data/patro")
	}
	return files, nil
}

type datedDay struct {
	date   string
	serial int32
	day    patroDay
}

// buildPatro checks the data and lays it out. It fails, naming the limit, when the data would
// overflow a bit field, and when a day inside the tithi range has no tithi (PATRO §3).
func buildPatro(cal *calendar, files map[int]*patroFile) (*patroTables, error) {
	bs := bsStarts(cal)
	var days []datedDay
	for _, f := range files {
		for date, day := range f.Days {
			s, err := serialOf(cal, &bs, date)
			if err != nil {
				return nil, err
			}
			if day.Tithi == nil && !day.Holiday && len(day.Events) == 0 {
				continue
			}
			days = append(days, datedDay{date, s, day})
		}
	}
	if len(days) == 0 {
		return nil, errors.New("patro data has no days")
	}
	slices.SortFunc(days, func(a, b datedDay) int { return int(a.serial - b.serial) })
	p := &patroTables{covFirst: days[0].serial, covLast: days[len(days)-1].serial}

	// Tithi names first, then event names, each in first-appearance order.
	ids := map[namePair]int{}
	intern := func(n namePair) int {
		if id, ok := ids[n]; ok {
			return id
		}
		ids[n] = len(p.pairs)
		p.pairs = append(p.pairs, n)
		return len(p.pairs) - 1
	}
	tithiFirst, tithiLast := int32(-1), int32(-1)
	for _, d := range days {
		if d.day.Tithi != nil {
			intern(*d.day.Tithi)
			if tithiFirst < 0 {
				tithiFirst = d.serial
			}
			tithiLast = d.serial
		}
	}
	p.tithiNames = len(p.pairs)
	if p.tithiNames > maxTithiNames {
		return nil, fmt.Errorf("%d distinct tithi names, the dayInfo bit field holds at most %d", p.tithiNames, maxTithiNames)
	}

	p.dayInfo = make([]uint32, p.covLast-p.covFirst+1)
	for _, d := range days {
		var rec uint32
		if d.day.Tithi != nil {
			rec |= uint32(ids[*d.day.Tithi] + 1) //nolint:gosec // below 63, checked above
		}
		if d.day.Holiday {
			rec |= holidayBit
		}
		if n := len(d.day.Events); n > 0 {
			if n > maxDayEvents {
				return nil, fmt.Errorf("%s has %d events, the dayInfo bit field holds at most %d", d.date, n, maxDayEvents)
			}
			first := len(p.eventName)
			if first+n > maxEvents {
				return nil, fmt.Errorf("more than %d events in total, the dayInfo bit field can't index them", maxEvents)
			}
			for _, e := range d.day.Events {
				p.eventName = append(p.eventName, uint16(intern(e))) //nolint:gosec // name ids are checked against the uint16 limit below
			}
			rec |= uint32(n)<<countShift | uint32(first)<<firstShift //nolint:gosec // both checked above
		}
		p.dayInfo[d.serial-p.covFirst] = rec
	}
	if len(p.pairs) > 1<<16 {
		return nil, fmt.Errorf("%d names don't fit the uint16 eventName table", len(p.pairs))
	}
	// Every day from the first to the last tithi day needs a tithi (PATRO §3).
	for s := tithiFirst; s <= tithiLast; s++ {
		if p.dayInfo[s-p.covFirst]&tithiMask == 0 {
			return nil, fmt.Errorf("tithi missing on %s inside the tithi range", dateOf(&bs, s))
		}
	}

	var text bytes.Buffer
	for _, n := range p.pairs {
		p.nameOff = append(p.nameOff, uint32(text.Len())) //nolint:gosec // the name pool fits uint32
		text.WriteString(n.En)
		p.nameOff = append(p.nameOff, uint32(text.Len())) //nolint:gosec // the name pool fits uint32
		text.WriteString(n.Np)
	}
	p.nameOff = append(p.nameOff, uint32(text.Len())) //nolint:gosec // the name pool fits uint32
	p.nameText = text.String()
	return p, nil
}

// serialOf returns the serial of a "YYYY-MM-DD" BS date.
func serialOf(cal *calendar, bs *[bsMonths + 1]int32, date string) (int32, error) {
	var y, m, d int
	if _, err := fmt.Sscanf(date, "%4d-%2d-%2d", &y, &m, &d); err != nil {
		return 0, fmt.Errorf("patro date %q: %w", date, err)
	}
	if y < minYear || y > maxYear || m < 1 || m > 12 || d < 1 || d > cal.lengths[y-minYear][m-1] {
		return 0, fmt.Errorf("patro date %q is not a valid BS date", date)
	}
	return bs[(y-minYear)*12+m-1] + int32(d) - 1, nil //nolint:gosec // 1..32
}

// dateOf returns serial s as "YYYY-MM-DD".
func dateOf(bs *[bsMonths + 1]int32, s int32) string {
	k := 0
	for bs[k+1] <= s {
		k++
	}
	return fmt.Sprintf("%04d-%02d-%02d", minYear+k/12, k%12+1, s-bs[k]+1)
}

func renderPatro(cal *calendar, p *patroTables) ([]byte, error) {
	bs := bsStarts(cal)
	var b bytes.Buffer
	b.WriteString(header)
	b.WriteString("\n// Source: shared/data/patro/*.json. Layout: ADR-0002 §2.\n\n")
	b.WriteString("package patro\n\n")
	fmt.Fprintf(&b, `// Coverage is serials covFirst..covLast (PATRO §2).
const (
	covFirst = %d
	covLast  = %d
)

// Bits of a dayInfo record.
const (
	tithiMask  = %d
	holidayBit = %d
	countShift = %d
	countMask  = %d
	firstShift = %d
)

`, p.covFirst, p.covLast, tithiMask, holidayBit, countShift, countMask, firstShift)

	b.WriteString("// dayInfo has one record per day of the coverage: tithi name id + 1, holiday bit, event\n")
	b.WriteString("// count and the index of the first event in eventName.\n")
	fmt.Fprintf(&b, "var dayInfo = [%d]uint32{\n", len(p.dayInfo))
	for i, rec := range p.dayInfo {
		fmt.Fprintf(&b, "\t%d, // %s\n", rec, dateOf(&bs, p.covFirst+int32(i))) //nolint:gosec // coverage fits int32
	}
	b.WriteString("}\n\n")

	b.WriteString("// eventName holds the name id of every event, day by day, in file order.\n")
	fmt.Fprintf(&b, "var eventName = [%d]uint16{", len(p.eventName))
	for i, id := range p.eventName {
		if i%16 == 0 {
			b.WriteString("\n\t")
		}
		fmt.Fprintf(&b, "%d, ", id)
	}
	b.WriteString("\n}\n\n")

	b.WriteString("// nameOff[2*i+l] and nameOff[2*i+l+1] bound name i in language l (0 English, 1 Nepali)\n")
	b.WriteString("// inside nameText. Tithi names come first.\n")
	fmt.Fprintf(&b, "var nameOff = [%d]uint32{", len(p.nameOff))
	for i, off := range p.nameOff {
		if i%12 == 0 {
			b.WriteString("\n\t")
		}
		fmt.Fprintf(&b, "%d, ", off)
	}
	b.WriteString("\n}\n\n")

	fmt.Fprintf(&b, "// tithiNames is the number of tithi names at the start of the name table.\nconst tithiNames = %d\n\n", p.tithiNames)
	b.WriteString("// nameText is every distinct name, English then Nepali for each.\n")
	b.WriteString("const nameText = " + strconv.Quote(p.nameText) + "\n")
	return gofmt("patro/patro_gen.go", b.Bytes())
}
