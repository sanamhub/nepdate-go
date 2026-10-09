# Roadmap (Go port)

Each phase is one PR (or a short series), sized so an agent finishes it in one session and a
reviewer reads it in about an hour. Task IDs `G<phase>-<n>`; branch `feat/G1-04-serial`; commit
`feat(nepdate): G1-04 serial conversion`.

```
G0 repo + CI ─► G1 core + conversion ─► G2 arithmetic/fiscal/range ─┐
                         └────────────► G3 format + parse ───────────┴─► G4 patro + interop ──► G5 docs, bench, release
```

Preconditions: hub task H0 merged, so `shared/VERSION` holds a hub commit SHA instead of `pending`;
hub H3 (reusable `shared-check` and `shared-sync`) merged before G0-05. G2 and G3 may run in
parallel; G4 needs both. Phases map one to one onto `shared/spec/PORT-PLAN.md` phases 0 to 5.
Hub H5 (C# test catalogue) is needed for the `TestCSharp` AC of G1 to G4 and hub H6 (C# benchmark
baseline) for the C# column in G5; if either is missing when a phase starts, do the rest of the
phase and report that AC as blocked at the checkpoint.

Every phase file ends with a **Checkpoint** section. The agent stops there, opens the PR and
waits for the human review; it does not start the next phase on its own.

| Phase | File | Output |
|---|---|---|
| G0 | `G0-repo-and-ci.md` | module, CI program, lint config, generator skeleton, community files |
| G1 | `G1-core-and-conversion.md` | `Date`, `Month`, errors, generated calendar table, BS↔AD, vectors, exhaustive |
| G2 | `G2-arithmetic-fiscal-range.md` | add days/months/years, diff, fiscal year, range |
| G3 | `G3-format-and-parse.md` | `Short`/`Long`/`Format` + `Append*`, strict and lenient parse, fuzz |
| G4 | `G4-patro-and-interop.md` | package `patro`, core-isolation gate, text marshalling, `Today`, `FromTime`/`Time` |
| G5 | `G5-docs-bench-release.md` | examples, benchmarks B1–B8 vs competitors and C# NepDate, release workflow, v0.1.0 |

## Definition of Done (every task)

1. Every acceptance criterion has a test (or a line in the PR saying why it can't).
2. `go run ./internal/ci` passes.
3. No exported identifier beyond `docs/API.md`, no dependency in the root module.
4. Allocation tests still report 0 for the paths in ADR-0002 §5, and the hot-path rules of
   ADR-0002 §6 hold (`go build -gcflags=-m` shows no new heap escape in the changed files).
5. `CHANGELOG.md` `[Unreleased]` updated for anything a user would notice.
6. Prose follows `.claude/skills/writing-style/SKILL.md`. PR has "What to look at first".

## Milestones

| Milestone | Phases | Exit |
|---|---|---|
| M1 converts | G0–G1 | month-boundary vectors + exhaustive round-trip pass |
| M2 parity | G2–G3 | all C# golden rows pass except `PARITY.md` deviations |
| M3 v0.1.0 | G4–G5 | core-isolation gate passes, BENCHMARK §6 targets met, tagged, on pkg.go.dev, benchmark table published |

## C# test catalogue by phase

`shared/spec/vectors/csharp-tests/catalog.tsv` (hub H5) has one row per `[Fact]` (395) and per
`[InlineData]` (147) of C# NepDate's tests at `cb05cb65`. Its `phase` column is the authority;
this table is the expected split by test file (counted from the C# sources on 2026-10-07), so the
reviewer can spot a file that went missing. ADR-0004 gives the Go test naming rule.

| C# test file | Facts | InlineData | Phase |
|---|---|---|---|
| `Core/NepaliDateConstructionTests` | 6 | 23 | G1 |
| `Core/NepaliDatePropertiesTests` | 9 | 0 | G1 |
| `Core/DictionaryIntegrityTests` | 7 | 4 | G1 |
| `Abilities/NepaliDateComparableTests` | 11 | 0 | G1 |
| `Core/OptimizationVerificationTests` | 59 | 74 | split: G1 (construction, comparison, weekday, min/max), G2 (subtraction), G3 (`ToString`, `ToUnicodeString`, `SmartDateParser`), D-10 (`BulkConvert_*`), D-12 (`IsLeapYear_*`) |
| `Core/NepaliDateArithmeticEdgeCaseTests` | 25 | 0 | G2 |
| `Core/NepaliDateManipulationTests` | 3 | 9 | G2 (fractional rows D-02, D-03) |
| `Core/FiscalYearTests` | 9 | 2 | G2 (D-08 shape) |
| `Core/FiscalYearInstanceMethodTests` | 38 | 7 | G2 (D-08 shape) |
| `Core/NepaliDateRangeTests` | 19 | 0 | G2 (set operations D-09) |
| `Core/NepaliDateRangeExtendedTests` | 64 | 0 | G2 (set operations, splitting, working days D-09) |
| `Abilities/NepaliDateFormattableTests` | 20 | 0 | G3 |
| `Abilities/NepaliDateParsableTests` | 5 | 0 | G3 |
| `Core/NepaliDateMonthNameTests` | 3 | 28 | G3 |
| `Core/SmartDateParserTests` | 14 | 0 | G3 (guessing rows D-06) |
| `Extensions/StringExtensionsTests` | 5 | 0 | G3 |
| `Core/CalendarDataTests` | 35 | 0 | G4 (internal-array rows `n/a`) |
| `Extensions/DateTimeExtensionsTests` | 5 | 0 | G4 |
| `Serialization/NepaliDateSerializationTests` | 9 | 0 | G4 (object form, Newtonsoft, XML: D-11) |
| `Core/NepaliDateIsRelativeTests` | 17 | 0 | G4 (D-14, through `FromTime`) |
| `Abilities/NepaliDateTypeConverterTests` | 18 | 0 | `n/a` (`TypeConverter` is .NET only) |
| `Core/BulkConvertTests` | 7 | 0 | D-10 |
| `Integration/MemoryUsageTests` | 4 | 0 | `n/a` (replaced by `alloc_test.go`) |
| `Integration/MemoryOptimizationBenchmark` | 3 | 0 | `n/a` (replaced by G5 benchmarks) |
| **Total** | **395** | **147** | |
