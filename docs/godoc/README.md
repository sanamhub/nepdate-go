# Generated package documentation

These files are the `go doc -all` output for `nepdate` and `patro`. They are committed so a change
to a doc comment shows up in review with the code that caused it.

They were generated with Go 1.27.1. Regenerate from the repository root:

```sh
go doc -all . > docs/godoc/nepdate.txt
go doc -all ./patro > docs/godoc/patro.txt
```

`TestGoDocPreview` compares the files with the command output byte for byte. On any Go version other
than 1.27.1 the test skips, and the CI summary reports that as a reduced check. The `api` job in
`ci.yml` and `release.yml` pins Go 1.27.1 and sets `NEPDATE_GODOC_REQUIRED=1`, which turns that skip
into a failure, so the files are enforced in CI even though the quality-gate jobs use `1.27`.
