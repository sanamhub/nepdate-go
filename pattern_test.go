package nepdate

import (
	"strings"
	"testing"
)

func TestFormatPatternGolden(t *testing.T) {
	rows := readTable(t, "shared/spec/vectors/csharp-golden/format-pattern.tsv", "\t", 451)
	skipped := 0
	for _, row := range rows {
		d := bsDate(t, row["bs"])
		// D-04: C# abbreviates Ashad and Ashoj both as "Ash"; we write Asa and Aso.
		if strings.Contains(row["pattern"], "MMM") && !strings.Contains(row["pattern"], "MMMM") &&
			(d.Month() == Ashad || d.Month() == Ashoj) {
			skipped++
			continue
		}
		if got := d.Format(row["pattern"], English); got != row["result"] {
			t.Errorf("%s Format(%q) = %q, want %q", row["bs"], row["pattern"], got, row["result"])
		}
	}
	t.Logf("%d rows skipped for D-04", skipped)
}

func TestFormatPatternExamples(t *testing.T) {
	d := MustNew(2081, Shrawan, 15)
	cases := []struct {
		pattern string
		lang    Lang
		want    string
	}{
		// FORMATTING §5
		{"yyyy-MM-dd", English, "2081-04-15"},
		{"M/d/yyyy", English, "4/15/2081"},
		{"yyyy'BS'", English, "2081BS"},
		{`yyyy\M\M-\d\d`, English, "2081MM-dd"},
		{"MMM yyyy", English, "Shr 2081"},
		{"dd.MM.yy", English, "15.04.81"},
		// standard patterns
		{"", English, "2081/04/15"},
		{"G", English, "2081/04/15"},
		{"g", Nepali, "२०८१/०४/१५"},
		{"d", English, "2081/04/15"},
		{"D", English, "Shrawan 15, 2081"},
		{"D", Nepali, "साउन १५, २०८१"},
		{"s", English, "2081-04-15"},
		{"s", Nepali, "२०८१-०४-१५"},      // FORMATTING §5: digits per lang
		{`yyyy-MM\`, English, "2081-04"}, // a trailing lone backslash writes nothing
		// token runs and literals
		{"y", English, "81"},
		{"yyy", English, "81"},
		{"yyyyy", English, "2081"},
		{"MMMM", Nepali, "साउन"},
		{"MMMMM", English, "Shrawan"},
		{"ddd", English, "15"},
		{"'unterminated", English, "unterminated"},
		{"''", English, ""},
		{"yyyy 'वि.सं.'", Nepali, "२०८१ वि.सं."},
		{"dd/MM साल yyyy", Nepali, "१५/०४ साल २०८१"},
		{`\`, English, ""},
		{"x", English, "x"},
	}
	for _, c := range cases {
		if got := d.Format(c.pattern, c.lang); got != c.want {
			t.Errorf("Format(%q, %d) = %q, want %q", c.pattern, c.lang, got, c.want)
		}
	}
	if got := (Date{}).Format("yyyy", English); got != "" {
		t.Errorf("zero Date Format = %q, want \"\" (D-01)", got)
	}
}
