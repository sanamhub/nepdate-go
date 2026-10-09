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
| G1-07 | done | f854c1b | `ci: 5/5 gates passed` | `TestCSharp` lists all 113 phase-1 rows: 68 run, 9 skip with a D-ID (D-06, D-12), 3 `n/a`, 33 pending: 32 rows of the phase-1 catalogue construct dates from strings (need `Parse`, G3-04) and 1 needs `Today` (G4-05). They turn into real tests in those tasks. `TestCSharpCatalogueCovered` checks the table against `catalog.tsv`. Exhaustive test: 0.01 s |
| G1 checkpoint | ready for review | f854c1b | `ci: 5/5 gates passed` (clean clone, Windows, Go 1.27.1, no `-race` locally) | 33 phase-1 C# rows pending G3-04 and G4-05 (see G1-07). Changelog updated in this commit |
| G1 light check | ok | | | reviewer approved the estimate constants, int64 estimate, uint32 Compare and `TestCSharpCatalogueCovered`; pending markers are now the hub rule (PORT-PLAN, hub 94946ac) |
| G2-01 | done | e2ac536 | `ci: 5/5 gates passed` | `FuzzAddDays` ran 30 s locally (`-parallel 2`), no finding. Zero Date: `AddDays` returns `ErrOutOfRange`, `DaysUntil` returns 0, `Next`/`Prev` return false (D-01) |
| G2-02 | done | 456efa2 | `ci: 5/5 gates passed` | none |
| G2-03 | done | 5e44f2d | `ci: 5/5 gates passed` | at most two `AddMonths` calls, no loop over days |
| G2-05 | done | 060711d | `ci: 5/5 gates passed` | Done before G2-04 because `FiscalYear.Range` and `QuarterRange` return a `Range`. `Days` inlines (`-gcflags=-m`: `can inline Range.Days`), and ranging over it allocates 0 times |
| G2-04 | done | bb881b3 | `ci: 5/5 gates passed` | `QuarterRange` returns `ErrOutOfRange` for a quarter outside Q1..Q4 (the API and spec name no error for it; reviewer please confirm). `Label` needs Devanagari digits, so the unexported `appendDigits` lands here; G3-01 exposes it as `Lang.AppendDigits` |
| G2-06 | done | 5037b72 | `ci: 5/5 gates passed` | 14 new allocation subtests, all 0. `TestConcurrentReads` adds `AddDays`, `AddMonths`, `Diff`. `TestCSharp` now 293 rows (113 phase 1 + 180 phase 2) |
| G2 checkpoint | ready for review | 5037b72 | `ci: 5/5 gates passed` (clean clone, Windows, Go 1.27.1, no `-race` locally) | `TestCSharp` 293 rows: 186 run, 107 skip (D-09 48, D-06 10, D-02 4, D-12 3, n/a 4, pending 38: 33 need `Parse`, 5 need `FromTime`/`Today`). G2-05 done before G2-04 |
| G2 light check | ok | | | reviewer confirmed `QuarterRange` returns `ErrOutOfRange` for a quarter outside Q1..Q4 (the "argument outside the supported domain" error; Rust and Kotlin can't express an invalid quarter); the doc comment names it |
| G3-01 | done | 970118c | `ci: 5/5 gates passed` | `AppendDigits` wraps the `appendDigits` added in G2-04 |
| G3-02 | done | d2b6841 | `ci: 5/5 gates passed` | Zero Date and unknown `DateOrder`/`Separator` format as "" (D-01: zero value of the result). Longest output measured: 68 bytes (Nepali long with weekday) |
| G3-03 | done | b00565a | `ci: 5/5 gates passed` | 444 of 451 golden rows run, 7 skip for D-04 (`MMM` in Ashad or Ashoj). Interpretation, please confirm: the standard pattern `s` also uses Nepali digits with `Nepali` (FORMATTING §5 says "digits per lang" only for the default display; D-15 adds Nepali output to every standard format). A trailing lone `\` writes nothing |
| G3-04 | done | f87308e | `ci: 5/5 gates passed` | all 57 `strict` cells pass. The 33 "pending G3-04" C# rows are now real tests. Successful `Parse` (ASCII and Devanagari) and `parseStrict([]byte)` measured 0 allocations, a failed one 1 |
| G3-05 | done | e735716 | `ci: 5/5 gates passed` | all 57 `lenient` cells and every PARSING §3 spelling pass. The strict scan returns a kind instead of a `*ParseError` (`scanStrict`), so a successful lenient parse allocates 0 times even though its strict attempt failed. Tokens live in a fixed `[8]` array |
| G3-06 | done | 810953c | `ci: 5/5 gates passed` | 60 s each with `-parallel 2`, no finding: `FuzzParse` `fuzz: elapsed: 1m0s, execs: 1995613`; `FuzzParseLenient` `fuzz: elapsed: 1m1s, execs: 245431`; `FuzzFormatRoundTrip` `fuzz: elapsed: 1m0s, execs: 2030443`. No failing input, so nothing under `testdata/fuzz/` |
| G3-07 | done | (next commit) | `ci: 5/5 gates passed` | 10 new zero-allocation subtests (cap-128 buffer) and `TestAllocOne` (`String`, `Short`, `Long`, `Format` exactly 1; failed `Parse` at most 1). `TestConcurrentReads` adds `AppendLong`, `Parse`, `ParseLenient`. `TestCSharp` 441 rows (148 phase 3) |
