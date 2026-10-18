// Command mutation-check runs Gremlins on the core package or patro and enforces the mutation
// gate.
//
//	go run ./tools/mutation-check -package=core
//	go run ./tools/mutation-check -package=patro
//	go run ./tools/mutation-check -package=patro -json report.json
//	go run ./tools/mutation-check -package=patro -update
//
// Gremlins is pinned here and fetched with go run, so the root module keeps no dependency on it.
// The run excludes generated files, which leaves the hand-written files of the selected package.
// The score is Gremlins' test efficacy, KILLED / (KILLED + LIVED). Survivors are the LIVED and NOT
// COVERED mutants. Each package has its own survivor baseline next to this file; -update rewrites
// the selected one from the current report.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	gremlinsPackage = "github.com/go-gremlins/gremlins/cmd/gremlins"
	gremlinsVersion = "v0.6.0"
	maxWorkers      = 4
	defaultPackage  = "core"

	// The baseline paths are relative to the module root, so the gate reads the same file from
	// any directory inside the repository.
	defaultSurvivorsPath      = "tools/mutation-check/survivors.txt"
	defaultPatroSurvivorsPath = "tools/mutation-check/survivors-patro.txt"
)

// mutationTarget is one package the gate can mutate and the baseline that belongs to it.
type mutationTarget struct {
	Key      string
	Package  string
	Dir      string
	Baseline string
}

var mutationTargets = map[string]mutationTarget{
	"core":  {Key: "core", Package: "nepdate", Dir: ".", Baseline: defaultSurvivorsPath},
	"patro": {Key: "patro", Package: "patro", Dir: "./patro", Baseline: defaultPatroSurvivorsPath},
}

// targetFor returns the target named by -package.
func targetFor(key string) (mutationTarget, error) {
	target, ok := mutationTargets[key]
	if !ok {
		return mutationTarget{}, fmt.Errorf("package = %q, want core or patro", key)
	}
	return target, nil
}

// defaultThreshold is just below the final G8 score (100.00). The survivor ceiling is not a
// separate number: it is the sum of the counts in the baseline file.
const defaultThreshold = 99.5

// mutation is one mutant in the Gremlins JSON report.
type mutation struct {
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

// file is one source file in the Gremlins JSON report.
type file struct {
	Name      string     `json:"file_name"`
	Mutations []mutation `json:"mutations"`
}

// report is the part of the Gremlins JSON report the gate uses.
type report struct {
	TestEfficacy float64 `json:"test_efficacy"`
	Total        int     `json:"mutants_total"`
	Killed       int     `json:"mutants_killed"`
	Lived        int     `json:"mutants_lived"`
	NotCovered   int     `json:"mutants_not_covered"`
	NotViable    int     `json:"mutants_not_viable"`
	Files        []file  `json:"files"`
}

// survivorKey identifies one baseline line. The source line and column are deliberately absent:
// moving a mutant within a file must not churn the baseline.
type survivorKey struct {
	File   string
	Type   string
	Status string
}

func main() {
	pkg := flag.String("package", defaultPackage, "package to mutate: core or patro")
	workers := flag.Int("workers", maxWorkers, "Gremlins workers, at most 4")
	threshold := flag.Float64("threshold", defaultThreshold, "minimum test efficacy; 0 disables the score gate")
	update := flag.Bool("update", false, "rewrite the survivor baseline from this run")
	out := flag.String("out", "", "keep the Gremlins JSON at this path")
	jsonIn := flag.String("json", "", "evaluate an existing Gremlins JSON report instead of running Gremlins")
	flag.Parse()

	if err := run(*pkg, *workers, *threshold, *update, *out, *jsonIn); err != nil {
		fmt.Fprintln(os.Stderr, "mutation-check:", err)
		os.Exit(1)
	}
}

func run(pkg string, workers int, threshold float64, update bool, out, jsonIn string) error {
	if workers < 1 || workers > maxWorkers {
		return fmt.Errorf("workers = %d, want 1 to %d", workers, maxWorkers)
	}
	target, err := targetFor(pkg)
	if err != nil {
		return err
	}
	root, err := moduleRoot()
	if err != nil {
		return err
	}
	baselinePath := filepath.Join(root, target.Baseline)
	if jsonIn != "" {
		path, err := filepath.Abs(jsonIn)
		if err != nil {
			return err
		}
		return checkReport(path, threshold, update, baselinePath, target.Package)
	}
	jsonPath := out
	cleanup := func() {}
	if jsonPath == "" {
		dir, err := os.MkdirTemp("", "nepdate-mutation-")
		if err != nil {
			return err
		}
		jsonPath = filepath.Join(dir, "mutation.json")
		cleanup = func() { _ = os.RemoveAll(dir) }
	} else {
		jsonPath, err = filepath.Abs(jsonPath)
		if err != nil {
			return err
		}
	}
	defer cleanup()
	// The Windows fallback below trusts an existing report, so a report left by an earlier run
	// at the -out path must not survive into this one.
	if err := os.Remove(jsonPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	args := []string{
		"run", gremlinsPackage + "@" + gremlinsVersion, "unleash",
		"--workers=" + strconv.Itoa(workers),
		"--test-cpu=1",
		"--timeout-coefficient=30",
		"-E", "/",
		"-E", `_gen\.go$`,
		"--output=" + jsonPath,
		target.Dir,
	}
	cmd := exec.Command("go", args...) //nolint:gosec // the toolchain and a pinned package
	cmd.Dir = root
	cmd.Env = withEnv(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("mutation package: %s (%s)\n", target.Key, target.Package)
	fmt.Println("$ go", strings.Join(args, " "))
	if err := cmd.Run(); err != nil {
		// Gremlins sometimes reports cleanup errors for its temporary Git copy on Windows after
		// writing a complete report. Keep the result when the JSON exists; otherwise fail.
		if _, statErr := os.Stat(jsonPath); statErr != nil {
			return fmt.Errorf("gremlins: %w", err)
		}
		fmt.Fprintf(os.Stderr, "mutation-check: warning: gremlins exited with %v; the report is readable\n", err)
	}

	return checkReport(jsonPath, threshold, update, baselinePath, target.Package)
}

// checkReport reads a Gremlins JSON report and applies the gate.
func checkReport(path string, threshold float64, update bool, baselinePath, packageName string) error {
	data, err := os.ReadFile(path) //nolint:gosec // the JSON file named by the caller
	if err != nil {
		return err
	}
	var rep report
	if err := json.Unmarshal(data, &rep); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	if !hasMutations(rep) {
		return fmt.Errorf("gremlins reported no mutants in %s; check the package path and exclusions", path)
	}
	return summarize(os.Stdout, rep, threshold, update, baselinePath, packageName)
}

// hasMutations reports whether the report holds at least one mutant. Gremlins sets
// mutants_total to the killed count, so it is 0 when every mutant is NOT COVERED.
func hasMutations(rep report) bool {
	for _, f := range rep.Files {
		if len(f.Mutations) > 0 {
			return true
		}
	}
	return false
}

// writef and writeln ignore write errors to the report writer. The caller writes to stdout or a
// test buffer, and a failed write cannot be acted on.
func writef(w io.Writer, format string, args ...any) { _, _ = fmt.Fprintf(w, format, args...) }
func writeln(w io.Writer, args ...any)               { _, _ = fmt.Fprintln(w, args...) }

// summarize prints the score, the survivor table and, unless update is set, compares the survivor
// counts with the baseline. The score is compared as the unrounded float; every baseline count is
// compared as an exact integer.
func summarize(w io.Writer, rep report, threshold float64, update bool, baselinePath, packageName string) error {
	counts := collectSurvivors(rep)
	lived, notCovered, timedOut := totals(counts)
	writef(w, "package: %s\n", packageName)
	writef(w, "mutation score: %.2f %% (killed %d, lived %d, not covered %d, timed out %d, not viable %d, total %d)\n",
		rep.TestEfficacy, rep.Killed, rep.Lived, rep.NotCovered, timedOut, rep.NotViable, rep.Total)
	survivors := lived + notCovered + timedOut
	writef(w, "survivors: %d (lived %d, not covered %d, timed out %d)\n", survivors, lived, notCovered, timedOut)

	byFile := make(map[string][3]int)
	for key, count := range counts {
		entry := byFile[key.File]
		switch key.Status {
		case "LIVED":
			entry[0] += count
		case "NOT COVERED":
			entry[1] += count
		case "TIMED OUT":
			entry[2] += count
		}
		byFile[key.File] = entry
	}

	files := make([]string, 0, len(byFile))
	for name := range byFile {
		files = append(files, name)
	}
	sort.Strings(files)
	writef(w, "\n%-28s %6s %12s %10s %6s\n", "file", "lived", "not covered", "timed out", "total")
	for _, name := range files {
		entry := byFile[name]
		writef(w, "%-28s %6d %12d %10d %6d\n", name, entry[0], entry[1], entry[2], entry[0]+entry[1]+entry[2])
	}
	details := survivorDetails(rep)
	if len(details) > 0 {
		writeln(w, "\nsurvivor details:")
		for _, line := range details {
			writeln(w, "  "+line)
		}
	}

	if update {
		if err := writeBaseline(baselinePath, counts, packageName); err != nil {
			return err
		}
		writef(w, "\nsurvivor baseline written: %s (%d keys, %d mutants)\n", baselinePath, len(counts), survivors)
	}

	var problems []error
	if threshold > 0 && rep.TestEfficacy < threshold {
		problems = append(problems, fmt.Errorf("test efficacy %.2f %% is below %.2f %%", rep.TestEfficacy, threshold))
	}
	if !update {
		if err := compareBaseline(baselinePath, counts); err != nil {
			problems = append(problems, err)
		}
	}
	if len(problems) > 0 {
		return errors.Join(problems...)
	}
	writeln(w, "\nmutation-check: ok")
	return nil
}

// collectSurvivors counts LIVED, NOT COVERED and TIMED OUT mutants by file, type and status.
func collectSurvivors(rep report) map[survivorKey]int {
	counts := make(map[survivorKey]int)
	for _, f := range rep.Files {
		name := canonicalFileName(f.Name)
		for _, m := range f.Mutations {
			if !isSurvivor(m.Status) {
				continue
			}
			counts[survivorKey{File: name, Type: m.Type, Status: m.Status}]++
		}
	}
	return counts
}

// totals returns the three status totals in the order used by the report.
func totals(counts map[survivorKey]int) (lived, notCovered, timedOut int) {
	for key, count := range counts {
		switch key.Status {
		case "LIVED":
			lived += count
		case "NOT COVERED":
			notCovered += count
		case "TIMED OUT":
			timedOut += count
		}
	}
	return lived, notCovered, timedOut
}

// survivorDetails lists each survivor with its source position, sorted by file, line and column.
func survivorDetails(rep report) []string {
	var details []string
	for _, f := range rep.Files {
		name := canonicalFileName(f.Name)
		for _, m := range f.Mutations {
			if isSurvivor(m.Status) {
				details = append(details, fmt.Sprintf("%s:%d:%d %s %s", name, m.Line, m.Column, m.Type, m.Status))
			}
		}
	}
	sort.Strings(details)
	return details
}

// isSurvivor reports whether the Gremlins status counts as an unkilled mutant for this gate.
func isSurvivor(status string) bool {
	switch status {
	case "LIVED", "NOT COVERED", "TIMED OUT":
		return true
	default:
		return false
	}
}

// canonicalFileName removes platform-specific path spelling. The run covers package nepdate only,
// so a bare file name is enough; the slash form keeps a report from another OS readable.
func canonicalFileName(name string) string {
	name = filepath.ToSlash(name)
	name = strings.TrimPrefix(name, "./")
	return name
}

// readBaseline reads the tab-separated baseline. Comment and blank lines are ignored.
func readBaseline(path string) (map[survivorKey]int, error) {
	f, err := os.Open(path) //nolint:gosec // the baseline path is fixed by the tool
	if err != nil {
		return nil, fmt.Errorf("read survivor baseline: %w (run `go run ./tools/mutation-check -update` to create it)", err)
	}
	defer func() { _ = f.Close() }()

	counts := make(map[survivorKey]int)
	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 4 {
			return nil, fmt.Errorf("%s:%d: want 4 tab-separated fields, got %d", path, lineNo, len(fields))
		}
		key := survivorKey{File: fields[0], Type: fields[1], Status: fields[2]}
		if key.File == "" || key.Type == "" || !isSurvivor(key.Status) {
			return nil, fmt.Errorf("%s:%d: unknown survivor key %q %q %q", path, lineNo, key.File, key.Type, key.Status)
		}
		count, err := strconv.Atoi(fields[3])
		if err != nil || count < 1 {
			return nil, fmt.Errorf("%s:%d: count %q is not a positive integer", path, lineNo, fields[3])
		}
		if _, duplicate := counts[key]; duplicate {
			return nil, fmt.Errorf("%s:%d: duplicate key %q %q %q", path, lineNo, key.File, key.Type, key.Status)
		}
		counts[key] = count
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read survivor baseline: %w", err)
	}
	return counts, nil
}

// writeBaseline writes the sorted survivor baseline for one package with a short header.
func writeBaseline(path string, counts map[survivorKey]int, packageName string) error {
	var b strings.Builder
	writef(&b, "# Gremlins %s survivor baseline for package %s.\n", gremlinsVersion, packageName)
	writeln(&b, "# Columns: file, mutator type, status, count (tab separated).")
	writeln(&b, "# Line numbers are not part of the key. Regenerate with -update for the selected package.")
	for _, key := range sortedKeys(counts) {
		writef(&b, "%s\t%s\t%s\t%d\n", key.File, key.Type, key.Status, counts[key])
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil { //nolint:gosec // the baseline path is fixed by the tool
		return fmt.Errorf("write survivor baseline: %w", err)
	}
	return nil
}

// compareBaseline reports every key whose count is higher, lower, new or missing. A lower count is
// a kill, but it still needs -update so the baseline records it.
func compareBaseline(path string, current map[survivorKey]int) error {
	baseline, err := readBaseline(path)
	if err != nil {
		return err
	}
	var changes []string
	for _, key := range unionKeys(baseline, current) {
		was, had := baseline[key]
		now, has := current[key]
		label := fmt.Sprintf("%s %s %s", key.File, key.Type, key.Status)
		switch {
		case !had:
			changes = append(changes, fmt.Sprintf("%s: %d, new key", label, now))
		case !has:
			changes = append(changes, fmt.Sprintf("%s: 0, baseline %d (missing)", label, was))
		case now > was:
			changes = append(changes, fmt.Sprintf("%s: %d, baseline %d (higher count)", label, now, was))
		case now < was:
			changes = append(changes, fmt.Sprintf("%s: %d, baseline %d (lower count)", label, now, was))
		}
	}
	if len(changes) == 0 {
		return nil
	}
	noun := "keys"
	if len(changes) == 1 {
		noun = "key"
	}
	return fmt.Errorf("survivor baseline differs in %d %s:\n  %s\nrun `go run ./tools/mutation-check -update` to accept the new baseline",
		len(changes), noun, strings.Join(changes, "\n  "))
}

// sortedKeys returns the baseline keys in file, type, status order.
func sortedKeys(counts map[survivorKey]int) []survivorKey {
	keys := make([]survivorKey, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	return sortKeys(keys)
}

// sortKeys sorts survivor keys in file, type, status order.
func sortKeys(keys []survivorKey) []survivorKey {
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].File != keys[j].File {
			return keys[i].File < keys[j].File
		}
		if keys[i].Type != keys[j].Type {
			return keys[i].Type < keys[j].Type
		}
		return keys[i].Status < keys[j].Status
	})
	return keys
}

// unionKeys returns the sorted union of the two count maps.
func unionKeys(a, b map[survivorKey]int) []survivorKey {
	seen := make(map[survivorKey]bool, len(a)+len(b))
	for key := range a {
		seen[key] = true
	}
	for key := range b {
		seen[key] = true
	}
	keys := make([]survivorKey, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	return sortKeys(keys)
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
