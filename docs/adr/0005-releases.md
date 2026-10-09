# ADR-0005: Releases

- Status: Accepted
- Date: 2026-09-25
- Deciders: Sanam Pakuwal

## Decision

- One module, so one tag releases both packages, `nepdate` and `patro`, with the same version
  (hub ADR-0004 §3). A data update that only changes `patro` is still a new patch tag.
- Semver tags `vX.Y.Z` on `main`, signed (SSH or GPG). Stay on v0 until the API settles: a v2 would
  need a `/v2` module path.
- `CHANGELOG.md` (Keep a Changelog). `go run ./internal/ci release-check vX.Y.Z` checks that the
  changelog has the section; Go has no version field to compare.
- `release.yml` on tag `v*`: preflight (release-check), full CI, then `environment: production`
  (required reviewer) creates the GitHub Release with notes from the changelog section, then asks
  the proxy to index the version:
  `GOPROXY=proxy.golang.org go list -m github.com/sanamhub/nepdate-go@vX.Y.Z`.
- No artifacts to sign: a Go library ships as source, and `sum.golang.org` records its hash.
  OpenSSF Scorecard runs weekly.
- **Never move or delete a published tag** (the checksum database would reject it). A bad release is
  fixed with a new tag plus a `retract` directive in `go.mod`.
- The weekly `shared-sync` PR, once merged, is released as a patch version.

## Consequences

- Nothing to publish except a tag; pkg.go.dev picks it up.
