# nepdate-go

Bikram Sambat (Nepali) calendar for Go. Part of the NepDate family: shared data and the
behaviour spec live in the hub [sanamhub/nepdate](https://github.com/sanamhub/nepdate); the C#
original is [NepDate](https://github.com/RajuPrasai/NepDate).

- Package `nepdate`, module `github.com/sanamhub/nepdate-go`. Patro (पात्रो, the Nepali almanac:
  tithi, public holidays and events) is the package `github.com/sanamhub/nepdate-go/patro` of the
  same module. The core never imports it, so a program that only converts dates links no patro
  data.
- Zero dependencies. Zero allocations on conversion, arithmetic, comparison, successful parsing,
  `Append*` formatting and patro lookups (measured with `-benchmem`).
- BS 1901-01-01 to 2199-12-30 (AD 1844-04-11 to 2143-04-15). Patro data covers BS 2077-01-01 to
  2084-08-20. Same results as C# NepDate and the other ports, except the deviations in
  [shared/spec/PARITY.md](shared/spec/PARITY.md).

## Install

```
go get github.com/sanamhub/nepdate-go
```

Package docs: <https://pkg.go.dev/github.com/sanamhub/nepdate-go>

## Quick start

The snippets use `fmt`, `log` and `time` from the standard library, plus the two imports shown
above. Construct a BS date and read its Gregorian date:

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

## License

MIT ([LICENSE](LICENSE)). Calendar data from NepDate by Raju Prasai and contributors, MIT
([shared/data/NOTICE](shared/data/NOTICE)). Not affiliated with the Government of Nepal.
