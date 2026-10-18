package nepdate

// Format returns d formatted by pattern (shared/spec/FORMATTING.md §5) in lang, or "" for the
// zero Date. The single-letter patterns "", "G", "g" and "d" give "YYYY/MM/DD", "D" the default
// Long format and "s" "YYYY-MM-DD". Any other pattern is read left to right: runs of y, M and d
// are the year, month and day; '...' and \x are literal; everything else is copied.
func (d Date) Format(pattern string, lang Lang) string {
	var buf [96]byte
	return string(d.AppendFormat(buf[:0], pattern, lang))
}

// AppendFormat is Format appending to b. It does not allocate when b has room.
func (d Date) AppendFormat(b []byte, pattern string, lang Lang) []byte {
	if d.IsZero() {
		return b
	}
	switch pattern {
	case "", "G", "g", "d":
		return d.AppendShort(b, YMD, Slash, true, lang)
	case "D":
		return d.AppendLong(b, LongOptions{}, lang)
	case "s":
		return d.AppendShort(b, YMD, Dash, true, lang)
	}
	for i := 0; i < len(pattern); {
		c := pattern[i]
		switch c {
		case '\\':
			// The escaped byte is copied; the rest of a multi-byte character follows as-is.
			if i+1 < len(pattern) {
				b = append(b, pattern[i+1])
			}
			i += 2
		case '\'':
			end := i + 1
			for end < len(pattern) && pattern[end] != '\'' {
				end++
			}
			b = append(b, pattern[i+1:end]...)
			i = end + 1
		case 'y', 'M', 'd':
			n := 1
			for i+n < len(pattern) && pattern[i+n] == c {
				n++
			}
			b = d.appendToken(b, c, n, lang)
			i += n
		default:
			b = append(b, c)
			i++
		}
	}
	return b
}

// appendToken writes a run of n letters c (y, M or d).
func (d Date) appendToken(b []byte, c byte, n int, lang Lang) []byte {
	// An if chain, not a switch: Gremlins finds no coverage block for a case expression, so it
	// cannot test one (G9-03).
	if c == 'y' && n >= 4 {
		return appendDigits(b, int(d.year), 4, lang)
	}
	if c == 'y' {
		return appendDigits(b, int(d.year)%100, 2, lang)
	}
	if c == 'M' && n >= 4 {
		return append(b, Month(d.month).Name(lang)...)
	}
	if c == 'M' && n == 3 {
		return append(b, Month(d.month).ShortName(lang)...)
	}
	if c == 'M' {
		return appendDigits(b, int(d.month), n, lang)
	}
	if n == 2 { // dd
		return appendDigits(b, int(d.day), 2, lang)
	}
	// d, ddd, dddd...: unpadded day
	return appendDigits(b, int(d.day), 0, lang)
}
