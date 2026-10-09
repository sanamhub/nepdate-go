package nepdate

// FiscalYear is a Nepal fiscal year, Shrawan 1 to the end of Ashad, named by its starting BS
// year: FY 2082/83 is FiscalYear(2082) (ALGORITHM §8b, PARITY D-08).
type FiscalYear int

// Quarter is a quarter of the fiscal year, Q1 = Shrawan to Ashoj up to Q4 = Baishakh to Ashad.
type Quarter uint8

// The fiscal quarters.
const (
	Q1 Quarter = 1 + iota
	Q2
	Q3
	Q4
)

// FiscalYearOf returns the fiscal year that contains d, or 0 for the zero Date.
func FiscalYearOf(d Date) FiscalYear {
	if d.IsZero() {
		return 0
	}
	if d.month >= uint8(Shrawan) {
		return FiscalYear(d.year)
	}
	return FiscalYear(d.year) - 1
}

// FiscalYear returns the fiscal year that contains d, or 0 for the zero Date.
func (d Date) FiscalYear() FiscalYear { return FiscalYearOf(d) }

// Quarter returns the fiscal quarter that contains d, or 0 for the zero Date.
func (d Date) Quarter() Quarter {
	if d.IsZero() {
		return 0
	}
	if d.month >= uint8(Shrawan) {
		return Quarter((d.month-uint8(Shrawan))/3) + Q1
	}
	return Q4
}

// Start returns Shrawan 1 of f. It returns ErrOutOfRange if that day is outside the range.
func (f FiscalYear) Start() (Date, error) {
	return New(int(f), Shrawan, 1)
}

// End returns the last day of Ashad of f+1. It returns ErrOutOfRange if that day is outside the
// range.
func (f FiscalYear) End() (Date, error) {
	first, err := New(int(f)+1, Ashad, 1)
	return first.LastOfMonth(), err
}

// Range returns every day of f. It returns ErrOutOfRange if either end is outside the range.
func (f FiscalYear) Range() (Range, error) {
	start, err := f.Start()
	if err != nil {
		return Range{}, err
	}
	end, err := f.End()
	if err != nil {
		return Range{}, err
	}
	return Range{start, end}, nil
}

// QuarterRange returns the days of quarter q of f. It returns ErrOutOfRange if the quarter is
// outside the supported range, and also if q is not Q1..Q4: ErrOutOfRange is the error for any
// argument outside the supported domain.
func (f FiscalYear) QuarterRange(q Quarter) (Range, error) {
	if q < Q1 || q > Q4 {
		return Range{}, ErrOutOfRange
	}
	months := q.Months()
	year := int(f)
	if q == Q4 {
		year++
	}
	first, err := New(year, months[0], 1)
	if err != nil {
		return Range{}, err
	}
	last, err := New(year, months[2], 1)
	if err != nil {
		return Range{}, err
	}
	return Range{first, last.LastOfMonth()}, nil
}

// Label returns the usual name of f, "2082/83" in English or "२०८२/८३" in Nepali.
func (f FiscalYear) Label(lang Lang) string {
	var buf [32]byte
	b := appendDigits(buf[:0], int(f), 0, lang)
	b = append(b, '/')
	next := (int(f) + 1) % 100
	if next < 0 {
		next = -next
	}
	return string(appendDigits(b, next, 2, lang))
}

// Months returns the three months of q in order, or three zero Months if q is not Q1..Q4.
func (q Quarter) Months() [3]Month {
	if q < Q1 || q > Q4 {
		return [3]Month{}
	}
	first := Shrawan + 3*Month(q-Q1)
	if q == Q4 {
		first = Baishakh
	}
	return [3]Month{first, first + 1, first + 2}
}
