package nepdate_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	nepdate "github.com/sanamhub/nepdate-go"
)

func ExampleMustNew() {
	d := nepdate.MustNew(2081, nepdate.Shrawan, 15)
	fmt.Println(d)
	// Output: 2081/04/15
}

func ExampleFromGregorian() {
	d, err := nepdate.FromGregorian(2024, time.August, 16)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(d)
	// Output: 2081/04/32
}

func ExampleToday() {
	today, err := nepdate.Today(time.FixedZone("NPT", 5*3600+45*60))
	if err != nil {
		log.Fatal(err)
	}
	// The date depends on the clock; the time of the same date does not.
	fmt.Println(today.Time(nil).Format("15:04"))
	// Output: 00:00
}

func ExampleParse() {
	d, err := nepdate.Parse("2081-04-15")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(d, d.Weekday())
	// Output: 2081/04/15 Tuesday
}

func ExampleWeekdayName() {
	fmt.Println(nepdate.WeekdayName(time.Tuesday, nepdate.English))
	fmt.Println(nepdate.WeekdayName(time.Tuesday, nepdate.Nepali))
	// Output:
	// Tuesday
	// मङ्गलवार
}

func ExampleWeekdayShortName() {
	fmt.Println(nepdate.WeekdayShortName(time.Tuesday, nepdate.English))
	fmt.Println(nepdate.WeekdayShortName(time.Tuesday, nepdate.Nepali))
	// Output:
	// Tue
	// मङ्गल
}

func ExampleFiscalYearOf() {
	d := nepdate.MustNew(2081, nepdate.Shrawan, 15)
	fmt.Println(nepdate.FiscalYearOf(d).Label(nepdate.English))
	// Output: 2081/82
}

func ExampleNewRange() {
	start := nepdate.MustNew(2081, nepdate.Shrawan, 1)
	end := nepdate.MustNew(2081, nepdate.Shrawan, 3)
	r, ok := nepdate.NewRange(start, end)
	fmt.Println(ok, r.Len())
	// Output: true 3
}

func ExampleMonthRange() {
	r, err := nepdate.MonthRange(2081, nepdate.Shrawan)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(r.Start(), r.End(), r.Len())
	// Output: 2081/04/01 2081/04/32 32
}

func ExampleYearRange() {
	r, err := nepdate.YearRange(2081)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(r.Start(), r.End(), r.Len())
	// Output: 2081/01/01 2081/12/31 366
}

func ExampleDate() {
	d := nepdate.MustNew(2081, nepdate.Shrawan, 15)
	y, m, day := d.Gregorian()
	fmt.Println(d.Year(), d.Month(), d.Day(), d.Weekday())
	fmt.Println(y, int(m), day)
	// Output:
	// 2081 Shrawan 15 Tuesday
	// 2024 7 30
}

func ExampleLang() {
	fmt.Println(string(nepdate.English.AppendDigits(nil, 2081, 4)))
	fmt.Println(string(nepdate.Nepali.AppendDigits(nil, 2081, 4)))
	// Output:
	// 2081
	// २०८१
}

func ExampleMonth() {
	fmt.Println(nepdate.Shrawan.Name(nepdate.English))
	fmt.Println(nepdate.Shrawan.Name(nepdate.Nepali))
	fmt.Println(nepdate.Ashad.ShortName(nepdate.English))
	// Output:
	// Shrawan
	// साउन
	// Asa
}

func ExampleParseErrorKind() {
	_, err := nepdate.Parse("2081/04/xx")
	var parseErr *nepdate.ParseError
	if errors.As(err, &parseErr) {
		switch parseErr.Kind {
		case nepdate.InvalidCharacter:
			fmt.Println("the date contains a character that is not a digit or a separator")
		case nepdate.WrongGroupCount:
			fmt.Println("the date has the wrong number of parts")
		case nepdate.NumberTooLong:
			fmt.Println("a number in the date is too long")
		case nepdate.Ambiguous:
			fmt.Println("the date is ambiguous")
		case nepdate.Unrecognized:
			fmt.Println("the date is not a recognized form")
		}
	}
	// Output: the date contains a character that is not a digit or a separator
}

func ExampleParseError() {
	_, err := nepdate.Parse("2081/04/xx")
	var parseErr *nepdate.ParseError
	if errors.As(err, &parseErr) {
		fmt.Println(parseErr.Input)
		fmt.Println(err)
	}
	// Output:
	// 2081/04/xx
	// nepdate: cannot parse "2081/04/xx": invalid character
}

func ExampleOverflow() {
	last := nepdate.MustNew(2081, nepdate.Shrawan, 32)
	clamped, err := last.AddMonths(1, nepdate.Clamp)
	if err != nil {
		log.Fatal(err)
	}
	spilled, err := last.AddMonths(1, nepdate.Spill)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(clamped, spilled)
	// Output: 2081/05/31 2081/06/01
}

func ExampleDiff() {
	start := nepdate.MustNew(2081, nepdate.Baishakh, 1)
	end := nepdate.MustNew(2081, nepdate.Shrawan, 15)
	diff := start.Diff(end)
	fmt.Println(diff.Years, diff.Months, diff.Days, diff.TotalDays)
	back := end.Diff(start)
	fmt.Println(back.Negative, back.TotalDays)
	// Output:
	// 0 3 14 108
	// true -108
}

func ExampleDateOrder() {
	d := nepdate.MustNew(2081, nepdate.Shrawan, 15)
	fmt.Println(d.Short(nepdate.DMY, nepdate.Dash, false, nepdate.English))
	fmt.Println(d.Short(nepdate.YMD, nepdate.Dash, true, nepdate.English))
	// Output:
	// 15-4-2081
	// 2081-04-15
}

func ExampleSeparator() {
	d := nepdate.MustNew(2081, nepdate.Shrawan, 15)
	fmt.Println(d.Short(nepdate.YMD, nepdate.Dot, true, nepdate.English))
	fmt.Println(d.Short(nepdate.YMD, nepdate.Space, true, nepdate.English))
	// Output:
	// 2081.04.15
	// 2081 04 15
}

func ExampleLongOptions() {
	d := nepdate.MustNew(2081, nepdate.Shrawan, 15)
	fmt.Println(d.Long(nepdate.LongOptions{}, nepdate.English))
	fmt.Println(d.Long(nepdate.LongOptions{Weekday: true, OmitYear: true}, nepdate.English))
	// Output:
	// Shrawan 15, 2081
	// Tuesday, Shrawan 15
}

func ExampleFiscalYear() {
	f := nepdate.MustNew(2081, nepdate.Shrawan, 15).FiscalYear()
	start, err := f.Start()
	if err != nil {
		log.Fatal(err)
	}
	end, err := f.End()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(f.Label(nepdate.English), start, end)
	// Output: 2081/82 2081/04/01 2082/03/32
}

func ExampleQuarter() {
	fmt.Println(nepdate.Q1.Months())
	// Output: [Shrawan Bhadra Ashoj]
}

func ExampleRange() {
	start := nepdate.MustNew(2081, nepdate.Shrawan, 1)
	end := nepdate.MustNew(2081, nepdate.Shrawan, 3)
	r, ok := nepdate.NewRange(start, end)
	if !ok {
		log.Fatal("start is after end")
	}
	fmt.Println(r.Start(), r.End(), r.Len(), r.Contains(end))
	// Output: 2081/04/01 2081/04/03 3 true
}

func ExampleDate_AddDays() {
	d := nepdate.MustNew(2081, nepdate.Shrawan, 15)
	later, err := d.AddDays(16)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(later)
	// Output: 2081/04/31
}

func ExampleDate_AddMonths() {
	d := nepdate.MustNew(2081, nepdate.Shrawan, 15)
	later, err := d.AddMonths(4, nepdate.Clamp)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(later)
	// Output: 2081/08/15
}

func ExampleDate_AddYears() {
	d := nepdate.MustNew(2080, nepdate.Falgun, 30)
	later, err := d.AddYears(1, nepdate.Clamp)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(later)
	// Output: 2081/11/29
}

func ExampleDate_DaysUntil() {
	start := nepdate.MustNew(2081, nepdate.Baishakh, 1)
	end := nepdate.MustNew(2081, nepdate.Shrawan, 15)
	fmt.Println(start.DaysUntil(end))
	// Output: 108
}

func ExampleDate_Diff() {
	start := nepdate.MustNew(2081, nepdate.Baishakh, 1)
	end := nepdate.MustNew(2081, nepdate.Shrawan, 15)
	diff := end.Diff(start)
	fmt.Println(diff.Years, diff.Months, diff.Days)
	// Output: 0 3 14
}

func ExampleDate_Short() {
	d := nepdate.MustNew(2081, nepdate.Shrawan, 15)
	fmt.Println(d.Short(nepdate.DMY, nepdate.Dash, false, nepdate.English))
	fmt.Println(d.Short(nepdate.YMD, nepdate.Slash, true, nepdate.Nepali))
	// Output:
	// 15-4-2081
	// २०८१/०४/१५
}

func ExampleDate_Long() {
	d := nepdate.MustNew(2081, nepdate.Shrawan, 15)
	fmt.Println(d.Long(nepdate.LongOptions{Weekday: true}, nepdate.Nepali))
	// Output: मङ्गलवार, साउन १५, २०८१
}

func ExampleDate_AppendShort() {
	d := nepdate.MustNew(2081, nepdate.Shrawan, 15)
	var b []byte
	b = d.AppendShort(b, nepdate.YMD, nepdate.Dot, true, nepdate.English)
	fmt.Println(string(b))
	// Output: 2081.04.15
}

func ExampleDate_MarshalText() {
	d := nepdate.MustNew(2081, nepdate.Shrawan, 15)
	payload, err := json.Marshal(struct {
		Date nepdate.Date `json:"date"`
	}{d})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(payload))
	// Output: {"date":"2081-04-15"}
}

func ExampleDate_UnmarshalText() {
	var payload struct {
		Date nepdate.Date `json:"date"`
	}
	if err := json.Unmarshal([]byte(`{"date":"2081-04-15"}`), &payload); err != nil {
		log.Fatal(err)
	}
	fmt.Println(payload.Date)
	// Output: 2081/04/15
}

func ExampleFiscalYear_QuarterRange() {
	f := nepdate.MustNew(2081, nepdate.Shrawan, 15).FiscalYear()
	r, err := f.QuarterRange(nepdate.Q1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(r.Start(), r.End(), r.Len())
	// Output: 2081/04/01 2081/06/30 93
}

func ExampleRange_Days() {
	start := nepdate.MustNew(2081, nepdate.Shrawan, 1)
	end := nepdate.MustNew(2081, nepdate.Shrawan, 3)
	r, ok := nepdate.NewRange(start, end)
	if !ok {
		log.Fatal("start is after end")
	}
	for d := range r.Days() {
		fmt.Println(d)
	}
	// Output:
	// 2081/04/01
	// 2081/04/02
	// 2081/04/03
}
