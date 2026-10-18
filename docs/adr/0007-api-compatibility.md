# ADR-0007: API compatibility gate

- Status: Accepted
- Date: 2026-10-18
- Deciders: Sanam Pakuwal
- Amends: ADR-0004 (gate list), ADR-0006 (gate list)

## Context

ADR-0004 and ADR-0006 define eight gates. None reads the exported surface as a binary contract:
the doc gate proves that every exported identifier has a comment, but not that a signature stayed
the same, and `docs/API.md` is checked only by review. A renamed function, a changed parameter or
an undocumented addition can pass all eight.

The module has two user-facing packages. `nepdate` and `patro` are the API; `internal/` is not
importable outside the module, and `tools/` holds commands rather than library packages.

## Decision

`go run ./internal/ci` runs a ninth gate after the docs gate:

9. **API compatibility.** `go run ./tools/api-check` compares each package with a committed apidiff
   export-data snapshot: `testdata/api/nepdate.snapshot` (7,486 bytes) and
   `testdata/api/patro.snapshot` (6,387 bytes). A removed or changed identifier fails. A compatible
   addition also fails until the snapshot is rewritten with
   `go run ./tools/api-check -update`. The gate then checks every exported declaration, method and
   struct field in both packages against `docs/API.md`, so an addition cannot be accepted by
   `-update` unless the contract page names it. Functions, methods and types must appear as their
   declaration (`func Name(`, `func (r Recv) Name(`, `type Name`); constants, variables and fields
   as a whole word.

The tool pins `golang.org/x/exp/cmd/apidiff` at
`v0.0.0-20261014143641-f0251eb4aa3e` and fetches it with `go run`, so the root module keeps zero
dependencies and no `go.sum`. Package snapshots are used instead of `apidiff -m` because module mode
also reports the command packages under `tools/`; only `nepdate` and `patro` are the public API.

`ci.yml` and `release.yml` each get an `api` job, and `release` needs it. The gate is also inside the
matrix `ci` job, so the local and CI gate lists stay the same.

## Consequences

- Adding, removing or changing an exported identifier fails locally and on GitHub.
- An addition needs two edits in the same change: `docs/API.md` and `-update` of the snapshot.
- The first run on a machine without the apidiff module cached needs the Go module proxy. A failed
  download fails the gate; it is not skipped.
- The snapshots are binary apidiff export data. A merge conflict in them is resolved by choosing
  either side and running `-update`, not by hand editing.
- Changing the tool pin re-runs `-update` so the snapshot format matches the tool.
