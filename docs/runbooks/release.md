# Runbook: release `nepdate-go`

Decided in ADR-0005. Owner: maintainer. Approver: a second person.

## Before tagging

1. `main` is green. `go run ./internal/ci` passes locally, and the commit's `ci` run, including the
   `mutation` job, is green. A tag is only made from such a commit. `release.yml` runs the mutation
   gate and the 3-minute fuzz budget again, and the `release` job needs both, so a red baseline or
   a crasher stops the release either way.
2. `CHANGELOG.md` has `## [X.Y.Z] - YYYY-MM-DD`, moved out of `[Unreleased]`.
3. `go run ./internal/ci release-check vX.Y.Z` passes. It checks the changelog section, the README
   import path and Go line, the ADR-0008 pin against `ci.yml`, `shared/VERSION`, the shared data
   NOTICE and that `go list -m all` names only this module.
4. `go run ./tools/data-status` passes. It prints the last BS and AD day of tithi, holiday and
   event data and the days left from today, the current UTC date (`-today=YYYY-MM-DD` takes an AD
   date instead). Holidays under 183 days exit 1 and block the release; tithi and events under 183
   days only warn, so refresh the data from the hub before they run out. `release.yml` runs both
   checks in preflight.

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

`release.yml` runs preflight, CI, the API snapshot check and the fuzz budget, waits for the
`production` approval, creates the GitHub Release from the changelog section, and asks
proxy.golang.org to fetch the version.

## After releasing

1. `go list -m github.com/sanamhub/nepdate-go@vX.Y.Z` works from a clean machine.
2. pkg.go.dev shows the version for both `github.com/sanamhub/nepdate-go` and
   `github.com/sanamhub/nepdate-go/patro` (it can take a few minutes). One tag releases both.
3. Add an empty `## [Unreleased]` to the changelog.

## Rolling back

Never move or delete the tag. Add `retract vX.Y.Z // reason` to `go.mod`, release `vX.Y.(Z+1)` with
the fix, and note both in the changelog.
