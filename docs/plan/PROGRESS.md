# Progress

One row per task (handover §3 step 7). The commit column names the task's own commit; a row is
written in the commit after it when the hash is not known yet.

| Task | Status | Commit | CI last line | Notes and open questions |
|---|---|---|---|---|
| G0-01 | done | c13098a | n/a (CI program comes in G0-03) | `go build`, `go vet`, `go list -m all` pass |
| G0-02 | done | (next commit) | n/a (CI program comes in G0-03) | two runs give identical bytes; `-check` exits 1 and names `calendar_gen.go` after one byte changes |
