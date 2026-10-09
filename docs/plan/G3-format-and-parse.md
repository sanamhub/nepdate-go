# G3: Formatting and parsing

Read first: `shared/spec/FORMATTING.md`, `shared/spec/PARSING.md`, `docs/API.md` (formatting,
parse), PARITY D-04, D-05, D-06, D-15, `shared/spec/PORT-PLAN.md` phase 3.

Golden files are in `shared/spec/vectors/csharp-golden/`.

## Tasks

### G3-01 Names and digits
Month and weekday names as package-level `[12]string` / `[7]string` arrays (FORMATTING §1).
`Lang.AppendDigits`: Devanagari digit `d` is the 3 bytes `0xE0 0xA5 0xA6+d`; plain byte arithmetic,
no `[]rune`. Width counts digits, not bytes.
AC: `Nepali.AppendDigits(nil, 2081, 0)` is `२०८१`; `Nepali.AppendDigits(nil, 5, 2)` is `०५`;
`English.AppendDigits(nil, 5, 2)` is `05`.

### G3-02 `Short`, `Long`, `String` + `Append*` twins
The `Append*` methods do the work: `append` of name strings and digit bytes into the caller's
slice, nothing else (ADR-0002 §6). The string versions call the `Append*` version on a stack
buffer (`var buf [96]byte`; the longest output, Nepali long format with weekday, is 68 bytes) and
convert once, so they allocate exactly the returned string. `String()` writes its 10 bytes with
one `append` of fixed digit positions (no loop). No `fmt`, no `strconv`, no `strings.Builder`.
AC: golden `dates.tsv` columns `to_string`, `unicode`, `long_en`, `long_en_weekday`, `long_ne`,
`long_ne_weekday`, `dmy_dash_nopad` (all 1018 rows); the FORMATTING §3 and §4 examples;
`AppendShort`, `AppendLong` and `AppendFormat` allocate 0 times into a buffer with enough capacity.

### G3-03 `Format` / `AppendFormat` (FORMATTING §5)
AC: golden `format-pattern.tsv` (451 rows), skipping only rows whose pattern has `MMM` and whose
month is 3 or 6 (D-04), with a comment naming D-04; the FORMATTING §5 examples.

### G3-04 Strict `Parse` (PARSING §1)
ASCII or Devanagari digits (decode the 3-byte sequences directly), at most 4 digits per group.
Write the parser once as an unexported generic function
`parseStrict[T string | []byte](s T) (Date, error)`; `Parse` calls it with a `string` and
`UnmarshalText` (G4) with the `[]byte`, so neither converts `string(b)` up front (ADR-0002 §6:
that conversion allocates, because the input escapes into `*ParseError` on failure). Build the
`*ParseError` (and its `Input` copy) only on the failure path.
AC: the `strict` column of every row (57) of `shared/spec/vectors/parse.tsv`. `ERR:<Kind>` for
`InvalidCharacter`, `WrongGroupCount`, `NumberTooLong`, `Ambiguous`, `Unrecognized` means
`errors.As(err, &pe)` with `pe.Kind` equal to that kind; `ERR:OutOfRange`, `ERR:InvalidMonth`,
`ERR:InvalidDay` mean `errors.Is` with the sentinel (ADR-0003). A successful `Parse` allocates
0 times.

### G3-05 `ParseLenient` (PARSING §2)
The closed month-name list from the spec as a package-level table; compare Latin names with
`strings.EqualFold`; Devanagari variants listed explicitly (no `x/text` normalisation: zero
dependencies). Tokens are index pairs into the input kept in a fixed `[8]token` array on the
stack (more tokens than that is `ERR:Unrecognized` by PARSING §2 step 5), so a successful lenient
parse allocates nothing either.
AC: the `lenient` column of every row of `parse.tsv`, with the same error mapping as G3-04.

### G3-06 Fuzz tests
`FuzzParse`, `FuzzParseLenient`: never panic; an error is a `*ParseError` or one of the three
validation sentinels, nothing else; a successful parse formats with `String()` and re-parses to
the same date. Seed with every `input` of `parse.tsv` (`f.Add`), and commit any failing input
the fuzzer finds under `testdata/fuzz/`.
AC: `go test -run '^$' -fuzz '^FuzzParse$' -fuzztime 60s .` and the same with
`'^FuzzParseLenient$'` find nothing (run locally; CI runs only the seed corpus through `go test`).
Add `FuzzFormatRoundTrip(serial int32, order, sep uint8, pad bool, lang uint8)`: for a valid
serial and every `Short` that keeps year, month and day as numbers in Y-M-D order, `Parse` of
the output gives the same date.

### G3-07 Allocation, race and C# catalogue tests
- `alloc_test.go` subtests (each `== 0`, buffer created outside the measured function with
  capacity 128): `AppendShort`, `AppendLong` (both languages, with weekday), `AppendFormat`
  (pattern `yyyy-MM-dd` and `D`), `Lang.AppendDigits`, `Parse` (ASCII and Devanagari input),
  `ParseLenient` (`15 Shrawan 2080`), `Month.ShortName`. And each `== 1` (the returned string):
  `String`, `Short`, `Long`, `Format`. A failed `Parse` is allowed 1 allocation (the
  `*ParseError`).
- `TestConcurrentReads` gains `AppendLong`, `Parse`, `ParseLenient`.
- `TestCSharp` gains every catalogue row tagged `3`. Expected sources:
  `NepaliDateFormattableTests`, `NepaliDateParsableTests`, `NepaliDateMonthNameTests`,
  `SmartDateParserTests`, `StringExtensionsTests` and the `ToString`, `ToUnicodeString`,
  `ToLongDateString`, string-constructor and `SmartDateParser_*` rows of
  `OptimizationVerificationTests`. Smart-parser rows that rely on guessing, `yy` → `20yy` or
  fuzzy month names are D-06; Devanagari-digit strict rows are D-05 (we accept what C# rejects);
  `MMM` rows for months 3 and 6 are D-04.

## Checkpoint (agent stops here for review)

The reviewer checks:

- [ ] `go run ./internal/ci` ends with `ci: 5/5 gates passed`.
- [ ] Every text column of golden `dates.tsv`, all 451 rows of `format-pattern.tsv` (only D-04
      rows skipped, each with a comment) and all 57 rows of `parse.tsv` (both columns) pass.
- [ ] Both fuzz targets ran for 60 s locally with no finding (paste the last fuzz line).
- [ ] `go test -run Alloc -v .`: the `Append*` and successful-parse subtests report 0.
- [ ] `go build -gcflags=-m . 2>&1 | grep -E 'format|parse' | grep escapes` shows only the
      `*ParseError` and the string conversions inside the string-returning methods.
- [ ] No `fmt`, `strconv`, `unicode/utf8` loop over runes or `[]rune` in `format.go` and
      `parse.go` (`grep`).
- [ ] `TestCSharp` row count for phase `3` matches the catalogue; skipped rows name D-04, D-05
      or D-06.
