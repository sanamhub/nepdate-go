package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writePackage(t *testing.T, src string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "p.go"), []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestMissingDocsClean(t *testing.T) {
	dir := writePackage(t, `package p

// F is documented.
func F() {}

// T is documented.
type T struct {
	// Field is documented.
	Field int
	Inline int // Inline is documented inline.
}

// M is documented.
func (T) M() {}

// Group is documented.
const (
	A = 1
	B = 2
)
`)
	missing, total, err := missingDocs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) != 0 {
		t.Fatalf("documented package reported %v", missing)
	}
	if total != 7 {
		t.Errorf("counted %d exported identifiers, want 7", total)
	}
}

func TestMissingDocsNegative(t *testing.T) {
	dir := writePackage(t, `package p

// F is documented.
func F() {}

func Missing() {}

// T is documented.
type T struct {
	// Field is documented.
	Field int
	Other int
}

// M is documented.
func (T) M() {}

func (T) MissingMethod() {}

// Group is documented.
const (
	C = 3
	D = 4
)

const (
	// A is documented.
	A = 1
	B = 2
)

var MissingVar = 1

func unexported() {}

type unexportedType struct{ ExportedFieldNoDoc int }
`)
	missing, _, err := missingDocs(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"p: constant B has no doc comment",
		"p: field T.Other has no doc comment",
		"p: function Missing has no doc comment",
		"p: method T.MissingMethod has no doc comment",
		"p: variable MissingVar has no doc comment",
	}
	if len(missing) != len(want) {
		t.Fatalf("got %d missing, want %d: %v", len(missing), len(want), missing)
	}
	for i, w := range want {
		if missing[i].String() != w {
			t.Errorf("missing[%d] = %q, want %q", i, missing[i].String(), w)
		}
	}
	err = missingDocsError(missing)
	t.Logf("gate error:\n%v", err)
	for _, w := range want {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("gate error does not contain %q:\n%v", w, err)
		}
	}
}
