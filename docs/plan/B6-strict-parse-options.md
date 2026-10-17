# B6 strict-parse options

Status: options for review. Nothing here is implemented.

B6 is `nepdate.Parse(text)` over the 64 `String()` texts of the benchmark dates, so every input is
exactly `YYYY/MM/DD`: 10 ASCII bytes, four year digits, a separator, two month digits, a separator,
two day digits (`benchmarks/inputs_test.go`). B6 measures 34.80 ns per date against C# NepDate's
21.55 ns, so it misses the ADR-0002 section 7 port target (every operation faster than C#), though
not the section 6 release blocker. The 2 s CPU profile puts 68.8 % of the samples in `scanStrict`
and 7.4 % in the inlined `devaAt` (`docs/benchmarks.md`).

The 34.80 ns baseline comes from a loaded machine (CPU load 1 to 71 %). Step 0 is a quiet rerun of
the current code with the BENCHMARK section 4 command and `benchstat`: if the current code already
meets 21.55 ns, B6 is a measurement artifact and no code changes.

## What the profile rules out

Removing only `devaAt` caps the gain at 7.4 %, about 32.2 ns, still above 21.55. The scan loop is
the target: its per-byte work is a three-byte Devanagari test, a switch over the digit and
separator cases, and the group, digit-count and overflow bookkeeping. The 17.8 ns ASCII-only
prototype in ADR-0002 section 8 shows that a direct ASCII path can clear the target.

## Option A: canonical-shape ASCII fast path (recommended)

At the top of `scanStrict`, accept exactly 10 bytes: digits at 0 to 3, 5 to 6 and 8 to 9, and a
strict separator at 4 and 7. Build the year, month and day with direct digit arithmetic, then call
`newFromInts` so the range and day checks and their error order stay exactly as now. Anything else
falls through to the existing loop unchanged, so the accepted language, the error kinds and the
zero-allocation guarantee stay with the current code. Put the path in `scanStrict` rather than
`parseStrict` so `ParseLenient` benefits too.

Expected: 16 to 20 ns per date, 1.7 to 2.2 times, under 21.55. The estimate takes the scan share
as 76.2 % and leaves the `newFromInts`, `New` and call overhead unchanged: a 70 % scan cut gives
`34.80 * (0.238 + 0.762 * 0.30)` = 16.3 ns, a 60 % cut gives 18.9 ns, and a 50 % cut only reaches
21.5 ns. The 17.8 ns prototype lands in the same range.

Risk: low. The fast path is a strict subset, so any shape it does not recognize takes the old
path. The duplicated digit and separator checks can drift, so a test must compare the fast path
with `scanStrict` for every date and separator, and `FuzzParse` must keep running. Code: about 25
lines plus the test.

## Option B: ASCII-specialized scanner

Keep the general loop but handle an ASCII byte without calling `devaAt`: branch on `c < 0x80` and
use a switch with the same digit and separator cases; handle the Devanagari lead byte in the other
branch. This removes the 7.4 % `devaAt` share and some of the switch work, but the loop and the
group bookkeeping remain.

Expected: 28 to 31 ns, 1.1 to 1.25 times, above 21.55. It does not meet the target on its own. It
is the lowest-risk change and leaves the parser shape unchanged, but the gain does not justify the
change unless it is combined with Option C. Code: about 40 changed lines.

## Option C: byte-class table

Replace the per-byte switch with a `[256]uint8` class table: digit value, separator, Devanagari
lead, invalid. The loop reads one class per byte and updates the group state; a Devanagari lead
consumes the next two bytes. This removes the nested `devaAt` checks and shortens the switch, but
it keeps a table lookup and the group state.

Expected: 21 to 24 ns, 1.45 to 1.65 times, on the edge of 21.55. It handles every shape and is the
only option that is not a shape special case. Risk: medium, because the loop is the parser and
every state transition must keep the current error kinds; the vectors, the C# catalogue and the
fuzz targets must pass unchanged. Code: about 60 lines.

## Comparison

| Option | Expected B6 | Speedup | Meets 21.55 | Code | Risk |
|---|---:|---:|---|---:|---|
| A canonical fast path | 16 to 20 ns | 1.7 to 2.2x | yes | about 25 lines | low |
| B ASCII-only `devaAt` skip | 28 to 31 ns | 1.1 to 1.25x | no | about 40 lines | low |
| C byte-class table | 21 to 24 ns | 1.45 to 1.65x | borderline | about 60 lines | medium |

The numbers are estimates from the profile shares and the 17.8 ns prototype. The only number that
counts is a new B6 run.

## Recommendation

Option A, then a quiet measurement. It targets the actual input shape, keeps the current scanner as
the fallback, and has the smallest correctness surface. Option C is the choice if the reviewer
rejects a shape special case; Option B is not enough alone.

## Measurement that confirms it

From `cmd` in `benchmarks/`, with other applications closed, use the command in
[`docs/benchmarks.md`](../benchmarks.md): `go test -c`, run with `/high /affinity 4`, 10 counts of
300 ms, then `benchstat -col /library`. Add `-test.cpuprofile` to check that the `scanStrict` and
`devaAt` shares fall. The acceptance test is the median B6 `ns/date` under 21.55 with 0 B/op and
0 allocs/op, and `go test ./...`, the parse vectors and the fuzz targets unchanged.
