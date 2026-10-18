# G11: Minimum Go version and runnable examples

Read first: ADR-0001, ADR-0002, ADR-0004, ADR-0006, [`docs/API.md`](../API.md),
`shared/spec/FORMATTING.md`, `shared/spec/PATRO.md` and `shared/spec/PARSING.md`.

Precondition: G10 merged at 2c5b28d, with the review commit 83959b8. The API gate now requires
declaration forms in `docs/API.md`, the fuzz jobs upload `testdata/fuzz/` on failure, and the fuzz
corpus holds 281 files. The local line is `ci: 9/9 gates passed, 1 reduced (tests without -race)`.
The machine is shared, so this phase runs no benchmarks.

`go.mod` says `go 1.24`, but every gate, workflow and tool runs under Go 1.27.1, so the minimum
version is never exercised. The module uses `iter`, `encoding.TextAppender` and `b.Loop`, all from
Go 1.24. The phase tests that claim, adds examples for the public surface, and makes the examples a
docs-gate obligation.

## Tasks

### G11-00 Plan
This file, the roadmap row and the progress rows. Commit before the Go 1.24 install or any example
change.

AC:
- The roadmap diagram and phase table name G11 and its three work tasks.
- The plan names the Go 1.24 test procedure, the example scope, the gate extension and the stop
  point.
- `docs/plan/PROGRESS.md` has a row for every G11 task before its work lands.
- No production code, test, workflow or `go.mod` change in this commit.

### G11-01 Minimum Go version
Install the latest Go 1.24 patch, 1.24.13, from the official Windows amd64 archive into
`D:\devcache\g11\go1.24`. Verify the SHA-256 from `go.dev/dl` and record `go version`. Use
`D:\devcache\g11\gocache`, `gomodcache`, `gopath` and `bin`, with `GOTOOLCHAIN=local`, so no
newer toolchain is downloaded and nothing is installed on C:.

Run `go test ./...` and `go run ./tools/consumer-check` with that toolchain. The consumer check
builds a throwaway module outside the checkout, imports both packages through `replace`, and must
still print its pinned date and event lines.

Add the oldest supported version to the `ci` matrix in `.github/workflows/ci.yml`: Go `1.24.x`,
`ubuntu-latest` only, running `go test ./...` and not the quality gates. The action steps stay
SHA-pinned, `permissions: contents: read` stays, and the existing 1.26, 1.27 and 386 entries do not
change.

If Go 1.24 fails, repair only test or tool code that uses a newer feature, without a public API or
behaviour change. If the library itself needs a newer release, raise the `go` line in `go.mod` to
the oldest version that passes, update README, CONTRIBUTING and the matrix, and record the evidence
in a new ADR. Do not edit ADR-0001.

AC:
- `D:\devcache\g11\go1.24\bin\go.exe version` reports `go version go1.24.13 windows/amd64`.
- With that toolchain, `go test ./...` exits 0 and `go run ./tools/consumer-check` exits 0.
- With that toolchain, `go list -m all` lists only the root module and `go.sum` is absent.
- `ci.yml` has a Go `1.24.x` matrix entry on `ubuntu-latest` that runs `go test ./...` only.
- `actionlint` accepts `ci.yml`.
- `go run ./internal/ci` under Go 1.27.1 still ends with
  `ci: 9/9 gates passed, 1 reduced (tests without -race)`.
- If the `go` line changed, ADR-0008 names the oldest passing version and the failing 1.24
  evidence, and the README states the supported range.
- If the line did not change, README still states the supported range and CONTRIBUTING no longer
  says Go 1.26 or later.

Result on 2026-10-18: Go 1.24.13 installed at `D:\devcache\g11\go1.24` from the official zip, SHA-256
`40b16bc8f00540a2cb02dff4de72b73e966fdd8d65f95e33d8e4080b48a2459a`. With `GOTOOLCHAIN=local` and
the G11 caches, `go test -count=1 ./...` passed all 10 packages and `go run ./tools/consumer-check`
printed the pinned lines. The `go` line stays 1.24, so no ADR was needed.

The first consumer run failed before reaching the module: `filepath.Rel` cannot compare the C:
temp directory with the E: checkout, so `outsideRepo` returned
`consumer-check: Rel: can't make C:\Users\... relative to E:\github\sanamhub\nepdate-go`. This
was a tool bug, not a Go 1.24 difference, and it also affects the Windows CI job, whose checkout is
on D: and temp on C:. `outsideRepo` now treats a different Windows volume as outside, with
`TestOutsideRepoDifferentVolume` on Windows. The positive and negative checks pass after the fix.

`ci.yml` adds `1.24.x` to the matrix on `ubuntu-latest` only, with a `go test ./...` step and the
other entries still running `go run ./internal/ci`. `actionlint` accepted the workflow.

### G11-02 Examples
Today there are six `Example*` functions: `ExampleNew`, `ExampleDate_Format`,
`ExampleDate_AppendLong`, `ExampleParseLenient`, `ExampleFromTime` and `ExampleOf`. Add runnable
examples with `// Output:` for every exported package-level function and every exported type:
16 functions and 16 types across `nepdate` and `patro`. Add method examples for the arithmetic,
`Diff`, the fiscal and range types, `Short`, `Long` and `Format`, `AppendDigits`, text marshalling
through `encoding/json`, and the `patro.Info` query methods.

Each example shows one realistic use and prints a fixed value. The inputs are hub-backed dates:
2081-04-15 (Tuesday, Dashami, three events), 2081-04-32 (AD 2024-08-16), 2081-01-01 (Panchami,
public holiday, New Year and Bisket Jatra), 2081-02-15 (Republic Day) and 2080-11-01 (five events).
`Today` prints only the midnight time from `Date.Time`, so its output does not depend on the clock.
No example calls an unexported helper or a test-only reference.

`ExampleNew`, `ExampleDate_Format`, `ExampleDate_AppendLong`, `ExampleParseLenient`,
`ExampleFromTime` and `ExampleOf` are the README quick start. Keep their bodies character for
character unless the README changes in the same commit, so `TestREADMEQuickStart` still passes.

AC:
- `go test -run '^Example' -v ./...` lists and passes every example, with at least one example
  for each of the 16 functions and 16 types.
- Every example function has an `// Output:` comment and a nonzero expected block.
- The recorded example count rises from 6 to the number in this file, and the file records it.
- `TestREADMEQuickStart` passes and no README block changed.
- `go test ./...` and `go vet ./...` pass.
- No public API or production behaviour change.

### G11-03 Example gate
Extend gate 8 (docs) in `internal/ci`; do not add a gate number. The new check parses the non-test
files of `nepdate` and `patro` for exported package-level functions and exported types, parses the
`_test.go` files for `Example*` functions, and fails when a function or type has no example whose
name starts with `Example<Name>` (an optional `_suffix` is allowed, so a method example such as
`ExampleDate_Format` covers the `Date` type).

The allow-list is a Go map with a comment for each entry. It is empty while every function and type
has a deterministic example. An allow-list entry that is not needed fails, so a stale exemption
cannot hide a missing example.

Add unit tests for a missing function, a missing type, a method-only example, an allow-listed name,
a stale allow-list entry and a name that only appears in prose.

AC:
- `go test ./internal/ci` passes, including the new tests.
- `TestGateCount` still reports 9 gates, and the gate order is unchanged.
- An exported function with no example fails gate 8, naming the package, kind and identifier.
- The allow-list is explicit, commented and empty when no entry is needed.
- A negative control adds an exported function and its doc comment without an example; gate 8
  fails with the exact line and exits 1.
- Restored, `go run ./internal/ci` ends with
  `ci: 9/9 gates passed, 1 reduced (tests without -race)` and the tree is clean.
- `golangci-lint` reports 0 issues and `actionlint` accepts the workflows.

## Checkpoint

Stop after G11-03. Report the commits, the real local `ci` line verbatim, the Go 1.24 result, the
example count before and after, the allow-list, the negative control, and every doubtful point. Do
not tag, publish, add a remote, change repository settings or start a later phase.

## Out of scope

No public API or behaviour change, no new dependency and no benchmark run. Fuzz, mutation, lint and
govulncheck keep their current budgets and jobs. GitHub CI is not measured because this checkout has
no remote.
