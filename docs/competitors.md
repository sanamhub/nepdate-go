# Other Go libraries for the Bikram Sambat calendar

Checked 2026-09-25 against GitHub, pkg.go.dev, proxy.golang.org and the sources; rechecked
2026-10-07 (pkg.go.dev searches `nepali`, `nepali date`, `bikram sambat`, `patro`; GitHub search
`nepali date language:Go` and `bikram sambat language:Go`; `@latest` on proxy.golang.org for each
module). Used by task G5-02 (benchmarks). None covers NepDate's range (1901–2199) or its features
(parsing, fiscal year, tithi). The awesome-go list has no entry.

## Native support: none

| Source | Bikram Sambat? | Checked |
|---|---|---|
| Go standard library `time` | No. Proleptic Gregorian only; since Go 1.23 its calendar arithmetic uses Neri and Schneider's formulas instead of tables (`src/time/time.go`, Go 1.27.1) | 2026-10-07 |
| `golang.org/x/text` v0.42.0 (2026-09-08) | No. Its only `date` package holds CLDR v32 time zone ids; no calendar systems | 2026-10-07 |
| Unicode CLDR / ICU | No. CLDR's `ca` key lists 18 calendars (`indian` is the Saka calendar); none is Bikram Sambat (https://github.com/unicode-org/cldr/blob/main/common/bcp47/calendar.xml). Go has no ICU binding in the standard library either | 2026-10-07 |
| `github.com/gilramir/gobabel/locale/ne` v0.1.0 (2026-09-27, MIT) | No. Registers CLDR's Nepali locale data (Gregorian month names, numbers) for `gobabel/dates` | 2026-10-07 |

So the Go port builds the calendar itself; there is nothing to delegate to.

## Libraries we benchmark against

| Module (pinned in `benchmarks/go.mod`) | License | BS range | Approach | Cost per conversion | State |
|---|---|---|---|---|---|
| `github.com/opensource-nepal/go-nepali v0.4.1` | GPL-3.0 | 2000–2099 | `[100][12]int8` month lengths | walks year by year; returns `*[3]int` (1 heap allocation) | stalled since 2025-04; most importers (5) |
| `github.com/srishanbhattarai/nepcal v1.3.0` | MIT | 1975–2100 | `map[int][]int` | linear walk + `time.Time` arithmetic; `fmt.Sprintf` in comparisons | abandoned (no release since 2021); most stars (~48) |
| `github.com/suprimkhatri77/go-bs v0.7.0` | MIT | 1979–2100 | generated `[122][12]uint8` | sums 12 months per year while walking, O(years × 12) | first release 2026-09, v0.7.0 on 2026-09-23, one author, no importers; closest in shape to ours |
| `github.com/JankariTech/GoBikramSambat` (pseudo-version 2020-05-07) | none (benchmark locally only, don't redistribute) | 1970–2100 | `map[int][13]int` | walks from 1 January; returns an interface | abandoned |

Calls to benchmark:

| Module | BS → AD | AD → BS |
|---|---|---|
| go-nepali | `dateConverter.NepaliToEnglish(y, m, d)` → `(*[3]int, error)` | `dateConverter.EnglishToNepali(y, m, d)` |
| nepcal | `nepcal.Date(y, nepcal.Month(m), d)` then `.Gregorian()` (conversion happens in `Date`) | `nepcal.FromGregorian(t)` |
| go-bs | `bs.BSToAD(bs.Date{Year: y, Month: m, Day: d})` → `time.Time` | `bs.ADToBS(t)` |
| GoBikramSambat | `bsdate.New(d, m, y)` then `.GetGregorianDate()` (day, month, year order) | `bsdate.NewFromGregorian(d, m, y)` |

Optional (string API, parse + convert): `github.com/samit22/calendarN/dateconv v1.4.0` (MIT):
`(&dateconv.Converter{}).EtoN("2006-01-02")`, `.NtoE(...)`. Not benchmarked by default: pulls cobra
and viper into the bench module.

## Correctness (why the table needs an "agrees" column)

Years whose month lengths differ from NepDate (our `shared/data/`):

| Module | Differing years |
|---|---|
| go-bs | 14: 1981, 1989, 1993, 1998, 2090–2099 |
| nepcal | 19: 1977, 1981, 1989, 1993, 1998, 2004, **2083**, 2087, 2090–2100 |
| GoBikramSambat | 30, including corrupt rows in 2030–2032 (`{32,32,…}`), 2055, 2062, 2081–2100 |
| go-nepali | 8: 2062, 2084–2090 |

nepcal's 2083 has Ashoj = 30 where NepDate has 31: if NepDate is right, nepcal is a day off from
about 2026-10-17. Other bugs seen in the sources: nepcal checks only the lower range bound and
not the day against the month length (issue #79) and can panic on a missing map key; go-nepali
accepts AD 2023-02-30; mitigo (`github.com/orted-org/mitigo`, no license, abandoned) has no upper
bound check in `BStoAD`.

## Not benchmarked

CLI-only or package `main`: `poudel/bsgo` (archived), `sandeepdhakal/go-nepali-date` (GPL),
`UdeshyaDhungana/nepdate`, `Ulleri-Tech/nepali-date-cli`, `KishanAdhikari11/Nepalidatecli`,
`abhiyankhanal/Date-CLI`, `rabinpoudyal/aaja`, `SunilNeupane77/nepali`,
`sumanchapai/patro` v1.2.1 (CLI over JSON scraped from Hamro Patro; archived 2026-07-17).
A service, not a library: `Sarojdhakal307/bs-calendar` (2026-09-28, MIT; REST API over a PostgreSQL
table of BS 1975–2100, 2084 onward marked projected). A bot: `mesaugat/bikram-sambat-progress`.
Not a date library: `PrameshKarki/nepali_utils`.

## Sources

https://github.com/opensource-nepal/go-nepali, https://github.com/srishanbhattarai/nepcal,
https://github.com/SuprimKhatri77/go-bs, https://github.com/JankariTech/GoBikramSambat,
https://github.com/orted-org/mitigo, https://github.com/samit22/calendarN,
https://github.com/topics/nepali-date?l=go, proxy.golang.org `@latest` for each module,
https://pkg.go.dev/search?q=nepali, https://pkg.go.dev/search?q=bikram+sambat,
https://pkg.go.dev/golang.org/x/text@v0.42.0/date, https://pkg.go.dev/github.com/gilramir/gobabel/locale/ne,
https://github.com/sumanchapai/patro, https://github.com/Sarojdhakal307/bs-calendar,
https://doi.org/10.1002/spe.3172 (Neri and Schneider).
