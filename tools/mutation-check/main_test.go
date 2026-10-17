package main

import (
	"bytes"
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

func runSummary(t *testing.T, rep report, baseline bool, threshold float64, maxSurvivors int) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	err := summarize(&buf, rep, baseline, threshold, maxSurvivors)
	return buf.String(), err
}

func TestSummarizeBaseline(t *testing.T) {
	out, err := runSummary(t, sampleReport(), true, 99, 0)
	if err != nil {
		t.Fatalf("baseline failed: %v", err)
	}
	for _, want := range []string{"mutation score: 80.00 %", "survivors: 3", "parse.go", "date.go", "time.go", "timed out", "baseline"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}

func TestSummarizeScoreGate(t *testing.T) {
	_, err := runSummary(t, sampleReport(), false, 81, 2)
	if err == nil || !strings.Contains(err.Error(), "below 81.00 %") {
		t.Fatalf("err = %v, want the score gate", err)
	}
}

func TestSummarizeSurvivorCeiling(t *testing.T) {
	_, err := runSummary(t, sampleReport(), false, 80, 2)
	if err == nil || !strings.Contains(err.Error(), "exceed the ceiling of 2") {
		t.Fatalf("err = %v, want the survivor ceiling", err)
	}
}

func TestSummarizePasses(t *testing.T) {
	out, err := runSummary(t, sampleReport(), false, 80, 3)
	if err != nil {
		t.Fatalf("passing report failed: %v", err)
	}
	if !strings.Contains(out, "mutation-check: ok") {
		t.Errorf("output lacks the ok line:\n%s", out)
	}
}

// TestScoreIsNotRounded is the K7 lesson: 95.4 must fail a 96 gate, even though it prints as
// 95.40 and a whole-number comparison would also reject it; 95.9 must fail even though a
// one-decimal rounding to 96 would pass.
func TestScoreIsNotRounded(t *testing.T) {
	rep := sampleReport()
	rep.TestEfficacy = 95.9
	if _, err := runSummary(t, rep, false, 96, 3); err == nil {
		t.Fatal("95.9 passed a 96 gate")
	}
	rep.TestEfficacy = 96
	if _, err := runSummary(t, rep, false, 96, 3); err != nil {
		t.Fatalf("96 failed a 96 gate: %v", err)
	}
}
