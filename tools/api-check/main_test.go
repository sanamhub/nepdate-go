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
	names, err := exportedNames(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := "ExportedConst,ExportedField,ExportedFunc,ExportedMethod,ExportedType,ExportedVar"
	if strings.Join(names, ",") != want {
		t.Errorf("exportedNames = %v, want %s", names, want)
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
