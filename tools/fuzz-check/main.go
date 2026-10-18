// Command fuzz-check runs every fuzz target in the module for a fixed budget.
//
//	go run ./tools/fuzz-check
//	go run ./tools/fuzz-check -time=30s -parallel=4
//	go run ./tools/fuzz-check -target='^FuzzParse$' -time=30s
//
// The tool lists the targets with go test -list, then runs one target at a time with
// -parallel=4, so at most four fuzz workers run. The default budget is 30 s per target. A new file
// under a package's testdata/fuzz is a crasher: it is left in place, because the next plain go test
// runs it as a regression test, and the tool stops at that target.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	maxParallel = 4
	minBudget   = time.Second
)

// target is one fuzz function and the package directory that owns its corpus.
type target struct {
	Package string
	Dir     string
	Name    string
}

// stats is the part of the fuzz progress output the report keeps.
type stats struct {
	Execs     int
	NewCorpus int
}

var (
	execsLine   = regexp.MustCompile(`execs:\s*([0-9]+)`)
	interesting = regexp.MustCompile(`(?:new interesting|new corpus):\s*([0-9]+)`)
	totalCorpus = regexp.MustCompile(`total:\s*([0-9]+)`)
	targetName  = regexp.MustCompile(`^Fuzz[A-Za-z0-9_]+$`)
)

func main() {
	budget := flag.Duration("time", 30*time.Second, "fuzztime per target, at least 1s")
	parallel := flag.Int("parallel", maxParallel, "fuzz workers per target, at most 4")
	targetPattern := flag.String("target", ".*", "regular expression for target names")
	flag.Parse()

	filter, err := regexp.Compile("^(?:" + *targetPattern + ")$")
	if err != nil {
		fmt.Fprintln(os.Stderr, "fuzz-check: -target:", err)
		os.Exit(2)
	}
	if err := run(*budget, *parallel, filter); err != nil {
		fmt.Fprintln(os.Stderr, "fuzz-check:", err)
		os.Exit(1)
	}
}

// run discovers and runs the selected targets. It stops at the first failed target.
func run(budget time.Duration, parallel int, filter *regexp.Regexp) error {
	if budget < minBudget {
		return fmt.Errorf("time = %v, want at least %v", budget, minBudget)
	}
	if parallel < 1 || parallel > maxParallel {
		return fmt.Errorf("parallel = %d, want 1 to %d", parallel, maxParallel)
	}
	root, err := moduleRoot()
	if err != nil {
		return err
	}
	all, err := discover(root)
	if err != nil {
		return err
	}
	var selected []target
	for _, t := range all {
		if filter.MatchString(t.Name) {
			selected = append(selected, t)
		}
	}
	if len(selected) == 0 {
		return errors.New("no fuzz targets found")
	}
	fmt.Printf("fuzz-check: %d target(s), %v each, -parallel=%d\n", len(selected), budget, parallel)

	totalExecs := 0
	for i, t := range selected {
		fmt.Printf("fuzz-check: target %d/%d: %s (%s)\n", i+1, len(selected), t.Name, t.Package)
		before, err := corpusFiles(t.Dir)
		if err != nil {
			return err
		}
		out, cmdErr := fuzz(root, t, budget, parallel)
		after, err := corpusFiles(t.Dir)
		if err != nil {
			return err
		}
		saved := addedFiles(before, after)
		s := parseStats(out)
		totalExecs += s.Execs
		if len(saved) > 0 {
			fmt.Printf("fuzz-check: %s saved %d corpus file(s):\n", t.Name, len(saved))
			for _, path := range saved {
				fmt.Println("  " + path)
			}
			return fmt.Errorf("%s failed and saved a crasher; it stays under testdata/fuzz as a regression test", t.Name)
		}
		if cmdErr != nil {
			return fmt.Errorf("%s failed: %w", t.Name, cmdErr)
		}
		fmt.Printf("fuzz-check: %s: execs %d, new corpus %d, crashers 0\n", t.Name, s.Execs, s.NewCorpus)
	}
	fmt.Printf("fuzz-check: %d/%d targets passed, 0 crashers, %d execs\n", len(selected), len(selected), totalExecs)
	return nil
}

// discover lists every test package and asks go test for its fuzz targets. Test packages are the
// only ones that can hold a target.
func discover(root string) ([]target, error) {
	out, err := output(root, "go", "list", "-f", "{{.ImportPath}}\t{{.Dir}}\t{{len .TestGoFiles}}\t{{len .XTestGoFiles}}", "./...")
	if err != nil {
		return nil, err
	}
	var targets []target
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Split(strings.TrimSpace(line), "\t")
		if len(fields) != 4 {
			continue
		}
		tests, err1 := strconv.Atoi(fields[2])
		xtests, err2 := strconv.Atoi(fields[3])
		if err1 != nil || err2 != nil || tests+xtests == 0 {
			continue
		}
		listed, err := output(root, "go", "test", "-list", "^Fuzz", fields[0])
		if err != nil {
			return nil, err
		}
		for _, name := range parseTargets(listed) {
			targets = append(targets, target{Package: fields[0], Dir: fields[1], Name: name})
		}
	}
	sort.Slice(targets, func(i, j int) bool {
		if targets[i].Package != targets[j].Package {
			return targets[i].Package < targets[j].Package
		}
		return targets[i].Name < targets[j].Name
	})
	return targets, nil
}

// parseTargets returns the fuzz test names in go test -list output.
func parseTargets(out string) []string {
	var names []string
	seen := make(map[string]bool)
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if !targetName.MatchString(line) || seen[line] {
			continue
		}
		seen[line] = true
		names = append(names, line)
	}
	return names
}

// fuzz runs one target. The output is echoed and returned so the caller can parse the last line.
func fuzz(root string, t target, budget time.Duration, parallel int) (string, error) {
	args := []string{
		"test",
		"-run=^$",
		"-fuzz=^" + regexp.QuoteMeta(t.Name) + "$",
		"-fuzztime=" + budget.String(),
		"-parallel=" + strconv.Itoa(parallel),
		t.Package,
	}
	fmt.Println("$ go", strings.Join(args, " "))
	cmd := exec.Command("go", args...) //nolint:gosec // the toolchain and package path come from go list
	cmd.Dir = root
	cmd.Env = withEnv(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
	var buf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &buf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &buf)
	err := cmd.Run()
	return buf.String(), err
}

// parseStats returns the final exec count and the number of corpus entries added during the run.
// Go repeats the cumulative "new interesting" count on each progress line, so the last value is
// the run total. The total-corpus form is a fallback for output that omits the delta.
func parseStats(out string) stats {
	var s stats
	for _, match := range execsLine.FindAllStringSubmatch(out, -1) {
		s.Execs, _ = strconv.Atoi(match[1])
	}
	found := false
	for _, match := range interesting.FindAllStringSubmatch(out, -1) {
		s.NewCorpus, _ = strconv.Atoi(match[1])
		found = true
	}
	if found {
		return s
	}
	firstTotal, lastTotal := -1, -1
	for _, match := range totalCorpus.FindAllStringSubmatch(out, -1) {
		n, _ := strconv.Atoi(match[1])
		if firstTotal < 0 {
			firstTotal = n
		}
		lastTotal = n
	}
	if firstTotal >= 0 && lastTotal >= firstTotal {
		s.NewCorpus = lastTotal - firstTotal
	}
	return s
}

// corpusFiles lists the files under dir/testdata/fuzz, relative to that directory. A missing
// directory is an empty corpus.
func corpusFiles(dir string) (map[string]bool, error) {
	root := filepath.Join(dir, "testdata", "fuzz")
	out := make(map[string]bool)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = true
		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return out, nil
	}
	return out, err
}

// addedFiles returns the sorted paths present in after but not before.
func addedFiles(before, after map[string]bool) []string {
	var added []string
	for path := range after {
		if !before[path] {
			added = append(added, path)
		}
	}
	sort.Strings(added)
	return added
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

// output runs a go command in root and returns its combined output.
func output(root string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...) //nolint:gosec // the go toolchain and arguments from go list
	cmd.Dir = root
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	if err != nil {
		return buf.String(), fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, buf.String())
	}
	return buf.String(), nil
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
