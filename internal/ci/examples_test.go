package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestExampleProblems(t *testing.T) {
	const source = `package p

func New() {}
func hidden() {}

type Date struct{}
type hidden struct{}

var ExportedVar int
`
	// ex returns an Example function that prints nothing and says so, which go test accepts.
	ex := func(name string) string { return "func " + name + "() {\n\t// Output:\n}\n" }
	cases := []struct {
		name  string
		tests string
		allow map[string]string
		want  []string
	}{
		{
			name: "missing function and type",
			want: []string{
				"p: function New has no Example function",
				"p: type Date has no Example function",
			},
		},
		{
			name:  "method example does not cover the type",
			tests: ex("ExampleNew") + ex("ExampleDate_Format"),
			want:  []string{"p: type Date has no Example function"},
		},
		{
			name:  "lowercase suffix covers the function and the type",
			tests: ex("ExampleNew_second") + ex("ExampleDate_utc"),
		},
		{
			name:  "example without output does not count",
			tests: "func ExampleNew() {}\n" + ex("ExampleDate"),
			want: []string{
				"p: ExampleNew has no output comment",
				"p: function New has no Example function",
			},
		},
		{
			name:  "method example without output is reported",
			tests: ex("ExampleNew") + ex("ExampleDate") + "func ExampleDate_Format() {}\n",
			want:  []string{"p: ExampleDate_Format has no output comment"},
		},
		{
			name:  "unordered output counts",
			tests: "func ExampleNew() {\n\t// Unordered output:\n\t// a\n}\n" + ex("ExampleDate"),
		},
		{
			name:  "allow-list removes a missing example",
			tests: ex("ExampleDate"),
			allow: map[string]string{"p.New": "a clock cannot be fixed here"},
		},
		{
			name:  "stale allow-list entry",
			tests: ex("ExampleNew") + ex("ExampleDate"),
			allow: map[string]string{"p.New": "no longer needed"},
			want:  []string{`allow-list entry "p.New" is not needed: New has an example`},
		},
		{
			name:  "unknown allow-list entry",
			tests: ex("ExampleNew") + ex("ExampleDate"),
			allow: map[string]string{"p.Nope": "typo"},
			want:  []string{`allow-list entry "p.Nope" does not name an exported function or type`},
		},
		{
			name:  "allow-list entry without a reason",
			tests: ex("ExampleDate"),
			allow: map[string]string{"p.New": ""},
			want:  []string{`allow-list entry "p.New" has no reason`},
		},
		{
			name:  "package example covers nothing",
			tests: ex("Example"),
			want: []string{
				"p: function New has no Example function",
				"p: type Date has no Example function",
			},
		},
		{
			name:  "prose does not count",
			tests: "// ExampleNew is mentioned in a comment.\n",
			want: []string{
				"p: function New has no Example function",
				"p: type Date has no Example function",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, "p.go", source)
			if c.tests != "" {
				writeFile(t, dir, "p_test.go", "package p_test\n\n"+c.tests)
			}
			got, total, err := exampleProblems([]string{dir}, c.allow)
			if err != nil {
				t.Fatal(err)
			}
			if total != 2 {
				t.Errorf("checked %d targets, want 2", total)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("problems = %v, want %v", got, c.want)
			}
		})
	}
}

func writeFile(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestExampleTargetName(t *testing.T) {
	cases := []struct{ name, want string }{
		{"", ""},
		{"_suffix", ""},
		{"New", "New"},
		{"New_second", "New"},
		{"Date_utc", "Date"},
		{"Date_Format", ""},
		{"Date_Format_more", ""},
		{"Date_", "Date"},
	}
	for _, c := range cases {
		if got := exampleTargetName(c.name); got != c.want {
			t.Errorf("exampleTargetName(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}
