# AGENTS.md: rules for AI coding agents (Go port)

You are implementing a planned project. Follow the plan; don't redesign it.

## 0. Before writing any code

1. Read `README.md`, this file, every ADR in `docs/adr/`, and `docs/API.md`.
2. Open `docs/plan/00-ROADMAP.md`, find your phase, read that phase file and the spec sections it
   lists under "Read first". Check its preconditions; if one is unmet, stop and say so.

## 1. Sources of truth (in priority order)

1. `shared/spec/*.md` + `shared/spec/vectors/*`: behaviour. Vectors win over prose (report the
   disagreement).
2. `docs/adr/*`: decisions. 3. `docs/plan/*`: tasks and acceptance criteria. 4. `docs/API.md`:
   the exported surface.

`shared/` is a read-only copy of the hub. **Never edit it.** A data or spec problem is reported to
the human as a hub issue. Never use the C# NepDate docs as a source (`shared/spec/PARITY.md` §1).

## 2. Hard rules

- Zero dependencies in the root module, tests included (standard `testing`, no testify). Competitor
  libraries live only in `benchmarks/go.mod`.
- No exported identifier beyond `docs/API.md`.
- Package `nepdate` never imports `patro` and holds no tithi, holiday or event data (hub ADR-0004;
  CI gate 6 in ADR-0004 checks it).
- Zero allocations on every path ADR-0002 §5 lists. `go test` asserts it with
  `testing.AllocsPerRun`; a new allocation is a failing test, not a review comment.
- Hot paths (ADR-0002 §6): no `fmt`, no `interface{}`/`any` parameters, no closures stored in
  package variables, no `map`, no `[]rune`, no `strings.Builder`. Check new code with
  `go build -gcflags=-m` before asking for review.
- No `fmt` anywhere in library code (tests may use it). No `panic` except in `Must*` helpers. No `unsafe`, no cgo.
- Don't edit `*_gen.go` by hand. Change `internal/gen` and run `go generate ./...`.
- Parity tests may skip rows **only** for deviations in `shared/spec/PARITY.md`, with a comment
  naming the D-ID. Don't weaken or delete a test to make CI pass.
- No features outside the plan. Secrets never go in code, logs or commits.

## 3. How to work

- One task ID at a time (`G1-03`). Test first, see it fail, then implement.
- Run `go run ./internal/ci` (G0-03) before saying you're done; paste its last line in the PR.
- PR description: task IDs, what changed and why, how tested, "What to look at first" (2–3 bullets).

## 4. Writing

All prose follows [`.claude/skills/writing-style/SKILL.md`](.claude/skills/writing-style/SKILL.md),
for every agent. Doc comments start with the identifier's name (Go convention) and say what
fails. **No AI attribution** in commits or PRs.
