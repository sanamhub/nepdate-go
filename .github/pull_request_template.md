<!-- What this changes and why, in two or three sentences. Name the issue (Closes #12). -->

## What to look at first

-

## How it was tested

## Checks

- [ ] `go run ./internal/ci` passes; last line:
- [ ] Tests added or updated for the change
- [ ] No exported identifier beyond `docs/API.md`, no dependency in the root module
- [ ] Allocation tests still report 0 for the paths in ADR-0002 §5
- [ ] `CHANGELOG.md` `[Unreleased]` updated, if a user would notice
