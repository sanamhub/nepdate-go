// Command api-check guards the exported API of nepdate and patro.
//
//	go run ./tools/api-check
//	go run ./tools/api-check -update
//
// Each package has a committed apidiff export-data snapshot under testdata/api. Any difference
// fails, including a compatible addition, until -update records it. Every exported identifier in
// both packages must also appear in docs/API.md, so -update cannot hide an undocumented addition.
// Functions, methods and types must appear as their declaration (`func Name(`,
// `func (r Recv) Name(` or `type Name`), so a common word such as Name or String elsewhere on the
// page does not count; constants, variables and struct fields must appear as a whole word.
// apidiff is pinned and fetched with go run, so the root module keeps no dependency on x/exp.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	modulePath      = "github.com/sanamhub/nepdate-go"
	apidiffPackage  = "golang.org/x/exp/cmd/apidiff"
	apidiffVersion  = "v0.0.0-20261014143641-f0251eb4aa3e"
	docsPath        = "docs/API.md"
	snapshotDir     = "testdata/api"
	nepdatePackage  = modulePath
	patroPackage    = modulePath + "/patro"
	nepdateSnapshot = snapshotDir + "/nepdate.snapshot"
	patroSnapshot   = snapshotDir + "/patro.snapshot"
	nepdateDir      = "."
	patroDir        = "patro"
	updateHint      = "run `go run ./tools/api-check -update` after updating docs/API.md"
)

// packageAPI pairs one package with its snapshot.
type packageAPI struct {
	Name     string
	Package  string
	Dir      string
	Snapshot string
}

var packages = []packageAPI{
	{Name: "nepdate", Package: nepdatePackage, Dir: nepdateDir, Snapshot: nepdateSnapshot},
	{Name: "patro", Package: patroPackage, Dir: patroDir, Snapshot: patroSnapshot},
}

func main() {
	update := flag.Bool("update", false, "rewrite both snapshots from the current packages")
	flag.Parse()

	if err := run(*update); err != nil {
		fmt.Fprintln(os.Stderr, "api-check:", err)
		os.Exit(1)
	}
}

// run compares or updates every package snapshot, then checks docs/API.md.
func run(update bool) error {
	root, err := moduleRoot()
	if err != nil {
		return err
	}
	for _, p := range packages {
		path := filepath.Join(root, filepath.FromSlash(p.Snapshot))
		if update {
			if err := writeSnapshot(root, p, path); err != nil {
				return err
			}
			info, err := os.Stat(path)
			if err != nil {
				return err
			}
			fmt.Printf("api-check: %s: wrote %s (%d bytes)\n", p.Name, p.Snapshot, info.Size())
			continue
		}
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("read %s: %w (%s)", p.Snapshot, err, updateHint)
		}
		diff, err := compare(root, p, path)
		if err != nil {
			return err
		}
		if diff != "" {
			return fmt.Errorf("%s: API differs from %s:\n%s\n%s", p.Name, p.Snapshot, diff, updateHint)
		}
		fmt.Printf("api-check: %s: no changes\n", p.Name)
	}
	return checkDocs(root)
}

// writeSnapshot runs apidiff -w and replaces the snapshot only after the command succeeds.
func writeSnapshot(root string, p packageAPI, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	_, err := apidiff(root, "-w", path, p.Package)
	if err != nil {
		return fmt.Errorf("write %s: %w", p.Snapshot, err)
	}
	return nil
}

// compare returns the apidiff report for one package, or "" when the API is unchanged.
func compare(root string, p packageAPI, snapshot string) (string, error) {
	return apidiff(root, snapshot, p.Package)
}

// apidiff runs the pinned command in root and returns standard output. apidiff exits 0 even when
// it reports removed identifiers, so the caller must read the output.
func apidiff(root string, args ...string) (string, error) {
	full := append([]string{"run", apidiffPackage + "@" + apidiffVersion}, args...)
	fmt.Println("$ go", strings.Join(full, " "))
	cmd := exec.Command("go", full...) //nolint:gosec // the toolchain, pin and paths are constants or checked
	cmd.Dir = root
	cmd.Env = withEnv(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("go %s: %w\n%s", strings.Join(full, " "), err, strings.TrimSpace(stderr.String()))
	}
	if text := strings.TrimSpace(stderr.String()); text != "" {
		fmt.Fprintln(os.Stderr, text)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// declKind says how an exported identifier must appear in docs/API.md.
type declKind int

const (
	kindValue  declKind = iota // constant, variable or struct field: a whole word
	kindFunc                   // package-level function: "func Name("
	kindMethod                 // method: "func (r Recv) Name(" or "func (r *Recv) Name("
	kindType                   // type: "type Name"
)

// exported is one exported identifier and, for a method, its receiver type.
type exported struct {
	Kind declKind
	Recv string
	Name string
}

// String is the identifier as an error message names it.
func (e exported) String() string {
	if e.Kind == kindMethod {
		return e.Recv + "." + e.Name
	}
	return e.Name
}

// documented reports whether doc names e in the form its kind requires.
func documented(doc string, e exported) bool {
	name := regexp.QuoteMeta(e.Name)
	switch e.Kind {
	case kindFunc:
		return regexp.MustCompile(`func ` + name + `\(`).MatchString(doc)
	case kindMethod:
		return regexp.MustCompile(`func \(\w+ \*?` + regexp.QuoteMeta(e.Recv) + `\) ` + name + `\(`).MatchString(doc)
	case kindType:
		return regexp.MustCompile(`type ` + name + `(?:[^A-Za-z0-9_]|$)`).MatchString(doc)
	case kindValue:
		return containsWord(doc, e.Name)
	}
	return false
}

// checkDocs requires every exported identifier in both packages to be named in docs/API.md.
func checkDocs(root string) error {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(docsPath))) //nolint:gosec // the contract path is fixed
	if err != nil {
		return fmt.Errorf("read %s: %w", docsPath, err)
	}
	doc := string(data)
	seen := make(map[exported]bool)
	for _, p := range packages {
		decls, err := exportedDecls(filepath.Join(root, p.Dir))
		if err != nil {
			return fmt.Errorf("%s: %w", p.Name, err)
		}
		for _, e := range decls {
			if seen[e] {
				continue
			}
			seen[e] = true
			if !documented(doc, e) {
				return fmt.Errorf("exported identifier %s is not in %s", e, docsPath)
			}
		}
	}
	fmt.Printf("api-check: %d exported identifiers are in %s\n", len(seen), docsPath)
	fmt.Println("api-check: ok")
	return nil
}

// exportedDecls returns the exported declarations, methods and struct fields in dir, sorted.
func exportedDecls(dir string) ([]exported, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	found := make(map[exported]bool)
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		if err != nil {
			return nil, err
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if !d.Name.IsExported() {
					continue
				}
				if d.Recv == nil {
					found[exported{Kind: kindFunc, Name: d.Name.Name}] = true
				} else if recv, ok := exportedReceiver(d.Recv.List[0].Type); ok {
					found[exported{Kind: kindMethod, Recv: recv, Name: d.Name.Name}] = true
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if !s.Name.IsExported() {
							continue
						}
						found[exported{Kind: kindType, Name: s.Name.Name}] = true
						collectFields(s.Type, found)
					case *ast.ValueSpec:
						for _, n := range s.Names {
							if n.IsExported() {
								found[exported{Kind: kindValue, Name: n.Name}] = true
							}
						}
					}
				}
			}
		}
	}
	decls := make([]exported, 0, len(found))
	for e := range found {
		decls = append(decls, e)
	}
	sort.Slice(decls, func(i, j int) bool { return decls[i].String() < decls[j].String() })
	return decls, nil
}

// collectFields adds exported struct field names under an exported type.
func collectFields(expr ast.Expr, found map[exported]bool) {
	ast.Inspect(expr, func(n ast.Node) bool {
		st, ok := n.(*ast.StructType)
		if !ok {
			return true
		}
		for _, f := range st.Fields.List {
			for _, name := range fieldNames(f) {
				if ast.IsExported(name) {
					found[exported{Kind: kindValue, Name: name}] = true
				}
			}
		}
		return true
	})
}

// fieldNames returns a field's names, or the embedded type name for a field without names.
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

// exportedReceiver returns the receiver type name and whether it is exported.
func exportedReceiver(expr ast.Expr) (string, bool) {
	for {
		switch t := expr.(type) {
		case *ast.StarExpr:
			expr = t.X
		case *ast.ParenExpr:
			expr = t.X
		case *ast.Ident:
			return t.Name, t.IsExported()
		default:
			return "", false
		}
	}
}

// containsWord reports whether name appears in text as a whole identifier.
func containsWord(text, name string) bool {
	re := regexp.MustCompile(`(?:^|[^A-Za-z0-9_])` + regexp.QuoteMeta(name) + `(?:[^A-Za-z0-9_]|$)`)
	return re.MatchString(text)
}

// moduleRoot returns the directory of the main module, so the tool works from any directory
// inside the repository.
func moduleRoot() (string, error) {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}") //nolint:gosec // the go toolchain, no variable arguments
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go list -m: %w (run this from inside the repository)", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// withEnv replaces the values of the named keys in base and appends the rest.
func withEnv(base []string, kv ...string) []string {
	env := make([]string, 0, len(base)+len(kv))
	keys := make(map[string]bool, len(kv))
	for _, pair := range kv {
		key, _, _ := strings.Cut(pair, "=")
		keys[strings.ToUpper(key)] = true
	}
	for _, pair := range base {
		key, _, _ := strings.Cut(pair, "=")
		if !keys[strings.ToUpper(key)] {
			env = append(env, pair)
		}
	}
	return append(env, kv...)
}
