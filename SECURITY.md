# Security policy

`nepdate-go` is a library with no dependencies, no network access and no file access. A security
problem is most likely a panic or an unbounded allocation on hostile input to `Parse`,
`ParseLenient` or `UnmarshalText`, or a problem in a workflow (`.github/workflows/`).

## Supported versions

The latest `v0.x` release gets fixes. Older tags are retracted in `go.mod` when they have a known
problem (ADR-0005).

## Reporting a vulnerability

Report privately through
[GitHub security advisories](https://github.com/sanamhub/nepdate-go/security/advisories/new).
Don't open a public issue. Expect a first reply within a week. Say which version and function are
affected and give an input that reproduces it.

A wrong date, holiday or tithi isn't a security problem: report it in the hub,
[`sanamhub/nepdate`](https://github.com/sanamhub/nepdate/issues/new/choose).
