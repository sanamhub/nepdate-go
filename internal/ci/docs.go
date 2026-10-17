package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// docIssue is one problem in a Markdown file.
type docIssue struct {
	file string
	line int
	msg  string
}

func (i docIssue) String() string {
	return fmt.Sprintf("%s:%d: %s", i.file, i.line, i.msg)
}

// docsGate checks README.md, CHANGELOG.md and every Markdown file under docs/: no em dash, every
// relative link target exists, and no task ID in a user-facing file.
func docsGate() error {
	issues, n, err := docsCheck(".")
	if err != nil {
		return err
	}
	if len(issues) > 0 {
		return docsError(issues)
	}
	fmt.Printf("ci: %d documents checked\n", n)
	return nil
}

func docsError(issues []docIssue) error {
	lines := make([]string, len(issues))
	for i, issue := range issues {
		lines[i] = issue.String()
	}
	return fmt.Errorf("%d documentation problems:\n  %s", len(issues), strings.Join(lines, "\n  "))
}

// links matches an inline Markdown link target. Targets with spaces in angle brackets are not
// used in this repository.
var links = regexp.MustCompile(`\]\(([^)\s]+)\)`)

// taskID matches a phase or task identifier, such as G6 or G6-01.
var taskID = regexp.MustCompile(`\bG[0-9]+(?:-[0-9]+)?\b`)

// docsCheck returns the problems in the documentation under root and the number of files checked.
func docsCheck(root string) ([]docIssue, int, error) {
	files, err := docFiles(root)
	if err != nil {
		return nil, 0, err
	}
	var issues []docIssue
	for _, file := range files {
		found, err := checkDoc(root, file)
		if err != nil {
			return nil, 0, err
		}
		issues = append(issues, found...)
	}
	sort.Slice(issues, func(i, j int) bool {
		if issues[i].file != issues[j].file {
			return issues[i].file < issues[j].file
		}
		return issues[i].line < issues[j].line
	})
	return issues, len(files), nil
}

// docFiles lists the Markdown files the gate covers, relative to root with forward slashes.
func docFiles(root string) ([]string, error) {
	var files []string
	for _, name := range []string{"README.md", "CHANGELOG.md"} {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			files = append(files, name)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	err := filepath.WalkDir(filepath.Join(root, "docs"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

// checkDoc checks one file. Line numbers are 1-based.
func checkDoc(root, file string) ([]docIssue, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file))) //nolint:gosec // a file under the repository root
	if err != nil {
		return nil, err
	}
	userFacing := isUserFacing(file)
	var issues []docIssue
	add := func(line int, format string, args ...any) {
		issues = append(issues, docIssue{file: file, line: line, msg: fmt.Sprintf(format, args...)})
	}
	inFence := false
	for i, line := range strings.Split(string(data), "\n") {
		n := i + 1
		if strings.Contains(line, "\u2014") {
			add(n, "em dash (U+2014); use a comma, colon, period or parentheses")
		}
		if userFacing {
			if id := taskID.FindString(line); id != "" {
				add(n, "task ID %q in a user-facing file", id)
			}
		}
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		for _, m := range links.FindAllStringSubmatch(line, -1) {
			target := m[1]
			if skipLink(target) {
				continue
			}
			target, _, _ = strings.Cut(target, "#")
			if target == "" {
				continue
			}
			full := filepath.Join(root, filepath.Dir(filepath.FromSlash(file)), filepath.FromSlash(target))
			rel, err := filepath.Rel(root, full)
			if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				add(n, "relative link %q escapes the repository", m[1])
				continue
			}
			if _, err := os.Stat(full); err != nil { //nolint:gosec // full is inside root, checked above
				add(n, "relative link %q does not exist", m[1])
			}
		}
	}
	return issues, nil
}

// skipLink reports whether a link target is not a repository-relative path: an absolute URL, a
// mailto link or a fragment of the same document.
func skipLink(target string) bool {
	switch {
	case strings.HasPrefix(target, "http://"), strings.HasPrefix(target, "https://"):
		return true
	case strings.HasPrefix(target, "mailto:"), strings.HasPrefix(target, "#"):
		return true
	case strings.HasPrefix(target, "/"), strings.Contains(target, "://"):
		return true
	}
	return false
}

// isUserFacing reports whether a file ships to users of the module. ADRs and plan files are for
// maintainers, so they keep task IDs for traceability.
func isUserFacing(file string) bool {
	switch file {
	case "README.md", "CHANGELOG.md", "docs/API.md", "docs/benchmarks.md", "docs/competitors.md":
		return true
	}
	return strings.HasPrefix(file, "docs/runbooks/")
}
