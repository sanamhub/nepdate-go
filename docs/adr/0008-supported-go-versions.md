# ADR-0008: Supported Go versions

- Status: Accepted
- Date: 2026-10-18
- Deciders: Sanam Pakuwal
- Amends: ADR-0001 (toolchain, point 2)

## Context

ADR-0001 sets `go 1.24` in `go.mod` and says CI runs the two supported releases, Go 1.26 and 1.27.
Until this decision no workflow ran Go 1.24, so the `go` line was a claim nobody tested. A local
run with Go 1.24.13, the latest 1.24 patch, ran `go test ./...` and the consumer check with
`GOTOOLCHAIN=local`. Both passed without a code change. The only failure was a cross-volume bug in
`tools/consumer-check`, which was not a Go version difference.

The gates use tools (golangci-lint, govulncheck, apidiff, the `go test` vet check for newer
standard-library symbols) that are pinned for the current Go release and do not promise to run on
1.24.

## Decision

1. The supported range is Go 1.24 or later. `go.mod` keeps `go 1.24`.
2. The `ci` matrix runs the full gates (`go run ./internal/ci`) on Go 1.26 and 1.27 on Linux,
   Windows and macOS, and on 1.27 for `GOARCH=386`, as before.
3. The oldest supported version runs `go test ./...` only, on `ubuntu-latest`, at the exact patch
   that was verified locally: `1.24.13`, not `1.24.x`. The step sets `GOTOOLCHAIN=local`, so a
   future `toolchain` line cannot silently test a newer Go. A later 1.24 patch is adopted by
   changing the pin after a local run, in the same change.
4. Raising the `go` line is an API decision for users and needs a new ADR that names the oldest
   passing version and why the old one failed. The pin and README move with it.

## Consequences

- A change that uses a standard-library feature newer than 1.24 fails the 1.24.13 job even when
  the 1.27 vet check misses it (for example a language change rather than a symbol).
- Lint and vulnerability results describe Go 1.27 builds only. A user on 1.24 gets tested
  behaviour, not linted builds.
- The 1.24 job does not run the race detector, the coverage floor, lint, or the docs and
  API gates. Those are version-independent or covered on 1.27; fuzz seed corpora still run as
  ordinary tests.
- The 1.24.13 pin goes stale as Go ships 1.24 patches; that is accepted, since the job tests the
  version we ran, and 1.24 itself is out of upstream support.
