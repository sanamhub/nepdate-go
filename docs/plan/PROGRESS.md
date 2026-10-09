# Progress

One row per task (handover §3 step 7). The commit column names the task's own commit; a row is
written in the commit after it when the hash is not known yet.

| Task | Status | Commit | CI last line | Notes and open questions |
|---|---|---|---|---|
| G0-01 | done | c13098a | n/a (CI program comes in G0-03) | `go build`, `go vet`, `go list -m all` pass |
| G0-02 | done | 762e16c | n/a (CI program comes in G0-03) | two runs give identical bytes; `-check` exits 1 and names `calendar_gen.go` after one byte changes |
| G0-03 | done | 546b74a | `ci: 5/5 gates passed` | lint and govulncheck skipped locally (not installed yet); `release-check v9.9.9` and `release-check 0.1.0` exit 1 |
| G0-04 | done | (next commit) | `ci: 5/5 gates passed` | golangci-lint 2.14.0 and govulncheck 1.8.0 installed with `go install`; lint reports 0 issues; gosec findings in the dev tools are suppressed with a reason each |
