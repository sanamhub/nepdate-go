package nepdate

// Parse reads a date in the strict form of shared/spec/PARSING.md §1: year, month and day in
// that order, ASCII or Devanagari digits (they may be mixed, PARITY D-05), at most 4 digits per
// number, separated by any of - / . _ \ space । |. It returns a *ParseError when the text is not
// in that form, and New's error when it names an invalid date. It does not allocate on success;
// a failed call allocates one *ParseError.
func Parse(s string) (Date, error) {
	return parseStrict(s)
}

// text is the input of the parsers. They are generic so UnmarshalText can parse its []byte
// without converting it to a string first, which would allocate (ADR-0002 §6).
type text interface{ ~string | ~[]byte }

// Devanagari digits and the danda are the 3 bytes 0xE0 0xA5 x.
const (
	devaLead     = 0xE0
	devaSecond   = 0xA5
	devaZero     = 0xA6 // U+0966
	devaNine     = 0xAF // U+096F
	devaDanda    = 0xA4 // U+0964 ।
	maxDigits    = 4
	strictGroups = 3
)

// devaAt returns the third byte of a Devanagari character (U+0940..U+097F) at s[i], or 0.
func devaAt[T text](s T, i int) byte {
	if i+2 < len(s) && s[i] == devaLead && s[i+1] == devaSecond {
		return s[i+2]
	}
	return 0
}

func isStrictSeparator(c byte) bool {
	switch c {
	case '-', '/', '.', '_', '\\', ' ', '|':
		return true
	}
	return false
}

func parseStrict[T text](s T) (Date, error) {
	d, kind, err := scanStrict(s)
	if kind != 0 {
		return Date{}, parseError(s, kind)
	}
	return d, err
}

// scanStrict is the strict parser without building the *ParseError, so ParseLenient can try it
// first and allocate nothing when it fails. A syntax failure returns its kind; a validation
// failure returns New's error. The canonical ASCII shape takes the fast path; everything else
// runs the general loop.
func scanStrict[T text](s T) (Date, ParseErrorKind, error) {
	if d, ok, err := scanCanonical(s); ok {
		return d, 0, err
	}
	return scanStrictLoop(s)
}

// scanStrictLoop is the general strict scanner: any strict separator, one to four digits per
// group, ASCII or Devanagari digits. It is the reference for the fast path's differential tests.
func scanStrictLoop[T text](s T) (Date, ParseErrorKind, error) {
	var groups [strictGroups]int
	n := 0 // groups started
	inGroup := false
	digits := 0 // digits in the current group
	for i := 0; i < len(s); {
		c := s[i]
		third := devaAt(s, i)
		var d int
		//nolint:gocritic // an if chain, not a switch, so Gremlins can test each condition (G9-03)
		if c >= '0' && c <= '9' {
			d = int(c - '0')
			i++
		} else if third >= devaZero && third <= devaNine {
			d = int(third - devaZero)
			i += 3
		} else if third == devaDanda {
			inGroup = false
			i += 3
			continue
		} else if isStrictSeparator(c) {
			inGroup = false
			i++
			continue
		} else {
			return Date{}, InvalidCharacter, nil
		}
		if !inGroup {
			if n == strictGroups {
				return Date{}, WrongGroupCount, nil
			}
			n++
			inGroup = true
			digits = 0
		}
		digits++
		if digits > maxDigits {
			return Date{}, NumberTooLong, nil
		}
		groups[n-1] = groups[n-1]*10 + d
	}
	if n < strictGroups {
		return Date{}, WrongGroupCount, nil
	}
	d, err := newFromInts(groups[0], groups[1], groups[2])
	return d, 0, err
}

// scanCanonical is the fast path for the 10-byte ASCII shape YYYY/MM/DD, with any strict
// separator. It reports false when the shape does not match, and the caller runs the general
// loop. The digit tests are byte comparisons, so a non-digit, a Devanagari digit or a danda falls
// through. newFromInts keeps the validation order and the errors of the general loop.
func scanCanonical[T text](s T) (Date, bool, error) {
	if len(s) != 10 {
		return Date{}, false, nil
	}
	if !isStrictSeparator(s[4]) || !isStrictSeparator(s[7]) {
		return Date{}, false, nil
	}
	y0, y1, y2, y3 := s[0], s[1], s[2], s[3]
	m0, m1 := s[5], s[6]
	d0, d1 := s[8], s[9]
	if y0-'0' > 9 || y1-'0' > 9 || y2-'0' > 9 || y3-'0' > 9 ||
		m0-'0' > 9 || m1-'0' > 9 || d0-'0' > 9 || d1-'0' > 9 {
		return Date{}, false, nil
	}
	year := int(y0-'0')*1000 + int(y1-'0')*100 + int(y2-'0')*10 + int(y3-'0')
	month := int(m0-'0')*10 + int(m1-'0')
	day := int(d0-'0')*10 + int(d1-'0')
	d, err := newFromInts(year, month, day)
	return d, true, err
}

// newFromInts is New for numbers that may not fit a Month: it keeps New's error order (year,
// then month, then day) without truncating a month like 257 to a uint8.
func newFromInts(y, m, d int) (Date, error) {
	if y < minYear || y > maxYear {
		return Date{}, ErrOutOfRange
	}
	if m < int(Baishakh) || m > int(Chaitra) {
		return Date{}, ErrInvalidMonth
	}
	return New(y, Month(m), d) //nolint:gosec // m is 1..12
}

// parseError builds the error of a failed parse. It is the only allocation of the parsers, and
// only on failure.
func parseError[T text](s T, kind ParseErrorKind) error {
	return &ParseError{Input: string(s), Kind: kind}
}
