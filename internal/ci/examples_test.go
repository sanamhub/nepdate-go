package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestExampleTargetName(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"Example", ""},
		{"ExampleNew", "New"},
		{"ExampleDate_Format", "Date"},
		{"ExampleDate_Format_more", "Date"},
		{"ExampleNew_second", "New"},
		{"Example_suffix", ""},
		{"NotAnExample", ""},
	}
	for _, c := range cases {
		if got := exampleTargetName(c.name); got != c.want {
			t.Errorf("exampleTargetName(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestExampleProblems(t *testing.T) {
	const source = `package p

func New() {}
func hidden() {}

type Date struct{}
type hidden struct{}

var ExportedVar int
`
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
			name:  "method example covers the type",
			tests: "func ExampleDate_Format() {}\n",
			want:  []string{"p: function New has no Example function"},
		},
		{
			name:  "suffix covers the function",
			tests: "func ExampleNew_second() {}\nfunc ExampleDate_Format() {}\n",
		},
		{
			name:  "allow-list removes a missing example",
			tests: "func ExampleDate_Format() {}\n",
			allow: map[string]string{"p.New": "a clock cannot be fixed here"},
		},
		{
			name:  "stale allow-list entry",
			tests: "func ExampleNew() {}\nfunc ExampleDate_Format() {}\n",
			allow: map[string]string{"p.New": "no longer needed"},
			want:  []string{`allow-list entry "p.New" is not needed: New has an example`},
		},
		{
			name:  "unknown allow-list entry",
			tests: "func ExampleNew() {}\nfunc ExampleDate_Format() {}\n",
			allow: map[string]string{"p.Nope": "typo"},
			want:  []string{`allow-list entry "p.Nope" does not name an exported function or type`},
		},
		{
			name:  "allow-list entry without a reason",
			tests: "func ExampleDate_Format() {}\n",
			allow: map[string]string{"p.New": ""},
			want:  []string{`allow-list entry "p.New" has no reason`},
		},
		{
			name:  "package example covers nothing",
			tests: "func Example() {}\n",
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
