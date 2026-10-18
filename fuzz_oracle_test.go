package nepdate

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
)

// The reference formatters below are written from shared/spec/FORMATTING.md. They do not call the
// production formatters, so a change in the package cannot make both sides agree by construction.

var (
	refMonthNames = [2][12]string{
		{"Baishakh", "Jestha", "Ashad", "Shrawan", "Bhadra", "Ashoj", "Kartik", "Mangsir", "Poush", "Magh", "Falgun", "Chaitra"},
		{"बैशाख", "जेठ", "असार", "साउन", "भदौ", "असोज", "कार्तिक", "मंसिर", "पुष", "माघ", "फागुन", "चैत"},
	}
	refMonthShort = [12]string{"Bai", "Jes", "Asa", "Shr", "Bha", "Aso", "Kar", "Man", "Pou", "Mag", "Fal", "Cha"}
	refWeekdays   = [2][7]string{
		{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"},
		{"आइतवार", "सोमवार", "मङ्गलवार", "बुधवार", "बिहिवार", "शुक्रवार", "शनिवार"},
	}
)

// refDigits maps ASCII digits to Devanagari for Nepali and leaves every other byte alone.
func refDigits(s string, lang Lang) string {
	if lang != Nepali {
		return s
	}
	var out []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			out = append(out, 0xE0, 0xA5, 0xA6+c-'0')
			continue
		}
		out = append(out, c)
	}
	return string(out)
}

// refNumber is appendDigits for a non-negative number, built with strconv and refDigits.
func refNumber(n, minWidth int, lang Lang) string {
	s := strconv.Itoa(n)
	for len(s) < minWidth {
		s = "0" + s
	}
	return refDigits(s, lang)
}

// refShort is FORMATTING §3.
func refShort(d Date, o DateOrder, sep Separator, pad bool, lang Lang) string {
	orders := [...][3]int{
		{0, 1, 2},
		{0, 2, 1},
		{1, 0, 2},
		{1, 2, 0},
		{2, 0, 1},
		{2, 1, 0},
	}
	seps := [...]byte{'/', '\\', '.', '_', '-', ' '}
	if d.IsZero() || int(o) >= len(orders) || int(sep) >= len(seps) {
		return ""
	}
	values := [...]int{d.Year(), int(d.Month()), d.Day()}
	widths := [...]int{4, 2, 2}
	var out []byte
	for i, f := range orders[o] {
		if i > 0 {
			out = append(out, seps[sep])
		}
		width := 0
		if pad {
			width = widths[f]
		}
		out = append(out, refNumber(values[f], width, lang)...)
	}
	return string(out)
}

// refLong is FORMATTING §4.
func refLong(d Date, opts LongOptions, lang Lang) string {
	if d.IsZero() {
		return ""
	}
	var out []byte
	if opts.Weekday {
		out = append(out, refWeekdays[lang][d.Weekday()]...)
		out = append(out, ", "...)
	}
	out = append(out, refMonthNames[lang][int(d.Month())-1]...)
	out = append(out, ' ')
	dayWidth := 2
	if opts.NoPad {
		dayWidth = 0
	}
	out = append(out, refNumber(d.Day(), dayWidth, lang)...)
	if !opts.OmitYear {
		out = append(out, ", "...)
		out = append(out, refNumber(d.Year(), 4, lang)...)
	}
	return string(out)
}

// refPattern is FORMATTING §5, including the standard single-letter patterns, escapes, quotes and
// the y, M and d runs.
func refPattern(d Date, pattern string, lang Lang) string {
	if d.IsZero() {
		return ""
	}
	switch pattern {
	case "", "G", "g", "d":
		return refShort(d, YMD, Slash, true, lang)
	case "D":
		return refLong(d, LongOptions{}, lang)
	case "s":
		return refShort(d, YMD, Dash, true, lang)
	}
	var out []byte
	for i := 0; i < len(pattern); {
		c := pattern[i]
		switch c {
		case '\\':
			if i+1 < len(pattern) {
				out = append(out, pattern[i+1])
			}
			i += 2
		case '\'':
			end := i + 1
			for end < len(pattern) && pattern[end] != '\'' {
				end++
			}
			out = append(out, pattern[i+1:end]...)
			i = end + 1
		case 'y', 'M', 'd':
			n := 1
			for i+n < len(pattern) && pattern[i+n] == c {
				n++
			}
			out = append(out, refToken(d, c, n, lang)...)
			i += n
		default:
			out = append(out, c)
			i++
		}
	}
	return string(out)
}

// refToken is one y, M or d run for refPattern.
func refToken(d Date, c byte, n int, lang Lang) string {
	switch c {
	case 'y':
		if n >= 4 {
			return refNumber(d.Year(), 4, lang)
		}
		return refNumber(d.Year()%100, 2, lang)
	case 'M':
		if n >= 4 {
			return refMonthNames[lang][int(d.Month())-1]
		}
		if n == 3 {
			if lang == Nepali {
				return refMonthNames[lang][int(d.Month())-1]
			}
			return refMonthShort[int(d.Month())-1]
		}
		return refNumber(int(d.Month()), n, lang)
	default:
		if n == 2 {
			return refNumber(d.Day(), 2, lang)
		}
		return refNumber(d.Day(), 0, lang)
	}
}

// refText is the text encoding of D-11, built with fmt so it is not the production writer.
func refText(d Date) string {
	if d.IsZero() {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", d.Year(), int(d.Month()), d.Day())
}

// refAppendDigits is the ADR-0002 digit rule: sign, then the decimal digits zero-padded to
// minWidth digits, then the Devanagari mapping for Nepali.
func refAppendDigits(n, minWidth int, lang Lang) string {
	u := uint64(n) //nolint:gosec // the magnitude is fixed below for a negative n
	sign := ""
	if n < 0 {
		sign = "-"
		u = -u
	}
	digits := strconv.FormatUint(u, 10)
	for len(digits) < minWidth {
		digits = "0" + digits
	}
	return sign + refDigits(digits, lang)
}

// FuzzFormatRoundTrip checks Short against refShort for every order, separator, padding and
// language, then parses the YMD output back. The other orders are checked against the oracle only,
// because Parse accepts YMD.
func FuzzFormatRoundTrip(f *testing.F) {
	f.Add(int32(-1), uint8(0), uint8(0), true, uint8(0))
	f.Add(int32(65746+105), uint8(0), uint8(0), true, uint8(0))
	f.Add(int32(0), uint8(0), uint8(5), false, uint8(1))
	f.Add(int32(maxSerial), uint8(0), uint8(1), true, uint8(1))
	for o := range 6 {
		f.Add(int32(65746), uint8(o), uint8(o), o%2 == 0, uint8(o%2))
	}
	f.Fuzz(func(t *testing.T, serial int32, order, sep uint8, pad bool, lang uint8) {
		if serial < -1 || serial > maxSerial {
			return
		}
		var d Date
		if serial >= 0 {
			d = fromSerial(serial)
		}
		o, sp, l := DateOrder(order%7), Separator(sep%7), Lang(lang%2)
		got := d.Short(o, sp, pad, l)
		want := refShort(d, o, sp, pad, l)
		if got != want {
			t.Fatalf("%v.Short(%d, %d, %v, %d) = %q, want %q", d, o, sp, pad, l, got, want)
		}
		if got := string(d.AppendShort([]byte("x"), o, sp, pad, l)); got != "x"+want {
			t.Fatalf("%v.AppendShort = %q, want %q", d, got, "x"+want)
		}
		if o != YMD || got == "" {
			return
		}
		back, err := Parse(got)
		if err != nil || back != d {
			t.Fatalf("Parse(%q) = %v, %v; want %v", got, back, err, d)
		}
	})
}

// FuzzLong checks Long and AppendLong against refLong for both languages and every option.
func FuzzLong(f *testing.F) {
	f.Add(int32(-1), uint8(0), uint8(0))
	f.Add(int32(0), uint8(0), uint8(0))
	f.Add(int32(65746), uint8(7), uint8(1))
	f.Add(int32(maxSerial), uint8(4), uint8(0))
	f.Fuzz(func(t *testing.T, serial int32, opts uint8, lang uint8) {
		if serial < -1 || serial > maxSerial {
			return
		}
		var d Date
		if serial >= 0 {
			d = fromSerial(serial)
		}
		o := LongOptions{Weekday: opts&1 != 0, OmitYear: opts&2 != 0, NoPad: opts&4 != 0}
		l := Lang(lang % 2)
		got := d.Long(o, l)
		want := refLong(d, o, l)
		if got != want {
			t.Fatalf("%v.Long(%+v, %d) = %q, want %q", d, o, l, got, want)
		}
		if got := string(d.AppendLong([]byte("x"), o, l)); got != "x"+want {
			t.Fatalf("%v.AppendLong = %q, want %q", d, got, "x"+want)
		}
	})
}

// FuzzFormatPattern checks Format and AppendFormat against refPattern for arbitrary patterns.
func FuzzFormatPattern(f *testing.F) {
	f.Add(int32(65746), "yyyy-MM-dd", uint8(0))
	f.Add(int32(65746), "M/d/yyyy", uint8(0))
	f.Add(int32(65746), `yyyy\M\M-\d\d`, uint8(1))
	f.Add(int32(maxSerial), "dd.MM.yy", uint8(0))
	f.Add(int32(0), "D", uint8(1))
	f.Add(int32(-1), "yyyy", uint8(0))
	f.Fuzz(func(t *testing.T, serial int32, pattern string, lang uint8) {
		if serial < -1 || serial > maxSerial {
			return
		}
		var d Date
		if serial >= 0 {
			d = fromSerial(serial)
		}
		l := Lang(lang % 2)
		got := d.Format(pattern, l)
		want := refPattern(d, pattern, l)
		if got != want {
			t.Fatalf("%v.Format(%q, %d) = %q, want %q", d, pattern, l, got, want)
		}
		if got := string(d.AppendFormat([]byte("x"), pattern, l)); got != "x"+want {
			t.Fatalf("%v.AppendFormat(%q) = %q, want %q", d, pattern, got, "x"+want)
		}
	})
}

// FuzzText checks String, MarshalText and AppendText against a fmt oracle, then parses the text
// back through UnmarshalText.
func FuzzText(f *testing.F) {
	f.Add(int32(-1))
	f.Add(int32(0))
	f.Add(int32(65746))
	f.Add(int32(maxSerial))
	f.Fuzz(func(t *testing.T, serial int32) {
		if serial < -1 || serial > maxSerial {
			return
		}
		var d Date
		if serial >= 0 {
			d = fromSerial(serial)
		}
		wantString := ""
		if !d.IsZero() {
			wantString = fmt.Sprintf("%04d/%02d/%02d", d.Year(), int(d.Month()), d.Day())
		}
		if got := d.String(); got != wantString {
			t.Fatalf("%v.String() = %q, want %q", d, got, wantString)
		}
		wantText := refText(d)
		text, err := d.MarshalText()
		if d.IsZero() {
			if !errors.Is(err, ErrOutOfRange) || text != nil {
				t.Fatalf("zero MarshalText = %q, %v", text, err)
			}
		} else if err != nil || string(text) != wantText {
			t.Fatalf("%v.MarshalText() = %q, %v; want %q", d, text, err, wantText)
		}
		appended, err := d.AppendText([]byte("x"))
		if d.IsZero() {
			if string(appended) != "x" || !errors.Is(err, ErrOutOfRange) {
				t.Fatalf("zero AppendText = %q, %v", appended, err)
			}
		} else if string(appended) != "x"+wantText || err != nil {
			t.Fatalf("%v.AppendText = %q, %v; want %q", d, appended, err, "x"+wantText)
		}
		if d.IsZero() {
			return
		}
		var back Date
		if err := back.UnmarshalText(text); err != nil || back != d {
			t.Fatalf("UnmarshalText(%q) = %v, %v; want %v", text, back, err, d)
		}
	})
}

// FuzzAppendDigits checks AppendDigits against refAppendDigits, including signs, padding and the
// invalid Lang values that fall back to ASCII.
func FuzzAppendDigits(f *testing.F) {
	f.Add(0, uint8(0), uint8(0))
	f.Add(2081, uint8(4), uint8(1))
	f.Add(-12, uint8(5), uint8(0))
	f.Add(-1, uint8(0), uint8(2))
	f.Add(1234567890, uint8(12), uint8(1))
	f.Fuzz(func(t *testing.T, n int, width uint8, lang uint8) {
		l := Lang(lang % 3)
		got := string(l.AppendDigits([]byte("x"), n, int(width)))
		want := "x" + refAppendDigits(n, int(width), l)
		if got != want {
			t.Fatalf("%d.AppendDigits(%d, %d) = %q, want %q", l, n, width, got, want)
		}
	})
}
