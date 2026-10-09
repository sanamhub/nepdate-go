package nepdate

import (
	"errors"
	"slices"
	"testing"
)

// checkDateCell compares a result with a golden cell: a date, or ERR for ErrOutOfRange.
func checkDateCell(t *testing.T, what string, got Date, err error, cell string) {
	t.Helper()
	if cell == "ERR" {
		if !errors.Is(err, ErrOutOfRange) {
			t.Errorf("%s = %v, %v; want ErrOutOfRange", what, got, err)
		}
		return
	}
	if want := bsDate(t, cell); err != nil || got != want {
		t.Errorf("%s = %v, %v; want %s", what, got, err, cell)
	}
}

func TestFiscalGolden(t *testing.T) {
	rows := readTable(t, "shared/spec/vectors/csharp-golden/dates.tsv", "\t", 1018)
	for _, row := range rows {
		d := bsDate(t, row["bs"])
		fy := FiscalYearOf(d)
		if fy != d.FiscalYear() {
			t.Fatalf("FiscalYearOf(%v) != %v.FiscalYear()", d, d)
		}
		start, err := fy.Start()
		checkDateCell(t, row["bs"]+" fiscal year start", start, err, row["fy_start"])
		end, err := fy.End()
		checkDateCell(t, row["bs"]+" fiscal year end", end, err, row["fy_end"])
		q, err := fy.QuarterRange(d.Quarter())
		if err != nil {
			t.Errorf("%s: QuarterRange(%d): %v", row["bs"], d.Quarter(), err)
			continue
		}
		checkDateCell(t, row["bs"]+" quarter start", q.Start(), nil, row["quarter_start"])
		checkDateCell(t, row["bs"]+" quarter end", q.End(), nil, row["quarter_end"])
		if !q.Contains(d) {
			t.Errorf("%s is not in its own quarter %v..%v", row["bs"], q.Start(), q.End())
		}
	}
}

func TestFiscalYear(t *testing.T) {
	if got := FiscalYear(2082).Label(English); got != "2082/83" {
		t.Errorf("Label(English) = %q", got)
	}
	if got := FiscalYear(2082).Label(Nepali); got != "२०८२/८३" {
		t.Errorf("Label(Nepali) = %q", got)
	}
	if got := FiscalYear(2099).Label(English); got != "2099/00" {
		t.Errorf("Label of 2099 = %q", got)
	}
	if _, err := FiscalYear(1900).Start(); !errors.Is(err, ErrOutOfRange) {
		t.Errorf("FiscalYear(1900).Start() error = %v", err)
	}
	if _, err := FiscalYear(2199).End(); !errors.Is(err, ErrOutOfRange) {
		t.Errorf("FiscalYear(2199).End() error = %v", err)
	}
	if _, err := FiscalYear(2199).Range(); !errors.Is(err, ErrOutOfRange) {
		t.Errorf("FiscalYear(2199).Range() error = %v", err)
	}
	r, err := FiscalYear(2081).Range()
	if err != nil || r.Start() != MustNew(2081, Shrawan, 1) || r.End() != MustNew(2082, Ashad, 1).LastOfMonth() {
		t.Errorf("FiscalYear(2081).Range() = %v..%v, %v", r.Start(), r.End(), err)
	}
	q4, err := FiscalYear(2081).QuarterRange(Q4)
	if err != nil || q4.Start() != MustNew(2082, Baishakh, 1) || q4.End() != MustNew(2082, Ashad, 1).LastOfMonth() {
		t.Errorf("Q4 of FY 2081 = %v..%v, %v", q4.Start(), q4.End(), err)
	}
	for _, q := range []Quarter{0, 5} {
		if _, err := FiscalYear(2081).QuarterRange(q); !errors.Is(err, ErrOutOfRange) {
			t.Errorf("QuarterRange(%d) error = %v", q, err)
		}
	}
	if FiscalYearOf(MustNew(2081, Ashad, 31)) != 2080 || FiscalYearOf(MustNew(2081, Shrawan, 1)) != 2081 {
		t.Error("the fiscal year does not start on Shrawan 1")
	}
	if FiscalYearOf(Date{}) != 0 || (Date{}).Quarter() != 0 {
		t.Error("zero Date has a fiscal year or quarter")
	}
}

func TestQuarterMonths(t *testing.T) {
	want := map[Quarter][3]Month{
		Q1: {Shrawan, Bhadra, Ashoj},
		Q2: {Kartik, Mangsir, Poush},
		Q3: {Magh, Falgun, Chaitra},
		Q4: {Baishakh, Jestha, Ashad},
		0:  {},
		5:  {},
	}
	for q, w := range want {
		if got := q.Months(); got != w {
			t.Errorf("Quarter(%d).Months() = %v, want %v", q, got, w)
		}
	}
	for m := Baishakh; m <= Chaitra; m++ {
		q := MustNew(2081, m, 1).Quarter()
		if ms := q.Months(); !slices.Contains(ms[:], m) {
			t.Errorf("month %d is in quarter %d with months %v", m, q, ms)
		}
	}
}

func TestAppendDigits(t *testing.T) {
	cases := []struct {
		n, width int
		lang     Lang
		want     string
	}{
		{2081, 0, Nepali, "२०८१"},
		{5, 2, Nepali, "०५"},
		{5, 2, English, "05"},
		{0, 0, English, "0"},
		{-5, 2, English, "-05"},
		{123, 2, English, "123"},
	}
	for _, c := range cases {
		if got := string(appendDigits(nil, c.n, c.width, c.lang)); got != c.want {
			t.Errorf("appendDigits(%d, %d, %d) = %q, want %q", c.n, c.width, c.lang, got, c.want)
		}
	}
}
