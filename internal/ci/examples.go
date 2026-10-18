package main

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// exampleAllowList names exported functions and types that cannot have a deterministic example.
// Keep it empty while every target can print a fixed value. An entry must give the reason, and it
// is rejected once the target has an example, so an exemption cannot outlive its cause.
var exampleAllowList = map[string]string{}

// exampleTarget is one exported function or type that needs an example.
type exampleTarget struct {
	pkg  string
	kind string
	name string
}

func (t exampleTarget) String() string {
	return t.pkg + ": " + t.kind + " " + t.name
}

// exampleGate extends gate 8: every exported package-level function and every exported type in
// nepdate and patro has an Example function named Example<Name> or Example<Name>_<suffix>, and every
// Example function has an output comment, so go test runs it and checks what it prints.
func exampleGate() error {
	problems, total, err := exampleProblems([]string{".", "./patro"}, exampleAllowList)
	if err != nil {
		return err
	}
	if len(problems) > 0 {
		return fmt.Errorf("%d example problems:\n  %s", len(problems), strings.Join(problems, "\n  "))
	}
	fmt.Printf("ci: %d exported functions and types have examples\n", total)
	return nil
}

// exampleProblems returns one line per missing example or bad allow-list entry, and the number of
// targets checked.
func exampleProblems(dirs []string, allow map[string]string) ([]string, int, error) {
	targets := make(map[string]exampleTarget)
	examples := make(map[string]bool)
	var problems []string
	for _, dir := range dirs {
		found, err := exportedTargets(dir)
		if err != nil {
			return nil, 0, err
		}
		for key, target := range found {
			targets[key] = target
		}
		names, silent, err := exampleNames(dir)
		if err != nil {
			return nil, 0, err
		}
		for name := range names {
			examples[name] = true
		}
		problems = append(problems, silent...)
	}

	for key, target := range targets {
		if examples[key] {
			continue
		}
		if _, ok := allow[key]; ok {
			continue
		}
		problems = append(problems, target.String()+" has no Example function")
	}
	for key, reason := range allow {
		target, ok := targets[key]
		if !ok {
			problems = append(problems, fmt.Sprintf("allow-list entry %q does not name an exported function or type", key))
			continue
		}
		if reason == "" {
			problems = append(problems, fmt.Sprintf("allow-list entry %q has no reason", key))
		}
		if examples[key] {
			problems = append(problems, fmt.Sprintf("allow-list entry %q is not needed: %s has an example", key, target.name))
		}
	}
	sort.Strings(problems)
	return problems, len(targets), nil
}

// exportedTargets returns the exported package-level functions and exported types in dir, keyed by
// package name and identifier.
func exportedTargets(dir string) (map[string]exampleTarget, error) {
	files, pkg, err := packageFiles(dir, false)
	if err != nil {
		return nil, err
	}
	targets := make(map[string]exampleTarget)
	add := func(kind, name string) {
		targets[pkg+"."+name] = exampleTarget{pkg: pkg, kind: kind, name: name}
	}
	for _, file := range files {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil && d.Name.IsExported() {
					add("function", d.Name.Name)
				}
			case *ast.GenDecl:
				if d.Tok != token.TYPE {
					continue
				}
				for _, spec := range d.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.IsExported() {
						add("type", ts.Name.Name)
					}
				}
			}
		}
	}
	return targets, nil
}

// exampleNames returns the declarations an Example function documents, keyed like exportedTargets,
// and one problem line per Example function without an output comment. Names follow go vet and
// go/doc: ExampleDate and ExampleDate_utc cover Date, while ExampleDate_Format documents the method
// Format and does not cover Date.
func exampleNames(dir string) (map[string]bool, []string, error) {
	files, pkg, err := packageFiles(dir, true)
	if err != nil {
		return nil, nil, err
	}
	names := make(map[string]bool)
	var silent []string
	for _, ex := range doc.Examples(files...) {
		if ex.Output == "" && !ex.EmptyOutput {
			silent = append(silent, fmt.Sprintf("%s: Example%s has no output comment", pkg, ex.Name))
			continue
		}
		if name := exampleTargetName(ex.Name); name != "" {
			names[pkg+"."+name] = true
		}
	}
	return names, silent, nil
}

// exampleTargetName returns the function or type an example documents, given the Example function
// name without its prefix, or "" for a package example or a method example. A final part after an
// underscore that does not start with an upper-case letter is a variant suffix and is dropped first,
// as go/doc does.
func exampleTargetName(name string) string {
	if i := strings.LastIndexByte(name, '_'); i >= 0 {
		if r, _ := utf8.DecodeRuneInString(name[i+1:]); !unicode.IsUpper(r) {
			name = name[:i]
		}
	}
	if strings.Contains(name, "_") {
		return ""
	}
	return name
}

// packageFiles returns the parsed files in dir and the package name. Test files are included only
// when tests is true; they may use an external _test package, which is why the package name is read
// from the first file.
func packageFiles(dir string, tests bool) ([]*ast.File, string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, "", err
	}
	fset := token.NewFileSet()
	var files []*ast.File
	pkg := ""
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		if !tests && strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ParseComments)
		if err != nil {
			return nil, "", err
		}
		if pkg == "" && !strings.HasSuffix(name, "_test.go") {
			pkg = file.Name.Name
		}
		files = append(files, file)
	}
	if len(files) == 0 {
		return nil, "", fmt.Errorf("%s has no Go files", dir)
	}
	if pkg == "" {
		pkg = files[0].Name.Name
	}
	return files, pkg, nil
}
