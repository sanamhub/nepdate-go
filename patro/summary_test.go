package patro

import (
	"encoding/json"
	"os"
	"testing"
)

// patroSummary is shared/spec/vectors/patro-summary.json (PATRO §3): the hub's counts for the
// data these tables were generated from. The tests compare the tables with it instead of pinning
// numbers that a data update changes.
type patroSummary struct {
	Schema           int    `json:"schema"`
	FirstYear        int    `json:"first_year"`
	LastYear         int    `json:"last_year"`
	FirstDay         string `json:"first_day"`
	LastDay          string `json:"last_day"`
	FirstSerial      int    `json:"first_serial"`
	LastSerial       int    `json:"last_serial"`
	Days             int    `json:"days"`
	TithiFirstDay    string `json:"tithi_first_day"`
	TithiLastDay     string `json:"tithi_last_day"`
	TithiFirstSerial int    `json:"tithi_first_serial"`
	TithiDays        int    `json:"tithi_days"`
	TithiNames       int    `json:"tithi_names"`
	Holidays         int    `json:"holidays"`
	EventDays        int    `json:"event_days"`
	Events           int    `json:"events"`
	EventNames       int    `json:"event_names"`
	Names            int    `json:"names"`
	NameBytes        int    `json:"name_bytes"`
}

func loadSummary(t *testing.T) *patroSummary {
	t.Helper()
	raw, err := os.ReadFile("../shared/spec/vectors/patro-summary.json") //nolint:gosec // test reads the repository's data
	if err != nil {
		t.Fatalf("shared/spec/vectors/patro-summary.json: %v", err)
	}
	var s patroSummary
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("shared/spec/vectors/patro-summary.json: %v", err)
	}
	if s.Schema != 1 {
		t.Fatalf("patro summary schema %d, want 1", s.Schema)
	}
	return &s
}
