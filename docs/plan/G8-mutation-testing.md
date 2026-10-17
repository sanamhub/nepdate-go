# G8: Mutation testing

Read first: ADR-0004, ADR-0006, [`G7-b6-fast-path.md`](G7-b6-fast-path.md).

Precondition: G7 merged on `main` (bebe42e). Mutation testing measures the tests, not the library:
a surviving mutant is a test gap, and the answer is a new test or a written equivalence argument.
Production code is never weakened to raise the score.

Tool: Gremlins v0.6.0, package `github.com/go-gremlins/gremlins/cmd/gremlins`. The version is
pinned in the `gremlinsVersion` constant of `tools/mutation-check`, which runs
`go run github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0`. The root module keeps zero
dependencies: the download is a build cache entry, not a `go.mod` requirement.

## Tasks

### G8-01 Plan
This file, the roadmap row and the progress row. Commit before the first mutation run.

AC:
- The plan lists the tool, its pin and every later task.
- The root module's `go.mod` and the absent `go.sum` are unchanged.

### G8-02 Baseline
`tools/mutation-check` runs Gremlins on the core package with `--workers=4 --test-cpu=1` and
`--output=<json>`. The score is the JSON `test_efficacy`, which is `KILLED / (KILLED + LIVED)`.
Survivors are the `LIVED` and `NOT COVERED` mutants, grouped by file.

AC:
- The run uses at most 4 workers and writes machine-readable JSON.
- The baseline score, the survivor count and the survivor table by file are recorded.
- Every survivor has a file, line, mutant type and status.

### G8-03 Kill or classify
Kill every survivor with a test, or argue in the report why the mutant is equivalent. An
equivalent mutant gets a comment in the test that covers the line, not a code change that hides
the mutation.

AC:
- Every survivor from G8-02 has a disposition: killed by a named test, or equivalent with a reason.
- The final score and survivor count are recorded.
- No production code is weakened: no check, error path or boundary is removed.

### G8-04 Gate
`tools/mutation-check` enforces both numbers:

- `test_efficacy` is compared as an unrounded float against a threshold set just below the final
  score.
- `LIVED + NOT COVERED` is compared as an exact integer against a survivor ceiling.

The K7 lesson: a rounded percentage can hide a regression. 95.4 % must fail a 96 % gate, and one
extra survivor must fail even when the rounded percentage stays the same. The CI job `mutation`
runs the gate with `--workers=4`.

AC:
- The gate fails below the score threshold and above the survivor ceiling.
- The CI job runs the gate with at most 4 workers.
- Negative control: delete one test that kills a mutant, show the gate fails, restore the test and
  show it passes.
- `go run ./internal/ci` still ends with `ci: 8/8 gates passed`.

## Out of scope

No tags, publishing, remote or repository settings changes. The gate runs in CI, not inside
`internal/ci`, because a mutation run takes minutes.
