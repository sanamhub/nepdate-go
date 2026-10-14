package nepdate

// ParseLenient reads a date written by a person (shared/spec/PARSING.md §2), such as
// "15 Shrawan 2080", "Shrawan 15, 2080", "१५ श्रावण २०८०" or "15.04.2080". It tries Parse first.
// Otherwise it splits the text into numbers and words, drops era markers (BS, वि.सं.) and the
// words गते and मिति, and accepts a closed list of month names (PARSING §3) or three numbers
// with the 4-digit year first or last. It never guesses (PARITY D-06): a 2-digit year is
// Ambiguous. It returns a *ParseError when no rule matches, and New's error when the date is
// invalid.
func ParseLenient(s string) (Date, error) {
	if d, kind, err := scanStrict(s); kind == 0 && err == nil {
		return d, nil
	}
	var toks lenientTokens
	if kind := toks.split(s); kind != 0 {
		return Date{}, parseError(s, kind)
	}
	return toks.date(s)
}

// lenientToken is a number or a month name; other words are only counted.
type lenientToken struct {
	value  int // the number, or the month
	digits int // digits in the number; 0 for a month name
}

// maxLenientTokens bounds the token array, which lives on the stack. A date needs at most
// three; more than eight is Unrecognized anyway (PARSING §2 step 5).
const maxLenientTokens = 8

type lenientTokens struct {
	toks    [maxLenientTokens]lenientToken
	n       int
	unknown bool // a word that is neither a month, an era marker nor गते/मिति
	tooMany bool
}

// split fills t from s and returns NumberTooLong for a number of more than 4 digits, or 0.
func (t *lenientTokens) split(s string) ParseErrorKind {
	for i := 0; i < len(s); {
		w := lenientSeparator(s, i)
		if w > 0 {
			i += w
			continue
		}
		end := i
		for end < len(s) && lenientSeparator(s, end) == 0 {
			end++
		}
		if kind := t.chunk(s[i:end]); kind != 0 {
			return kind
		}
		i = end
	}
	return 0
}

// chunk handles the text between two separators. Digits and dots only: numbers split on the
// dots ("15.04.2080"). Anything else is one word, dots and digits included, so "B.S." is an
// era marker and "Shr.15" or "15Shrawan" match nothing (PARSING §2).
func (t *lenientTokens) chunk(c string) ParseErrorKind {
	for i := 0; i < len(c); {
		if c[i] == '.' {
			i++
			continue
		}
		_, w := digitAt(c, i)
		if w == 0 {
			t.word(c)
			return 0
		}
		i += w
	}
	for i := 0; i < len(c); {
		if c[i] == '.' {
			i++
			continue
		}
		value, digits := 0, 0
		for i < len(c) && c[i] != '.' {
			d, w := digitAt(c, i)
			value = value*10 + d
			digits++
			i += w
		}
		if digits > maxDigits {
			return NumberTooLong
		}
		t.add(lenientToken{value: value, digits: digits})
	}
	return 0
}

func (t *lenientTokens) add(tok lenientToken) {
	if t.n == maxLenientTokens {
		t.tooMany = true
		return
	}
	t.toks[t.n] = tok
	t.n++
}

func (t *lenientTokens) word(w string) {
	for _, m := range ignoredWords {
		if foldEqual(w, m) {
			return
		}
	}
	for _, m := range monthWords {
		if foldEqual(w, m.name) {
			t.add(lenientToken{value: int(m.month)})
			return
		}
	}
	t.unknown = true
}

// date applies PARSING §2 steps 3 to 6 to the tokens.
func (t *lenientTokens) date(s string) (Date, error) {
	if t.unknown || t.tooMany {
		return Date{}, parseError(s, Unrecognized)
	}
	var nums [3]lenientToken
	nNums, nMonths, month := 0, 0, 0
	for _, tok := range t.toks[:t.n] {
		if tok.digits == 0 {
			nMonths++
			month = tok.value
			continue
		}
		if nNums == len(nums) {
			return Date{}, parseError(s, Unrecognized)
		}
		nums[nNums] = tok
		nNums++
	}
	switch {
	case nMonths == 1 && nNums == 2:
		a, b := nums[0], nums[1]
		if b.digits == 4 {
			a, b = b, a
		}
		if a.digits != 4 || b.digits > 2 {
			return Date{}, parseError(s, Ambiguous)
		}
		return newFromInts(a.value, month, b.value)
	case nMonths == 0 && nNums == 3:
		switch {
		case nums[0].digits == 4:
			return newFromInts(nums[0].value, nums[1].value, nums[2].value)
		case nums[2].digits == 4:
			return newFromInts(nums[2].value, nums[1].value, nums[0].value)
		default:
			return Date{}, parseError(s, Ambiguous)
		}
	default:
		return Date{}, parseError(s, Unrecognized)
	}
}

// The Kelvin sign U+212A is the 3 bytes 0xE2 0x84 0xAA.
const (
	kelvinLead   = 0xE2
	kelvinSecond = 0x84
	kelvinThird  = 0xAA
)

// foldEqual reports whether token equals entry under the folding of PARSING §2: A to Z fold to
// lower case and the Kelvin sign folds to 'k', which is what Python str.lower() maps onto ASCII.
// Nothing else folds, so unlike strings.EqualFold the long s U+017F does not match 's'. Every
// entry is already lower case.
func foldEqual(token, entry string) bool {
	j := 0
	for i := 0; i < len(token); {
		c := token[i]
		width := 1
		switch {
		case c >= 'A' && c <= 'Z':
			c += 'a' - 'A'
		case c == kelvinLead && i+2 < len(token) && token[i+1] == kelvinSecond &&
			token[i+2] == kelvinThird:
			c, width = 'k', 3
		}
		if j >= len(entry) || entry[j] != c {
			return false
		}
		i += width
		j++
	}
	return j == len(entry)
}

// digitAt returns the value and byte width of an ASCII or Devanagari digit at s[i], or width 0.
func digitAt(s string, i int) (value, width int) {
	if c := s[i]; c >= '0' && c <= '9' {
		return int(c - '0'), 1
	}
	if third := devaAt(s, i); third >= devaZero && third <= devaNine {
		return int(third - devaZero), 3
	}
	return 0, 0
}

// lenientSeparator returns the byte width of a separator at s[i], or 0: the strict
// separators except '.', plus ','.
func lenientSeparator(s string, i int) int {
	switch c := s[i]; {
	case c == ',':
		return 1
	case c != '.' && isStrictSeparator(c):
		return 1
	case devaAt(s, i) == devaDanda:
		return 3
	}
	return 0
}

// ignoredWords are dropped by the lenient parser: era markers, गते (day) and मिति (date). Latin
// entries are lower case, as foldEqual needs.
var ignoredWords = [...]string{
	"bs", "b.s.", "b.s", "vs", "v.s.", "v.s",
	"बि.सं.", "वि.सं.", "बि.सं", "वि.सं",
	"गते", "मिति",
}

// monthWords is the closed month-name list of PARSING §3. Latin names match in any case.
var monthWords = [...]struct {
	name  string
	month Month
}{
	{"baishakh", Baishakh},
	{"baisakh", Baishakh},
	{"baishak", Baishakh},
	{"vaishakh", Baishakh},
	{"vaisakh", Baishakh},
	{"bai", Baishakh},
	{"बैशाख", Baishakh},
	{"वैशाख", Baishakh},
	{"बैसाख", Baishakh},
	{"jestha", Jestha},
	{"jeth", Jestha},
	{"jyestha", Jestha},
	{"jes", Jestha},
	{"जेठ", Jestha},
	{"जेष्ठ", Jestha},
	{"ज्येष्ठ", Jestha},
	{"ashad", Ashad},
	{"ashadh", Ashad},
	{"asar", Ashad},
	{"asadh", Ashad},
	{"asa", Ashad},
	{"असार", Ashad},
	{"आषाढ", Ashad},
	{"असाढ", Ashad},
	{"shrawan", Shrawan},
	{"shravan", Shrawan},
	{"srawan", Shrawan},
	{"sawan", Shrawan},
	{"saun", Shrawan},
	{"shr", Shrawan},
	{"साउन", Shrawan},
	{"श्रावण", Shrawan},
	{"सावन", Shrawan},
	{"bhadra", Bhadra},
	{"bhadau", Bhadra},
	{"bha", Bhadra},
	{"भदौ", Bhadra},
	{"भाद्र", Bhadra},
	{"भाद्रपद", Bhadra},
	{"ashoj", Ashoj},
	{"asoj", Ashoj},
	{"ashwin", Ashoj},
	{"aso", Ashoj},
	{"असोज", Ashoj},
	{"आश्विन", Ashoj},
	{"असौज", Ashoj},
	{"kartik", Kartik},
	{"kattik", Kartik},
	{"kar", Kartik},
	{"कार्तिक", Kartik},
	{"कात्तिक", Kartik},
	{"mangsir", Mangsir},
	{"mansir", Mangsir},
	{"margashirsha", Mangsir},
	{"man", Mangsir},
	{"मंसिर", Mangsir},
	{"मङ्सिर", Mangsir},
	{"मार्गशीर्ष", Mangsir},
	{"poush", Poush},
	{"paush", Poush},
	{"push", Poush},
	{"pus", Poush},
	{"pou", Poush},
	{"पुष", Poush},
	{"पौष", Poush},
	{"पुस", Poush},
	{"magh", Magh},
	{"mag", Magh},
	{"माघ", Magh},
	{"falgun", Falgun},
	{"phalgun", Falgun},
	{"fagun", Falgun},
	{"fal", Falgun},
	{"फागुन", Falgun},
	{"फाल्गुन", Falgun},
	{"chaitra", Chaitra},
	{"chait", Chaitra},
	{"cha", Chaitra},
	{"चैत", Chaitra},
	{"चैत्र", Chaitra},
}
