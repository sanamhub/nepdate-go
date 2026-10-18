# Contributing

This is the Go port of the NepDate family. Behaviour is fixed by the shared spec in `shared/spec/`
and decisions by the ADRs in `docs/adr/`, so most changes follow a task in `docs/plan/`. The rules
for agents, which also apply to people, are in [`AGENTS.md`](AGENTS.md).

## Data and spec changes go to the hub

`shared/` is a read-only copy of [`sanamhub/nepdate`](https://github.com/sanamhub/nepdate). A wrong
month length, holiday or tithi, or a question about the spec, goes there as an issue. Never edit
`shared/` here, and never edit a `*_gen.go` file by hand: change `internal/gen` and run
`go generate ./...`.

## Checks

Go 1.26 or later. Before a PR:

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

## Commits and PRs

- [Conventional Commits](https://www.conventionalcommits.org/): `type(scope): summary`, with the
  plan's task ID first when there is one (`feat(nepdate): <task ID> serial conversion`).
- One task per PR. Fill in the PR template, including "What to look at first".
- Prose, comments and commit messages follow
  [`.claude/skills/writing-style/SKILL.md`](.claude/skills/writing-style/SKILL.md): no em dashes,
  no filler, exact numbers and IDs, and no AI attribution lines.

By contributing you agree that your contribution is licensed under the MIT license
([LICENSE](LICENSE)).
