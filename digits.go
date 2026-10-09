package nepdate

// appendDigits appends n in decimal, zero-padded to at least minWidth digits, with ASCII digits
// for English and Devanagari digits for Nepali. A Devanagari digit d is the 3 bytes
// 0xE0 0xA5 0xA6+d (U+0966 + d), so no rune conversion is needed.
func appendDigits(b []byte, n, minWidth int, lang Lang) []byte {
	u := uint64(n) //nolint:gosec // the two's complement of a negative n is negated below
	if n < 0 {
		b = append(b, '-')
		u = -u
	}
	var buf [20]byte // 2^64 has 20 decimal digits
	i := len(buf)
	for u >= 10 {
		i--
		buf[i] = byte(u%10) + '0'
		u /= 10
	}
	i--
	buf[i] = byte(u) + '0'
	for pad := minWidth - (len(buf) - i); pad > 0; pad-- {
		b = appendDigit(b, '0', lang)
	}
	for _, c := range buf[i:] {
		b = appendDigit(b, c, lang)
	}
	return b
}

func appendDigit(b []byte, c byte, lang Lang) []byte {
	if lang == Nepali {
		return append(b, 0xE0, 0xA5, 0xA6+c-'0')
	}
	return append(b, c)
}
