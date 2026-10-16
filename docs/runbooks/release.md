# Runbook: release `nepdate-go`

Decided in ADR-0005. Owner: maintainer. Approver: a second person.

## Before tagging

1. `main` is green. `go run ./internal/ci` passes locally.
2. `CHANGELOG.md` has `## [X.Y.Z] - YYYY-MM-DD`, moved out of `[Unreleased]`.
3. `go run ./internal/ci release-check vX.Y.Z` passes.

## Dry run (no tag)

`release.yml` also has a `workflow_dispatch` input `version`. Running it with `version: vX.Y.Z`
runs the preflight and full CI and skips the release and proxy jobs, so a version can be checked
before the human tags it. Trigger it from the Actions tab, or run the preflight locally with
`go run ./internal/ci release-check vX.Y.Z`. The dispatch path needs the GitHub remote.

## Releasing

```bash
git tag -s vX.Y.Z -m "vX.Y.Z"
git push origin vX.Y.Z
```

`release.yml` runs preflight and CI, waits for the `production` approval, creates the GitHub Release
from the changelog section, and asks proxy.golang.org to fetch the version.

## After releasing

1. `go list -m github.com/sanamhub/nepdate-go@vX.Y.Z` works from a clean machine.
2. pkg.go.dev shows the version for both `github.com/sanamhub/nepdate-go` and
   `github.com/sanamhub/nepdate-go/patro` (it can take a few minutes). One tag releases both.
3. Add an empty `## [Unreleased]` to the changelog.

## Rolling back

Never move or delete the tag. Add `retract vX.Y.Z // reason` to `go.mod`, release `vX.Y.(Z+1)` with
the fix, and note both in the changelog.
