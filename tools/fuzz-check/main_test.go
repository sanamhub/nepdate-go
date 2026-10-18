package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestParseTargets(t *testing.T) {
	out := "FuzzParse\nFuzzParse\nFuzzAddDays\n" +
		"ok  \tgithub.com/sanamhub/nepdate-go\t0.2s\n" +
		"?   \tgithub.com/sanamhub/nepdate-go/patro\t[no test files]\n"
	got := parseTargets(out)
	want := []string{"FuzzParse", "FuzzAddDays"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("parseTargets = %v, want %v", got, want)
	}
}

func TestParseStatsUsesTotalCorpus(t *testing.T) {
	out := "fuzz: elapsed: 0s, execs: 0 (0/sec), new interesting: 0 (total: 11)\r\n" +
		"fuzz: elapsed: 3s, execs: 120 (40/sec), new interesting: 2 (total: 13)\r\n" +
		"fuzz: elapsed: 30s, execs: 900 (30/sec), new interesting: 3 (total: 14)\r\n"
	got := parseStats(out)
	if got.Execs != 900 || got.NewCorpus != 3 {
		t.Errorf("parseStats = %+v, want execs 900 and 3 new entries", got)
	}
}

func TestParseStatsKeepsCumulativeCountOnce(t *testing.T) {
	out := "fuzz: elapsed: 3s, execs: 50 (16/sec), new interesting: 4 (total: 8)\n" +
		"fuzz: elapsed: 6s, execs: 100 (16/sec), new interesting: 4 (total: 8)\n"
	got := parseStats(out)
	if got.Execs != 100 || got.NewCorpus != 4 {
		t.Errorf("parseStats = %+v, want execs 100 and 4 new entries", got)
	}
}

func TestParseStatsFallsBackToTotalCorpus(t *testing.T) {
	out := "fuzz: elapsed: 1s, execs: 50 (50/sec), total: 8\n" +
		"fuzz: elapsed: 2s, execs: 100 (50/sec), total: 14\n"
	got := parseStats(out)
	if got.Execs != 100 || got.NewCorpus != 6 {
		t.Errorf("parseStats = %+v, want execs 100 and 6 new entries", got)
	}
}

func TestCorpusDiffFindsCrasher(t *testing.T) {
	dir := t.TempDir()
	before, err := corpusFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(before) != 0 {
		t.Fatalf("a missing corpus listed %v", before)
	}
	path := filepath.Join(dir, "testdata", "fuzz", "FuzzParse", "crasher")
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("go test fuzz v1\nstring(\"x\")\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	after, err := corpusFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	added := addedFiles(before, after)
	if len(added) != 1 || added[0] != "FuzzParse/crasher" {
		t.Errorf("addedFiles = %v, want [FuzzParse/crasher]", added)
	}
}

func TestRunRejectsWideOrShortRuns(t *testing.T) {
	filter := regexp.MustCompile(".*")
	if err := run(0, 4, filter); err == nil {
		t.Error("a zero budget passed")
	}
	if err := run(30*time.Second, 5, filter); err == nil {
		t.Error("parallel 5 passed")
	}
	if err := run(30*time.Second, 0, filter); err == nil {
		t.Error("parallel 0 passed")
	}
}
