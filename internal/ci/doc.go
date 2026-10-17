package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// missingDoc is one exported identifier without a doc comment.
type missingDoc struct {
	pkg  string
	kind string
	name string
}

func (m missingDoc) String() string {
	return m.pkg + ": " + m.kind + " " + m.name + " has no doc comment"
}

// docGate checks the exported surface of both packages. It runs go doc -all so the gate fails if
// the command cannot render a package, then walks the same declarations with go/ast.
func docGate() error {
	total := 0
	for _, dir := range []string{".", "./patro"} {
		if err := quiet("go", "doc", "-all", dir); err != nil {
			return err
		}
		missing, n, err := missingDocs(dir)
		if err != nil {
			return err
		}
		total += n
		if len(missing) > 0 {
			return missingDocsError(missing)
		}
	}
	fmt.Printf("ci: %d exported identifiers have doc comments\n", total)
	return nil
}

func missingDocsError(missing []missingDoc) error {
	lines := make([]string, len(missing))
	for i, m := range missing {
		lines[i] = m.String()
	}
	return fmt.Errorf("%d exported identifiers have no doc comment:\n  %s",
		len(missing), strings.Join(lines, "\n  "))
}

// missingDocs returns every exported identifier in dir that has no doc comment, the number of
// exported identifiers checked, and a parse error. A doc comment on a const, var or type group
// covers every name in the group, which is how go doc shows it.
func missingDocs(dir string) ([]missingDoc, int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, 0, err
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ParseComments)
		if err != nil {
			return nil, 0, err
		}
		files = append(files, f)
	}
	if len(files) == 0 {
		return nil, 0, fmt.Errorf("%s has no Go files", dir)
	}
	pkgName := files[0].Name.Name
	for _, f := range files[1:] {
		if f.Name.Name != pkgName {
			return nil, 0, fmt.Errorf("%s holds package %s and %s, want one package", dir, pkgName, f.Name.Name)
		}
	}

	var (
		missing []missingDoc
		total   int
	)
	add := func(kind, name string) {
		total++
		missing = append(missing, missingDoc{pkg: pkgName, kind: kind, name: name})
	}
	for _, file := range files {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				name, exported := funcName(d)
				if !exported {
					continue
				}
				kind := "function"
				if d.Recv != nil {
					kind = "method"
				}
				if d.Doc == nil {
					add(kind, name)
				} else {
					total++
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if !s.Name.IsExported() {
							continue
						}
						if s.Doc == nil && d.Doc == nil {
							add("type", s.Name.Name)
						} else {
							total++
						}
						checkFields(s.Type, s.Name.Name, &total, &missing, pkgName)
					case *ast.ValueSpec:
						for _, n := range s.Names {
							if !n.IsExported() {
								continue
							}
							if s.Doc == nil && s.Comment == nil && d.Doc == nil {
								add(kindOf(d.Tok), n.Name)
							} else {
								total++
							}
						}
					}
				}
			}
		}
	}
	sort.Slice(missing, func(i, j int) bool { return missing[i].String() < missing[j].String() })
	return missing, total, nil
}

// checkFields visits every struct field under an exported type, including nested structs and
// embedded fields, which go doc also shows.
func checkFields(typ ast.Expr, owner string, total *int, missing *[]missingDoc, pkgName string) {
	ast.Inspect(typ, func(n ast.Node) bool {
		st, ok := n.(*ast.StructType)
		if !ok {
			return true
		}
		for _, f := range st.Fields.List {
			names := fieldNames(f)
			if len(names) == 0 {
				continue
			}
			for _, name := range names {
				if !ast.IsExported(name) {
					continue
				}
				*total++
				if f.Doc == nil && f.Comment == nil {
					*missing = append(*missing, missingDoc{
						pkg:  pkgName,
						kind: "field",
						name: owner + "." + name,
					})
				}
			}
		}
		return true
	})
}

// fieldNames returns the field name, or the embedded type's name for a field without names.
func fieldNames(f *ast.Field) []string {
	if len(f.Names) > 0 {
		names := make([]string, len(f.Names))
		for i, n := range f.Names {
			names[i] = n.Name
		}
		return names
	}
	switch t := f.Type.(type) {
	case *ast.Ident:
		return []string{t.Name}
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return []string{id.Name}
		}
	case *ast.SelectorExpr:
		return []string{t.Sel.Name}
	}
	return nil
}

// funcName returns the name go doc shows for a function or method, and whether it is exported.
// A method on an unexported receiver is not part of go doc's surface.
func funcName(d *ast.FuncDecl) (string, bool) {
	if !d.Name.IsExported() {
		return "", false
	}
	if d.Recv == nil || len(d.Recv.List) == 0 {
		return d.Name.Name, true
	}
	recv := receiverName(d.Recv.List[0].Type)
	if recv == "" || !ast.IsExported(recv) {
		return "", false
	}
	return recv + "." + d.Name.Name, true
}

func receiverName(t ast.Expr) string {
	for {
		switch x := t.(type) {
		case *ast.StarExpr:
			t = x.X
		case *ast.ParenExpr:
			t = x.X
		case *ast.Ident:
			return x.Name
		default:
			return ""
		}
	}
}

func kindOf(tok token.Token) string {
	if tok == token.CONST {
		return "constant"
	}
	return "variable"
}

// quiet runs a command with its output discarded, returning its error and standard error.
func quiet(name string, args ...string) error {
	cmd := exec.Command(name, args...) //nolint:gosec // constant tool and arguments
	var stderr bytes.Buffer
	cmd.Stdout = io.Discard
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return nil
}
