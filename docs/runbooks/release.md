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
runs the preflight and full CI and skips the release and verify jobs, so a version can be checked
before the maintainer tags it. Trigger it from the Actions tab, or run the preflight locally with
`go run ./internal/ci release-check vX.Y.Z`. The dispatch path needs the GitHub remote.

## Releasing

```bash
git tag -s vX.Y.Z -m "vX.Y.Z"
git push origin vX.Y.Z
```

`release.yml` runs preflight, CI, the API snapshot check and the fuzz budget, waits for the
`production` approval and creates the GitHub Release from the changelog section, with install and
verification notes. Then it calls `verify-published.yml`, which fetches the version through
proxy.golang.org and sum.golang.org with no replace directive (`go run ./tools/consumer-check
-published vX.Y.Z`) and converts a date and reads a patro event on Linux, macOS and Windows. It
waits up to 30 minutes for the proxy. `verify-published.yml` can also be run by hand from the
Actions tab for any released version, or with `latest`.

`verify-published.yml` also runs on a schedule, always for `latest`. Every Monday at 05:23 UTC it
installs that version on Linux and runs the same conversion and patro check. Every day at 03:17 UTC
it checks the proxy metadata only: the version must still be listed, and its `.info` and `.zip`
files must answer a HEAD request, so no module is downloaded. A failed scheduled run opens or
updates one issue labelled `published-check`, titled `Published nepdate-go <version> does not
install or run`. Create that label once. A green run is silent.

## Pre-releases

A tag with a semver pre-release part, such as `v0.1.0-beta.1`, goes through the same workflow. The
changelog section is `## [0.1.0-beta.1] - YYYY-MM-DD`; a stable section does not satisfy a beta tag
and the other way round. The GitHub Release is marked as a pre-release. The go command prefers a
stable version for `@latest`, but while no stable version exists, `@latest` resolves to the highest
pre-release, so the first beta is what a plain `go get` installs until 0.1.0 ships.

## After releasing

1. `go list -m github.com/sanamhub/nepdate-go@vX.Y.Z` works from a clean machine.
2. pkg.go.dev shows the version for both `github.com/sanamhub/nepdate-go` and
   `github.com/sanamhub/nepdate-go/patro` (it can take a few minutes). One tag releases both.
3. Add an empty `## [Unreleased]` to the changelog.

## Rolling back

Never move or delete the tag. Add `retract vX.Y.Z // reason` to `go.mod`, release `vX.Y.(Z+1)` with
the fix, and note both in the changelog.
