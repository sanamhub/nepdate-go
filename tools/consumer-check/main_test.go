package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCheckOutput(t *testing.T) {
	cases := []struct {
		name string
		out  string
		ok   bool
	}{
		{"exact", wantOutput, true},
		{"trailing newline", wantOutput + "\n", true},
		{"wrong AD date", strings.Replace(wantOutput, "2024-07-30", "2024-07-31", 1), false},
		{"wrong event", strings.Replace(wantOutput, "Kheer Khane Din", "Nothing", 1), false},
		{"missing event", "date: 2081/04/15 -> AD 2024-07-30", false},
		{"extra line", wantOutput + "\nextra", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := checkOutput(c.out)
			if (err == nil) != c.ok {
				t.Fatalf("checkOutput(%q) = %v, want ok=%v", c.out, err, c.ok)
			}
		})
	}
}

func TestOutsideRepo(t *testing.T) {
	root := t.TempDir()
	inside := filepath.Join(root, "sub")
	if err := os.MkdirAll(inside, 0o750); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := outsideRepo(root, root); err == nil {
		t.Error("the repository root passed as outside")
	}
	if err := outsideRepo(root, inside); err == nil {
		t.Error("a directory inside the repository passed as outside")
	}
	if err := outsideRepo(root, outside); err != nil {
		t.Errorf("a sibling directory failed: %v", err)
	}
}

func TestOutsideRepoDifferentVolume(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("volume names exist only on Windows")
	}
	if err := outsideRepo(`D:\checkout`, `C:\Users\runner\Temp\consumer`); err != nil {
		t.Errorf("a directory on another volume failed: %v", err)
	}
	if err := outsideRepo(`\\host\share\checkout`, `\\host\other\consumer`); err != nil {
		t.Errorf("a directory on another UNC share failed: %v", err)
	}
	// The same volume spelled in another case is still the same volume.
	if err := outsideRepo(`E:\checkout`, `e:\checkout\consumer`); err == nil {
		t.Error("a directory inside the repository passed as outside when the drive letter case differs")
	}
	if err := outsideRepo(`\\host\share\checkout`, `\\HOST\SHARE\checkout\consumer`); err == nil {
		t.Error("a directory inside a UNC checkout passed as outside when the share case differs")
	}
}

func TestWriteModule(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "checkout")
	if err := writeModule(dir, root, modulePath); err != nil {
		t.Fatal(err)
	}
	mod, err := os.ReadFile(filepath.Join(dir, "go.mod")) //nolint:gosec // a file this test just wrote
	if err != nil {
		t.Fatal(err)
	}
	want := "replace " + modulePath + " => " + filepath.ToSlash(root)
	if !strings.Contains(string(mod), want) {
		t.Errorf("go.mod lacks %q:\n%s", want, mod)
	}
	if !strings.Contains(string(mod), "require "+modulePath+" v0.0.0") {
		t.Errorf("go.mod lacks the require line:\n%s", mod)
	}
	src, err := os.ReadFile(filepath.Join(dir, "main.go")) //nolint:gosec // a file this test just wrote
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(src), "IMPORT_PATH") {
		t.Error("main.go still has the import placeholder")
	}
	if !strings.Contains(string(src), `"`+modulePath+`"`) {
		t.Errorf("main.go does not import %s:\n%s", modulePath, src)
	}
}

func TestWriteModuleBreaksImport(t *testing.T) {
	dir := t.TempDir()
	if err := writeModule(dir, t.TempDir(), brokenPath); err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(filepath.Join(dir, "main.go")) //nolint:gosec // a file this test just wrote
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), `"`+brokenPath+`"`) {
		t.Errorf("main.go does not import the broken path:\n%s", src)
	}
}
