# Contributing

Bug reports, fixes and documentation improvements are welcome. For a new feature or an API change,
open an issue first so the design can be agreed before you write code. Behaviour is fixed by the
shared spec in `shared/spec/` and decisions by the ADRs in `docs/adr/`.

This project follows the [Code of Conduct](CODE_OF_CONDUCT.md). Report security problems as
described in [SECURITY.md](SECURITY.md), not in a public issue.

## Data and spec changes go to the hub

`shared/` is a read-only copy of [`sanamhub/nepdate`](https://github.com/sanamhub/nepdate). A wrong
month length, holiday or tithi, or a question about the spec, goes there as an issue. Never edit
`shared/` here, and never edit a `*_gen.go` file by hand: change `internal/gen` and run
`go generate ./...`.

## Checks

The module supports Go 1.24 or later (ADR-0008); the gates need Go 1.26 or later, and CI runs them
on 1.26 and 1.27 and the tests alone on 1.24.13. Before a PR:

```bash
go run ./internal/ci
```

The last line reads `ci: <n>/<n> gates passed`. `golangci-lint` and `govulncheck` are skipped when
they aren't installed, and the line then says so, for example `ci: 7/9 gates passed, 2 skipped
(golangci-lint, govulncheck)`. CI runs both in their own jobs. The race detector needs cgo, so on a
machine without a C compiler the tests run without `-race` and the line ends with `1 reduced (tests
without -race)`. The Linux CI job runs them with `-race`. The API gate fetches its pinned apidiff
tool with `go run`, so its first run needs the module proxy or a warm module cache.

`go run ./tools/fuzz-check` runs every fuzz target for 30 s. `testdata/fuzz/` holds the committed
corpus, which plain `go test` also runs as regression inputs. A file that appears there during a
fuzz run is a crasher: keep it, fix the bug, and commit both. To add inputs that found new
coverage, copy them from `$(go env GOCACHE)/fuzz/github.com/sanamhub/nepdate-go/<target>/`.

- Zero dependencies in the root module, tests included.
- No exported identifier beyond [`docs/API.md`](docs/API.md).
- The paths listed in ADR-0002 §5 allocate nothing; `alloc_test.go` checks it.
- Never weaken, skip or delete a test to get CI green. A skipped parity row names its deviation
  from `shared/spec/PARITY.md`.

## Pull requests

- Keep a pull request to one change. Fill in the PR template, including "What to look at first".
- Add a line to the `Unreleased` section of [CHANGELOG.md](CHANGELOG.md) for anything a user would
  notice.
- A significant design decision gets an ADR in `docs/adr/`. Accepted ADRs are not edited; a later
  ADR supersedes an earlier one.

## Commit messages

[Conventional Commits](https://www.conventionalcommits.org/): `type(scope): summary`, imperative,
lower case, no trailing period. Types: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `build`,
`ci`, `perf`. The body says why, not what.

## Releases

Releasing and rolling back are described in
[`docs/runbooks/release.md`](docs/runbooks/release.md).

By contributing you agree that your contribution is licensed under the MIT license
([LICENSE](LICENSE)).
