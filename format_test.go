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
