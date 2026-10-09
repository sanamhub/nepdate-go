# Progress

One row per task (handover §3 step 7). The commit column names the task's own commit; the hash
is filled in by the next commit.

| Task | Status | Commit | CI last line | Notes and open questions |
|---|---|---|---|---|
| G0-01 | done | c13098a | n/a (CI program comes in G0-03) | `go build`, `go vet`, `go list -m all` pass |
| G0-02 | done | 762e16c | n/a (CI program comes in G0-03) | two runs give identical bytes; `-check` exits 1 and names `calendar_gen.go` after one byte changes |
| G0-03 | done | 546b74a | `ci: 5/5 gates passed` | lint and govulncheck skipped locally (not installed yet); `release-check v9.9.9` and `release-check 0.1.0` exit 1 |
| G0-04 | done | 4b5b802 | `ci: 5/5 gates passed` | golangci-lint 2.14.0 and govulncheck 1.8.0 installed with `go install`; lint reports 0 issues; gosec findings in the dev tools are suppressed with a reason each |
| G0-05 | blocked (partial) | e024eae | `ci: 5/5 gates passed` | Done: `ci.yml` (os x Go 1.26/1.27 plus linux/386, lint and govulncheck jobs, actions pinned by SHA), Dependabot, PR and issue templates, `SECURITY.md`, `CONTRIBUTING.md`. Blocked on hub H3: the `shared-check` job in `ci.yml` and `shared-sync.yml` need the hub reusable workflows. Not verifiable locally: "CI green on the first PR" needs the GitHub repository (human task). |
| G0 checkpoint | ready for review | e024eae | `ci: 5/5 gates passed` (clean clone, Windows, Go 1.27.1; gate 3 warned and ran without `-race`, `CGO_ENABLED=0`) | G0-05 partial: shared-check and shared-sync wait for hub H3 |
| G0 light check | ok | | | reviewer reran CI: `ci: 5/5 gates passed`; G0-05 stays blocked on hub H3 |
| G1-01 | done | e2c1ed0 | `ci: 5/5 gates passed` | Estimate constants found by the generator: BS mul 34449 add -560448, AD mul 34451 add 2963025 (ADR-0002 §3 quotes 34443 from the prototype; the generator picks the multiplier nearest 2^20 / mean month length and the middle of its valid add interval, then verifies every serial) |
| G1-02 | done | 28520c2 | `ci: 5/5 gates passed` | `Lang` gets only its type and constants here; `AppendDigits` is G3-01 |
| G1-03 | done | 51ce26b | `ci: 5/5 gates passed` | none |
| G1-04 | done | 802535c | `ci: 5/5 gates passed` | `go build -gcflags=-m`: `serial.go:6:6: can inline toSerial`, `serial.go:13:6: can inline fromSerial` |
| G1-05 | done | 64167ee | `ci: 5/5 gates passed` | Compare packs the fields into one uint32 and uses `cmp.Compare` (same result as field-by-field, inlines) |
| G1-06 | done | 3868159 | `ci: 5/5 gates passed` | The oracle reads `bs-calendar.json` itself (not the generated tables) and is checked against ALGORITHM §9 by `TestReferenceOracle` |
| G1-07 | done | (next commit) | `ci: 5/5 gates passed` | `TestCSharp` lists all 113 phase-1 rows: 68 run, 9 skip with a D-ID (D-06, D-12), 3 `n/a`, 33 pending: 32 rows of the phase-1 catalogue construct dates from strings (need `Parse`, G3-04) and 1 needs `Today` (G4-05). They turn into real tests in those tasks. `TestCSharpCatalogueCovered` checks the table against `catalog.tsv`. Exhaustive test: 0.01 s |
