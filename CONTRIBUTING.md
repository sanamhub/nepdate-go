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

The last line reads `ci: <n>/<n> gates passed`. `golangci-lint` and `govulncheck` are skipped with a
warning when they aren't installed; CI always runs them.

- Zero dependencies in the root module, tests included.
- No exported identifier beyond [`docs/API.md`](docs/API.md).
- The paths listed in ADR-0002 §5 allocate nothing; `alloc_test.go` checks it.
- Never weaken, skip or delete a test to get CI green. A skipped parity row names its deviation
  from `shared/spec/PARITY.md`.

## Commits and PRs

- [Conventional Commits](https://www.conventionalcommits.org/): `type(scope): summary`, with the
  task ID first when there is one (`feat(nepdate): G1-04 serial conversion`).
- One task per PR. Fill in the PR template, including "What to look at first".
- Prose, comments and commit messages follow
  [`.claude/skills/writing-style/SKILL.md`](.claude/skills/writing-style/SKILL.md): no em dashes,
  no filler, exact numbers and IDs, and no AI attribution lines.

By contributing you agree that your contribution is licensed under the MIT license
([LICENSE](LICENSE)).
