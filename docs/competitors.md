# Other Go libraries for the Bikram Sambat calendar

Checked 2026-11-02 against pkg.go.dev, the Go module proxy, GitHub and the sources. The selection
follows [`shared/spec/BENCHMARK.md`](../shared/spec/BENCHMARK.md) section 7. Usage is the "Imported
by" count on pkg.go.dev, read on 2026-11-02. The bar is 3 importing packages, first release at least
12 months before 2026-11-02, a release or commit in the last 36 months, a license, not archived or
deprecated, and a build on the current toolchain.

## Candidate table

| Module | Version | Imported by | First release | Last change | License | Verdict |
|---|---|---:|---|---|---|---|
| `github.com/opensource-nepal/go-nepali` | v0.4.1 | 5 | 2021 | 2025-04-07 | GPL-3.0 | selected |
| `github.com/srishanbhattarai/nepcal` | v1.3.0 | 1 | 2021 | 2021-12-02 | MIT | fails 3 importers and the 36-month bar |
| `github.com/suprimkhatri77/go-bs` | v0.7.0 | 0 | 2026-09-23 | 2026-09-23 | MIT | fails 3 importers and the 12-month bar |
| `github.com/JankariTech/GoBikramSambat` | pseudo-version 2020-05-07 | 0 | 2020 | 2020-05-07 | none | fails a license, 3 importers and the 36-month bar |
| `github.com/samit22/calendarN/dateconv` | v1.4.0 | 0 | 2022 | 2026 | MIT | fails 3 importers; pulls cobra and viper into the bench module |
| `github.com/orted-org/mitigo` | pseudo-version | 0 | 2020 | 2020 | none | fails a license, 3 importers and the 36-month bar |

CLI-only programs are not libraries: `github.com/sandeepdhakal/go-nepali-date`,
`github.com/SunilNeupane77/nepali`, `github.com/sumanchapai/ere`,
`github.com/mesaugat/bikram-sambat-progress`, `github.com/opensource-nepal/ad2bs` and
`github.com/opensource-nepal/bs2ad` are package `main` and are not candidates.

## Selected library

`github.com/opensource-nepal/go-nepali` v0.4.1 is the only candidate that passes. The port measures
B1 with `dateConverter.NepaliToEnglish(y, m, d)` and B2 with
`dateConverter.EnglishToNepali(y, m, d)`. It returns a `*[3]int` and one heap allocation per
conversion. Its AD range starts at 1944-01-01, so the first AD input, 1943-04-14, is rejected and
counts as a disagreement.

## Correctness

The 64 benchmark dates include years where go-nepali's month table differs from `shared/data`:
2062, 2084, 2085, 2086, 2087, 2088, 2089 and 2090. The report lists the exact dates and counts. A
fast wrong answer is listed, never ranked first.

## Sources

https://pkg.go.dev/github.com/opensource-nepal/go-nepali/dateConverter?tab=importedby,
https://pkg.go.dev/github.com/srishanbhattarai/nepcal/nepcal?tab=importedby,
https://pkg.go.dev/github.com/suprimkhatri77/go-bs?tab=importedby,
https://pkg.go.dev/github.com/JankariTech/GoBikramSambat,
https://github.com/opensource-nepal/go-nepali, https://github.com/srishanbhattarai/nepcal,
https://github.com/SuprimKhatri77/go-bs, https://github.com/samit22/calendarN,
https://proxy.golang.org for each module's versions.
