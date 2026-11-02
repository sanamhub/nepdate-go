package nepdate

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	orderByVector = map[string]DateOrder{
		"Ymd": YMD, "Ydm": YDM, "Myd": MYD, "Mdy": MDY, "Dym": DYM, "Dmy": DMY,
	}
	separatorByVector = map[string]Separator{
		"Slash": Slash, "Backslash": Backslash, "Dot": Dot,
		"Underscore": Underscore, "Dash": Dash, "Space": Space,
	}
)

func langByVector(t *testing.T, s string) Lang {
	t.Helper()
	switch s {
	case "en":
		return English
	case "ne":
		return Nepali
	}
	t.Fatalf("unknown language %q", s)
	return English
}

func orderBy(t *testing.T, s string) DateOrder {
	t.Helper()
	o, ok := orderByVector[s]
	if !ok {
		t.Fatalf("unknown order %q", s)
	}
	return o
}

func separatorBy(t *testing.T, s string) Separator {
	t.Helper()
	sep, ok := separatorByVector[s]
	if !ok {
		t.Fatalf("unknown separator %q", s)
	}
	return sep
}

// dateOf parses a vector date without failing the test, so an ERR row can name an invalid date.
func dateOf(s string) (Date, bool) {
	parts := strings.Split(s, "-")
	if len(parts) != 3 {
		return Date{}, false
	}
	var n [3]int
	for i, p := range parts {
		v, err := strconv.Atoi(p)
		if err != nil {
			return Date{}, false
		}
		n[i] = v
	}
	d, err := New(n[0], Month(n[1]), n[2]) //nolint:gosec // New rejects a month outside 1..12
	return d, err == nil
}

func vectorDate(d Date) string {
	return d.Short(YMD, Dash, true, English)
}

func rangeText(r Range) string {
	return vectorDate(r.Start()) + ".." + vectorDate(r.End())
}

func rangeOf(t *testing.T, spec string) (Range, bool) {
	t.Helper()
	sep := ".."
	if strings.Contains(spec, ",") {
		sep = ","
	}
	a, b, ok := strings.Cut(spec, sep)
	if !ok {
		t.Fatalf("bad range %q", spec)
	}
	start, ok := dateOf(a)
	if !ok {
		return Range{}, false
	}
	end, ok := dateOf(b)
	if !ok {
		return Range{}, false
	}
	return NewRange(start, end)
}

func joinDates(dates []Date) string {
	parts := make([]string, len(dates))
	for i, d := range dates {
		parts[i] = vectorDate(d)
	}
	return strings.Join(parts, ";")
}

func dateList(dates []Date) string {
	if len(dates) == 0 {
		return "none"
	}
	return joinDates(dates)
}

func joinRanges(ranges []Range) string {
	if len(ranges) == 0 {
		return "none"
	}
	parts := make([]string, len(ranges))
	for i, r := range ranges {
		parts[i] = rangeText(r)
	}
	return strings.Join(parts, ";")
}

func collect(seq func(func(Date) bool)) []Date {
	var dates []Date
	for d := range seq {
		dates = append(dates, d)
	}
	return dates
}

func TestYearLengthVectors(t *testing.T) {
	rows := readTable(t, "shared/spec/vectors/year-length.tsv", "\t", 299)
	for _, row := range rows {
		year, err := strconv.Atoi(row["bs_year"])
		if err != nil {
			t.Fatal(err)
		}
		want, err := strconv.Atoi(row["year_length"])
		if err != nil {
			t.Fatal(err)
		}
		d := MustNew(year, Baishakh, 1)
		if got := d.YearLength(); got != want {
			t.Errorf("%d: YearLength() = %d, want %d", year, got, want)
		}
		sum := 0
		for m := Baishakh; m <= Chaitra; m++ {
			sum += MustNew(year, m, 1).MonthLength()
		}
		if got := d.YearLength(); got != sum {
			t.Errorf("%d: YearLength() = %d, months sum to %d", year, got, sum)
		}
	}
	if got := (Date{}).YearLength(); got != 0 {
		t.Errorf("zero Date YearLength() = %d", got)
	}
}

func TestDayDiffVectors(t *testing.T) {
	rows := readTable(t, "shared/spec/vectors/csharp-golden/day-diff.tsv", "\t", 16)
	for _, row := range rows {
		a := bsDate(t, row["a"])
		b := bsDate(t, row["b"])
		wantUntil, _ := strconv.Atoi(row["days_until"])
		wantMinus, _ := strconv.Atoi(row["operator_minus"])
		if got := a.DaysUntil(b); got != wantUntil {
			t.Errorf("%s.DaysUntil(%s) = %d, want %d", a, b, got, wantUntil)
		}
		if got := b.DaysUntil(a); got != wantMinus {
			t.Errorf("%s - %s = %d, want %d", a, b, got, wantMinus)
		}
		if got := a.Diff(b).TotalDays; got != wantUntil {
			t.Errorf("%s.Diff(%s).TotalDays = %d, want %d", a, b, got, wantUntil)
		}
	}
}

func TestYMDIntVectors(t *testing.T) {
	rows := readTable(t, "shared/spec/vectors/csharp-golden/ymd-int.tsv", "\t", 14)
	for _, row := range rows {
		switch row["kind"] {
		case "date":
			d := bsDate(t, row["input"])
			want, err := strconv.Atoi(row["ymd"])
			if err != nil {
				t.Fatal(err)
			}
			if got := d.YMD(); got != want {
				t.Errorf("%s.YMD() = %d, want %d", d, got, want)
			}
			if row["result"] == "ERR" {
				t.Errorf("%s: a valid date has no ERR result", row["input"])
			}
		case "int":
			n, err := strconv.Atoi(row["ymd"])
			if err != nil {
				t.Fatal(err)
			}
			got, err := FromYMD(n)
			if row["result"] == "ERR" {
				if err == nil {
					t.Errorf("FromYMD(%d) = %s, want ERR", n, got)
				}
				continue
			}
			want := bsDate(t, row["result"])
			if err != nil || got != want {
				t.Errorf("FromYMD(%d) = %s, %v; want %s", n, got, err, want)
			}
		default:
			t.Fatalf("unknown kind %q", row["kind"])
		}
	}
	if got := (Date{}).YMD(); got != 0 {
		t.Errorf("zero Date YMD() = %d", got)
	}
}

func TestBulkVectors(t *testing.T) {
	rows := readTable(t, "shared/spec/vectors/csharp-golden/bulk.tsv", "\t", 6)
	for _, row := range rows {
		input := strings.Split(row["input"], ";")
		var got string
		var err error
		switch row["operation"] {
		case "to_nepali_dates", "batch_to_nepali_dates":
			times := make([]time.Time, len(input))
			for i, s := range input {
				times[i], err = time.Parse("2006-01-02", s)
				if err != nil {
					t.Fatal(err)
				}
			}
			var dates []Date
			dates, err = ToNepaliDates(times)
			if err == nil {
				got = joinDates(dates)
			}
		case "to_english_dates_from_strings":
			var times []time.Time
			times, err = ToEnglishDatesFromStrings(input)
			if err == nil {
				parts := make([]string, len(times))
				for i, tm := range times {
					parts[i] = tm.Format("2006-01-02")
				}
				got = strings.Join(parts, ";")
			}
		case "to_english_dates_from_dates", "batch_to_english_dates":
			dates := make([]Date, len(input))
			for i, s := range input {
				dates[i] = bsDate(t, s)
			}
			var times []time.Time
			times, err = ToEnglishDates(dates)
			if err == nil {
				parts := make([]string, len(times))
				for i, tm := range times {
					parts[i] = tm.Format("2006-01-02")
				}
				got = strings.Join(parts, ";")
			}
		default:
			t.Fatalf("unknown operation %q", row["operation"])
		}
		if row["output"] == "ERR" {
			if err == nil {
				t.Errorf("%s: got %q, want ERR", row["operation"], got)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: %v", row["operation"], err)
		}
		if got != row["output"] {
			t.Errorf("%s = %q, want %q", row["operation"], got, row["output"])
		}
	}
}

func TestSerializationVectors(t *testing.T) {
	rows := readTable(t, "shared/spec/vectors/csharp-golden/serialization.tsv", "\t", 11)
	for _, row := range rows {
		switch row["kind"] {
		case "write":
			d := bsDate(t, row["input"])
			got, err := json.Marshal(d)
			if err != nil || string(got) != row["json"] {
				t.Errorf("Marshal(%s) = %s, %v; want %s", d, got, err, row["json"])
			}
		case "read":
			var d Date
			err := json.Unmarshal([]byte(row["input"]), &d)
			if row["result"] == "ERR" {
				if err == nil {
					t.Errorf("Unmarshal(%s) = %s, want ERR", row["input"], d)
				}
				continue
			}
			want := bsDate(t, row["result"])
			if err != nil || d != want {
				t.Errorf("Unmarshal(%s) = %s, %v; want %s", row["input"], d, err, want)
			}
		case "roundtrip":
			d := bsDate(t, row["input"])
			payload, err := json.Marshal(d)
			if err != nil || string(payload) != row["json"] {
				t.Errorf("Marshal(%s) = %s, %v; want %s", d, payload, err, row["json"])
			}
			var back Date
			if err := json.Unmarshal(payload, &back); err != nil || back != bsDate(t, row["result"]) {
				t.Errorf("round trip %s = %s, %v", d, back, err)
			}
		default:
			t.Fatalf("unknown kind %q", row["kind"])
		}
	}
}

func TestFormatMatrixVectors(t *testing.T) {
	rows := readTable(t, "shared/spec/vectors/csharp-golden/format-matrix.tsv", "\t", 102)
	for _, row := range rows {
		d := bsDate(t, row["bs"])
		lang := langByVector(t, row["lang"])
		switch row["kind"] {
		case "short":
			pad := row["pad"] == "true"
			got := d.Short(orderBy(t, row["order"]), separatorBy(t, row["separator"]), pad, lang)
			if got != row["result"] {
				t.Errorf("%s.Short(%s, %s, %v, %s) = %q, want %q",
					d, row["order"], row["separator"], pad, row["lang"], got, row["result"])
			}
		case "try_format":
			capacity, err := strconv.Atoi(row["capacity"])
			if err != nil {
				t.Fatal(err)
			}
			buf := make([]byte, 0, capacity)
			var out []byte
			if row["format"] == "default" {
				out = d.AppendShort(buf, YMD, Slash, true, lang)
			} else {
				out = d.AppendFormat(buf, row["format"], lang)
			}
			got := row["result"]
			if len(out) > capacity {
				got = "ERR"
			} else if string(out) != row["result"] {
				got = string(out)
			}
			if got != row["result"] {
				t.Errorf("%s.Append(%q) in %d bytes = %q, want %q", d, row["format"], capacity, got, row["result"])
			}
		default:
			t.Fatalf("unknown kind %q", row["kind"])
		}
	}
}

func TestRangeFactoryVectors(t *testing.T) {
	rows := readTable(t, "shared/spec/vectors/csharp-golden/range-factories.tsv", "\t", 27)
	for _, row := range rows {
		var got Range
		var err error
		ok := true
		switch row["factory"] {
		case "single_day":
			d, valid := dateOf(row["argument"])
			if !valid {
				ok = false
			} else {
				got = SingleDay(d)
			}
		case "from_day_count":
			parts := strings.Split(row["argument"], ",")
			if len(parts) != 2 {
				t.Fatalf("bad from_day_count argument %q", row["argument"])
			}
			d, valid := dateOf(parts[0])
			days, convErr := strconv.Atoi(parts[1])
			if !valid || convErr != nil {
				ok = false
			} else {
				got, err = FromDayCount(d, days)
			}
		case "for_month":
			parts := strings.Split(row["argument"], ",")
			if len(parts) != 2 {
				t.Fatalf("bad for_month argument %q", row["argument"])
			}
			year, _ := strconv.Atoi(parts[0])
			month, _ := strconv.Atoi(parts[1])
			got, err = MonthRange(year, Month(month)) //nolint:gosec // MonthRange rejects a month outside 1..12
		case "for_calendar_year":
			year, _ := strconv.Atoi(row["argument"])
			got, err = YearRange(year)
		case "for_fiscal_year":
			year, _ := strconv.Atoi(row["argument"])
			got, err = FiscalYear(year).Range()
		case "construct":
			got, ok = rangeOf(t, row["argument"])
		default:
			t.Fatalf("unknown factory %q", row["factory"])
		}
		if row["result"] == "ERR" {
			if err == nil && ok {
				t.Errorf("%s(%s) = %s, want ERR", row["factory"], row["argument"], rangeText(got))
			}
			continue
		}
		if err != nil || !ok {
			t.Errorf("%s(%s): %v, ok %v", row["factory"], row["argument"], err, ok)
			continue
		}
		if gotText := rangeText(got); gotText != row["result"] {
			t.Errorf("%s(%s) = %s, want %s", row["factory"], row["argument"], gotText, row["result"])
		}
	}
}

func TestRangeVectors(t *testing.T) {
	rows := readTable(t, "shared/spec/vectors/csharp-golden/ranges.tsv", "\t", 54)
	for _, row := range rows {
		start, startOK := dateOf(row["start"])
		end, endOK := dateOf(row["end"])
		if !startOK || !endOK {
			if row["result"] != "ERR" {
				t.Fatalf("%s..%s: invalid bounds for a row that expects a value", row["start"], row["end"])
			}
			continue
		}
		r, ok := NewRange(start, end)
		switch row["operation"] {
		case "construct":
			if row["result"] == "ERR" {
				if ok {
					t.Errorf("NewRange(%s, %s) = %s, want ERR", start, end, rangeText(r))
				}
				continue
			}
			if !ok || rangeText(r) != row["result"] {
				t.Errorf("NewRange(%s, %s) = %s, %v; want %s", start, end, rangeText(r), ok, row["result"])
			}
		case "length":
			if !ok {
				t.Fatalf("%s..%s is invalid", start, end)
			}
			want, _ := strconv.Atoi(row["result"])
			if got := r.Len(); got != want {
				t.Errorf("%s.Len() = %d, want %d", rangeText(r), got, want)
			}
		case "contains_date":
			if !ok {
				t.Fatalf("%s..%s is invalid", start, end)
			}
			d, valid := dateOf(row["argument"])
			if !valid {
				t.Fatalf("bad argument %q", row["argument"])
			}
			if got := r.Contains(d); strconv.FormatBool(got) != row["result"] {
				t.Errorf("%s.Contains(%s) = %v, want %s", rangeText(r), d, got, row["result"])
			}
		case "contains_range":
			if !ok {
				t.Fatalf("%s..%s is invalid", start, end)
			}
			other, valid := rangeOf(t, row["argument"])
			if !valid {
				t.Fatalf("bad argument range %q", row["argument"])
			}
			if got := r.ContainsRange(other); strconv.FormatBool(got) != row["result"] {
				t.Errorf("%s.ContainsRange(%s) = %v, want %s", rangeText(r), rangeText(other), got, row["result"])
			}
		case "overlaps":
			if !ok {
				t.Fatalf("%s..%s is invalid", start, end)
			}
			other, valid := rangeOf(t, row["argument"])
			if !valid {
				t.Fatalf("bad argument range %q", row["argument"])
			}
			if got := r.Overlaps(other); strconv.FormatBool(got) != row["result"] {
				t.Errorf("%s.Overlaps(%s) = %v, want %s", rangeText(r), rangeText(other), got, row["result"])
			}
		case "is_adjacent_to":
			if !ok {
				t.Fatalf("%s..%s is invalid", start, end)
			}
			other, valid := rangeOf(t, row["argument"])
			if !valid {
				t.Fatalf("bad argument range %q", row["argument"])
			}
			if got := r.IsAdjacentTo(other); strconv.FormatBool(got) != row["result"] {
				t.Errorf("%s.IsAdjacentTo(%s) = %v, want %s", rangeText(r), rangeText(other), got, row["result"])
			}
		case "intersect":
			if !ok {
				t.Fatalf("%s..%s is invalid", start, end)
			}
			other, valid := rangeOf(t, row["argument"])
			if !valid {
				t.Fatalf("bad argument range %q", row["argument"])
			}
			got, have := r.Intersect(other)
			if row["result"] == "none" {
				if have {
					t.Errorf("%s.Intersect(%s) = %s, want none", rangeText(r), rangeText(other), rangeText(got))
				}
				continue
			}
			if !have || rangeText(got) != row["result"] {
				t.Errorf("%s.Intersect(%s) = %s, %v; want %s", rangeText(r), rangeText(other), rangeText(got), have, row["result"])
			}
		case "union":
			if !ok {
				t.Fatalf("%s..%s is invalid", start, end)
			}
			other, valid := rangeOf(t, row["argument"])
			if !valid {
				t.Fatalf("bad argument range %q", row["argument"])
			}
			if got := r.Union(other); rangeText(got) != row["result"] {
				t.Errorf("%s.Union(%s) = %s, want %s", rangeText(r), rangeText(other), rangeText(got), row["result"])
			}
		case "except":
			if !ok {
				t.Fatalf("%s..%s is invalid", start, end)
			}
			other, valid := rangeOf(t, row["argument"])
			if !valid {
				t.Fatalf("bad argument range %q", row["argument"])
			}
			if got := joinRanges(r.Except(other)); got != row["result"] {
				t.Errorf("%s.Except(%s) = %s, want %s", rangeText(r), rangeText(other), got, row["result"])
			}
		case "equals":
			if !ok {
				t.Fatalf("%s..%s is invalid", start, end)
			}
			other, valid := rangeOf(t, row["argument"])
			if !valid {
				t.Fatalf("bad argument range %q", row["argument"])
			}
			if got := strconv.FormatBool(r == other); got != row["result"] {
				t.Errorf("%s == %s = %s, want %s", rangeText(r), rangeText(other), got, row["result"])
			}
		case "to_string":
			if !ok {
				t.Fatalf("%s..%s is invalid", start, end)
			}
			if got := r.String(); got != row["result"] {
				t.Errorf("%s.String() = %q, want %q", rangeText(r), got, row["result"])
			}
		case "enumerate":
			if !ok {
				t.Fatalf("%s..%s is invalid", start, end)
			}
			if got := joinDates(collect(r.Days())); got != row["result"] {
				t.Errorf("%s.Days() = %q, want %q", rangeText(r), got, row["result"])
			}
		default:
			t.Fatalf("unknown operation %q", row["operation"])
		}
	}
}

func TestRangeSplitVectors(t *testing.T) {
	rows := readTable(t, "shared/spec/vectors/csharp-golden/range-splits.tsv", "\t", 9)
	for _, row := range rows {
		r, ok := rangeOf(t, row["start"]+".."+row["end"])
		if !ok {
			t.Fatalf("%s..%s is invalid", row["start"], row["end"])
		}
		var got []Range
		switch row["operation"] {
		case "split_by_month":
			got = r.SplitByMonth()
		case "split_by_fiscal_quarter":
			got = r.SplitByFiscalQuarter()
		default:
			t.Fatalf("unknown operation %q", row["operation"])
		}
		if text := joinRanges(got); text != row["result"] {
			t.Errorf("%s.%s() = %s, want %s", rangeText(r), row["operation"], text, row["result"])
		}
	}
}

func TestRangeDayVectors(t *testing.T) {
	rows := readTable(t, "shared/spec/vectors/csharp-golden/range-days.tsv", "\t", 18)
	for _, row := range rows {
		r, ok := rangeOf(t, row["start"]+".."+row["end"])
		if !ok {
			t.Fatalf("%s..%s is invalid", row["start"], row["end"])
		}
		var dates []Date
		var err error
		switch row["operation"] {
		case "dates_with_interval":
			interval, convErr := strconv.Atoi(row["argument"])
			if convErr != nil {
				t.Fatal(convErr)
			}
			var seq func(func(Date) bool)
			seq, err = r.DatesWithInterval(interval)
			if err == nil {
				dates = collect(seq)
			}
		case "working_days":
			dates = collect(r.WorkingDays(row["argument"] == "true"))
		case "weekend_days":
			dates = collect(r.WeekendDays(row["argument"] == "true"))
		case "enumerate":
			dates = collect(r.Days())
		default:
			t.Fatalf("unknown operation %q", row["operation"])
		}
		if row["result"] == "ERR" {
			if err == nil {
				t.Errorf("%s(%s) = %s, want ERR", row["operation"], row["argument"], joinDates(dates))
			}
			continue
		}
		if err != nil {
			t.Errorf("%s(%s): %v", row["operation"], row["argument"], err)
		}
		if got := dateList(dates); got != row["result"] {
			t.Errorf("%s(%s) = %q, want %q", row["operation"], row["argument"], got, row["result"])
		}
	}
}

func TestRangeFormatVectors(t *testing.T) {
	rows := readTable(t, "shared/spec/vectors/csharp-golden/range-format.tsv", "\t", 18)
	for _, row := range rows {
		r, ok := rangeOf(t, row["start"]+".."+row["end"])
		if row["result"] == "ERR" {
			if ok {
				t.Errorf("NewRange(%s, %s) = %s, want ERR", row["start"], row["end"], rangeText(r))
			}
			continue
		}
		if !ok {
			t.Fatalf("%s..%s is invalid", row["start"], row["end"])
		}
		var got string
		switch row["kind"] {
		case "default":
			got = r.String()
		case "short":
			got = r.Format(orderBy(t, row["order"]), separatorBy(t, row["separator"]), langByVector(t, row["lang"]))
		default:
			t.Fatalf("unknown kind %q", row["kind"])
		}
		if got != row["result"] {
			t.Errorf("%s.%s = %q, want %q", rangeText(r), row["kind"], got, row["result"])
		}
	}
}

func TestClockHelpersAtFixedInstant(t *testing.T) {
	clock := time.Date(2024, time.July, 30, 18, 30, 0, 0, time.UTC)
	today, err := todayAt(clock, npt)
	if err != nil {
		t.Fatal(err)
	}
	if want := MustNew(2081, Shrawan, 16); today != want {
		t.Fatalf("todayAt = %s, want %s", today, want)
	}
	yesterday, ok := today.Prev()
	if !ok {
		t.Fatal("today has no previous day")
	}
	tomorrow, ok := today.Next()
	if !ok {
		t.Fatal("today has no next day")
	}
	if !today.isTodayAt(clock, npt) || !yesterday.isYesterdayAt(clock, npt) || !tomorrow.isTomorrowAt(clock, npt) {
		t.Error("relative predicates disagree with the fixed clock")
	}
	if (Date{}).isTodayAt(clock, npt) || (Date{}).isYesterdayAt(clock, npt) || (Date{}).isTomorrowAt(clock, npt) {
		t.Error("a zero Date is relative to the clock")
	}
	month, err := currentMonthAt(clock, npt)
	if err != nil || rangeText(month) != "2081-04-01..2081-04-32" {
		t.Errorf("currentMonthAt = %s, %v", rangeText(month), err)
	}
	fy, err := currentFiscalYearAt(clock, npt)
	if err != nil || rangeText(fy) != "2081-04-01..2082-03-32" {
		t.Errorf("currentFiscalYearAt = %s, %v", rangeText(fy), err)
	}
	year, err := currentCalendarYearAt(clock, npt)
	if err != nil || rangeText(year) != "2081-01-01..2081-12-31" {
		t.Errorf("currentCalendarYearAt = %s, %v", rangeText(year), err)
	}
}

func TestRelativePredicatesUseLocation(t *testing.T) {
	clock := time.Date(2024, time.July, 30, 18, 30, 0, 0, time.UTC)
	// The same instant is still 30 July in UTC but already 31 July in Nepal.
	if !MustNew(2081, Shrawan, 15).isTodayAt(clock, time.UTC) {
		t.Error("UTC date is not today in UTC")
	}
	if MustNew(2081, Shrawan, 15).isTodayAt(clock, npt) {
		t.Error("UTC date is today in Nepal")
	}
	if !MustNew(2081, Shrawan, 16).isTodayAt(clock, npt) {
		t.Error("Nepal date is not today in Nepal")
	}
}

func TestTryParse(t *testing.T) {
	if d, ok := TryParse("2081/04/15"); !ok || d != MustNew(2081, Shrawan, 15) {
		t.Errorf("TryParse(valid) = %s, %v", d, ok)
	}
	if d, ok := TryParse("2081/04"); ok || !d.IsZero() {
		t.Errorf("TryParse(invalid) = %s, %v", d, ok)
	}
	if d, ok := TryParseLenient("15 Shrawan 2081"); !ok || d != MustNew(2081, Shrawan, 15) {
		t.Errorf("TryParseLenient(valid) = %s, %v", d, ok)
	}
	if d, ok := TryParseLenient("not a date"); ok || !d.IsZero() {
		t.Errorf("TryParseLenient(invalid) = %s, %v", d, ok)
	}
}

func TestRangeZeroValues(t *testing.T) {
	var r Range
	if r.String() != "" || r.Len() != 0 || r.Format(YMD, Slash, English) != "" {
		t.Error("zero Range has output")
	}
	if r.Union(SingleDay(Min)) != SingleDay(Min) {
		t.Error("zero Range changes a union")
	}
	if r.ContainsRange(SingleDay(Min)) {
		t.Error("zero Range contains a range")
	}
	if r.Overlaps(SingleDay(Min)) || r.IsAdjacentTo(SingleDay(Min)) {
		t.Error("zero Range overlaps or touches a range")
	}
	if _, ok := r.Intersect(SingleDay(Min)); ok {
		t.Error("zero Range intersects a range")
	}
	if got := r.Except(SingleDay(Min)); got != nil {
		t.Errorf("zero Range.Except = %v", got)
	}
	if got := r.SplitByMonth(); got != nil {
		t.Errorf("zero Range.SplitByMonth = %v", got)
	}
	if got := r.SplitByFiscalQuarter(); got != nil {
		t.Errorf("zero Range.SplitByFiscalQuarter = %v", got)
	}
	if _, err := r.DatesWithInterval(1); err != nil {
		t.Errorf("zero Range.DatesWithInterval: %v", err)
	}
}
