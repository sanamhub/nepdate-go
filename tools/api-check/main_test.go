package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportedNames(t *testing.T) {
	dir := t.TempDir()
	src := `package p

const ExportedConst = 1
var ExportedVar = 1

type ExportedType struct {
	ExportedField int
	hidden        int
}

type hidden struct{}

func ExportedFunc() {}

func (ExportedType) ExportedMethod() {}
func (hidden) HiddenMethod()         {}
`
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	decls, err := exportedDecls(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, len(decls))
	for i, e := range decls {
		got[i] = e.String()
	}
	want := "ExportedConst,ExportedField,ExportedFunc,ExportedType,ExportedType.ExportedMethod,ExportedVar"
	if strings.Join(got, ",") != want {
		t.Errorf("exportedDecls = %v, want %s", got, want)
	}
}

func TestDocumentedNeedsTheDeclaration(t *testing.T) {
	doc := `func Parse(s string)
func (d Date) Short(o DateOrder)
func (d *Date) UnmarshalText(b []byte)
type Range struct{}
const ( YMD DateOrder = iota )
The Name of a month and the String form.
`
	cases := []struct {
		e    exported
		want bool
	}{
		{exported{Kind: kindFunc, Name: "Parse"}, true},
		{exported{Kind: kindMethod, Recv: "Date", Name: "Short"}, true},
		{exported{Kind: kindMethod, Recv: "Date", Name: "UnmarshalText"}, true},
		{exported{Kind: kindType, Name: "Range"}, true},
		{exported{Kind: kindValue, Name: "YMD"}, true},
		// A word in prose is not a declaration.
		{exported{Kind: kindFunc, Name: "Name"}, false},
		{exported{Kind: kindMethod, Recv: "Month", Name: "String"}, false},
		{exported{Kind: kindType, Name: "Name"}, false},
		// A method on another receiver does not document this one.
		{exported{Kind: kindMethod, Recv: "Range", Name: "Short"}, false},
		// A method does not document a function of the same name.
		{exported{Kind: kindFunc, Name: "Short"}, false},
	}
	for _, c := range cases {
		if got := documented(doc, c.e); got != c.want {
			t.Errorf("documented(%v) = %v, want %v", c.e, got, c.want)
		}
	}
}

func TestContainsWord(t *testing.T) {
	text := "func Parse, ParseLenient and ParseDate; type ParseError"
	cases := []struct {
		name string
		want bool
	}{
		{"Parse", true},
		{"ParseLenient", true},
		{"ParseDate", true},
		{"ParseError", true},
		{"Par", false},
		{"Lenient", false},
	}
	for _, c := range cases {
		if got := containsWord(text, c.name); got != c.want {
			t.Errorf("containsWord(%q) = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestCheckDocsReportsMissingName(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "pkg"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o750); err != nil {
		t.Fatal(err)
	}
	src := "package p\n\nfunc Documented() {}\nfunc Missing() {}\n"
	if err := os.WriteFile(filepath.Join(root, "pkg", "a.go"), []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "API.md"), []byte("func Documented()\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	old := packages
	defer func() { packages = old }()
	packages = []packageAPI{{Name: "p", Dir: "pkg"}}

	err := checkDocs(root)
	if err == nil || !strings.Contains(err.Error(), "Missing") {
		t.Fatalf("checkDocs = %v, want a failure naming Missing", err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "API.md"), []byte("func Documented()\nfunc Missing()\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := checkDocs(root); err != nil {
		t.Errorf("documented API failed: %v", err)
	}
}
