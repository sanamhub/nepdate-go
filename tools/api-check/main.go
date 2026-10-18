// Command api-check guards the exported API of nepdate and patro.
//
//	go run ./tools/api-check
//	go run ./tools/api-check -update
//
// Each package has a committed apidiff export-data snapshot under testdata/api. Any difference
// fails, including a compatible addition, until -update records it. Every exported identifier in
// both packages must also appear in docs/API.md, so -update cannot hide an undocumented addition.
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

// checkDocs requires every exported name in both packages to appear as a word in docs/API.md.
func checkDocs(root string) error {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(docsPath))) //nolint:gosec // the contract path is fixed
	if err != nil {
		return fmt.Errorf("read %s: %w", docsPath, err)
	}
	doc := string(data)
	seen := make(map[string]bool)
	total := 0
	for _, p := range packages {
		names, err := exportedNames(filepath.Join(root, p.Dir))
		if err != nil {
			return fmt.Errorf("%s: %w", p.Name, err)
		}
		for _, name := range names {
			if seen[name] {
				continue
			}
			seen[name] = true
			total++
			if !containsWord(doc, name) {
				return fmt.Errorf("exported identifier %s is not in %s", name, docsPath)
			}
		}
	}
	fmt.Printf("api-check: %d exported names are in %s\n", total, docsPath)
	fmt.Println("api-check: ok")
	return nil
}

// exportedNames returns the distinct exported declarations, methods and struct fields in dir.
func exportedNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	found := make(map[string]bool)
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
				if d.Recv == nil || exportedReceiver(d.Recv.List[0].Type) {
					found[d.Name.Name] = true
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if !s.Name.IsExported() {
							continue
						}
						found[s.Name.Name] = true
						collectFields(s.Type, found)
					case *ast.ValueSpec:
						for _, n := range s.Names {
							if n.IsExported() {
								found[n.Name] = true
							}
						}
					}
				}
			}
		}
	}
	names := make([]string, 0, len(found))
	for name := range found {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// collectFields adds exported struct field names under an exported type.
func collectFields(expr ast.Expr, found map[string]bool) {
	ast.Inspect(expr, func(n ast.Node) bool {
		st, ok := n.(*ast.StructType)
		if !ok {
			return true
		}
		for _, f := range st.Fields.List {
			for _, name := range fieldNames(f) {
				if ast.IsExported(name) {
					found[name] = true
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

// exportedReceiver reports whether a method receiver is an exported type.
func exportedReceiver(expr ast.Expr) bool {
	for {
		switch t := expr.(type) {
		case *ast.StarExpr:
			expr = t.X
		case *ast.ParenExpr:
			expr = t.X
		case *ast.Ident:
			return t.IsExported()
		default:
			return false
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
