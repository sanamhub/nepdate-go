// Command mutation-check runs Gremlins on the core package and enforces the mutation gate.
//
//	go run ./tools/mutation-check -baseline
//	go run ./tools/mutation-check
//
// Gremlins is pinned here and fetched with go run, so the root module keeps no dependency on it.
// The run excludes generated files and every subdirectory, which leaves the hand-written files of
// package nepdate. The score is Gremlins' test efficacy, KILLED / (KILLED + LIVED). Survivors are
// the LIVED and NOT COVERED mutants.
package main

import (
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
)

// The gate defaults are just below the final G8 score (100.00) and exactly at the final survivor
// count (51). Zero disables the score check and a negative ceiling disables the survivor ceiling,
// which is how the baseline runs.
const (
	defaultThreshold    = 99.5
	defaultMaxSurvivors = 51
)

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

func main() {
	baseline := flag.Bool("baseline", false, "report the score and survivors without enforcing the gate")
	workers := flag.Int("workers", maxWorkers, "Gremlins workers, at most 4")
	threshold := flag.Float64("threshold", defaultThreshold, "minimum test efficacy; 0 disables the score gate")
	maxSurvivors := flag.Int("max-survivors", defaultMaxSurvivors, "maximum LIVED plus NOT COVERED plus TIMED OUT; negative disables the ceiling")
	out := flag.String("out", "", "keep the Gremlins JSON at this path")
	jsonIn := flag.String("json", "", "evaluate an existing Gremlins JSON report instead of running Gremlins")
	flag.Parse()

	if err := run(*baseline, *workers, *threshold, *maxSurvivors, *out, *jsonIn); err != nil {
		fmt.Fprintln(os.Stderr, "mutation-check:", err)
		os.Exit(1)
	}
}

func run(baseline bool, workers int, threshold float64, maxSurvivors int, out, jsonIn string) error {
	if workers < 1 || workers > maxWorkers {
		return fmt.Errorf("workers = %d, want 1 to %d", workers, maxWorkers)
	}
	if jsonIn != "" {
		path, err := filepath.Abs(jsonIn)
		if err != nil {
			return err
		}
		return checkReport(path, baseline, threshold, maxSurvivors)
	}
	root, err := moduleRoot()
	if err != nil {
		return err
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

	args := []string{
		"run", gremlinsPackage + "@" + gremlinsVersion, "unleash",
		"--workers=" + strconv.Itoa(workers),
		"--test-cpu=1",
		"--timeout-coefficient=10",
		"-E", "/",
		"-E", `_gen\.go$`,
		"--output=" + jsonPath,
		".",
	}
	cmd := exec.Command("go", args...) //nolint:gosec // the toolchain and a pinned package
	cmd.Dir = root
	cmd.Env = withEnv(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println("$ go", strings.Join(args, " "))
	if err := cmd.Run(); err != nil {
		// Gremlins sometimes reports cleanup errors for its temporary Git copy on Windows after
		// writing a complete report. Keep the result when the JSON exists; otherwise fail.
		if _, statErr := os.Stat(jsonPath); statErr != nil {
			return fmt.Errorf("gremlins: %w", err)
		}
		fmt.Fprintf(os.Stderr, "mutation-check: warning: gremlins exited with %v; the report is readable\n", err)
	}

	return checkReport(jsonPath, baseline, threshold, maxSurvivors)
}

// checkReport reads a Gremlins JSON report and applies the gate.
func checkReport(path string, baseline bool, threshold float64, maxSurvivors int) error {
	data, err := os.ReadFile(path) //nolint:gosec // the JSON file named by the caller
	if err != nil {
		return err
	}
	var rep report
	if err := json.Unmarshal(data, &rep); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	if rep.Total == 0 {
		return fmt.Errorf("gremlins reported 0 mutants in %s; check the package path and exclusions", path)
	}
	return summarize(os.Stdout, rep, baseline, threshold, maxSurvivors)
}

// writef and writeln ignore write errors to the report writer. The caller writes to stdout or a
// test buffer, and a failed write cannot be acted on.
func writef(w io.Writer, format string, args ...any) { _, _ = fmt.Fprintf(w, format, args...) }
func writeln(w io.Writer, args ...any)               { _, _ = fmt.Fprintln(w, args...) }

// summarize prints the score, the survivor table and, unless baseline is set, applies the gate.
// The score is compared as the unrounded float; the survivor count is an exact integer.
func summarize(w io.Writer, rep report, baseline bool, threshold float64, maxSurvivors int) error {
	lived, notCovered, timedOut := 0, 0, 0
	byFile := make(map[string][3]int)
	var details []string
	for _, f := range rep.Files {
		for _, m := range f.Mutations {
			counts := byFile[f.Name]
			switch m.Status {
			case "LIVED":
				lived++
				counts[0]++
				details = append(details, fmt.Sprintf("%s:%d:%d %s LIVED", f.Name, m.Line, m.Column, m.Type))
			case "NOT COVERED":
				notCovered++
				counts[1]++
				details = append(details, fmt.Sprintf("%s:%d:%d %s NOT COVERED", f.Name, m.Line, m.Column, m.Type))
			case "TIMED OUT":
				timedOut++
				counts[2]++
				details = append(details, fmt.Sprintf("%s:%d:%d %s TIMED OUT", f.Name, m.Line, m.Column, m.Type))
			}
			byFile[f.Name] = counts
		}
	}
	writef(w, "mutation score: %.2f %% (killed %d, lived %d, not covered %d, timed out %d, not viable %d, total %d)\n",
		rep.TestEfficacy, rep.Killed, rep.Lived, rep.NotCovered, timedOut, rep.NotViable, rep.Total)
	survivors := lived + notCovered + timedOut
	writef(w, "survivors: %d (lived %d, not covered %d, timed out %d)\n", survivors, lived, notCovered, timedOut)

	files := make([]string, 0, len(byFile))
	for name := range byFile {
		files = append(files, name)
	}
	sort.Strings(files)
	writef(w, "\n%-28s %6s %12s %10s %6s\n", "file", "lived", "not covered", "timed out", "total")
	for _, name := range files {
		counts := byFile[name]
		writef(w, "%-28s %6d %12d %10d %6d\n", name, counts[0], counts[1], counts[2], counts[0]+counts[1]+counts[2])
	}
	if len(details) > 0 {
		writeln(w, "\nsurvivor details:")
		for _, line := range details {
			writeln(w, "  "+line)
		}
	}

	if baseline {
		writeln(w, "\nmutation-check: baseline (gate not enforced)")
		return nil
	}
	var problems []error
	if threshold > 0 && rep.TestEfficacy < threshold {
		problems = append(problems, fmt.Errorf("test efficacy %.2f %% is below %.2f %%", rep.TestEfficacy, threshold))
	}
	if maxSurvivors >= 0 && survivors > maxSurvivors {
		problems = append(problems, fmt.Errorf("survivors %d exceed the ceiling of %d", survivors, maxSurvivors))
	}
	if len(problems) > 0 {
		return errors.Join(problems...)
	}
	writeln(w, "\nmutation-check: ok")
	return nil
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
