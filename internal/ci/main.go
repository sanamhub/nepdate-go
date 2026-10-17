// Command ci runs every quality gate of ADR-0004 in order and stops at the first failure.
//
//	go run ./internal/ci                       run the gates
//	go run ./internal/ci release-check v0.1.0  check CHANGELOG.md has a dated 0.1.0 section
//
// The last line is always "ci: <passed>/<total> gates passed".
package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// minCoverage is the statement coverage each library package needs (ADR-0004 gate 3).
const minCoverage = 90.0

type gate struct {
	name string
	run  func() error
}

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] != "release-check" || len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "usage: go run ./internal/ci [release-check vX.Y.Z]")
			os.Exit(2)
		}
		log, err := os.ReadFile("CHANGELOG.md")
		if err == nil {
			err = releaseCheck(string(log), os.Args[2])
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "release-check:", err)
			os.Exit(1)
		}
		fmt.Println("release-check: ok")
		return
	}
	os.Exit(runGates(gates()))
}

func gates() []gate {
	return []gate{
		{"format and vet", func() error {
			if err := gofmtGate("."); err != nil {
				return err
			}
			return command("go", "vet", "./...")
		}},
		{"golangci-lint", func() error { return optional("golangci-lint", "run") }},
		{"tests and coverage", testGate},
		{"generated files", func() error { return command("go", "run", "./internal/gen", "-check") }},
		{"govulncheck", func() error { return optional("govulncheck", "./...") }},
		{"core isolation", isolationGate},
		{"doc comments", docGate},
		{"docs", docsGate},
	}
}

func runGates(gs []gate) int {
	passed := 0
	for i, g := range gs {
		fmt.Printf("ci: gate %d/%d: %s\n", i+1, len(gs), g.name)
		if err := g.run(); err != nil {
			fmt.Fprintf(os.Stderr, "ci: gate %d (%s) failed: %v\n", i+1, g.name, err)
			fmt.Printf("ci: %d/%d gates passed\n", passed, len(gs))
			return 1
		}
		passed++
	}
	fmt.Printf("ci: %d/%d gates passed\n", passed, len(gs))
	return 0
}

// command runs name with args, printing the command line, with output going to the terminal.
func command(name string, args ...string) error {
	fmt.Println("$", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...) //nolint:gosec // the gate list is fixed in this file
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// output runs name with args and returns its standard output, echoing it to the terminal.
func output(name string, args ...string) (string, error) {
	fmt.Println("$", name, strings.Join(args, " "))
	var out bytes.Buffer
	cmd := exec.Command(name, args...) //nolint:gosec // the gate list is fixed in this file
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if _, werr := os.Stdout.Write(out.Bytes()); werr != nil && err == nil {
		err = werr
	}
	return out.String(), err
}

// optional runs a tool that CI always installs but a developer machine may not have. A missing
// tool counts as passed, with a warning.
func optional(name string, args ...string) error {
	if _, err := exec.LookPath(name); err != nil {
		fmt.Printf("ci: warning: %s not on PATH, gate skipped (CI always runs it)\n", name)
		return nil
	}
	return command(name, args...)
}

func gofmtGate(dir string) error {
	out, err := output("gofmt", "-l", dir)
	if err != nil {
		return err
	}
	if strings.TrimSpace(out) != "" {
		return errors.New("files above are not gofmt-ed")
	}
	return nil
}

func testGate() error {
	cgo, err := goEnv("CGO_ENABLED")
	if err != nil {
		return err
	}
	arch, err := goEnv("GOARCH")
	if err != nil {
		return err
	}
	args := []string{"test", "-shuffle=on", "./..."}
	if raceSupported(cgo, arch) {
		args = []string{"test", "-race", "-shuffle=on", "./..."}
	} else {
		fmt.Printf("ci: warning: race detector needs cgo on amd64 or arm64 (CGO_ENABLED=%s, GOARCH=%s); running without -race\n", cgo, arch)
	}
	if err := command("go", args...); err != nil {
		return err
	}
	pkgs := []string{"."}
	if st, err := os.Stat("patro"); err == nil && st.IsDir() {
		pkgs = append(pkgs, "./patro")
	}
	out, err := output("go", append([]string{"test", "-cover"}, pkgs...)...)
	if err != nil {
		return err
	}
	return coverageOK(coverageOf(out))
}

func goEnv(key string) (string, error) {
	out, err := exec.Command("go", "env", key).Output() //nolint:gosec // key is a constant at every call site
	return strings.TrimSpace(string(out)), err
}

func raceSupported(cgo, arch string) bool {
	return cgo == "1" && (arch == "amd64" || arch == "arm64")
}

type coverage struct {
	pkg     string
	percent float64
	none    bool // "coverage: [no statements]"
}

// A package with statements but no test files prints "coverage: 0.0% of statements", so the
// bare "[no test files]" line means the package has no statements either.
var coverageLine = regexp.MustCompile(`^(?:ok|\?)\s+(\S+)\s.*(?:coverage: (?:([0-9.]+)% of statements|(\[no statements\]))|(\[no test files\])$)`)

func coverageOf(out string) []coverage {
	var res []coverage
	for _, line := range strings.Split(out, "\n") {
		m := coverageLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		c := coverage{pkg: m[1], none: m[3] != "" || m[4] != ""}
		if !c.none {
			c.percent, _ = strconv.ParseFloat(m[2], 64)
		}
		res = append(res, c)
	}
	return res
}

func coverageOK(cs []coverage) error {
	if len(cs) == 0 {
		return errors.New("no coverage line in the go test output")
	}
	for _, c := range cs {
		if !c.none && c.percent < minCoverage {
			return fmt.Errorf("%s: coverage %.1f %% is below %.1f %%", c.pkg, c.percent, minCoverage)
		}
	}
	return nil
}

// isolationMarker is an event name that only the patro tables hold (hub ADR-0004).
const isolationMarker = "World Day Against Human Trafficking"

const patroPackage = "github.com/sanamhub/nepdate-go/patro"

// isolationGate is ADR-0004 gate 6: the core doesn't depend on patro, a core-only binary holds
// no patro names, and a binary that uses patro does, so the check can't pass by accident.
func isolationGate() error {
	// The core-only program is listed too: a blank import of patro there links no patro data
	// (nothing references it), so only the dependency list catches it.
	deps, err := output("go", "list", "-deps", ".", "./internal/coreonly")
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "nepdate-isolation")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	core := filepath.Join(dir, "coreonly")
	withPatro := filepath.Join(dir, "withpatro")
	if err := command("go", "build", "-o", core, "./internal/coreonly"); err != nil {
		return err
	}
	if err := command("go", "build", "-o", withPatro, "./internal/withpatro"); err != nil {
		return err
	}
	coreBin, err := os.ReadFile(core) //nolint:gosec // a binary this gate just built
	if err != nil {
		return err
	}
	patroBin, err := os.ReadFile(withPatro) //nolint:gosec // a binary this gate just built
	if err != nil {
		return err
	}
	return isolationCheck(deps, coreBin, patroBin)
}

func isolationCheck(deps string, coreBin, patroBin []byte) error {
	for _, line := range strings.Fields(deps) {
		if strings.TrimSpace(line) == patroPackage {
			return errors.New("package nepdate or internal/coreonly depends on " + patroPackage)
		}
	}
	if bytes.Contains(coreBin, []byte(isolationMarker)) {
		return fmt.Errorf("the core-only binary contains %q", isolationMarker)
	}
	if !bytes.Contains(patroBin, []byte(isolationMarker)) {
		return fmt.Errorf("the patro binary lacks %q, so the core-only check proves nothing", isolationMarker)
	}
	return nil
}

var (
	tagPattern = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
	dateSuffix = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}\s*$`)
)

// releaseCheck succeeds only if tag is vX.Y.Z and the changelog has a line
// "## [X.Y.Z] - YYYY-MM-DD" (ADR-0005).
func releaseCheck(changelog, tag string) error {
	if !tagPattern.MatchString(tag) {
		return fmt.Errorf("%q is not a tag of the form vX.Y.Z", tag)
	}
	prefix := "## [" + tag[1:] + "] - "
	for _, line := range strings.Split(changelog, "\n") {
		line = strings.TrimRight(line, "\r")
		if rest, ok := strings.CutPrefix(line, prefix); ok && dateSuffix.MatchString(rest) {
			return nil
		}
	}
	return fmt.Errorf("CHANGELOG.md has no line %q followed by a YYYY-MM-DD date", prefix)
}
