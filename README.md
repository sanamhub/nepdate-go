# nepdate-go

Bikram Sambat (Nepali calendar) for Go. Part of the NepDate family: shared data and spec live
in the hub [`sanamhub/nepdate`](https://github.com/sanamhub/nepdate); the C# original is
[NepDate](https://github.com/RajuPrasai/NepDate).

- Module `github.com/sanamhub/nepdate-go`, package `nepdate`. The patro (पात्रो, the Nepali
  almanac: tithi, public holidays and events) is a separate package of the same module,
  `github.com/sanamhub/nepdate-go/patro` (hub ADR-0004). The core never imports it, so a program
  that only converts dates links no patro data.
- Zero dependencies. Zero allocations on conversion, arithmetic, comparison, successful parsing,
  `Append*` formatting and patro lookups (ADR-0002, measured with `-benchmem`).
- BS 1901-01-01 to 2199-12-30 (AD 1844-04-11 to 2143-04-15). Same results as C# NepDate and every
  other port, except the deviations in [shared/spec/PARITY.md](shared/spec/PARITY.md), proved by
  the shared vectors.
- Patro data covers BS 2077-01-01 to 2084-08-20 only
  ([shared/spec/PATRO.md](shared/spec/PATRO.md)).

**Status:** planning complete, implementation not started (next: phase G0).

## Where to start

| You are | Read |
|---|---|
| an AI agent implementing a phase | [AGENTS.md](AGENTS.md) → [docs/plan/00-ROADMAP.md](docs/plan/00-ROADMAP.md) |
| reviewing the API | [docs/API.md](docs/API.md) |
| comparing with other Go libraries | [docs/competitors.md](docs/competitors.md) |
| wondering "why is it like this?" | [docs/adr/README.md](docs/adr/README.md) |
| looking for the algorithm or data | [shared/spec/README.md](shared/spec/README.md) (changes go to the hub) |

## Layout

```
shared/          hub copy: data/ (calendar JSON), spec/ (specs, vectors), VERSION (hub commit)
*.go             package nepdate               (created in phase G0)
patro/           package patro                 (phase G4)
internal/gen/    table generator, run by go generate
internal/ci/     the CI program: go run ./internal/ci
benchmarks/      separate module: competitor benchmarks (phase G5)
docs/            API.md, competitors.md, adr/, plan/, runbooks/
```

## License

MIT ([LICENSE](LICENSE)). Calendar data from NepDate by Raju Prasai and contributors, MIT
([shared/data/NOTICE](shared/data/NOTICE)). Not affiliated with the Government of Nepal.
