package nepdate

// DateOrder is the order of the three numbers in Short (shared/spec/FORMATTING.md §3).
type DateOrder uint8

// The six orders, in C#'s DateFormats order.
const (
	YMD DateOrder = iota
	YDM
	MYD
	MDY
	DYM
	DMY
)

// Separator is the character between the numbers in Short.
type Separator uint8

// The separators: '/', '\', '.', '_', '-' and ' '.
const (
	Slash Separator = iota
	Backslash
	Dot
	Underscore
	Dash
	Space
)

// LongOptions changes Long. The zero value is C#'s default: no weekday, year shown, day
// padded to 2 digits.
type LongOptions struct {
	Weekday  bool // start with the weekday name
	OmitYear bool // leave out ", <year>"
	NoPad    bool // write the day without a leading zero
}

const (
	fieldYear = iota
	fieldMonth
	fieldDay
)

var (
	separatorBytes = [...]byte{Slash: '/', Backslash: 0x5C, Dot: '.', Underscore: '_', Dash: '-', Space: ' '}
	orderFields    = [...][3]uint8{
		YMD: {fieldYear, fieldMonth, fieldDay},
		YDM: {fieldYear, fieldDay, fieldMonth},
		MYD: {fieldMonth, fieldYear, fieldDay},
		MDY: {fieldMonth, fieldDay, fieldYear},
		DYM: {fieldDay, fieldYear, fieldMonth},
		DMY: {fieldDay, fieldMonth, fieldYear},
	}
)

// String returns d as "YYYY/MM/DD" with ASCII digits (FORMATTING §2), or "" for the zero Date.
func (d Date) String() string {
	var buf [16]byte
	return string(d.appendYMD(buf[:0], '/'))
}

// appendYMD writes the 10 bytes "YYYY<sep>MM<sep>DD", or nothing for the zero Date.
func (d Date) appendYMD(b []byte, sep byte) []byte {
	if d.IsZero() {
		return b
	}
	y, m, day := d.year, d.month, d.day
	//nolint:gosec // each operand is one decimal digit, 0..9
	return append(b,
		byte('0'+y/1000), byte('0'+y/100%10), byte('0'+y/10%10), byte('0'+y%10), sep,
		'0'+m/10, '0'+m%10, sep,
		'0'+day/10, '0'+day%10)
}

// Short returns the three numbers of d in order o, separated by sep, in lang's digits. With
// pad the year has 4 digits and the month and day 2. It returns "" for the zero Date or an
// unknown order or separator.
func (d Date) Short(o DateOrder, sep Separator, pad bool, lang Lang) string {
	var buf [96]byte
	return string(d.AppendShort(buf[:0], o, sep, pad, lang))
}

// AppendShort is Short appending to b. It does not allocate when b has room.
func (d Date) AppendShort(b []byte, o DateOrder, sep Separator, pad bool, lang Lang) []byte {
	if d.IsZero() || int(o) >= len(orderFields) || int(sep) >= len(separatorBytes) {
		return b
	}
	for i, f := range orderFields[o] {
		if i > 0 {
			b = append(b, separatorBytes[sep])
		}
		b = d.appendField(b, f, pad, lang)
	}
	return b
}

func (d Date) appendField(b []byte, f uint8, pad bool, lang Lang) []byte {
	switch f {
	case fieldYear:
		return appendDigits(b, int(d.year), padWidth(pad, 4), lang)
	case fieldMonth:
		return appendDigits(b, int(d.month), padWidth(pad, 2), lang)
	default:
		return appendDigits(b, int(d.day), padWidth(pad, 2), lang)
	}
}

func padWidth(pad bool, width int) int {
	if pad {
		return width
	}
	return 0
}

// Long returns d as "[Weekday, ]Month DD[, YYYY]" in lang (FORMATTING §4), or "" for the zero
// Date.
func (d Date) Long(opts LongOptions, lang Lang) string {
	var buf [96]byte
	return string(d.AppendLong(buf[:0], opts, lang))
}

// AppendLong is Long appending to b. It does not allocate when b has room.
func (d Date) AppendLong(b []byte, opts LongOptions, lang Lang) []byte {
	if d.IsZero() {
		return b
	}
	if opts.Weekday {
		b = append(b, WeekdayName(d.Weekday(), lang)...)
		b = append(b, ", "...)
	}
	b = append(b, Month(d.month).Name(lang)...)
	b = append(b, ' ')
	b = appendDigits(b, int(d.day), padWidth(!opts.NoPad, 2), lang)
	if !opts.OmitYear {
		b = append(b, ", "...)
		b = appendDigits(b, int(d.year), 4, lang)
	}
	return b
}
