# G6: Publish readiness

Read first: ADR-0004, ADR-0005, `docs/runbooks/release.md`, `shared/spec/BENCHMARK.md` section 6.

Precondition: G5 merged on `main` (05b6db1), including the release workflow fixes: a tag or
dispatch input reaches the shell only through an environment variable, and the release jobs also
require `github.event_name == 'push'`.

This phase proves the module works from outside the repository and that its docs are ready for
pkg.go.dev. It publishes nothing: no tag, no release, no remote and no repository setting changes.

## Tasks

### G6-01 Consumer check
`tools/consumer-check` writes a throwaway module outside the repository, imports
`github.com/sanamhub/nepdate-go` and `github.com/sanamhub/nepdate-go/patro` through a `replace`
directive to this checkout, and runs it with `GOPROXY=off`, so only the local checkout can satisfy
the import. The program converts BS 2081-04-15 and prints the first patro event of that day.

`-dir` keeps the module at a named directory; the default is a temporary directory that is
removed. `-negative` writes the same program with a broken import path and requires the build to
fail, which proves the check runs the program and does not read a cached result.

AC:
- `go run ./tools/consumer-check` exits 0 and prints the date and the event.
- `go run ./tools/consumer-check -negative` exits 0 only when the broken import fails.
- The module directory is outside the repository, and `go list -m` inside it reports the checkout
  as the directory of `github.com/sanamhub/nepdate-go`.
- A CI job runs both modes. The release workflow runs the same job before the release job.

### G6-02 Doc comment gate
Run `go doc -all` on `nepdate` and `patro`, then check the exported declarations with `go/ast`,
the source model that `go/doc` and `go doc` render. Every exported declaration, method and struct
field needs a non-empty doc comment. A declaration group doc covers every name in the group.

Revive checks declarations and methods, not struct fields, so the new gate adds the field check
and fails on any exported identifier that `go doc` shows without a comment.

AC:
- Every exported identifier in both packages has a doc comment.
- A temporary package with a missing function, method, const or field comment fails the check,
  and the error names the identifier. This is the negative control.
- `go run ./internal/ci` ends with `ci: 7/7 gates passed` at this commit.

### G6-03 Docs gate
Check `README.md`, `CHANGELOG.md` and every Markdown file under `docs/`:

- no em dash (U+2014);
- every relative link target exists (a `#fragment` is ignored; absolute URLs are not checked);
- no task ID in a user-facing file.

User-facing files are `README.md`, `CHANGELOG.md`, `docs/API.md`, `docs/benchmarks.md`,
`docs/competitors.md` and `docs/runbooks/`. ADRs and plan files keep task IDs for traceability.

AC:
- The repository's own docs pass.
- A temporary tree with an em dash, a broken relative link or a task ID in a user-facing file
  fails, and the error names the file and line. This is the negative control.
- `go run ./internal/ci` ends with `ci: 8/8 gates passed` at this commit.

### G6-04 B6 strict-parse options
`docs/plan/B6-strict-parse-options.md` gives two or three designs for the B6 miss: 34.80 ns
against C# NepDate's 21.55 ns, with 68.8 % of the profile samples in `scanStrict` and 7.4 % in
`devaAt` on ASCII input. Each option names its expected B6 time, its risk, the tests it needs and
the measurement that would confirm it.

Do not implement an option. The reviewer picks one.

AC:
- The file covers the options with numbers and a recommendation.
- No parser code or test changes in this task.

## Checkpoint (agent stops here for review)

The reviewer checks:

- [ ] `go run ./internal/ci` ends with `ci: 8/8 gates passed`.
- [ ] `go run ./tools/consumer-check` prints one date and one event, and `-negative` reports the
      broken import.
- [ ] The doc gate fails after a doc comment is removed from an exported struct field.
- [ ] The docs gate fails after an em dash, a broken relative link or a task ID is added to a
      user-facing file.
- [ ] `docs/plan/B6-strict-parse-options.md` gives each option an expected B6 number.
- [ ] Nothing was tagged, published, pushed to a new remote or changed in the repository settings.
