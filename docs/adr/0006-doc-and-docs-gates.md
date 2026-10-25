# ADR-0006: Doc comment and docs gates

- Status: Accepted
- Date: 2026-10-18
- Deciders: Sanam Pakuwal
- Amends: ADR-0004 (gate list)

## Context

ADR-0004 lists six CI gates. Publish readiness showed two kinds of defect that none of them catch.
First, five exported struct fields had no doc comment, and revive passed because it checks only
declarations and methods. Second, the user-facing pages had three broken relative links and an
internal tracking ID that means nothing to a reader.

## Decision

`go run ./internal/ci` runs two more gates after the six from ADR-0004.

7. **Doc comments.** Every exported constant, variable, type, function, method and struct field in
   the module's packages has a doc comment. The gate runs `go doc -all` to prove the docs render,
   then checks the same declarations with `go/ast`, which sees fields that revive does not.
8. **Docs.** Every Markdown page under the repository root list (README, CHANGELOG, CONTRIBUTING,
   SECURITY) and under `docs/` has no em dash (U+2014), and every inline relative link resolves to
   a file inside the repository. User-facing pages, which are the root list plus every page under
   `docs/` except `docs/adr/`, also contain no task ID. A new page under `docs/`
   is user-facing by default.

The release workflow runs the same eight gates, and also the consumer check
(`tools/consumer-check`), which builds a separate module against this one with `GOPROXY=off`.

## Consequences

- Adding an exported identifier without a doc comment, or a docs page with a stale link, fails CI
  locally and on GitHub.
- ADRs are not checked for task IDs.
- The docs gate checks inline links only, not link fragments or reference-style links. En dashes,
  as in `1901–2199`, are allowed.
