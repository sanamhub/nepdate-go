package nepdate_test

import (
	"fmt"
	"log"
	"time"

	nepdate "github.com/sanamhub/nepdate-go"
)

func ExampleNew() {
	d, err := nepdate.New(2081, nepdate.Shrawan, 32)
	if err != nil {
		log.Fatal(err)
	}
	y, m, day := d.Gregorian()
	fmt.Println(y, int(m), day)
	// Output: 2024 8 16
}

func ExampleDate_Format() {
	d, _ := nepdate.New(2081, nepdate.Shrawan, 15)
	fmt.Println(d.Format("MMM d, yyyy", nepdate.Nepali))
	// Output: साउन १५, २०८१
}

func ExampleDate_AppendLong() {
	d, _ := nepdate.New(2081, nepdate.Shrawan, 15)
	var buf []byte
	buf = d.AppendLong(buf, nepdate.LongOptions{Weekday: true}, nepdate.Nepali)
	fmt.Println(string(buf))
	// Output: मङ्गलवार, साउन १५, २०८१
}

func ExampleParseLenient() {
	d, err := nepdate.ParseLenient("15 Shrawan 2080")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(d)
	// Output: 2080/04/15
}

func ExampleFromTime() {
	d, err := nepdate.FromTime(time.Date(2024, time.August, 16, 0, 0, 0, 0, time.UTC))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(d)
	// Output: 2081/04/32
}
