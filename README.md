<p align="center">
  <img src="https://raw.githubusercontent.com/sanamhub/nepdate-go/main/docs/assets/nepdate-go-128.png" width="128" alt="NepDate for Go logo">
</p>

<h1 align="center">nepdate-go</h1>

<p align="center">
  <a href="https://github.com/sanamhub/nepdate-go/actions/workflows/ci.yml"><img src="https://github.com/sanamhub/nepdate-go/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://pkg.go.dev/github.com/sanamhub/nepdate-go"><img src="https://pkg.go.dev/badge/github.com/sanamhub/nepdate-go.svg" alt="Go Reference"></a>
  <a href="https://github.com/sanamhub/nepdate-go/releases"><img src="https://img.shields.io/github/v/release/sanamhub/nepdate-go?include_prereleases&sort=semver&logo=go" alt="Release"></a>
  <a href="https://github.com/sanamhub/nepdate-go/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue" alt="License: MIT"></a>
</p>

Bikram Sambat (Nepali) calendar for Go. Part of the NepDate family: shared data and the
behaviour spec live in the hub [sanamhub/nepdate](https://github.com/sanamhub/nepdate); the C#
original is [NepDate](https://github.com/RajuPrasai/NepDate).

> **Status:** 0.2.0-rc.1, a release candidate. The API is listed in
> [docs/API.md](https://github.com/sanamhub/nepdate-go/blob/main/docs/API.md) and may still change
> before 0.2.0. Patro data (tithi, holidays, events) is informational, not an official calendar.

Go 1.24 or later. CI runs the tests on Go 1.24.13 and the full gates on Go 1.26 and 1.27.

- Package `nepdate`, module `github.com/sanamhub/nepdate-go`. Patro (पात्रो, the Nepali almanac:
  tithi, public holidays and events) is the package `github.com/sanamhub/nepdate-go/patro` of the
  same module. The core never imports it, so a program that only converts dates links no patro
  data.
- Zero dependencies. Zero allocations on conversion, arithmetic, comparison, successful parsing,
  `Append*` formatting and patro lookups (measured with `-benchmem`).
- BS 1901-01-01 to 2199-12-30 (AD 1844-04-11 to 2143-04-15). Patro data covers BS 2077-01-01 to
  2084-08-20. Same results as C# NepDate and the other ports, except the deviations in
  [shared/spec/PARITY.md](https://github.com/sanamhub/nepdate-go/blob/main/shared/spec/PARITY.md).

## Install

```
go get github.com/sanamhub/nepdate-go@v0.2.0-rc.1
```

```go
import (
	"github.com/sanamhub/nepdate-go"       // package nepdate
	"github.com/sanamhub/nepdate-go/patro" // only if you need tithi, holidays or events
)
```

Package docs: <https://pkg.go.dev/github.com/sanamhub/nepdate-go>

## Quick start

The snippets use `fmt`, `log`, `time` and `errors` from the standard library, plus the imports
above.
Construct a BS date and read its Gregorian date:

```go
	d, err := nepdate.New(2081, nepdate.Shrawan, 32)
	if err != nil {
		log.Fatal(err)
	}
	y, m, day := d.Gregorian()
	fmt.Println(y, int(m), day)
```

Format in Nepali:

```go
	d, _ := nepdate.New(2081, nepdate.Shrawan, 15)
	fmt.Println(d.Format("MMM d, yyyy", nepdate.Nepali))
```

Append a long form without allocating:

```go
	d, _ := nepdate.New(2081, nepdate.Shrawan, 15)
	var buf []byte
	buf = d.AppendLong(buf, nepdate.LongOptions{Weekday: true}, nepdate.Nepali)
	fmt.Println(string(buf))
```

Parse what a person wrote:

```go
	d, err := nepdate.ParseLenient("15 Shrawan 2080")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(d)
```

Convert a Go time:

```go
	d, err := nepdate.FromTime(time.Date(2024, time.August, 16, 0, 0, 0, 0, time.UTC))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(d)
```

Read the almanac for a date:

```go
	d, _ := nepdate.New(2081, nepdate.Baishakh, 1)
	i := patro.Of(d)
	t, _ := i.Tithi()
	fmt.Println(t.Name(nepdate.Nepali), i.IsPublicHoliday())
	for n := 0; n < i.EventCount(); n++ {
		fmt.Println(i.Event(n).Name(nepdate.English))
	}
```

A Date marshals to JSON as `"YYYY-MM-DD"`. Saving an unset (zero) Date returns an error, so use
`*Date` for an optional field; nil marshals as null.

## More of the API

The full surface, in the same groups as the shared spec, is in
[docs/API.md](https://github.com/sanamhub/nepdate-go/blob/main/docs/API.md).

Dates and conversion:

```go
d, _ := nepdate.New(2081, nepdate.Shrawan, 15)
y, m, day := d.Gregorian()
n, _ := nepdate.FromYMD(d.YMD())
fmt.Println(y, int(m), day, n.YearLength())
// 2024 7 30 366
```

Arithmetic and differences:

```go
d := nepdate.MustNew(2081, nepdate.Shrawan, 32)
clamped, _ := d.AddMonths(2, nepdate.Clamp)
spilled, _ := d.AddMonths(2, nepdate.Spill)
fmt.Println(clamped, spilled, d.DaysUntil(nepdate.Max))
// 2081/06/30 2081/07/02 43340
```

Formatting and parsing:

```go
d := nepdate.MustNew(2081, nepdate.Shrawan, 15)
fmt.Println(d.Format("MMM d, yyyy", nepdate.Nepali))
parsed, ok := nepdate.TryParseLenient("15 Shrawan 2081")
fmt.Println(parsed, ok)
// साउन १५, २०८१
// 2081/04/15 true
```

Ranges and fiscal years:

```go
r, _ := nepdate.FromDayCount(nepdate.MustNew(2081, nepdate.Shrawan, 10), 11)
fy := nepdate.FiscalYearOf(r.Start())
fmt.Println(r.Format(nepdate.YMD, nepdate.Dash, nepdate.English), fy.Label(nepdate.English))
// 2081-04-10 - 2081-04-20 2081/82
```

Errors and the zero value:

```go
var d nepdate.Date
_, err := d.AddDays(1)
fmt.Println(d.IsZero(), errors.Is(err, nepdate.ErrOutOfRange))
// true true
```

## More

| Topic | Where |
|---|---|
| Guides and examples | [nepdate.com/docs/go](https://nepdate.com/docs/go/) |
| Every exported identifier | [docs/API.md](https://github.com/sanamhub/nepdate-go/blob/main/docs/API.md) |
| Speed and allocations, with the C# baseline | [docs/benchmarks.md](https://github.com/sanamhub/nepdate-go/blob/main/docs/benchmarks.md) |
| What changed in each version | [CHANGELOG.md](https://github.com/sanamhub/nepdate-go/blob/main/CHANGELOG.md) |
| Building, gates and pull requests | [CONTRIBUTING.md](https://github.com/sanamhub/nepdate-go/blob/main/CONTRIBUTING.md) |
| Reporting a vulnerability | [SECURITY.md](https://github.com/sanamhub/nepdate-go/blob/main/SECURITY.md) |
| A wrong date, holiday or tithi | the hub's [data error form](https://github.com/sanamhub/nepdate/issues/new/choose) |
| The other languages | [sanamhub/nepdate](https://github.com/sanamhub/nepdate#packages) |

## License

MIT ([LICENSE](https://github.com/sanamhub/nepdate-go/blob/main/LICENSE)). Calendar data from
NepDate by Raju Prasai and contributors, MIT
([shared/data/NOTICE](https://github.com/sanamhub/nepdate-go/blob/main/shared/data/NOTICE)). Not
affiliated with the Government of Nepal.
