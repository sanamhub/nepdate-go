# G4: Patro package and standard-library interop

Read first: `shared/spec/PATRO.md`, `shared/spec/PORT-PLAN.md` phase 4, hub ADR-0004 (patro in a
separate package), ADR-0001 §1, ADR-0002 §2 and §5 (patro layout and allocation rules),
`docs/API.md` (`patro`, text encoding, time interop), ADR-0003.

Package `patro` is `github.com/sanamhub/nepdate-go/patro` in directory `patro/`, in the root
module (no second `go.mod`). It imports `nepdate`; `nepdate` must never import `patro`. No type
named `Patro` is needed: Go has package-level functions, so `patro.Coverage()` and
`patro.PublicHolidays()` play that role (hub ADR-0004 §5).

## Tasks

### G4-01 Patro tables (`patro/patro_gen.go`)
Extend `internal/gen` to read `shared/data/patro/*.json` and write `patro/patro_gen.go`
(`package patro`) with the layout of ADR-0002 §2:

- `const covFirst, covLast` (serials of 2077-01-01 and 2084-08-20: 64285 and 67077).
- `dayInfo [2793]uint32`, one record per day of the coverage: bits 0–5 tithi name id + 1 (0 =
  no tithi), bit 6 public holiday, bits 7–10 event count, bits 11–22 index of the day's first
  entry in `eventName`. A `// 2081-01-01` comment per entry.
- `eventName [3032]uint16`: name ids of all events, day by day, in file order.
- `const nameText = "..."` (all 753 deduplicated (np, en) pairs, tithi names first, as one string
  constant, 59,800 bytes today) and `nameOff [1507]uint32`: pair `i` in language `l` (0 =
  English, 1 = Nepali, the `nepdate.Lang` values) is `nameText[nameOff[2*i+l]:nameOff[2*i+l+1]]`.

The generator fails (exit 1, message naming the limit) if a tithi day is missing inside the tithi
range (PATRO §3), if there are 63 or more tithi names, 16 or more events on one day, 4096 or more
events in total, so a data update can never overflow a bit
field silently. `-check` covers this file too. Nothing from `shared/data/patro/` goes into any
file of package `nepdate`.
AC: counts T = 2190, H = 294, E = 1671, 3032 events, 753 pairs (PATRO §3; a test in
`internal/gen` counts them from the generated tables); the generator exits non-zero on a copy of
the data with one tithi day removed, and on a copy with 16 events on one day (unit tests in
`internal/gen`).

### G4-02 Package `patro`
`Of`, `Info.Tithi`, `IsPublicHoliday`, `Events() iter.Seq[Event]`, `EventCount`, `Event(i)`,
`Coverage`, `PublicHolidays`, `Tithi.Name`, `Event.Name`. `Of` returns the empty `Info` for
`d.IsZero()`; otherwise the serial is `nepdate.Min.DaysUntil(d)` (exported, inlined, no
allocation), so `patro` needs no unexported access to `nepdate` and no extra internal package.
`Of` is then one subtraction, one bounds check and one array read. `PublicHolidays` walks
`dayInfo` and yields the days with bit 6 set.
AC: every row of the PATRO §4 anchors (tithi, holiday and events in order, in both languages);
`Coverage()` is 2077-01-01..2084-08-20; ranging over `PublicHolidays()` yields 294 dates in
ascending order; `Of(nepdate.Date{})` and `Of` of 2070-06-15 and 2090-01-01 are empty;
`Info{}.Event(0).Name(nepdate.English) == ""`.

### G4-03 Core isolation gate (hub ADR-0004, ADR-0004 gate 6)
Add `internal/coreonly/main.go` (`package main`, imports only `github.com/sanamhub/nepdate-go`,
converts one date and prints it) and `internal/withpatro/main.go` (imports `patro`, prints the
English events of 2081-04-15). Add gate 6 to `internal/ci` exactly as ADR-0004 describes. Gate 3
now also measures `./patro`.
AC: `go run ./internal/ci` ends with `ci: 6/6 gates passed`; `go test .` passes with the `patro/`
directory deleted from a scratch copy of the repo; adding
`import _ "github.com/sanamhub/nepdate-go/patro"` to `internal/coreonly/main.go` makes gate 6
fail (try it locally, don't commit it).

### G4-04 Text encoding
`MarshalText`, `AppendText`, `UnmarshalText` as `YYYY-MM-DD`; `UnmarshalText` calls the generic
strict parser with the `[]byte` directly (G3-04).
AC: `json.Marshal(struct{ D nepdate.Date }{d})` gives `{"D":"2081-04-15"}`; a `map[nepdate.Date]int`
round-trips through JSON; `UnmarshalText([]byte("2081-04-33"))` returns `ErrInvalidDay`;
`AppendText` and a successful `UnmarshalText` allocate 0 times.
`FuzzUnmarshalText`: same properties as `FuzzParse` (ADR-0004).

### G4-05 `time` interop
`FromTime` (uses `t.Date()` in `t`'s location), `Time(loc)`, `Today(loc)`. `Today(loc)` is
`FromTime(time.Now().In(loc))`, so the clock-dependent AC is tested through `FromTime`.
`Today(nil)` loads `Asia/Kathmandu` once through a package-level `sync.OnceValue` (falling back to
`time.FixedZone("NPT", 20700)`), because `time.LoadLocation` reads the zone database and allocates
on every call. `Today` is not a hot path; it is in the API for convenience.
AC: `FromTime` of 2024-07-30 18:30 UTC is BS 2081-04-15 in UTC but 2081-04-16 in
`time.FixedZone("NPT", 20700)`; the doc comment says why; `MustNew(2081, Shrawan, 15).Time(nil)`
is 2024-07-30 00:00 UTC; `FromTime` allocates 0 times.

### G4-06 Allocation, race and C# catalogue tests
- `patro/alloc_test.go` subtests (each `== 0`): `Of`, `Info.Tithi` plus `Tithi.Name` (both
  languages), `IsPublicHoliday`, a full `for e := range info.Events()` reading every `Name`, the
  `Event(i)` loop, `Coverage`, a full range over `PublicHolidays()`.
- `TestConcurrentReads` in `patro`: 8 goroutines call `Of` and read names for every day of the
  coverage.
- `TestCSharp` gains every catalogue row tagged `4`. Expected sources: `CalendarDataTests` (rows
  about C#'s internal arrays, such as `CalendarOffsets_*`, `EventData_SliceStart_*`, are replaced by
  the generator count tests and named in the skip text as `n/a: internal layout`),
  `DateTimeExtensionsTests` (`FromTime`), `NepaliDateSerializationTests` (string form only; the
  object form, Newtonsoft and XML rows are D-11), and `NepaliDateIsRelativeTests` (D-14, tested
  through `FromTime` with a fixed clock).

## Checkpoint (agent stops here for review)

The reviewer checks:

- [ ] `go run ./internal/ci` ends with `ci: 6/6 gates passed`.
- [ ] Gate 6 fails with the blank import added to `internal/coreonly` (the PR says it was tried).
- [ ] `go run ./internal/gen -check` exits 0; `patro/patro_gen.go` has no `func init`; the
      bit-field limit tests in `internal/gen` pass.
- [ ] Every PATRO §4 anchor passes in both languages.
- [ ] `go test -run Alloc -v ./patro`: every subtest reports 0.
- [ ] `go list -deps . | grep patro` prints nothing.
- [ ] `TestCSharp` row count for phase `4` matches the catalogue; skipped rows name D-11, D-14 or
      `n/a`.
- [ ] `FromTime` doc comment explains the UTC versus Nepal one-day difference.
