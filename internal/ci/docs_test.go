package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeDocs creates a documentation tree and returns its root.
func writeDocs(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, body := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o750); err != nil {
		t.Fatal(err)
	}
	return root
}

func cleanDocs() map[string]string {
	return map[string]string{
		"README.md":                "# Project\n\n[API](docs/API.md), [install](#install), [web](https://pkg.go.dev/x).\n",
		"CHANGELOG.md":             "# Changelog\n\n## [0.1.0] - 2026-10-17\n\n- First release.\n",
		"docs/API.md":              "# API\n\n[README](../README.md). The range 1901–2199 is allowed here.\n",
		"docs/runbooks/release.md": "# Release\n\nTag `vX.Y.Z` and push it.\n",
		"docs/adr/0001.md":         "# ADR\n\nDecided in G6-01; the next phase keeps its task ID.\n",
		"docs/plan/x.md":           "# Plan\n\nG6-02 is a task ID in a plan file.\n",
	}
}

func issueText(issues []docIssue) string {
	lines := make([]string, len(issues))
	for i, issue := range issues {
		lines[i] = issue.String()
	}
	return strings.Join(lines, "\n")
}

func TestDocsCheckClean(t *testing.T) {
	issues, n, err := docsCheck(writeDocs(t, cleanDocs()))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("%d issues in a clean tree:\n%s", len(issues), issueText(issues))
	}
	if n != 6 {
		t.Errorf("checked %d files, want 6", n)
	}
}

func TestDocsCheckEmDash(t *testing.T) {
	files := cleanDocs()
	files["README.md"] = "# Project\n\nThis is an em dash \u2014 in the README.\n"
	issues, _, err := docsCheck(writeDocs(t, files))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0].line != 3 || !strings.Contains(issues[0].msg, "em dash") {
		t.Fatalf("issues = %v, want one em dash on line 3", issues)
	}
}

func TestDocsCheckBrokenLink(t *testing.T) {
	files := cleanDocs()
	files["docs/API.md"] = "# API\n\n[missing](../missing.md).\n"
	issues, _, err := docsCheck(writeDocs(t, files))
	if err != nil {
		t.Fatal(err)
	}
	want := `docs/API.md:3: relative link "../missing.md" does not exist`
	if len(issues) != 1 || issues[0].String() != want {
		t.Fatalf("issues = %v, want %q", issues, want)
	}
}

func TestDocsCheckLinkEscapesRepo(t *testing.T) {
	files := cleanDocs()
	files["docs/API.md"] = "# API\n\n[out](../../outside.md)\n"
	issues, _, err := docsCheck(writeDocs(t, files))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || !strings.Contains(issues[0].msg, "escapes the repository") {
		t.Fatalf("issues = %v, want a link that escapes the repository", issues)
	}
}

func TestDocsCheckTaskID(t *testing.T) {
	files := cleanDocs()
	files["docs/benchmarks.md"] = "# Benchmarks\n\nMeasured by task G6-01.\n"
	issues, _, err := docsCheck(writeDocs(t, files))
	if err != nil {
		t.Fatal(err)
	}
	want := `docs/benchmarks.md:3: task ID "G6-01" in a user-facing file`
	if len(issues) != 1 || issues[0].String() != want {
		t.Fatalf("issues = %v, want %q", issues, want)
	}
}

func TestDocsCheckTaskIDAllowedInPlanAndADR(t *testing.T) {
	issues, _, err := docsCheck(writeDocs(t, cleanDocs()))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("task IDs in plan and ADR files were reported:\n%s", issueText(issues))
	}
}

func TestDocsCheckFencedLink(t *testing.T) {
	files := cleanDocs()
	files["docs/API.md"] = "# API\n\n```\n[x](missing.md)\n```\n"
	issues, _, err := docsCheck(writeDocs(t, files))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("a link in a code fence was checked:\n%s", issueText(issues))
	}

	files["docs/API.md"] = "# API\n\n```\n[x]\n```\n\n[y](missing.md)\n"
	issues, _, err = docsCheck(writeDocs(t, files))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || !strings.Contains(issues[0].msg, `"missing.md"`) {
		t.Fatalf("issues = %v, want the link outside the fence", issues)
	}
}

func TestIsUserFacing(t *testing.T) {
	for _, file := range []string{"README.md", "CHANGELOG.md", "docs/API.md", "docs/benchmarks.md", "docs/competitors.md", "docs/runbooks/release.md"} {
		if !isUserFacing(file) {
			t.Errorf("%s is not user-facing", file)
		}
	}
	for _, file := range []string{"docs/adr/0001.md", "docs/plan/G6.md", "docs/plan/PROGRESS.md"} {
		if isUserFacing(file) {
			t.Errorf("%s is user-facing", file)
		}
	}
}
