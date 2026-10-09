package nepdate

import "testing"

func TestLangAppendDigits(t *testing.T) {
	cases := []struct {
		lang     Lang
		n, width int
		want     string
	}{
		{Nepali, 2081, 0, "२०८१"},
		{Nepali, 5, 2, "०५"},
		{English, 5, 2, "05"},
		{English, 2081, 4, "2081"},
		{English, 7, 4, "0007"},
		{Nepali, 1234567890, 0, "१२३४५६७८९०"},
	}
	for _, c := range cases {
		if got := string(c.lang.AppendDigits([]byte("x"), c.n, c.width)); got != "x"+c.want {
			t.Errorf("%d.AppendDigits(%d, %d) = %q, want %q", c.lang, c.n, c.width, got, "x"+c.want)
		}
	}
}

func TestFormatGolden(t *testing.T) {
	rows := readTable(t, "shared/spec/vectors/csharp-golden/dates.tsv", "\t", 1018)
	for _, row := range rows {
		d := bsDate(t, row["bs"])
		got := map[string]string{
			"to_string":       d.String(),
			"unicode":         d.Short(YMD, Slash, true, Nepali),
			"long_en":         d.Long(LongOptions{}, English),
			"long_en_weekday": d.Long(LongOptions{Weekday: true}, English),
			"long_ne":         d.Long(LongOptions{}, Nepali),
			"long_ne_weekday": d.Long(LongOptions{Weekday: true}, Nepali),
			"dmy_dash_nopad":  d.Short(DMY, Dash, false, English),
		}
		for col, g := range got {
			if g != row[col] {
				t.Errorf("%s %s = %q, want %q", row["bs"], col, g, row[col])
			}
		}
	}
}

func TestShortExamples(t *testing.T) {
	d := MustNew(2080, Bhadra, 15)
	cases := []struct {
		o    DateOrder
		sep  Separator
		pad  bool
		lang Lang
		want string
	}{
		// FORMATTING §3
		{YMD, Dash, true, English, "2080-05-15"},
		{DMY, Dash, false, English, "15-5-2080"},
		{YMD, Slash, true, Nepali, "२०८०/०५/१५"},
		// every order and separator
		{YDM, Backslash, true, English, `2080\15\05`},
		{MYD, Dot, true, English, "05.2080.15"},
		{MDY, Underscore, false, English, "5_15_2080"},
		{DYM, Space, true, English, "15 2080 05"},
		{DMY, Slash, true, English, "15/05/2080"},
		// unknown order or separator: nothing
		{DateOrder(6), Slash, true, English, ""},
		{YMD, Separator(6), true, English, ""},
	}
	for _, c := range cases {
		if got := d.Short(c.o, c.sep, c.pad, c.lang); got != c.want {
			t.Errorf("Short(%d, %d, %v, %d) = %q, want %q", c.o, c.sep, c.pad, c.lang, got, c.want)
		}
	}
}

func TestLongExamples(t *testing.T) {
	cases := []struct {
		d      Date
		opts   LongOptions
		en, ne string
	}{
		// FORMATTING §4
		{MustNew(2081, Shrawan, 15), LongOptions{}, "Shrawan 15, 2081", "साउन १५, २०८१"},
		{MustNew(2081, Shrawan, 15), LongOptions{Weekday: true}, "Tuesday, Shrawan 15, 2081", "मङ्गलवार, साउन १५, २०८१"},
		{MustNew(1901, Baishakh, 1), LongOptions{}, "Baishakh 01, 1901", "बैशाख ०१, १९०१"},
		{MustNew(2079, Jestha, 6), LongOptions{Weekday: true, OmitYear: true, NoPad: true}, "Friday, Jestha 6", "शुक्रवार, जेठ ६"},
	}
	for _, c := range cases {
		if got := c.d.Long(c.opts, English); got != c.en {
			t.Errorf("%v.Long(%+v, English) = %q, want %q", c.d, c.opts, got, c.en)
		}
		if got := c.d.Long(c.opts, Nepali); got != c.ne {
			t.Errorf("%v.Long(%+v, Nepali) = %q, want %q", c.d, c.opts, got, c.ne)
		}
	}
}

func TestFormatZeroDate(t *testing.T) {
	var z Date
	if z.String() != "" || z.Short(YMD, Dash, true, English) != "" || z.Long(LongOptions{}, English) != "" {
		t.Error("the zero Date formats as something other than \"\" (D-01)")
	}
	if got := z.AppendLong([]byte("x"), LongOptions{}, Nepali); string(got) != "x" {
		t.Errorf("zero AppendLong = %q", got)
	}
}

func TestLongestOutputFitsBuffer(t *testing.T) {
	// The longest output, Nepali long with weekday, must fit the 96-byte stack buffer.
	longest := 0
	for s := int32(0); s <= maxSerial; s += 7 {
		for w := range 7 {
			if s+int32(w) > maxSerial {
				break
			}
			n := len(fromSerial(s+int32(w)).Long(LongOptions{Weekday: true}, Nepali))
			longest = max(longest, n)
		}
	}
	if longest > 96 {
		t.Errorf("longest Long output is %d bytes", longest)
	}
	t.Logf("longest Nepali long format with weekday: %d bytes", longest)
}
