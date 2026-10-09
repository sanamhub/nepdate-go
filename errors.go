package nepdate

import "errors"

// Sentinel errors. Test for them with errors.Is.
var (
	// ErrOutOfRange reports a date outside BS 1901-01-01..2199-12-30, a Gregorian date outside
	// AD 1844-04-11..2143-04-15, an operation whose result leaves the range, or the zero Date.
	ErrOutOfRange = errors.New("nepdate: date outside BS 1901-01-01..2199-12-30")
	// ErrInvalidMonth reports a month outside 1..12.
	ErrInvalidMonth = errors.New("nepdate: month must be 1..12")
	// ErrInvalidDay reports a day outside 1 to the length of its month.
	ErrInvalidDay = errors.New("nepdate: day outside the month")
	// ErrInvalidGregorian reports an AD date that does not exist, such as 2023-02-29.
	ErrInvalidGregorian = errors.New("nepdate: invalid Gregorian date")
)

// ParseErrorKind says why the syntax step of Parse or ParseLenient failed
// (shared/spec/PARSING.md §4).
type ParseErrorKind uint8

// The parse error kinds.
const (
	InvalidCharacter ParseErrorKind = iota + 1 // a character no rule accepts
	WrongGroupCount                            // not exactly three groups
	NumberTooLong                              // a number with more than 4 digits
	Ambiguous                                  // the input fits more than one reading
	Unrecognized                               // the lenient rules found no date
)

// ParseError is returned when the input is not a date in the accepted syntax. A string that
// parses but names an invalid date returns ErrOutOfRange, ErrInvalidMonth or ErrInvalidDay
// instead.
type ParseError struct {
	Input string
	Kind  ParseErrorKind
}

var parseErrorText = [...]string{
	InvalidCharacter: "invalid character",
	WrongGroupCount:  "wrong number of groups",
	NumberTooLong:    "number longer than 4 digits",
	Ambiguous:        "ambiguous",
	Unrecognized:     "unrecognized",
}

// Error returns a message naming the input and the kind.
func (e *ParseError) Error() string {
	msg := `nepdate: cannot parse "` + e.Input + `"`
	if e.Kind >= InvalidCharacter && int(e.Kind) < len(parseErrorText) {
		msg += ": " + parseErrorText[e.Kind]
	}
	return msg
}
