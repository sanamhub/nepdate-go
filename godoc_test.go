package nepdate_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

const goDocVersion = "go1.27.1"

// TestGoDocPreview compares the committed go doc output with the command output. It runs only on
// the toolchain that generated the files, so a different Go version skips instead of producing a
// formatting diff that says nothing about the docs. CI sets NEPDATE_GODOC_REQUIRED in the job that
// pins that toolchain, where a skip would hide a stale preview, so there it fails instead.
func TestGoDocPreview(t *testing.T) {
	if runtime.Version() != goDocVersion {
		if os.Getenv("NEPDATE_GODOC_REQUIRED") != "" {
			t.Fatalf("pkg.go.dev preview needs %s, this is %s, and NEPDATE_GODOC_REQUIRED is set", goDocVersion, runtime.Version())
		}
		t.Skipf("pkg.go.dev preview needs %s, this is %s", goDocVersion, runtime.Version())
	}
	for _, c := range []struct {
		pkg  string
		path string
	}{
		{".", "docs/godoc/nepdate.txt"},
		{"./patro", "docs/godoc/patro.txt"},
	} {
		out, err := exec.Command("go", "doc", "-all", c.pkg).Output() //nolint:gosec // fixed tool and package paths
		if err != nil {
			t.Fatalf("go doc -all %s: %v", c.pkg, err)
		}
		want, err := os.ReadFile(c.path) //nolint:gosec // a committed file under docs/godoc
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(out, want) {
			t.Errorf("%s is stale; regenerate with `go doc -all %s > %s`\n%s",
				c.path, c.pkg, c.path, firstDiff(want, out))
		}
	}
}

// firstDiff names the first line that differs, or the length difference.
func firstDiff(want, got []byte) string {
	wantLines := strings.Split(string(want), "\n")
	gotLines := strings.Split(string(got), "\n")
	for i := 0; i < len(wantLines) && i < len(gotLines); i++ {
		if wantLines[i] != gotLines[i] {
			return fmt.Sprintf("line %d:\n  want %q\n  got  %q", i+1, wantLines[i], gotLines[i])
		}
	}
	return fmt.Sprintf("line count: want %d, got %d", len(wantLines), len(gotLines))
}
