package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sampleReport() report {
	return report{
		TestEfficacy: 80,
		Total:        10,
		Killed:       8,
		Lived:        1,
		NotCovered:   1,
		Files: []file{
			{Name: "parse.go", Mutations: []mutation{
				{Line: 10, Column: 2, Type: "CONDITIONALS_NEGATION", Status: "LIVED"},
				{Line: 20, Column: 4, Type: "ARITHMETIC_BASE", Status: "KILLED"},
			}},
			{Name: "date.go", Mutations: []mutation{
				{Line: 5, Column: 3, Type: "ARITHMETIC_BASE", Status: "NOT COVERED"},
			}},
			{Name: "time.go", Mutations: []mutation{
				{Line: 47, Column: 32, Type: "ARITHMETIC_BASE", Status: "TIMED OUT"},
			}},
		},
	}
}

func baselineFile(t *testing.T, rep report) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "survivors.txt")
	if err := writeBaseline(path, collectSurvivors(rep)); err != nil {
		t.Fatalf("write baseline: %v", err)
	}
	return path
}

func runSummary(t *testing.T, rep report, threshold float64, update bool, baselinePath string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	err := summarize(&buf, rep, threshold, update, baselinePath)
	return buf.String(), err
}

func TestSummarizeBaselineMatch(t *testing.T) {
	out, err := runSummary(t, sampleReport(), 80, false, baselineFile(t, sampleReport()))
	if err != nil {
		t.Fatalf("exact baseline failed: %v", err)
	}
	for _, want := range []string{"mutation score: 80.00 %", "survivors: 3", "parse.go", "date.go", "time.go", "timed out", "mutation-check: ok"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}

func TestSummarizeScoreGate(t *testing.T) {
	_, err := runSummary(t, sampleReport(), 81, false, baselineFile(t, sampleReport()))
	if err == nil || !strings.Contains(err.Error(), "below 81.00 %") {
		t.Fatalf("err = %v, want the score gate", err)
	}
}

func TestBaselineHigherCountFails(t *testing.T) {
	rep := sampleReport()
	rep.Files[0].Mutations = append(rep.Files[0].Mutations,
		mutation{Line: 30, Column: 1, Type: "CONDITIONALS_NEGATION", Status: "LIVED"})
	_, err := runSummary(t, rep, 0, false, baselineFile(t, sampleReport()))
	if err == nil {
		t.Fatal("a higher survivor count passed")
	}
	for _, want := range []string{
		"parse.go CONDITIONALS_NEGATION LIVED: 2, baseline 1 (higher count)",
		"go run ./tools/mutation-check -update",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error lacks %q:\n%v", want, err)
		}
	}
}

func TestBaselineNewKeyFails(t *testing.T) {
	rep := sampleReport()
	rep.Files[1].Mutations = append(rep.Files[1].Mutations,
		mutation{Line: 9, Column: 1, Type: "CONDITIONALS_NEGATION", Status: "NOT COVERED"})
	_, err := runSummary(t, rep, 0, false, baselineFile(t, sampleReport()))
	if err == nil {
		t.Fatal("a new survivor key passed")
	}
	for _, want := range []string{
		"date.go CONDITIONALS_NEGATION NOT COVERED: 1, new key",
		"go run ./tools/mutation-check -update",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error lacks %q:\n%v", want, err)
		}
	}
}

func TestBaselineLowerCountFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "survivors.txt")
	counts := collectSurvivors(sampleReport())
	counts[survivorKey{File: "parse.go", Type: "CONDITIONALS_NEGATION", Status: "LIVED"}] = 2
	if err := writeBaseline(path, counts); err != nil {
		t.Fatalf("write baseline: %v", err)
	}
	_, err := runSummary(t, sampleReport(), 0, false, path)
	if err == nil {
		t.Fatal("a lower survivor count passed without recording it")
	}
	for _, want := range []string{
		"parse.go CONDITIONALS_NEGATION LIVED: 1, baseline 2 (lower count)",
		"go run ./tools/mutation-check -update",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error lacks %q:\n%v", want, err)
		}
	}
}

func TestBaselineMissingKeyFails(t *testing.T) {
	path := baselineFile(t, sampleReport())
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, []byte("fiscal.go\tARITHMETIC_BASE\tNOT COVERED\t1\n")...)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = runSummary(t, sampleReport(), 0, false, path)
	if err == nil {
		t.Fatal("a missing survivor key passed")
	}
	want := "fiscal.go ARITHMETIC_BASE NOT COVERED: 0, baseline 1 (missing)"
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error lacks %q:\n%v", want, err)
	}
}

func TestBaselineUpdateWritesCurrentCounts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "survivors.txt")
	out, err := runSummary(t, sampleReport(), 80, true, path)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if !strings.Contains(out, "survivor baseline written") {
		t.Errorf("output lacks the update line:\n%s", out)
	}
	got, err := readBaseline(path)
	if err != nil {
		t.Fatalf("read updated baseline: %v", err)
	}
	want := collectSurvivors(sampleReport())
	if len(got) != len(want) {
		t.Fatalf("baseline has %d keys, want %d", len(got), len(want))
	}
	for key, count := range want {
		if got[key] != count {
			t.Errorf("%v count = %d, want %d", key, got[key], count)
		}
	}
}

// TestScoreIsNotRounded is the K7 lesson: 95.9 must fail a 96 gate even though rounding to a
// whole number would pass it, and exactly 96 must pass.
func TestScoreIsNotRounded(t *testing.T) {
	rep := sampleReport()
	rep.TestEfficacy = 95.9
	if _, err := runSummary(t, rep, 96, false, baselineFile(t, rep)); err == nil {
		t.Fatal("95.9 passed a 96 gate")
	}
	rep.TestEfficacy = 96
	if _, err := runSummary(t, rep, 96, false, baselineFile(t, rep)); err != nil {
		t.Fatalf("96 failed a 96 gate: %v", err)
	}
}
