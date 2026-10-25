package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	releaseModule = "github.com/sanamhub/nepdate-go"
	releaseSHA    = "1d7ca176b1d909883add8054355174349887efe3"
)

func TestReleaseReadinessPasses(t *testing.T) {
	root := releaseFixture(t)
	if err := releaseReadiness(root, "v0.1.0", fixedModules(releaseModule+"\n")); err != nil {
		t.Fatalf("valid release failed: %v", err)
	}
}

func TestReleaseReadinessNegativeControls(t *testing.T) {
	cases := []struct {
		name   string
		file   string
		old    string
		new    string
		want   string
		module func(string) (string, error)
	}{
		{
			name: "CHANGELOG section",
			file: "CHANGELOG.md",
			old:  "0.1.0",
			new:  "0.2.0",
			want: "CHANGELOG: CHANGELOG.md has no line",
		},
		{
			name: "README import path",
			file: "README.md",
			old:  "go get " + releaseModule,
			new:  "go get example.com/wrong",
			want: "README import path: does not contain `go get " + releaseModule + "`",
		},
		{
			name: "README go line",
			file: "README.md",
			old:  "Go 1.24 or later",
			new:  "Go 1.25 or later",
			want: "README go line: does not contain `Go 1.24 or later`",
		},
		{
			name: "ADR-0008 pin",
			file: "docs/adr/0008-supported-go-versions.md",
			old:  "1.24.13",
			new:  "1.24.14",
			want: "ADR-0008 pin: ADR-0008 pins 1.24.14 but ci.yml pins 1.24.13",
		},
		{
			name: "shared version",
			file: "shared/VERSION",
			old:  releaseSHA,
			new:  "pending",
			want: `shared/VERSION: is "pending"`,
		},
		{
			name: "NOTICE",
			file: "shared/data/NOTICE",
			old:  "data/patro/",
			new:  "data/missing/",
			want: `NOTICE: lacks "data/patro/"`,
		},
		{
			name:   "module dependency",
			want:   `go list -m all: lists "github.com/sanamhub/nepdate-go\nexample.com/dep v1.0.0"`,
			module: fixedModules(releaseModule + "\nexample.com/dep v1.0.0\n"),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := releaseFixture(t)
			if c.file != "" {
				editReleaseFile(t, root, c.file, c.old, c.new)
			}
			modules := fixedModules(releaseModule + "\n")
			if c.module != nil {
				modules = c.module
			}
			err := releaseReadiness(root, "v0.1.0", modules)
			if err == nil {
				t.Fatal("the broken input passed the release check")
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error lacks %q:\n%v", c.want, err)
			}
			// Each case breaks one input, so exactly one check may fail.
			if !strings.HasPrefix(err.Error(), "1 release problems:") {
				t.Errorf("want exactly one failed check, got:\n%v", err)
			}
		})
	}
}

// TestReleaseReadinessRepo runs the check on the real repository files, so drift between the
// README, ADR-0008, ci.yml and go.mod fails the normal test gate, not only a release. go list is
// replaced because the module list is covered by the broken-input cases above.
// TestReleaseReadinessRepo checks the repository's own release inputs against the newest dated
// changelog section, so the test follows the changelog instead of naming a version.
func TestReleaseReadinessRepo(t *testing.T) {
	root := filepath.Join("..", "..")
	changelog, err := os.ReadFile(filepath.Join(root, "CHANGELOG.md")) //nolint:gosec // the repository's own changelog
	if err != nil {
		t.Fatal(err)
	}
	version := newestReleasedVersion(string(changelog))
	if version == "" {
		t.Fatal("CHANGELOG.md has no dated version section")
	}
	if err := releaseReadiness(root, "v"+version, fixedModules(releaseModule+"\n")); err != nil {
		t.Fatalf("repository release inputs for v%s: %v", version, err)
	}
}

// newestReleasedVersion returns the version of the first "## [X] - YYYY-MM-DD" line, skipping
// [Unreleased], or "" when there is none.
func newestReleasedVersion(changelog string) string {
	for _, line := range strings.Split(changelog, "\n") {
		rest, ok := strings.CutPrefix(strings.TrimRight(line, "\r"), "## [")
		if !ok {
			continue
		}
		version, date, ok := strings.Cut(rest, "] - ")
		if ok && dateSuffix.MatchString(date) {
			return version
		}
	}
	return ""
}

func TestNewestReleasedVersion(t *testing.T) {
	log := "# Changelog\n\n## [Unreleased]\n\n## [0.1.0-beta.1] - 2026-10-21\n\n## [0.0.9] - 2026-01-01\n"
	if got := newestReleasedVersion(log); got != "0.1.0-beta.1" {
		t.Errorf("newestReleasedVersion = %q, want 0.1.0-beta.1", got)
	}
	if got := newestReleasedVersion("## [Unreleased]\n"); got != "" {
		t.Errorf("newestReleasedVersion with no release = %q, want empty", got)
	}
}

func releaseFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"go.mod": "module " + releaseModule + "\n\ngo 1.24\n",
		"README.md": "# nepdate-go\n\nGo 1.24 or later.\n\n" +
			"```\ngo get " + releaseModule + "\n```\n",
		"CHANGELOG.md":                           "# Changelog\n\n## [0.1.0] - 2026-10-18\n",
		"docs/adr/0008-supported-go-versions.md": "# ADR-0008\n\nUse Go 1.24.13.\n",
		".github/workflows/ci.yml":               "go: [\"1.24.13\"]\n",
		"shared/VERSION":                         releaseSHA + "\n",
		"shared/data/NOTICE":                     "data/calendar/ and data/patro/ are MIT License, from NepDate.\n",
	}
	for name, body := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func editReleaseFile(t *testing.T, root, name, old, replacement string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	data, err := os.ReadFile(path) //nolint:gosec // a fixture under t.TempDir
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), old) {
		t.Fatalf("%s lacks %q", name, old)
	}
	data = []byte(strings.Replace(string(data), old, replacement, 1))
	if err := os.WriteFile(path, data, 0o600); err != nil { //nolint:gosec // a fixture under t.TempDir
		t.Fatal(err)
	}
}

func fixedModules(out string) func(string) (string, error) {
	return func(string) (string, error) { return out, nil }
}
