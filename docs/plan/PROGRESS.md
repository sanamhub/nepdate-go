# Progress

One row per task (handover §3 step 7). The commit column names the task's own commit; the hash
is filled in by the next commit.

| Task | Status | Commit | CI last line | Notes and open questions |
|---|---|---|---|---|
| G0-01 | done | c13098a | n/a (CI program comes in G0-03) | `go build`, `go vet`, `go list -m all` pass |
| G0-02 | done | 762e16c | n/a (CI program comes in G0-03) | two runs give identical bytes; `-check` exits 1 and names `calendar_gen.go` after one byte changes |
| G0-03 | done | 546b74a | `ci: 5/5 gates passed` | lint and govulncheck skipped locally (not installed yet); `release-check v9.9.9` and `release-check 0.1.0` exit 1 |
| G0-04 | done | 4b5b802 | `ci: 5/5 gates passed` | golangci-lint 2.14.0 and govulncheck 1.8.0 installed with `go install`; lint reports 0 issues; gosec findings in the dev tools are suppressed with a reason each |
| G0-05 | blocked (partial) | e024eae | `ci: 5/5 gates passed` | Done: `ci.yml` (os x Go 1.26/1.27 plus linux/386, lint and govulncheck jobs, actions pinned by SHA), Dependabot, PR and issue templates, `SECURITY.md`, `CONTRIBUTING.md`. Blocked on hub H3: the `shared-check` job in `ci.yml` and `shared-sync.yml` need the hub reusable workflows. Not verifiable locally: "CI green on the first PR" needs the GitHub repository (human task). |
| G0 checkpoint | ready for review | e024eae | `ci: 5/5 gates passed` (clean clone, Windows, Go 1.27.1; gate 3 warned and ran without `-race`, `CGO_ENABLED=0`) | G0-05 partial: shared-check and shared-sync wait for hub H3 |
