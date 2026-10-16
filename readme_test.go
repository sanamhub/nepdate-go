package nepdate_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// TestREADMEQuickStart checks that every Go block in the README quick start is, character for
// character, the body of one of the Example functions, so the snippets compile and run.
func TestREADMEQuickStart(t *testing.T) {
	blocks := readmeGoBlocks(t)
	want := []struct{ file, name string }{
		{"example_test.go", "ExampleNew"},
		{"example_test.go", "ExampleDate_Format"},
		{"example_test.go", "ExampleDate_AppendLong"},
		{"example_test.go", "ExampleParseLenient"},
		{"example_test.go", "ExampleFromTime"},
		{"patro/example_test.go", "ExampleOf"},
	}
	if len(blocks) != len(want) {
		t.Fatalf("README quick start has %d Go blocks, want %d", len(blocks), len(want))
	}
	for i, w := range want {
		body := exampleBody(t, w.file, w.name)
		if got := "\n" + blocks[i]; got != body {
			t.Errorf("README block %d is not the body of %s\nreadme (%d bytes):\n%s\nexample (%d bytes):\n%s", i, w.name, len(got), got, len(body), body)
		}
	}
}

func exampleBody(t *testing.T, path, name string) string {
	t.Helper()
	src, err := os.ReadFile(path) //nolint:gosec // path is a constant at every call site
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Name.Name != name || fd.Body == nil {
			continue
		}
		start := fset.File(fd.Body.Pos()).Offset(fd.Body.Lbrace) + 1
		end := fset.File(fd.Body.Pos()).Offset(fd.Body.Rbrace)
		body := string(src[start:end])
		// The // Output: assertion is test metadata, not part of the snippet a user reads.
		if i := strings.Index(body, "\n\t// Output:"); i >= 0 {
			body = body[:i+1]
		}
		return body
	}
	t.Fatalf("%s: no function %s", path, name)
	return ""
}

func readmeGoBlocks(t *testing.T) []string {
	t.Helper()
	src, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	var blocks []string
	var inQuick, inBlock bool
	var b strings.Builder
	for _, line := range strings.Split(string(src), "\n") {
		switch {
		case strings.HasPrefix(line, "## Quick start"):
			inQuick = true
		case inQuick && strings.HasPrefix(line, "## "):
			inQuick = false
			inBlock = false
		case inQuick && strings.HasPrefix(line, "```go"):
			inBlock = true
		case inBlock && strings.HasPrefix(line, "```"):
			blocks = append(blocks, b.String())
			b.Reset()
			inBlock = false
		case inBlock:
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	return blocks
}
