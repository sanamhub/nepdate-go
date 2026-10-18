package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	goModuleLine  = regexp.MustCompile(`(?m)^module\s+(\S+)\s*$`)
	goVersionLine = regexp.MustCompile(`(?m)^go\s+([0-9]+\.[0-9]+)\s*$`)
	go124Patch    = regexp.MustCompile(`\b1\.24\.[0-9]+\b`)
	hubSHA        = regexp.MustCompile(`^[0-9a-f]{40}$`)
)

// releaseReadiness checks every file a release depends on. It prints one line per check and
// returns every failed check in one error.
func releaseReadiness(root, tag string, modules func(string) (string, error)) error {
	mod, err := readReleaseFile(root, "go.mod")
	if err != nil {
		return fmt.Errorf("release-check: %w", err)
	}
	module, goversion, err := goModInfo(mod)
	if err != nil {
		return fmt.Errorf("release-check: %w", err)
	}

	var problems []string
	check := func(name string, found []string) {
		if len(found) == 0 {
			fmt.Printf("release-check: %s: ok\n", name)
			return
		}
		fmt.Printf("release-check: %s: failed\n", name)
		for _, p := range found {
			problems = append(problems, name+": "+p)
		}
	}

	if changelog, err := readReleaseFile(root, "CHANGELOG.md"); err != nil {
		check("CHANGELOG", []string{err.Error()})
	} else if err := releaseCheck(changelog, tag); err != nil {
		check("CHANGELOG", []string{err.Error()})
	} else {
		check("CHANGELOG", nil)
	}

	if readme, err := readReleaseFile(root, "README.md"); err != nil {
		check("README import path", []string{err.Error()})
		check("README go line", []string{err.Error()})
	} else {
		check("README import path", readmeImportProblems(readme, module))
		check("README go line", readmeVersionProblems(readme, goversion))
	}

	adr, adrErr := readReleaseFile(root, "docs/adr/0008-supported-go-versions.md")
	ci, ciErr := readReleaseFile(root, ".github/workflows/ci.yml")
	var pins []string
	if adrErr != nil {
		pins = append(pins, adrErr.Error())
	}
	if ciErr != nil {
		pins = append(pins, ciErr.Error())
	}
	if adrErr == nil && ciErr == nil {
		pins = pinProblems(adr, ci, goversion)
	}
	check("ADR-0008 pin", pins)

	if version, err := readReleaseFile(root, "shared/VERSION"); err != nil {
		check("shared/VERSION", []string{err.Error()})
	} else {
		check("shared/VERSION", versionProblems(version))
	}

	if notice, err := readReleaseFile(root, "shared/data/NOTICE"); err != nil {
		check("NOTICE", []string{err.Error()})
	} else {
		check("NOTICE", noticeProblems(notice))
	}

	if out, err := modules(root); err != nil {
		check("go list -m all", []string{err.Error()})
	} else if got := strings.TrimSpace(out); got != module {
		check("go list -m all", []string{fmt.Sprintf("lists %q, want only %q", got, module)})
	} else {
		check("go list -m all", nil)
	}

	if len(problems) > 0 {
		return fmt.Errorf("%d release problems:\n  %s", len(problems), strings.Join(problems, "\n  "))
	}
	return nil
}

// readReleaseFile reads a repository-relative file.
func readReleaseFile(root, name string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name))) //nolint:gosec // name is a constant relative path
	if err != nil {
		return "", fmt.Errorf("%s: %w", name, err)
	}
	return string(data), nil
}

// goModInfo returns the module path and the go directive.
func goModInfo(mod string) (string, string, error) {
	m := goModuleLine.FindStringSubmatch(mod)
	v := goVersionLine.FindStringSubmatch(mod)
	if m == nil || v == nil {
		return "", "", errors.New("go.mod needs a module line and a go line")
	}
	return m[1], v[1], nil
}

// readmeImportProblems checks that the README shows the module's import path.
func readmeImportProblems(readme, module string) []string {
	if !strings.Contains(readme, "go get "+module) {
		return []string{fmt.Sprintf("does not contain `go get %s`", module)}
	}
	return nil
}

// readmeVersionProblems checks that the README states the go.mod version.
func readmeVersionProblems(readme, goversion string) []string {
	if !strings.Contains(readme, "Go "+goversion+" or later") {
		return []string{fmt.Sprintf("does not contain `Go %s or later`", goversion)}
	}
	return nil
}

// pinProblems checks that ADR-0008 and ci.yml name the same 1.24 patch.
func pinProblems(adr, ci, goversion string) []string {
	var problems []string
	adrPins := uniqueStrings(go124Patch.FindAllString(adr, -1))
	ciPins := uniqueStrings(go124Patch.FindAllString(ci, -1))
	if len(adrPins) != 1 {
		problems = append(problems, fmt.Sprintf("ADR-0008 names %v, want one 1.24 patch", adrPins))
	}
	if len(ciPins) != 1 {
		problems = append(problems, fmt.Sprintf("ci.yml names %v, want one 1.24 patch", ciPins))
	}
	if len(adrPins) == 1 && len(ciPins) == 1 {
		if adrPins[0] != ciPins[0] {
			problems = append(problems, fmt.Sprintf("ADR-0008 pins %s but ci.yml pins %s", adrPins[0], ciPins[0]))
		}
		if !strings.HasPrefix(adrPins[0], goversion+".") {
			problems = append(problems, fmt.Sprintf("pin %s does not match go.mod go %s", adrPins[0], goversion))
		}
	}
	return problems
}

// versionProblems checks that shared/VERSION is a hub commit SHA.
func versionProblems(version string) []string {
	version = strings.TrimSpace(version)
	if !hubSHA.MatchString(version) {
		return []string{fmt.Sprintf("is %q, want a 40-character lowercase hex hub SHA", version)}
	}
	return nil
}

// noticeProblems checks that the shared data NOTICE covers both data directories.
func noticeProblems(notice string) []string {
	var problems []string
	for _, marker := range []string{"data/calendar/", "data/patro/", "MIT License", "NepDate"} {
		if !strings.Contains(notice, marker) {
			problems = append(problems, fmt.Sprintf("lacks %q", marker))
		}
	}
	return problems
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]bool, len(in))
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// moduleList returns go list -m all from root. The release must have no dependency.
func moduleList(root string) (string, error) {
	cmd := exec.Command("go", "list", "-m", "all") //nolint:gosec // fixed tool and arguments
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go list -m all: %w", err)
	}
	return string(out), nil
}
