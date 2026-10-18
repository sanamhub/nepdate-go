package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const changelog = `# Changelog

## [Unreleased]

## [0.1.0] - 2026-11-02

### Added
- First release.
`

func TestReleaseCheck(t *testing.T) {
	cases := []struct {
		arg string
		ok  bool
	}{
		{"v0.1.0", true},
		{"v9.9.9", false},
		{"0.1.0", false},
		{"v0.1", false},
		{"v0.1.0-rc.1", false},
		{"v01.1.0", false},
		{"", false},
	}
	for _, c := range cases {
		err := releaseCheck(changelog, c.arg)
		if (err == nil) != c.ok {
			t.Errorf("releaseCheck(%q) = %v, want ok=%v", c.arg, err, c.ok)
		}
	}
}

func TestReleaseCheckNeedsDate(t *testing.T) {
	for _, log := range []string{
		"## [0.1.0]\n",
		"## [0.1.0] - \n",
		"## [0.1.0] - 2026-11\n",
		"## [0.1.0] - soon\n",
		"### [0.1.0] - 2026-11-02\n",
	} {
		if err := releaseCheck(log, "v0.1.0"); err == nil {
			t.Errorf("changelog %q accepted", log)
		}
	}
	if err := releaseCheck("## [0.1.0] - 2026-11-02\n", "v0.1.0"); err != nil {
		t.Errorf("minimal changelog rejected: %v", err)
	}
}

func TestCoverageOf(t *testing.T) {
	out := "ok  \tgithub.com/sanamhub/nepdate-go\t0.3s\tcoverage: 93.4% of statements\n" +
		"ok  \tgithub.com/sanamhub/nepdate-go/patro\t0.2s\tcoverage: [no statements]\n"
	got := coverageOf(out)
	if len(got) != 2 {
		t.Fatalf("got %d packages, want 2: %v", len(got), got)
	}
	if got[0].pkg != "github.com/sanamhub/nepdate-go" || got[0].percent != 93.4 || got[0].none {
		t.Errorf("first line parsed as %+v", got[0])
	}
	if got[1].pkg != "github.com/sanamhub/nepdate-go/patro" || !got[1].none {
		t.Errorf("second line parsed as %+v", got[1])
	}
}

func TestCoverageGateFailsBelowMinimum(t *testing.T) {
	if err := coverageOK([]coverage{{pkg: "p", percent: 89.9}}); err == nil {
		t.Error("89.9 % passed")
	}
	if err := coverageOK([]coverage{{pkg: "p", percent: 90.0}, {pkg: "q", none: true}}); err != nil {
		t.Errorf("90.0 %% and [no statements] failed: %v", err)
	}
	if err := coverageOK(nil); err == nil {
		t.Error("no coverage line passed")
	}
}

func TestFormatGateFailsOnUnformattedFile(t *testing.T) {
	dir := t.TempDir()
	write := func(name, src string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("ok.go", "package p\n\nfunc F() {}\n")
	if err := gofmtGate(dir); err != nil {
		t.Fatalf("formatted tree failed: %v", err)
	}
	write("bad.go", "package p\nfunc  G( ) {}\n")
	if err := gofmtGate(dir); err == nil {
		t.Fatal("unformatted file passed")
	}
}

func TestRaceSupported(t *testing.T) {
	cases := []struct {
		cgo, arch string
		want      bool
	}{
		{"1", "amd64", true},
		{"1", "arm64", true},
		{"0", "amd64", false},
		{"1", "386", false},
	}
	for _, c := range cases {
		if got := raceSupported(c.cgo, c.arch); got != c.want {
			t.Errorf("raceSupported(%q, %q) = %v", c.cgo, c.arch, got)
		}
	}
}

func TestIsolationCheck(t *testing.T) {
	marker := []byte(isolationMarker)
	ok := isolationCheck("github.com/sanamhub/nepdate-go\nerrors\n", []byte("core"), append([]byte("x"), marker...))
	if ok != nil {
		t.Errorf("clean core failed: %v", ok)
	}
	if err := isolationCheck("github.com/sanamhub/nepdate-go/patro\n", []byte("core"), marker); err == nil {
		t.Error("core depending on patro passed")
	}
	if err := isolationCheck("", append([]byte("core"), marker...), marker); err == nil {
		t.Error("core binary with patro names passed")
	}
	if err := isolationCheck("", []byte("core"), []byte("patro without names")); err == nil {
		t.Error("patro binary without the names passed")
	}
}

func TestRunGatesCountsSkipsSeparately(t *testing.T) {
	ok := func() error { return nil }
	skip := func() error { return fmt.Errorf("%w: tool not on PATH", errSkipped) }
	fail := func() error { return errors.New("boom") }

	if code := runGates([]gate{{"a", ok}, {"b", skip}, {"c", ok}}); code != 0 {
		t.Errorf("run with a skipped gate exited %d, want 0", code)
	}
	if code := runGates([]gate{{"a", ok}, {"b", skip}, {"c", fail}}); code != 1 {
		t.Errorf("run with a failed gate exited %d, want 1", code)
	}
}

func TestSummaryNamesSkippedGates(t *testing.T) {
	cases := []struct {
		passed, total int
		skipped       []string
		want          string
	}{
		{8, 8, nil, "ci: 8/8 gates passed"},
		{6, 8, []string{"golangci-lint", "govulncheck"}, "ci: 6/8 gates passed, 2 skipped (golangci-lint, govulncheck)"},
		{2, 8, []string{"golangci-lint"}, "ci: 2/8 gates passed, 1 skipped (golangci-lint)"},
	}
	for _, c := range cases {
		if got := summary(c.passed, c.total, c.skipped); got != c.want {
			t.Errorf("summary(%d, %d, %v) = %q, want %q", c.passed, c.total, c.skipped, got, c.want)
		}
	}
}
